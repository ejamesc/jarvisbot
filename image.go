package jarvisbot

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

const GOOGLE_IMAGE_API_URL = "http://ajax.googleapis.com/ajax/services/search/images?v=1.0&rsz=5&imgsz=small|medium|large&q="

func (j *JarvisBot) ImageSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.bot.SendMessage(msg.Chat, "/img: Get an image\nHere are some commands to try: \n* pappy dog\n\n\U0001F4A1 You could also use this format for faster results:\n/img pappy dog", so)
		return
	}

	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	resp, err := http.Get(GOOGLE_IMAGE_API_URL + q)
	if err != nil {
		j.log.Printf("failure retrieving images from Google for query '%s': %s", q, err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		j.log.Printf("failure reading json results from Google image search for query '%s': %s", q, err)
		return
	}

	searchRes := struct {
		ResponseData struct {
			Results []imgResult `json:"results"`
		} `json:"responseData"`
	}{}
	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for image search query '%s': %s", q, err)
		return
	}

	if len(searchRes.ResponseData.Results) > 0 {
		// Randomly select an image
		n := rand.Intn(len(searchRes.ResponseData.Results))
		r := searchRes.ResponseData.Results[n]
		u, err := r.imgUrl()
		j.log.Printf("[%s] img url: %s", time.Now().Format(time.RFC3339), u.String())
		if err != nil {
			j.log.Printf("error generating url based on search result %v: %s", r, err)
			return
		}

		j.sendPhotoFromURL(u, msg)
	}
}

type imgResult struct {
	UnescapedURL string `json:"unescapedUrl"`
	URL          string `json:"url"`
	Width        string `json:"width"`
	Height       string `json:"height"`
}

func (i *imgResult) imgUrl() (*url.URL, error) {
	if i.UnescapedURL != "" {
		return url.Parse(i.UnescapedURL)
	} else {
		return url.Parse(i.URL)
	}
}
