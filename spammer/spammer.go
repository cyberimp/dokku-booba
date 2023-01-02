package spammer

import (
	"github.com/cyberimp/dokku-booba/danbooru"
	"github.com/dghubble/sling"
	"os"
)

type (
	PostParams struct {
		Photo     string `url:"photo,omitempty"`
		Video     string `url:"video,omitempty"`
		ParseMode string `url:"parse_mode,omitempty"`
		ChatID    int    `url:"chat_id,omitempty"`
		Caption   string `url:"caption,omitempty"`
	}

	Spammer struct {
		token string
	}
	TGResponse struct {
		Ok bool `json:"ok"`
	}
)

const baseUrl = "https://api.telegram.org"

func (s *Spammer) Init() {
	token := os.Getenv("TG_TOKEN")
	s.token = token
}

func (s *Spammer) Post(chatID int, post *danbooru.BooruPost) error {
	postParams := &PostParams{ParseMode: "Markdown"}
	caption := post.GetMarkdown()
	postParams.Caption = caption
	postParams.ChatID = chatID
	mode := "Photo"
	if post.FileExt == "mp4" || post.FileExt == "gif" {
		postParams.Video = post.FileUrl
		mode = "Video"
	} else {
		postParams.Photo = post.FileUrl
	}

	resp := &TGResponse{}

	_, err := sling.New().Get(baseUrl).Path("/bot" + s.token + "/send" + mode).QueryStruct(postParams).ReceiveSuccess(resp)
	if err != nil {
		return err
	}
	return nil
}
