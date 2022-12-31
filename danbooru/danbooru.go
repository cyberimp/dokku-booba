package danbooru

import (
	"errors"
	"github.com/dghubble/sling"
	"net/http"
	"os"
)

const (
	magicTags = "solo breasts 1girl -loli score:>50"
	baseUrl   = "https://danbooru.donmai.us/posts.json"
)

type (
	Query struct {
		Page   int    `url:"page,omitempty"`
		Limit  int    `url:"limit,omitempty"`
		Tags   string `url:"tags,omitempty"`
		Login  string `url:"login,omitempty"`
		ApiKey string `url:"api_key,omitempty"`
	}

	BooruClient struct {
		client *http.Client
		login  string
		apiKey string
	}
	id struct {
		Id string `json:"id"`
	}
)

func GetClient() (*BooruClient, error) {
	login := os.Getenv("DANBOORU_LOGIN")
	apiKey := os.Getenv("DANBOORU_API_KEY")
	if apiKey == "" && login != "" {
		return nil, errors.New("empty api key with non-empty login")
	}
	client := BooruClient{client: new(http.Client), login: login, apiKey: apiKey}
	return &client, nil
}

func (c *BooruClient) GetBooba() ([]string, error) {
	var (
		result []string
		idArr  []id
	)
	params := &Query{
		Page:   0,
		Limit:  200,
		Tags:   magicTags,
		Login:  c.login,
		ApiKey: c.apiKey,
	}

	_, err := sling.New().Get(baseUrl).QueryStruct(params).ReceiveSuccess(idArr)
	if err != nil {
		return nil, err
	}
	for _, i := range idArr {
		result = append(result, i.Id)
	}

	return result, nil
}
