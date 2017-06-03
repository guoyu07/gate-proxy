// ------------------------------------
// Constants
// ------------------------------------
const APIS_REQUEST = 'APIS_REQUEST'
const APIS_SUCCESS = 'APIS_SUCCESS'
const APIS_FAILED = 'APIS_FAILED'
// Add cluster
const API_ADD = 'API_ADD'
// Update cluster
const API_UPDATE = 'API_UPDATE'

// Delete cluster
const API_DELETE = 'API_DELETE'

// ------------------------------------
// Actions
// ------------------------------------

export const addApi = (value) => ({
  type: API_ADD,
  payload: value
})

export const updateApi = (value) => ({
  type: API_UPDATE,
  payload: value
})

export const deleteApi = (value) => ({
  type: API_DELETE,
  payload: value
})

function requestStart () {
  return {
    type: APIS_REQUEST
  }
}

export const requestSuccess = (value) => ({
  type: APIS_SUCCESS,
  payload: value
})

export const requestFailed = (value) => ({
  type: APIS_FAILED,
  payload: value
})

export function fetchApis () {
  return (dispatch, getState) => {
    if (getState().apis.fetching) return
    dispatch(requestStart())
    return fetch(`/v1/apis`)
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
  [APIS_REQUEST]: (state) => {
    return ({ ...state, fetching: true, msg: '' })
  },
  [APIS_SUCCESS]: (state, action) => {
    return ({ ...state, fetching: false, items: action.payload })
  },
  [APIS_FAILED]: (state, action) => {
    return ({ ...state, fetching: false, msg: action.payload })
  },
  [API_ADD]: (state, action) => {
    return ({ ...state,
      items: state.items.concat(action.payload) })
  },
  [API_UPDATE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].method === action.payload.method && state.items[i].url === action.payload.url) {
        state.items[i] = action.payload.info
      }
    }
    return ({ ...state, items: state.items })
  },
  [API_DELETE]: (state, action) => {
    // edit
    for (let i = 0; i < state.items.length; i++) {
      if (state.items[i].method === action.payload.method && state.items[i].url === action.payload.url) {
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
