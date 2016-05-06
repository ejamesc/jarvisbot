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

const googleSearchAPI = "https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q="

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

	key, ok := j.keys["custom_search_api_key"]
	if !ok {
		j.log.Printf("error retrieving custom_search_api_key")
		return
	}
	cx, ok := j.keys["custom_search_id"]
	if !ok {
		j.log.Printf("error retrieving custom_search_id")
	}

	searchURL := fmt.Sprintf(googleSearchAPI, key, cx)
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	resp, err := http.Get(searchURL + q)
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
		Items []searchResult `json:"items"`
	}{}
	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for search query '%s': %s", q, err)
		return
	}

	resMsg := ""
	if len(searchRes.Items) > 0 {
		for _, v := range searchRes.Items {
			u, err := v.getUrl()
			if err == nil {
				resMsg = resMsg + fmt.Sprintf("%s - %s\n", u, v.Title)
			} else {
				continue
			}
		}
		j.SendMessage(msg.Chat, resMsg, nil)
	} else {
		var errorRes struct {
			Error struct {
				Code int `json:"code"`
			} `json:"error"`
		}
		err = json.Unmarshal(jsonBody, &errorRes)
		if err == nil && errorRes.Error.Code == 403 {
			j.SendMessage(msg.Chat, "Sorry about this! I've hit my Google Custom Search API limits. \U0001F62D My creator is working on this issue here: https://github.com/ejamesc/jarvisbot/issues/21", &telebot.SendOptions{ReplyTo: *msg.Message})
		} else {
			j.SendMessage(msg.Chat, "My search returned nothing. \U0001F622", &telebot.SendOptions{ReplyTo: *msg.Message})
		}
	}

}

type searchResult struct {
	Title string `json:"title"`
	URL   string `json:"link"`
}

func (s *searchResult) getUrl() (*url.URL, error) {
	return url.Parse(s.URL)
}
