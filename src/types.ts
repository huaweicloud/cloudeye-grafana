import {DataQuery, DataSourceJsonData} from '@grafana/data';

export interface MyQuery extends DataQuery {
  region?: string;
  namespace?: string;
  dimstr?: string;
  metricName?: string;
  filter?: string;
  period?: string;
  from?: number;
  to?: number;
}

export const defaultQuery: Partial<MyQuery> = {
  region: "cn-east-3",
  namespace: "SYS.ECS",
  dimstr: "instance_id,1b674d59-0a56-4bf5-ad77-c5f8c63e9324",
  metricName: "cpu_util",
  filter: "average",
  period: "1",
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  iamEndpoint?: string;
  cesEndpoint?: string;
  region?: string;
  projectId?: string;
  metaConfEnabled?: boolean;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  accessKey?: string;
  secretKey?: string;
}
