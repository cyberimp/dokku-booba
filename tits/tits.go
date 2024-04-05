package tits

import (
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/cyberimp/dokku-booba/repo"
	"github.com/cyberimp/dokku-booba/spammer"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	client *danbooru.BooruClient
	spam   *spammer.Spammer
	rep    *repo.BoobaRepo
	err    error
)

func init() {
	start := time.Now()
	spam = new(spammer.Spammer)
	spam.Init()

	client, err = danbooru.GetClient()

	if err != nil {
		log.Fatal(err.Error())
	}

	boobas, err := client.GetBooba()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("total %d boobs, took %s !", len(boobas), time.Since(start))

	rep = new(repo.BoobaRepo)
	rep.InitCache(boobas)
}

func PostTits(chatID int) {
	start := time.Now()

	rep.IncViews()
	rep.AddChat(chatID)

	magicChat, err := strconv.Atoi(os.Getenv("CHAT_ID"))
	if err != nil {
		log.Fatal("Error parsing CHAT_ID env:", err)
	}

	post := new(danbooru.BooruPost)
	post.FileExt = "invalid"
	post.ID = 0

	retry := true

	for retry {
		for post.BadExt() || rep.CheckBayan(chatID, post.ID) {
			res, err := rep.GetBooba()
			if err != nil {
				log.Fatal(err)
			}

			post, err = client.GetPost(res)
			if err != nil {
				return
			}
		}

		log.Printf("Posting https://danbooru.donmai.us/posts/%d to chat #%d...", post.ID, chatID)
		err = spam.Post(chatID, post)

		if err == nil || strings.Contains(err.Error(), "not enough rights") {
			retry = false
		}

		if err != nil {
			log.Print("...Failed! Error:", err)
		}

		err = nil

		trim := 100
		if chatID == magicChat {
			trim *= 10
		}

		rep.AddBayan(chatID, post.ID, trim)

	}
	log.Printf("Posting this took %s !", time.Since(start))
}

func GetStats() (int, int) {
	return rep.GetStats()
}

func GetRequests() []repo.ReqData {
	return rep.GetRequests()
}
