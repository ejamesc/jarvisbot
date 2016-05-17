package jarvisbot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/tucnak/telebot"
)

const rateLimit = 3
const countKey = "count"
const timestampKey = "timestamp"

// CollectPing sends a message requesting all users in a group to respond.
// This allows Jarvis to collect their usernames
func (j *JarvisBot) CollectPing(msg *message) {
	if !msg.Chat.IsGroupChat() {
		j.SendMessage(msg.Chat, "This feature only works in group chats, sorry!", nil)
		return
	}

	if msg.IsReply() {
		so := &telebot.SendOptions{ReplyTo: *msg.Message}

		if msg.Sender.Username == "" {
			j.SendMessage(msg.Chat, "I'm afraid you don't have a username, "+msg.Sender.FirstName+". You should create one if you'd like @ notifications.", so)
		} else {
			if !j.usernameExistsForChat(&msg.Chat, &msg.Sender) {
				err := j.saveUserToDB(&msg.Chat, &msg.Sender)
				if err != nil {
					j.log.Printf("[%s] error saving user to db: %s", time.Now().Format(time.RFC3339), err)
					return
				}
			}
			j.SendMessage(msg.Chat, "Thanks, "+msg.Sender.FirstName+". I've saved @"+msg.Sender.Username+" for use later.", so)
		}
	} else {
		so := &telebot.SendOptions{ReplyMarkup: telebot.ReplyMarkup{ForceReply: true}}
		j.SendMessage(msg.Chat, "/pingsetup: Sets up ping functionality\nPlease reply to this so I can store your username \U0001F60A", so)
	}
}

// Ping returns a list of all usernames in a given chat, along with a message.
// This is used for alerting all members in a group.
func (j *JarvisBot) Ping(msg *message) {
	if !msg.Chat.IsGroupChat() {
		j.SendMessage(msg.Chat, "This feature only works in group chats, sorry!", nil)
		return
	}

	// We don't accept pings from people without usernames.
	if msg.Sender.Username == "" {
		j.SendMessage(msg.Chat, "You can only ping if you have a username yourself, sorry!", nil)
		return
	}

	if !j.groupBucketExists(&msg.Chat) {
		j.SendMessage(msg.Chat, "I don't have any records for this group. Perhaps run /pingsetup first?", nil)
		return
	}

	if j.canSendWithinTimeLimit(&msg.Chat) {
		usernames, err := j.getAllUsernamesForChat(&msg.Chat, &msg.Sender)
		if err != nil {
			j.log.Printf("[%s] error getting all usernames for chat %s: %s", time.Now().Format(time.RFC3339), msg.Chat.Title, err)
			return
		}

		if usernames == "" {
			j.SendMessage(msg.Chat, "I'm afraid I don't currently have any other usernames for this group apart from yours, "+msg.Sender.FirstName+".", &telebot.SendOptions{ReplyTo: *msg.Message})
			return
		}

		message := ""
		for _, v := range msg.Args {
			message = message + v + " "
		}

		// We update the last saved time and hour count.
		j.updateLastSentTime(&msg.Chat)
		j.SendMessage(msg.Chat, usernames+message, nil)
	} else {
		j.SendMessage(msg.Chat, "Rate limit reached! You can only send "+strconv.Itoa(rateLimit)+" pings to a group every hour.", nil)
	}
}

// saveUserToDB saves the given user to the given chat bucket in Bolt.
func (j *JarvisBot) saveUserToDB(chat *telebot.Chat, sender *telebot.User) error {
	groupID, userID, username := strconv.Itoa(int(chat.ID)), strconv.Itoa(sender.ID), sender.Username
	if sender.Username == "" {
		return fmt.Errorf("user has no username!")
	}

	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(group_usernames_bucket_name)
		gb, err := b.CreateBucketIfNotExists([]byte(groupID))
		if err != nil {
			return err
		}

		err = gb.Put([]byte(userID), []byte(username))
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

// saveUsernameSafely checks if a user exists in a given group chat, and saves if he/she
// doesn't currently exist.
// We try as much as possible to avoid opening a Read-Write transaction, because
// Bolt can only have one Read-Write transaction at any time.
func (j *JarvisBot) saveUsernameSafely(chat *telebot.Chat, sender *telebot.User) {
	if sender.Username == "" || !chat.IsGroupChat() {
		return
	}

	if !j.usernameExistsForChat(chat, sender) {
		err := j.saveUserToDB(chat, sender)
		if err != nil {
			j.log.Printf("[%s] error saving user %s to group %s: %s", time.Now().Format(time.RFC3339), sender.Username, chat.Title, err)
		}
	}
}

// usernameExistsForChat checks if a username exists for a given chat.
// This returns false if the userID exists but the username has changed.
func (j *JarvisBot) usernameExistsForChat(chat *telebot.Chat, sender *telebot.User) bool {
	res := false
	// Guard against possibility that chat isn't a group chat.
	// We're lazy here, returning true implies that we don't want to run a save.
	if !chat.IsGroupChat() {
		return true
	}

	groupID, senderID, username := strconv.Itoa(int(chat.ID)), strconv.Itoa(sender.ID), sender.Username

	j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(group_usernames_bucket_name)
		gb := b.Bucket([]byte(groupID))
		if gb == nil {
			return nil
		}

		v := gb.Get([]byte(senderID))
		if v != nil && string(v) == username {
			res = true
			return nil
		}

		return nil
	})

	return res
}

