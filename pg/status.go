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

func calcSpeed(stamina, max int) (speed time.Duration) {
	seconds := max + 1 - stamina
	speed = time.Duration(seconds) * time.Second
	return
}

func (c *Creature) resetAction() {
	c.action = HELPLESS
}

func (c Creature) isDead() bool {
	return c.hp <= 0
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

func (c Creature) isOnStatus(actionOrEffect Status) bool {
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
		if len(c.effects) > 0 {
			if c.isOnStatus(STUNNED) {
				response.Performed = STUNNED
			} else if c.isOnStatus(EXAUSTED) {
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

func (res *InvokeRes) update(enemyRessponse InvokeRes) {
	if isEffect(res.Performed) {
		res.Success = false
		return
	}

	switch res.Performed {
	case ATTACK:
		res.Success = enemyRessponse.LifeOffset < 0
	case DEFEND:
		res.Success = (enemyRessponse.GainEffect != HELPLESS || res.LifeOffset < 0)
	case DODGE:
		res.Success = res.LifeOffset == 0
	}
}

func (attacking *Creature) attack(enemy Creature) (response InvokeRes) {
	attacking.useEnergy(&response)

	if enemy.action == ATTACK {
		attacking.hp -= enemy.damage
		response.LifeOffset = -enemy.damage
	}

	return
}

func (defending *Creature) defend(enemy Creature) (response InvokeRes) {
	if enemy.action == ATTACK {
		response.LifeOffset = -enemy.damage / 2
		defending.hp += response.LifeOffset
		return
	}
	defending.gainEnergy(&response)

	return
}

func (dodging *Creature) dodge(enemy Creature) (response InvokeRes) {
	if dodging.duration > enemy.duration {
		switch enemy.action {
		case ATTACK:
			dodging.hp -= enemy.damage
			response.LifeOffset = -enemy.damage
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
		sleeping.hp -= enemy.damage
		response.LifeOffset = -enemy.damage
	case DEFEND:
		if !sleeping.isOnStatus(STUNNED) {
			response.GainEffect = stun(sleeping)
		}
	}
	sleeping.gainEnergy(&response)

	return
}
