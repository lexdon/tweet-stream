// Based on https://github.com/dghubble/gologin/tree/master/examples/twitter

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/dghubble/sessions"
	"golang.org/x/net/context"
)

// Config configures the main ServeMux.
type Config struct {
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	Port                  string
}

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

const (
	sessionName    = "tweet-stream"
	sessionSecret  = "very secret session secret"
	sessionUserKey = "twitterID"
)

func main() {
	// Environment variables
	config := &Config{
		TwitterConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		TwitterConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		Port: os.Getenv("TWEET_STREAM_SERVER_PORT"),
	}

	// Validate presence of config values
	if config.TwitterConsumerKey == "" {
		log.Fatal("TWITTER_CONSUMER_KEY env variable not defined")
	}
	if config.TwitterConsumerSecret == "" {
		log.Fatal("TWITTER_CONSUMER_SECRET env variable not defined")
	}

	// TODO: Better port validation
	if config.Port == "" {
		config.Port = "8080"
	}

	// Start server
	log.Printf("Starting server listening on %s\n", config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", config.Port), New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	// authorizedAdapter := auth.Authorize([]byte(sessionSecret))

	// r.HandleFunc("/", HomePageHandler)

	// // TODO: Wrap in session check handler
	// r.Handle("/demo/", http.StripPrefix("/demo/", http.FileServer(http.Dir("../client/dist"))))
	// r.HandleFunc("/ws", StreamHandler)
	// r.HandleFunc("/authorize", AuthorizeHandler)
	// r.HandleFunc("/oauth_callback", OauthCallbackHandler)
}

// New returns a new ServeMux with app routes.
func New(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/stream", requireLogin(http.HandlerFunc(streamHandler)))
	// mux.HandleFunc("/logout", logoutHandler)
	// 1. Register Twitter login and callback handlers
	oauth1Config := &oauth1.Config{
		ConsumerKey:    config.TwitterConsumerKey,
		ConsumerSecret: config.TwitterConsumerSecret,
		CallbackURL:    fmt.Sprintf("http://localhost:%s/twitter/callback", config.Port),
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}
	mux.Handle("/twitter/login", ctxh.NewHandler(twitter.LoginHandler(oauth1Config, nil)))
	mux.Handle("/twitter/callback", ctxh.NewHandler(twitter.CallbackHandler(oauth1Config, issueSession(), nil)))
	return mux
}

// issueSession issues a cookie session after successful Twitter login
func issueSession() ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		twitterUser, err := twitter.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Implement a success handler to issue some form of session
		session := sessionStore.New(sessionName)
		session.Values[sessionUserKey] = twitterUser.ID
		session.Save(w)
		http.Redirect(w, req, "/stream", http.StatusFound)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// welcomeHandler shows a welcome message and login button.
func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	if isAuthenticated(req) {
		http.Redirect(w, req, "/stream", http.StatusFound)
		return
	}
	page, _ := ioutil.ReadFile("index.html")
	fmt.Fprintf(w, string(page))
}

// streamHandler shows protected user content.
func streamHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, `<p>You are logged in!</p><form action="/logout" method="post"><input type="submit" value="Logout"></form>`)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		sessionStore.Destroy(w, sessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

// requireLogin redirects unauthenticated users to the login route.
func requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !isAuthenticated(req) {
			http.Redirect(w, req, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// isAuthenticated returns true if the user has a signed session cookie.
func isAuthenticated(req *http.Request) bool {
	if _, err := sessionStore.Get(req, sessionName); err == nil {
		return true
	}
	return false
}
