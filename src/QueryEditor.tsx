import {defaults} from 'lodash';

import React, {PureComponent} from 'react';
import {InlineFormLabel, SegmentAsync, Select} from '@grafana/ui';
import {QueryEditorProps} from '@grafana/data';
import {DataSource} from './datasource';
import {defaultQuery, MyDataSourceOptions, MyQuery} from './types';
import './css/common.css'

// @ts-ignore
type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  componentDidMount() {
    const {onRunQuery} = this.props;
    onRunQuery();
  }

  onFilterChange = (item: any) => {
    const {onChange, onRunQuery, query} = this.props;
    onChange({...query, filter: item.value});
    onRunQuery();
  };

  onPeriodChange = (item: any) => {
    const {onChange, onRunQuery, query} = this.props;
    onChange({...query, period: item.value});
    onRunQuery();
  };

  onRegionChange = (item: any) => {
    const {onChange, query} = this.props;
    onChange({...query, region: item.value});
  };

  onNamespaceChange = (item: any) => {
    const {onChange, query} = this.props;
    onChange({...query, namespace: item.value});
  };

  onDimstrChange = (item: any) => {
    const {query, onChange, onRunQuery} = this.props;
    onChange({...query, dimstr: item.value});
    onRunQuery();
  };

  onMetricNameChange = (item: any) => {
    const {query, onRunQuery, onChange} = this.props;
    onChange({...query, metricName: item.value});
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const datasource = this.props.datasource;

    return (
      <div className="gf-form">
        <div className="gf-form-inline">
          <InlineFormLabel width={5} tooltip={<p>Select Region</p>}>
            Region
          </InlineFormLabel>
          <SegmentAsync
            loadOptions={() => datasource.listRegions()}
            placeholder="region"
            value={query.region}
            allowCustomValue={false}
            onChange={this.onRegionChange}
          />

          <InlineFormLabel width={5} tooltip={<p>Select namespace</p>}>
            Namespace
          </InlineFormLabel>
          <SegmentAsync
            loadOptions={() => datasource.listNamespaces(query.region)}
            placeholder="namespace"
            value={query.namespace}
            allowCustomValue={false}
            onChange={this.onNamespaceChange}
          />

          <InlineFormLabel width={5} tooltip={<p>Select dims</p>}>
            dimstr
          </InlineFormLabel>
          <SegmentAsync
            loadOptions={() => datasource.listDims(query.region, query.namespace, '', '')}
            placeholder="dimstr"
            value={query.dimstr}
            allowCustomValue={false}
            onChange={this.onDimstrChange}
          />

          <InlineFormLabel width={5} tooltip={<p>Select metrics</p>}>
            metrics
          </InlineFormLabel>
          <SegmentAsync
            loadOptions={() => datasource.listMetrics(query.region, query.namespace, query.dimstr)}
            placeholder="metrics"
            value={query.metricName}
            allowCustomValue={false}
            onChange={this.onMetricNameChange}
          />

          <InlineFormLabel width={5} tooltip={<p>filter</p>}>
            filter
          </InlineFormLabel>
          <Select
            width={15}
            options={datasource.listFilterOptions()}
            placeholder="filter"
            value={query.filter}
            allowCustomValue={false}
            onChange={this.onFilterChange}
          />

          <InlineFormLabel width={5} tooltip={<p>period</p>}>
            period
          </InlineFormLabel>
          <Select
            width={15}
            options={datasource.listPeriodOptions()}
            placeholder="period"
            value={query.period}
            allowCustomValue={false}
            onChange={this.onPeriodChange}
          />
        </div>
      </div>
    );
  }
}
