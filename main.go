package main

import (
	"encoding/json"
	"fmt"
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/cyberimp/dokku-booba/repo"
	"github.com/cyberimp/dokku-booba/spammer"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
)

type tgInfo struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			Id int `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	spam := new(spammer.Spammer)
	spam.Init()

	client, err := danbooru.GetClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	start := time.Now()

	boobas, err := client.GetBooba()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print(boobas)
	log.Printf("total %d boobs, took %s !", len(boobas), time.Since(start))

	start = time.Now()
	rep := new(repo.BoobaRepo)
	rep.InitCache(boobas)
	log.Printf("saving cache took %s !", time.Since(start))

	start = time.Now()
	res, err := rep.GetBooba()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("got id %d, in %s !", res, time.Since(start))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hi")
	})

	token := os.Getenv("TG_TOKEN")

	http.HandleFunc("/"+token, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		res, err := rep.GetBooba()
		if err != nil {
			log.Fatal(err)
		}

		var m tgInfo
		err = json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			return
		}
		post, err := client.GetPost(res)
		if err != nil {
			return
		}
		err = spam.Post(m.Message.Chat.Id, post)
		if err != nil {
			return
		}
		log.Print(res)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
