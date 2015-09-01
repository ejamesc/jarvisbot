package jarvisbot

func (j *JarvisBot) SayHello(msg *message) {
	j.bot.SendMessage(msg.Chat, "Hello there, "+msg.Sender.FirstName+"!", nil)
}

func (j *JarvisBot) Echo(msg *message) {
	response := ""
	for _, s := range msg.Args {
		response = response + s + " "
	}
	j.bot.SendMessage(msg.Chat, response, nil)
}
