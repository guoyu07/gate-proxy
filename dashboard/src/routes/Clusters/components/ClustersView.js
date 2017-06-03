import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { Table, Spin, Form, Modal, Input, Breadcrumb, Button, message } from 'antd'

const FormItem = Form.Item
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
const ClusterModifyForm = Form.create()(
  (props) => {
    const { visible, onCancel, onCreate, form, title, modify, loading } = props
    const { getFieldDecorator } = form
    return (
      <Modal
        visible={visible}
        title={title}
        confirmLoading={loading}
        okText='保存'
        onCancel={onCancel}
        onOk={onCreate}
        maskClosable={false}
      >
        <Form>
          <FormItem
            {...formItemLayout}
            label='ClusterName'
          >
            {getFieldDecorator('name', {
              rules: [{ required: true }]
            })(
              <Input disabled={modify} />
            )}
          </FormItem>
          <FormItem
            {...formItemLayout}
            label='Description'>
            {getFieldDecorator('description', {
              rules: [{ required: true }]
            })(
              <Input type='textarea' rows={4} />
            )}
          </FormItem>
        </Form>
      </Modal>
    )
  }
)

export default class Clusters extends Component {
  constructor (props, context) {
    super(props, context)
    this.state = {
      visible: false,
      title: '',
      loading: false,
      delLoading: '',
      modify: false
    }
  }
  componentDidMount () {
    this.props.fetchClusters()
  }
  /**
   * 获取主表格列属性
   */
  getClusterTableColums = () => (
    [
      {
        title: '集群名称',
        dataIndex: 'clusterName',
        key: 'clusterName'
      },
      {
        title: '简介',
        dataIndex: 'description',
        key: 'description'
      },
      {
        title: '服务数量',
        dataIndex: 'backendNum',
        key: 'backendNum'
      },
      {
        title: '操作',
        key: 'backends',
        render: (text, record) => (
          <span>
            <a onClick={() => this.goBackends(record)}>服务列表</a>
            <span className='ant-divider' />
            <a onClick={() => this.showForm(record)}>编辑</a>
            <span className='ant-divider' />
            <Button
              type='danger'
              onClick={() => this.del(text)}
              loading={this.state.delLoading === text.clusterName}>
              删除
            </Button>
          </span>
          )
      }
    ]
  )
  goBackends = (record) => {
    const { router } = this.props
    router.push({
      pathname: `/backends/${record.clusterName}`
    })
  }
  showForm = (data) => {
    const form = this.form
    let title
    let modify
    if (data) {
      form.setFieldsValue({
        name: data.clusterName,
        description: data.description
      })
      modify = true
      title = `编辑:[${data.clusterName}]`
    } else {
      title = `添加集群`
      modify = false
    }
    this.setState({
      title,
      modify,
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
      let uri
      if (this.state.modify) {
        uri = '/v1/cluster/update'
      } else {
        uri = '/v1/cluster'
      }
      this.setState({ loading: true })
      fetch(uri,
        { method: 'POST',
          body: JSON.stringify(values)
        })
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          this.state.modify ? this.props.updateCluster(values) : this.props.addCluster(values)
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
    this.setState({ delLoading: value.clusterName })
    fetch('/v1/cluster/delete',
      { method: 'POST',
        body: JSON.stringify({
          name: value.clusterName
        })
      })
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          this.props.deleteCluster(value)
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
    const { clusters: { fetching, items } } = this.props
    return (
      <div>
        <div className='breadcrump-box'>
          <div className='left'>
            <Breadcrumb separator='>'>
              <Breadcrumb.Item>首页</Breadcrumb.Item>
              <Breadcrumb.Item>集群</Breadcrumb.Item>
            </Breadcrumb>
          </div>
          <div className='right'>
            <Button type='primary' onClick={() => this.showForm(null)}>添加集群</Button>
          </div>
        </div>
        <Spin spinning={fetching}>
          <Table
            columns={this.getClusterTableColums()}
            dataSource={items}
            rowKey={(record) => record.clusterName}
          />
        </Spin>
        <ClusterModifyForm
          ref={this.saveFormRef}
          visible={this.state.visible}
          title={this.state.title}
          modify={this.state.modify}
          loading={this.state.loading}
          onCancel={this.handleCancel}
          onCreate={this.handleCreate}
        />
      </div>
    )
  }
}

Clusters.propTypes = {
  router: PropTypes.object.isRequired,
  clusters: PropTypes.object.isRequired,
  fetchClusters: PropTypes.func,
  addCluster: PropTypes.func,
  updateCluster:PropTypes.func,
  deleteCluster:PropTypes.func
}
