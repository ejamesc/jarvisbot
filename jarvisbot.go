package jarvisbot

import (
	"log"
	"os"
	"strings"

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

// Wrapper struct for a message
type message struct {
	Cmd  string
	Args []string
	*telebot.Message
}

func (j *JarvisBot) Router(msg *telebot.Message) {
	jmsg := parseMessage(msg)

	switch {
	case jmsg.Cmd == "/img":
		j.log.Println("IMG")
	case jmsg.Cmd == "/hello":
		j.SayHello(jmsg)
	case jmsg.Cmd == "/gif":
		j.log.Println("GIF")
	case jmsg.Cmd == "/xchg":
		j.Exchange(jmsg)
	}
}

func (j *JarvisBot) SayHello(msg *message) {
	j.bot.SendMessage(msg.Chat, "Hello there, "+msg.Sender.FirstName+"!", nil)
}

func parseMessage(msg *telebot.Message) *message {
	msgTokens := strings.Split(msg.Text, " ")
	cmd, args := msgTokens[0], msgTokens[1:]
	return &message{Cmd: cmd, Args: args, Message: msg}
}
