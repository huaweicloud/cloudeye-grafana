package plugin

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/region"
)

func GetCESClient(c *CloudEyeSettings) *ces.CesClient {
	creBuilder := basic.NewCredentialsBuilder().
		WithAk(c.AK).
		WithSk(c.SK)

	httpConfig := config.DefaultHttpConfig().WithIgnoreSSLVerification(true)
	// 单region模式
	if c.ProjectID != "" && c.CESEndpoint != "" {
		creBuilder = creBuilder.WithProjectId(c.ProjectID)
		clientBuilder := ces.CesClientBuilder().
			WithCredential(creBuilder.Build()).
			WithHttpConfig(httpConfig).
			WithEndpoint(c.CESEndpoint)
		return ces.NewCesClient(clientBuilder.Build())
	}

	// 多region模式，只支持华为云, region列表依赖SDK
	clientBuilder := ces.CesClientBuilder().
		WithCredential(creBuilder.Build()).
		WithHttpConfig(httpConfig).
		WithRegion(region.ValueOf(c.Region))
	return ces.NewCesClient(clientBuilder.Build())
}

type CESClient struct {
	Client *ces.CesClient
	Region string
}

type DataQueryParam struct {
	Region     string
	Namespace  string
	DimStr     string
	MetricName string
	Filter     string
	Period     string
	From       int64
	To         int64
	RefID      string
}

func getValueByFilter(v model.DatapointForBatchMetric, filter string) float64 {
	switch filter {
	case "average":
		return *(v.Average)
	case "min":
		return *(v.Min)
	case "max":
		return *(v.Max)
	case "sum":
		return *(v.Sum)
	}
	return 0
}

func (c *CESClient) BatchQuery(refIDs []string, req *model.BatchListMetricDataRequest) (*backend.QueryDataResponse, error) {
	res, err := c.Client.BatchListMetricData(req)
	if err != nil {
		return nil, err
	}

	response := backend.NewQueryDataResponse()
	for i, each := range *(res.Metrics) {
		timeDuration := make([]time.Time, 0, len(each.Datapoints))
		values := make([]float64, 0, len(each.Datapoints))

		for _, v := range each.Datapoints {
			timeDuration = append(timeDuration, time.Unix(v.Timestamp/1000, 0))
			values = append(values, getValueByFilter(v, req.Body.Filter))
		}

		yLabel := map[string]string{
			"namespace": *each.Namespace,
		}
		for _, dim := range *each.Dimensions {
			yLabel[dim.Name] = dim.Value
		}

		fields := []*data.Field{
			data.NewField("time", nil, timeDuration),
			data.NewField(each.MetricName, yLabel, values),
		}
		frame := data.NewFrame("")
		frame.Fields = append(frame.Fields, fields...)
		eachRes := backend.DataResponse{}
		eachRes.Frames = append(eachRes.Frames, frame)
		response.Responses[refIDs[i]] = eachRes
	}
	return response, nil
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func (c *CESClient) ListNamespaces() []string {
	return c.ListMeta(&QueryParam{})
}

func (c *CESClient) ListDims(namespace string) []string {
	return c.ListMeta(&QueryParam{Namespace: namespace})
}

func (c *CESClient) ListMetrics(namespace, dimStr string) []string {
	return c.ListMeta(&QueryParam{Namespace: namespace, DimStr: dimStr})
}

func (c *CESClient) Check() error {
	var limit int32 = 1
	_, err := c.Client.ListAlarms(&model.ListAlarmsRequest{
		Limit: &limit,
	})

	return err
}

func getDimStr(dims []model.MetricsDimension) string {
	var dimsList []string
	for _, dim := range dims {
		dimsList = append(dimsList, fmt.Sprintf("%s,%s", dim.Name, dim.Value))
	}
	return strings.Join(dimsList, ".")
}

type QueryParam struct {
	Region    string
	Namespace string
	DimStr    string
}

func getMetaUtil(param *QueryParam) MetaUtil {
	if param.Namespace == "" && param.DimStr == "" {
		return &NsCache
	}

	if param.DimStr != "" {
		return &MCache
	}

	return &DmCache
}

func (c *CESClient) ListMeta(param *QueryParam) []string {
	metaUtil := getMetaUtil(param)
	metaCache := metaUtil.getCache()
	param.Region = c.Region
	key := metaUtil.buildKey(param)
	reqParam := metaUtil.buildQuery(param)

	isMetaExist, metaList, meta := getMeta(metaCache, key, reqParam)
	if metaList != nil || meta != nil {
		return meta
	}

	newCache := metaUtil.newCachedMeta()

	endFlag := make(chan bool)
	defer close(endFlag)
	go func() {
		for {
			res, err := c.Client.ListMetrics(reqParam)
			if err != nil {
				log.DefaultLogger.Error("ListMetrics error", "detail", err)
				metaUtil.setDefaultMeta(param, newCache)
				metaCache.Data.Store(key, newCache)
				endFlag <- true
				return
			}
			metrics := *(res.Metrics)
			if len(metrics) == 0 {
				newCache.Finished = true
				metaCache.Data.Store(key, newCache)
				endFlag <- true
				return
			}

			procMetaList(metrics, metaUtil, isMetaExist, &metaList)
			reqParam.Start = &(res.MetaData.Marker)
			newCache.Marker = res.MetaData.Marker
			newCache.Meta = metaList
			newCache.UpdateTime = getTimestamp()
		}
	}()
	select {
	case <-time.After(5 * time.Second):
		log.DefaultLogger.Error("ListMeta timeout", "query params", *param)
		if len(metaList) == 0 {
			metaUtil.setDefaultMeta(param, newCache)
			metaCache.Data.Store(key, newCache)
			return metaCache.getCachedMeta(key).Meta
		}
		return metaList
	case <-endFlag:
		return metaList
	}
}

func procMetaList(metrics []model.MetricInfoList, metaUtil MetaUtil, isMetaExist map[string]bool, metaList *[]string){
	for _, metric := range metrics {
		if len(metric.Dimensions) > 3 {
			continue
		}

		element := metaUtil.getRespElem(metric)
		if !isMetaExist[element] {
			isMetaExist[element] = true
			*metaList = append(*metaList, element)
		}
	}
}

func getMeta(metaCache *MetaCache, key string, reqParam *model.ListMetricsRequest) (map[string]bool, []string, []string) {
	isMetaExist := make(map[string]bool)
	var metaList []string
	cachedMeta := metaCache.getCachedMeta(key)
	if cachedMeta != nil {
		if !cachedMeta.isExpired() {
			return nil, nil, cachedMeta.Meta
		}

		// 大租户可能被流控，接着上次的marker继续请求
		if !cachedMeta.Finished {
			reqParam.Start = &cachedMeta.Marker
			metaList = cachedMeta.Meta
			for i := range metaList {
				isMetaExist[metaList[i]] = true
			}
		}
	}
	return isMetaExist, metaList, nil
}
