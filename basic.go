package jarvisbot

import "github.com/tucnak/telebot"

// SayHello says hi.
func (j *JarvisBot) SayHello(msg *message) {
	j.bot.SendMessage(msg.Chat, "Hello there, "+msg.Sender.FirstName+"!", nil)
}

// Echo parrots back the argument given by the user.
func (j *JarvisBot) Echo(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.bot.SendMessage(msg.Chat, "/echo Jarvis Parrot Mode \U0001F426\nWhat do you want me to parrot?\n\n", so)
	}
	response := ""
	for _, s := range msg.Args {
		response = response + s + " "
	}
	j.bot.SendMessage(msg.Chat, response, nil)
}

// Clear returns a message that clears out the folder
func (j *JarvisBot) Clear(msg *message) {
	j.bot.SendMessage(msg.Chat, "Lol, sure.\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\nCleared.", nil)
}

// Source returns a link to Jarvis's source code.
func (j *JarvisBot) Source(msg *message) {
	j.bot.SendMessage(msg.Chat, "Touch me: https://github.com/ejamesc/jarvisbot", nil)
}
