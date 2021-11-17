package plugin

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var NsCache = NamespaceCache{MetaCache{
	Data: sync.Map{},
}}

var DmCache = DimensionCache{MetaCache{
	Data: sync.Map{},
}}

var MCache = MetricCache{MetaCache{
	Data: sync.Map{},
}}

type MetaCache struct {
	Data sync.Map // key: string, value: CachedMeta
}

type CachedMeta struct {
	Name       string
	Meta       []string
	Finished   bool // 是否查完
	Marker     string
	TTL        int64
	UpdateTime int64
}

func (cache *CachedMeta) isExpired() bool {
	return cache.UpdateTime+cache.TTL < getTimestamp()
}

func (c *MetaCache) getCachedMeta(key string) *CachedMeta {
	if v, ok := c.Data.Load(key); ok {
		if cache, ok := v.(*CachedMeta); ok {
			return cache
		}
	}
	return nil
}

func resetCache() {
	NsCache.Data = sync.Map{}
	DmCache.Data = sync.Map{}
	MCache.Data = sync.Map{}
	runtime.GC()
}

type MetaUtil interface {
	newCachedMeta() *CachedMeta
	buildKey(*QueryParam) string
	getCache() *MetaCache
	getRespElem(model.MetricInfoList) string
	buildQuery(*QueryParam) *model.ListMetricsRequest
	setDefaultMeta(*QueryParam, *CachedMeta)
}

type NamespaceCache struct {
	MetaCache
}

func (c *NamespaceCache) newCachedMeta() *CachedMeta {
	return &CachedMeta{TTL: 30 * 60 * 1000}
}

func (c *NamespaceCache) getCache() *MetaCache {
	return &(NsCache.MetaCache)
}

func (c *NamespaceCache) buildKey(params *QueryParam) string {
	return params.Region
}

func (c *NamespaceCache) getRespElem(metric model.MetricInfoList) string {
	return metric.Namespace
}

func (c *NamespaceCache) buildQuery(param *QueryParam) *model.ListMetricsRequest {
	return &model.ListMetricsRequest{}
}

func (c *NamespaceCache) setDefaultMeta(param *QueryParam, cache *CachedMeta) {
	cache.Meta = GetMeta().Namespaces[param.Region]
}

type DimensionCache struct {
	MetaCache
}

func (c *DimensionCache) newCachedMeta() *CachedMeta {
	return &CachedMeta{TTL: 10 * 60 * 1000}
}

func (c *DimensionCache) getCache() *MetaCache {
	return &(DmCache.MetaCache)
}

func (c *DimensionCache) buildKey(param *QueryParam) string {
	return fmt.Sprintf("%s|%s", param.Region, param.Namespace)
}

func (c *DimensionCache) getRespElem(metric model.MetricInfoList) string {
	return getDimStr(metric.Dimensions)
}

func (c *DimensionCache) buildQuery(param *QueryParam) *model.ListMetricsRequest {
	return &model.ListMetricsRequest{
		Namespace: &param.Namespace,
	}
}

func (c *DimensionCache) setDefaultMeta(param *QueryParam, cache *CachedMeta) {
	return
}

type MetricCache struct {
	MetaCache
}

func (c *MetricCache) newCachedMeta() *CachedMeta {
	return &CachedMeta{TTL: 5 * 60 * 1000}
}

func (c *MetricCache) getCache() *MetaCache {
	return &(MCache.MetaCache)
}

func (c *MetricCache) buildKey(param *QueryParam) string {
	return fmt.Sprintf("%s|%s|%s", param.Region, param.Namespace, param.DimStr)
}

func (c *MetricCache) getRespElem(metric model.MetricInfoList) string {
	return metric.MetricName
}

func (c *MetricCache) buildQuery(param *QueryParam) *model.ListMetricsRequest {
	dims := strings.Split(param.DimStr, ".")
	reqParam := &model.ListMetricsRequest{
		Namespace: &param.Namespace,
	}
	switch len(dims) {
	case 3:
		reqParam.Dim2 = &dims[2]
		fallthrough
	case 2:
		reqParam.Dim1 = &dims[1]
		fallthrough
	case 1:
		reqParam.Dim0 = &dims[0]
	}
	return reqParam
}

func (c *MetricCache) setDefaultMeta(param *QueryParam, cache *CachedMeta) {
	return
}
