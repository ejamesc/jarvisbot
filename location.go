package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

const placeSearchAPI = "https://maps.googleapis.com/maps/api/place/textsearch/json?query=%s&key=%s"

// LocationSearch takes a search string and returns a location.
func (j *JarvisBot) LocationSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/loc: Does a location search\nHere are some commands to try: \n* serangoon gardens\n* rail mall singapore\n\n\U0001F4A1 You could also use this format for faster results:\n/loc chomp chomp food", so)
		return
	}

	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	key, ok := j.keys["maps_api_key"]
	if !ok {
		j.log.Printf("[%s] tried to do a location search, but no Google api key!", time.Now().Format(time.RFC3339))
		return
	}

	urlString := fmt.Sprintf(placeSearchAPI, q, key)
	resp, err := http.Get(urlString)
	if err != nil {
		j.log.Printf("failure retrieving videos from Google Place Search for query '%s': %s", q, err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		j.log.Printf("failure reading json results from Google Place Search for query '%s': %s", q, err)
		return
	}

	searchRes := struct {
		Results []struct {
			Geometry struct {
				Location struct {
					Lat float32 `json:"lat"`
					Lng float32 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
	}{}

	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for Google Place Search query '%s': %s", q, err)
		return
	}

	if len(searchRes.Results) > 0 {
		n := rand.Intn(len(searchRes.Results))
		r := searchRes.Results[n]
		loc := &telebot.Location{Longitude: r.Geometry.Location.Lng, Latitude: r.Geometry.Location.Lat}

		j.bot.SendLocation(msg.Chat, loc, nil)
	}

}
