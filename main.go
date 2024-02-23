package main

import (
	"encoding/json"
	"fmt"
	"github.com/cyberimp/dokku-booba/tits"
	_ "github.com/heroku/x/hmetrics/onload"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type (
	tgInfo struct {
		Message struct {
			Text string `json:"text"`
			From struct {
				Username  string `json:"username,omitempty"`
				FirstName string `json:"first_name,omitempty"`
				LastName  string `json:"last_name,omitempty"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				Title     string `json:"title,omitempty"`
				Username  string `json:"username,omitempty"`
				FirstName string `json:"first_name,omitempty"`
				LastName  string `json:"last_name,omitempty"`
			} `json:"chat"`
		} `json:"message"`
	}
	chatData struct {
		Chats int `json:"chats,omitempty"`
		Priv  int `json:"priv,omitempty"`
	}
)

func handle(c chan os.Signal) {
	chat, err := strconv.Atoi(os.Getenv("CHAT_ID"))
	if err != nil {
		log.Fatal(err)
	}
	for {
		<-c
		log.Print("Got SIGUSR1 from worker!")
		tits.PostTits(chat)
	}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	go handle(c)

	remote, err := url.Parse("http://static.tiddies.pics/")
	if err != nil {
		panic(err)
	}

	reverseProxy := http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			fmt.Printf("[reverse proxy server] received request at: %s\n", time.Now())

			// set req Host, URL and Request URI to forward a request to the origin server
			req.Host = remote.Host
			req.URL.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.RequestURI = ""

			// save the response from the origin server
			originServerResponse, err := http.DefaultClient.Do(req)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprint(rw, err)
				return
			}
			rw.Header().Add("Content-Type", originServerResponse.Header.Get("Content-Type"))
			rw.Header().Add("Content-Length", originServerResponse.Header.Get("Content-Length"))
			// return response to the client
			rw.WriteHeader(http.StatusOK)
			_, _ = io.Copy(rw, originServerResponse.Body)
		},
	)

	http.HandleFunc("/", reverseProxy)

	http.HandleFunc(
		"/hi", func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, "Hi")
		},
	)

	http.HandleFunc(
		"/stats.json", func(w http.ResponseWriter, r *http.Request) {
			data := chatData{0, 0}
			data.Chats, data.Priv = tits.GetStats()
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(data)
			if err != nil {
				log.Fatal("Error sending json:", err)
			}
		},
	)

	token := os.Getenv("TG_TOKEN")

	http.HandleFunc(
		"/"+token, func(w http.ResponseWriter, r *http.Request) {
			type Response struct {
				Method string `json:"method"`
				Action string `json:"action"`
				ChatID int    `json:"chat_id"`
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			var m tgInfo
			err := json.NewDecoder(r.Body).Decode(&m)
			if err != nil {
				return
			}

			log.Printf("%+v", m)

			if !strings.HasPrefix(m.Message.Text, "/tits") {
				return
			}

			resp := Response{Method: "sendChatAction", ChatID: m.Message.Chat.ID, Action: "upload_photo"}

			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				log.Print("error encoding response:", err)
			}

			go tits.PostTits(m.Message.Chat.ID)
		},
	)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
