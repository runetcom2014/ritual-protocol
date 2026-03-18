package main

import (
	"fmt"
	"math"
)

type ConstellationRite struct{}

func init() { registerRite(ConstellationRite{}) }

func (r ConstellationRite) Tag() string { return "828a4d063ec9712f75234c9f:CONSTELLATION" }

func (r ConstellationRite) Encode(payload []interface{}) ([]byte, error) {
	if len(payload) < 2 {
		return nil, fmt.Errorf("constellation: expected [rotation, [starIdx, ...]]")
	}
	rotF, ok := payload[0].(float64)
	if !ok { return nil, fmt.Errorf("constellation: payload[0] must be number") }
	raw, ok := payload[1].([]interface{})
	if !ok { return nil, fmt.Errorf("constellation: payload[1] must be array") }
	out := make([]byte, 1+len(raw))
	out[0] = byte(rotF)
	for i, v := range raw {
		f, ok := v.(float64)
		if !ok { return nil, fmt.Errorf("constellation: star index %d must be number", i) }
		out[i+1] = byte(f)
	}
	return out, nil
}

func (r ConstellationRite) Entropy(rite *RiteState) float64 {
	if rite.Payload == nil { return 0 }
	raw, ok := rite.Payload[1].([]interface{})
	if !ok { return 0 }
	rotBits  := math.Log2(float64(ConstellationSteps))
	starBits := float64(len(raw)) * math.Log2(float64(ConstellationStarCount))
	return rotBits + starBits
}

func (r ConstellationRite) Dataset() interface{} {
	return map[string]interface{}{
		"stars": ConstellationStars,
		"steps": ConstellationSteps,
	}
}

// StarsAt returns 25 star data rotated by step (0..17)
func StarsAt(step int) []StarData {
	angle := float64(step) * (2 * math.Pi / float64(ConstellationSteps))
	cos, sin := math.Cos(angle), math.Sin(angle)
	result := make([]StarData, len(ConstellationStars))
	for i, s := range ConstellationStars {
		dx, dy := s.X-0.5, s.Y-0.5
		result[i] = StarData{
			X:     0.5 + dx*cos - dy*sin,
			Y:     0.5 + dx*sin + dy*cos,
			Size:  s.Size,
			Color: s.Color,
			Name:  s.Name,
		}
	}
	return result
}

const ConstellationSteps     = 18
const ConstellationStarCount = 25

type StarSize  string
type StarColor string

const (
	StarSizeSmall  StarSize = "small"
	StarSizeMedium StarSize = "medium"
	StarSizeLarge  StarSize = "large"
)

const (
	StarColorBlue   StarColor = "blue"
	StarColorWhite  StarColor = "white"
	StarColorYellow StarColor = "yellow"
	StarColorOrange StarColor = "orange"
	StarColorRed    StarColor = "red"
)

type StarData struct {
	X     float64   `json:"x"`
	Y     float64   `json:"y"`
	Size  StarSize  `json:"size"`
	Color StarColor `json:"color"`
	Name  string    `json:"name"`
}

var ConstellationStars = []StarData{
	{0.50, 0.08, StarSizeLarge,  StarColorBlue,   "Sirius"},
	{0.72, 0.14, StarSizeLarge,  StarColorBlue,   "Rigel"},
	{0.83, 0.28, StarSizeLarge,  StarColorRed,    "Betelgeuse"},
	{0.88, 0.45, StarSizeMedium, StarColorOrange, "Arcturus"},
	{0.80, 0.62, StarSizeMedium, StarColorBlue,   "Vega"},
	{0.65, 0.74, StarSizeMedium, StarColorYellow, "Capella"},
	{0.50, 0.80, StarSizeLarge,  StarColorYellow, "Canopus"},
	{0.35, 0.74, StarSizeMedium, StarColorBlue,   "Spica"},
	{0.20, 0.62, StarSizeMedium, StarColorYellow, "Pollux"},
	{0.12, 0.45, StarSizeMedium, StarColorRed,    "Antares"},
	{0.17, 0.28, StarSizeMedium, StarColorBlue,   "Fomalhaut"},
	{0.28, 0.14, StarSizeMedium, StarColorWhite,  "Deneb"},
	{0.62, 0.32, StarSizeSmall,  StarColorBlue,   "Regulus"},
	{0.42, 0.22, StarSizeMedium, StarColorOrange, "Aldebaran"},
	{0.74, 0.38, StarSizeSmall,  StarColorWhite,  "Procyon"},
	{0.38, 0.38, StarSizeSmall,  StarColorWhite,  "Achernar"},
	{0.26, 0.50, StarSizeSmall,  StarColorBlue,   "Mimosa"},
	{0.58, 0.50, StarSizeSmall,  StarColorBlue,   "Acrux"},
	{0.50, 0.42, StarSizeSmall,  StarColorBlue,   "Hadar"},
	{0.44, 0.58, StarSizeSmall,  StarColorWhite,  "Alnilam"},
	{0.66, 0.58, StarSizeSmall,  StarColorWhite,  "Mintaka"},
	{0.32, 0.28, StarSizeSmall,  StarColorWhite,  "Alnitak"},
	{0.70, 0.22, StarSizeSmall,  StarColorBlue,   "Mirzam"},
	{0.22, 0.38, StarSizeSmall,  StarColorOrange, "Avior"},
	{0.78, 0.55, StarSizeMedium, StarColorYellow, "Rigil Kent"},
}