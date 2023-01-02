package main

import (
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/cyberimp/dokku-booba/repo"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
	log.Printf("got id %s, in %s !", res, time.Since(start))

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	err = router.Run(":" + port)
	if err != nil {
		return
	}
}
