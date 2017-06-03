import { connect } from 'react-redux'
import { fetchClusters, addCluster, updateCluster, deleteCluster } from './../modules/clusters'

import Clusters from '../components/ClustersView'

const mapDispatchtoProps = {
  fetchClusters,
  addCluster,
  updateCluster,
  deleteCluster
}

const mapStateToProps = (state) => ({
  clusters: state.clusters
})

export default connect(mapStateToProps, mapDispatchtoProps)(Clusters)
