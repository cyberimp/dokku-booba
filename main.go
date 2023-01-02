package main

import (
	"fmt"
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/cyberimp/dokku-booba/repo"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

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
	r := new(repo.BoobaRepo)
	r.InitCache(boobas)
	log.Printf("saving cache took %s !", time.Since(start))

	start = time.Now()
	res, err := r.GetBooba()
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

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
