package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/NicoNex/echotron/v3"
)

// TEMP: here will bw saved the battle history of a player
var lastBattle = make(map[int64][]string, 0)

// TEMP: generate the entire battle history ready to be displayed
func GenPlayerHistory(ownerID int64) string {
	return strings.Join(lastBattle[ownerID], "\n  ➖➖➖➖➖➖➖➖➖")
}

// TEMP: add a new report to the battle history of a player
func addToPlayerHistory(ownerID int64, msgReport string) {
	lastBattle[ownerID] = append(lastBattle[ownerID], msgReport)
}

// Make the actions (ME, ATTACK, GUARD ecc.. ) more pretty
func Prettfy(rawAction string, conditional bool, emoji int8) (pretty string) {
	var selectEmoji = map[string]string{
		"ME":       "👤",
		"ENEMY":    "👤",
		"GUARD":    "👁‍🗨",
		"ATTACK":   "⚔️",
		"DEFEND":   "🛡",
		"DODGE":    "➰",
		"STUNNED":  "💫",
		"EXAUSTED": "🥵",
		"HELPLESS": "😵",
	}

	switch rawAction {
	case "HELPLESS":
		pretty = "Unable to fight"
	case "GUARD":
		pretty = "On Guard"
	default:
		pretty = fmt.Sprint(string(rawAction[0]), strings.ToLower(rawAction[1:]))
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
	life, stamina, max, _, err := GetPlayerInfo(userID)
	if err != nil {
		return ""
	}
	return fmt.Sprint(
		"❤:", "<code>", life, "</code>",
		" ⚡:[<code>", strings.Repeat("#", int(stamina)), strings.Repeat(" ", int(max-stamina)), "</code>]",
	)
}

// Generate the info bar with the action of the player
func genActionBar(userID int64) string {
	move, err := GetPlayerAction(userID)
	if err != nil {
		return ""
	}
	return fmt.Sprint("<b>", Prettfy(move, true, 1), "</b>")
}

// Warn a user of the new status of the opponent
func (b *bot) SpyAction(toUserID, opponentID int64, move string) {
	text := fmt.Sprint(
		"👁‍🗨 <b>", GenUserLink(opponentID, "Enemy"), " is ",
		strings.ToLower(Prettfy(move, true, 1)), "</b>\n",
		"Hurry up and prepare your counter-move!\n",
		"\n<i>You are able to recive this notification because you are on guard</i>",
	)
	UpdateStatus(toUserID, text, false)
}

// Notify the result of a perform
func (b *bot) NotifyBattleReport(report BattleReport) {
	for i, current := range report.PlayersInfo {
		enemy := report.PlayersInfo[1-i]
		DisplayReport(current, enemy)
		if !report.EndDuel {
			DisplayStatus(current.UserID, false)
		}
	}
}

// Display the last battle report deleting the previous
func DisplayReport(current, enemy PlayerReport) {
	var text string

	if current.GainEffect != nil {
		text = "\n<b>You got " + Prettfy(*current.GainEffect, false, 1) + "</b>"
	}

	text += "\nYou %s" + Prettfy(current.Performed, false, 1) + "%s"
	switch current.Performed {
	case "HELPLESS", "STUNNED", "EXAUSTED":
		text = fmt.Sprintf(text, "<b>were ", "</b>")
	default:
		if current.Success {
			text = fmt.Sprintf(text, "<b>", " successfully</b>\nmeanwhile")
		} else {
			text = fmt.Sprintf(text, "<b>tried to ", "</b> but...")
		}
	}

	text += "\nEnemy <b>was " + Prettfy(enemy.Performed, true, 1) + "</b>"

	if enemy.GainEffect != nil {
		text += "\n<b>Enemy got " + Prettfy(*enemy.GainEffect, false, 1) + "</b>"
	}

	text += "\n\n" + GenOffsetInfoBar(current.LifeOff, current.StaminaOff, enemy.LifeOff)

	addToPlayerHistory(current.UserID, text)
	UpdateReport(current.UserID, text)
}

// Display the current status of a user
func DisplayStatus(toUserID int64, newMessage bool) {
	var text string

	enemyID, _ := GetOpponentID(toUserID)

	text = fmt.Sprint(
		"🏷 <b>You</b>: ", genInfoBar(toUserID), "\n\n",
		"👤 <b>", GenUserLink(enemyID, "Enemy"), "</b>",
	)
	if onGurad, _ := IsPlayerOnGuard(toUserID); onGurad {
		text += " current status: " + genActionBar(enemyID) + "\n"
	} else {
		text += ": "
	}
	text += genInfoBar(enemyID)

	UpdateStatus(toUserID, text, newMessage)
}

// Generate the inline keyboard with all the actions
func genActionKbd(move string) (markup echotron.InlineKeyboardMarkup) {
	var (
		mainActions = []string{"GUARD", "ATTACK", "DEFEND", "DODGE"}
		row         []echotron.InlineKeyboardButton
	)

	for i, action := range mainActions {
		btn := echotron.InlineKeyboardButton{CallbackData: "/action " + action}

		if move == action {
			btn.Text = "▶️ " + Prettfy(action, false, 0) + " ◀️"
		} else {
			btn.Text = Prettfy(action, false, -1)
		}
		row = append(row, btn)

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

// Generate the inline keyboard with all the actions
func genRematchKbd(opponentID int64) (markup *echotron.MessageReplyMarkup) {
	var kbd echotron.InlineKeyboardMarkup

	kbd.InlineKeyboard = [][]echotron.InlineKeyboardButton{{
		{Text: "📜 Battle history", CallbackData: "/history"},
		{Text: "🔄 Rematch", CallbackData: fmt.Sprint("/inviteid ", opponentID, " rematch")},
	}}

	return &echotron.MessageReplyMarkup{kbd}
}

// Notify the users that the duel is starting
func (b *bot) NotifyAcceptDuel(firstID, secondID int64) {
	var IDs = [2]int64{firstID, secondID}

	for i, currentID := range IDs {
		delete(lastBattle, currentID)
		user := GenUserLink(IDs[1-i], b.GetUserName(IDs[1-i]))
		b.SendMessage(
			fmt.Sprint("Duel against ", user, " is now starting 🏁"),
			currentID,
			&echotron.MessageOptions{ParseMode: echotron.HTML},
		)
		DisplayStatus(currentID, true)
		UpdateReport(currentID, "Enemy is approching...\n<i>Here will be displayed the report of the last clash. Now is still empty</i>")
	}
}

// Notify the users of the end of a match by draw
func (b *bot) NotifyDraw(player1ID, player2ID int64) {
	var IDs = []int64{player1ID, player2ID}

	for i, id := range IDs {
		enemy := GenUserLink(IDs[1-i], b.GetUserName(IDs[1-i]))
		res, _ := b.SendMessage(
			fmt.Sprint("⚖️ <b>The match is a draw</b> in the battle against ", enemy, "\n"),
			id,
			&echotron.MessageOptions{ParseMode: echotron.HTML},
		)
		b.EditMessageReplyMarkup(echotron.NewMessageID(id, res.Result.ID), genRematchKbd(IDs[1-i]))
		addToPlayerHistory(id, "⚖️ <b>The match is a draw</b>")
	}
}

// Notify the users of the win / lost of a match
func (b *bot) NotifyEndDuel(winnerID int64) {
	var opt = echotron.MessageOptions{ParseMode: echotron.HTML}

	looserID, err := GetOpponentID(winnerID)
	if err != nil {
		log.Println("NotifyEndDuel", "GetOpponentID", err)
		return
	}
	winnerName, looserName := b.GetUserName(winnerID), b.GetUserName(looserID)

	text := fmt.Sprint(
		"🥇 <b>You win</b> in the battle against ", GenUserLink(looserID, looserName), "\n",
		"<i>Congratulation ", winnerName, " the big spirit of the war is proud of you</i>",
	)
	addToPlayerHistory(winnerID, "🥇 <b>You win</b>")
	res, _ := b.SendMessage(text, winnerID, &opt)
	b.EditMessageReplyMarkup(echotron.NewMessageID(winnerID, res.Result.ID), genRematchKbd(looserID))

	text = fmt.Sprint(
		"☠ <b>You loose</b> the battle against ", GenUserLink(winnerID, winnerName), "\n",
		"<i>I hope that the guardian spirit can assist you in the next battle</i>",
	)
	addToPlayerHistory(looserID, "☠ <b>You loose</b>")
	res, _ = b.SendMessage(text, looserID, &opt)
	b.EditMessageReplyMarkup(echotron.NewMessageID(looserID, res.Result.ID), genRematchKbd(winnerID))
}

// Notify the users of the withdrawn of one of the two
func (b *bot) NotifyCancel() {
	var opt = echotron.MessageOptions{ParseMode: echotron.HTML}

	winnerID, err := GetOpponentID(b.chatID)
	if err != nil {
		log.Println("NotifyCancel", "GetOpponentID", err)
		return
	}

	text := fmt.Sprint(
		"🏳️ <b>You flee</b> from the battle against ", GenUserLink(winnerID, b.GetUserName(winnerID)), "\n",
		"<i>The big spirit of the war will not like this behaviour...</i>",
	)
	b.SendMessage(text, b.chatID, &opt)
	delete(lastBattle, b.chatID)

	text = fmt.Sprint(
		"🏃 <b>Your ", GenUserLink(b.chatID, "opponent"), " has withdrawn</b>\n",
		"<i>Probably you are too strong for him or maybe he doesn't like your face...</i>",
	)
	b.SendMessage(text, winnerID, &opt)
	delete(lastBattle, winnerID)
}
