package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tucnak/telebot"
)

const GOOGLE_SEARCH_API = "http://ajax.googleapis.com/ajax/services/search/web?v=1.0&q="

func (j *JarvisBot) GoogleSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/google: Do a Google search\nHere are some commands to try: \n* best chicken rice\n* mee siam mai hum\n\n\U0001F4A1 You could also use this format for faster results:\n/g mee siam mai hum", so)
		return
	}

	j.bot.SendChatAction(msg.Chat, telebot.Typing)
	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	resp, err := http.Get(GOOGLE_SEARCH_API + q)
	if err != nil {
		j.log.Printf("failure retrieving search results from Google for query '%s': %s", q, err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		j.log.Printf("failure reading json results from Google search for query '%s': %s", q, err)
		return
	}

	searchRes := struct {
		ResponseData struct {
			Results []searchResult `json:"results"`
		} `json:"responseData"`
	}{}

	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for search query '%s': %s", q, err)
		return
	}

	resMsg := ""
	if len(searchRes.ResponseData.Results) > 0 {
		for _, v := range searchRes.ResponseData.Results {
			u, err := v.getUrl()
			if err == nil {
				resMsg = resMsg + fmt.Sprintf("%s - %s\n", u, v.Title)
			} else {
				continue
			}
		}
		j.SendMessage(msg.Chat, resMsg, nil)
	}

}

type searchResult struct {
	Title        string `json:"titleNoFormatting"`
	UnescapedURL string `json:"unescapedUrl"`
	URL          string `json:"url"`
}

func (s *searchResult) getUrl() (*url.URL, error) {
	if s.UnescapedURL != "" {
		return url.Parse(s.UnescapedURL)
	} else {
		return url.Parse(s.URL)
	}
}
