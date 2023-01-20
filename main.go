package main

import (
	"encoding/json"
	"fmt"
	"github.com/cyberimp/dokku-booba/tits"
	_ "github.com/heroku/x/hmetrics/onload"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type (
	tgInfo struct {
		Message struct {
			Text string `json:"text"`
			Chat struct {
				ID int `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	}
	chatData struct {
		Chats int
		Priv  int
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

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := chatData{0, 0}

		data.Chats, data.Priv = tits.GetStats()
		tmpl, _ := template.ParseFiles("templates/index.html")
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Fatal("error parsing template:", err)
		}
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hi")
	})

	token := os.Getenv("TG_TOKEN")

	http.HandleFunc("/"+token, func(w http.ResponseWriter, r *http.Request) {
		type Response struct {
			Method string `json:"method"`
			Action string `json:"action"`
			ChatID int    `json:"chat_id"`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		log.Print(r.Body)

		var m tgInfo
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			return
		}

		if !strings.HasPrefix(m.Message.Text, "/tits") {
			return
		}

		resp := Response{Method: "sendChatAction", ChatID: m.Message.Chat.ID, Action: "upload_photo"}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Print("error encoding response:", err)
		}

		go tits.PostTits(m.Message.Chat.ID)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
