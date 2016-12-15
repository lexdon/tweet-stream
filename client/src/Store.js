import { createStore } from 'redux'
import streamApp from './Reducers'

const Store = createStore(streamApp)

let unsubscribe = Store.subscribe(() => console.log(Store.getState()))

export default Store