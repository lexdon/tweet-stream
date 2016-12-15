# Description

This application is a proof of concept utilizing the Public streams part of the Twitter Streaming APIs to stream all new tweets regarding a specific topic (defined by the value of the environment variable `TWEET_STREAM_TRACK`).

These updates are then streamed to a React client using Server-Sent Events. On the client, you're able to filter Tweets based on their contents (a simple contains text filter for now).

The application uses "Sign in with Twitter". You therefore need a valid Twitter user account to access the Twitter streaming web application.

Additionally, you need to set the `TWITTER_CONSUMER_KEY` and `TWITTER_CONSUMER_SECRET` environment variables to enable the server to communicate with the Twitter API. These can be obtained at [https://apps.twitter.com/](https://apps.twitter.com/).

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

## Server

### Dependencies

- go 1.7.3

### How to install

```
go get -u golang.org/x/net/context
go get -u github.com/dghubble/go-twitter/twitter
go get -u github.com/dghubble/oauth1
go get -u github.com/dghubble/gologin
go get -u github.com/dghubble/sessions
```

# How to run

## Environment variables

The following environment variables need to be set:

<table>
  <thead>
    <tr>
      <th>Environment variable</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>TWITTER_CONSUMER_KEY</code></td>
      <td>The Twitter consumer key</td>
    </tr>
    <tr>
      <td><code>TWITTER_CONSUMER_SECRET</code></td>
      <td>The Twitter consumer secret</td>
    </tr>
    <tr>
      <td><code>TWEET_STREAM_TRACK</code></td>
      <td>Which topic to track in the stream</td>
    </tr>
    <tr>
      <td><code>TWEET_STREAM_SERVER_PORT</code></td>
      <td>Port to expose the server on (optional)</td>
    </tr>
  </tbody>
</table>

## Build & run

```
cd .\client 
npm run build
cd ..\server
go run .\main.go
```

Then visit [localhost:8080](localhost:8080)

# Challenges

- Can get a firehose of tweets depending on what you're tracking

# Remaining work before production ready

- Creater Docker development image
- Create Docker production image
- One step build
- Nicer GUI
- More advanced filtering/sorting
- Vendor Go dependencies (with e.g. glide, gb)
- Write tests
- Run load tests with `-race` enabled
- Implement persistent session storage (e.g. Redis) & CSRF protection (e.g. gorilla/csrf)
- Take a closer look at the use of go channels and make sure we don't end up leaking goroutines
- Remove unsafe casts (use context and middleware to propagate session state instead)
- Add EventSource polyfill to client
- Implement rate limiting on the server for "firehose" Twitter topics
- Implement max length for tweets array on client to prevent rendering performance issues
- Use reselect to limit processing of existing tweets when a new tweet arrives on the client
- Add the ability to choose what topic to track on the client (e.g. by specifying a query string parameter on the stream request to the server)

# Nice to have 
- Implement Twitter auth in popup