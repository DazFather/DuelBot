package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Register user chatID -> inviteID
var invites = make(map[int64][]rune, 0)

// Get an inviteID in []rune
func getInviteID(userID int64) []rune {
	return invites[userID]
}

// Set a inviteID of a user to a new value
func setInviteID(userID int64, newInviteID []rune) {
	invites[userID] = newInviteID
}

// Increment the passed char (must be in range: 0-9, A-Y, a-y)
func incrChar(char *rune) {
	switch *char {
	case '9':
		*char = 'A'
	case 'Z':
		*char = 'a'
	default:
		*char++
	}
}

// Return a copy of the inviteID with 3 random chars before and after
func addRandChars(rawInviteID []rune) (inviteID []rune) {
	var chunk [2][3]rune

	rand.Seed(time.Now().UnixNano())
	for j := 0; j < 2; j++ {
		for i := 0; i < 3; i++ {
			switch rand.Intn(3) {
			case 0:
				chunk[j][i] = rune(rand.Intn(10) + 48)
			case 1:
				chunk[j][i] = rune(rand.Intn(26) + 65)
			case 2:
				chunk[j][i] = rune(rand.Intn(26) + 97)
			}
		}
	}
	inviteID = append(inviteID, chunk[0][:]...)
	inviteID = append(inviteID, rawInviteID...)
	inviteID = append(inviteID, chunk[1][:]...)
	return
}

// increment the passed inviteID
func incrInviteID(inviteID []rune) {
	var i = len(inviteID) - 1

	for ; i >= 0; i-- {
		if inviteID[i] != 'z' {
			incrChar(&inviteID[i])
			break
		}
		inviteID[i] = '0'
	}
	if i == 0 && inviteID[i] == '0' {
		inviteID = append([]rune{'1'}, inviteID...)
	}
}

// Generate and return a new inviteID as a string
func NewInviteID(userID int64) string {
	var inviteID = getInviteID(userID)
	if inviteID == nil {
		inviteID = []rune{'0'}
	} else {
		inviteID = inviteID[3 : len(inviteID)-3]
	}

	incrInviteID(inviteID)
	inviteID = addRandChars(inviteID)
	setInviteID(userID, inviteID)
	return string(inviteID)
}

// Check the validity of an inviteID of a user
func isValidInviteID(userID int64, inviteID string) bool {
	return string(getInviteID(userID)) == inviteID
}

// Check the validity of a Invitation, errorMessage == "" if is all okay
func (b *bot) IsInvitionValid(chatID int64, inviteID string) (errorMessage string) {
	var botID int64
	if res, err := b.GetMe(); err != nil {
		botID = res.Result.ID
	}

	switch true {
	case b.chatID == chatID:
		errorMessage = "This link is not for you. Send it to who you want to duel"

	case botID == chatID:
		errorMessage = "Thanks for the invitation but I'm quite busy now. Maybe another time"

	case !isValidInviteID(chatID, inviteID):
		errorMessage = "This link is expired"

	case IsPlayerBusy(chatID):
		errorMessage = "Your opponent might be already engaged in another fight. Brawls are still not allowed"

	case IsPlayerBusy(b.chatID):
		errorMessage = "You are already engaged in another fight. Brawls are still not allowed"
	}

	return
}

// Generate a new invitiation link
func (b *bot) GenInvitationLink() string {
	var botUser string
	if res, err := b.GetMe(); err == nil {
		botUser = res.Result.Username
	}
	return fmt.Sprint("https://t.me/", botUser, "?start=joinDuel_", b.chatID, "_", NewInviteID(b.chatID))
}
