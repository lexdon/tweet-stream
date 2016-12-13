# Environment variables

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

# Challenges

- Can get a firehose of tweets depending on what you're tracking

# Remaining work before production ready

- Run load tests with `-race` enabled
- Implement persistent session storage (e.g. Redis) & CSRF protection (e.g. gorilla/csrf)
- Implement Twitter auth in popup 
    - [http://stackoverflow.com/questions/1878529/twitter-oauth-via-a-popup](http://stackoverflow.com/questions/1878529/twitter-oauth-via-a-popup)
    - [http://clarkdave.net/2012/10/2012-10-30-twitter-oauth-authorisation-in-a-popup/](http://clarkdave.net/2012/10/2012-10-30-twitter-oauth-authorisation-in-a-popup/)

# Nice to have

- One step build
- Nicer GUI