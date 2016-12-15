import React, { PropTypes } from 'react';
import { connect } from 'react-redux'
import { filterContains } from './ActionCreators'
import Store from './Store'

const StreamFilter = (props) => {
  return (
      <div className="filter">
        <p>Filter: Contains</p>
        <input type='text' text={props.filter} onChange={event => {
            Store.dispatch(filterContains(event.target.value))
        }} />
      </div>
    )
}

StreamFilter.propTypes = {
    onChange: PropTypes.func.isRequired,
    filter: PropTypes.string.isRequired
}

const mapStateToProps = (state) => {
    return {
        filter: state.containsFilter
    }
}

const mapDispatchToProps = (dispatch) => {
    return {}
}

const StreamFilterContainer = connect(
    mapStateToProps,
    mapDispatchToProps
)(StreamFilter)

export default StreamFilterContainer