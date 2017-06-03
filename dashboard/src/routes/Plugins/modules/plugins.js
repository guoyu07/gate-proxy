// ------------------------------------
// Constants
// ------------------------------------
const PLUGINS_REQUEST = 'PLUGINS_REQUEST'
const PLUGINS_SUCCESS = 'PLUGINS_SUCCESS'
const PLUGINS_FAILED = 'PLUGINS_FAILED'
// Add cluster
const PLUGIN_ADD = 'PLUGIN_ADD'
// Update cluster
const PLUGIN_UPDATE = 'PLUGIN_UPDATE'

// Delete cluster
const PLUGIN_DELETE = 'PLUGIN_DELETE'

// ------------------------------------
// Actions
// ------------------------------------

export const addPlugin = (value) => ({
  type: PLUGIN_ADD,
  payload: value
})

export const updatePlugin = (value) => ({
  type: PLUGIN_UPDATE,
  payload: value
})

export const deletePlugin = (value) => ({
  type: PLUGIN_DELETE,
  payload: value
})

function requestStart () {
  return {
    type: PLUGINS_REQUEST
  }
}

export const requestSuccess = (value) => ({
  type: PLUGINS_SUCCESS,
  payload: value
})

export const requestFailed = (value) => ({
  type: PLUGINS_FAILED,
  payload: value
})

export function fetchPlugins () {
  return (dispatch, getState) => {
    if (getState().plugins.fetching) return
    dispatch(requestStart())
    return fetch('/v1/plugins')
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
  [PLUGINS_REQUEST]: (state) => {
    return ({ ...state, fetching: true, msg: '' })
  },
  [PLUGINS_SUCCESS]: (state, action) => {
    return ({ ...state, fetching: false, items: action.payload })
  },
  [PLUGINS_FAILED]: (state, action) => {
    return ({ ...state, fetching: false, msg: action.payload })
  },
  [PLUGIN_ADD]: (state, action) => {
    return ({ ...state,
      items: state.items.concat({
        clusterName: action.payload.name,
        description: action.payload.description })
    })
  },
  [PLUGIN_UPDATE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].clusterName === action.payload.name) {
        state.items[i].description = action.payload.description
      }
    }
    return ({ ...state })
  },
  [PLUGIN_DELETE]: (state, action) => {
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
