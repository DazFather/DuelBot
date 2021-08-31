package pg

import (
	"time"
)

type Status int8

// Action
const (
	HELPLESS Status = iota // default status: Do nothing
	GUARD
	DEFEND
	ATTACK
	DODGE
)

// Symptoms - Effects
const (
	STUNNED Status = (iota + 1) * -1
	EXAUSTED
)

// Response after two creature fight each other
type InvokeRes struct {
	LifeOffset    int    // Difference of health points
	StaminaOffset int    // Difference of stamina points
	GainEffect    Status // If creature got a new effect (default: HELPLESS)
	Performed     Status // Performed action (default: HELPLESS)
}

// Effect
type effect struct {
	symptom Status // type of effect
	turns   int8   // how many "turns" will last
}

// Creature
type Creature struct {
	hp         int           // health points
	damage     uint          // (constant) max damage that is capable of dealing
	stamina    uint          // how fast it is. It fill influence the duration
	maxStamina uint          // max level of stamina
	action     Status        // action he is doing
	duration   time.Duration // duration of the action
	effects    []effect      // list of effects
}
