package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/NicoNex/echotron/v3"
)

type bot struct {
	chatID int64
	echotron.API
}

// TOKEN is the bot's token.
const TOKEN = ""

func newBot(chatID int64) echotron.Bot {
	api := echotron.NewAPI(TOKEN)

	if res, err := api.GetChat(chatID); err == nil {
		if res.Result != nil && res.Result.Type == "private" {
			AddNewPlayer(chatID, res.Result.FirstName)
		}
	}

	return &bot{chatID, api}
}

func (b *bot) handleStart() {
	var botUsername string

	if res, err := b.GetMe(); err == nil && res.Result != nil {
		botUsername = res.Result.Username
	}
	text := fmt.Sprint(
		"👋 <b>Welcome adventurer to @", botUsername, "</b> <code>[BETA]</code>\n",
		"If you are new I suggest you to see how to play using /help\n",
		"Use me inline in to engage a fight against another user\n",
		"\n💟 Created with love by @DazFather in Go using <a href=\"https://github.com/NicoNex/echotron\">echotron</a>",
	)
	b.SendMessage(text, b.chatID, &echotron.MessageOptions{
		ParseMode:             echotron.HTML,
		DisableWebPagePreview: true,
	})
}

func (b *bot) handleInvite(update *echotron.Update) {
	if update.InlineQuery.ChatType == "private" {
		b.AnswerInlineQuery(
			update.InlineQuery.ID,
			[]echotron.InlineQueryResult{
				&echotron.InlineQueryResultArticle{
					Type:        echotron.ARTICLE,
					ID:          string(b.chatID),
					Title:       "Engage a duel",
					Description: "Invite this user to a duel",
					ReplyMarkup: echotron.InlineKeyboardMarkup{
						[][]echotron.InlineKeyboardButton{
							{{Text: "✅ Accept", CallbackData: fmt.Sprint("/accept ", b.chatID)}},
						},
					},
					InputMessageContent: echotron.InputTextMessageContent{
						MessageText: "Do you have the guts to face me in a duel?",
					},
				},
			},
			&echotron.InlineQueryOptions{
				IsPersonal:        true,
				SwitchPmText:      "What is this?",
				SwitchPmParameter: "https://t.me/nbortkbijrjbwm_bot?start=noob",
			},
		)
	}
}

func (b *bot) handleAccept(payload []string) {
	var userID int64

	if rawID, err := strconv.Atoi(payload[0]); err == nil {
		userID = int64(rawID)
	}

	// Check if player is busy in another duel or not
	if !EngageDuel(b.chatID, userID) {
		b.SendMessage("You or your opponent might be already engaged in another fight. Brawls are still not allowed", b.chatID, nil)
		return
	}

	NotifyAcceptDuel(b.chatID, userID)
}

func (b *bot) handleAction(payload []string) {
	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("Calm down warrior... you are not in a fight anymore", b.chatID, nil)
		return
	}

	switch payload[0] {
	case "ME":
		DisplayStatus(b.chatID, b.chatID, false)

	case "ENEMY":
		DisplayStatus(b.chatID, GetOpponentID(b.chatID), false)

	case "GUARD", "ATTACK", "DEFEND", "DODGE":
		SetPlayerMoves(b.chatID, payload[0])
		enemyID := GetOpponentID(b.chatID)
		if IsPlayerOnGuard(enemyID) {
			b.SpyAction(enemyID, b.chatID, payload[0])
		}
		msg := fmt.Sprint(
			"You are now ", Prettfy(payload[0], true, 1),
			"\nPress or type \"Confirm\" to perform",
		)
		UpdateStatus(b.chatID, msg, false)

	default:
		b.SendMessage("¯\\_(ツ)_/¯", b.chatID, nil)
	}
}

func (b *bot) TEMP_handleInvite(payload []string) {
	var (
		name   = fmt.Sprintf("[%s](tg://user?id=%d)", GetPlayerName(b.chatID), b.chatID)
		opt    = &echotron.MessageOptions{ParseMode: echotron.MarkdownV2}
		userID int64
	)

	opt.BaseOptions.ReplyMarkup = echotron.InlineKeyboardMarkup{
		[][]echotron.InlineKeyboardButton{
			{{Text: "✅ Accept", CallbackData: fmt.Sprint("/accept ", b.chatID)}},
		},
	}

	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		userID = int64(rawID)
	}

	_, err := b.SendMessage(fmt.Sprint("😏 Do you have the guts to face ", name, " in a duel?"), userID, opt)
	if err != nil {
		b.SendMessage("🚫 This user never fight before. We duel, not murder", b.chatID, nil)
	} else {
		b.SendMessage("✅ Invitation sent.", b.chatID, nil)
	}
}

func (b *bot) handleConfirm() {
	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("¯\\_(ツ)_/¯ You are not in a battle", b.chatID, nil)
	}
	if isEnded, err := PlayersPerformMove(b.chatID); isEnded && err == nil {
		b.NotifyEndDuel(IsPlayerWinner(b.chatID))
		EndDuel(b.chatID)
	}
}

func (b *bot) handleFlee() {
	if IsPlayerBusy(b.chatID) {
		b.NotifyCancel()
		EndDuel(b.chatID)
	} else {
		b.SendMessage("¯\\_(ツ)_/¯ You are not in a battle", b.chatID, nil)
	}
}

// Manage the incoming inputs (uptate) from Telegram
func (b *bot) Update(update *echotron.Update) {
	var command, payload = extractCommand(update)

	/* Work in progress...
	if update.InlineQuery != nil {
		b.handleInvite(update)
	} */

	// Outside a duel
	switch command {
	case "/start":
		b.handleStart()

	case "/help":
		b.SendMessage("😈 No one is gonna help ya", b.chatID, nil)

	case "/invite":
		b.TEMP_handleInvite(payload)

	case "/accept":
		b.handleAccept(payload)

	case "/id":
		b.SendMessage(fmt.Sprint(b.chatID), b.chatID, nil)
	}

	// Inside a duel
	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("¯\\_(ツ)_/¯ You are not in a battle", b.chatID, nil)
		return
	}
	switch command {
	case "/action":
		b.handleAction(payload)

	case "Confirm":
		b.handleConfirm()

	case "/end", "/flee":
		b.handleFlee()
	}
}

func main() {
	if TOKEN == "" {
		fmt.Println("Missing TOKEN value")
	}
	dsp := echotron.NewDispatcher(TOKEN, newBot)
	log.Println(dsp.Poll())
}
