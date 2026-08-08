package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"
	
	"xcosh/internal"
	// Mengimpor modul wallet dari direktori storage/wallet Anda
	"xcosh/storage/wallet"
)

// Block merepresentasikan struktur data satu blok dalam blockchain mandiri.
type Block struct {
	Index        int64     // Tinggi/nomor blok
	Timestamp    int64     // Waktu blok dibuat (Unix Epoch)
	Data         []byte    // Payload transaksi atau data blok
	PrevHash     []byte    // Hash dari blok sebelumnya (memastikan immutability)
	Hash         []byte    // Hash unik blok ini hasil Keccak-256
	Nonce        int64     // Angka penambangan Proof-of-Work
	Difficulty   uint      // Tingkat kesulitan (jumlah byte nol di awal hash)
	MinerAddress string    // Alamat dompet penambang (tahan kuantum/Dilithium)
}

// Blockchain merepresentasikan struktur rantai utama dengan pengaman konkurensi (Mutex).
type Blockchain struct {
	mu         sync.Mutex
	blocks     []*Block
	difficulty uint
}

// CalculateKeccak256 menghasilkan hash Keccak-256 murni standar industri.
func CalculateKeccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// SetHash menghitung ulang dan mengisi field Hash berdasarkan header blok saat ini.
func (b *Block) SetHash() {
	timestampBytes := []byte(strconv.FormatInt(b.Timestamp, 10))
	indexBytes := []byte(strconv.FormatInt(b.Index, 10))
	nonceBytes := []byte(strconv.FormatInt(b.Nonce, 10))
	diffBytes := []byte(strconv.FormatUint(uint64(b.Difficulty), 10))

	headers := bytes.Join([][]byte{
		indexBytes,
		b.PrevHash,
		b.Data,
		timestampBytes,
		nonceBytes,
		diffBytes,
		[]byte(b.MinerAddress),
	}, []byte{})

	b.Hash = CalculateKeccak256(headers)
}

// MineBlock menjalankan algoritma Proof-of-Work (PoW) hingga hash memenuhi target kesulitan.
func (b *Block) MineBlock() {
	target := bytes.Repeat([]byte{0}, int(b.Difficulty))

	for {
		b.SetHash()
		if bytes.HasPrefix(b.Hash, target) {
			break
		}
		b.Nonce++
	}
}

// NewGenesisBlock membuat blok pertama (Genesis Block) secara hardcode dengan pengamanan mutlak.
func NewGenesisBlock(difficulty uint, minerAddress string) *Block {
	genesis := &Block{
		Index:        0,
		Timestamp:    1710000000,
		Data:         []byte("Genesis Block - Koin Mandiri Tahan Kuantum"),
		PrevHash:     []byte{0},
		Nonce:        0,
		Difficulty:   difficulty,
		MinerAddress: minerAddress,
	}
	genesis.MineBlock()
	return genesis
}

// NewBlockchain menginisialisasi blockchain baru dengan blok genesis.
func NewBlockchain(difficulty uint, minerAddress string) *Blockchain {
	return &Blockchain{
		blocks:     []*Block{NewGenesisBlock(difficulty, minerAddress)},
		difficulty: difficulty,
	}
}

// AddBlock memvalidasi dan menambang blok baru ke dalam rantai dengan aman (Thread-Safe).
func (bc *Blockchain) AddBlock(data string, minerAddress string) (*Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	prevBlock := bc.blocks[len(bc.blocks)-1]
	
	newBlock := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Data:         []byte(data),
		PrevHash:     prevBlock.Hash,
		Nonce:        0,
		Difficulty:   bc.difficulty,
		MinerAddress: minerAddress,
	}

	newBlock.MineBlock()

	if err := bc.validateBlock(newBlock, prevBlock); err != nil {
		return nil, err
	}

	bc.blocks = append(bc.blocks, newBlock)
	return newBlock, nil
}

