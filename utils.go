package jarvisbot

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

const TEMPDIR = "temp"

// sendPhotoFromURL is a helper function to send a photo from a URL to a chat.
// Photos are temporarily stored in a temp folder in the same directory, and
// are deleted after being sent to Telegram.
func (j *JarvisBot) sendPhotoFromURL(url *url.URL, msg *message) {
	urlPath := strings.Split(url.Path, "/")
	imgName := urlPath[len(urlPath)-1]
	ext := strings.ToLower(path.Ext(imgName))

	// If the URL doesn't end with a valid image filename, stop.
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" && ext != ".gif" {
		j.log.Printf("[%s] invalid image filename: %s", time.Now().Format(time.RFC3339), ext)
		return
	}

	j.bot.SendChatAction(msg.Chat, telebot.UploadingPhoto)
	resp, err := http.Get(url.String())
	if err != nil {
		j.log.Printf("[%s] error retrieving image:\n%s", time.Now().Format(time.RFC3339), err)
		return
	}
	defer resp.Body.Close()

	// Grab current executing directory
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

	imgFilePath := filepath.Join(tmpDirPath, imgName)
	file, err := os.Create(imgFilePath)
	if err != nil {
		j.log.Printf("error creating image file")
		return
	}
	defer func() {
		file.Close()
		os.Remove(imgFilePath)
	}()

	// Use io.Copy to just dump the response body to the file.
	// This supports huge files.
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
		j.bot.SendPhoto(msg.Chat, photo, nil)
	}
}
