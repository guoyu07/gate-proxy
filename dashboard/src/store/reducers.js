import { combineReducers } from 'redux'
import locationReducer from './location'

export const makeRootReducer = (asyncReducers) => {
  return combineReducers({
    location: locationReducer,
    ...asyncReducers
  })
}

export const injectReducer = (store, reducers) => {
  for (let i = 0; i < reducers.length; i++) {
    store.asyncReducers[reducers[i].key] = reducers[i].reducer
  }
  store.replaceReducer(makeRootReducer(store.asyncReducers))
}

export default makeRootReducer
