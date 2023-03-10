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

	retry := true

	for retry {
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

		log.Printf("Posting https://danbooru.donmai.us/posts/%d to chat #%d...", post.ID, chatID)
		err = spam.Post(chatID, post)

		if err == nil || strings.Contains(err.Error(), "not enough rights") {
			retry = false
		}

		if err != nil {
			log.Print("...Failed! Error:", err)
		}

		err = nil

		_, err = conn.Exec(context.Background(), "INSERT INTO antibayan (chat_id, post_id) VALUES($1, $2)", chatID, post.ID)
		if err != nil {
			log.Fatal("error adding post: ", err)
		}

		posts = append(posts, post.ID)

		if (len(posts) > 100 && chatID != magicChat) || len(posts) > 1000 {
			id := new(int)
			row := conn.QueryRow(context.Background(), "SELECT id FROM antibayan WHERE chat_id = $1 ORDER BY id LIMIT 1", chatID)
			err := row.Scan(id)
			if err != nil {
				log.Fatal("error finding first post: ", err)
			}
			_, err = conn.Exec(context.Background(), "DELETE FROM antibayan WHERE id = $1", id)
			if err != nil {
				log.Fatal("error deleting post: ", err)
			}
		}
	}
	log.Printf("Posting this took %s !", time.Since(start))
}

func checkBayan(id int, posts []int) bool {
	for _, post := range posts {
		if post == id {
			return true
		}
	}
	return false
}

func GetStats() (int, int) {
	var (
		chats int
		priv  int
	)
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

	row := conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM (SELECT DISTINCT chat_id FROM antibayan) AS temp")
	err = row.Scan(&chats)
	if err != nil {
		log.Fatal("Error counting chats:", err)
		return 0, 0
	}

	row = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM (SELECT DISTINCT chat_id FROM antibayan) AS temp WHERE temp.chat_id < 0")
	err = row.Scan(&priv)
	if err != nil {
		log.Fatal("Error counting chats:", err)
		return 0, 0
	}

	return chats, priv
}
