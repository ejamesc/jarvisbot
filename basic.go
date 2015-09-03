package jarvisbot

import "github.com/tucnak/telebot"

func (j *JarvisBot) SayHello(msg *message) {
	j.bot.SendMessage(msg.Chat, "Hello there, "+msg.Sender.FirstName+"!", nil)
}

func (j *JarvisBot) Echo(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyMarkup: telebot.ReplyMarkup{ForceReply: true}}
		j.bot.SendMessage(msg.Chat, "/echo Jarvis Parrot Mode \U0001F426\nWhat do you want me to parrot?\n\n", so)
	}
	response := ""
	for _, s := range msg.Args {
		response = response + s + " "
	}
	j.bot.SendMessage(msg.Chat, response, nil)
}

func (j *JarvisBot) Clear(msg *message) {
	j.bot.SendMessage(msg.Chat, "Lol, sure.\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\nCleared.", nil)
}
