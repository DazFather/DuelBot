package pg

import (
	"errors"
	"time"
)

// Create a new creature
func NewCreature(power, stamina, maxStamina, health uint) Creature {
	c := Creature{
		damage:     power,
		stamina:    stamina,
		maxStamina: maxStamina,
		hp:         int(health),
	}
	c.resetAction()
	return c
}

// Chek if a creature is dead
func (c Creature) IsDead() bool {
	return c.hp <= 0
}

/* Two creature perform their actions aginst each other and it returns:
 * winner - a flag who indicates the winner creature (0 -> none, -1 -> draw, 1 -> c1, 2 -> c2)
 * responses - the responses of the actions performed (c1 -> responses[0], c2 -> responses[1])
 */
func PerformAction(c1, c2 *Creature) (winner int8, responses [2]InvokeRes) {
	responses[0] = c1.perform(*c2)
	responses[1] = c2.perform(*c1)
	responses[0].update(responses[1])
	responses[1].update(responses[0])

	c1.resetAction()
	c2.resetAction()

	p1Dead, p2Dead := c1.IsDead(), c2.IsDead()
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
func (c *Creature) SetAction(action Status) (time.Duration, error) {
	if c.IsOnStatus(STUNNED) || c.IsOnStatus(EXAUSTED) {
		return c.duration, nil
	}

	switch action {
	case GUARD:
		c.duration = time.Duration(0)
	case ATTACK, DODGE:
		c.duration = calcSpeed(c.stamina, c.maxStamina)
	case DEFEND:
		c.duration = 1 * time.Second
	default:
		return time.Duration(0), errors.New("Invalid action")
	}

	c.action = action
	return c.duration, nil
}

// Get the stats of a creature
func (c Creature) GetInfo() (life int, agility, maxStamina, damage uint) {
	life = c.hp
	agility = c.stamina
	maxStamina = c.maxStamina
	damage = c.damage

	return
}

// Get the creature setted action and effects
func (c Creature) GetStatus() (action Status, effects []Status) {
	action = c.action
	for _, eff := range c.effects {
		effects = append(effects, eff.symptom)
	}

	return
}

// Check if a creature is on particular status (can be action or effect)
func (c Creature) IsOnStatus(actionOrEffect Status) bool {
	if isEffect(actionOrEffect) {
		for _, eff := range c.effects {
			if eff.symptom == actionOrEffect {
				return true
			}
		}
		return false
	}

	return c.action == actionOrEffect
}
