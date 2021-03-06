package pg

import (
	"time"
)

func (c *Creature) useEnergy(response *InvokeRes) (isExausted bool) {
	if c.stamina <= 0 {
		response.GainEffect = fatigue(c)
		return true
	}
	c.stamina -= 1
	response.StaminaOffset -= 1
	return
}

func (c *Creature) gainEnergy(response *InvokeRes) {
	if c.stamina < c.maxStamina {
		c.stamina += 1
		response.StaminaOffset += 1
	}
}

func calcSpeed(stamina, max uint) (speed time.Duration) {
	seconds := max + 1 - stamina
	speed = time.Duration(seconds) * time.Second
	return
}

func (c *Creature) resetAction() {
	c.action = HELPLESS
}

func isEffect(actionOrEffect Status) bool {
	return actionOrEffect < 0
}

func isEnergyIntensive(action Status) bool {
	return action > DEFEND
}

func stun(c *Creature) Status {
	c.effects = append(c.effects, effect{symptom: STUNNED, turns: 1})
	c.action = HELPLESS
	c.duration = 0 * time.Second

	return STUNNED
}

func fatigue(c *Creature) Status {
	c.effects = append(c.effects, effect{symptom: EXAUSTED, turns: 1})
	c.action = HELPLESS
	c.duration = 0 * time.Second

	return EXAUSTED
}

func (c *Creature) reduceEffects() {
	var newEffects []effect

	if len(c.effects) == 0 {
		return
	}

	for _, eff := range c.effects {
		if eff.turns > 0 {
			eff.turns--
			newEffects = append(newEffects, eff)
		}
	}

	c.effects = newEffects
}

func (c *Creature) perform(enemy Creature) (response InvokeRes) {
	switch c.action {
	case ATTACK:
		response = c.attack(enemy)
	case DEFEND:
		response = c.defend(enemy)
	case DODGE:
		response = c.dodge(enemy)
	case GUARD, HELPLESS:
		response = c.sleep(enemy)
		// I'm not sure if I should delete this or not...
		if len(c.effects) > 0 {
			if c.IsOnStatus(STUNNED) {
				response.Performed = STUNNED
			} else if c.IsOnStatus(EXAUSTED) {
				response.Performed = EXAUSTED
			}
			c.reduceEffects()
			return
		}
	}
	c.reduceEffects()

	response.Performed = c.action
	return
}

func (attacking *Creature) attack(enemy Creature) (response InvokeRes) {
	attacking.useEnergy(&response)

	if enemy.action == ATTACK {
		attacking.hp -= int(enemy.damage)
		response.LifeOffset = -int(enemy.damage)
	}

	return
}

func (defending *Creature) defend(enemy Creature) (response InvokeRes) {
	if enemy.action == ATTACK {
		response.LifeOffset = -int(enemy.damage) / 2
		defending.hp += response.LifeOffset
		return
	}
	defending.gainEnergy(&response)

	return
}

func (dodging *Creature) dodge(enemy Creature) (response InvokeRes) {
	if dodging.stamina < enemy.stamina {
		switch enemy.action {
		case ATTACK:
			dodging.hp -= int(enemy.damage)
			response.LifeOffset = -int(enemy.damage)
		case DEFEND:
			response.GainEffect = stun(dodging)
		}
	}

	dodging.useEnergy(&response)

	return
}

func (sleeping *Creature) sleep(enemy Creature) (response InvokeRes) {
	switch enemy.action {
	case ATTACK:
		sleeping.hp -= int(enemy.damage)
		response.LifeOffset = -int(enemy.damage)
	case DEFEND:
		if !sleeping.IsOnStatus(STUNNED) {
			response.GainEffect = stun(sleeping)
		}
	}
	sleeping.gainEnergy(&response)

	return
}
