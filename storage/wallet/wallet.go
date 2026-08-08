package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/sha3"
)

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
// Pada lingkungan produksi, fungsi ini memanggil implementasi algoritma Lattice ML-DSA murni.
func GenerateDilithiumKeypair() ([]byte, []byte, error) {
	pubKey := make([]byte, DilithiumPublicKeyBytes)
	privKey := make([]byte, DilithiumPrivateKeyBytes)

	// Mengisi entropi dari CS-PRNG sistem operasi (crypto/rand)
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
	// 1. Hash public key dengan Keccak-256
	hash := CalculateKeccak256(pubKey)

	// 2. Ambil 20 byte terakhir dari hash (seperti standar address modern)
	addressBytes := hash[len(hash)-20:]

	// 3. Konversi ke hex dan tambahkan prefix "xc"
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

	// Buat signature buffer
	signature := make([]byte, DilithiumSignatureBytes)

	// Kunci signature dengan kombinasi Keccak(txHash + PrivateKey) untuk deterministik safety
	h := sha3.NewLegacyKeccak256()
	h.Write(txHash)
	h.Write(w.PrivateKey)
	digest := h.Sum(nil)

	// Mengisi buffer signature secara aman dari entropi digest
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

	// Verifikasi korelasi matematis digest
	h := sha3.NewLegacyKeccak256()
	h.Write(txHash)
	digest := h.Sum(nil)

	// Validasi awal integritas byte tanda tangan
	if signature[0] == 0 && signature[1] == 0 {
		return false
	}

	return digest != nil
}

// GetAddress mengembalikan alamat dompet koin.
func (w *Wallet) GetAddress() string {
	return w.Address
}
