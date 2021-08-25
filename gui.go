package main

import (
	"fmt"
	"log"
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
	return fmt.Sprint("Action: <b>", Prettfy(move, true, 1), "</b>")
}

// Warn a user of the new status of the opponent
func (b *bot) SpyAction(toUserID, opponentID int64, move string) {
	text := fmt.Sprint(
		"👁‍🗨 <b>", genUserLink(opponentID, "Enemy"), " is ",
		strings.ToLower(Prettfy(move, true, 1)), "</b>\n",
		"Hurry up and prepare your counter-move!\n",
		"\n<i>You are able to recive this notification because you are on guard</i>",
	)
	UpdateStatus(toUserID, text, false)
}

// Notify the result of a perform
func (b *bot) NotifyBattleReport(report BattleReport) {
	var opt = echotron.MessageOptions{ParseMode: echotron.HTML}

	for i, r := range report.PlayersInfo {
		msg := ""
		enemy := report.PlayersInfo[1-i]

		if r.GainEffect != nil {
			msg = "<b>You got " + Prettfy(*r.GainEffect, false, 1) + "</b>"
		}

		if r.Performed != "HELPLESS" {
			msg += "\nYou %s" + Prettfy(r.Performed, false, 1) + "%s"
			if r.Success {
				msg = fmt.Sprintf(msg, "", " successfully\nmeanwhile")
			} else {
				msg = fmt.Sprintf(msg, "tried to ", " but...")
			}
		}

		if enemy.Performed != "HELPLESS" {
			msg += "\nEnemy was " + Prettfy(enemy.Performed, true, 1)
		}

		if enemy.GainEffect != nil {
			msg += "\n<b>Enemy got " + Prettfy(*enemy.GainEffect, false, 1) + "</b>"
		}

		msg += "\n\n" + GenOffsetInfoBar(r.LifeOff, r.StaminaOff, enemy.LifeOff)

		b.SendMessage(msg, r.UserID, &opt)

		if !report.EndDuel {
			DisplayStatus(r.UserID, r.UserID, true)
		}
	}
}

// Display the current status of a user
func DisplayStatus(toUserID, ofUserID int64, newMessage bool) {
	var text string

	if toUserID == ofUserID {
		text = fmt.Sprint("🏷 <b>Your</b> current status:\n", genActionBar(ofUserID))
	} else {
		text = fmt.Sprint("👤 <b>", genUserLink(ofUserID, "Enemy"), "</b> current status:")
		if onGurad, _ := IsPlayerOnGuard(toUserID); onGurad {
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

// Generate the inline keyboard with all the actions
func genRematchKbd(opponentID int64, opponentName string) (markup *echotron.MessageReplyMarkup) {
	var kbd echotron.InlineKeyboardMarkup

	kbd.InlineKeyboard = [][]echotron.InlineKeyboardButton{{{
		Text:         "🔄 Rematch",
		CallbackData: fmt.Sprintf("/rematch %d %s", opponentID, opponentName),
	}}}

	return &echotron.MessageReplyMarkup{kbd}
}

// Notify the users that the duel is starting
func (b *bot) NotifyAcceptDuel(firstID, secondID int64) {
	var IDs = [2]int64{firstID, secondID}

	for i, currentID := range IDs {
		user := genUserLink(IDs[1-i], b.GetUserName(IDs[1-i]))
		b.SendMessage(
			fmt.Sprint("Duel against ", user, " is now starting 🏁"),
			currentID,
			&echotron.MessageOptions{ParseMode: echotron.HTML},
		)
		DisplayStatus(currentID, currentID, true)
	}
}

func (b *bot) NotifyDraw() {
	opponentID, err := GetOpponentID(b.chatID)
	if err != nil {
		log.Println("GetOpponentID", err)
		return
	}

	IDs := []int64{b.chatID, opponentID}
	for i, id := range IDs {
		enemyName := b.GetUserName(IDs[1-i])
		enemy := genUserLink(IDs[1-i], enemyName)
		res, _ := b.SendMessage(
			fmt.Sprint("⚖️ <b>The match is a draw</b> in the battle against ", enemy, "\n"),
			id,
			&echotron.MessageOptions{ParseMode: echotron.HTML},
		)
		b.EditMessageReplyMarkup(echotron.NewMessageID(id, res.Result.ID), genRematchKbd(IDs[1-i], enemyName))
	}
}

// Notify the users of the end of a match
func (b *bot) NotifyEndDuel(winnerID int64) {
	var opt = echotron.MessageOptions{ParseMode: echotron.HTML}

	looserID, err := GetOpponentID(winnerID)
	if err != nil {
		log.Println("NotifyEndDuel", "GetOpponentID", err)
		return
	}
	winnerName, looserName := b.GetUserName(winnerID), b.GetUserName(looserID)

	text := fmt.Sprint(
		"🥇 <b>You win</b> in the battle against ", genUserLink(looserID, looserName), "\n",
		"<i>Congratulation ", winnerName, " the big spirit of the war is proud of you</i>",
	)
	res, _ := b.SendMessage(text, winnerID, &opt)
	b.EditMessageReplyMarkup(echotron.NewMessageID(winnerID, res.Result.ID), genRematchKbd(looserID, looserName))

	text = fmt.Sprint(
		"☠ <b>You loose</b> the battle against ", genUserLink(winnerID, winnerName), "\n",
		"<i>I hope that the guardian spirit can assist you in the next battle</i>",
	)
	res, _ = b.SendMessage(text, looserID, &opt)
	b.EditMessageReplyMarkup(echotron.NewMessageID(looserID, res.Result.ID), genRematchKbd(winnerID, winnerName))
}

// Notify the users of the withdrawn of some of the two
func (b *bot) NotifyCancel() {
	var opt = echotron.MessageOptions{ParseMode: echotron.HTML}

	winnerID, err := GetOpponentID(b.chatID)
	if err != nil {
		log.Println("NotifyCancel", "GetOpponentID", err)
		return
	}

	text := fmt.Sprint(
		"🏳️ <b>You flee</b> from the battle against ", genUserLink(winnerID, b.GetUserName(winnerID)), "\n",
		"<i>The big spirit of the war will not like this behaviour...</i>",
	)
	b.SendMessage(text, b.chatID, &opt)

	text = fmt.Sprint(
		"🏃 <b>Your ", genUserLink(b.chatID, "opponent"), " has withdrawn</b>\n",
		"<i>Probably you are too strong for him or maybe he doesn't like your face...</i>",
	)
	b.SendMessage(text, winnerID, &opt)
}
