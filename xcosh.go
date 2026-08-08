package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"

	"xcosh/internal"
	"xcosh/p2p"
	"xcosh/storage/db"
	"xcosh/storage/wallet"
)

// Block merepresentasikan struktur data satu blok dalam blockchain mandiri.
type Block struct {
	Index        int64                  // Tinggi/nomor blok
	Timestamp    int64                  // Waktu blok dibuat (Unix Epoch)
	Transactions []internal.Transaction // Payload transaksi sah dari mempool
	PrevHash     []byte                 // Hash dari blok sebelumnya (memastikan immutability)
	Hash         []byte                 // Hash unik blok ini hasil Keccak-256
	Nonce        int64                  // Angka penambangan Proof-of-Work
	Difficulty   uint                   // Tingkat kesulitan (jumlah byte nol di awal hash)
	MinerAddress string                 // Alamat dompet penambang (tahan kuantum/Dilithium)
}

// Blockchain merepresentasikan struktur rantai utama dengan pengaman konkurensi (Mutex) dan penyimpanan disk.
type Blockchain struct {
	Mu         sync.Mutex
	Blocks     []*Block
	Difficulty uint
	Mempool    *internal.Mempool
	Storage    *db.BlockStorage
	P2PNode    *p2p.P2PNode
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

	var txDataBytes []byte
	for _, tx := range b.Transactions {
		txDataBytes = append(txDataBytes, []byte(tx.ID+tx.Sender+tx.Recipient+strconv.FormatUint(tx.Amount, 10))...)
	}

	headers := bytes.Join([][]byte{
		indexBytes,
		b.PrevHash,
		txDataBytes,
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
		Transactions: []internal.Transaction{},
		PrevHash:     []byte{0},
		Nonce:        0,
		Difficulty:   difficulty,
		MinerAddress: minerAddress,
	}
	genesis.MineBlock()
	return genesis
}

// NewBlockchain menginisialisasi blockchain baru dengan blok genesis, mempool, disk database, dan p2p node.
func NewBlockchain(difficulty uint, minerAddress string, dbPath string, p2pPort string) (*Blockchain, error) {
	storage, err := db.NewBlockStorage(dbPath)
	if err != nil {
		return nil, err
	}

	genesis := NewGenesisBlock(difficulty, minerAddress)
	
	// Simpan genesis block ke disk
	blockBytes, _ := json.Marshal(genesis)
	_ = storage.SaveBlock([]byte(strconv.FormatInt(genesis.Index, 10)), blockBytes)

	p2pNode := p2p.NewP2PNode(p2pPort)

	chain := &Blockchain{
		Blocks:     []*Block{genesis},
		Difficulty: difficulty,
		Mempool:    internal.NewMempool(),
		Storage:    storage,
		P2PNode:    p2pNode,
	}

	// Tangani event ketika blok baru diterima dari jaringan P2P tetangga
	p2pNode.NewBlockCb = func(blockData []byte) {
		var incomingBlock Block
		if err := json.Unmarshal(blockData, &incomingBlock); err == nil {
			chain.Mu.Lock()
			defer chain.Mu.Unlock()
			
			prevBlock := chain.Blocks[len(chain.Blocks)-1]
			if incomingBlock.Index == prevBlock.Index+1 {
				chain.Blocks = append(chain.Blocks, &incomingBlock)
				
				// Simpan ke disk lokal
				blockBytes, _ := json.Marshal(&incomingBlock)
				_ = chain.Storage.SaveBlock([]byte(strconv.FormatInt(incomingBlock.Index, 10)), blockBytes)
				
				fmt.Printf("\n[P2P] Blok baru diterima & disinkronkan dari jaringan! Index: %d\n", incomingBlock.Index)
			}
		}
	}

	return chain, nil
}

// MinePendingTransactions mengambil transaksi dari mempool, menambangnya, menyimpannya ke disk, & broadcast P2P.
func (bc *Blockchain) MinePendingTransactions(minerAddress string) (*Block, error) {
	bc.Mu.Lock()
	defer bc.Mu.Unlock()

	pendingTxs := bc.Mempool.GetBlockTransactions()
	if len(pendingTxs) == 0 {
		return nil, errors.New("tidak ada transaksi di mempool untuk ditambang")
	}

	prevBlock := bc.Blocks[len(bc.Blocks)-1]

	newBlock := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: pendingTxs,
		PrevHash:     prevBlock.Hash,
		Nonce:        0,
		Difficulty:   bc.Difficulty,
		MinerAddress: minerAddress,
	}

	newBlock.MineBlock()

	if err := bc.validateBlock(newBlock, prevBlock); err != nil {
		return nil, err
	}

	bc.Blocks = append(bc.Blocks, newBlock)

	// Simpan blok baru secara permanen ke disk (BoltDB)
	blockBytes, err := json.Marshal(newBlock)
	if err == nil {
		_ = bc.Storage.SaveBlock([]byte(strconv.FormatInt(newBlock.Index, 10)), blockBytes)
	}

	// Broadcast blok baru ke seluruh peer P2P yang terhubung di jaringan
	go bc.P2PNode.Broadcast("NEW_BLOCK", newBlock)

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
	bc.Mu.Lock()
	defer bc.Mu.Unlock()

	fmt.Println("\n================ [ STATUS RANTAI BLOCKCHAIN XCOSH ] ================")
	for _, block := range bc.Blocks {
		fmt.Printf("Blok Index   : %d\n", block.Index)
		fmt.Printf("Timestamp    : %d\n", block.Timestamp)
		fmt.Printf("Jumlah Tx    : %d transaksi\n", len(block.Transactions))
		for idx, tx := range block.Transactions {
			fmt.Printf("  -> Tx [%d] : %s mengirim %d XCOSH ke %s\n", idx+1, tx.Sender[:12], tx.Amount, tx.Recipient[:12])
		}
		fmt.Printf("Prev Hash    : %x\n", block.PrevHash)
		fmt.Printf("Hash         : %x\n", block.Hash)
		fmt.Printf("Nonce        : %d\n", block.Nonce)
		fmt.Printf("Miner Wallet : %s\n", block.MinerAddress)
		fmt.Println("-------------------------------------------------------------------")
	}
}

