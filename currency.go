package jarvisbot

import (
	"strconv"
	"strings"
)

var ENDPOINT = "https://openexchangerates.org/api/latest.json?app_id="

func (j *JarvisBot) Exchange(msg *message) {
	amount, fromCurr, toCurr := parseArgs(msg.Args)
	if amount == 0.0 || fromCurr == "" || toCurr == "" {
		j.bot.SendMessage(msg.Chat, "I didn't understand that. Some sample commands that work include: \n/xchg 10 sgd in usd\n/xchg 100 vnd to sgd\n/xchg 21 usd how much arr?", nil)
	}
}

func parseArgs(args []string) (amount float64, fromCurr, toCurr string) {
	amount = 0.0
	fromCurr, toCurr = "", ""
	for _, a := range args {
		if strings.ToLower(a) == "usd" {
			if fromCurr != "" {
				fromCurr = a
			} else if toCurr != "" {
				toCurr = a
			}
		} else if f, err := strconv.ParseFloat(a, 64); err != nil {
			amount = f
		}
	}
	if toCurr == "" && fromCurr != "" {
		toCurr = "SGD"
	}
	return amount, fromCurr, toCurr
}
