package main

import (
	"errors"
	"fmt"
	"log"
	"os"
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

// Update the menuID with the status of the player
func UpdateStatus(userID int64, text string, newMessage bool) (err error) {
	var (
		menuID int
		b      = echotron.NewAPI(TOKEN)
		res    echotron.APIResponseMessage
	)

	if !newMessage {
		menuID, err = GetPlayerMenuID(userID)
		if err != nil {
			log.Println("UpdateStatus", "GetPlayerMenuID", err)
			return
		}
		messageID := echotron.NewMessageID(userID, menuID)
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

// Vreating an HTML link to the specified user
func genUserLink(userID int64, name string) string {
	return fmt.Sprint("<a href=\"tg://user?id=", userID, "\">", name, "</a>")
}

// It try to detect id a string is a vaild token
func validateToken(token string) error {
	match, err := regexp.MatchString(`\d+:[\w\-]+`, token)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("Wrong format for TOKEN value")
	}

	return nil
}

/* Load the token from using the command line arguments (os.Args)
 * put the token next to the executable file name on the console or
 * use put readfrom followed by the file in witch there is the token (ex. .\DuelBot.exe readfrom mytoken.txt)
 */
func LoadToken() (token string, err error) {
	switch len(os.Args) {
	case 1:
		return "", errors.New("Missing TOKEN value")

	case 2:
		token = os.Args[1]

	case 3:
		if strings.ToUpper(os.Args[1]) != "--READFROM" {
			return "", errors.New("Invalid format")
		}
		if content, err := os.ReadFile(os.Args[2]); err != nil {
			return "", err
		} else {
			token = strings.TrimSpace(string(content))
		}

	default:
		return "", errors.New("Too many arguments")
	}

	err = validateToken(token)
	return
}

// Get the name of a player
func (b *bot) GetPlayerName(chatID int64) (name string) {
	res, err := b.GetChat(chatID)
	if err != nil || res.Result == nil {
		return "Unknown User"
	}
	name = ParseName(res.Result.FirstName)
	if name == "" {
		name = "Unknown User"
	}

	return
}
