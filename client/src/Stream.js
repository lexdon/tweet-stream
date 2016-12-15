import React, { PropTypes } from 'react';
import { connect } from 'react-redux'

const Stream = (props) => {
  return (
      <div className="Stream">
        <p>Recent tweets</p>
        <ul>
            {props.tweets.map(function(tweet) {
                return <li>{tweet.text}</li>
            })}
        </ul>
      </div>
    )
}

Stream.propTypes = {
    tweets: PropTypes.array.isRequired
}

const mapStateToProps = (state) => {
    let filteredTweets = state.tweets.filter(tweet => tweet.text.includes(state.filterContains))

    return {
        tweets: filteredTweets
    }
}

const mapDispatchToProps = (dispatch) => {
    return {}
}

const StreamContainer = connect(
    mapStateToProps,
    mapDispatchToProps
)(Stream)

export default StreamContainer