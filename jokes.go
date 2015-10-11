package jarvisbot

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/ejamesc/telebot"
	"github.com/kardianos/osext"
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

// sendFileWrapper checks if the file exists and writes it before returning the response function.
func (j *JarvisBot) sendFileWrapper(assetName string, filetype string) (ResponseFunc, error) {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		j.log.Printf("error retrieving pwd: %s", err)
	}

	_, filename := path.Split(assetName)
	filePath := path.Join(pwd, TEMPDIR, filename)
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
			j.bot.SendPhoto(msg.Chat, &telebot.Photo{Thumbnail: telebot.Thumbnail{File: file}}, nil)
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
