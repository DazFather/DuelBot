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

// TOKEN is the Telegram API bot's token.
var TOKEN string

func newBot(chatID int64) echotron.Bot {
	return &bot{chatID, echotron.NewAPI(TOKEN)}
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
			"\nUse me inline or use the command /invite to fight against your friends",
			"\n\n💟 Created with love by @DazFather in Go using <a href=\"https://github.com/NicoNex/echotron\">echotron</a>.",
			" I'm also <a href=\"https://github.com/DazFather/DuelBot\">open source</a>",
		)
		opt.DisableWebPagePreview = true
	} else {
		switch payload[0] {
		case "noob":
			b.handleHelp()
			return
		case "joinDuel":
			b.handleAccept(-1, payload[1:])
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
	if update.InlineQuery == nil {
		return
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
						{{Text: "✅ Accept", URL: b.GenInvitationLink()}},
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

func (b *bot) handleAccept(msgID int, payload []string) {
	var userID int64

	if len(payload) != 2 {
		b.SendMessage("¯\\_(ツ)_/¯ Wrong format", b.chatID, nil)
		return
	}
	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		userID = int64(rawID)
	}

	if errMess := b.IsInvitionValid(userID, payload[1]); errMess != "" {
		b.SendMessage(errMess, b.chatID, nil)
		return
	}

	// Check if player is busy in another duel or not
	if !EngageDuel(b.chatID, userID) {
		b.SendMessage("You or your opponent might be already engaged in another fight. Brawls are still not allowed", b.chatID, nil)
		return
	}

	b.DeleteMessage(b.chatID, msgID)
	b.NotifyAcceptDuel(b.chatID, userID)
}

func (b *bot) handleAction(payload []string) {
	var (
		duration time.Duration
		enemyID  int64
		err      error
	)

	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("Calm down warrior... you are not in a fight anymore", b.chatID, nil)
		return
	}

	switch payload[0] {
	case "GUARD", "ATTACK", "DEFEND", "DODGE":
		// Don't do nothing if is the same action
		if move, _ := GetPlayerAction(b.chatID); move == payload[0] {
			return
		}

		// Set the player moves and update status
		duration, err = SetPlayerMoves(b.chatID, payload[0])
		if err != nil {
			log.Println("handleAction", "SetPlayerMoves", err)
			return
		}
		DisplayStatus(b.chatID, false)

		// If enemy is on guard spy the action
		enemyID, err = GetOpponentID(b.chatID)
		if err != nil {
			log.Println("handleAction", "GetOpponentID", err)
			return
		}

		if onGurad, _ := IsPlayerOnGuard(enemyID); onGurad {
			b.SpyAction(enemyID, b.chatID, payload[0])
		}

		// Loop looking for a change in the enemy status
		start := time.Now()
		for time.Since(start).Milliseconds() < duration.Milliseconds() {
			// if enemy is ready stop the loop
			if ready, err := IsPlayerReady(enemyID); err != nil {
				log.Println("handleAction", "IsPlayerReady", err)
				return
			} else if ready {
				break
			}

			time.Sleep(time.Duration(500) * time.Millisecond)
			// if the user change action quit this process
			if move, _ := GetPlayerAction(b.chatID); move != payload[0] {
				return
			}
		}

		// Perform the action between players and his opponent
		report, err := PlayersPerformMove(b.chatID)
		if err != nil {
			log.Println("handleAction", "PlayersPerformMove", err)
			return
		}

		// Notify users
		b.NotifyBattleReport(report)
		if !report.EndDuel {
			return
		}
		if report.WinnerID == nil {
			b.NotifyDraw()
		} else {
			b.NotifyEndDuel(*report.WinnerID)
		}

		// End duel (if duel ended)
		EndDuel(b.chatID)

	default:
		b.SendMessage("¯\\_(ツ)_/¯", b.chatID, nil)
	}
}

func (b *bot) TEMP_handleInvite(payload []string) {
	var (
		name   = fmt.Sprintf("[%s](tg://user?id=%d)", b.GetUserName(b.chatID), b.chatID)
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
	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("What are you running away from? There is no battle", b.chatID, nil)
		return
	}
	b.NotifyCancel()
	EndDuel(b.chatID)
}

func (b *bot) handleReject(update *echotron.Update, payload []string) {
	var (
		messageID int
		inviterID int64
		guestUser string
	)

	if len(payload) != 2 {
		b.SendMessage("Wrong format", b.chatID, nil)
	}

	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		inviterID = int64(rawID)
	}
	if rawID, err := strconv.Atoi(payload[1]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		messageID = rawID
	}

	b.EditMessageText(
		"✖️ <i>You delcine this invitation</i>",
		*extractMessageIDOpt(update),
		&echotron.MessageTextOptions{ParseMode: echotron.HTML},
	)

	guestUser = GenUserLink(b.chatID, extractName(update))
	b.DeleteMessage(inviterID, messageID)

	b.SendMessage(
		"✖️ <b>"+guestUser+" decline your invitation</b>\nProbably he was too afraid of you to accept",
		inviterID,
		&echotron.MessageOptions{ParseMode: echotron.HTML},
	)
}

