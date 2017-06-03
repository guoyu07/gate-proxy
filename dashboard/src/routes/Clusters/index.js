import { injectReducer } from '../../store/reducers'

export default (store) => ({
  path: 'clusters',
  getComponent (nextState, cb) {
    require.ensure([], (require) => {
      const Clusters = require('./containers/ClustersContainer').default
      const reducer = require('./modules/clusters').default
      injectReducer(store, [{ key: 'clusters', reducer }])
      cb(null, Clusters)
    }, 'clusters')
  }
})
