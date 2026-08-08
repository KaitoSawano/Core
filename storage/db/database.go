package db

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type BlockStorage struct {
	DB *bolt.DB
}

// NewBlockStorage membuka atau membuat file database disk lokal (xcosh.db)
func NewBlockStorage(dbPath string) (*BlockStorage, error) {
	database, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka database disk: %v", err)
	}

	// Buat bucket untuk menyimpan blok dan metadata chain
	err = database.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("blocks"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("meta"))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BlockStorage{DB: database}, nil
}

// SaveBlock menyimpan data blok terstruktur (JSON) ke disk berdasarkan Index atau Hash
func (bs *BlockStorage) SaveBlock(key []byte, blockData []byte) error {
	return bs.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		return b.Put(key, blockData)
	})
}

// GetBlock mengambil data blok dari disk berdasarkan key
func (bs *BlockStorage) GetBlock(key []byte) ([]byte, error) {
	var data []byte
	err := bs.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		val := b.Get(key)
		if val != nil {
			data = make([]byte, len(val))
			copy(data, val)
		}
		return nil
	})
	return data, err
}

// Close menutup koneksi database secara aman
func (bs *BlockStorage) Close() {
	bs.DB.Close()
}
