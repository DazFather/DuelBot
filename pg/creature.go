package pg

type Status int8

type Creature struct {
	damage  int
	health  int
	speed   int
	action  Status
	perform invokeFn
}

const MAX_SPEED = 10

func isEffectStatus(actionOrEffect Status) bool {
	return actionOrEffect < 0
}

func isIntensive(action Status) bool {
	return action > DEFEND
}

func (c *Creature) useEnergy(response *InvokeRes) (isExausted bool) {
	if c.speed <= 0 {
		response.GainEffect = fatigue(c)
		return true
	}
	c.speed -= 1
	response.SpeedOffset -= 1
	return false
}

func (c *Creature) gainEnergy(response *InvokeRes) {
	if c.speed < MAX_SPEED {
		c.speed += 1
		response.SpeedOffset += 1
	}
}

func stun(c *Creature) *Status {
	c.action = STUNNED
	c.perform = c.sleep
	return &c.action
}

func fatigue(c *Creature) *Status {
	c.action = EXAUSTED
	c.perform = c.sleep
	return &c.action
}

func (c Creature) isStunned() bool {
	return c.action == STUNNED
}

func (c Creature) isExausted() bool {
	return c.action == EXAUSTED || c.speed <= 0
}

func (c Creature) isDead() bool {
	return c.health <= 0
}

func (c Creature) isReady() bool {
	return c.perform != nil
}

func (c *Creature) resetAction() {
	c.perform = c.sleep
	c.action = GUARD
}

func (c Creature) getEffects() []Status {
	var eff []Status

	if c.isStunned() {
		eff = append(eff, STUNNED)
	}
	if c.isExausted() {
		eff = append(eff, EXAUSTED)
	}

	return nil
}

func (res *InvokeRes) update(action Status, enemyRessponse InvokeRes) {
	res.Performed = action

	if isEffectStatus(action) {
		res.Success = false
		return
	}

	switch action {
	case ATTACK:
		res.Success = enemyRessponse.LifeOffset < 0
	case DEFEND:
		res.Success = (enemyRessponse.GainEffect != nil || res.LifeOffset < 0)
	case DODGE:
		res.Success = res.LifeOffset == 0
	}
}
