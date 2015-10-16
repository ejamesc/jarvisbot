package jarvisbot

import (
	"strings"

	"github.com/tucnak/telebot"
)

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
	j.bot.SendMessage(msg.Chat, "Lol, sure."+strings.Repeat("\n", 41)+"Cleared.", nil)
}

// Source returns a link to Jarvis's source code.
func (j *JarvisBot) Source(msg *message) {
	j.bot.SendMessage(msg.Chat, "Touch me: https://github.com/ejamesc/jarvisbot", nil)
}

// Start returns some help text.
func (j *JarvisBot) Start(msg *message) {
	j.bot.SendMessage(msg.Chat, `Hi there! I can help you with the following things:

/img - gets an image
/gif - gets a gif
/google - does a Google search
/xchg - does an exchange rate conversion
/youtube - does a Youtube search
/clear - clears your NSFW images for you
/psi - returns the current PSI numbers
/echo - parrots stuff back at you

Give these commands a try!`, nil)
}

func (j *JarvisBot) Help(msg *message) {
	j.bot.SendMessage(msg.Chat, `Some commands:

/img - gets an image
/gif - gets a gif
/google - does a Google search
/xchg - does an exchange rate conversion
/youtube - does a Youtube search
/clear - clears your NSFW images for you
/psi - returns the current PSI numbers
/echo - parrots stuff back at you
`, nil)
}
