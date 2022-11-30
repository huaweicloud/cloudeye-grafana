import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  FieldType,
  MutableDataFrame,
  SelectableValue
} from '@grafana/data';
import {DataSourceWithBackend, getTemplateSrv} from '@grafana/runtime';
import {MyDataSourceOptions, MyQuery} from './types';


export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  metaConfEnabled: boolean;

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.metaConfEnabled = instanceSettings.jsonData.metaConfEnabled || false;
  }

  // @ts-ignore
  async query(options: DataQueryRequest<MyQuery>): Promise<DataQueryResponse> {
    if (!this.variableIsExist('filter') || !this.variableIsExist('period')) {
      const promises = this.listMetricDataByCustom(options);
      return Promise.all(promises).then((data: any) => ({data}))
    }
    const promise = this.listMetricDataByTemplate(options);
    // @ts-ignore
    return promise.then((data: any) => ({data}));
  }

  // 不使用模板，使用自定义dashboard的场景查询监控数据
  listMetricDataByCustom(options: any) {
    const promises = options.targets.map((target: any) => {
      const metrics: Array<any> = [];
      const metric: any = {};
      const refIDs: Array<any> = [];
      metric.namespace = target.namespace;
      metric.dimensions = this.parseDims(target.dimstr);
      metric.metric_name = target.metricName;
      refIDs.push(target.refId);
      metrics.push(metric);
      const reqBody: any = {
        metrics,
        refIDs,
        from: options.range.from.valueOf(),
        to: options.range.to.valueOf(),
        filter: target.filter || 'average',
        period: target.period || '1',
        region: target.region || 'cn-east-3'
      }
      // @ts-ignore
      return this.listMetricData(reqBody).then(response => {
        if (response && response.data && response.data.results) {
          for (let ref in response.data.results) {
            if (Object.prototype.hasOwnProperty.call(response.data.results, ref)) {
              const label: any = {
                namespace: target.namespace
              }
              metric.dimensions.forEach((dim: any) => {
                label[dim.name] = dim.value;
              });
              const frame = new MutableDataFrame({
                refId: ref,
                fields: [
                  {name: "Time", type: FieldType.time},
                  {name: metric.metric_name, type: FieldType.number, labels: label},
                ],
              });

              const times = response.data.results[ref].frames[0].data.values[0];
              const values = response.data.results[ref].frames[0].data.values[1];
              times.forEach((time: any, index: any) => {
                frame.appendRow([time, values[index]]);
              });
              return frame;
            }
          }
        }
      })
    });
    return promises;
  }

  // 使用模板生成dashboard场景查询监控数据
  listMetricDataByTemplate(options: any) {
    const metrics: Array<any> = [];
    const refIDs: Array<any> = [];
    const queriesMap: any = {};
    // @ts-ignore
    options.targets.forEach(target => {
      if (!target.region || !target.namespace || !target.dimstr || !target.metricName) {
        return [];
      }
      const metric: any = {};
      metric.namespace = target.namespace;
      metric.dimensions = this.handleDimStr(target);
      metric.metric_name = target.metricName;
      metrics.push(metric);
      refIDs.push(target.refId);
      queriesMap[target.refId] = metric;
    });
    const reqBody = {
      metrics,
      refIDs,
      from: options.range.from.valueOf(),
      to: options.range.to.valueOf(),
      filter: this.getVarValue('filter', 'average'),
      period: this.getVarValue('period', '1'),
      region: this.getVarValue('region', 'cn-east-3')
    };
    if (metrics.length === 0 || refIDs.length === 0) {
      return new Promise(resolve => resolve).then(() => {
        return {data: []}
      })
    }
    const promise = this.listMetricData(reqBody).then(response => {
      const frames: Array<any> = [];
      if (response && response.data && response.data.results) {
        for (let ref in response.data.results) {
          if (Object.prototype.hasOwnProperty.call(response.data.results, ref)) {
            const label: any = {
              namespace: queriesMap[ref].namespace
            }
            queriesMap[ref].dimensions.forEach((dim: any) => {
              label[dim.name] = dim.value;
            });
            const frame = new MutableDataFrame({
              refId: ref,
              fields: [
                {name: "Time", type: FieldType.time},
                {name: queriesMap[ref].metric_name, type: FieldType.number, labels: label},
              ],
            });

            const times = response.data.results[ref].frames[0].data.values[0];
            const values = response.data.results[ref].frames[0].data.values[1];
            times.forEach((time: any, index: any) => {
              frame.appendRow([time, values[index]]);
            });
            frames.push(frame);
          }
        }
      }
      return frames;
    });
    return promise;
  }

  // @ts-ignore
  async metricFindQuery(query: any, options?: any) {
    const templateVariables = getTemplateSrv().getVariables();
    if (query == 'listRegions()') {
      return await this.listRegions();
    }

    if (query == 'listFilterOptions()') {
      return this.listFilterOptions();
    }

    if (query == 'listPeriodOptions()') {
      return this.listPeriodOptions();
    }

    if (query.indexOf('listDims') >= 0) {
      const listDimsParams = this.getTemplateVars(query);
      const regionVar = listDimsParams[0] ? listDimsParams[0] : '';
      const namespace = listDimsParams[1] ? listDimsParams[1] : '';
      let dimsName = listDimsParams[2] ? listDimsParams[2] : '';
      const tagDimName = listDimsParams[3] ? listDimsParams[3] : '';
      let region = '';
      templateVariables.forEach((item: any) => {
        if (regionVar.indexOf("$" + item.name) >= 0) {
          region = item.current.value;
        }
        if (dimsName.indexOf("$" + item.name) >= 0) {
          dimsName = dimsName.replace("$" + item.name, item.current.value)
        }
      });
      return await this.listDims(region, namespace, dimsName, tagDimName);
    }
  }

  listFilterOptions(): any[] {
    return [
      {text: '平均值', label: '平均值', value: 'average'},
      {text: '最小值', label: '最小值', value: 'min'},
      {text: '最大值', label: '最大值', value: 'max'},
      {text: '求和值', label: '求和值', value: 'sum'},
    ];
  }

  listPeriodOptions(): any[] {
    return [
      {text: '原始粒度', label: '原始粒度', value: '1'},
      {text: '5min粒度', label: '5min粒度', value: '300'},
      {text: '1h粒度', label: '1h粒度', value: '3600'},
    ];
  }

  async listRegions(): Promise<Array<SelectableValue<string>>> {
    return this.getResource('regions').then(({regions}) => {
      return regions ? regions.map((item: string) => ({text: item, label: item, value: item})) : [];
    });
  }

  async listNamespaces(region: string | undefined): Promise<Array<SelectableValue<string>>> {
    return this.getResource('namespaces', {region: region}).then(({namespaces}) => {
      return namespaces ? namespaces.map((item: string) => ({text: item, label: item, value: item})) : [];
    });
  }

  async listDims(region: string | undefined, namespace: string | undefined, dimsName: string, tagDimName: string): Promise<Array<SelectableValue<string>>> {
    return this.getResource('dimensions', {region: region, namespace: namespace}).then(({dimensions}) => {
      const dims = Object.values(dimensions);
      const result: Array<SelectableValue<string>> = [];
      if (dimsName === '') {
        dims.forEach((item: any) => {
          result.push({label: item, value: item});
        });
        return result;
      }
      const newDimsName = dimsName.replace(':', ',')
      const tagDimsName = this.getOrderedDimNames(newDimsName);
      const preDim = this.parseDims(newDimsName)

      dims.forEach((item: any) => {
        const itemDimsName = this.getOrderedDimNames(item);
        let valid = true;
        preDim.forEach((dim: any) => {
          if (item.indexOf(dim.name + "," + dim.value) === -1) {
            valid = false
          }
        })
        if (tagDimsName === itemDimsName && valid) {
          const dimension = this.parseDims(item)
          let value = '';
          dimension.forEach((dim: any) => {
            if (dim.name === tagDimName) {
              value = dim.value;
              return;
            }
          });
          result.push({text: value, label: item, value: item});
        }
      })
      return result;
    });
  }

  getOrderedDimNames(dimsStr: string | null): string {
    if (dimsStr) {
      const dimNames: Array<Object> = [];
      const dims = dimsStr.split('.');
      dims.forEach((item: any) => {
        const temp = item.split(',');
        if (temp.length === 0) {
          return;
        }
        dimNames.push(temp[0])
      });
      return dimNames.sort().join('.');
    }
    return "";
  }


  async listMetrics(region: string | undefined, namespace: string | undefined, dimstr: string | undefined): Promise<Array<SelectableValue<string>>> {
    return this.getResource('metrics', {region: region, namespace: namespace, dimstr: dimstr}).then(({metrics}) => {
      return metrics ? metrics.map((item: string) => ({label: item, value: item})) : [];
    });
  }

  async listMetricData(reqBody: any): Promise<any> {
    return this.postResource('metric-data', reqBody);
  }

  getTemplateVars(queryStr: string): Array<string> {
    const leftIndex: number = queryStr.indexOf('(');
    const rightIndex: number = queryStr.indexOf(')');
    const params = queryStr.substring(leftIndex + 1, rightIndex);
    return params ? params.split(',') : [];
  }

  // dimstr:string to dimsions:Array
  parseDims(dimStr: string): Array<any> {
    if (dimStr) {
      const dimsions: Array<Object> = [];
      const dims = dimStr.split('.');
      dims.forEach((item: any) => {
        const temp = item.split(',');
        if (temp.length == 2) {
          dimsions.push({name: temp[0], value: temp[1]})
        }
      });
      return dimsions;
    }
    return [];
  }

  handleDimStr(target: any): any[] {
    const dimsionsInTarget: any = this.parseDims(target.dimstr);
    if (dimsionsInTarget.length > 0) {
      const queries = getTemplateSrv().getVariables();
      let targetQuery;
      queries.forEach((item: any) => {
            const queryDims = this.getOrderedDimNames(item.current.value)
            const targetDims = this.getOrderedDimNames(target.dimstr)
            if (queryDims === targetDims) {
              targetQuery = item;
            }
          }
      );
      // @ts-ignore
      return targetQuery && targetQuery.current.value ? this.parseDims(targetQuery.current.value) : dimsionsInTarget;
    }
    return dimsionsInTarget || [];
  }

  getVarValue(key: string, defaultValue: string): string {
    const queries = getTemplateSrv().getVariables();
    const targetQuery: any = queries.find(item => item.name === key);
    return targetQuery && targetQuery.current.value ? targetQuery.current.value : defaultValue;
  };

  variableIsExist(key: string): boolean {
    const variables = getTemplateSrv().getVariables();
    const targetVar: any = variables.find(item => item.name === key);
    return targetVar && targetVar.name ? true : false;
  }
}