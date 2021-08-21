package main

import (
	"errors"
	"time"

	"DuelBot/pg"
)

type Player struct {
	name    string
	stats   pg.Creature
	menuID  int
	enemyID int64
}

var (
	Players = make(map[int64]*Player, 0)

	toString = map[pg.Status]string{
		pg.GUARD:    "GUARD",
		pg.ATTACK:   "ATTACK",
		pg.DEFEND:   "DEFEND",
		pg.DODGE:    "DODGE",
		pg.STUNNED:  "STUNNED",
		pg.EXAUSTED: "EXAUSTED",
		pg.HELPLESS: "GUARD",
	}
)

func stdCreature() pg.Creature {
	const (
		defAttack     = 5
		defStamina    = 6
		defMaxStamina = 10
		defHealth     = 20
	)
	return pg.NewCreature(defAttack, defStamina, defMaxStamina, defHealth)
}

func SelectPlayer(userID int64) *Player {
	return Players[userID]
}

func IsPlayerReady(ownerID int64) bool {
	return !IsPlayerOnGuard(ownerID)
}

func IsPlayerOnGuard(ownerID int64) bool {
	status, _ := Players[ownerID].stats.GetStatus()
	return status == pg.GUARD || status == pg.HELPLESS
}

func AddNewPlayer(ownerID int64, playerName string) {
	Players[ownerID] = &Player{
		name:   playerName,
		stats:  stdCreature(),
		menuID: -1,
	}
	return
}

func GetPlayerMenuID(ownerID int64) (menuID int) {
	return Players[ownerID].menuID
}

func SetPlayerMenuID(ownerID int64, newMenuID int) {
	Players[ownerID].menuID = newMenuID
}

func SetPlayerMoves(ownerID int64, move string) (time.Duration, error) {
	var toStatus = map[string]pg.Status{
		"GUARD":  pg.GUARD,
		"ATTACK": pg.ATTACK,
		"DEFEND": pg.DEFEND,
		"DODGE":  pg.DODGE,
	}

	action := toStatus[move]
	if action == pg.HELPLESS {
		return time.Duration(0), errors.New("Invalid player move: \"" + move + "\"")
	}

	return Players[ownerID].stats.SetAction(action), nil
}

func EngageDuel(firstOwnerID, secondOwnerID int64) bool {
	if IsPlayerBusy(firstOwnerID) || IsPlayerBusy(secondOwnerID) {
		return false
	}

	Players[firstOwnerID].enemyID = secondOwnerID
	Players[secondOwnerID].enemyID = firstOwnerID
	return true
}

func sendNotifications(ownerID int64, responses [2]pg.InvokeRes) {
	var IDs = []int64{ownerID, Players[ownerID].enemyID}
	var moves [2]string

	for i, r := range responses {
		if r.HasGotEffected() {
			moves[i] = toString[r.GainEffect]
			continue
		}
		moves[i] = toString[r.Performed]
	}
	for i, subj := range responses {
		NotifyResult(
			IDs[i],
			subj.Success,
			moves[i],
			moves[1-i],
			GenOffsetInfoBar(
				subj.LifeOffset,
				subj.StaminaOffset,
				responses[1-i].LifeOffset,
			),
		)
	}
}

func PlayersPerformMove(ownerID int64) (end bool, winnerID int64, err error) {
	opponentID := GetOpponentID(ownerID)
	if Players[opponentID] == nil {
		return true, 0, errors.New("Invalid opponent")
	}

	winFlag, responses := pg.PerformAction(&Players[ownerID].stats, &Players[opponentID].stats)
	sendNotifications(ownerID, responses)
	switch winFlag {
	case 0: // Match is still going...
		return
	//case -1: Everyone died match is a draw
	case 1: // Player 1 (owner) win
		winnerID = ownerID
	case 2: // Player 2 (opponent) win
		winnerID = opponentID
	}
	end = true
	return
}

func IsPlayerWinner(ownerID int64) bool {
	health, _, _ := GetOpponentOf(ownerID).stats.GetInfo()
	return health <= 0
}

func EndDuel(userID int64) {
	GetOpponentOf(userID).refresh()
	SelectPlayer(userID).refresh()
}

func (p *Player) refresh() {
	p.stats = stdCreature()
	p.enemyID = 0
	p.menuID = -1
}

func GetOpponentOf(ownerID int64) *Player {
	return Players[Players[ownerID].enemyID]
}

func GetOpponentID(userID int64) int64 {
	return Players[userID].enemyID
}

func GetPlayerInfo(ownerID int64) (life, agility, maxStamina int) {
	return Players[ownerID].stats.GetInfo()
}

func GetPlayerName(ownerID int64) string {
	return Players[ownerID].name
}

func GetPlayerAction(ownerID int64) string {
	current, effects := Players[ownerID].stats.GetStatus()

	if current == pg.HELPLESS && len(effects) > 0 {
		current = effects[0]
	}
	return toString[current]
}

func IsPlayerBusy(ownerID int64) bool {
	if Players[ownerID] != nil {
		return Players[ownerID].enemyID != 0
	}
	return true
}
