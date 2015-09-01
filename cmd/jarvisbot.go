package main

import (
	"log"
	"os"
	"time"

	"github.com/ejamesc/jarvisbot"
	"github.com/tucnak/telebot"
)

func main() {
	bot, err := telebot.NewBot(API_KEY)
	if err != nil {
		log.Fatalf("Error creating new bot, dude %s", err)
	}

	logger := log.New(os.Stdout, "[jarvis] ", 0)
	jb := jarvisbot.InitJarvis(bot, logger, nil)

	messages := make(chan telebot.Message)
	bot.Listen(messages, 1*time.Second)

	for message := range messages {
		jb.Router(&message)
	}
}