// getAllUsernamesForChat returns all usernames for a chat, sans the sender's.
func (j *JarvisBot) getAllUsernamesForChat(chat *telebot.Chat, sender *telebot.User) (string, error) {
	uArray, groupID, senderID := []string{}, strconv.Itoa(int(chat.ID)), strconv.Itoa(sender.ID)

	if !chat.IsGroupChat() {
		return "", fmt.Errorf("chat %s is not a group chat!", groupID)
	}

	err := j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(group_usernames_bucket_name)
		gb := b.Bucket([]byte(groupID))
		if gb == nil {
			return fmt.Errorf("error retrieving bucket for group ID %s", groupID)
		}

		gb.ForEach(func(k, v []byte) error {
			key, curr := string(k), string(v)
			if key != senderID && key != timestampKey && key != countKey {
				uArray = append(uArray, curr)
			}
			return nil
		})
		return nil
	})

	if err != nil {
		return "", err
	}

	res := ""
	for _, v := range uArray {
		res = res + "@" + v + " "
	}
	return res, nil
}

// Check if a bucket for the group chat already exists.
func (j *JarvisBot) groupBucketExists(chat *telebot.Chat) bool {
	res, groupID := false, strconv.Itoa(int(chat.ID))

	j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(group_usernames_bucket_name)
		gb := b.Bucket([]byte(groupID))

		if gb != nil {
			res = true
		}
		return nil
	})
	return res
}

// canSendWithinTimeLimit checks the time limit for the given group chat.
// We limit pings per group chat to rateLimit every hour.
func (j *JarvisBot) canSendWithinTimeLimit(chat *telebot.Chat) bool {
	res := false
	groupID := strconv.Itoa(int(chat.ID))

	err := j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(group_usernames_bucket_name)
		gb := b.Bucket([]byte(groupID))

		if gb == nil {
			return fmt.Errorf("invariant error - time limit checked for %s when no group bucket exists", groupID)
		}

		timeString := gb.Get([]byte(timestampKey))
		if timeString == nil {
			res = true
			return nil
		}

		lastTime, err := time.Parse(time.RFC3339, string(timeString))
		if err != nil {
			return fmt.Errorf("problem parsing timestring for group ID %s: %s", groupID, err)
		}

		if time.Since(lastTime) > time.Hour {
			res = true
		} else {
			if count := gb.Get([]byte(countKey)); count != nil {
				numCount, err := strconv.Atoi(string(count))
				if err != nil {
					return err
				}
				res = (numCount < rateLimit)
			}
		}
		return nil
	})

	if err != nil {
		j.log.Printf("[%s] error grabbing last-send-time from db: %s", time.Now().Format(time.RFC3339), err)
	}

	return res
}

// updateLastSentTime updates the last sent time for a given chat.
func (j *JarvisBot) updateLastSentTime(chat *telebot.Chat) error {
	if !chat.IsGroupChat() {
		return fmt.Errorf("chat is not a group chat!")
	}

	groupID := strconv.Itoa(int(chat.ID))
	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(group_usernames_bucket_name)
		gb := b.Bucket([]byte(groupID))
		if gb == nil {
			return fmt.Errorf("invariant error - group bucket does not exist, yet attempt to update timestamp")
		}

		timeString := gb.Get([]byte(timestampKey))
		if timeString == nil {
			return resetTime(gb)
		} else {
			lastTime, err := time.Parse(time.RFC3339, string(timeString))
			if err != nil {
				return err
			}
			if time.Since(lastTime) > time.Hour {
				return resetTime(gb)
			} else {
				countBytes := gb.Get([]byte(countKey))
				if countBytes == nil {
					return fmt.Errorf("invariant error - timestamp was saved but no count exists")
				}
				lastCount, err := strconv.Atoi(string(countBytes))
				if err != nil {
					return err
				}
				newCount := strconv.Itoa(lastCount + 1)
				err = gb.Put([]byte(countKey), []byte(newCount))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

// Helpers
func resetTime(b *bolt.Bucket) error {
	err := b.Put([]byte(timestampKey), []byte(time.Now().Format(time.RFC3339)))
	if err != nil {
		return err
	}
	err = b.Put([]byte(countKey), []byte(strconv.Itoa(1)))
	if err != nil {
		return err
	}
	return nil
}
