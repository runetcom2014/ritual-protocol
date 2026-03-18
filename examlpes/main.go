package main

/*
#cgo LDFLAGS: ritual.dll
#include <stdlib.h>
extern int   RitualNew();
extern void  RitualFree(int h);
extern void  RitualFreeString(char* s);
extern char* RitualAddRite(int h, char* name);
extern char* RitualUpdateRite(int h, int id, char* payloadJSON);
extern char* RitualRemoveRite(int h, int id);
extern char* RitualGetRitePayload(int h, int id);
extern char* RitualFinalize(int h);
extern char* RitualGetState(int h);
extern char* RitualGetEntropy(int h);
extern char* RitualGetRiteDataset(char* name);
*/
import "C"
import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	webview "github.com/webview/webview_go"
)

var handle C.int

func gostr(cs *C.char) string {
	s := C.GoString(cs)
	C.RitualFreeString(cs)
	return s
}

func main() {
	handle = C.RitualNew()
	defer C.RitualFree(handle)

	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("Ritual Protocol v1")
	w.SetSize(1200, 800, webview.HintNone)

	w.Bind("ritualReset", func() {
		C.RitualFree(handle)
		handle = C.RitualNew()
	})

	w.Bind("ritualAddRite", func(name string) string {
		cs := C.CString(name)
		defer C.free(unsafe.Pointer(cs))
		return gostr(C.RitualAddRite(handle, cs))
	})

	w.Bind("ritualUpdateRite", func(riteID int, payload []interface{}) string {
		b, _ := json.Marshal(payload)
		cs := C.CString(string(b))
		defer C.free(unsafe.Pointer(cs))
		return gostr(C.RitualUpdateRite(handle, C.int(riteID), cs))
	})

	w.Bind("ritualGetRitePayload", func(riteID int) string {
		return gostr(C.RitualGetRitePayload(handle, C.int(riteID)))
	})

	w.Bind("ritualRemoveRite", func(riteID int) string {
		return gostr(C.RitualRemoveRite(handle, C.int(riteID)))
	})

	w.Bind("ritualFinalize", func() string {
		s := gostr(C.RitualFinalize(handle))
		var parsed map[string]interface{}
		json.Unmarshal([]byte(s), &parsed)
		if keyHex, ok := parsed["key"].(string); ok {
			keyBytes, _ := hex.DecodeString(keyHex)
			var master [32]byte
			copy(master[:], keyBytes)
			bundle, err := DeriveKeys(master)
			if err == nil {
				parsed["keys"] = bundle
			}
		}
		b, _ := json.Marshal(parsed)
		return string(b)
	})

	w.Bind("ritualDeriveKeys", func(keyHex string) string {
		keyBytes, err := hex.DecodeString(keyHex)
		if err != nil { return fmt.Sprintf(`{"error":"%s"}`, err.Error()) }
		var master [32]byte
		copy(master[:], keyBytes)
		bundle, err := DeriveKeys(master)
		if err != nil { return fmt.Sprintf(`{"error":"%s"}`, err.Error()) }
		b, _ := json.Marshal(bundle)
		return string(b)
	})

	w.Bind("ritualGetRiteData", func(name string) string {
		cs := C.CString(name)
		defer C.free(unsafe.Pointer(cs))
		return gostr(C.RitualGetRiteDataset(cs))
	})

	w.Bind("ritualGetState", func() string {
		return gostr(C.RitualGetState(handle))
	})

	w.Bind("ritualGetEntropy", func() string {
		return gostr(C.RitualGetEntropy(handle))
	})

	exePath, err := os.Executable()
	if err != nil { fmt.Println("Error:", err); return }
	htmlPath := filepath.Join(filepath.Dir(exePath), "ritual-ui.html")
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil { fmt.Println("Error reading ritual-ui.html:", err); return }
	w.SetHtml(string(htmlBytes))

	fmt.Println("Ritual Protocol running...")
	w.Run()
}