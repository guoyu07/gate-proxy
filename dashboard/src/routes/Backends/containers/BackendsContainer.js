import { connect } from 'react-redux'
import { fetchBackends, addBackend, updateBackend, deleteBackend } from './../modules/backends'

import Backends from '../components/BackendsView'

const mapDispatchtoProps = {
  fetchBackends,
  addBackend,
  updateBackend,
  deleteBackend
}

const mapStateToProps = (state) => ({
  backends: state.backends
})

export default connect(mapStateToProps, mapDispatchtoProps)(Backends)
