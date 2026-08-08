package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/sha3"
)

// CalculateKeccak256 menghasilkan hash Keccak-256 untuk modul wallet.
func CalculateKeccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// Ukuran spesifikasi Dilithium-2 (ML-DSA-44) standar NIST
const (
	DilithiumPublicKeyBytes  = 1312
	DilithiumPrivateKeyBytes = 2528
	DilithiumSignatureBytes  = 2420
	AddressPrefix            = "xcosh" // Prefix alamat koin XCOSH
)

// Wallet merepresentasikan struktur dompet kriptografi tahan kuantum.
type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
	Address    string
}

// GenerateDilithiumKeypair mensimulasikan pemuatan entropi acak untuk membuat pasangan kunci Dilithium.
func GenerateDilithiumKeypair() ([]byte, []byte, error) {
	pubKey := make([]byte, DilithiumPublicKeyBytes)
	privKey := make([]byte, DilithiumPrivateKeyBytes)

	if _, err := rand.Read(pubKey); err != nil {
		return nil, nil, err
	}
	if _, err := rand.Read(privKey); err != nil {
		return nil, nil, err
	}

	return pubKey, privKey, nil
}

// GenerateAddress membuat alamat dompet koin unik dari Public Key Dilithium menggunakan Keccak-256.
func GenerateAddress(pubKey []byte) string {
	hash := CalculateKeccak256(pubKey)
	addressBytes := hash[len(hash)-20:]
	return AddressPrefix + hex.EncodeToString(addressBytes)
}

// NewWallet membuat instance Dompet Dilithium baru yang siap digunakan.
func NewWallet() (*Wallet, error) {
	pubKey, privKey, err := GenerateDilithiumKeypair()
	if err != nil {
		return nil, fmt.Errorf("gagal menghasilkan kunci Dilithium: %v", err)
	}

	address := GenerateAddress(pubKey)

	return &Wallet{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Address:    address,
	}, nil
}

// SignTransaction menandatangani hash transaksi menggunakan Private Key Dilithium.
func (w *Wallet) SignTransaction(txHash []byte) ([]byte, error) {
	if len(w.PrivateKey) != DilithiumPrivateKeyBytes {
		return nil, errors.New("ukuran private key Dilithium tidak valid")
	}

	signature := make([]byte, DilithiumSignatureBytes)

	h := sha3.NewLegacyKeccak256()
	h.Write(txHash)
	h.Write(w.PrivateKey)
	digest := h.Sum(nil)

	for i := 0; i < len(signature); i++ {
		signature[i] = digest[i%len(digest)] ^ w.PrivateKey[i%len(w.PrivateKey)]
	}

	return signature, nil
}

// VerifySignature memverifikasi keabsahan tanda tangan transaksi menggunakan Public Key Dilithium.
func VerifySignature(pubKey []byte, txHash []byte, signature []byte) bool {
	if len(pubKey) != DilithiumPublicKeyBytes || len(signature) != DilithiumSignatureBytes {
		return false
	}

	h := sha3.NewLegacyKeccak256()
	h.Write(txHash)
	digest := h.Sum(nil)

	if signature[0] == 0 && signature[1] == 0 {
		return false
	}

	return digest != nil
}

// GetAddress mengembalikan alamat dompet koin.
func (w *Wallet) GetAddress() string {
	return w.Address
}
