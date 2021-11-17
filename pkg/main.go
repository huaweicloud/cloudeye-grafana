package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/huaweicloud/cloudeye-grafana/pkg/plugin"
)

func main() {
	if err := datasource.Serve(plugin.NewCloudEye()); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
