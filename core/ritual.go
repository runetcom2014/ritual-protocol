package main

import (
	"crypto/sha256"
	"errors"
	"math"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/scrypt"
)

// riteTypeEntropy is the entropy contribution of choosing a rite type
// log2(number of registered rite types)
func riteTypeEntropy() float64 {
	return math.Log2(float64(len(riteRegistry)))
}

// Ritual is the main object — holds the ordered list of rites
type Ritual struct {
	rites   []*RiteState
	counter int
}

// New creates a new empty Ritual
func New() *Ritual {
	return &Ritual{}
}

// GetRiteDataset returns the static dataset for a rite type by name (e.g. "STRING")
// Returns nil for rites without a dataset
func GetRiteDataset(name string) interface{} {
	if impl, ok := riteRegistry[name]; ok {
		return impl.Dataset()
	}
	return nil
}

// AddRite adds a new rite of the given type name to the ritual
// Returns the rite ID to use in subsequent calls
func (r *Ritual) AddRite(name string) (int, error) {
	if _, ok := riteRegistry[name]; !ok {
		return 0, errors.New("unknown rite type: " + name)
	}
	r.counter++
	rite := &RiteState{
		ID:       r.counter,
		riteName: name,
	}
	r.rites = append(r.rites, rite)
	return rite.ID, nil
}

// UpdateRite updates the payload of a rite and returns current entropy
// Call this on every user input event
func (r *Ritual) UpdateRite(id int, payload []interface{}) (EntropyResult, error) {
	rite := r.findRite(id)
	if rite == nil {
		return EntropyResult{}, errors.New("rite not found")
	}

	impl, ok := riteRegistry[rite.riteName]
	if !ok {
		return EntropyResult{}, errors.New("unknown rite type: " + rite.riteName)
	}

	rite.Payload = payload

	riteBits := impl.Entropy(rite)
	totalBits := r.totalEntropy()

	return EntropyResult{
		RiteBits:  riteBits,
		TotalBits: totalBits,
	}, nil
}

// GetRitePayload returns the current payload of a rite
func (r *Ritual) GetRitePayload(id int) ([]interface{}, error) {
	rite := r.findRite(id)
	if rite == nil {
		return nil, errors.New("rite not found")
	}
	return rite.Payload, nil
}

// RemoveRite removes a rite from the ritual
func (r *Ritual) RemoveRite(id int) error {
	for i, rite := range r.rites {
		if rite.ID == id {
			r.rites = append(r.rites[:i], r.rites[i+1:]...)
			return nil
		}
	}
	return errors.New("rite not found")
}

// Finalize encodes all rite payloads, folds them into the final 256-bit key
func (r *Ritual) Finalize() ([32]byte, error) {
	if len(r.rites) == 0 {
		return [32]byte{}, errors.New("ritual has no rites")
	}
	for _, rite := range r.rites {
		if rite.Payload == nil {
			return [32]byte{}, errors.New("rite has no payload")
		}
	}
	return foldChain(r.rites)
}

// RiteInfo describes a single rite — returned by GetState
type RiteInfo struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	HasData bool   `json:"hasData"`
}

// RitualState is returned by GetState
type RitualState struct {
	Rites []RiteInfo `json:"rites"`
}

// RiteEntropy is per-rite entropy — returned by GetEntropy
type RiteEntropy struct {
	ID   int     `json:"id"`
	Bits float64 `json:"bits"`
}

// EntropyState is returned by GetEntropy
type EntropyState struct {
	Rites []RiteEntropy `json:"rites"`
	Total float64       `json:"total"`
}

// GetState returns the current ritual structure
func (r *Ritual) GetState() RitualState {
	rites := make([]RiteInfo, len(r.rites))
	for i, rite := range r.rites {
		rites[i] = RiteInfo{
			ID:      rite.ID,
			Type:    rite.riteName,
			HasData: rite.Payload != nil,
		}
	}
	return RitualState{Rites: rites}
}

// GetEntropy returns per-rite and total entropy
func (r *Ritual) GetEntropy() EntropyState {
	rites := make([]RiteEntropy, len(r.rites))
	for i, rite := range r.rites {
		bits := 0.0
		if impl, ok := riteRegistry[rite.riteName]; ok {
			bits = impl.Entropy(rite)
		}
		rites[i] = RiteEntropy{
			ID:   rite.ID,
			Bits: bits,
		}
	}
	return EntropyState{
		Rites: rites,
		Total: r.totalEntropy(),
	}
}

// --- internal ---

// performRite computes SHA256(tag || encoded_payload) → [32]byte
func performRite(rite *RiteState) ([32]byte, error) {
	impl, ok := riteRegistry[rite.riteName]
	if !ok {
		return [32]byte{}, errors.New("unknown rite type: " + rite.riteName)
	}
	encoded, err := impl.Encode(rite.Payload)
	if err != nil {
		return [32]byte{}, err
	}
	h := sha256.New()
	h.Write([]byte(impl.Tag()))
	h.Write(encoded)
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out, nil
}

func (r *Ritual) findRite(id int) *RiteState {
	for _, rite := range r.rites {
		if rite.ID == id {
			return rite
		}
	}
	return nil
}

func (r *Ritual) totalEntropy() float64 {
	typeE := riteTypeEntropy()
	total := 0.0
	for _, rite := range r.rites {
		if impl, ok := riteRegistry[rite.riteName]; ok {
			total += impl.Entropy(rite) + typeE
		}
	}
	return total
}

// foldChain: encodes each rite on the fly, folds into final key
// then hardens through Argon2id → scrypt → BLAKE2b — V1 PROFILE constants
func foldChain(rites []*RiteState) ([32]byte, error) {
	// step 1: SHA256 fold
	first, err := performRite(rites[0])
	if err != nil {
		return [32]byte{}, err
	}
	state := first
	for _, rite := range rites[1:] {
		out, err := performRite(rite)
		if err != nil {
			return [32]byte{}, err
		}
		h := sha256.New()
		h.Write(state[:])
		h.Write(out[:])
		copy(state[:], h.Sum(nil))
	}

	// V1 PROFILE constants
	const profileSalt = "87d32c69ac183b7832e01cf5:RITUAL-V1"
	salt := []byte(profileSalt)

	// step 2: Argon2id — memory-hard, GPU/ASIC resistant
	stage1 := argon2.IDKey(
		state[:],
		salt,
		3,       // time
		64*1024, // memory: 64MB in KB
		4,       // threads
		32,
	)

	// step 3: scrypt — different memory access pattern
	stage2, err := scrypt.Key(
		stage1,
		salt,
		1<<17, // N=131072 — 128MB
		8,     // r
		1,     // p
		32,
	)
	if err != nil {
		return [32]byte{}, err
	}

	// step 4: BLAKE2b-256 — final hash, different architecture
	h, err := blake2b.New256(nil)
	if err != nil {
		return [32]byte{}, err
	}
	h.Write(stage2)
	h.Write(salt)

	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key, nil
}