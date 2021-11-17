# cloudeye-grafana
[![LICENSE](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/huaweicloud/cloudeye-grafana/blob/master/LICENSE)

cloudeye-grafana是[华为云监控服务](https://support.huaweicloud.com/ces/)为适配Grafana开发的datasource插件，通过[华为云SDK](https://github.com/huaweicloud/huaweicloud-sdk-go-v3)获取监控数据。

# 快速入门

## 1. 安装
> 安装前准备:  
> a. 已安装Grafana版本 >=7, [grafana官方下载地址](https://grafana.com/grafana/download)  
> b. 从[release页面](https://github.com/huaweicloud/cloudeye-grafana/releases)下载cloudeye-grafana-{version}.tar.gz

### 1.1 从release安装
a. 将下载的插件包放到grafana的plugin目录(见conf/defaults.ini中的plugins配置路径), 解压缩cloudeye-grafana-{version}.tar.gz, 需要注意目录权限和grafana运行权限保持一致
  
b. 修改 conf/defaults.ini 允许未签名插件运行    
> allow_loading_unsigned_plugins = huawei-cloudeye-grafana  
   
c. 重启grafana

## 2. 配置cloudeye-grafana数据源
> 配置前准备:  
> a. [获取AK/SK](https://support.huaweicloud.com/devg-apisign/api-sign-provide-aksk.html)  
> b. (可选)Specific Region Mode模式下需要[获取project_id](https://support.huaweicloud.com/devg-apisign/api-sign-provide-proid.html)  
> c. (可选)Specific Region Mode模式下需要[获取CES Endpoint和RegionID](https://developer.huaweicloud.com/endpoint)

### 2.1 配置数据源
a. 进入grafana的数据源配置页面(Data Sources),点击Add data source进入配置表单页面,填入数据源名称cloudeye-grafana,
在数据源列表中选择cloudeye-grafana。

b. 当前支持两种模式，可按需选择配置。
> Huaweicloud Mode（华为云多region模式）：配置IAM Access Key、IAM Secret Key

> Specific Region Mode（单region模式）：配置CES Endpoint、Region ID、Project ID、IAM Access Key、IAM Secret Key

c. (可选)如果需要开启通过配置文件读取指标元数据，需要点击Get Metric Meta From Conf按钮开启，并按下文配置指标元数据列表。

d. 点击Save & test按钮，如果显示Data source is working，说明数据源配置成功，可以开始在grafana中访问华为云监控的数据了。

## 3. (可选)配置指标元数据列表
为了提升查询体验，对于资源列表变化实时性不高、资源量大的租户，可以提前将资源列表配置在dist/metric.yaml文件中,区域/服务/资源/指标列表以配置文件为准。  
> a. [云监控支持的服务指标列表](https://support.huaweicloud.com/usermanual-ces/zh-cn_topic_0202622212.html)     
> b. [华为云支持region列表](https://developer.huaweicloud.com/endpoint)  
> c. 按metric.yaml样例配置完成后，重启grafana
    
## 4. 导入dashboard模板
为简便租户配置，本插件提供了ECS、ELB、RDS服务的Dashboard预设模板，见： cloudeye-grafana/src/templates目录
