package main

import (
	"encoding/base64"
	"fmt"
)

type FileRite struct{}

func init() { registerRite(FileRite{}) }

func (r FileRite) Tag() string { return "43f2b66b11e9586a274fdcee:FILE" }

// payload: [sliceB64, salt, filename, offset]
// filename and offset are UI-only — stored in payload for state restoration, not hashed
func (r FileRite) Encode(payload []interface{}) ([]byte, error) {
	if len(payload) < 2 {
		return nil, fmt.Errorf("file: expected [sliceB64, salt, filename, offset]")
	}
	sliceB64, ok := payload[0].(string)
	if !ok { return nil, fmt.Errorf("file: payload[0] must be string") }
	salt, ok := payload[1].(string)
	if !ok { return nil, fmt.Errorf("file: payload[1] must be string") }

	slice, _ := base64.StdEncoding.DecodeString(sliceB64)
	saltB := []byte(salt)
	out := make([]byte, len(slice)+len(saltB))
	copy(out, slice)
	copy(out[len(slice):], saltB)
	return out, nil
}

func (r FileRite) Entropy(rite *RiteState) float64 {
	if rite.Payload == nil { return 0 }
	salt, ok := rite.Payload[1].(string)
	if !ok { return 0 }

	var offsetBits float64
	if len(rite.Payload) >= 4 {
		if offset, ok := rite.Payload[3].(float64); ok && offset > 0 {
			s := fmt.Sprintf("%d", int(offset))
			offsetBits = StringRite{}.Entropy(&RiteState{Payload: []interface{}{s}})
		}
	}

	return 7.0 + StringRite{}.Entropy(&RiteState{Payload: []interface{}{salt}}) + offsetBits
}

func (r FileRite) Dataset() interface{} { return nil }