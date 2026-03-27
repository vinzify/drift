package scene

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

func (o *Orrery) Draw(screen tcell.Screen) {
	if o.w <= 0 || o.h <= 0 {
		return
	}

	o.clearScratch()
	o.drawStars()
	o.drawOrbits()

	for x := 0; x < o.pw; x++ {
		for y := 0; y < o.ph; y++ {
			if b := o.trail[x][y]; b > 0.05 {
				o.stampPixel(x, y, o.trailOwner[x][y], b)
			}
		}
	}

	o.drawSun()
	for _, body := range o.bodies {
		o.drawBody(body)
	}
	o.drawAsteroid()
	o.drawUFO()

	for cx := 0; cx < o.w; cx++ {
		for cy := 0; cy < o.h; cy++ {
			ch, style, ok := o.buildBrailleCell(cx, cy)
			if ok {
				screen.SetContent(cx, cy, ch, nil, style)
			}
		}
	}
}

func (o *Orrery) allocBuffers() {
	o.trail = make([][]float64, o.pw)
	o.pixels = make([][]float64, o.pw)
	o.trailOwner = make([][]uint8, o.pw)
	o.pixelOwner = make([][]uint8, o.pw)
	for x := 0; x < o.pw; x++ {
		o.trail[x] = make([]float64, o.ph)
		o.pixels[x] = make([]float64, o.ph)
		o.trailOwner[x] = make([]uint8, o.ph)
		o.pixelOwner[x] = make([]uint8, o.ph)
	}
}

func (o *Orrery) clearScratch() {
	for x := range o.pixels {
		for y := range o.pixels[x] {
			o.pixels[x][y] = 0
			o.pixelOwner[x][y] = 0
		}
	}
}

func (o *Orrery) drawStars() {
	for _, star := range o.stars {
		brightness := 0.16
		if star.static {
			brightness = 0.24
		} else {
			brightness = 0.18 + 0.08*math.Sin(o.time*0.45+star.phase)
		}
		o.stampDisc(star.x, star.y, 0.35, star.paletteIdx, brightness, false)
	}
}

func (o *Orrery) drawOrbits() {
	for _, radius := range o.orbitRadii {
		samples := int(clamp64(radius*4.2, 80, 280))
		for step := 0; step < samples; step++ {
			angle := (float64(step) / float64(samples)) * 2 * math.Pi
			x, y := o.pointOnOrbit(radius, angle)
			o.stampPixel(int(x+0.5), int(y+0.5), orreryOrbitOwner, 0.16)
		}
	}
}

func (o *Orrery) stampTrail(body *orreryBody) {
	o.stampDisc(body.x, body.y, clamp64(body.size*0.55, 0.5, 1.3), uint8(body.paletteIdx), 0.14, true)
}

func (o *Orrery) drawSun() {
	o.stampDisc(o.centerX, o.centerY, 4.6, orrerySunOwner, 1.0, false)
	o.stampDisc(o.centerX, o.centerY, 7.2, orrerySunOwner, 0.32, false)
}

func (o *Orrery) drawBody(body orreryBody) {
	o.stampDisc(body.x, body.y, body.size, uint8(body.paletteIdx), 0.95, false)
	if body.hasRing {
		o.stampRing(body.x, body.y, body.size+1.4, body.size+2.4, uint8(body.paletteIdx), 0.42)
	}
}

func (o *Orrery) drawAsteroid() {
	if !o.asteroid.active {
		return
	}
	o.stampDisc(o.asteroid.x, o.asteroid.y, o.asteroid.size, orreryAsteroidOwner, 0.88, false)
}

func (o *Orrery) drawUFO() {
	if !o.ufo.active {
		return
	}

	o.stampEllipse(o.ufo.x, o.ufo.y+0.2, 4.2, 1.4, orreryUFOOwner, 0.72, false)
	o.stampEllipseRing(o.ufo.x, o.ufo.y+0.15, 4.0, 1.2, 4.8, 1.7, orreryUFOOwner, 0.28)
	o.stampEllipse(o.ufo.x, o.ufo.y-1.3, 1.7, 0.9, orreryUFODomeOwner, 0.64, false)
	o.stampEllipse(o.ufo.x, o.ufo.y+0.9, 2.3, 0.45, orreryUFOOwner, 0.40, false)
}

