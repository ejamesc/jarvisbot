package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/ejamesc/jarvisbot"
	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

func main() {
	// Grab current executing directory
	// In most cases it's the folder in which the Go binary is located.
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("error getting executable folder: %s", err)
	}
	configJSON, err := ioutil.ReadFile(path.Join(pwd, "config.json"))
	if err != nil {
		log.Fatalf("error reading config file! Boo: %s", err)
	}

	logger := log.New(os.Stdout, "[jarvis] ", 0)

	jb := jarvisbot.InitJarvis(configJSON, logger)
	defer jb.CloseDB()

	jb.AddFunction("/laugh", jb.SendLaugh)
	jb.AddFunction("/neverforget", jb.NeverForget)
	jb.AddFunction("/touch", jb.Touch)
	jb.AddFunction("/hanar", jb.Hanar)
	jb.AddFunction("/logic", jb.SendLogic)
	jb.AddFunction("/yank", jb.Yank)
	jb.AddFunction("/tellthatto", jb.TellThatTo)
	jb.AddFunction("/kanjiklub", jb.TellThatTo)
	jb.AddFunction("/ducks", jb.SendImage("quack quack motherfucker"))
	jb.AddFunction("/chickens", jb.SendImage("cluck cluck motherfucker"))

	jb.GoSafely(func() {
		logger.Println("Scheduling exchange rate update")
		for {
			time.Sleep(1 * time.Hour)
			jb.RetrieveAndSaveExchangeRates()
			logger.Printf("[%s] exchange rates updated!", time.Now().Format(time.RFC3339))
		}
	})

	messages := make(chan telebot.Message)
	jb.Listen(messages, 1*time.Second)

	for message := range messages {
		jb.Router(message)
	}
}
