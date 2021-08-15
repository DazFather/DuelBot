package pg

// Action
const (
	HELPLESS Status = iota
	GUARD
	ATTACK
	DEFEND
	DODGE
)

//Effects
const (
	STUNNED Status = (iota + 1) * -1
	EXAUSTED
)

type InvokeRes struct {
	LifeOffset  int
	SpeedOffset int
	GainEffect  *Status
	Success     bool
	Performed   Status
}

// Create a new creature with standard stats
func NewCreature(power, stamina, hp int) Creature {
	c := Creature{
		damage: power,
		speed:  stamina,
		health: hp,
	}
	c.resetAction()
	return c
}

// Creature will try to prepare a certain action (choose between GUARD, ATTACK, DEFEND, DODGE)
func (c *Creature) SetAction(moves Status) {
	if c.isStunned() {
		return
	}
	if c.isExausted() && isIntensive(moves) {
		return
	}

	var toFunction = map[Status]invokeFn{
		ATTACK: c.attack,
		DEFEND: c.defend,
		DODGE:  c.dodge,
	}

	fn := toFunction[moves]
	if fn == nil {
		fn = c.sleep
	}
	c.perform = fn

	c.action = moves
}

func PerformAction(firstCreature, secondCreature *Creature) (winner int8, responses []InvokeRes) {
	responses = append(responses, firstCreature.perform(*secondCreature))
	responses = append(responses, secondCreature.perform(*firstCreature))
	responses[0].update(firstCreature.action, responses[1])
	responses[1].update(secondCreature.action, responses[0])

	firstCreature.resetAction()
	secondCreature.resetAction()

	p1Dead, p2Dead := firstCreature.isDead(), secondCreature.isDead()
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

func (c Creature) GetInfo() (life, agility int) {
	return c.health, c.speed
}

func (c Creature) GetStatus() (action Status, effects []Status) {
	effects = c.getEffects()
	if len(effects) > 0 {
		action = HELPLESS
		return
	}

	action = c.action

	return
}
