# Challenges

- Can get a firehose of tweets depending on what you're tracking

# Remaining work before production ready

- Run tests with `-race` enabled
- Implement persistent session storage (e.g. Redis) & CSRF protection (e.g. gorilla/csrf)
- Implement Twitter auth in popup 
    - [http://stackoverflow.com/questions/1878529/twitter-oauth-via-a-popup](http://stackoverflow.com/questions/1878529/twitter-oauth-via-a-popup)
    - [http://clarkdave.net/2012/10/2012-10-30-twitter-oauth-authorisation-in-a-popup/](http://clarkdave.net/2012/10/2012-10-30-twitter-oauth-authorisation-in-a-popup/)

# Nice to have

- One step build
- Nicer GUI