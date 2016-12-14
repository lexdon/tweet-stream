# TODO

- Find out how too serve the web app from the Go server without routing issues
- Create Websocket listener when clicking "stream" button on Stream-page
- Render streamed tweets
- Implement filter mechanism

# Installation instructions

These instructions assume a Windows environment (using PowerShell).

## Client

### Dependencies

- node 7.1.0
- npm 4.0.3

### How to install

```
cd .\client
npm install
```

### Run

```
cd .\client
npm start
```

## Server

### Dependencies

- go 1.7.3

### How to install

```
go get -u github.com/gorilla/mux
go get -u github.com/gorilla/sessions
go get -u github.com/gorilla/websocket
go get -u github.com/mrjones/oauth
```

# How to run

## Environment variables

The following environment variables need to be set:

<table>
  <thead>
    <tr>
      <th>Environment variable<th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>TWITTER_KEY</code><td>
      <td>The Twitter consumer key</td>
    </tr>
    <tr>
      <td><code>TWITTER_SECRET</code><td>
      <td>The Twitter consumer secret</td>
    </tr>
    <tr>
      <td><code>SESSION_SECRET</code><td>
      <td>Secret used to encrypt/decrypt secure session cookies</td>
    </tr>
  </tbody>
</table>

## Commands

```
cd .\server
go run .\main.go .\stream.go
```

# Challenges

- Can get a firehose of tweets depending on what you're tracking

# Remaining work before production ready

- Creater Docker development image
- Create Docker production image
- Vendor Go dependencies (with e.g. glide, gb)
- Write tests
- Run load tests with `-race` enabled
- Implement persistent session storage (e.g. Redis) & CSRF protection (e.g. gorilla/csrf)
- Implement Twitter auth in popup 
    - [http://stackoverflow.com/questions/1878529/twitter-oauth-via-a-popup](http://stackoverflow.com/questions/1878529/twitter-oauth-via-a-popup)
    - [http://clarkdave.net/2012/10/2012-10-30-twitter-oauth-authorisation-in-a-popup/](http://clarkdave.net/2012/10/2012-10-30-twitter-oauth-authorisation-in-a-popup/)

# Nice to have

- One step build
- Nicer GUI