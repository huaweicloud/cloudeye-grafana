package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"gopkg.in/yaml.v2"
)

type MetaConf struct {
	Regions    []string                       `yaml:"regions"`
	Namespaces map[string][]string            `yaml:"namespaces"` // key: region, value: namespaceList
	Dimensions map[string]map[string][]string `yaml:"dimensions"` // key: region|namespace, value: map[dimKey]dimValues
	Metrics    map[string][]string            `yaml:"metrics"`    // key: namespace|dimKey, value: metrics
}

var GetMeta = initMetaConf()

func initMetaConf() func() *MetaConf {
	var metaConf MetaConf
	var once sync.Once
	return func() *MetaConf {
		once.Do(func() {
			ex, err := os.Executable()
			if err != nil {
				log.DefaultLogger.Error("Get executable file path error", "err", err)
				return
			}
			path := filepath.Dir(ex)
			metaConfPath := filepath.Join(path, "./metric.yaml")

			c, err := loadConf(metaConfPath)
			if err != nil {
				log.DefaultLogger.Error("Load meta conf error", "err", err)
				return
			}
			metaConf = *c
		})
		return &metaConf
	}
}

func loadConf(fPath string) (*MetaConf, error) {
	var conf MetaConf
	if err := loadConfFromFile(fPath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func loadConfFromFile(fPath string, c interface{}) error {
	bs, err := ioutil.ReadFile(fPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, c)
}

func LoadDimensions(region, namespace string) []string {
	dims := GetMeta().Dimensions[fmt.Sprintf("%s|%s", region, namespace)]
	var res []string
	for k, v := range dims {
		dimKeys := strings.Split(k, ",")
		for i := range v {
			dimValues := strings.Split(v[i], ",")
			if len(dimKeys) != len(dimValues) {
				continue
			}
			dimStrs := make([]string, 0, len(dimKeys))
			for j := range dimKeys {
				dimStrs = append(dimStrs, fmt.Sprintf("%s,%s", dimKeys[j], dimValues[j]))
			}

			res = append(res, strings.Join(dimStrs, "."))
		}
	}
	return res
}

func LoadMetrics(namespace, dimStr string) []string {
	dims := strings.Split(dimStr, ".")
	dimKeys := make([]string, 0, len(dims))
	for _, dim := range dims {
		eachDim := strings.Split(dim, ",")
		if len(eachDim) != 2 {
			continue
		}
		dimKeys = append(dimKeys, eachDim[0])
	}

	return GetMeta().Metrics[fmt.Sprintf("%s|%s", namespace, strings.Join(dimKeys, ","))]
}
