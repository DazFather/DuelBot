package pg

import (
	"time"
)

// Create a new creature with standard stats
func NewCreature(power, stamina, maxStamina, health int) Creature {
	c := Creature{
		damage:     power,
		stamina:    stamina,
		maxStamina: maxStamina,
		hp:         health,
	}
	c.resetAction()
	return c
}

func PerformAction(c1, c2 *Creature) (winner int8, responses [2]InvokeRes) {
	responses[0] = c1.perform(*c2)
	responses[1] = c2.perform(*c1)
	responses[0].update(responses[1])
	responses[1].update(responses[0])

	c1.resetAction()
	c2.resetAction()

	p1Dead, p2Dead := c1.isDead(), c2.isDead()
	switch true {
	case !p1Dead && !p2Dead:
		winner = 0
	case p1Dead && p2Dead:
		winner = -1
	case p1Dead:
		winner = 2
	case p2Dead:
		winner = 1
	}
	return
}

// Creature will try to prepare a certain action (choose between GUARD, ATTACK, DEFEND, DODGE)
func (c *Creature) SetAction(action Status) time.Duration {
	if c.isOnStatus(STUNNED) || c.isOnStatus(EXAUSTED) {
		return c.duration
	}

	if action == GUARD {
		c.duration = 0 * time.Second
	} else if isEnergyIntensive(action) {
		c.duration = calcSpeed(c.stamina, c.maxStamina)
	} else {
		c.duration = 1 * time.Second
	}

	c.action = action

	return c.duration
}

func (c Creature) GetInfo() (life, agility, maxStamina int) {
	life = c.hp
	agility = c.stamina
	maxStamina = c.maxStamina

	return
}

func (c Creature) GetStatus() (action Status, effects []Status) {
	action = c.action
	for _, eff := range c.effects {
		effects = append(effects, eff.symptom)
	}

	return
}

func (r InvokeRes) HasGotEffected() bool {
	return r.Performed == HELPLESS && r.GainEffect != HELPLESS
}
