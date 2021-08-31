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

const defAction = pg.GUARD

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

	toStatus = reverseMap(toString)
)

func reverseMap(original map[pg.Status]string) (reversed map[string]pg.Status) {
	reversed = make(map[string]pg.Status)
	for key, val := range original {
		reversed[val] = key
	}
	return
}

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
	current, _, _ := players[ownerID].stats.GetStatus()

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

// Check if a player exist and is busy on a duel or not
func IsPlayerBusy(ownerID int64) bool {
	if isPlayerExistent(ownerID) {
		return players[ownerID].enemyID != 0
	}
	return false
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
		return !players[ownerID].stats.IsOnStatus(defAction), nil
	}
	err = errors.New("Player does not exist")
	return
}

// Check if player actually exist
func isPlayerExistent(ownerID int64) bool {
	return players[ownerID] != nil
}

// Check if an action was executed successfully
func isSuccessfull(response, enemyRessponse pg.InvokeRes) bool {
	switch response.Performed {
	case pg.ATTACK:
		return enemyRessponse.LifeOffset < 0
	case pg.DEFEND:
		return enemyRessponse.GainEffect != pg.HELPLESS || response.LifeOffset < 0
	case pg.DODGE:
		return response.LifeOffset == 0
	}

	// in case of pg.GUARD, pg.HELPLESS, pg.STUNNED, pg.EXAUSTED:
	return false
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
func SetPlayerMoves(ownerID int64, move string) (duration time.Duration, err error) {
	var action = toStatus[move]

	// Chek if player is on a fight
	if !IsPlayerBusy(ownerID) {
		return time.Duration(0), errors.New("Player does not exist or is not in a duel")
	}

	// Don't do anything if is the same action
	if players[ownerID].stats.IsOnStatus(action) {
		return
	}

	// If player is STUNNED or EXAUSTED or some type of effects that put the pg KO
	if players[ownerID].stats.IsOnStatus(pg.HELPLESS) {
		return time.Duration(0), errors.New("Unable to set moves, player cannot fight")
	}

	// Setting player action
	duration, err = players[ownerID].stats.SetAction(action)
	if err != nil {
		return time.Duration(0), err
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

// Engage a duel between two players saving them on the register
func EngageDuel(firstOwnerID, secondOwnerID int64) bool {
	if IsPlayerBusy(firstOwnerID) || IsPlayerBusy(secondOwnerID) {
		return false
	}

	AddNewPlayer(firstOwnerID, secondOwnerID)
	players[firstOwnerID].stats.SetAction(defAction)
	AddNewPlayer(secondOwnerID, firstOwnerID)
	players[secondOwnerID].stats.SetAction(defAction)
	return true
}

// Generating the report
func genReport(firstOwnerID, secondOwnerID int64, winFlag int8, responses [2]pg.InvokeRes) BattleReport {
	var report BattleReport

	// Filling match details
	for i, res := range responses {
		report.PlayersInfo = append(report.PlayersInfo, PlayerReport{
			LifeOff:    res.LifeOffset,
			StaminaOff: res.StaminaOffset,
			Performed:  toString[res.Performed],
			Success:    isSuccessfull(res, responses[1-i]),
		})
		// Check if during the battle a creature got a new effect
		if res.GainEffect != pg.HELPLESS {
			val := toString[res.GainEffect]
			report.PlayersInfo[i].GainEffect = &val
		}
	}

	// Filling the owners IDs
	report.PlayersInfo[0].UserID = firstOwnerID
	report.PlayersInfo[1].UserID = secondOwnerID

	// Filling end-duel related infos
	switch winFlag {
	case 0: // Match is still going...
		return report
	//case -1: Everyone died match is a draw
	case 1, 2: // Player 1 (owner) or Player 2 (opponent) win
		report.WinnerID = &report.PlayersInfo[winFlag-1].UserID
	}
	report.EndDuel = true

	return report
}

// Execute the action of a player against his opponent and vice versa
func PlayersPerformMove(ownerID, opponentID int64) (report BattleReport, err error) {
	var (
		ready bool
		// Grab the action of the player and it's duration
		action, duration, _ = players[ownerID].stats.GetStatus()
	)

	switch action {
	case pg.DEFEND:
		// If enemy is ready players will perform the action or else
		ready, err = IsPlayerReady(opponentID)
		if err != nil {
			return
		}
		if !ready {
			return
		}
	case pg.ATTACK, pg.DODGE:
		// Loop looking for a change in the enemy status
		start := time.Now()
		for time.Since(start).Milliseconds() < duration.Milliseconds() {
			// if enemy is ready stop the loop
			ready, err = IsPlayerReady(opponentID)
			if err != nil {
				return
			}
			if ready {
				break
			}

			time.Sleep(time.Duration(500) * time.Millisecond)
			// if the user change action quit this process
			if current, _, _ := players[ownerID].stats.GetStatus(); action != current {
				return
			}
		}
	default:
		return
	}

	// Perform the action between players and generate the BattleReport
	winFlag, responses := pg.PerformAction(&players[ownerID].stats, &players[opponentID].stats)
	report = genReport(ownerID, opponentID, winFlag, responses)
	if winFlag == 0 {
		// Set players on default action
		players[ownerID].stats.SetAction(defAction)
		players[opponentID].stats.SetAction(defAction)
	}

	return
}
