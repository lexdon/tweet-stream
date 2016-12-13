// Based on https://github.com/rodreegez/go-signin-with-twitter/blob/master/server.go
// and https://github.com/NOX73/go-twitter-stream-api/blob/master/twitter_api.go

package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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

func init() {
	flag.IntVar(&port, "port", 8080, "port to listen to")

	// Needed to store structs in secure cookie
	gob.Register(&oauth.RequestToken{})
	gob.Register(&oauth.AccessToken{})
}

func main() {
	r := mux.NewRouter()

	// r.Handle("/", http.FileServer(http.Dir("static")))
	r.HandleFunc("/", HomePageHandler)
	r.HandleFunc("/stream", StreamHandler)
	r.HandleFunc("/authorize", AuthorizeHandler)
	r.HandleFunc("/oauth_callback", OauthCallbackHandler)
	server := &http.Server{Handler: r}

	fmt.Println("Listening on port", port)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if nil != err {
		log.Fatalln(err)
	}
	if err := server.Serve(listener); nil != err {
		log.Fatalln(err)
	}
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	// Check if access token is already present in session cookie.
	session, err := store.Get(r, cookieKey)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	token, ok := session.Values["access_token"]
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, ok := token.(*oauth.AccessToken)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// We have an access token and can attempt to create the stream
	// TODO: Get track parameter from query string
	streamEndpoint := "https://stream.twitter.com/1.1/statuses/filter.json?track=porsche"

	req, err := http.NewRequest("GET", streamEndpoint, nil)
	if err != nil {
		log.Print(err)
		// TODO: React correctly to different status codes
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	client, err := c.MakeHttpClient(accessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err, resp)
	}

	defer resp.Body.Close()

	bodyReader := bufio.NewReader(resp.Body)

	fmt.Println("Parsing stream")

	for {
		var part []byte //Part of line
		var prefix bool //Flag. Readln readed only part of line.

		part, prefix, err := bodyReader.ReadLine()
		if err != nil {
			break
		}

		if len(part) == 0 {
			continue
		}

		buffer := append([]byte(nil), part...)

		for prefix && err == nil {
			part, prefix, err = bodyReader.ReadLine()
			buffer = append(buffer, part...)
		}
		if err != nil {
			break
		}

		tweet := &Tweet{
			Body: string(buffer),
		}

		message := &Message{
			Response: resp,
			Tweet:    tweet,
		}

		//ch <- message

		fmt.Println("New message received")
		fmt.Printf("%v\n", message)
	}
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
		http.Redirect(w, r, "/stream", http.StatusFound)
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

	http.Redirect(w, r, "/stream", http.StatusFound)
}
