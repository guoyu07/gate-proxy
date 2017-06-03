import React, { Component } from 'react'
import PropTypes from 'prop-types'
import {
  Modal,
  Checkbox,
  Switch,
  Button,
  Form,
  Input,
  Row,
  Col,
  Select,
  Steps } from 'antd'

const FormItem = Form.Item
const Step = Steps.Step
const Option = Select.Option
const CheckboxGroup = Checkbox.Group
const formItemLayout = {
  labelCol: {
    xs: { span: 24 },
    sm: { span: 6 }
  },
  wrapperCol: {
    xs: { span: 24 },
    sm: { span: 12 }
  }
}
const formNodeItemLayout = {
  labelCol: {
    sm: { span: 4 }
  },
  wrapperCol: {
    sm: { span: 20 }
  }
}
const formNodeParamItemLayout = {
  labelCol: {
    xs: { span: 10 }
  },
  wrapperCol: {
    xs: { span: 14 }
  }
}
const FormNodeParamItemLabelLayout = {
  labelCol: {
    xs: { span: 8 }
  },
  wrapperCol: {
    xs: { span: 16 }
  }
}
const paramInfo = {
  attr: '',
  required: false,
  from: 1,
  to: 1,
  toName: '',
  validation: ''
}
const nodeInfo = {
  cluster: '',
  attr: '',
  rewrite: '',
  paramGroup: [
    paramInfo
  ]
}

export default class ApiModalForm extends Component {
  constructor (props, context) {
    super(props, context)
    this.state = {
      currentStep: 0,
      name: '',
      method: '',
      url: '',
      domain: '',
      handlers: [],
      nodeGroup: []
    }
  }
  /**
   * @description 处理插件数据
   */
  handlePluginOptions = (plugins) => {
    let pluginsOptions = []
    for (let i = 0; i < plugins.length; i++) {
      pluginsOptions.push({
        label: plugins[i].name + '-' + plugins[i].version,
        value: plugins[i].name,
        disabled: plugins[i].private
      })
    }
    return pluginsOptions
  }

