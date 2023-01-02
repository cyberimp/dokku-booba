package danbooru

import (
	"errors"
	"github.com/dghubble/sling"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	magicTags = "solo breasts 1girl -loli score:>50"
	baseUrl   = "https://danbooru.donmai.us/"
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

	BooruPost struct {
		FileExt   string `json:"file_ext"`
		Character string `json:"tag_string_character"`
		Artist    string `json:"tag_string_artist"`
		Copyright string `json:"tag_string_copyright"`
		FileUrl   string `json:"large_file_url"`
	}

	id struct {
		Id int `json:"id"`
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

func (c *BooruClient) GetBooba() ([]int, error) {
	var (
		result []int
	)
	params := &Query{
		Page:   0,
		Limit:  200,
		Tags:   magicTags,
		Login:  c.login,
		ApiKey: c.apiKey,
	}

	idArr := new([]id)
	for i := 0; i < 30; i++ {
		params.Page = i
		_, err := sling.New().Get(baseUrl).Path("posts.json").QueryStruct(params).ReceiveSuccess(idArr)
		if err != nil {
			return nil, err
		}
		for _, i := range *idArr {
			result = append(result, i.Id)
		}
	}

	return result, nil
}

func (c *BooruClient) GetPost(id int) (*BooruPost, error) {
	params := &Query{
		Login:  c.login,
		ApiKey: c.apiKey,
	}
	result := new(BooruPost)
	_, err := sling.New().Get(baseUrl).Path("posts/" + strconv.Itoa(id) + ".json").QueryStruct(params).ReceiveSuccess(result)
	return result, err
}

func clearUnderscore(s string) string {
	result := strings.Replace(s, " ", "\n", -1)
	result = strings.Replace(result, "_", " ", -1)
	return result
}
func (p *BooruPost) GetMarkdown() string {
	artist := clearUnderscore(p.Artist)
	copyright := clearUnderscore(p.Copyright)
	character := clearUnderscore(p.Character)
	result := "*Artist:* `" + artist + "`"
	result += "\n*Origin:* `" + copyright + "`"
	if character != "" {
		result += "\n*Character:* `" + character + "`"
	}
	return result
}

func (p *BooruPost) CheckExt() bool {
	if p.FileExt == "gif" || p.FileExt == "jpg" || p.FileExt == "jpeg" || p.FileExt == "png" || p.FileExt == "mp4" {
		return true
	}
	return false
}
