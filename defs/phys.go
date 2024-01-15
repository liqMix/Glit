package defs

type Move struct {
	x, y   float64
	vX, vY float64
}

type Action struct {
	Thing  Thingi
	Move   *Move
	Glitch bool
}

func HasCollision(a, b Thingi, move *Move) bool {
	if move == nil {
		return false
	}
	if a == nil || b == nil {
		return false
	}

	// Check for collision
	aX, aY := a.Position()
	aW, aH := a.Girth()
	bX, bY := b.Position()
	bW, bH := b.Girth()

	// Use move plus the girth and positions of the two things
	// to determine if there is overlap on either side of a bounding box
	aX += move.vX
	aY += move.vY

	return (aX-aW/2 < bX+bW/2 && aX+aW/2 > bX-bW/2) && (aY-aH/2 < bY+bH/2 && aY+aH/2 > bY-bH/2)
}