// Fungsi utama daemon XCOSH terintegrasi penuh
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

	// 2. Inisialisasi Blockchain + Disk Database + P2P Node (Port 19333)
	var initialDifficulty uint = 2
	myChain, err := NewBlockchain(initialDifficulty, minerWallet.GetAddress(), "xcosh.db", "19333")
	if err != nil {
		fmt.Printf("Gagal menginisialisasi blockchain disk: %v\n", err)
		return
	}
	defer myChain.Storage.Close()

	// Nyalakan P2P Listener di background pada port 19333
	go myChain.P2PNode.StartListener()

	// 3. Buat Dompet Alice & Bob untuk Simulasi Transaksi
	aliceWallet, _ := wallet.NewWallet()
	bobWallet, _ := wallet.NewWallet()

	fmt.Printf("[+] Dompet Alice : %s\n", aliceWallet.GetAddress())
	fmt.Printf("[+] Dompet Bob   : %s\n", bobWallet.GetAddress())
	fmt.Println("-------------------------------------------------------------")

	// 4. Simulasi Penandatanganan Transaksi Tahan Kuantum
	txPayload := fmt.Sprintf("%s:%s:%d:%d", aliceWallet.GetAddress(), bobWallet.GetAddress(), uint64(50), time.Now().Unix())
	txHash := CalculateKeccak256([]byte(txPayload))

	signature, err := aliceWallet.SignTransaction(txHash)
	if err != nil {
		fmt.Printf("Gagal tanda tangan transaksi: %v\n", err)
		return
	}

	// 5. Verifikasi Keabsahan Tanda Tangan Dilithium
	isValid := wallet.VerifySignature(aliceWallet.PublicKey, txHash, signature)
	fmt.Printf("[+] Verifikasi Tanda Tangan Alice (Dilithium): %v\n", isValid)

	if isValid {
		tx := internal.Transaction{
			ID:        fmt.Sprintf("%x", txHash[:8]),
			Sender:    aliceWallet.GetAddress(),
			Recipient: bobWallet.GetAddress(),
			Amount:    50,
			Timestamp: time.Now().Unix(),
			Signature: signature,
		}

		// Masukkan transaksi ke mempool
		err = myChain.Mempool.AddTransaction(tx)
		if err != nil {
			fmt.Printf("Gagal memasukkan transaksi ke mempool: %v\n", err)
		} else {
			fmt.Println("[+] Transaksi berhasil masuk ke Mempool!")
		}

		// Tambang transaksi dalam mempool ke blok baru, simpan ke disk, & broadcast P2P
		fmt.Println("[*] Menambang blok baru, menyimpan ke disk, & broadcast P2P...")
		_, err = myChain.MinePendingTransactions(minerWallet.GetAddress())
		if err != nil {
			fmt.Printf("Gagal menambang blok: %v\n", err)
		} else {
			fmt.Println("[+] Blok transaksi berhasil ditambang & di-broadcast ke port 19333!")
		}
	}

	// Cetak rantai blockchain akhir
	myChain.PrintChain()

	// 6. Jalankan RPC/API Server di background (Port 8333)
	rpcServer := internal.NewRPCServer("8333", func() map[string]interface{} {
		myChain.Mu.Lock()
		defer myChain.Mu.Unlock()
		return map[string]interface{}{
			"coin":          "XCOSH",
			"version":       "1.0.0-post-quantum",
			"blocks_count":  len(myChain.Blocks),
			"difficulty":    myChain.Difficulty,
			"mempool_size":  len(myChain.Mempool.PendingTxs),
			"p2p_port":      "19333",
			"active_peers":  len(myChain.P2PNode.Peers),
			"miner_address": minerWallet.GetAddress(),
		}
	})
	rpcServer.Start()

	// Menjaga agar daemon terus berjalan sebagai proses latar depan (blocking)
	fmt.Println("[*] Daemon XCOSH berjalan penuh (P2P Port: 19333 | RPC Port: 8333)...")
	select {}
}