func (o *Orrery) stampDisc(cx, cy, radius float64, owner uint8, brightness float64, trail bool) {
	minX := int(cx - radius - 1)
	maxX := int(cx + radius + 1)
	minY := int(cy - radius - 1)
	maxY := int(cy + radius + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			falloff := 1 - dist/math.Max(radius, 0.01)
			value := brightness * (0.45 + falloff*0.55)
			if trail {
				o.stampTrailPixel(x, y, owner, value)
			} else {
				o.stampPixel(x, y, owner, value)
			}
		}
	}
}

func (o *Orrery) stampRing(cx, cy, inner, outer float64, owner uint8, brightness float64) {
	minX := int(cx - outer - 1)
	maxX := int(cx + outer + 1)
	minY := int(cy - outer - 1)
	maxY := int(cy + outer + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := float64(x) - cx
			dy := (float64(y) - cy) * 0.72
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < inner || dist > outer {
				continue
			}
			falloff := 1 - math.Abs(dist-(inner+outer)/2)/math.Max((outer-inner)/2, 0.01)
			o.stampPixel(x, y, owner, brightness*(0.5+falloff*0.5))
		}
	}
}

func (o *Orrery) stampEllipse(cx, cy, rx, ry float64, owner uint8, brightness float64, trail bool) {
	minX := int(cx - rx - 1)
	maxX := int(cx + rx + 1)
	minY := int(cy - ry - 1)
	maxY := int(cy + ry + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := (float64(x) - cx) / math.Max(rx, 0.01)
			dy := (float64(y) - cy) / math.Max(ry, 0.01)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 1 {
				continue
			}
			falloff := 1 - dist
			value := brightness * (0.45 + falloff*0.55)
			if trail {
				o.stampTrailPixel(x, y, owner, value)
			} else {
				o.stampPixel(x, y, owner, value)
			}
		}
	}
}

func (o *Orrery) stampEllipseRing(cx, cy, innerRx, innerRy, outerRx, outerRy float64, owner uint8, brightness float64) {
	minX := int(cx - outerRx - 1)
	maxX := int(cx + outerRx + 1)
	minY := int(cy - outerRy - 1)
	maxY := int(cy + outerRy + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			outerDX := (float64(x) - cx) / math.Max(outerRx, 0.01)
			outerDY := (float64(y) - cy) / math.Max(outerRy, 0.01)
			outerDist := math.Sqrt(outerDX*outerDX + outerDY*outerDY)
			if outerDist > 1 {
				continue
			}

			innerDX := (float64(x) - cx) / math.Max(innerRx, 0.01)
			innerDY := (float64(y) - cy) / math.Max(innerRy, 0.01)
			innerDist := math.Sqrt(innerDX*innerDX + innerDY*innerDY)
			if innerDist < 1 {
				continue
			}

			falloff := 1 - outerDist
			o.stampPixel(x, y, owner, brightness*(0.5+falloff*0.5))
		}
	}
}

func (o *Orrery) stampTrailPixel(px, py int, owner uint8, brightness float64) {
	if px < 0 || px >= o.pw || py < 0 || py >= o.ph {
		return
	}
	if brightness > o.trail[px][py] {
		o.trail[px][py] = brightness
		o.trailOwner[px][py] = owner
	}
}

func (o *Orrery) stampPixel(px, py int, owner uint8, brightness float64) {
	if px < 0 || px >= o.pw || py < 0 || py >= o.ph {
		return
	}
	if brightness > o.pixels[px][py] {
		o.pixels[px][py] = brightness
		o.pixelOwner[px][py] = owner
	}
}
