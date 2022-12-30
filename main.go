package main

import (
	"dokku-booba/danbooru"
	"log"
	"net/http"
	"os"

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

	boobas, err := client.GetBooba()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print(boobas)

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.Run(":" + port)
}
