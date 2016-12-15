import React from 'react'
import logo from './logo.svg'
import './App.css'
import Stream from './Stream'
import StreamFilter from './StreamFilter'

const App = (props) => {
  return (
      <div className="App">
        <div className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <h2>Tweet Stream</h2>
        </div>
        <StreamFilter />
        <Stream />
      </div>
    )
}

export default App
