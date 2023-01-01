package main

import (
	"github.com/cyberimp/dokku-booba/danbooru"
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

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.Run(":" + port)
}
