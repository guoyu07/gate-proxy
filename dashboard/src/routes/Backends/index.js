import { injectReducer } from '../../store/reducers'

export default (store) => ({
  path: 'backends/:clusterName',
  getComponent (nextState, cb) {
    require.ensure([], (require) => {
      const Backends = require('./containers/BackendsContainer').default
      const reducer = require('./modules/backends').default
      injectReducer(store, [{ key: 'backends', reducer }])
      cb(null, Backends)
    }, 'backends')
  }
})
