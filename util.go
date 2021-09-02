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
	switch true {
	case update.Message != nil:
		if update.Message.Text != "" {
			return update.Message.Text
		}
		return update.Message.Caption
	case update.EditedMessage != nil:
		return update.EditedMessage.Text
	case update.ChannelPost != nil:
		return update.ChannelPost.Text
	case update.EditedChannelPost != nil:
		return update.EditedChannelPost.Text
	case update.CallbackQuery != nil:
		return update.CallbackQuery.Data
	}

	return ""
}

// Get the Message.ID of an update (that is not inline-based)
func extractMessageID(update *echotron.Update) int {
	switch true {
	case update.Message != nil:
		return update.Message.ID
	case update.EditedMessage != nil:
		return update.EditedMessage.ID
	case update.ChannelPost != nil:
		return update.ChannelPost.ID
	case update.EditedChannelPost != nil:
		return update.EditedChannelPost.ID
	case update.CallbackQuery != nil && update.CallbackQuery.Message != nil:
		return update.CallbackQuery.Message.ID
	}

	return -1
}

// Generate and return the MessageIDOptions of a given update using the ID and SenderChat
func extractMessageIDOpt(update *echotron.Update) *echotron.MessageIDOptions {
	var (
		message *echotron.Message
		msgID   echotron.MessageIDOptions
		userID  int64
	)

	switch true {
	case update.Message != nil:
		message = update.Message
		userID = message.From.ID
	case update.EditedMessage != nil:
		message = update.EditedMessage
		userID = message.From.ID
	case update.ChannelPost != nil:
		message = update.ChannelPost
		userID = message.SenderChat.ID
	case update.EditedChannelPost != nil:
		message = update.EditedChannelPost
		userID = message.SenderChat.ID
	case update.InlineQuery != nil:
		msgID = echotron.NewInlineMessageID(update.InlineQuery.ID)
		return &msgID
	case update.ChosenInlineResult != nil:
		msgID = echotron.NewInlineMessageID(update.ChosenInlineResult.ResultID)
		return &msgID
	case update.CallbackQuery != nil:
		message = update.CallbackQuery.Message
		if message == nil {
			msgID = echotron.NewInlineMessageID(update.CallbackQuery.ID)
			return &msgID
		}
		userID = update.CallbackQuery.From.ID
	}

	if message == nil {
		return nil
	}
	msgID = echotron.NewMessageID(userID, message.ID)
	return &msgID
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

// Return the parsed FirstName of the user who sent the message
func extractName(update *echotron.Update) (FirstName string) {
	var user *echotron.User

	switch true {
	case update.Message != nil:
		user = update.Message.From
	case update.EditedMessage != nil:
		user = update.EditedMessage.From
	case update.ChannelPost != nil:
		user = update.ChannelPost.From
	case update.EditedChannelPost != nil:
		user = update.EditedChannelPost.From
	case update.InlineQuery != nil:
		user = update.InlineQuery.From
	case update.ChosenInlineResult != nil:
		user = update.ChosenInlineResult.From
	case update.CallbackQuery != nil:
		user = update.CallbackQuery.From
	}

	if user == nil {
		return "Unknown User"
	}
	FirstName = parseName(user.FirstName)
	if FirstName == "" {
		return "Unnamed User"
	}

	return
}

// Update the menuID with the status of the player
func UpdateStatus(userID int64, text string, newMessage bool) (err error) {
	var (
		menuID int
		b      = echotron.NewAPI(TOKEN)
		res    echotron.APIResponseMessage
		move   string
	)

	move, err = GetPlayerAction(userID)
	if err != nil {
		log.Println("UpdateStatus", "GetPlayerAction", err)
		return
	}

	if !newMessage {
		menuID, err = GetPlayerMenuID(userID)
		if err != nil {
			log.Println("UpdateStatus", "GetPlayerMenuID", err)
			return
		}
		messageID := echotron.NewMessageID(userID, menuID)
		_, err = b.EditMessageText(text, messageID, &echotron.MessageTextOptions{
			ParseMode:   echotron.HTML,
			ReplyMarkup: genActionKbd(move),
		})
	}

	if newMessage || err != nil {
		res, err = b.SendMessage(text, userID, &echotron.MessageOptions{
			ParseMode:   echotron.HTML,
			BaseOptions: echotron.BaseOptions{ReplyMarkup: genActionKbd(move)},
		})
		if err != nil || res.Result == nil {
			log.Println("UpdateStatus", err)
			return
		}

		SetPlayerMenuID(userID, res.Result.ID)
	}

	return
}

// Put the escaping sequence on name
func parseName(rawName string) string {
	var rx = regexp.MustCompile(`[\*\[\]\(\)\` + "`" + `~>#+\-=|{}.!]`)

	return rx.ReplaceAllString(rawName, "\\$0")
}

// Vreating an HTML link to the specified user
func GenUserLink(userID int64, name string) string {
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
func (b *bot) GetUserName(chatID int64) (name string) {
	res, err := b.GetChat(chatID)
	if err != nil || res.Result == nil {
		return "Unnamed User"
	}
	name = parseName(res.Result.FirstName)
	if name == "" {
		name = "Unnamed User"
	}

	return
}

// Try to edit a message if it can't or IDO == nil send a new one
func (b *bot) DisplayMessage(text string, IDO *echotron.MessageIDOptions, linkPreview bool, kbd *echotron.InlineKeyboardMarkup) (res echotron.APIResponseMessage, err error) {
	var (
		editOpt echotron.MessageTextOptions
		sendOpt echotron.MessageOptions
	)

	if IDO != nil {
		editOpt = echotron.MessageTextOptions{
			ParseMode:             echotron.HTML,
			DisableWebPagePreview: !linkPreview,
		}

		if kbd != nil {
			editOpt.ReplyMarkup = *kbd
		}

		res, err = b.EditMessageText(text, *IDO, &editOpt)
	}
	if err != nil || res.Result == nil {
		sendOpt = echotron.MessageOptions{
			ParseMode:             echotron.HTML,
			DisableWebPagePreview: !linkPreview,
		}

		if kbd != nil {
			sendOpt.BaseOptions = echotron.BaseOptions{ReplyMarkup: *kbd}
		}

		res, err = b.SendMessage(text, b.chatID, &sendOpt)
	}
	return
}
