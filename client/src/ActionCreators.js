import {ADD_TWEET, FILTER_CONTAINS} from './ActionTypes'

export function addTweet(tweet) {
    return {
        type: ADD_TWEET,
        tweet
    }
}

export function filterContains(filter) {
    return {
        type: FILTER_CONTAINS,
        filter
    }
}