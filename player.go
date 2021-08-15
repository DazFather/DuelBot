package main

import (
	"errors"

	"DuelBot/pg"
)

const (
	STD_ATTACK  = 5
	STD_STAMINA = 6
	STD_HEALTH  = 20
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
		pg.HELPLESS: "HELPLESS",
	}
)

func stdCreature() pg.Creature {
	return pg.NewCreature(STD_ATTACK, STD_STAMINA, STD_HEALTH)
}

func SelectPlayer(userID int64) *Player {
	return Players[userID]
}

func (p Player) isBusy() bool {
	return p.enemyID != 0
}

func (p Player) isOnGuard() bool {
	status, _ := p.stats.GetStatus()
	return status == pg.GUARD
}

func IsPlayerOnGuard(ownerID int64) bool {
	return Players[ownerID].isOnGuard()
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

func SetPlayerMoves(ownerID int64, action string) {
	var toStatus = map[string]pg.Status{
		"GUARD":  pg.GUARD,
		"ATTACK": pg.ATTACK,
		"DEFEND": pg.DEFEND,
		"DODGE":  pg.DODGE,
	}

	Players[ownerID].stats.SetAction(toStatus[action])
}

func EngageDuel(firstOwnerID, secondOwnerID int64) bool {
	if Players[firstOwnerID].isBusy() || Players[secondOwnerID].isBusy() {
		return false
	}

	Players[firstOwnerID].enemyID = secondOwnerID
	Players[secondOwnerID].enemyID = firstOwnerID
	return true
}

func sendNotifications(ownerID int64, responses []pg.InvokeRes) {
	var IDs = []int64{ownerID, Players[ownerID].enemyID}

	for i, subj := range responses {

		NotifyResult(
			IDs[i],
			subj.Success,
			toString[subj.Performed],
			toString[responses[len(IDs)-1-i].Performed],
			GenOffsetInfoBar(subj.LifeOffset, subj.SpeedOffset),
		)
	}
}

func PlayersPerformMove(ownerID int64) (end bool, err error) {
	opponent := GetOpponentOf(ownerID)
	if opponent == nil {
		return true, errors.New("Invalid opponent")
	}

	winFlag, responses := pg.PerformAction(&Players[ownerID].stats, &opponent.stats)
	sendNotifications(ownerID, responses)

	return winFlag != 0, nil
}

func IsPlayerWinner(ownerID int64) bool {
	health, _ := GetOpponentOf(ownerID).stats.GetInfo()
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

func GetPlayerInfo(ownerID int64) (life, agility int) {
	return Players[ownerID].stats.GetInfo()
}

func GetPlayerName(ownerID int64) string {
	return Players[ownerID].name
}

func GetPlayerAction(ownerID int64) string {
	current, effects := Players[ownerID].stats.GetStatus()

	if current == pg.HELPLESS {
		if val := toString[effects[0]]; val != "" {
			return val
		}
	}
	return toString[current]
}

func IsPlayerBusy(ownerID int64) bool {
	return Players[ownerID].isBusy()
}
