import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { withRouter } from 'react-router'
import { Menu } from 'antd'

class Header extends Component {
  constructor (props) {
    super(props)
    this.onSelect = this.onSelect.bind(this)
    this.state = {
      current: '/'
    }
  }
  onSelect = (e) => {
    this.props.router.push(e.key)
  }
  getCurrentKey = () => {
    const key = this.props.router.getCurrentLocation().pathname
    if (key === '/') {
      return key
    } else if (key.indexOf('/cluster') !== -1 || key.indexOf('backend') !== -1) {
      return '/clusters'
    } else if (key.indexOf('/api') !== -1) {
      return '/apis'
    } else {
      return key
    }
  }
  render () {
    return (
      <div>
        <Menu
          onClick={this.onSelect}
          selectedKeys={[this.getCurrentKey()]}
          mode='horizontal'
        >
          <Menu.Item key='/'>
            首页
          </Menu.Item>
          <Menu.Item key='/clusters'>
            集群列表
          </Menu.Item>
          <Menu.Item key='/apis'>
            路由表
          </Menu.Item>
        </Menu>
      </div>
    )
  }
}
Header.propTypes = {
  router: PropTypes.object.isRequired
}
export default withRouter(Header)