  renderBaseForm = () => {
    const { form } = this.props
    return (
      <div>
        <FormItem
          {...formItemLayout}
          label='Name'>
          {form.getFieldDecorator('name', {
            rules: [{ required: true }]
          })(<Input placeholder='e.g login api' onChange={(e) => this.setState({ name: e.target.value })} />)}
        </FormItem>
        <FormItem
          {...formItemLayout}
          label='Method'>
          {form.getFieldDecorator('method', {
            rules: [{ required: true }]
          })(
            <Select placeholder='e.g GET' onSelect={(method) => this.setState({ method })}>
              <Option key='GET' value='GET'>GET</Option>
              <Option key='POST' value='POST'>POST</Option>
              <Option key='PUT' value='PUT'>PUT</Option>
              <Option key='DELETE' value='DELETE'>DELETE</Option>
            </Select>
            )}
        </FormItem>
        <FormItem
          {...formItemLayout}
          label='URL'>
          {form.getFieldDecorator('url', {
            rules: [{ required: true }]
          })(<Input placeholder='e.g /user/v0.1/login' onChange={(e) => this.setState({ url: e.target.value })} />)}
        </FormItem>
        <FormItem
          {...formItemLayout}
          label='Domain'>
          {form.getFieldDecorator('domain')(
            <Input
              placeholder='e.g www.goodsogood.com'
              onChange={(e) => this.setState({ domain: e.target.value })}
          />
        )}
        </FormItem>
      </div>
    )
  }
  renderPluginForm = () => {
    const { form, plugins } = this.props
    return (
      <FormItem
        {...formItemLayout}
        label='Plugin'
      >
        {form.getFieldDecorator('handlers')(
          <CheckboxGroup
            options={this.handlePluginOptions(plugins.items)}
            onChange={(handlers) => this.setState({ handlers })} />
      )}
      </FormItem>
    )
  }
  setNodeVal = (index, key, value) => {
    const nodeGroup = this.state.nodeGroup
    nodeGroup[index][key] = value
    this.setState({ nodeGroup })
  }
  setParamVal = (nodeIndex, paramIndex, key, val) => {
    const nodeGroup = this.state.nodeGroup
    nodeGroup[nodeIndex].paramGroup[paramIndex][key] = val
    this.setState({ nodeGroup })
  }
  addNode = () => {
    const nodeGroup = this.state.nodeGroup
    nodeGroup.push(nodeInfo)
    this.setState({ nodeGroup })
  }
  removeNode = (index) => {
    const nodeGroup = this.state.nodeGroup
    nodeGroup.splice(index, 1)
    this.setState({ nodeGroup })
  }
  addParamWithNode = (index) => {
    const nodeGroup = this.state.nodeGroup
    nodeGroup[index].paramGroup.push(paramInfo)
    this.setState({ nodeGroup })
  }
  removeParamWithNode = (nodeIndex, index) => {
    const nodeGroup = this.state.nodeGroup
    nodeGroup[nodeIndex].paramGroup.splice(index, 1)
    this.setState({ nodeGroup })
  }
  generatorValidFields = () => {
    let fields = []
    const nodeGroup = this.state.nodeGroup
    for (let i = 0; i < nodeGroup.length; i++) {
      fields = fields.concat([`nodeGroup[${i}].cluster`, `nodeGroup[${i}].rewrite`])
      for (let pi = 0; pi < nodeGroup[i].paramGroup.length; pi++) {
        fields = fields.concat([
          `nodeGroup[${i}].paramGroup[${pi}].attr`,
          `nodeGroup[${i}].paramGroup[${pi}].from`,
          `nodeGroup[${i}].paramGroup[${pi}].to`
        ])
      }
    }
    return fields
  }
  renderNodeForm = () => {
    const { form, clusters } = this.props
    const nodeGroup = this.state.nodeGroup
    return (
      nodeGroup.map((item, index) => {
        return (
          <Row key={`node-${index}`} >
            <Col span={4}><strong>{`节点[${index}]`}</strong></Col>
            <Col span={17}>
              <FormItem
                {...formNodeItemLayout}
                label='Cluster'
                colon={false}
                >
                {form.getFieldDecorator(`nodeGroup[${index}].cluster`, {
                  rules: [{ required: true, message: 'cluster is required' }]
                })(
                  <Select placeholder='e.g cluster.' onSelect={(value) => this.setNodeVal(index, 'cluster', value)}>
                    {clusters.items.map(item => <Option
                      key={item.clusterName}
                      value={item.clusterName}
                      disabled={item.exist}>{item.clusterName}</Option>)}
                  </Select>
                  )}
              </FormItem>
              <FormItem
                {...formNodeItemLayout}
                label='Attr'
                colon={false}>
                {form.getFieldDecorator(`nodeGroup[${index}].attr`)(
                  <Input placeholder='e.g userInfo' onChange={(e) => this.setNodeVal(index, 'attr', e.target.value)} />
                  )}
              </FormItem>
              <FormItem
                {...formNodeItemLayout}
                label='Rewrite'
                colon={false}>
                {form.getFieldDecorator(`nodeGroup[${index}].rewrite`, {
                  rules: [{ required: true, message: 'rewrite is required' }]
                })(<Input placeholder='e.g /u' onChange={(e) => this.setNodeVal(index, 'rewrite', e.target.value)} />)}
              </FormItem>
            </Col>
            <Col span={2} offset={1}>
              { index === 0
              ? <Button type='primary' shape='circle' icon='plus' onClick={() => this.addNode()} />
              : <Button type='primary' shape='circle' icon='minus' onClick={() => this.removeNode(index)} />}
            </Col>
            {nodeGroup[index].paramGroup && nodeGroup[index].paramGroup.map((param, pi) => {
              return (
                <Col span={24}>
                  <Row key={`node-${index}-${pi}`} >
                    <Col span={2} offset={5}>{`参数: ${pi}`}</Col>
                    <Col span={14} className='param-container'>
                      <Row key={`node-${index}-${pi}-detail`} >
                        <Col span={12}>
                          <FormItem {...FormNodeParamItemLabelLayout} label={`Attr`} colon={false} >
                            {form.getFieldDecorator(`nodeGroup[${index}].paramGroup[${pi}].attr`, {
                              rules: [{ required: true, message: 'attr is required' }]
                            })(
                              <Input placeholder='e.g param attr'
                                onChange={(e) => this.setParamVal(index, pi, 'attr', e.target.value)}
                              />)}
                          </FormItem>
                        </Col>
                        <Col span={12}>
                          <FormItem {...formNodeParamItemLayout} label={`Required`} colon={false}>
                            {form.getFieldDecorator(`nodeGroup[${index}].paramGroup[${pi}].required`)(
                              <Switch
                                defaultChecked={nodeGroup[index].paramGroup[pi].required}
                                onChange={(state) => this.setParamVal(index, pi, 'required', state)}
                              />)}
                          </FormItem>
                        </Col>
                        <Col span={12}>
                          <FormItem {...FormNodeParamItemLabelLayout} label={`From`} colon={false}>
                            {form.getFieldDecorator(`nodeGroup[${index}].paramGroup[${pi}].from`, {
                              rules: [{ required: true, message: 'from is required' }]
                            })(
                              <Select placeholder='e.g Param From .'
                                onChange={(value) => this.setParamVal(index, pi, 'from', value)}>
                                <Option value={1}>Header</Option>
                                <Option value={2}>Query</Option>
                                <Option value={3}>Body</Option>
                              </Select>)}
                          </FormItem>
                        </Col>
                        <Col span={12}>
                          <FormItem {...FormNodeParamItemLabelLayout} label={`To`} colon={false}>
                            {form.getFieldDecorator(`nodeGroup[${index}].paramGroup[${pi}].to`, {
                              rules: [{ required: true, message: 'to is required' }]
                            })(
                              <Select placeholder='e.g Param To .'
                                onChange={(value) => this.setParamVal(index, pi, 'to', value)}>
                                <Option value={1}>Header</Option>
                                <Option value={2}>Query</Option>
                                <Option value={3}>Body</Option>
                              </Select>)}
                          </FormItem>
                        </Col>
                        <Col span={12}>
                          <FormItem {...FormNodeParamItemLabelLayout} label={`ToAttr`} colon={false}>
                            {form.getFieldDecorator(`nodeGroup[${index}].paramGroup[${pi}].toName`)(
                              <Input placeholder='e.g param userId'
                                onChange={(e) => this.setParamVal(index, pi, 'toName', e.target.value)} />)}
                          </FormItem>
                        </Col>
                        <Col span={12}>
                          <FormItem {...FormNodeParamItemLabelLayout} label={`Rule`} colon={false}>
                            {form.getFieldDecorator(`nodeGroup[${index}].paramGroup[${pi}].validation`)(
                              <Input placeholder='e.g param /d+'
                                onChange={(e) => this.setParamVal(index, pi, 'validation', e.target.value)} />)}
                          </FormItem>
                        </Col>
                      </Row>
                    </Col>
                    <Col span={2} offset={1}>
                      {pi === 0
                      ? <Button shape='circle' icon='plus' onClick={() => this.addParamWithNode(index)} />
                      : <Button shape='circle' icon='minus' onClick={() => this.removeParamWithNode(index, pi)} />}
                    </Col>
                  </Row>
                </Col>
              )
            })}
          </Row>
        )
      })
    )
  }
  /**
   * @description 渲染底部
   */
  renderFooter = () => {
    const { form, defaultData } = this.props
    switch (this.state.currentStep) {
      case 0:
        return (
          <Button key='nextStep' type='primary' size='large' onClick={() => {
            form.validateFields(['name', 'method', 'url'], (err, values) => {
              if (err) {
                return
              }
              let handlers = []
              if (this.state.handlers.length) {
                handlers = this.state.handlers
              } else if (defaultData && defaultData.handlers) {
                handlers = defaultData.handlers
              }
              const timeout = setTimeout(() => {
                form.setFieldsValue({
                  handlers
                })
                clearTimeout(timeout)
              }, 500)
              this.setState({
                ...values,
                handlers,
                currentStep: 1
              })
            })
          }
          }>
            下一步
          </Button>
        )
      case 1:
        return (
          <div>
            <Button key='previous' size='large' onClick={() => {
              this.setState({ currentStep: 0 })
              const timeout = setTimeout(() => {
                form.setFieldsValue({
                  name: this.state.name,
                  method: this.state.method,
                  url: this.state.url,
                  domain: this.state.domain
                })
                clearTimeout(timeout)
              }, 500)
            }}>
              上一步
            </Button>
            <Button
              key='nextStep'
              type='primary'
              size='large'
              onClick={() => {
                let nodeGroup = []
                if (this.state.nodeGroup.length) {
                  nodeGroup = this.state.nodeGroup
                } else if (defaultData && defaultData.nodeGroup) {
                  nodeGroup = defaultData.nodeGroup
                } else {
                  nodeGroup[0] = nodeInfo
                }
                this.setState({ nodeGroup, currentStep: 2 })
                if (nodeGroup.length) {
                  const timeout = setTimeout(() => {
                    form.setFieldsValue({
                      nodeGroup
                    })
                    clearTimeout(timeout)
                  }, 500)
                }
              }}>
              下一步
            </Button>
          </div>
        )
      case 2:
        return (
          <div>
            <Button key='previous' size='large' onClick={() => {
              this.setState({ currentStep: 1 })
              const timeout = setTimeout(() => {
                form.setFieldsValue({
                  handlers: this.state.handlers
                })
                clearTimeout(timeout)
              }, 500)
            }}>
              上一步
            </Button>
            <Button key='submit' type='primary' size='large' onClick={() => {
              form.validateFields(this.generatorValidFields(), (err, values) => {
                if (err) {
                  return
                }
                this.props.onCreate({
                  name: this.state.name,
                  method: this.state.method,
                  url: this.state.url,
                  domain: this.state.domain,
                  handlers: this.state.handlers,
                  nodeGroup: this.state.nodeGroup
                })
              })
            }}>
              保存
            </Button>
          </div>
        )
      default:
        return null
    }
  }
  renderFormContent = (step) => {
    switch (step) {
      case 0:
        return this.renderBaseForm()
      case 1:
        return this.renderPluginForm()
      case 2:
        return this.renderNodeForm()
      default:
        return null
    }
  }
  render () {
    const { visible, onCancel, onCreate, title, loading } = this.props
    return (
      <Modal
        visible={visible}
        title={title}
        width={720}
        confirmLoading={loading}
        okText='保存'
        onCancel={() => {
          this.props.form.resetFields()
          const timeout = setTimeout(() => {
            this.setState({ currentStep: 0, handlers: [], nodeGroup: [] })
            clearTimeout(timeout)
          }, 200)
          onCancel()
        }}
        afterClose={() => {
          this.props.form.resetFields()
          const timeout = setTimeout(() => {
            this.setState({ currentStep: 0, handlers: [], nodeGroup: [] })
            clearTimeout(timeout)
          }, 200)
        }}
        onOk={onCreate}
        maskClosable={false}
        footer={this.renderFooter()}
      >
        <Steps current={this.state.currentStep}>
          <Step key={1} title='基础配置' />
          <Step key={2} title='插件配置' />
          <Step key={3} title='节点配置' />
        </Steps>
        <Form style={{ marginTop: '10px' }}>
          {this.renderFormContent(this.state.currentStep)}
        </Form>
      </Modal>
    )
  }
}

ApiModalForm.propTypes = {
  clusters: PropTypes.object,
  plugins: PropTypes.object,
  defaultData: PropTypes.object,
  visible: PropTypes.bool,
  onCancel: PropTypes.func,
  onCreate: PropTypes.func,
  form: PropTypes.object,
  title: PropTypes.string,
  modify: PropTypes.bool,
  loading: PropTypes.bool
}
