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
					"\nTap on \"Inline invitation\" or simply type <code>", username, "</code> in any chat to <i>auto",
					"magiacally✨</i> generate an invitation message",
					"\nIf you prefer to create your own instead, you can generate",
					" a new invitation link using the button below",
				),
				extractMessageIDOpt(update),
				false,
				&echotron.InlineKeyboardMarkup{
					InlineKeyboard: [][]echotron.InlineKeyboardButton{
						{{Text: "✨ Inline invitation", SwitchInlineQuery: "DuellingRobot"}},
						{{Text: "🔗 Invite link", CallbackData: "/invite refresh"}},
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
		text = fmt.Sprint(
			"<b>What is DuelBot❓</b>\n",
			"DuelBot is a Telegram bot where you can fight your friends in real-time.\n",
			"It's currently under development by ", GenUserLink(169090723, "@DazFather"),
			" and it's still on beta so it might be pretty unstable and things are goning to",
			" change in future. Also it's <a href=\"https://github.com/DazFather/DuelBot\">",
			"open source</a>, so feel free to contribute.\n",
			"\nUse the buttons below to navigate into the help section. Tap on \"Next ⏭\"",
			" if you want to know more about how the duelling mechanics works",
		)

	case "1":
		prev.Text, prev.CallbackData = "⏮ Prev.", "/help 0"
		next.Text, next.CallbackData = "Next ⏭", "/help 2"
		text = fmt.Sprint(
			"<b>How to play - Stats 🧮</b>\n",
			"Every player have two main stats:\n",
			"❤️ <b>health</b> - that start at 20 and it reduce every time you recive",
			" a damage. If it reach 0 you loose. Currently there is no way to heal\n",
			"⚡ <b>stamina bar</b> - that start at 6 and it cap at 10. It also reduce",
			" itself when you make an action that require energy like <i>dodging</i>",
			" or <i>attacking</i> and it influence the speed of execution of these, ",
			"so more stamina you got, faster it is. If it reach 0 you become <i>exausted</i>",
			" and will not be able to move untill the opponent do something.\nYou can",
			"gain some stamina back by <i>defending</i>\n",
			"\nThere is also <b>damage</b> (\"⚔\") that rapresent how much damage ",
			"that you can deal to an enemy. It's value is always 5 but if the enemy ",
			"is <i>attacking</i>, it will recive just 2. ",
			"(half of the damage of the opponent rounded down)\n",
			"\nEvery time you clash against the opponent you recive a report ",
			"where is how your stats modified and the damage you dealt",
			"\n\n⚠ <i>This bot is still on beta so things can change in future</i>",
		)

	case "2":
		prev.Text, prev.CallbackData = "⏮ Prev.", "/help 1"
		next.Text, next.CallbackData = "Play 🕹", "/start invitationInfo"
		text = fmt.Sprint(
			"<b>How to play - Actions 💪</b>\n",
			"There are four actions you can set yourself douring a duel:\n",
			"👁‍🗨 <b>on guard</b> - it's the default one. Althow is pretty useless when",
			" you clash (if enemy is <i>defending</i> you also get <i>stunned</i>)",
			", it allow you to see and get notified when the opponent change his",
			" action so you can use that time to quickly set your counter-move\n",
			"🛡 <b>defend</b> - it allow you to gain 1 stamina back and recive half ",
			"of the damage when hit. When you clash against an enemy it allow you",
			" to <i>stunn</i> it if it's <i>on guard</i> or if it's <i>dodging</i>",
			" but you have more stamina\n",
			"⚔ <b>attack</b> - you deal damage to the enemy if is not <i>defending</i> is 5\n",
			"➰ <b>dodge</b> - it allow you to not recive any damage if the enemy ",
			"is <i>attacking</i> but only if you are faster",
			"\n\n⚠ <i>This bot is still on beta so things can change in future</i>",
		)

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
				Type:        echotron.INLINE_ARTICLE,
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

// Handle the accepting of an incoming match request
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

// Handle the rejecting of an incoming match request
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

	if len(payload) != 1 {
		b.SendMessage("Wrong format", b.chatID, nil)
		return
	}

	// Grab enemyID
	enemyID, err = GetOpponentID(b.chatID)
	if err != nil {
		b.SendMessage("Calm down warrior... you are not in a fight anymore", b.chatID, nil)
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
	if onGurad, _ := IsPlayerOnGuard(enemyID); onGurad {
		b.SpyAction(enemyID, b.chatID, payload[0])
	}

	// If action does not require any duration (already done) just quit
	if duration == time.Duration(0) || payload[0] == "DEFEND" {
		return
	}

	// Perform the action between players and his opponent
	report, err := PlayersPerformMove(b.chatID, enemyID)
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
		b.NotifyDraw(report.PlayersInfo[0].UserID, report.PlayersInfo[1].UserID)
	} else {
		b.NotifyEndDuel(*report.WinnerID)
	}

	// End duel (if duel ended)
	EndDuel(b.chatID)
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

// Handle the sending of the entire last battle history
func (b *bot) handleBattleHistory() {
	var history = GenPlayerHistory(b.chatID)

	if history == "" {
		history = "<i>There is nothing to see here</i>"
	}
	b.SendMessage(history, b.chatID, &echotron.MessageOptions{ParseMode: echotron.HTML})
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

	case "/history":
		b.handleBattleHistory()

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
