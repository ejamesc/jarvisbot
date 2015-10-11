package jarvisbot

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kardianos/osext"
	"github.com/satori/go.uuid"
	"github.com/tucnak/telebot"
)

const TEMPDIR = "temp"

// sendPhotoFromURL is a helper function to send a photo from a URL to a chat.
// Photos are temporarily stored in a temp folder in the same directory, and
// are deleted after being sent to Telegram.
func (j *JarvisBot) sendPhotoFromURL(url *url.URL, msg *message) {
	errSO := &telebot.SendOptions{ReplyTo: *msg.Message}

	urlPath := strings.Split(url.Path, "/")
	imgName := urlPath[len(urlPath)-1]
	ext := strings.ToLower(path.Ext(imgName))

	// If the URL doesn't end with a valid image filename, stop.
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" && ext != ".gif" {
		j.log.Printf("[%s] invalid image filename: %s", time.Now().Format(time.RFC3339), ext)
		j.bot.SendMessage(msg.Chat, "I got an image with an invalid image extension, I'm afraid: "+url.String(), errSO)
		return
	}

	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
	resp, err := http.Get(url.String())
	if err != nil {
		j.log.Printf("[%s] error retrieving image:\n%s", time.Now().Format(time.RFC3339), err)
		j.bot.SendMessage(msg.Chat, "I encountered a problem when retrieving the image: "+url.String(), errSO)

		return
	}
	defer resp.Body.Close()

	// Grab current executing directory.
	// In most cases it's the folder in which the Go binary is located.
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		j.log.Printf("error grabbing pwd \n%s", err)
		return
	}

	// Test if temporary directory exists
	// If it doesn't exist, create it.
	tmpDirPath := filepath.Join(pwd, TEMPDIR)
	if _, err := os.Stat(tmpDirPath); os.IsNotExist(err) {
		j.log.Printf("[%s] creating temporary directory", time.Now().Format(time.RFC3339))
		mkErr := os.Mkdir(tmpDirPath, 0775)
		if mkErr != nil {
			j.log.Printf("[%s] error creating temporary directory\n%s", time.Now().Format(time.RFC3339), err)
			return
		}
	}

	// We generate a random uuid to prevent race conditions
	imgFilePath := filepath.Join(tmpDirPath, uuid.NewV4().String()+ext)
	file, err := os.Create(imgFilePath)
	if err != nil {
		j.log.Printf("error creating image file")
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			j.log.Printf("error closing file: %s", err)
		}
		err = os.Remove(imgFilePath)
		if err != nil {
			j.log.Printf("error removing %s: %s", imgFilePath, err)
		}
	}()

	// io.Copy supports copying large files.
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		j.log.Printf("error writing request body to file: %s", err)
		return
	}

	tFile, err := telebot.NewFile(imgFilePath)
	if err != nil {
		j.log.Printf("error creating new Telebot file: %s", err)
		return
	}

	if ext == ".gif" {
		j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
		doc := &telebot.Document{File: tFile, Preview: telebot.Thumbnail{File: tFile}, Mime: "image/gif"}
		j.bot.SendDocument(msg.Chat, doc, nil)
	} else {
		photo := &telebot.Photo{Thumbnail: telebot.Thumbnail{File: tFile}}
		err := j.bot.SendPhoto(msg.Chat, photo, nil)
		if err != nil {
			j.log.Printf("[%s] error sending picture: %s", time.Now().Format(time.RFC3339), err.Error())
		}
	}
}

// Test function to explore db.
func (j *JarvisBot) Retrieve(msg *message) {
	if j.ratesAreEmpty() {
		j.RetrieveAndSaveExchangeRates()
	}

	err := j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(exchange_rate_bucket_name)
		b.ForEach(func(k, v []byte) error {
			f, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return err
			}
			fmt.Printf("key=%s, value=%s\n", string(k), f)
			return nil
		})
		return nil
	})

	if err != nil {
		j.log.Println(err)
	}
}
