package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"
)

var (
	store   = map[C.int]*Ritual{}
	storeMu sync.Mutex
	counter C.int
)

func main() {}

func storeGet(h C.int) *Ritual {
	storeMu.Lock()
	defer storeMu.Unlock()
	return store[h]
}

func storeAdd(r *Ritual) C.int {
	storeMu.Lock()
	defer storeMu.Unlock()
	counter++
	store[counter] = r
	return counter
}

func storeDel(h C.int) {
	storeMu.Lock()
	defer storeMu.Unlock()
	delete(store, h)
}

func cok(v interface{}) *C.char {
	b, _ := json.Marshal(v)
	return C.CString(string(b))
}

func cerr(msg string) *C.char {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return C.CString(string(b))
}

//export RitualNew
func RitualNew() C.int {
	return storeAdd(New())
}

//export RitualFree
func RitualFree(h C.int) {
	storeDel(h)
}

//export RitualFreeString
func RitualFreeString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

//export RitualAddRite
func RitualAddRite(h C.int, name *C.char) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	id, err := r.AddRite(C.GoString(name))
	if err != nil { return cerr(err.Error()) }
	return cok(map[string]int{"id": id})
}

//export RitualUpdateRite
func RitualUpdateRite(h C.int, riteID C.int, payloadJSON *C.char) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	var payload []interface{}
	if err := json.Unmarshal([]byte(C.GoString(payloadJSON)), &payload); err != nil {
		return cerr("invalid payload: " + err.Error())
	}
	result, err := r.UpdateRite(int(riteID), payload)
	if err != nil { return cerr(err.Error()) }
	return cok(result)
}

//export RitualRemoveRite
func RitualRemoveRite(h C.int, riteID C.int) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	if err := r.RemoveRite(int(riteID)); err != nil { return cerr(err.Error()) }
	return cok(map[string]bool{"ok": true})
}

//export RitualGetRitePayload
func RitualGetRitePayload(h C.int, riteID C.int) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	payload, err := r.GetRitePayload(int(riteID))
	if err != nil { return cerr(err.Error()) }
	return cok(payload)
}

//export RitualFinalize
func RitualFinalize(h C.int) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	key, err := r.Finalize()
	if err != nil { return cerr(err.Error()) }
	return cok(map[string]interface{}{
		"key":       fmt.Sprintf("%x", key),
		"totalBits": r.GetEntropy().Total,
	})
}

//export RitualGetState
func RitualGetState(h C.int) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	return cok(r.GetState())
}

//export RitualGetEntropy
func RitualGetEntropy(h C.int) *C.char {
	r := storeGet(h)
	if r == nil { return cerr("invalid handle") }
	return cok(r.GetEntropy())
}

//export RitualGetRiteDataset
func RitualGetRiteDataset(name *C.char) *C.char {
	return cok(GetRiteDataset(C.GoString(name)))
}