module github.com/huaweicloud/cloudeye-grafana

go 1.19

require (
	github.com/agiledragon/gomonkey/v2 v2.11.0
	github.com/grafana/grafana-plugin-sdk-go v0.166.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.92
	github.com/stretchr/testify v1.8.4
	gopkg.in/yaml.v3 v3.0.1
)

replace (
	github.com/apache/thrift => github.com/apache/thrift v0.16.0
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.9.3
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.14.0
	go.etcd.io/etcd => go.etcd.io/etcd v3.4.15+incompatible
)
