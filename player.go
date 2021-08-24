package main

import (
	"errors"
	"time"

	"DuelBot/pg"
)

type Player struct {
	stats   pg.Creature
	menuID  int
	enemyID int64
}

var (
	players = make(map[int64]*Player, 0)

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

// It adds a player to the register
func AddNewPlayer(ownerID, enemyID int64) {
	const (
		defAttack     = 5
		defStamina    = 6
		defMaxStamina = 10
		defHealth     = 20
	)

	players[ownerID] = &Player{
		stats:   pg.NewCreature(defAttack, defStamina, defMaxStamina, defHealth),
		menuID:  -1,
		enemyID: enemyID,
	}
	return
}

// It ends the duel and clean the values from the register
func EndDuel(userID int64) error {
	if !IsPlayerBusy(userID) {
		return errors.New("Player is not in a duel")
	}
	oppID, err := GetOpponentID(userID)
	if err != nil {
		return err
	}

	delete(players, oppID)
	delete(players, userID)
	return nil
}

// Check if player actually exist
func isPlayerExistent(ownerID int64) bool {
	return players[ownerID] != nil
}

// Check if a player exist and is busy on a duel or not
func IsPlayerBusy(ownerID int64) bool {
	if isPlayerExistent(ownerID) {
		return players[ownerID].enemyID != 0
	}
	return false
}

// Engage a duel between two players saving them on the register
func EngageDuel(firstOwnerID, secondOwnerID int64) bool {
	if IsPlayerBusy(firstOwnerID) || IsPlayerBusy(secondOwnerID) {
		return false
	}

	AddNewPlayer(firstOwnerID, secondOwnerID)
	players[firstOwnerID].stats.SetAction(pg.GUARD)
	AddNewPlayer(secondOwnerID, firstOwnerID)
	players[secondOwnerID].stats.SetAction(pg.GUARD)
	return true
}

// Get the message ID of the last menu
func GetPlayerMenuID(ownerID int64) (menuID int, err error) {
	if isPlayerExistent(ownerID) {
		return players[ownerID].menuID, nil
	}
	err = errors.New("Player does not exist")
	return
}

// Get the stats of a player
func GetPlayerInfo(ownerID int64) (life int, agility, maxStamina, damage uint, err error) {
	if isPlayerExistent(ownerID) {
		life, agility, maxStamina, damage = players[ownerID].stats.GetInfo()
		return
	}
	err = errors.New("Player does not exist")
	return
}

// Get the move that a player is going to execute / has already executed
func GetPlayerAction(ownerID int64) (move string, err error) {
	if !isPlayerExistent(ownerID) {
		err = errors.New("Player does not exist")
		return
	}
	current, _ := players[ownerID].stats.GetStatus()

	if current == pg.HELPLESS {
		if players[ownerID].stats.IsOnStatus(pg.STUNNED) {
			current = pg.STUNNED
		} else if players[ownerID].stats.IsOnStatus(pg.EXAUSTED) {
			current = pg.EXAUSTED
		}
	}

	return toString[current], nil
}

// Get the enemy chatID of a player
func GetOpponentID(userID int64) (opponentID int64, err error) {
	if IsPlayerBusy(userID) {
		return players[userID].enemyID, nil
	}
	err = errors.New("Original player is not in duel")
	return
}

// Check if a player is on guard (return err if does not exist)
func IsPlayerOnGuard(ownerID int64) (onGurad bool, err error) {
	if IsPlayerBusy(ownerID) {
		return players[ownerID].stats.IsOnStatus(pg.GUARD), nil
	}
	err = errors.New("Player does not exist")
	return
}

// Check if a player setted his moves
func IsPlayerReady(ownerID int64) (onGurad bool, err error) {
	if IsPlayerBusy(ownerID) {
		return !players[ownerID].stats.IsOnStatus(pg.GUARD), nil
	}
	err = errors.New("Player does not exist")
	return
}

// Set a new value for the message ID of the menu in use
func SetPlayerMenuID(ownerID int64, newMenuID int) (err error) {
	if !IsPlayerBusy(ownerID) {
		return errors.New("Player does not exist or is not in a duel")
	}
	players[ownerID].menuID = newMenuID
	return
}

/* Set the player moves (GUARD, ATTACK, DEFEND or DODGE) and return it's duration.
* Error if unable to perform (STUNNED or EXAUSTED) or if the move is invalid
 */
func SetPlayerMoves(ownerID int64, move string) (time.Duration, error) {
	var action pg.Status

	if !IsPlayerBusy(ownerID) {
		return time.Duration(0), errors.New("Player does not exist or is not in a duel")
	}

	// If player is STUNNED or EXAUSTED or some type of effects that put the pg KO
	if players[ownerID].stats.IsOnStatus(pg.HELPLESS) {
		return time.Duration(0), errors.New("Unable to set moves, player cannot fight")
	}

	switch move {
	case "GUARD":
		action = pg.GUARD
	case "ATTACK":
		action = pg.ATTACK
	case "DEFEND":
		action = pg.DEFEND
	case "DODGE":
		action = pg.DODGE
	default:
		return time.Duration(0), errors.New("Invalid player move: \"" + move + "\"")
	}

	return players[ownerID].stats.SetAction(action)
}

type BattleReport struct {
	EndDuel     bool
	WinnerID    *int64
	PlayersInfo []PlayerReport
}

type PlayerReport struct {
	UserID     int64
	LifeOff    int
	StaminaOff int
	GainEffect *string
	Performed  string
	Success    bool
}

func PlayersPerformMove(ownerID int64) (report BattleReport, err error) {
	var opponentID int64

	opponentID, err = GetOpponentID(ownerID)
	if err != nil {
		return
	}

	winFlag, responses := pg.PerformAction(&players[ownerID].stats, &players[opponentID].stats)

	// Generating the report
	for i, res := range responses {
		report.PlayersInfo = append(report.PlayersInfo, PlayerReport{
			LifeOff:    res.LifeOffset,
			StaminaOff: res.StaminaOffset,
			Performed:  toString[res.Performed],
			Success:    res.Success,
		})
		// Check if during the battle a creature got a new effect
		if res.Performed == pg.HELPLESS && res.GainEffect != pg.HELPLESS {
			val := toString[res.GainEffect]
			report.PlayersInfo[i].GainEffect = &val
		}
	}
	report.PlayersInfo[0].UserID = ownerID
	report.PlayersInfo[1].UserID = opponentID

	players[ownerID].stats.SetAction(pg.GUARD)
	players[opponentID].stats.SetAction(pg.GUARD)

	switch winFlag {
	case 0: // Match is still going...
		return

	//case -1: Everyone died match is a draw

	case 1, 2: // Player 1 (owner) or Player 2 (opponent) win
		report.WinnerID = &report.PlayersInfo[winFlag-1].UserID
	}
	report.EndDuel = true

	return
}
