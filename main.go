package main

import (
	"encoding/json"
	"fmt"
	"github.com/cyberimp/dokku-booba/tits"
	_ "github.com/heroku/x/hmetrics/onload"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type tgInfo struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func handle(c chan os.Signal) {
	for {
		<-c
		log.Print("Got SIGUSR1 from worker!")
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hi")
	})

	token := os.Getenv("TG_TOKEN")

	http.HandleFunc("/"+token, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		var m tgInfo
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			return
		}

		if !strings.HasPrefix(m.Message.Text, "/tits") {
			return
		}

		tits.PostTits(m.Message.Chat.ID)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
