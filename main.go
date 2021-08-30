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

// Create a new bot
func newBot(chatID int64) echotron.Bot {
	return &bot{chatID, echotron.NewAPI(TOKEN)}
}

// Handle the start message and redirect links to their helper funcions
func (b *bot) handleStart(update *echotron.Update, payload []string) {
	var username string
	if res, err := b.GetMe(); err == nil && res.Result != nil {
		username = "@" + res.Result.Username
	} else {
		return
	}

	switch len(payload) {
	case 0:
		text := fmt.Sprint(
			"👋 <b>Welcome duelist to ", username, "</b> <code>[BETA]</code>\n",
			"If you are new I suggest you to see how to play using /help\n",
			"\nUse me inline or use the command /invite to fight against your friends",
			"\n\n💟 Created with love by @DazFather in Go using <a href=\"https://github.com/NicoNex/echotron\">echotron</a>.",
			" I'm also <a href=\"https://github.com/DazFather/DuelBot\">open source</a>",
		)

		kbd := echotron.InlineKeyboardMarkup{
			InlineKeyboard: [][]echotron.InlineKeyboardButton{{
				{Text: "❓ How to play", CallbackData: "/help"},
				{Text: "🕹 Play with others", CallbackData: "/start invitationInfo"},
			}},
		}

		b.DisplayMessage(text, extractMessageIDOpt(update), false, &kbd)
		return
	case 1:
		switch payload[0] {
		case "noob":
			b.handleHelp(update, []string{"0"})
			return

		case "invitationInfo":
			b.DisplayMessage(
				fmt.Sprint(
					"💬 <b>Invite other users to a duel</b>",
					"\nJust type <code>", username, "</code> in any chat to <i>auto",
					"magiacally✨</i> generate an invitation message",
					"\nIf you prefer to create your own instead you can generate",
					" a new invitation link using the button below",
				),
				extractMessageIDOpt(update),
				false,
				&echotron.InlineKeyboardMarkup{
					InlineKeyboard: [][]echotron.InlineKeyboardButton{
						{{Text: "🔗 New invite link", CallbackData: "/invite refresh"}},
						{{Text: "✨ Inline invitation", SwitchInlineQuery: "DuellingRobot"}},
						{{Text: "🔙 Main menu", CallbackData: "/start"}},
					},
				},
			)
			return
		}
	case 3:
		if payload[0] == "joinDuel" {
			b.handleAccept(-1, payload[1:])
			return
		}
	}
	b.SendMessage("¯\\_(ツ)_/¯ Wrong format", b.chatID, nil)
}

// Handle the tutorial
func (b *bot) handleHelp(update *echotron.Update, payload []string) {
	var (
		prev, next echotron.InlineKeyboardButton
		text       string
	)

	if len(payload) == 0 {
		payload = append(payload, "0")
	} else if len(payload) != 1 {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}

	switch payload[0] {
	case "close":
		b.DeleteMessage(b.chatID, extractMessageID(update))
		return
	case "0":
		prev.Text, prev.CallbackData = "🔙 Main menu", "/start"
		next.Text, next.CallbackData = "Next ⏭", "/help 1"
		text = "No one"

	case "1":
		prev.Text, prev.CallbackData = "⏮ Prev.", "/help 0"
		next.Text, next.CallbackData = "Next ⏭", "/help 2"
		text = "is gonna"

	case "2":
		prev.Text, prev.CallbackData = "⏮ Prev.", "/help 1"
		next.Text, next.CallbackData = "Play 🕹", "/start invitationInfo"
		text = "help ya"

	default:
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}

	kbd := echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{prev, next},
			{{Text: "❌ Close", CallbackData: "/help close"}},
		},
	}

	b.DisplayMessage(text, extractMessageIDOpt(update), false, &kbd)
}

// Handle the inline invitation
func (b *bot) handleInviteInline(update *echotron.Update) {
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

// Handle the request of a new invite link
func (b *bot) handleInviteLink(update *echotron.Update, payload []string) {
	var kbd = echotron.InlineKeyboardMarkup{
		InlineKeyboard: [][]echotron.InlineKeyboardButton{
			{{Text: "🔂 Refresh", CallbackData: "/invite refresh"}},
			{{Text: "🔙 Go Back", CallbackData: "/start invitationInfo"}},
		},
	}

	if len(payload) > 1 {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}

	text := fmt.Sprint(
		"🔗 <b>Invitation link</b>",
		"\nThis is your invitation link, you can send it to your friends so when",
		" one of them click on it a duel will start against him. If you change",
		" your idea you can refresh for a new one so the old one will stop working.\n",
		"\n ", b.GenInvitationLink(), "\n",
		"\n⚠️<i>Refreshing, opening again this section or using the bot inline ",
		"will generate a new invite that will make the previous invalid.</i> ",
		"<a href=\"https://telegra.ph/DuelBot---I-care-about-Privacy-08-26\">",
		"Because I care about privacy</a>",
	)

	if len(payload) == 1 && payload[0] == "refresh" {
		b.DisplayMessage(text, extractMessageIDOpt(update), false, &kbd)
	} else {
		b.SendMessage(
			text,
			b.chatID,
			&echotron.MessageOptions{
				ParseMode:             echotron.HTML,
				BaseOptions:           echotron.BaseOptions{ReplyMarkup: kbd},
				DisableWebPagePreview: true,
			},
		)
	}
}

// Handle the sending a match request to a specified userID
func (b *bot) handleInviteUserID(update *echotron.Update, payload []string) {
	var (
		opt      = echotron.MessageOptions{ParseMode: echotron.HTML}
		msgID    = extractMessageID(update)
		userName = GenUserLink(b.chatID, extractName(update))
		userID   int64
		text     string
	)

	switch len(payload) {
	case 1:
		text = "🗡 <b>" + userName + " want to challenge you in a duel</b>"
	case 2:
		if payload[1] != "rematch" {
			b.SendMessage("Wrong format", b.chatID, nil)
			return
		}
		text = "🗡 <b>" + userName + " is challenging you for a rematch</b>"
	default:
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}

	if rawID, err := strconv.Atoi(payload[0]); err != nil {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	} else {
		userID = int64(rawID)
	}
	if res, _ := b.GetMe(); res.Result.ID == userID {
		b.SendMessage("Sorry I'm too busy now, maybe another time", b.chatID, nil)
		return
	}
	if userID == b.chatID {
		b.SendMessage("Do you have a double personality? 👀", b.chatID, nil)
		return
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

	b.SendMessage(text+"\nWhat are you going to do?", userID, &opt)
}

// Handle the accepting of a incoming match request
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

// Handle the rejecting of a incoming match request
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

// Handle the changing action inside a duel
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

// Handle the exit from a duel
func (b *bot) handleFlee() {
	if !IsPlayerBusy(b.chatID) {
		b.SendMessage("What are you running away from? There is no battle", b.chatID, nil)
		return
	}
	b.NotifyCancel()
	EndDuel(b.chatID)
}

// Manage the incoming inputs (uptate) from Telegram
func (b *bot) Update(update *echotron.Update) {
	var command, payload = extractCommand(update)

	// Inviting a user with inline mode
	if update.InlineQuery != nil {
		b.handleInviteInline(update)
		return
	}

	switch command {
	// Outside a duel
	case "/start":
		b.handleStart(update, payload)

	case "/help":
		b.handleHelp(update, payload)

	case "/invite":
		b.handleInviteLink(update, payload)

	case "/inviteid":
		b.handleInviteUserID(update, payload)

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
