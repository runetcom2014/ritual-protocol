package main

import (
	"fmt"
	"math"
	"unicode"
)

type StringRite struct{}

func init() { registerRite(StringRite{}) }

func (r StringRite) Tag() string { return "81cc991c40419971d2f7b754:STRING" }

func (r StringRite) Encode(payload []interface{}) ([]byte, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("string: expected [string]")
	}
	s, ok := payload[0].(string)
	if !ok {
		return nil, fmt.Errorf("string: payload[0] must be string")
	}
	return []byte(s), nil
}

func (r StringRite) Entropy(rite *RiteState) float64 {
	if rite.Payload == nil { return 0 }
	s, ok := rite.Payload[0].(string)
	if !ok || len(s) == 0 { return 0 }
	runes := []rune(s)
	unique := make(map[rune]struct{})
	for _, r := range runes { unique[r] = struct{}{} }
	alpha := 0
	hasLower, hasUpper, hasDigit, hasSpecial := false, false, false, false
	for _, r := range runes {
		if unicode.IsLower(r)        { hasLower   = true
		} else if unicode.IsUpper(r) { hasUpper   = true
		} else if unicode.IsDigit(r) { hasDigit   = true
		} else                       { hasSpecial = true }
	}
	if hasLower   { alpha += 26 }
	if hasUpper   { alpha += 26 }
	if hasDigit   { alpha += 10 }
	if hasSpecial { alpha += 33 }
	if alpha == 0 { alpha = 26 }
	if len(unique) == 1 {
		return math.Log2(float64(alpha) * float64(len(runes)))
	}
	return float64(len(runes)) * math.Log2(float64(alpha))
}

func (r StringRite) Dataset() interface{} { return nil }