package scene

import "github.com/gdamore/tcell/v2"

func (o *Orrery) buildBrailleCell(cx, cy int) (rune, tcell.Style, bool) {
	var mask uint8
	paletteEnergy := make([]float64, len(o.theme.Palette))
	orbitEnergy := 0.0
	sunEnergy := 0.0
	starEnergy := 0.0
	asteroidEnergy := 0.0
	ufoEnergy := 0.0
	ufoDomeEnergy := 0.0
	total := 0.0

	for subRow := 0; subRow < 4; subRow++ {
		for subCol := 0; subCol < 2; subCol++ {
			px := cx*2 + subCol
			py := cy*4 + subRow
			if px >= o.pw || py >= o.ph {
				continue
			}

			brightness := o.pixels[px][py]
			if brightness < 0.06 {
				continue
			}

			mask |= uint8(1) << brailleOffsets[subRow][subCol]
			total += brightness

			switch owner := o.pixelOwner[px][py]; {
			case owner == orrerySunOwner:
				sunEnergy += brightness
			case owner == orreryOrbitOwner:
				orbitEnergy += brightness
			case owner == orreryStarOwner:
				starEnergy += brightness
			case owner == orreryAsteroidOwner:
				asteroidEnergy += brightness
			case owner == orreryUFODomeOwner:
				ufoDomeEnergy += brightness
			case owner == orreryUFOOwner:
				ufoEnergy += brightness
			default:
				paletteEnergy[int(owner)%len(o.theme.Palette)] += brightness
			}
		}
	}

	if mask == 0 {
		return 0, tcell.StyleDefault, false
	}

	var color RGBColor
	switch {
	case sunEnergy >= orbitEnergy && sunEnergy >= starEnergy && sunEnergy >= maxPaletteEnergy(paletteEnergy):
		color = Lerp(o.theme.Palette[0], o.theme.Bright, clamp64(0.72+sunEnergy*0.18, 0, 1))
	case ufoDomeEnergy >= ufoEnergy && ufoDomeEnergy >= asteroidEnergy && ufoDomeEnergy >= orbitEnergy && ufoDomeEnergy >= starEnergy && ufoDomeEnergy >= maxPaletteEnergy(paletteEnergy):
		color = Lerp(o.theme.Palette[1%len(o.theme.Palette)], o.theme.Bright, clamp64(0.52+ufoDomeEnergy*0.24, 0, 1))
	case ufoEnergy >= asteroidEnergy && ufoEnergy >= orbitEnergy && ufoEnergy >= starEnergy && ufoEnergy >= maxPaletteEnergy(paletteEnergy):
		color = Lerp(o.theme.Palette[3%len(o.theme.Palette)], o.theme.Bright, clamp64(0.56+ufoEnergy*0.22, 0, 1))
	case asteroidEnergy >= orbitEnergy && asteroidEnergy >= starEnergy && asteroidEnergy >= maxPaletteEnergy(paletteEnergy):
		color = Lerp(o.theme.Palette[2%len(o.theme.Palette)], o.theme.Bright, clamp64(0.42+asteroidEnergy*0.24, 0, 1))
	case orbitEnergy >= starEnergy && orbitEnergy >= maxPaletteEnergy(paletteEnergy):
		color = Lerp(o.theme.Dim[0], o.theme.Bright, clamp64(orbitEnergy*0.35, 0.18, 0.4))
	case starEnergy >= maxPaletteEnergy(paletteEnergy):
		color = Lerp(o.theme.Dim[1%len(o.theme.Dim)], o.theme.Palette[1%len(o.theme.Palette)], clamp64(starEnergy*0.45, 0.12, 0.28))
	default:
		best := 0
		for i := 1; i < len(paletteEnergy); i++ {
			if paletteEnergy[i] > paletteEnergy[best] {
				best = i
			}
		}
		color = Lerp(o.theme.Palette[best], o.theme.Bright, clamp64(total*0.16, 0, 0.32))
	}

	return '\u2800' | rune(mask), color.Style(), true
}

func maxPaletteEnergy(vals []float64) float64 {
	best := 0.0
	for _, v := range vals {
		if v > best {
			best = v
		}
	}
	return best
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
