package jarvisbot

//go:generate go-bindata -pkg $GOPACKAGE -o assets.go data/

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

var exchange_rate_bucket_name = []byte("rates")

// JarvisBot is the main struct. All response funcs bind to this.
type JarvisBot struct {
	Name string // The name of the bot registered with Botfather
	bot  *telebot.Bot
	log  *log.Logger
	fmap FuncMap
	db   *bolt.DB
	keys map[string]string
}

// Wrapper struct for a message
type message struct {
	Cmd  string
	Args []string
	*telebot.Message
}

// A FuncMap is a map of command strings to response functions.
// It is use for routing comamnds to responses.
type FuncMap map[string]ResponseFunc

type ResponseFunc func(m *message)

// Initialise a JarvisBot.
// lg is optional.
func InitJarvis(name string, bot *telebot.Bot, lg *log.Logger, config map[string]string) *JarvisBot {
	// We'll use random numbers throughout JarvisBot
	rand.Seed(time.Now().UTC().UnixNano())

	if lg == nil {
		lg = log.New(os.Stdout, "[jarvis] ", 0)
	}
	j := &JarvisBot{Name: name, bot: bot, log: lg, keys: config}

	j.fmap = j.getDefaultFuncMap()

	// Setup database
	// Get current executing folder
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		lg.Fatalf("cannot retrieve present working directory: %s", err)
	}

	db, err := bolt.Open(path.Join(pwd, "jarvis.db"), 0600, nil)
	if err != nil {
		lg.Fatal(err)
	}
	j.db = db
	createAllBuckets(db)

	// Ensure temp directory is created.
	// This is used to store media temporarily.
	tmpDirPath := filepath.Join(pwd, TEMPDIR)
	if _, err := os.Stat(tmpDirPath); os.IsNotExist(err) {
		j.log.Printf("[%s] creating temporary directory", time.Now().Format(time.RFC3339))
		mkErr := os.Mkdir(tmpDirPath, 0775)
		if mkErr != nil {
			j.log.Printf("[%s] error creating temporary directory\n%s", time.Now().Format(time.RFC3339), err)
		}
	}

	return j
}

// Get the built-in, default FuncMap.
func (j *JarvisBot) getDefaultFuncMap() FuncMap {
	return FuncMap{
		"/hello":  j.SayHello,
		"/echo":   j.Echo,
		"/xchg":   j.Exchange,
		"/x":      j.Exchange,
		"/clear":  j.Clear,
		"/c":      j.Clear,
		"/img":    j.ImageSearch,
		"/psi":    j.PSI,
		"/source": j.Source,
		"/google": j.GoogleSearch,
		"/g":      j.GoogleSearch,
		"/gif":    j.GifSearch,
	}
}

// Add a response function to the FuncMap
func (j *JarvisBot) AddFunction(command string, resp ResponseFunc) error {
	if !strings.Contains(command, "/") {
		return fmt.Errorf("not a valid command string - it should be of the format /something")
	}
	j.fmap[command] = resp
	return nil
}

// Route received Telegram messages to the appropriate response functions.
func (j *JarvisBot) Router(msg *telebot.Message) {
	jmsg := j.parseMessage(msg)
	execFn := j.fmap[jmsg.Cmd]

	if execFn != nil {
		j.GoSafely(func() { execFn(jmsg) })
	}
}

func (j *JarvisBot) CloseDB() {
	j.db.Close()
}

// GoSafely is a utility wrapper to recover and log panics in goroutines.
// If we use naked goroutines, a panic in any one of them crashes
// the whole program. Using GoSafely prevents this.
func (j *JarvisBot) GoSafely(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]

				j.log.Printf("PANIC: %s\n%s", err, stack)
			}
		}()

		fn()
	}()
}

// Ensure all buckets needed by jarvisbot are created.
func createAllBuckets(db *bolt.DB) error {
	// Check all buckets have been created
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(exchange_rate_bucket_name)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return err
}

// Helper to parse incoming messages and return JarvisBot messages
func (j *JarvisBot) parseMessage(msg *telebot.Message) *message {
	cmd := ""
	args := []string{}

	if msg.IsReply() {
		// We use a hack. All reply-to messages have the command it's replying to as the
		// part of the message.
		r := regexp.MustCompile(`\/\w*`)
		res := r.FindString(msg.ReplyTo.Text)
		for k, _ := range j.fmap {
			if res == k {
				cmd = k
				args = strings.Split(msg.Text, " ")
				break
			}
		}
	} else {
		msgTokens := strings.Split(msg.Text, " ")
		cmd, args = strings.ToLower(msgTokens[0]), msgTokens[1:]
		// Deal with commands of the form command@JarvisBot, which appear in
		// group chats.
		if strings.Contains(cmd, "@") {
			c := strings.Split(cmd, "@")
			cmd = c[0]
		}
	}

	return &message{Cmd: cmd, Args: args, Message: msg}
}
