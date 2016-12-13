import { injectReducer } from '../../store/reducers'

export default (store) => ({
  path : 'stream',
  /*  Async getComponent is only invoked when route matches   */
  getComponent (nextState, cb) {
    /*  Webpack - use 'require.ensure' to create a split point
        and embed an async module loader (jsonp) when bundling   */
    require.ensure([], (require) => {
      /*  Webpack - use require callback to define
          dependencies for bundling   */
      const Stream = require('./containers/StreamContainer').default
      const reducer = require('./modules/stream').default

      /*  Add the reducer to the store on key 'stream'  */
      injectReducer(store, { key: 'stream', reducer })

      /*  Return getComponent   */
      cb(null, Stream)

    /* Webpack named bundle   */
    }, 'stream')
  }
})
