package pg

type invokeFn func(Creature) InvokeRes

func (attacking *Creature) attack(enemy Creature) (response InvokeRes) {
	attacking.useEnergy(&response)

	if enemy.action == ATTACK {
		attacking.health -= enemy.damage
		response.LifeOffset = -enemy.damage
	}

	return
}

func (defending *Creature) defend(enemy Creature) (response InvokeRes) {
	if enemy.action == ATTACK {
		response.LifeOffset = -enemy.damage / 2
		defending.health += response.LifeOffset
		return
	}
	defending.gainEnergy(&response)

	return
}

func (dodging *Creature) dodge(enemy Creature) (response InvokeRes) {
	if dodging.speed <= enemy.speed {
		switch enemy.action {
		case ATTACK:
			dodging.health -= enemy.damage
			response.LifeOffset = -enemy.damage
		case DEFEND:
			response.GainEffect = stun(dodging)
		}
	}

	dodging.useEnergy(&response)

	return
}

func (c *Creature) sleep(enemy Creature) (response InvokeRes) {
	switch enemy.action {
	case ATTACK:
		c.health -= enemy.damage
		response.LifeOffset = -enemy.damage
	case DEFEND:
		if !c.isStunned() {
			response.GainEffect = stun(c)
		}
	}
	c.gainEnergy(&response)

	return
}
