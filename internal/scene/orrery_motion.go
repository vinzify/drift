package scene

import (
	"math"
	"math/rand"
)

func (o *Orrery) buildStars() {
	if o.pw <= 0 || o.ph <= 0 {
		o.stars = nil
		return
	}

	seed := int64(o.w*44549 ^ o.h*98317 ^ 0x45a11)
	rng := rand.New(rand.NewSource(seed))
	count := maxInt((o.w*o.h)/90, 36)
	stars := make([]orreryStar, 0, count)

	addStar := func(xMin, xMax float64, idx int) {
		for attempt := 0; attempt < 8; attempt++ {
			x := xMin + rng.Float64()*(xMax-xMin)
			y := rng.Float64() * float64(o.ph)
			if math.Abs(x-o.centerX) < 20 && math.Abs(y-o.centerY) < 14 {
				continue
			}
			stars = append(stars, orreryStar{
				x:          x,
				y:          y,
				paletteIdx: uint8(idx % len(o.theme.Palette)),
				phase:      rng.Float64() * 2 * math.Pi,
				static:     rng.Intn(3) == 0,
			})
			return
		}
	}

	leftCount := count / 2
	rightCount := count - leftCount
	for i := 0; i < leftCount; i++ {
		addStar(0, o.centerX-12, i)
	}
	for i := 0; i < rightCount; i++ {
		addStar(o.centerX+12, float64(o.pw), leftCount+i)
	}
	o.stars = stars
}

func (o *Orrery) buildBodies() {
	count := o.effectiveBodyCount()
	seed := int64(o.w*73856093 ^ o.h*19349663 ^ count*83492791)
	rng := rand.New(rand.NewSource(seed))

	maxRadius := math.Min(float64(o.pw)*0.26, float64(o.ph)*0.26)
	minRadius := 10.0
	step := (maxRadius - minRadius) / float64(maxInt(count, 1))
	if step < 6 {
		step = 6
	}

	sizePattern := []float64{1.0, 1.3, 1.7, 1.5, 2.0, 2.4, 2.1, 1.8}
	o.bodies = make([]orreryBody, count)
	o.orbitRadii = make([]float64, count)
	for i := 0; i < count; i++ {
		frac := float64(i+1) / float64(count+1)
		radius := minRadius + step*float64(i+1)
		speed := 1.45 / math.Pow(radius+4, 0.72)

		o.bodies[i] = orreryBody{
			radius:     radius,
			angle:      rng.Float64() * 2 * math.Pi,
			speed:      speed,
			size:       sizePattern[minInt(i, len(sizePattern)-1)],
			paletteIdx: i % len(o.theme.Palette),
			hasRing:    i == minInt(5, count-1),
		}
		o.orbitRadii[i] = radius

		if i < 3 {
			o.bodies[i].radius += frac * 4
			o.orbitRadii[i] = o.bodies[i].radius
		}
	}
}

func (o *Orrery) resetAsteroid() {
	o.asteroid = orreryAsteroid{
		active:   false,
		cooldown: 4 + o.rng.Float64()*5,
		size:     1.0,
	}
}

func (o *Orrery) resetUFO() {
	o.ufo = orreryUFO{
		active:   false,
		cooldown: 9 + o.rng.Float64()*10,
	}
}

func (o *Orrery) spawnAsteroid() {
	margin := 18.0
	side := o.rng.Intn(4)
	maxFlybyRadius := 24.0
	if len(o.orbitRadii) > 0 {
		maxFlybyRadius = o.orbitRadii[len(o.orbitRadii)-1] + 14
	}
	flybyRadius := 16.0 + o.rng.Float64()*(maxFlybyRadius-16.0)
	flybyAngle := o.rng.Float64() * 2 * math.Pi
	targetX := o.centerX + math.Cos(flybyAngle)*flybyRadius
	targetY := o.centerY + math.Sin(flybyAngle)*flybyRadius

	var x, y float64
	switch side {
	case 0:
		x = -margin
		y = o.rng.Float64() * float64(o.ph)
	case 1:
		x = float64(o.pw) + margin
		y = o.rng.Float64() * float64(o.ph)
	case 2:
		x = o.rng.Float64() * float64(o.pw)
		y = -margin
	default:
		x = o.rng.Float64() * float64(o.pw)
		y = float64(o.ph) + margin
	}

	dx := targetX - x
	dy := targetY - y
	dist := math.Sqrt(dx*dx + dy*dy)
	speed := 14 + o.rng.Float64()*7

	o.asteroid = orreryAsteroid{
		active: true,
		x:      x,
		y:      y,
		vx:     dx / math.Max(dist, 0.01) * speed,
		vy:     dy / math.Max(dist, 0.01) * speed,
		size:   0.7 + o.rng.Float64()*1.5,
	}
}

