import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { Table, Modal, Breadcrumb, Badge, Button, Form, Input, Select, InputNumber, Spin, Switch, message } from 'antd'
import Moment from 'moment'

const FormItem = Form.Item
const Option = Select.Option
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
const BackendForm = Form.create()(
  (props) => {
    const { visible, onCancel, onCreate, form, title, loading, defaultData } = props
    const { getFieldDecorator } = form
    return (
      <Modal
        visible={visible}
        title={title}
        confirmLoading={loading}
        okText='保存'
        onCancel={onCancel}
        onOk={onCreate}
      >
        <Form>
          <FormItem
            {...formItemLayout}
            label='Schema'
          >
            {getFieldDecorator('schema', {
              rules: [{ required: true }]
            })(
              <Select style={{ width: 120 }}>
                <Option value='http'>http</Option>
                <Option value='https'>https</Option>
              </Select>
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='Addr'>
            {getFieldDecorator('addr', {
              rules: [{ required: true }]
            })(
              <Input />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='MaxQPS'>
            {getFieldDecorator('maxQPS', {
              rules: [{ required: true }]
            })(
              <InputNumber min={1} />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='Timeout'>
            {getFieldDecorator('timeout', {
              rules: [{ required: true }]
            })(
              <InputNumber min={1} />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='HeartDisabled'>
            {getFieldDecorator('heartDisabled', {})(
              <Switch
                defaultChecked={defaultData ? defaultData.heartDisabled : false}
              />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='HeartDuration'>
            {getFieldDecorator('heartDuration', {})(
              <InputNumber min={1} />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='HeartPath'>
            {getFieldDecorator('heartPath', {})(
              <Input />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='HeartResponseBody'>
            {getFieldDecorator('heartResponseBody', {})(
              <Input />
            )}
          </FormItem>
        </Form>
      </Modal>
    )
  }
)
export default class BackendsView extends Component {
  constructor (props, context) {
    super(props, context)
    this.state = {
      visible: false,
      title: '',
      loading: false,
      modify: false,
      delLoading: '',
      currentData: {}
    }
  }
  componentDidMount () {
    const { fetchBackends, router } = this.props

    fetchBackends(router.params.clusterName)
  }
  /**
   * 获取主表格列属性
   */
  getBackendsTableColums = () => (
    [
      {
        title: 'Schema',
        dataIndex: 'schema',
        key: 'schema'
      },
      {
        title: 'Addr',
        dataIndex: 'addr',
        key: 'addr'
      },
      {
        title: 'QPS',
        dataIndex: 'QPS',
        key: 'qps'
      },
      {
        title: 'MaxQPS',
        dataIndex: 'maxQPS',
        key: 'maxQPS'
      },
      {
        title: 'Waiting',
        dataIndex: 'waiting',
        key: 'waiting'
      },
      {
        title: 'LastHeartbeat',
        dataIndex: 'lastHeartTime',
        key: 'lastHeartTime',
        render: (record) => <span>{Moment.unix(record).format('lll')}</span>
      },
      {
        title: 'HeartDisabled',
        dataIndex: 'heartDisabled',
        key: 'heartDisabled',
        render: (record) => <Badge status={record ? 'success' : 'error'} />
      },
      {
        title: 'Status',
        dataIndex: 'status',
        key: 'status',
        render: (record) => <Badge status={record ? 'success' : 'error'} />
      },
      {
        title: '操作',
        key: 'backends',
        render: (text, record) => (
          <span>
            <a onClick={() => this.showForm(text)}>编辑</a>
            <span className='ant-divider' />
            <Button
              type='danger'
              onClick={() => this.del(text)}
              loading={this.state.delLoading === text.addr}>
              删除
            </Button>
          </span>
          )
      }
    ]
  )
  showForm = (data) => {
    const form = this.form
    const { router } = this.props
    let title = router.params.clusterName
    let modify
    if (data) {
      form.setFieldsValue({
        schema: data.schema,
        addr: data.addr,
        maxQPS: data.maxQPS,
        timeout: data.Timeout,
        heartDisabled: data.heartDisabled,
        heartDuration: data.heartDuration,
        heartPath: data.heartPath,
        heartResponseBody: data.heartResponseBody
      })
      title = `[${title}]编辑:${data.addr}`
      modify = true
    } else {
      title = `[${title}]添加服务`
      modify = false
    }
    this.setState({
      title,
      modify,
      currentData: data,
      visible: true
    })
  }
  saveFormRef = (form) => {
    this.form = form
  }
  handleCancel = () => {
    this.form.resetFields()
    this.setState({ visible: false })
  }
  handleCreate = () => {
    const form = this.form
    form.validateFields((err, values) => {
      if (err) {
        return
      }
      if (!values.heartDisabled) {
        if (values.heartDuration === undefined) {
          message.error('HeartDuration is required')
          return
        }
        if (values.heartPath === undefined) {
          message.error('HeartPath is required')
          return
        }
      }
      values.QPS = 0
      values.waiting = 0
      let uri
      let postData
      if (this.state.modify) {
        uri = '/v1/backend/update'
        postData = {
          addr: this.state.currentData.addr,
          backendInfo: values
        }
      } else {
        uri = '/v1/backend'
        postData = values
      }
      values.clusterName = this.props.router.params.clusterName
      this.setState({ loading: true })
      fetch(uri,
        { method: 'POST',
          body: JSON.stringify(postData)
        })
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          this.state.modify ? this.props.updateBackend(postData) : this.props.addBackend(values)
          this.setState({ visible: false, loading: false })
          message.success(json.message)
        } else {
          this.setState({ loading: false })
          message.error(json.message)
        }
      })
      .catch(err => {
        this.setState({ loading: false })
        message.error(err)
      })
    })
  }
  del = (value) => {
    this.setState({ delLoading: value.addr })
    fetch('/v1/backend/delete',
      { method: 'POST',
        body: JSON.stringify({
          clusterName: this.props.router.params.clusterName,
          addr: value.addr
        })
      })
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          this.props.deleteBackend(value)
          this.setState({ delLoading: '' })
          message.success(json.message)
        } else {
          this.setState({ delLoading: '' })
          message.error(json.message)
        }
      })
      .catch(err => {
        this.setState({ delLoading: '' })
        message.error(err)
      })
  }
  render () {
    const { router, backends: { fetching, items } } = this.props
    return (
      <div>
        <div className='breadcrump-box'>
          <div className='left'>
            <Breadcrumb separator='>'>
              <Breadcrumb.Item>首页</Breadcrumb.Item>
              <Breadcrumb.Item href='/clusters'>集群</Breadcrumb.Item>
              <Breadcrumb.Item>{router.params.clusterName}</Breadcrumb.Item>
            </Breadcrumb>
          </div>
          <div className='right'>
            <Button type='primary' onClick={() => this.showForm(null)}>添加服务</Button>
          </div>
        </div>
        <Spin spinning={fetching}>
          <Table
            columns={this.getBackendsTableColums()}
            dataSource={items}
            rowKey={(record) => record.addr}
        />
        </Spin>
        <BackendForm
          ref={this.saveFormRef}
          visible={this.state.visible}
          title={this.state.title}
          modify={this.state.modify}
          defaultData={this.state.currentData}
          loading={this.state.loading}
          onCancel={this.handleCancel}
          onCreate={this.handleCreate}
        />
      </div>
    )
  }
}

BackendsView.propTypes = {
  router: PropTypes.object.isRequired,
  backends: PropTypes.object,
  fetchBackends: PropTypes.func,
  addBackend: PropTypes.func,
  updateBackend: PropTypes.func,
  deleteBackend: PropTypes.func
}
