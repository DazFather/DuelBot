package main

import (
	"fmt"
	"strings"

	"github.com/NicoNex/echotron/v3"
)

// Make the actions (ME, ATTACK, GUARD ecc.. ) more pretty
func Prettfy(rawAction string, conditional bool, emoji int8) string {
	var selectEmoji = map[string]string{
		"ME":       "👤",
		"ENEMY":    "👤",
		"GUARD":    "👁‍🗨",
		"ATTACK":   "⚔️",
		"DEFEND":   "🛡",
		"DODGE":    "➰",
		"STUNNED":  "💫",
		"EXAUSTED": "🥵",
		"HELPLESS": "🆘",
	}
	pretty := fmt.Sprint(string(rawAction[0]), strings.ToLower(rawAction[1:]))

	if rawAction == "GUARD" {
		pretty = "On " + pretty
	}

	if conditional {
		switch rawAction {
		case "DODGE":
			pretty = pretty[:len(pretty)-1] + "ing"
		case "ATTACK", "DEFEND":
			pretty += "ing"
		}
	}

	switch emoji {
	case -1:
		pretty = selectEmoji[rawAction] + pretty
	case 1:
		pretty += selectEmoji[rawAction]
	}

	return pretty
}

// Generate the info bar with the changed status
func GenOffsetInfoBar(lifeOffset, staminaOffset, damageDealt int) string {
	var bar []string

	if lifeOffset > 0 {
		bar = append(bar, fmt.Sprint("❤ +", lifeOffset))
	} else if lifeOffset < 0 {
		bar = append(bar, fmt.Sprint("💔 ", lifeOffset))
	}

	if staminaOffset > 0 {
		bar = append(bar, fmt.Sprint("⚡ +", staminaOffset))
	} else if staminaOffset < 0 {
		bar = append(bar, fmt.Sprint("⚡ ", staminaOffset))
	}

	if damageDealt < 0 {
		bar = append(bar, fmt.Sprint("🗡 ", damageDealt*-1))
	}

	return strings.Join(bar, "|")
}

// Generate the info bar with life points and stamina bar
func genInfoBar(userID int64) string {
	life, stamina, max := GetPlayerInfo(userID)
	return fmt.Sprint(
		"❤:", "<code>", life, "</code>",
		" ⚡:[<code>", strings.Repeat("#", stamina), strings.Repeat(" ", max-stamina), "</code>]",
	)
}

// Generate the info bar with the action of the player
func genActionBar(userID int64) string {
	return fmt.Sprint("Action: <b>", Prettfy(GetPlayerAction(userID), true, 1), "</b>")
}

// Warn a user of the new status of the opponent
func (b *bot) SpyAction(toUserID, opponentID int64, move string) {
	text := fmt.Sprint(
		"👁‍🗨 <b><a href=\"tg://user?id=", opponentID, "\">Enemy</a> is ",
		strings.ToLower(Prettfy(move, true, 1)), "</b>\n",
		"Hurry up and prepare your counter-move!\n",
		"\n<i>You are able to recive this notification because you are on guard</i>",
	)
	UpdateStatus(toUserID, text, false)
}

// Notify the Result of a perform
func NotifyResult(userID int64, success bool, move, enemyMoves, infos string) {
	var descr string

	descr = fmt.Sprint(
		"<b>You %s ", strings.ToLower(Prettfy(move, false, 1)), "</b>\n",
		"<i>%s the enemy %s</i>\n", infos,
	)

	switch move {
	case "STUNNED", "EXAUSTED":
		descr = fmt.Sprintf(descr, "got", "meanwhile", strings.ToLower(enemyMoves))
	default:
		if success {
			enemyMoves = "was " + strings.ToLower(Prettfy(enemyMoves, true, 0))
			descr = fmt.Sprintf(descr, "successfully", "meanwhile", enemyMoves)
		} else {
			descr = fmt.Sprintf(descr, "tried to", "but", strings.ToLower(enemyMoves))
		}
	}

	b := echotron.NewAPI(TOKEN)
	b.SendMessage(descr, userID, &echotron.MessageOptions{ParseMode: echotron.HTML})
	DisplayStatus(userID, userID, true)
}

