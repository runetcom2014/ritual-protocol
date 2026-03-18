package main

import (
	"fmt"
	"math"
)

type SequenceRite struct{}

func init() { registerRite(SequenceRite{}) }

func (r SequenceRite) Tag() string { return "0968cec3b90372c01b651b05:SEQUENCE" }

func (r SequenceRite) Encode(payload []interface{}) ([]byte, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("sequence: expected [[idx, ...]]")
	}
	raw, ok := payload[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sequence: payload[0] must be array")
	}
	seq := make([]byte, len(raw))
	for i, v := range raw {
		f, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("sequence: index %d must be number", i)
		}
		seq[i] = byte(f)
	}
	return seq, nil
}

func (r SequenceRite) Entropy(rite *RiteState) float64 {
	if rite.Payload == nil { return 0 }
	raw, ok := rite.Payload[0].([]interface{})
	if !ok || len(raw) == 0 { return 0 }
	return float64(len(raw)) * math.Log2(float64(len(SequenceAlphabet)))
}

func (r SequenceRite) Dataset() interface{} {
	return map[string]interface{}{
		"symbols": SequenceAlphabet,
		"emoji":   SequenceEmoji,
	}
}

var SequenceAlphabet = []string{
	"fire", "water", "earth", "lightning", "human",
	"eye", "hand", "heart", "cross", "time",
	"sun", "star", "key", "shield", "scales",
	"plus", "cycle", "equal", "home", "danger",
	"up", "down", "left", "right", "question",
}

var SequenceEmoji = map[string]string{
	"fire": "🔥", "water": "💧", "earth": "🌍", "lightning": "⚡", "human": "👤",
	"eye": "👁", "hand": "✋", "heart": "❤️", "cross": "✖️", "time": "⏳",
	"sun": "☀️", "star": "⭐", "key": "🔑", "shield": "🛡", "scales": "⚖️",
	"plus": "+", "cycle": "🔄", "equal": "=", "home": "🏠", "danger": "⚠️",
	"up": "⬆️", "down": "⬇️", "left": "⬅️", "right": "➡️", "question": "❓",
}