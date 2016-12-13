import { connect } from 'react-redux'
import Stream from '../components/Stream'
import { stream } from '../modules/stream'

const mapDispathToProps = {
    stream
}

const mapStateToProps = (state) => ({
    stream: state.stream
})

export default connect(mapStateToProps, mapDispathToProps)(Stream)