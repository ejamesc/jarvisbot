package jarvisbot

import (
	"log"
	"os"

	"github.com/tucnak/telebot"
)

type JarvisBot struct {
	bot *telebot.Bot
	log *log.Logger
}

func InitJarvis(bot *telebot.Bot, lg *log.Logger) *JarvisBot {
	if lg == nil {
		lg = log.New(os.Stdout, "[jarvis] ", 0)
	}
	return &JarvisBot{bot: bot, log: lg}
}

func (j *JarvisBot) Router(message telebot.Message) {
	parsedMessage := message.Text
}
