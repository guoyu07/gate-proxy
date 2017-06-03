import { injectReducer } from '../../store/reducers'

export default (store) => ({
  path: 'apis',
  getComponent (nextState, cb) {
    require.ensure([], (require) => {
      const Backends = require('./containers/ApisContainer').default
      const reducer = require('./modules/apis').default
      const cluster = require('../Clusters/modules/clusters').default
      const plugins = require('../Plugins/modules/plugins').default
      injectReducer(store, [
        { key: 'apis', reducer },
        { key: 'clusters', reducer: cluster },
        { key: 'plugins', reducer: plugins }
      ])
      cb(null, Backends)
    }, 'apis')
  }
})
