# Ritual Protocol

A cryptographic protocol for deterministic 256-bit key generation through
a sequence of interactive actions called rites.

## Core idea

The key is reproduced through action, not retrieved from storage.
The same sequence of rites with the same parameters always produces
the same key — eliminating an entire class of threats related to
storage compromise.

## How it works

A ritual is an ordered list of rites. Each rite has a type and parameters.
The key is derived by folding the rite hashes through a KDF chain:

```
rite_hash = SHA256(type_tag || encode(payload))
state     = fold(rite_hashes)
key       = Argon2id → scrypt → BLAKE2b-256
```

## Rite types

| Type | Description |
|------|-------------|
| STRING | Text passphrase |
| SEQUENCE | Ordered sequence of symbols |
| FILE | A 512-byte slice from a file |
| CONSTELLATION | Star map rotation and selected stars |
| CITYTIME | City and time of day |
| RUNEGRID | Runes placed on a 3×3 grid |

## V1 Profile

The V1 Profile is the specific set of constants (type tags, domain separator,
KDF parameters, rite datasets) that defines interoperability. Any implementation
using these constants will produce identical keys for identical ritual inputs.

Commercial use of this software or any V1-compatible implementation
requires a license. See [LICENSE](LICENSE).

## Specification

Full protocol specification: [10.5281/zenodo.19090390]

## Build

Requirements: Go 1.21+, GCC (MinGW on Windows)

### Core library

```bash
# Windows
go build -buildmode=c-shared -o ritual.dll .

# Linux / macOS
go build -buildmode=c-shared -o ritual.so .
```

Produces `ritual.dll` / `ritual.so` and `ritual.h`.

### Example application

```bash
# Copy ritual.dll (or ritual.so) into examples/
cd examples
go build -o R.exe .   # Windows
go build -o R .       # Linux / macOS
```

## C API

```c
int   RitualNew()
void  RitualFree(int handle)
void  RitualFreeString(char* s)

char* RitualAddRite(int handle, char* name)
char* RitualUpdateRite(int handle, int riteID, char* payloadJSON)
char* RitualRemoveRite(int handle, int riteID)
char* RitualGetRitePayload(int handle, int riteID)
char* RitualFinalize(int handle)
char* RitualGetState(int handle)
char* RitualGetEntropy(int handle)
char* RitualGetRiteDataset(char* name)
```

All functions returning `char*` must be freed with `RitualFreeString`.
Payloads are JSON arrays. Results are JSON objects.

## Security

- No key storage — the key exists only during finalization
- Brute force cost: Argon2id (64 MB) → scrypt (128 MB) → BLAKE2b-256 per attempt
- Recommended minimum entropy: 80 bits across the ritual

See the specification for full threat model.

## Authorship verification

The author's Ed25519 public key is embedded in the specification document.
To verify authorship, request a signed message from the author and verify with:
