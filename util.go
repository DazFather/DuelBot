package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/NicoNex/echotron/v3"
)

// Returns the text contained in the given update.
func extractText(update *echotron.Update) string {
	if update.Message != nil {
		if update.Message.Text != "" {
			return update.Message.Text
		}
		return update.Message.Caption
	} else if update.EditedMessage != nil {
		return update.EditedMessage.Text
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Data
	}

	return ""
}

// Return the /command and the payload (other element separated by ' ' or '_' if /start)
func extractCommand(update *echotron.Update) (command string, payload []string) {
	var (
		text = extractText(update)
		ind  = strings.IndexRune(text, ' ')
	)

	if ind == -1 {
		return text, nil
	}

	command = text[:ind]
	payload = append(payload, text[ind+1:])
	if strings.ContainsRune(text[ind+1:], ' ') {
		payload = strings.Split(text[ind+1:], " ")
	} else if command == "/start" && strings.ContainsRune(text[ind+1:], '_') {
		payload = strings.Split(text[ind+1:], "_")
	}
	return
}

func UpdateStatus(userID int64, text string, newMessage bool) (err error) {
	var (
		b   = echotron.NewAPI(TOKEN)
		res echotron.APIResponseMessage
	)

	if !newMessage {
		messageID := echotron.NewMessageID(userID, GetPlayerMenuID(userID))
		_, err = b.EditMessageText(text, messageID, &echotron.MessageTextOptions{
			ParseMode:   echotron.HTML,
			ReplyMarkup: genActionKbd(),
		})
	}

	if newMessage || err != nil {
		res, err = b.SendMessage(text, userID, &echotron.MessageOptions{
			ParseMode:   echotron.HTML,
			BaseOptions: echotron.BaseOptions{ReplyMarkup: genActionKbd()},
		})
		if err != nil || res.Result == nil {
			log.Println("UpdateStatus", err)
			return
		}

		SetPlayerMenuID(userID, res.Result.ID)
	}

	return
}

func ParseName(rawName string) string {
	var rx = regexp.MustCompile(`[\*\[\]\(\)\` + "`" + `~>#+\-=|{}.!]`)

	return rx.ReplaceAllString(rawName, "\\$0")
}

func genUserLink(userID int64) string {
	return fmt.Sprint("<a href=\"tg://user?id=", userID, "\">", GetPlayerName(userID), "</a>")
}
