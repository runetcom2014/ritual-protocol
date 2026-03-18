package main

import (
	"fmt"
	"math"
)

type CityTimeRite struct{}

func init() { registerRite(CityTimeRite{}) }

func (r CityTimeRite) Tag() string { return "f09801cecce8373e22030aab:CITYTIME" }

func (r CityTimeRite) Encode(payload []interface{}) ([]byte, error) {
	if len(payload) < 2 {
		return nil, fmt.Errorf("citytime: expected [city, hhmm]")
	}
	city, ok := payload[0].(string)
	if !ok { return nil, fmt.Errorf("citytime: payload[0] must be string") }
	hhmmF, ok := payload[1].(float64)
	if !ok { return nil, fmt.Errorf("citytime: payload[1] must be number") }
	hhmm := uint16(hhmmF)
	c := []byte(city)
	out := make([]byte, len(c)+3)
	copy(out, c)
	out[len(c)]   = 0x00
	out[len(c)+1] = byte(hhmm >> 8)
	out[len(c)+2] = byte(hhmm & 0xff)
	return out, nil
}

func (r CityTimeRite) Entropy(rite *RiteState) float64 {
	if rite.Payload == nil { return 0 }
	city, ok := rite.Payload[0].(string)
	if !ok || len(city) == 0 { return 0 }
	return math.Log2(float64(len(CityList))) + math.Log2(float64(24*60))
}

func (r CityTimeRite) Dataset() interface{} {
	return map[string]interface{}{
		"cities": CityList,
	}
}