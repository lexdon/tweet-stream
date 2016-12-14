import { connect } from 'react-redux'
import Stream from '../components/Stream'
import { openTweetStream } from '../modules/stream'

const mapDispathToProps = {
    openTweetStream
}

const mapStateToProps = (state) => ({
    stream: state.stream
})

export default connect(mapStateToProps, mapDispathToProps)(Stream)