// validateBlock memeriksa apakah blok baru sah secara kriptografi dan struktural.
func (bc *Blockchain) validateBlock(newBlock, prevBlock *Block) error {
	if newBlock.Index != prevBlock.Index+1 {
		return errors.New("indeks blok tidak berurutan dengan benar")
	}
	if !bytes.Equal(newBlock.PrevHash, prevBlock.Hash) {
		return errors.New("hash blok sebelumnya (PrevHash) tidak cocok")
	}
	
	target := bytes.Repeat([]byte{0}, int(newBlock.Difficulty))
	if !bytes.HasPrefix(newBlock.Hash, target) {
		return errors.New("bukti kerja (PoW) pada blok tidak valid atau belum memenuhi target")
	}

	return nil
}

// PrintChain mencetak seluruh informasi rantai blok ke terminal.
func (bc *Blockchain) PrintChain() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	fmt.Println("\n================ [ STATUS RANTAI BLOCKCHAIN ] ================")
	for _, block := range bc.blocks {
		fmt.Printf("Blok Index   : %d\n", block.Index)
		fmt.Printf("Timestamp    : %d\n", block.Timestamp)
		fmt.Printf("Data         : %s\n", string(block.Data))
		fmt.Printf("Prev Hash    : %x\n", block.PrevHash)
		fmt.Printf("Hash         : %x\n", block.Hash)
		fmt.Printf("Nonce        : %d\n", block.Nonce)
		fmt.Printf("Miner Wallet : %s\n", block.MinerAddress)
		fmt.Println("-------------------------------------------------------------")
	}
}

// Fungsi utama terintegrasi penuh dengan Dompet Dilithium
func main() {
	fmt.Println("=============================================================")
	fmt.Println("         MEMULAI DAEMON CORE XCOSH (POST-QUANTUM)            ")
	fmt.Println("=============================================================")

	// 1. Buat Dompet Penambang menggunakan Modul Dilithium
	minerWallet, err := wallet.NewWallet()
	if err != nil {
		fmt.Printf("Gagal membuat dompet miner: %v\n", err)
		return
	}

	fmt.Println("[+] Dompet Miner Berhasil Dibuat!")
	fmt.Printf("    Alamat (Address) : %s\n", minerWallet.GetAddress())
	fmt.Println("-------------------------------------------------------------")

	// 2. Inisialisasi Blockchain dengan Alamat Dompet Asli
	var initialDifficulty uint = 2
	myChain := NewBlockchain(initialDifficulty, minerWallet.GetAddress())

	// 3. Buat Dompet Alice & Bob untuk Simulasi Transaksi
	aliceWallet, _ := wallet.NewWallet()
	bobWallet, _ := wallet.NewWallet()

	fmt.Printf("[+] Dompet Alice : %s\n", aliceWallet.GetAddress())
	fmt.Printf("[+] Dompet Bob   : %s\n", bobWallet.GetAddress())
	fmt.Println("-------------------------------------------------------------")

	// 4. Simulasi Penandatanganan Transaksi Tahan Kuantum
	txData := fmt.Sprintf("Kirim 50 XCOSH dari %s ke %s", aliceWallet.GetAddress(), bobWallet.GetAddress())
	txHash := CalculateKeccak256([]byte(txData))

	signature, err := aliceWallet.SignTransaction(txHash)
	if err != nil {
		fmt.Printf("Gagal tanda tangan transaksi: %v\n", err)
		return
	}

	// 5. Verifikasi Keabsahan Tanda Tangan Dilithium
	isValid := wallet.VerifySignature(aliceWallet.PublicKey, txHash, signature)
	fmt.Printf("[+] Verifikasi Tanda Tangan Alice (Dilithium): %v\n", isValid)

	if isValid {
		_, err = myChain.AddBlock(txData, minerWallet.GetAddress())
		if err != nil {
			fmt.Printf("Gagal menambah blok: %v\n", err)
		}
	}

	// Cetak rantai blockchain akhir
	myChain.PrintChain()
}