func (o *Orrery) updateAsteroid(dt float64) {
	if !o.asteroid.active {
		o.asteroid.cooldown -= dt
		if o.asteroid.cooldown <= 0 {
			o.spawnAsteroid()
		}
		return
	}

	dx := o.centerX - o.asteroid.x
	dy := o.centerY - o.asteroid.y
	dist2 := dx*dx + dy*dy
	dist := math.Sqrt(math.Max(dist2, 1))
	accel := 1500 / math.Max(dist2, 180)

	o.asteroid.vx += dx / dist * accel * dt
	o.asteroid.vy += dy / dist * accel * dt

	sunAvoidRadius := 12.0
	if dist < sunAvoidRadius {
		repel := 5200 / math.Max(dist2, 50)
		o.asteroid.vx -= dx / dist * repel * dt
		o.asteroid.vy -= dy / dist * repel * dt
	}

	maxSpeed := 24.0
	speed := math.Sqrt(o.asteroid.vx*o.asteroid.vx + o.asteroid.vy*o.asteroid.vy)
	if speed > maxSpeed {
		o.asteroid.vx = o.asteroid.vx / speed * maxSpeed
		o.asteroid.vy = o.asteroid.vy / speed * maxSpeed
	}

	o.asteroid.x += o.asteroid.vx * dt
	o.asteroid.y += o.asteroid.vy * dt

	o.stampDisc(o.asteroid.x, o.asteroid.y, clamp64(o.asteroid.size*0.6, 0.5, 1.5), orreryAsteroidOwner, 0.22, true)

	margin := 26.0
	if o.asteroid.x < -margin || o.asteroid.x > float64(o.pw)+margin || o.asteroid.y < -margin || o.asteroid.y > float64(o.ph)+margin {
		o.resetAsteroid()
	}
}

func (o *Orrery) spawnUFO() {
	margin := 24.0
	side := o.rng.Intn(2)
	radius := o.orbitRadii[maxInt(len(o.orbitRadii)-2, 0)] + 8 + o.rng.Float64()*10
	angle := -0.9 + o.rng.Float64()*1.8
	targetX, targetY := o.pointOnOrbit(radius, angle)

	x := -margin
	if side == 1 {
		x = float64(o.pw) + margin
	}
	y := targetY - 8 - o.rng.Float64()*10
	if y < 10 {
		y = 10
	}
	if y > float64(o.ph)-10 {
		y = float64(o.ph) - 10
	}

	o.ufo = orreryUFO{
		active:     true,
		x:          x,
		y:          y,
		targetX:    targetX,
		targetY:    targetY,
		hoverTime:  1.4 + o.rng.Float64()*1.4,
		wobbleSeed: o.rng.Float64() * 2 * math.Pi,
	}
}

func (o *Orrery) updateUFO(dt float64) {
	if !o.ufo.active {
		o.ufo.cooldown -= dt
		if o.ufo.cooldown <= 0 {
			o.spawnUFO()
		}
		return
	}

	if !o.ufo.departing {
		dx := o.ufo.targetX - o.ufo.x
		dy := o.ufo.targetY - o.ufo.y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 2.5 {
			speed := 16.0
			o.ufo.vx = dx / math.Max(dist, 0.01) * speed
			o.ufo.vy = dy / math.Max(dist, 0.01) * speed
			o.ufo.x += o.ufo.vx * dt
			o.ufo.y += o.ufo.vy * dt
		} else {
			o.ufo.hoverTime -= dt
			o.ufo.x = o.ufo.targetX + math.Sin(o.time*2.6+o.ufo.wobbleSeed)*1.8
			o.ufo.y = o.ufo.targetY + math.Sin(o.time*4.1+o.ufo.wobbleSeed)*0.8
			if o.ufo.hoverTime <= 0 {
				o.ufo.departing = true
				o.ufo.vx = 44 + o.rng.Float64()*18
				if o.ufo.x > o.centerX {
					o.ufo.vx *= -1
				}
				o.ufo.vy = -6 + o.rng.Float64()*12
			}
		}
	} else {
		o.ufo.x += o.ufo.vx * dt
		o.ufo.y += o.ufo.vy * dt
		o.ufo.vx *= math.Pow(1.02, dt*60)
	}

	margin := 30.0
	if o.ufo.x < -margin || o.ufo.x > float64(o.pw)+margin || o.ufo.y < -margin || o.ufo.y > float64(o.ph)+margin {
		o.resetUFO()
	}
}
