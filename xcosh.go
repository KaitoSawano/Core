package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"
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
	mu           sync.Mutex
	blocks       []*Block
	difficulty   uint
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
	// Membuat target prefix berupa byte nol sejumlah b.Difficulty
	target := bytes.Repeat([]byte{0}, int(b.Difficulty))

	for {
		b.SetHash()
		// Periksa apakah byte awal hash sesuai dengan target kesulitan
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
		Timestamp:    1710000000, // Timestamp tetap untuk genesis
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

	// Lakukan penambangan PoW
	newBlock.MineBlock()

	// Validasi integritas blok sebelum dimasukkan ke rantai
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
	
	// Validasi ulang hash blok untuk memastikan tidak ada manipulasi data
	target := bytes.Repeat([]byte{0}, int(newBlock.Difficulty))
	if !bytes.HasPrefix(newBlock.Hash, target) {
		return errors.New("bukti kerja (PoW) pada blok tidak valid atau belum memenuhi target")
	}

	return nil
}

// GetLatestBlock mengembalikan blok terakhir yang ada di rantai dengan aman.
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.blocks[len(bc.blocks)-1]
}

// PrintChain mencetak seluruh informasi rantai blok ke terminal (untuk keperluan debugging node).
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

// Fungsi utama sementara untuk menguji modul file pertama ini
func main() {
	fmt.Println("Memulai Daemon Core Koin Mandiri...")

	// Tentukan tingkat kesulitan awal PoW (misal: 2 byte awal harus bernilai 0)
	var initialDifficulty uint = 2
	minerWallet := "dilithium_pubkey_testnet_xyz9812376"

	// Inisialisasi Blockchain
	myChain := NewBlockchain(initialDifficulty, minerWallet)

	// Tambah blok transaksi pertama
	_, err := myChain.AddBlock("Transaksi #1: Alice mengirim 50 Koin ke Bob", minerWallet)
	if err != nil {
		fmt.Printf("Gagal menambah blok: %v\n", err)
	}

	// Tambah blok transaksi kedua
	_, err = myChain.AddBlock("Transaksi #2: Bob mengirim 10 Koin ke Charlie", minerWallet)
	if err != nil {
		fmt.Printf("Gagal menambah blok: %v\n", err)
	}

	// Tampilkan seluruh rantai
	myChain.PrintChain()
}
