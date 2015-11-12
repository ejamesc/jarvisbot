package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

const SGP_URL = "http://sgp.si/now.json"

func (j *JarvisBot) PSI(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.Typing)

	resp, err := http.Get(SGP_URL)
	if err != nil {
		j.log.Printf("[%s] error retrieving PSI: %s", time.Now().Format(time.RFC3339), err)
		return
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		j.log.Printf("[%s] PSI: problem reading response body: %s", time.Now().Format(time.RFC3339), err)
		return
	}

	psi := struct {
		LastGenerated string     `json:"time"`
		North         psiReading `json:"north"`
		South         psiReading `json:"south"`
		West          psiReading `json:"west"`
		East          psiReading `json:"east"`
		Central       psiReading `json:"central"`
		Overall       struct {
			PM2_5_1h string `json:"pm2_5_1h"`
			PSI_24h  string `json:"psi_24h"`
			PSI_3h   int    `json:"PSI_3h"`
		} `json:"overall"`
	}{}

	err = json.Unmarshal(jsonBody, &psi)
	if err != nil {
		j.log.Printf("failure unmarshalling json for psi %s: %s", string(jsonBody), err)
		return
	}

	if len(msg.Args) == 0 {
		j.SendMessage(msg.Chat, "\U0001F4AD PSI Readings:\n* PSI (3hr): "+strconv.Itoa(psi.Overall.PSI_3h)+"\n* PSI (24hr): "+psi.Overall.PSI_24h+"\n* PM2.5 (1hr): "+psi.Overall.PM2_5_1h+"\n\n\U0001F550 "+psi.LastGenerated, nil)
	} else {
		direction := ""
		for _, v := range msg.Args {
			d := strings.TrimSpace(strings.ToLower(v))
			if d == "north" || d == "south" || d == "west" || d == "east" || d == "central" || d == "n" || d == "s" || d == "w" || d == "e" || d == "c" {
				direction = d
			}
		}

		formatText := "\U0001F4AD PSI Readings (%s):\n* PSI (24hr): %s\n\n* PM2.5 (24hr): %s\n* PM10 (24hr): %s\n\n* SO2 (24hr): %s\n* NO2 (1hr): %s\n* O3 (8hr): %s\n* CO (8hr): %s\n\n\U0001F550 %s"
		msgText := ""
		switch direction {
		case "north", "n":
			msgText = fmt.Sprintf(formatText, "North", strconv.Itoa(psi.North.PSI_24h), strconv.Itoa(psi.North.PM2_5_24h), strconv.Itoa(psi.North.PM10_24h), strconv.Itoa(psi.North.SO2_24), strconv.Itoa(psi.North.No2_1h), strconv.Itoa(psi.North.O3_8h), strconv.FormatFloat(psi.North.Co_8h, 'f', 2, 64), psi.LastGenerated)
		case "south", "s":
			msgText = fmt.Sprintf(formatText, "South", strconv.Itoa(psi.South.PSI_24h), strconv.Itoa(psi.South.PM2_5_24h), strconv.Itoa(psi.South.PM10_24h), strconv.Itoa(psi.South.SO2_24), strconv.Itoa(psi.South.No2_1h), strconv.Itoa(psi.South.O3_8h), strconv.FormatFloat(psi.South.Co_8h, 'f', 2, 64), psi.LastGenerated)
		case "west", "w":
			msgText = fmt.Sprintf(formatText, "West", strconv.Itoa(psi.West.PSI_24h), strconv.Itoa(psi.West.PM2_5_24h), strconv.Itoa(psi.West.PM10_24h), strconv.Itoa(psi.West.SO2_24), strconv.Itoa(psi.West.No2_1h), strconv.Itoa(psi.West.O3_8h), strconv.FormatFloat(psi.West.Co_8h, 'f', 2, 64), psi.LastGenerated)
		case "east", "e":
			msgText = fmt.Sprintf(formatText, "East", strconv.Itoa(psi.East.PSI_24h), strconv.Itoa(psi.East.PM2_5_24h), strconv.Itoa(psi.East.PM10_24h), strconv.Itoa(psi.East.SO2_24), strconv.Itoa(psi.East.No2_1h), strconv.Itoa(psi.East.O3_8h), strconv.FormatFloat(psi.East.Co_8h, 'f', 2, 64), psi.LastGenerated)
		case "central", "c":
			msgText = fmt.Sprintf(formatText, "Central", strconv.Itoa(psi.Central.PSI_24h), strconv.Itoa(psi.Central.PM2_5_24h), strconv.Itoa(psi.Central.PM10_24h), strconv.Itoa(psi.Central.SO2_24), strconv.Itoa(psi.Central.No2_1h), strconv.Itoa(psi.Central.O3_8h), strconv.FormatFloat(psi.Central.Co_8h, 'f', 2, 64), psi.LastGenerated)
		default:
			msgText = "I didn't understand that. Try:\n* /psi east\n* /psi central\n* /psi c"
		}
		j.SendMessage(msg.Chat, msgText, nil)
	}

}

type psiReading struct {
	PSI_24h int `json:"psi_24h"`

	PM2_5_24h int `json:"pm2_5_24h"`
	PM10_24h  int `json:"pm10_24h"`

	SO2_24 int     `json:"so2_24h"`
	No2_1h int     `json:"no2_1h"`
	O3_8h  int     `json:"o3_8h"`
	Co_8h  float64 `json:"co_8h"`
}
