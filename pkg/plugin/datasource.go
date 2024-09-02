package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

type commonConf struct {
	ProjectID       string `json:"projectId"`
	CESEndpoint     string `json:"cesEndpoint"`
	Region          string `json:"region"`
	MetaConfEnabled bool   `json:"metaConfEnabled"`
}

type CloudEyeSettings struct {
	ProjectID       string `json:"projectId"`
	CESEndpoint     string `json:"cesEndpoint"`
	Region          string `json:"region"`
	MetaConfEnabled bool   `json:"metaConfEnabled"`
	AK              string `json:"accessKey"`
	SK              string `json:"secretKey"`
}

type CustomBatchListMetricDataRequestBody struct {
	model.BatchListMetricDataRequestBody
	RefIDs []string `json:"refIDs"`
	Region string   `json:"region"`
}

func recoverWrapper(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.DefaultLogger.Info(fmt.Sprintf("Path:%s, cost %d ms", r.URL.Path, time.Since(start).Milliseconds()))
			if err := recover(); err != nil {
				log.DefaultLogger.Error("Panic recovered", "err", err)
			}
		}()
		handler(w, r)
	}
}

// NewCloudEye returns datasource.ServeOpts.
func NewCloudEye() datasource.ServeOpts {
	log.DefaultLogger.Info("Creating cloudEye datasource")

	im := datasource.NewInstanceManager(newCloudEyeInstance)
	data := &CloudEyeDatasource{
		im: im,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/regions", recoverWrapper(data.listRegions))
	mux.HandleFunc("/namespaces", recoverWrapper(data.listNamespaces))
	mux.HandleFunc("/dimensions", recoverWrapper(data.listDims))
	mux.HandleFunc("/metrics", recoverWrapper(data.listMetrics))
	mux.HandleFunc("/metric-data", recoverWrapper(data.listMetricData))

	httpResourceHandler := httpadapter.New(mux)
	return datasource.ServeOpts{
		CallResourceHandler: httpResourceHandler,
		QueryDataHandler:    data,
		CheckHealthHandler:  data,
	}
}

type CloudEyeDatasource struct {
	im instancemgmt.InstanceManager
}

func (ds *CloudEyeDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()
	return response, nil
}

func (ds *CloudEyeDatasource) listMetricData(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	cfg, err := LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}

	refIDs, batchReq := buildCustomBatchQueryParams(bodyBytes, cfg)
	cesClient := &CESClient{Client: GetCESClient(cfg)}
	res, err := cesClient.BatchQuery(refIDs, batchReq)
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}
	writeResult(rw, "data", res, nil)
}

func buildCustomBatchQueryParams(reqBodyBytes []byte, setting *CloudEyeSettings) ([]string, *model.BatchListMetricDataRequest) {
	var reqBody CustomBatchListMetricDataRequestBody
	err := json.Unmarshal(reqBodyBytes, &reqBody)
	if err != nil {
		return nil, nil
	}
	setting.Region = reqBody.Region
	return reqBody.RefIDs, &model.BatchListMetricDataRequest{
		Body: &reqBody.BatchListMetricDataRequestBody,
	}
}

func buildHealthCheckRes(err error) *backend.CheckHealthResult {
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (ds *CloudEyeDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	conf, err := LoadSettings(req.PluginContext)
	if err != nil {
		log.DefaultLogger.Error("LoadSettings failed", "err", err.Error())
		return buildHealthCheckRes(err), err
	}

	cesClient := &CESClient{Client: GetCESClient(conf)}
	err = cesClient.Check()
	resetCache()
	return buildHealthCheckRes(err), err
}

func writeResult(rw http.ResponseWriter, path string, val interface{}, err error) {
	response := make(map[string]interface{})
	code := http.StatusOK
	if err != nil {
		response["error"] = err.Error()
		code = http.StatusBadRequest
	} else {
		response[path] = val
	}

	body, err := json.Marshal(response)
	if err != nil {
		body = []byte(err.Error())
		code = http.StatusInternalServerError
	}
	_, err = rw.Write(body)
	if err != nil {
		code = http.StatusInternalServerError
	}
	rw.WriteHeader(code)
}

func (ds *CloudEyeDatasource) listRegions(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	cfg, err := LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}
	if cfg.Region != "" && cfg.CESEndpoint != "" {
		res := []string{cfg.Region}
		writeResult(rw, "regions", res, nil)
		return
	}
	writeResult(rw, "regions", GetMeta().Regions, nil)
}

func (ds *CloudEyeDatasource) listNamespaces(rw http.ResponseWriter, req *http.Request) {
	log.DefaultLogger.Info("List namespaces", "URL", req.URL.String())
	ctx := req.Context()
	cfg, err := LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}
	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}
	reqRegion := params.Get("region")
	cfg.Region = reqRegion

	if cfg.MetaConfEnabled {
		writeResult(rw, "namespaces", GetMeta().Namespaces[reqRegion], nil)
		return
	}

	cesClient := &CESClient{Client: GetCESClient(cfg), Region: reqRegion}
	res := cesClient.ListNamespaces()
	writeResult(rw, "namespaces", res, nil)
}

func (ds *CloudEyeDatasource) listDims(rw http.ResponseWriter, req *http.Request) {
	log.DefaultLogger.Info("List dimensions", "URL", req.URL.String())
	ctx := req.Context()
	cfg, err := LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}

	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}
	reqRegion := params.Get("region")
	reqNamespace := params.Get("namespace")
	cfg.Region = reqRegion

	if cfg.MetaConfEnabled {
		writeResult(rw, "dimensions", LoadDimensions(reqRegion, reqNamespace), nil)
		return
	}

	cesClient := &CESClient{Client: GetCESClient(cfg), Region: reqRegion}
	res := cesClient.ListDims(reqNamespace)
	writeResult(rw, "dimensions", res, nil)
}

func (ds *CloudEyeDatasource) listMetrics(rw http.ResponseWriter, req *http.Request) {
	log.DefaultLogger.Info("List metrics", "URL", req.URL.String())
	ctx := req.Context()
	cfg, err := LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}

	params, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		writeResult(rw, "", nil, err)
		return
	}
	reqRegion := params.Get("region")
	cfg.Region = reqRegion

	if cfg.MetaConfEnabled {
		writeResult(rw, "metrics", LoadMetrics(params.Get("namespace"), params.Get("dimstr")), nil)
		return
	}
	cesClient := &CESClient{Client: GetCESClient(cfg), Region: reqRegion}
	res := cesClient.ListMetrics(params.Get("namespace"), params.Get("dimstr"))
	writeResult(rw, "metrics", res, nil)
}

type instanceSettings struct {
}

func newCloudEyeInstance(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creating a a new instance to allow plugin authors
	// to cleanup.
}

func LoadSettings(ctx backend.PluginContext) (*CloudEyeSettings, error) {
	var conf commonConf
	setting := ctx.DataSourceInstanceSettings
	err := json.Unmarshal(setting.JSONData, &conf)
	if err != nil {
		return nil, err
	}

	secDataMap := setting.DecryptedSecureJSONData
	config := &CloudEyeSettings{
		CESEndpoint:     conf.CESEndpoint,
		Region:          conf.Region,
		ProjectID:       conf.ProjectID,
		MetaConfEnabled: conf.MetaConfEnabled,
		AK:              secDataMap["accessKey"],
		SK:              secDataMap["secretKey"],
	}

	return config, nil
}
