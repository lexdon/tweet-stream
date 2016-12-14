// ------------------------------------
// Constants
// ------------------------------------
export const OPEN_TWEET_STREAM = 'OPEN_TWEET_STREAM'

// ------------------------------------
// Actions
// ------------------------------------
export function openTweetStream(value = '') {
  return {
    type    : OPEN_TWEET_STREAM,
    payload : value
  }
}

/*  This is a thunk, meaning it is a function that immediately
    returns a function for lazy evaluation. It is incredibly useful for
    creating async actions, especially when combined with redux-thunk!

    NOTE: This is solely for demonstration purposes. In a real application,
    you'd probably want to dispatch an action of COUNTER_DOUBLE and let the
    reducer take care of this logic.  */

export const doubleAsync = () => {
  return (dispatch, getState) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        dispatch(increment(getState().counter))
        resolve()
      }, 200)
    })
  }
}

export const actions = {
  openTweetStream
}

// ------------------------------------
// Action Handlers
// ------------------------------------
const ACTION_HANDLERS = {
    [OPEN_TWEET_STREAM] : (state, action) => {
    if (!state.streaming) {
      var exampleSocket = new WebSocket("ws://www.localhost:8080/ws");
      exampleSocket.onopen = function (event) {
        console.log("Connection open!")
      };
      
      exampleSocket.onmessage = function(event) {
        // TODO: Trigger action
        console.log("Received data!")
        console.log(event.data);
      }

      // TODO: Implement ping/pong

      return {
        tweets: state.tweets,
        streaming: true
      }
    }

    return state    
  }
}

// ------------------------------------
// Reducer
// ------------------------------------
const initialState = {
    tweets: [],
    streaming: false
}

export default function streamReducer (state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]

  return handler ? handler(state, action) : state
}
