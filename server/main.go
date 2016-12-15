// Based on https://github.com/dghubble/gologin/tree/master/examples/twitter

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/go-twitter/twitter"
	oauth1login "github.com/dghubble/gologin/oauth1"
	gologinTwitter "github.com/dghubble/gologin/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/dghubble/sessions"
	"golang.org/x/net/context"
)

// Config specifies the configuration of the server application
type Config struct {
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	Port                  string
	Track                 string
}

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

const (
	sessionName            = "tweet-stream"
	sessionSecret          = "very secret session secret"
	sessionUserKey         = "twitterID"
	sessionAccessTokenKey  = "twitterAccessToken"
	sessionAccessSecretKey = "twitterAccessSecret"
)

var (
	config *Config
)

func main() {
	// Environment variables
	config = &Config{
		TwitterConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		TwitterConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		Port:  os.Getenv("TWEET_STREAM_SERVER_PORT"),
		Track: os.Getenv("TWEET_STREAM_TRACK"),
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

	// Launch pprof server
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	// Start server
	log.Printf("Starting server listening on %s\n", config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", config.Port), New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// New returns a new ServeMux with app routes.
func New(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/app", requireLogin(http.StripPrefix("/app", http.FileServer(http.Dir("../client/build")))))

	// Easy fix to serve static assets for the web app's index.html
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../client/build/static"))))

	// TODO: Require login for /api/stream
	mux.HandleFunc("/api/stream", streamHandler)

	mux.HandleFunc("/logout", logoutHandler)

	// 1. Register Twitter login and callback handlers
	oauth1Config := &oauth1.Config{
		ConsumerKey:    config.TwitterConsumerKey,
		ConsumerSecret: config.TwitterConsumerSecret,
		CallbackURL:    fmt.Sprintf("http://localhost:%s/twitter/callback", config.Port),
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}
	mux.Handle("/twitter/login", ctxh.NewHandler(gologinTwitter.LoginHandler(oauth1Config, nil)))
	mux.Handle("/twitter/callback", ctxh.NewHandler(gologinTwitter.CallbackHandler(oauth1Config, issueSession(), nil)))
	return mux
}

// streamHandler streams Twitter updates in realtime to the client via SSE
func streamHandler(rw http.ResponseWriter, req *http.Request) {
	// Get session values
	session, err := sessionStore.Get(req, sessionName)
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	// TODO: Remove unsafe cast and propagate these through the context instead
	accessToken := session.Values[sessionAccessTokenKey].(string)
	accessSecret := session.Values[sessionAccessSecretKey].(string)

	// Make sure the writer supports flushing
	flusher, ok := rw.(http.Flusher)

	if !ok {
		log.Println("Streaming unsupported!")
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	//rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Listen to connection close
	notify := rw.(http.CloseNotifier).CloseNotify()

	// Create channel to forward messages to the client on
	messageChan := make(chan []byte, 100)

	fmt.Println("Creating Twitter client")

	// Open stream to twitter
	oauthConfig := oauth1.NewConfig(config.TwitterConsumerKey, config.TwitterConsumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	twitterClient := twitter.NewClient(httpClient)

	fmt.Println("Twitter client created")

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if msg, err := json.Marshal(tweet); err != nil {
			log.Println(err.Error())
		} else {
			fmt.Printf("Received tweet: %s\n", tweet.Text)
			messageChan <- msg
		}
	}
	demux.DM = func(dm *twitter.DirectMessage) {
		fmt.Println(dm.SenderID)
	}
	demux.Event = func(event *twitter.Event) {
		fmt.Printf("%#v\n", event)
	}

	fmt.Println("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{config.Track},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := twitterClient.Streams.Filter(filterParams)
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

SSE:
	for {
		select {
		case <-time.After(10 * time.Second):
			fmt.Println("Heartbeat")
			rw.Write([]byte(":ping\n\n"))
			//rw.Write([]byte("data: heartbeat\n\n"))
			flusher.Flush()
		case <-notify:
			break SSE
		case msg := <-messageChan:
			_, err := fmt.Fprintf(rw, "data: %s\n\n", msg)
			if err != nil {
				log.Println(err.Error())
			}
			flusher.Flush()
		}
	}

	fmt.Println("Stopping Stream...")
	stream.Stop()
	fmt.Println("Stream Stopped")
}

// issueSession issues a cookie session after successful Twitter login
func issueSession() ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		accessToken, accessSecret, err := oauth1login.AccessTokenFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		twitterUser, err := gologinTwitter.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Implement a success handler to issue some form of session
		session := sessionStore.New(sessionName)
		session.Values[sessionUserKey] = twitterUser.ID
		session.Values[sessionAccessTokenKey] = accessToken
		session.Values[sessionAccessSecretKey] = accessSecret
		session.Save(w)
		http.Redirect(w, req, "/app", http.StatusFound)
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
		http.Redirect(w, req, "/app", http.StatusFound)
		return
	}
	page, _ := ioutil.ReadFile("index.html")
	fmt.Fprintf(w, string(page))
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
