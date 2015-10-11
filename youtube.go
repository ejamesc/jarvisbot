package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ejamesc/telebot"
)

const YOUTUBE_SEARCH_ENDPOINT = "https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=4&type=video&q=%s&key=%s"
const YOUTUBE_VIDEO_BASE = "https://www.youtube.com/watch?v="

func (j *JarvisBot) YoutubeSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.bot.SendMessage(msg.Chat, "/youtube: Does a Youtube search\nHere are some commands to try: \n* unbelievable spouse for house\n* okgo\n\n\U0001F4A1 You could also use this format for faster results:\n/yt okgo", so)
		return
	}

	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	key, ok := j.keys["youtube_api_key"]
	if !ok {
		j.log.Printf("[%s] tried to do a video search, but no Youtube api key!", time.Now().Format(time.RFC3339))
		return
	}

	urlString := fmt.Sprintf(YOUTUBE_SEARCH_ENDPOINT, q, key)
	resp, err := http.Get(urlString)
	if err != nil {
		j.log.Printf("failure retrieving videos from Youtube for query '%s': %s", q, err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		j.log.Printf("failure reading json results from Youtube video search for query '%s': %s", q, err)
		return
	}

	searchRes := struct {
		Items []struct {
			Id struct {
				VideoId string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
		} `json:"items"`
	}{}

	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for Youtube search query '%s': %s", q, err)
		return
	}

	resMsg := ""
	if len(searchRes.Items) > 0 {
		for _, v := range searchRes.Items {
			resMsg = resMsg + fmt.Sprintf("%s%s - %s\n", YOUTUBE_VIDEO_BASE, v.Id.VideoId, v.Snippet.Title)
		}
		j.bot.SendMessage(msg.Chat, resMsg, nil)
	}
}
