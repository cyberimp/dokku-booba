package tits

import (
	"context"
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/cyberimp/dokku-booba/repo"
	"github.com/cyberimp/dokku-booba/spammer"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"strconv"
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
	magicChat, err := strconv.Atoi(os.Getenv("CHAT_ID"))
	if err != nil {
		log.Fatal("Error parsing CHAT_ID env:", err)
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}

	defer func(conn *pgx.Conn, ctx context.Context) {
		err = conn.Close(ctx)
		if err != nil {
			log.Fatal("Error closing connection:", err)
		}
	}(conn, context.Background())

	post := new(danbooru.BooruPost)
	post.FileExt = "invalid"
	post.ID = 0

	allPosts, err := conn.Query(context.Background(), "SELECT post_id FROM antibayan WHERE chat_id = $1", chatID)
	var posts []int
	cur := new(int)
	for allPosts.Next() {
		err := allPosts.Scan(cur)
		if err != nil {
			log.Fatal("error scanning posts", err)
		}
		posts = append(posts, *cur)
	}

	for post.BadExt() || checkBayan(post.ID, posts) {
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

	_, err = conn.Exec(context.Background(), "INSERT INTO antibayan (chat_id, post_id) VALUES($1, $2)", chatID, post.ID)
	if err != nil {
		log.Fatal("error adding post", err)
	}

	if (len(posts) > 100 && chatID != magicChat) || len(posts) > 1000 {
		id := new(int)
		row := conn.QueryRow(context.Background(), "SELECT id FROM antibayan WHERE chat_id = $1 LIMIT 1", chatID)
		err := row.Scan(id)
		if err != nil {
			log.Fatal("error finding first post", err)
		}
		_, err = conn.Exec(context.Background(), "DELETE FROM antibayan WHERE id = $1", id)
		if err != nil {
			log.Fatal("error deleting post", err)
		}
	}
}

func checkBayan(id int, posts []int) bool {
	for _, post := range posts {
		if post == id {
			return true
		}
	}
	return false
}
