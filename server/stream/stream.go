package stream

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mrjones/oauth"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	// Check if access token is already present in session cookie.
	// TODO: Extract into middleware
	session, err := store.Get(r, cookieKey)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	token, ok := session.Values["access_token"]
	if !ok {
		log.Println("No access token found in session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, ok := token.(*oauth.AccessToken)
	if !ok {
		log.Println("Unable to cast token to *oauth.AccessToken")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Upgrade to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		log.Println(err.Error())
		return
	}

	defer conn.Close()

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

	messages := make(chan *Message, 100)
	defer close(messages)

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case message, ok := <-messages:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				encodedMessage, err := json.Marshal(message)
				if err != nil {
					log.Println(err.Error())
					return
				}

				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Println(err.Error())
					return
				}

				_, err = w.Write(encodedMessage)
				if err != nil {
					log.Println(err.Error())
					return
				}

				if err := w.Close(); err != nil {
					log.Println(err.Error())
					return
				}
			}
		}
	}()

	// Parse stream
	fmt.Println("### Parsing stream")

	for {
		var part []byte // Part of line
		var prefix bool // Flag. Readln readed only part of line.

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

		// DEBUG

		encodedMessage, err := json.Marshal(message)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Println("### Encoded tweet")
		fmt.Println(string(encodedMessage))

		// END DEBUG

		messages <- message

		fmt.Println("New message received")
		fmt.Printf("%v\n", message)
	}
}
