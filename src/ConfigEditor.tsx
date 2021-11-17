import React, {ChangeEvent, PureComponent} from 'react';
import {InlineField, InlineFieldRow, InlineSwitch, LegacyForms, Tab, TabContent, TabsBar} from '@grafana/ui';
import {DataSourcePluginOptionsEditorProps} from '@grafana/data';
import {MyDataSourceOptions, MySecureJsonData} from './types';
import './css/common.css'

const {SecretFormField, FormField} = LegacyForms;

// @ts-ignore
interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData<unknown>> {
}

interface ConfigEditorState {
  tabs: Array<any>;
}

export class ConfigEditor extends PureComponent<Props, ConfigEditorState> {
  constructor(props: Props) {
    super(props);
    this.state = {
      tabs: [
        {label: "Huaweicloud Mode", active: true},
        {label: "Specific Region Mode", active: false}
      ]
    }
  }

  onCESEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onOptionsChange, options} = this.props;
    const jsonData = {
      ...options.jsonData,
      cesEndpoint: event.target.value,
    };
    onOptionsChange({...options, jsonData});
  };

  onRegionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onOptionsChange, options} = this.props;
    const jsonData = {
      ...options.jsonData,
      region: event.target.value,
    };
    onOptionsChange({...options, jsonData});
  };

  onProjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onOptionsChange, options} = this.props;
    const jsonData = {
      ...options.jsonData,
      projectId: event.target.value,
    };
    onOptionsChange({...options, jsonData});
  };

  onAKChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onOptionsChange, options} = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        accessKey: event.target.value,
      },
    });
  };

  onSKChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onOptionsChange, options} = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        secretKey: event.target.value,
      },
    });
  };

  onResetAK = () => {
    const {onOptionsChange, options} = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        accessKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        accessKey: '',
      },
    });
  };

  onResetSK = () => {
    const {onOptionsChange, options} = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        secretKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        secretKey: '',
      },
    });
  };

  onMetaConfChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onOptionsChange, options} = this.props;
    options.jsonData.metaConfEnabled = !options.jsonData.metaConfEnabled
    const jsonData = {
      ...options.jsonData,
      metaConfEnabled: options.jsonData.metaConfEnabled,
    };
    onOptionsChange({...options, jsonData});
  };

  resetForm() {
    const {onOptionsChange, options} = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        accessKey: false,
        secretKey: false

      },
      secureJsonData: {
        ...options.secureJsonData,
        accessKey: '',
        secretKey: ''
      },
      jsonData: {
        iamEndpoint: '',
        cesEndpoint: '',
        projectId: '',
        region: ''
      }
    });
  }

  onTabChange(index: number) {
    if (index === 0) {
      const {onOptionsChange, options} = this.props;
      const jsonData = {
        ...options.jsonData,
        region: 'cn-east-3',
      };
      onOptionsChange({...options, jsonData});
    }
    this.resetForm()
    const res: Array<object> = [];
    this.state.tabs.forEach((tab: any, idx: number) => {
      const temp = {
        label: tab.label,
        active: idx === index
      }
      res.push(temp)
    })
    this.setState({tabs: res})
  }

  render() {
    const {options} = this.props;
    const {jsonData, secureJsonFields} = options;
    const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;
    return (
      <div>
        <TabsBar>
          {this.state.tabs.map((tab: any, index: number) => {
            return (
              <Tab
                css="margin: 1px"
                key={index}
                label={tab.label}
                active={tab.active}
                onChangeTab={() => this.onTabChange(index)}/>
            )
          })}
        </TabsBar>
        <TabContent>
          {this.state.tabs[0].active &&
          <div className="gf-form-group">
              <div className="form-line-style">
                  <div className="gf-form-inline">
                      <div className="gf-form">
                          <SecretFormField
                              isConfigured={(secureJsonFields && secureJsonFields.accessKey) as boolean}
                              value={secureJsonData.accessKey || ''}
                              label="IAM Access Key"
                              placeholder="access key"
                              labelWidth={10}
                              inputWidth={20}
                              onReset={this.onResetAK}
                              onChange={this.onAKChange}
                          />
                          <SecretFormField
                              isConfigured={(secureJsonFields && secureJsonFields.secretKey) as boolean}
                              value={secureJsonData.secretKey || ''}
                              label="IAM Secret Key"
                              placeholder="secret key"
                              labelWidth={10}
                              inputWidth={20}
                              onReset={this.onResetSK}
                              onChange={this.onSKChange}
                          />
                      </div>
                  </div>
              </div>
          </div>
          }
          {this.state.tabs[1].active &&
          <div className="gf-form-group">
              <div className="form-line-style">
                  <FormField
                      label="CES endpoint"
                      labelWidth={10}
                      inputWidth={20}
                      onChange={this.onCESEndpointChange}
                      value={jsonData.cesEndpoint || ''}
                      placeholder="https://ces.cn-east-3.myhuaweicloud.com"
                  />
                  <FormField
                      label="Region ID"
                      labelWidth={10}
                      inputWidth={20}
                      onChange={this.onRegionChange}
                      value={jsonData.region || ''}
                      placeholder="cn-east-3"
                  />
                  <FormField
                      label="Project ID"
                      labelWidth={10}
                      inputWidth={20}
                      onChange={this.onProjectIdChange}
                      value={jsonData.projectId || ''}
                      placeholder="project id"
                  />
              </div>
              <div className="gf-form-inline">
                  <div className="gf-form">
                      <SecretFormField
                          isConfigured={(secureJsonFields && secureJsonFields.accessKey) as boolean}
                          value={secureJsonData.accessKey || ''}
                          label="IAM Access Key"
                          placeholder="access key"
                          labelWidth={10}
                          inputWidth={20}
                          onReset={this.onResetAK}
                          onChange={this.onAKChange}
                      />
                      <SecretFormField
                          isConfigured={(secureJsonFields && secureJsonFields.secretKey) as boolean}
                          value={secureJsonData.secretKey || ''}
                          label="IAM Secret Key"
                          placeholder="secret key"
                          labelWidth={10}
                          inputWidth={20}
                          onReset={this.onResetSK}
                          onChange={this.onSKChange}
                      />
                  </div>
              </div>
          </div>
          }
        </TabContent>
        <InlineFieldRow>
          <InlineField label="Get Metric Meta From Conf File" tooltip="打开开关后，通过metric.yaml配置获取区域/服务/资源/指标列表">
            <InlineSwitch css="" onChange={this.onMetaConfChange} value={jsonData.metaConfEnabled || false}/>
          </InlineField>
        </InlineFieldRow>
      </div>
    );
  };
};
