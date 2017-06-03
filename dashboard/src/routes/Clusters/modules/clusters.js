// ------------------------------------
// Constants
// ------------------------------------
const CLUSTERS_REQUEST = 'CLUSTERS_REQUEST'
const CLUSTERS_SUCCESS = 'CLUSTERS_SUCCESS'
const CLUSTERS_FAILED = 'CLUSTERS_FAILED'
// Add cluster
const CLUSTER_ADD = 'CLUSTER_ADD'
// Update cluster
const CLUSTER_UPDATE = 'CLUSTER_UPDATE'

// Delete cluster
const CLUSTER_DELETE = 'CLUSTER_DELETE'

// ------------------------------------
// Actions
// ------------------------------------

export const addCluster = (value) => ({
  type: CLUSTER_ADD,
  payload: value
})

export const updateCluster = (value) => ({
  type: CLUSTER_UPDATE,
  payload: value
})

export const deleteCluster = (value) => ({
  type: CLUSTER_DELETE,
  payload: value
})

function requestStart () {
  return {
    type: CLUSTERS_REQUEST
  }
}

export const requestSuccess = (value) => ({
  type: CLUSTERS_SUCCESS,
  payload: value
})

export const requestFailed = (value) => ({
  type: CLUSTERS_FAILED,
  payload: value
})

export function fetchClusters () {
  return (dispatch, getState) => {
    if (getState().clusters.fetching) return
    dispatch(requestStart())
    return fetch('/v1/clusters')
      .then(data => data.json())
      .then(json => {
        if (json.code === 0) {
          dispatch(requestSuccess(json.data))
        } else {
          dispatch(requestFailed(json.message))
        }
      })
      .catch(err => dispatch(requestFailed(err)))
  }
}

// ------------------------------------
// Action Handlers
// ------------------------------------
const ACTION_HANDLERS = {
  [CLUSTERS_REQUEST]: (state) => {
    return ({ ...state, fetching: true, msg: '' })
  },
  [CLUSTERS_SUCCESS]: (state, action) => {
    return ({ ...state, fetching: false, items: action.payload })
  },
  [CLUSTERS_FAILED]: (state, action) => {
    return ({ ...state, fetching: false, msg: action.payload })
  },
  [CLUSTER_ADD]: (state, action) => {
    return ({ ...state,
      items: state.items.concat({
        clusterName: action.payload.name,
        description: action.payload.description })
    })
  },
  [CLUSTER_UPDATE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].clusterName === action.payload.name) {
        state.items[i].description = action.payload.description
      }
    }
    return ({ ...state })
  },
  [CLUSTER_DELETE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].clusterName === action.payload.clusterName) {
        state.items.splice(i, 1)
      }
    }
    return ({ ...state, items: state.items })
  }
}

// ------------------------------------
// Reducer
// ------------------------------------
const initialState = {
  fetching: false,
  items: [],
  msg: ''
}
export default function (state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]

  return handler ? handler(state, action) : state
}