// Display the current status of a user
func DisplayStatus(toUserID, ofUserID int64, newMessage bool) {
	var text string

	if toUserID == ofUserID {
		text = fmt.Sprint("🏷 <b>Your</b> current status:\n", genActionBar(ofUserID))
	} else {
		text = fmt.Sprint("👤 <b><a href=\"tg://user?id=", ofUserID, "\">Enemy</a></b> current status:")
		if GetPlayerAction(toUserID) == "GUARD" {
			text += "\n" + genActionBar(ofUserID)
		}
	}
	text = fmt.Sprint(text, "\n", genInfoBar(ofUserID))

	UpdateStatus(toUserID, text, newMessage)
}

// Generate the inline keyboard with all the actions
func genActionKbd() (markup echotron.InlineKeyboardMarkup) {
	var (
		mainActions = []string{"ME", "ENEMY", "GUARD", "ATTACK", "DEFEND", "DODGE"}
		row         []echotron.InlineKeyboardButton
	)

	for i, action := range mainActions {
		row = append(row,
			echotron.InlineKeyboardButton{
				Text:         Prettfy(action, false, -1),
				CallbackData: fmt.Sprint("/action ", action),
			},
		)
		if (i+1)%2 == 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, row)
			row = nil
		}
	}

	if row != nil {
		markup.InlineKeyboard = append(markup.InlineKeyboard, row)
	}

	return
}

// Notify the users that the duel is starting
func NotifyAcceptDuel(firstID, secondID int64) {
	var (
		b   = echotron.NewAPI(TOKEN)
		IDs = [2]int64{firstID, secondID}
	)

	for i, currentID := range IDs {
		b.SendMessage(
			fmt.Sprint("Duel against ", genUserLink(IDs[1-i]), " is now starting 🏁"),
			currentID,
			&echotron.MessageOptions{ParseMode: echotron.HTML},
		)
		DisplayStatus(currentID, currentID, true)
	}
}

func (b *bot) NotifyDraw() {
	var IDs = []int64{b.chatID, GetOpponentID(b.chatID)}
	for i, id := range IDs {
		b.SendMessage(
			"⚖️ <b>The match is a draw</b> in the battle against "+genUserLink(IDs[1-i])+"\n",
			id,
			&echotron.MessageOptions{ParseMode: echotron.HTML},
		)
	}
}

// Notify the users of the end of a match
func (b *bot) NotifyEndDuel(winnerID int64) {
	var (
		opt      = echotron.MessageOptions{ParseMode: echotron.HTML}
		looserID = GetOpponentID(winnerID)
	)

	text := fmt.Sprint(
		"🥇 <b>You win</b> in the battle against ", genUserLink(looserID), "\n",
		"<i>Congratulation ", GetPlayerName(winnerID), " the big spirit of the war is proud of you</i>",
	)
	b.SendMessage(text, winnerID, &opt)

	text = fmt.Sprint(
		"☠ <b>You loose</b> the battle against ", genUserLink(winnerID), "\n",
		"<i>I hope that the guardian spirit can assist you in the next battle</i>",
	)
	b.SendMessage(text, looserID, &opt)
}

// Notify the users of the withdrawn of some of the two
func (b *bot) NotifyCancel() {
	var (
		opt      = echotron.MessageOptions{ParseMode: echotron.HTML}
		winnerID = GetOpponentID(b.chatID)
	)

	text := fmt.Sprint(
		"🏳️ <b>You flee</b> from the battle against ", genUserLink(winnerID), "\n",
		"<i>The big spirit of the war will not like this behaviour...</i>",
	)
	b.SendMessage(text, b.chatID, &opt)

	text = fmt.Sprint(
		"🏃 <b>Your <a href=\"tg://user?id=", b.chatID, "\">opponent</a> has withdrawn</b>\n",
		"<i>Probably you are too strong for him or maybe he doesn't like your face...</i>",
	)
	b.SendMessage(text, winnerID, &opt)
}
