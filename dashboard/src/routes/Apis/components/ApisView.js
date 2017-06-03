import React, { Component } from 'react'
import PropTypes from 'prop-types'
import {
  Table,
  Breadcrumb,
  Button,
  Tag,
  Form,
  Spin,
  message
} from 'antd'

import ApiModalForm from './ApiModalForm'

const ApiModalFormComponent = Form.create()(ApiModalForm)

export default class ApisView extends Component {
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
    const { fetchApis, fetchClusters, fetchPlugins } = this.props
    fetchApis()
    fetchClusters()
    fetchPlugins()
  }
  /**
   * 获取主表格列属性
   */
  getBackendsTableColums = () => (
    [
      {
        title: 'Name',
        dataIndex: 'name',
        key: 'name'
      },
      {
        title: 'Method',
        dataIndex: 'method',
        key: 'method'
      },
      {
        title: 'URL',
        dataIndex: 'url',
        key: 'url'
      },
      {
        title: 'Handlers',
        key: 'handles',
        render: (record) => {
          if (record.handlers === null) {
            return null
          }
          return (
            record.handlers.map(item => <Tag key={`${record.url}-${item}`} color='#f50'>{item}</Tag>)
          )
        }
      },
      {
        title: 'Nodes',
        key: 'nodes',
        render: (record) => <span>{record.nodeGroup ? record.nodeGroup.length : 0}</span>
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
    let title
    let modify
    if (data) {
      this.form.setFieldsValue({
        name: data.name,
        method: data.method,
        url: data.url,
        domain: data.domain
      })
      modify = true
      title = `编辑:[${data.url}]`
    } else {
      title = `添加路由`
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
  handleCreate = (info) => {
    let uri
    let values
    if (this.state.modify) {
      uri = '/v1/api/update'
      values = {
        info,
        method: this.state.currentData.method,
        url: this.state.currentData.url
      }
    } else {
      uri = '/v1/api'
      values = info
    }
    this.setState({ loading: true })
    fetch(uri,
      {
        method: 'POST',
        body: JSON.stringify(values)
      })
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          this.state.modify ? this.props.updateApi(values) : this.props.addApi(values)
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
  }
  del = (value) => {
    this.setState({ delLoading: value.addr })
    fetch('/v1/api/delete',
      { method: 'POST',
        body: JSON.stringify({
          method: value.method,
          url: value.url
        })
      })
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          this.props.deleteApi(value)
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
    const { apis: { fetching, items } } = this.props
    return (
      <div>
        <div className='breadcrump-box'>
          <div className='left'>
            <Breadcrumb separator='>'>
              <Breadcrumb.Item>首页</Breadcrumb.Item>
              <Breadcrumb.Item>路由</Breadcrumb.Item>
            </Breadcrumb>
          </div>
          <div className='right'>
            <Button type='primary' onClick={() => this.showForm(null)}>添加路由</Button>
          </div>
        </div>
        <Spin spinning={fetching}>
          <Table
            columns={this.getBackendsTableColums()}
            dataSource={items}
            rowKey={(record) => record.url}
        />
        </Spin>
        <ApiModalFormComponent
          ref={this.saveFormRef}
          visible={this.state.visible}
          title={this.state.title}
          modify={this.state.modify}
          loading={this.state.loading}
          defaultData={this.state.currentData}
          onCancel={this.handleCancel}
          clusters={this.props.clusters}
          plugins={this.props.plugins}
          onCreate={(entry) => this.handleCreate(entry)}
        />
      </div>
    )
  }
}

ApisView.propTypes = {
  apis: PropTypes.object,
  clusters: PropTypes.object,
  fetchApis: PropTypes.func,
  fetchClusters: PropTypes.func,
  fetchPlugins: PropTypes.func,
  plugins: PropTypes.object,
  addApi: PropTypes.func,
  updateApi: PropTypes.func,
  deleteApi: PropTypes.func
}
