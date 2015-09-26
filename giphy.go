package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

const GIPHY_API_URL = "http://api.giphy.com/v1/gifs/search?q="
const GIPHY_PUBLIC_BETA_KEY = "dc6zaTOxFJmzC"

func (j *JarvisBot) GifSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.bot.SendMessage(msg.Chat, "/gif: Get a gif\nHere are some commands to try: \n* dance dance\n\n\U0001F4A1 You could also use this format for faster results:\n/gif dance dance", so)
		return
	}

	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
	// TODO: Change this to listen on channel for status
	go func() {
		time.Sleep(6 * time.Second)
		j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
	}()

	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	key := ""
	if j.keys["giphy_api_key"] != "" {
		key = j.keys["giphy_api_key"]
	} else {
		key = GIPHY_PUBLIC_BETA_KEY
	}
	resp, err := http.Get(GIPHY_API_URL + q + "&api_key=" + key)
	if err != nil {
		j.log.Printf("failure retrieving images from Giphy for query '%s': %s", q, err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		j.log.Printf("failure reading json results from Giphy image search for query '%s': %s", q, err)
		return
	}

	searchRes := struct {
		Data []gifResult `json:"data"`
	}{}
	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for giphy search query '%s': %s", q, err)
		return
	}

	if len(searchRes.Data) > 0 {
		// Randomly select an image
		n := rand.Intn(len(searchRes.Data))
		r := searchRes.Data[n]
		u, err := r.getUrl()
		j.log.Printf("[%s] gif url: %s", time.Now().Format(time.RFC3339), u.String())
		if err != nil {
			j.log.Printf("error generating url based on search result %v: %s", r, err)
			return
		}

		j.sendPhotoFromURL(u, msg)
	} else {
		j.bot.SendMessage(msg.Chat, "My gif search returned nothing. \U0001F622", nil)
	}
}

type gifResult struct {
	Images struct {
		DownsizedLarge struct {
			Url  string `json:"url"`
			Size string `json:"size"`
		} `json:"downsized_large"`
		Downsized struct {
			Url  string `json:"url"`
			Size string `json:"size"`
		} `json:"downsized"`
		Original struct {
			Url  string `json:"url"`
			Size string `json:"size"`
		} `json:"original"`
	} `json:"images"`
}

// Return the smallest sized gif
func (g *gifResult) getUrl() (*url.URL, error) {
	sizes := map[string]int{}

	if g.Images.Downsized.Size != "" {
		downSizedSize, err := strconv.Atoi(g.Images.Downsized.Size)
		if err == nil {
			sizes["downsized"] = downSizedSize
		}
	}
	if g.Images.DownsizedLarge.Size != "" {
		downSizedLargeSize, err := strconv.Atoi(g.Images.DownsizedLarge.Size)
		if err == nil {
			sizes["downsizedlarge"] = downSizedLargeSize
		}
	}
	if g.Images.Original.Size != "" {
		originalSize, err := strconv.Atoi(g.Images.Original.Size)
		if err == nil {
			sizes["original"] = originalSize
		}
	}

	selectedGIF := ""
	min := 0
	for k, v := range sizes {
		if v < min || min == 0 {
			selectedGIF = k
			min = v
		}
	}

	switch selectedGIF {
	case "downsized":
		return url.Parse(g.Images.Downsized.Url)
	case "downsizedlarge":
		return url.Parse(g.Images.DownsizedLarge.Url)
	case "original":
		return url.Parse(g.Images.Original.Url)
	}

	return nil, fmt.Errorf("something went horribly wrong, couldn't find an image url")
}
