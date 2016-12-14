// Based on https://github.com/rodreegez/go-signin-with-twitter/blob/master/server.go
// and https://github.com/NOX73/go-twitter-stream-api/blob/master/twitter_api.go

package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/mrjones/oauth"
)

type Message struct {
	Error    error
	Response *http.Response
	Tweet    *Tweet
}

type TweetJSON struct {
	Text string
	User struct {
		Id                      int
		Screen_name             string
		Name                    string
		Description             string
		Profile_image_url_https string
	}
}

type Tweet struct {
	Body string
	JSON *TweetJSON
}

var port int
var cookieKey string = "tweet-stream"

var c = oauth.NewConsumer(
	os.Getenv("TWITTER_KEY"),
	os.Getenv("TWITTER_SECRET"),
	oauth.ServiceProvider{
		RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
		AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
	},
)

var notAuthenticatedTemplate = template.Must(template.New("").Parse(`
<html><body>
Auth w/ Twitter:
<form action="/authorize" method="POST"><input type="submit" value="Ok, authorize this app with my id"/></form>
</body></html>
`))

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	flag.IntVar(&port, "port", 8080, "port to listen to")

	// Needed to store structs in secure cookie
	gob.Register(&oauth.RequestToken{})
	gob.Register(&oauth.AccessToken{})
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", HomePageHandler)

	// TODO: Wrap in session check handler
	r.Handle("/demo/", http.StripPrefix("/demo/", http.FileServer(http.Dir("../client/dist"))))
	r.HandleFunc("/ws", StreamHandler)
	r.HandleFunc("/authorize", AuthorizeHandler)
	r.HandleFunc("/oauth_callback", OauthCallbackHandler)

	fmt.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	// Check if access token is already present in session cookie.
	session, err := store.Get(r, cookieKey)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	token, ok := session.Values["access_token"]

	if !ok {
		// Not authenticated, show login button
		notAuthenticatedTemplate.Execute(w, nil)
		return
	}

	if _, ok = token.(*oauth.AccessToken); ok {
		// Already authenticated
		fmt.Println("### Aleady authenticated, redirecting")
		//http.Error(w, "error string", http.StatusNotFound)
		http.Redirect(w, r, "/demo/stream", http.StatusFound)
	} else {
		log.Println(
			"Access token retrieved from secure cookie " +
				"cannot be casted to type *oauth.AccessToken")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	tokenUrl := fmt.Sprintf("http://%s/oauth_callback", r.Host)

	requestToken, requestUrl, err := c.GetRequestTokenAndUrl(tokenUrl)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	session, err := store.Get(r, cookieKey)
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
	session, _ := store.Get(r, cookieKey)

	// Retrieve query string parameters
	values := r.URL.Query()
	verificationCode := values.Get("oauth_verifier")
	tokenKey := values.Get("oauth_token")

	// Retrieve the request token from the session cookie
	token, ok := session.Values[tokenKey].(*oauth.RequestToken)

	if !ok {
		log.Println("")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
