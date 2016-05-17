package jarvisbot

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"

	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

// jokes.go contain joke functions. Not part of jarvisbot's default funcmap

// SendLaugh returns Jon's laugh. Thx, @jhowtan
func (j *JarvisBot) SendLaugh(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingAudio)

	fn, err := j.sendFileWrapper("data/laugh.ogg", "ogg")
	if err != nil {
		j.log.Printf("error sending file: %s", err)
		return
	}
	fn(msg)
}

// NeverForget returns the Barisan Socialis flag. Thx, @shawntan
func (j *JarvisBot) NeverForget(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)

	fn, err := j.sendFileWrapper("data/barisan.jpg", "photo")
	if err != nil {
		j.log.Printf("error sending file: %s", err)
		return
	}
	fn(msg)
}

// Hanar returns a picture of a Hanar. Thx, @shawntan
// Context: 'hanar, hanar, hanar' means 'yeah I get it, stop nagging'
// in Singaporean Hokkien. The picture sent is a picture of a creature in
// Mass Effect called a Hanar.
func (j *JarvisBot) Hanar(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)

	fn, err := j.sendFileWrapper("data/hanar.jpg", "photo")
	if err != nil {
		j.log.Printf("error sending file: %s", err)
		return
	}
	fn(msg)
}

// TellThatTo returns a picture of Kanjiklub
func (j *JarvisBot) TellThatTo(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)

	fn, err := j.sendFileWrapper("data/kanjiklub.jpg", "photo")
	if err != nil {
		j.log.Printf("error sending file: %s", err)
		return
	}
	fn(msg)
}

// Touch allows Jarvis to be touched. Thx, @rahulg
func (j *JarvisBot) Touch(msg *message) {
	messages := []string{
		"Why, thank you!",
		"Stop touching me!",
		"Ooh, that feels *so* good.",
		"\U0001f60a",
		"AAAAAHHHHHHHH!!!!!\n\nOh, frightened me for a moment, there.",
		"Ouch! Watch it!",
		"Noice!\nhttps://www.youtube.com/watch?v=rQnYi3z56RE",
	}
	n := rand.Intn(len(messages))
	j.SendMessage(msg.Chat, messages[n], nil)
}

// sendFileWrapper checks if the file exists and writes it before returning the response function.
func (j *JarvisBot) sendFileWrapper(assetName string, filetype string) (ResponseFunc, error) {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		j.log.Printf("error retrieving pwd: %s", err)
	}

	_, filename := path.Split(assetName)
	filePath := path.Join(pwd, tempDir, filename)
	// Check if file exists, if it doesn't exist, create it
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileData, err := Asset(assetName)
		if err != nil {
			err = fmt.Errorf("error retrieving asset %s: %s", assetName, err)
			return nil, err
		}
		err = ioutil.WriteFile(filePath, fileData, 0775)
		if err != nil {
			err = fmt.Errorf("error creating %s: %s", assetName, err)
			return nil, err
		}
	}

	return func(msg *message) {
		file, err := telebot.NewFile(filePath)
		if err != nil {
			j.log.Printf("error reading %s: %s", filePath, err)
			return
		}

		if filetype == "ogg" {
			j.bot.SendAudio(msg.Chat, &telebot.Audio{File: file, Mime: "audio/ogg"}, nil)
		} else if filetype == "photo" {
			j.bot.SendPhoto(msg.Chat, &telebot.Photo{File: file, Thumbnail: telebot.Thumbnail{File: file}}, nil)
		} else if filetype == "gif" {
			doc := &telebot.Document{File: file, Preview: telebot.Thumbnail{File: file}, Mime: "image/gif"}
			j.bot.SendDocument(msg.Chat, doc, nil)

		}
	}, nil
}

// Returns pictures of keyword.
// Used for joke functions, e.g. /ducks, because I like to say that
func (j *JarvisBot) SendImage(keyword string) ResponseFunc {
	return func(msg *message) {
		msg.Args = []string{keyword}
		j.ImageSearch(msg)
	}
}
