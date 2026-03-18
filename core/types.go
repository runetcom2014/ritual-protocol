package main

import "strings"

const ProtocolVersion = "v1"

// Rite is the interface every rite type must implement
type Rite interface {
	Tag() string // "PREFIX:NAME" — used as type_tag in hashing
	Encode(payload []interface{}) ([]byte, error)
	Entropy(rite *RiteState) float64
	Dataset() interface{}
}

// riteRegistry maps rite name (e.g. "STRING") to its implementation
// populated automatically via registerRite() calls in each rite's init()
var riteRegistry = map[string]Rite{}

func registerRite(r Rite) {
	name := strings.SplitN(r.Tag(), ":", 2)[1]
	riteRegistry[name] = r
}

type EntropyResult struct {
	RiteBits  float64 `json:"RiteBits"`
	TotalBits float64 `json:"TotalBits"`
}

type RiteState struct {
	ID       int
	riteName string
	Payload  []interface{}
}