import { connect } from 'react-redux'
import { fetchApis, addApi, updateApi, deleteApi } from './../modules/apis'
import { fetchClusters } from '../../Clusters/modules/clusters'
import { fetchPlugins } from '../../Plugins/modules/plugins'

import Apis from '../components/ApisView'

const mapDispatchtoProps = {
  fetchApis,
  addApi,
  updateApi,
  deleteApi,
  fetchClusters,
  fetchPlugins
}

const mapStateToProps = (state) => ({
  apis: state.apis,
  clusters: state.clusters,
  plugins: state.plugins
})

export default connect(mapStateToProps, mapDispatchtoProps)(Apis)
