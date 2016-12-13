import React from 'react'

export const Stream = (props) => (
  <div style={{ margin: '0 auto' }} >
    <h2>Stream</h2>
    {/*<button className='btn btn-default' onClick={props.increment}>
      Increment
    </button>
    {' '}
    <button className='btn btn-default' onClick={props.doubleAsync}>
      Double (Async)
    </button>*/}
  </div>
)

Stream.propTypes = {
  counter     : React.PropTypes.number.isRequired,
  doubleAsync : React.PropTypes.func.isRequired,
  increment   : React.PropTypes.func.isRequired
}

export default Stream