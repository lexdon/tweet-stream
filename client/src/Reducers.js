import { ADD_TWEET, FILTER_CONTAINS } from './ActionTypes'

const initialState = {
    tweets: [],
    filterContains: ''
}

function streamApp(state = initialState, action) {
    switch (action.type) {
        case ADD_TWEET:
            return Object.assign({}, state, {
                tweets: [
                    ...state.tweets,
                    action.tweet
                ]
            })
        case FILTER_CONTAINS: 
            return Object.assign({}, state, {
                 filterContains: action.filter
            }) 
        default:
            return state
    }
}

export default streamApp