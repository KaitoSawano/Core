package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RPCServer merepresentasikan peladen HTTP API untuk node XCOSH
type RPCServer struct {
	Port         string
	GetInfoFunc  func() map[string]interface{}
	AddNodeFunc  func(addr string) // Callback untuk menangani perintah addnode
}

// NewRPCServer menginisialisasi instance RPC server baru
func NewRPCServer(port string, getInfo func() map[string]interface{}, addNode func(addr string)) *RPCServer {
	return &RPCServer{
		Port:        port,
		GetInfoFunc: getInfo,
		AddNodeFunc: addNode,
	}
}

// Start menjalankan peladen HTTP di background
func (s *RPCServer) Start() {
	mux := http.NewServeMux()

	// Endpoint untuk melihat status node dan rantai blok
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := s.GetInfoFunc()
		json.NewEncoder(w).Encode(info)
	})

	// Endpoint health check
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("XCOSH Daemon RPC Server is Online!"))
	})

	// Endpoint baru untuk addnode
	mux.HandleFunc("/addnode", func(w http.ResponseWriter, r *http.Request) {
		addr := r.URL.Query().Get("addr")
		if addr == "" {
			http.Error(w, "Parameter 'addr' diperlukan", http.StatusBadRequest)
			return
		}

		if s.AddNodeFunc != nil {
			s.AddNodeFunc(addr)
			fmt.Fprintf(w, "Berhasil mencoba menghubungkan ke peer: %s", addr)
		} else {
			http.Error(w, "AddNode handler tidak dikonfigurasi", http.StatusInternalServerError)
		}
	})

	fmt.Printf("[+] RPC Server aktif mendengarkan pada port %s...\n", s.Port)
	go func() {
		if err := http.ListenAndServe(":"+s.Port, mux); err != nil {
			fmt.Printf("Gagal menjalankan RPC Server: %v\n", err)
		}
	}()
}
