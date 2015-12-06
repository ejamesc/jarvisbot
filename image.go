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

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/width"

	"github.com/tucnak/telebot"
)

const googleImageApiUrl = "https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&searchType=image&q="
const googleImageSafeApiUrl = "https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&searchType=image&safe=high&q="

const yaoYujianID = 36972523

var shawnTanRE *regexp.Regexp
var shawnRE *regexp.Regexp
var TAN_RE *regexp.Regexp

func (j *JarvisBot) ImageSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/img: Get an image\nHere are some commands to try: \n* pappy dog\n\n\U0001F4A1 You could also use this format for faster results:\n/img pappy dog", so)
		return
	}

	quitRepeat := j.RepeatChatAction(msg, telebot.UploadingPhoto)
	defer func() { quitRepeat <- true }()

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

	searchURL := fmt.Sprintf(googleImageApiUrl, key, cx)
	if msg.Sender.ID == yaoYujianID {
		// @yyjhao loves spamming "Shawn Tan", replace it with his name in queries
		// This will usually return an image of his magnificent face
		rawQuery = dealWithYujian(rawQuery)
		searchURL = fmt.Sprintf(googleImageSafeApiUrl, key, cx)
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	resp, err := http.Get(searchURL + q)
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
		Items []imgResult `json:"items"`
	}{}
	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for image search query '%s': %s", q, err)
		return
	}

	if len(searchRes.Items) > 0 {
		// Randomly select an image
		n := rand.Intn(len(searchRes.Items))
		r := searchRes.Items[n]
		u, err := r.imgUrl()
		j.log.Printf("[%s] img url: %s", time.Now().Format(time.RFC3339), u.String())
		if err != nil {
			j.log.Printf("error generating url based on search result %v: %s", r, err)
			return
		}

		j.sendPhotoFromURL(u, msg)
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
			j.SendMessage(msg.Chat, "My image search returned nothing. \U0001F622", &telebot.SendOptions{ReplyTo: *msg.Message})
		}
	}
}

type imgResult struct {
	URL   string `json:"link"`
	Image struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"image"`
}

func (i *imgResult) imgUrl() (*url.URL, error) {
	return url.Parse(i.URL)
}

// For Yujian
func init() {
	shawnTanRE = regexp.MustCompile("([Ss][Hh][Aa][Ww][Nn]).*([Tt][Aa][Nn])|([Tt][Aa][Nn]).*([Ss][Hh][Aa][Ww][Nn])")
	shawnRE = regexp.MustCompile("[Ss][Hh][Aa][Ww][Nn]")
	TAN_RE = regexp.MustCompile("[Tt][Aa][Nn]")
}

func dealWithYujian(rawQuery string) string {
	// We create a transformer that reduces all latin runes to their canonical form
	t := runes.If(runes.In(unicode.Latin), width.Fold, nil)
	rawQuery, _, _ = transform.String(t, rawQuery)

	if shawnTanRE.MatchString(rawQuery) {
		rawQuery = shawnRE.ReplaceAllLiteralString(rawQuery, "Yujian")
		rawQuery = TAN_RE.ReplaceAllLiteralString(rawQuery, "Yao")
	} else if tq := strings.Replace(rawQuery, " ", "", -1); shawnTanRE.MatchString(tq) {
		rawQuery = shawnRE.ReplaceAllLiteralString(tq, "Yujian")
		rawQuery = TAN_RE.ReplaceAllLiteralString(rawQuery, "Yao")
	}

	return rawQuery
}
