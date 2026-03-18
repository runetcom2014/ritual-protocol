package main

import (
	"fmt"
	"math"
)

type RuneGridRite struct{}

func init() { registerRite(RuneGridRite{}) }

func (r RuneGridRite) Tag() string { return "2e53405c69848a3e657412a3:RUNEGRID" }

func (r RuneGridRite) Encode(payload []interface{}) ([]byte, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("runegrid: expected [[cell, runeIdx], ...]")
	}
	out := make([]byte, len(payload)*2)
	for i, v := range payload {
		pair, ok := v.([]interface{})
		if !ok || len(pair) < 2 {
			return nil, fmt.Errorf("runegrid: placement %d must be [cell, runeIdx]", i)
		}
		cell, ok1 := pair[0].(float64)
		rune_, ok2 := pair[1].(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("runegrid: placement %d values must be numbers", i)
		}
		out[i*2]   = byte(cell)
		out[i*2+1] = byte(rune_)
	}
	return out, nil
}

func (r RuneGridRite) Entropy(rite *RiteState) float64 {
	if rite.Payload == nil { return 0 }
	n := len(rite.Payload)
	if n == 0 { return 0 }

	unique := make(map[float64]struct{})
	for _, v := range rite.Payload {
		pair, ok := v.([]interface{})
		if !ok || len(pair) < 2 { continue }
		if ri, ok := pair[1].(float64); ok {
			unique[ri] = struct{}{}
		}
	}

	posBits := 0.0
	for i := 0; i < n; i++ {
		posBits += math.Log2(float64(RuneGridSize - i))
	}

	var runeBits float64
	if len(unique) == 1 {
		runeBits = math.Log2(float64(RuneAlphabetSize))
	} else {
		runeBits = float64(n) * math.Log2(float64(RuneAlphabetSize))
	}

	return posBits + runeBits
}

func (r RuneGridRite) Dataset() interface{} {
	return map[string]interface{}{
		"runes":     RuneAlphabet,
		"runeNames": RuneNames,
		"gridSize":  RuneGridSize,
		"gridCount": RuneGridCount,
	}
}

const RuneGridSize     = 9
const RuneGridCount    = 4
const RuneAlphabetSize = 24

var RuneAlphabet = []string{
	"ᚠ","ᚢ","ᚦ","ᚨ","ᚱ","ᚲ","ᚷ","ᚹ",
	"ᚺ","ᚾ","ᛁ","ᛃ","ᛇ","ᛈ","ᛉ","ᛊ",
	"ᛏ","ᛒ","ᛖ","ᛗ","ᛚ","ᛜ","ᛞ","ᛟ",
}

var RuneNames = []string{
	"fehu","uruz","thurisaz","ansuz","raidho","kenaz","gebo","wunjo",
	"hagalaz","nauthiz","isa","jera","eihwaz","perthro","algiz","sowilo",
	"tiwaz","berkano","ehwaz","mannaz","laguz","ingwaz","dagaz","othalan",
}