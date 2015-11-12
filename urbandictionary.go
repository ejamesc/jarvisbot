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

const UD_SEARCH_ENDPOINT = "http://api.urbandictionary.com/v0/define?term=%s"

func (j *JarvisBot) UrbanDictSearch(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/urbandict: Does an Urban Dictonary search\nHere are some commands to try: \n* fleek\n* sapiosexual\n\n\U0001F4A1 You could also use this format for faster results:\n/ud fleek", so)
		return
	}

	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}
	rawQuery = strings.TrimSpace(rawQuery)
	q := url.QueryEscape(rawQuery)

	urlString := fmt.Sprintf(UD_SEARCH_ENDPOINT, q)
	resp, err := http.Get(urlString)
	if err != nil {
		j.log.Printf("failure retrieving videos from Urban Dictionary for query '%s': %s", q, err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		j.log.Printf("failure reading json results from Urban Dictionary video search for query '%s': %s", q, err)
		return
	}

	searchRes := struct {
		List []struct {
			Word       string `json:"word"`
			Definition string `json:"definition"`
			Example    string `json:"example"`
		} `json:"list"`
	}{}

	err = json.Unmarshal(jsonBody, &searchRes)
	if err != nil {
		j.log.Printf("failure unmarshalling json for Urban Dictionary search query '%s': %s", q, err)
		return
	}

	resMsg := "%s\n\nDefinition:\n%s\n\nExample:\n%s"
	if len(searchRes.List) > 0 {
		r := searchRes.List[0]
		rm := fmt.Sprintf(resMsg, r.Word, r.Definition, r.Example)
		j.SendMessage(msg.Chat, rm, nil)
	} else {
		j.SendMessage(msg.Chat, "My Urban Dictionary search returned nothing. \U0001F622", nil)
	}
}
