package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/tucnak/telebot"
)

const GOOGLE_IMAGE_API_URL = "http://ajax.googleapis.com/ajax/services/search/images?v=1.0&rsz=5&imgsz=small|medium|large&q="

const YAO_YUJIAN_ID = 36972523

var SHAWN_TAN_RE *regexp.Regexp
var SHAWN_RE *regexp.Regexp
var TAN_RE *regexp.Regexp

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

	if msg.Sender.ID == YAO_YUJIAN_ID {
		// @yyjhao loves spamming "Shawn Tan", replace it with his name in queries
		// This will usually return an image of his magnificent face
		rq, err := dealWithYujian(rawQuery)
		if err != nil {
			so := &telebot.SendOptions{ReplyTo: *msg.Message}
			j.bot.SendMessage(msg.Chat, "Nope. You're using characters outside of the Latin and East Asian character set. So, Nope.", so)
			return
		}
		rawQuery = rq
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
	} else {
		j.bot.SendMessage(msg.Chat, "My image search returned nothing. \U0001F622", &telebot.SendOptions{ReplyTo: *msg.Message})
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

// For Yujian
func init() {
	SHAWN_TAN_RE = regexp.MustCompile("([Ss][Hh][Aa][Ww][Nn]).*([Tt][Aa][Nn])|([Tt][Aa][Nn]).*([Ss][Hh][Aa][Ww][Nn])")
	SHAWN_RE = regexp.MustCompile("[Ss][Hh][Aa][Ww][Nn]")
	TAN_RE = regexp.MustCompile("[Tt][Aa][Nn]")
}

func dealWithYujian(rawQuery string) (string, error) {
	if isNotValidRange(rawQuery) {
		return "", fmt.Errorf("Yujian tried using an invalid range.")
	}

	if SHAWN_TAN_RE.MatchString(rawQuery) {
		rawQuery = SHAWN_RE.ReplaceAllLiteralString(rawQuery, "Yujian")
		rawQuery = TAN_RE.ReplaceAllLiteralString(rawQuery, "Yao")
	} else if tq := strings.Replace(rawQuery, " ", "", -1); SHAWN_TAN_RE.MatchString(tq) {
		rawQuery = SHAWN_RE.ReplaceAllLiteralString(tq, "Yujian")
		rawQuery = TAN_RE.ReplaceAllLiteralString(rawQuery, "Yao")
	}

	return rawQuery, nil
}

func isNotValidRange(input string) bool {
	for _, c := range input {
		if !unicode.In(c, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Yi, unicode.Hangul, unicode.Bopomofo, unicode.Latin, unicode.Space) {
			return true
		}
	}
	return false
}
