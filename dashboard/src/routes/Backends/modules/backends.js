// ------------------------------------
// Constants
// ------------------------------------
const BACKENDS_REQUEST = 'BACKENDS_REQUEST'
const BACKENDS_SUCCESS = 'BACKENDS_SUCCESS'
const BACKENDS_FAILED = 'BACKENDS_FAILED'
// Add cluster
const BACKEND_ADD = 'BACKEND_ADD'
// Update cluster
const BACKEND_UPDATE = 'BACKEND_UPDATE'

// Delete cluster
const BACKEND_DELETE = 'BACKEND_DELETE'

// ------------------------------------
// Actions
// ------------------------------------

export const addBackend = (value) => ({
  type: BACKEND_ADD,
  payload: value
})

export const updateBackend = (value) => ({
  type: BACKEND_UPDATE,
  payload: value
})

export const deleteBackend = (value) => ({
  type: BACKEND_DELETE,
  payload: value
})

function requestStart () {
  return {
    type: BACKENDS_REQUEST
  }
}

export const requestSuccess = (value) => ({
  type: BACKENDS_SUCCESS,
  payload: value
})

export const requestFailed = (value) => ({
  type: BACKENDS_FAILED,
  payload: value
})

export function fetchBackends (clusterName) {
  return (dispatch, getState) => {
    if (getState().backends.fetching) return
    dispatch(requestStart())
    return fetch(`/v1/backends/${clusterName}`)
      .then(data => data.json())
      .then(json => {
        if (json.code === 0 && json.data) {
          dispatch(requestSuccess(json.data))
        } else {
          dispatch(requestFailed(json.message || ''))
        }
      })
      .catch(err => dispatch(requestFailed(err)))
  }
}

// ------------------------------------
// Action Handlers
// ------------------------------------
const ACTION_HANDLERS = {
  [BACKENDS_REQUEST]: (state) => {
    return ({ ...state, fetching: true, msg: '' })
  },
  [BACKENDS_SUCCESS]: (state, action) => {
    return ({ ...state, fetching: false, items: action.payload })
  },
  [BACKENDS_FAILED]: (state, action) => {
    return ({ ...state, fetching: false, msg: action.payload })
  },
  [BACKEND_ADD]: (state, action) => {
    action.payload.status = action.payload.heartDisabled
    return ({ ...state,
      items: state.items.concat(action.payload) })
  },
  [BACKEND_UPDATE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].addr === action.payload.addr) {
        state.items[i] = action.payload.backendInfo
      }
    }
    return ({ ...state, items: state.items })
  },
  [BACKEND_DELETE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].addr === action.payload.addr) {
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
