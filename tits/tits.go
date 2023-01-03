package tits

import (
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/cyberimp/dokku-booba/repo"
	"github.com/cyberimp/dokku-booba/spammer"
	"log"
	"time"
)

var (
	client *danbooru.BooruClient
	spam   *spammer.Spammer
	rep    *repo.BoobaRepo
	err    error
)

func init() {
	spam = new(spammer.Spammer)
	spam.Init()

	client, err = danbooru.GetClient()

	if err != nil {
		log.Fatal(err.Error())
	}

	start := time.Now()
	boobas, err := client.GetBooba()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("total %d boobs, took %s !", len(boobas), time.Since(start))

	rep = new(repo.BoobaRepo)
	rep.InitCache(boobas)
}

func PostTits(chatID int) {
	post := new(danbooru.BooruPost)
	post.FileExt = "invalid"

	for post.CheckExt() {
		res, err := rep.GetBooba()
		if err != nil {
			log.Fatal(err)
		}

		post, err = client.GetPost(res)
		if err != nil {
			return
		}
	}

	err = spam.Post(chatID, post)
	if err != nil {
		return
	}
	log.Printf("Posting https://danbooru.donmai.us/posts/%d to chat #%d", post.ID, chatID)
}
