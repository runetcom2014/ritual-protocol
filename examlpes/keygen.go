package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/hkdf"
	"crypto/sha256"
	"io"

	bip39 "github.com/tyler-smith/go-bip39"
)

// deriveKey derives a subkey from master using HKDF with the given domain
func deriveKey(master [32]byte, domain string, length int) []byte {
	h := hkdf.New(sha256.New, master[:], nil, []byte("87d32c69ac183b7832e01cf5:RITUAL-V1:"+domain))
	out := make([]byte, length)
	io.ReadFull(h, out)
	return out
}

// KeyBundle holds all derived keys
type KeyBundle struct {
	Ed25519Private string `json:"ed25519_private"`
	Ed25519Public  string `json:"ed25519_public"`
	SSHPrivate     string `json:"ssh_private"`
	SSHPublic      string `json:"ssh_public"`
	AES256         string `json:"aes256"`
	BIP39Mnemonic  string `json:"bip39_mnemonic"`
	SSLCSR         string `json:"ssl_csr"`
	SSLPrivate     string `json:"ssl_private"`
}

// DeriveKeys derives all key formats from a 32-byte master key
func DeriveKeys(master [32]byte) (KeyBundle, error) {
	var bundle KeyBundle

	// --- Ed25519 ---
	seed := deriveKey(master, "ed25519", ed25519.SeedSize)
	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)

	privDER, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil { return bundle, fmt.Errorf("ed25519 private: %w", err) }
	bundle.Ed25519Private = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))

	pubDER, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil { return bundle, fmt.Errorf("ed25519 public: %w", err) }
	bundle.Ed25519Public = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))

	// --- SSH ---
	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil { return bundle, fmt.Errorf("ssh public: %w", err) }
	bundle.SSHPublic = string(ssh.MarshalAuthorizedKey(sshPub))

	sshPriv, err := ssh.MarshalPrivateKey(privKey, "")
	if err != nil { return bundle, fmt.Errorf("ssh private: %w", err) }
	bundle.SSHPrivate = string(pem.EncodeToMemory(sshPriv))

	// --- AES-256 ---
	aesKey := deriveKey(master, "aes256", 32)
	bundle.AES256 = hex.EncodeToString(aesKey)

	// --- BIP39 ---
	entropy := deriveKey(master, "bip39", 32)
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil { return bundle, fmt.Errorf("bip39: %w", err) }
	bundle.BIP39Mnemonic = mnemonic

	// --- SSL CSR (P-256) ---
	ecSeed := deriveKey(master, "ssl-p256", 32)
	ecKey, err := deterministicP256(ecSeed)
	if err != nil { return bundle, fmt.Errorf("ssl key: %w", err) }

	ecDER, err := x509.MarshalECPrivateKey(ecKey)
	if err != nil { return bundle, fmt.Errorf("ssl private: %w", err) }
	bundle.SSLPrivate = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: ecDER}))

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: "ritual-protocol"},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, ecKey)
	if err != nil { return bundle, fmt.Errorf("csr: %w", err) }
	bundle.SSLCSR = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER}))

	return bundle, nil
}

// deterministicP256 generates a P-256 key from a 32-byte seed
func deterministicP256(seed []byte) (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256()
	// use seed as scalar — ensure it's in range [1, n-1]
	k := new(big.Int).SetBytes(seed)
	n := curve.Params().N
	k.Mod(k, new(big.Int).Sub(n, big.NewInt(1)))
	k.Add(k, big.NewInt(1))

	kBytes := make([]byte, 32)
	k.FillBytes(kBytes)

	// use standard library via a seeded reader trick
	reader := &deterministicReader{data: kBytes, pos: 0}
	return ecdsa.GenerateKey(curve, reader)
}

type deterministicReader struct {
	data []byte
	pos  int
}

func (r *deterministicReader) Read(p []byte) (int, error) {
	n := copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		// pad with zeros if needed
		for i := n; i < len(p); i++ { p[i] = 0 }
		return len(p), nil
	}
	return n, nil
}

// unused — kept for reference
var _ = time.Now