package jarvisbot

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"

	"github.com/boltdb/bolt"
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

// SendLogic returns Tuan's logical meme, based on when Ian called him
// illogical.
func (j *JarvisBot) SendLogic(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)

	fn, err := j.sendFileWrapper("data/logic.jpg", "photo")
	if err != nil {
		j.log.Printf("error sending file: %s", err)
		return
	}
	fn(msg)
}

// Yank returns a gif of @jellykaya being yanked off-stage by @vishnup
func (j *JarvisBot) Yank(msg *message) {
	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)

	fn, err := j.sendFileWrapper("data/yank.gif", "gif")
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

	fileId, err := j.getCachedFileID(assetName)
	if err != nil {
		j.log.Printf("error retreiving cached file_id for %s", assetName)
	}

	if fileId != "" {
		file := telebot.File{FileID: fileId}

		return func(msg *message) {
			if filetype == "ogg" {
				audio := telebot.Audio{File: file, Mime: "audio/ogg"}
				j.bot.SendAudio(msg.Chat, &audio, nil)
			} else if filetype == "photo" {
				photo := telebot.Photo{File: file}
				j.bot.SendPhoto(msg.Chat, &photo, nil)
			} else if filetype == "gif" {
				doc := telebot.Document{File: file, Mime: "image/gif"}
				j.bot.SendDocument(msg.Chat, &doc, nil)
			}
		}, nil
	} else {

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

			var newFileId string
			if filetype == "ogg" {
				audio := telebot.Audio{File: file, Mime: "audio/ogg"}
				j.bot.SendAudio(msg.Chat, &audio, nil)
				newFileId = audio.FileID
			} else if filetype == "photo" {
				photo := telebot.Photo{File: file, Thumbnail: telebot.Thumbnail{File: file}}
				j.bot.SendPhoto(msg.Chat, &photo, nil)
				newFileId = photo.FileID
			} else if filetype == "gif" {
				doc := telebot.Document{File: file, Preview: telebot.Thumbnail{File: file}, Mime: "image/gif"}
				j.bot.SendDocument(msg.Chat, &doc, nil)
				newFileId = doc.FileID
			}
			err = j.cacheFileID(assetName, newFileId)
			if err != nil {
				j.log.Println("error caching file_id '%s' for '%s': %s", fileId, assetName, err)
			}
		}, nil

	}

}

func (j *JarvisBot) cacheFileID(assetName string, fileId string) error {
	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(file_cache_bucket_name)
		err := b.Put([]byte(assetName), []byte(fileId))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (j *JarvisBot) getCachedFileID(assetName string) (string, error) {
	var fileId string

	err := j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(file_cache_bucket_name)
		v := b.Get([]byte(assetName))
		fileId = string(v[:])
		return nil
	})
	if err != nil {
		return "", err
	}

	return fileId, nil

}

// Returns pictures of keyword.
// Used for joke functions, e.g. /ducks, because I like to say that
func (j *JarvisBot) SendImage(keyword string) ResponseFunc {
	return func(msg *message) {
		msg.Args = []string{keyword}
		j.ImageSearch(msg)
	}
}
