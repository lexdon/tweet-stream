import { connect } from 'react-redux'
import Stream from '../components/Stream'

const mapDispathToProps = {

}

const mapStateToProps = (state) => ({
    stream: state.stream
})

export default connect(mapStateToProps, mapDispathToProps)(Stream)