package internal

import (
	"errors"
	"fmt"
	"time"
)

// Transaction merepresentasikan struktur data transaksi antar dompet XCOSH.
type Transaction struct {
	ID        string    // Hash unik transaksi
	Sender    string    // Alamat dompet pengirim (Dilithium Address)
	Recipient string    // Alamat dompet penerima (Dilithium Address)
	Amount    uint64    // Jumlah koin yang dikirim
	Timestamp int64     // Waktu transaksi dibuat
	Signature []byte    // Tanda tangan digital tahan kuantum (Dilithium Signature)
}

// Mempool menampung transaksi sah yang belum dimasukkan ke dalam blok.
type Mempool struct {
	PendingTxs []Transaction
}

// NewMempool menginisialisasi mempool baru.
func NewMempool() *Mempool {
	return &Mempool{
		PendingTxs: make([]Transaction, 0),
	}
}

// AddTransaction memvalidasi dan memasukkan transaksi ke dalam mempool.
func (m *Mempool) AddTransaction(tx Transaction) error {
	if tx.Amount <= 0 {
		return errors.New("jumlah transaksi harus lebih besar dari 0")
	}
	if tx.Sender == tx.Recipient {
		return errors.New("pengirim dan penerima tidak boleh alamat yang sama")
	}
	if len(tx.Signature) == 0 {
		return errors.New("transaksi ditolak: tidak memiliki tanda tangan digital yang sah")
	}

	m.PendingTxs = append(m.PendingTxs, tx)
	return nil
}

// GetBlockTransactions mengambil dan mengosongkan antrean mempool untuk dimasukkan ke blok baru.
func (m *Mempool) GetBlockTransactions() []Transaction {
	txs := m.PendingTxs
	m.PendingTxs = make([]Transaction, 0) // Reset mempool
	return txs
}
