package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/oauth1"
	"github.com/gorilla/sessions"
	"github.com/mrjones/oauth"
	"golang.org/x/net/context"
)

var c = oauth.NewConsumer(
	os.Getenv("TWITTER_KEY"),
	os.Getenv("TWITTER_SECRET"),
	oauth.ServiceProvider{
		RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
		AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
	},
)

type Adapter func(http.Handler) http.Handler

const (
	AccessTokenKey string = "accessToken"
	CookieKey      string = "tweet-stream"
)

var (
	twitterKey    string
	twitterSecret string
	store         *sessions.CookieStore
	consumer      *oauth.Consumer
)

func AuthenticateWithTwitter(consumerKey, consumerSecret, sessionSecret string) Adapter {
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	oauth1.NewToken("accessToken", "accessSecret")

	store = sessions.NewCookieStore([]byte(sessionSecret))
	consumer = oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, string(CookieKey))
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			token, ok := session.Values["access_token"]
			if !ok {
				log.Println("No access token found in session")
				authorizeHandler(w, r)
				return
			}

			accessToken, ok := token.(*oauth.AccessToken)
			if !ok {
				log.Println("Unable to cast token to *oauth.AccessToken")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), AccessTokenKey, accessToken)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	tokenUrl := fmt.Sprintf("http://%s/oauth_callback", r.Host)

	requestToken, requestUrl, err := c.GetRequestTokenAndUrl(tokenUrl)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	session, err := store.Get(r, CookieKey)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	session.Values[requestToken.Token] = requestToken
	err = session.Save(r, w)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Redirect user to Twitter sign in
	http.Redirect(w, r, requestUrl, http.StatusFound)
}

func OauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session cookie
	session, _ := store.Get(r, CookieKey)

	// Retrieve query string parameters
	values := r.URL.Query()
	verificationCode := values.Get("oauth_verifier")
	tokenKey := values.Get("oauth_token")

	// Retrieve the request token from the session cookie
	token, ok := session.Values[tokenKey].(*oauth.RequestToken)

	if !ok {
		log.Println("")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Request an access token
	accessToken, err := c.AuthorizeToken(
		token,
		verificationCode,
	)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Store access token for future requests
	session.Values["access_token"] = accessToken

	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/demo/stream", http.StatusFound)
}
