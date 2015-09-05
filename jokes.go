package jarvisbot

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

// jokes.go contain joke functions. Not part of jarvisbot's default funcmap

// SendLaugh returns Jon's laugh. Thx, @jhowtan
func (j *JarvisBot) SendLaugh(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingAudio)
	laughData, err := Asset("data/laugh.ogg")
	if err != nil {
		j.log.Printf("error retrieving laughtrack: %s", err)
		return
	}

	fn := j.sendFileWrapper(laughData, "laugh.ogg", "ogg")
	fn(msg)
}

// NeverForget returns the Barisan Socialis flag. Thx, @shawntan
func (j *JarvisBot) NeverForget(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
	barisanData, err := Asset("data/barisan.jpg")
	if err != nil {
		j.log.Printf("error retrieving barisan logo: %s", err)
		return
	}

	fn := j.sendFileWrapper(barisanData, "barisan.jpg", "photo")
	fn(msg)
}

// sendFileWrapper writes the laugh file before returning the response function
func (j *JarvisBot) sendFileWrapper(laughFile []byte, filename string, filetype string) ResponseFunc {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		j.log.Printf("error writing laughFile: %s", err)
	}

	filePath := path.Join(pwd, TEMPDIR, filename)
	// Check if filepath exists, if it doesn't exist, create it
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = ioutil.WriteFile(filePath, laughFile, 0775)
		if err != nil {
			j.log.Printf("error creating %s: %s", filename, err)
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
	}
}
