package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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
			AddNewPlayer(chatID, ParseName(res.Result.FirstName))
		}
	}

	return &bot{chatID, api}
}

func (b *bot) handleStart(payload []string) {
	var (
		opt           = echotron.MessageOptions{ParseMode: echotron.HTML}
		text, botUser string
	)

	if len(payload) == 0 {
		if res, err := b.GetMe(); err == nil && res.Result != nil {
			botUser = "@" + res.Result.Username
		}
		text = fmt.Sprint(
			"👋 <b>Welcome duelist to ", botUser, "</b> <code>[BETA]</code>\n",
			"If you are new I suggest you to see how to play using /help\n",
			"Use me inline in to engage a fight against another user\n",
			"\n💟 Created with love by @DazFather in Go using <a href=\"https://github.com/NicoNex/echotron\">echotron</a>.",
			" I'm also <a href=\"https://github.com/DazFather/DuelBot\">open source</a>",
		)
		opt.DisableWebPagePreview = true
	} else {
		switch payload[0] {
		case "noob":
			b.handleHelp()
			return
		case "acceptInvite":
			b.handleAccept(payload[1:])
			return
		default:
			text = "¯\\_(ツ)_/¯ Wrong format"
		}
	}
	b.SendMessage(text, b.chatID, &opt)
}

func (b *bot) handleHelp() {
	b.SendMessage("😈 No one is gonna help ya", b.chatID, nil)
}

func (b *bot) handleInvite(update *echotron.Update) {
	var botUser string

	if update.InlineQuery == nil {
		return
	}

	if res, err := b.GetMe(); err == nil {
		botUser = res.Result.Username
	}
	b.AnswerInlineQuery(
		update.InlineQuery.ID,
		[]echotron.InlineQueryResult{
			&echotron.InlineQueryResultArticle{
				Type:        echotron.ARTICLE,
				ID:          string(b.chatID),
				Title:       "Engage a duel",
				Description: "Invite this user to a duel",
				HideURL:     false,
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					[][]echotron.InlineKeyboardButton{
						{{Text: "✅ Accept", URL: fmt.Sprint("https://t.me/", botUser, "?start=acceptInvite_", b.chatID)}},
					},
				},
				InputMessageContent: echotron.InputTextMessageContent{
					MessageText: "Do you have the guts to face me in a duel?",
				},
			},
		},
		&echotron.InlineQueryOptions{
			CacheTime:         0,
			IsPersonal:        true,
			SwitchPmText:      "What is this?",
			SwitchPmParameter: "noob",
		},
	)

}

func (b *bot) handleAccept(payload []string) {
	var userID int64

	if len(payload) != 1 {
		b.SendMessage("¯\\_(ツ)_/¯ Wrong format", b.chatID, nil)
		return
	}
	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		userID = int64(rawID)
		if res, _ := b.GetMe(); res.Result.ID == userID {
			b.SendMessage("Sorry I'm too busy now, maybe another time", b.chatID, nil)
			return
		} else if userID == b.chatID {
			b.SendMessage("Do you have a double personality? 👀", b.chatID, nil)
			return
		}
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
		// Don't do nothing if is the same action
		if payload[0] == GetPlayerAction(b.chatID) {
			return
		}

		// Set the player moves and update status
		duration, err := SetPlayerMoves(b.chatID, payload[0])
		if err != nil {
			log.Println("handleAction", "SetPlayerMoves", err)
			return
		}
		DisplayStatus(b.chatID, b.chatID, false)

		// If enemy is on guard spy the action
		enemyID := GetOpponentID(b.chatID)
		if IsPlayerOnGuard(enemyID) {
			b.SpyAction(enemyID, b.chatID, payload[0])
		}

		// Loop looking for a change in the enemy status
		for i := 0; i < int(duration.Seconds())*2; i++ {
			// if enemy is ready stop the loop
			if IsPlayerReady(enemyID) {
				break
			}
			time.Sleep(time.Duration(500) * time.Millisecond)
			// if the user change action quit this process
			if payload[0] != GetPlayerAction(b.chatID) {
				return
			}
		}
		if end, winnerID, _ := PlayersPerformMove(b.chatID); end {
			if winnerID == 0 {
				b.NotifyDraw()
			} else {
				b.NotifyEndDuel(winnerID)
			}
			EndDuel(b.chatID)
		}

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

	if len(payload) != 1 {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}
	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		userID = int64(rawID)
		if res, _ := b.GetMe(); res.Result.ID == userID {
			b.SendMessage("Do you have a double personality? 👀", b.chatID, nil)
			return
		} else if userID == b.chatID {
			b.SendMessage("Sorry I'm too busy now, maybe another time", b.chatID, nil)
			return
		}
	}

	_, err := b.SendMessage(fmt.Sprint("😏 Do you have the guts to face ", name, " in a duel?"), userID, opt)
	if err != nil {
		b.SendMessage("🚫 This user never fight before. We duel, not murder", b.chatID, nil)
	} else {
		b.SendMessage("✅ Invitation sent.", b.chatID, nil)
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

	// Inviting a user with inline mode
	if update.InlineQuery != nil {
		b.handleInvite(update)
		return
	}

	// Outside a duel
	switch command {
	case "/start":
		b.handleStart(payload)
		return

	case "/help":
		b.handleHelp()
		return

	case "/invite":
		b.TEMP_handleInvite(payload)
		return

	case "/accept":
		b.handleAccept(payload)
		return

	case "/id":
		b.SendMessage(fmt.Sprint(b.chatID), b.chatID, nil)
		return
	}

	// Inside a duel
	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("¯\\_(ツ)_/¯ You are not in a battle", b.chatID, nil)
		return
	}
	switch command {
	case "/action":
		b.handleAction(payload)

	case "/end", "/flee":
		b.handleFlee()
	}
}

func main() {
	if TOKEN == "" {
		fmt.Println("Missing TOKEN value")
		return
	}
	dsp := echotron.NewDispatcher(TOKEN, newBot)
	log.Println(dsp.Poll())
}
