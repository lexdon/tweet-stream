import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';
import './index.css';
import { Provider } from 'react-redux'
import { addTweet } from './ActionCreators'
import Store from './Store'

const logout = () => {
  fetch("/logout", {
    method: "POST"
  })
}

const props = {
  logout
}

ReactDOM.render(
  <Provider store={Store}>
    <App props={props} />
  </Provider>,
  document.getElementById('root')
);


// Initiate EventSource
var eventSource = new EventSource("http://localhost:8080/api/stream");

eventSource.onmessage = function(e) {
  Store.dispatch(addTweet(JSON.parse(e.data)))
}

eventSource.onerror = function(e) {
  console.error("EventSource failed")
}

console.info("EventSource initialized")