func (b *bot) handleRematch(update *echotron.Update, payload []string) {
	var (
		opt    = echotron.MessageOptions{ParseMode: echotron.HTML}
		userID int64
		msgID  = extractMessageID(update)
	)

	if len(payload) != 1 {
		b.SendMessage("Wrong format", b.chatID, nil)
		fmt.Println("HEYA")
		return
	}

	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		fmt.Println("HEYA2")
		return
	} else {
		userID = int64(rawID)
	}

	opt.BaseOptions.ReplyMarkup = echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{{
			{Text: "✅ Accept", CallbackData: fmt.Sprintf("/accept %d %s", b.chatID, NewInviteID(b.chatID))},
			{Text: "❌ Decline", CallbackData: fmt.Sprintf("/reject %d %d", b.chatID, msgID)},
		}},
	}

	b.EditMessageText(
		"<b>Invitation sent</b>\n... waiting for a reply ⏳",
		echotron.NewMessageID(b.chatID, msgID),
		&echotron.MessageTextOptions{ParseMode: echotron.HTML},
	)

	b.SendMessage(
		fmt.Sprint(
			"🗡 <b>", GenUserLink(b.chatID, extractName(update)), " is challenging you for a rematch</b>",
			"\nWhat are you going to do?",
		),
		userID,
		&opt,
	)
}

func (b *bot) handleInviteLink(update *echotron.Update, payload []string) {
	var (
		err error
		res echotron.APIResponseMessage
		kbd = echotron.InlineKeyboardMarkup{
			InlineKeyboard: [][]echotron.InlineKeyboardButton{{
				{Text: "🔂 Refresh", CallbackData: "/invite refresh"},
			}},
		}
	)

	if len(payload) > 1 {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}

	if len(payload) == 1 && payload[0] == "refresh" {
		res, err = b.EditMessageText(
			"Here is your invitation link: "+b.GenInvitationLink(),
			*extractMessageIDOpt(update),
			&echotron.MessageTextOptions{
				ParseMode:   echotron.HTML,
				ReplyMarkup: kbd,
			},
		)

	}
	if len(payload) == 0 || err != nil || res.Result == nil {
		b.SendMessage(
			"Here is your invitation link: "+b.GenInvitationLink(),
			b.chatID,
			&echotron.MessageOptions{
				ParseMode:   echotron.HTML,
				BaseOptions: echotron.BaseOptions{ReplyMarkup: kbd},
			},
		)
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

	switch command {
	// Outside a duel
	case "/start":
		b.handleStart(payload)

	case "/help":
		b.handleHelp()

	case "/invite":
		b.handleInviteLink(update, payload)

	case "/rematch":
		b.handleRematch(update, payload)

	case "/inviteid":
		b.TEMP_handleInvite(payload)

	case "/accept":
		b.handleAccept(extractMessageID(update), payload)
	case "/reject":
		b.handleReject(update, payload)

	case "/id":
		b.SendMessage(fmt.Sprint(b.chatID), b.chatID, nil)

	// Inside a duel
	case "/action":
		b.handleAction(payload)

	case "/end", "/flee":
		b.handleFlee()
	}
}

func main() {
	if rawToken, err := LoadToken(); err != nil {
		fmt.Println(err)
		return
	} else {
		TOKEN = rawToken
	}
	dsp := echotron.NewDispatcher(TOKEN, newBot)
	log.Println(dsp.Poll())
}
