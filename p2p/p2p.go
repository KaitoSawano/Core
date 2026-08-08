package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Message mendefinisikan struktur paket data yang dikirim antar node P2P
type Message struct {
	Type    string          `json:"type"`    // Contoh: "PING", "NEW_BLOCK", "NEW_TX"
	Payload json.RawMessage `json:"payload"` // Isi data sesuai tipe pesan
}

// P2PNode merepresentasikan instance node jaringan P2P
type P2PNode struct {
	ListenPort string
	Peers      map[string]net.Conn
	mu         sync.Mutex
	NewBlockCb func(blockBytes []byte)
}

// NewP2PNode menginisialisasi node P2P baru
func NewP2PNode(port string) *P2PNode {
	return &P2PNode{
		ListenPort: port,
		Peers:      make(map[string]net.Conn),
	}
}

// StartListener mulai mendengarkan koneksi masuk dari node lain
func (node *P2PNode) StartListener() {
	listener, err := net.Listen("tcp", ":"+node.ListenPort)
	if err != nil {
		fmt.Printf("Gagal memulai P2P listener pada port %s: %v\n", node.ListenPort, err)
		return
	}
	defer listener.Close()

	fmt.Printf("[+] P2P Node aktif mendengarkan koneksi masuk di port %s...\n", node.ListenPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go node.handleConnection(conn)
	}
}

// handleConnection mengelola data yang masuk dari peer yang terhubung
func (node *P2PNode) handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	node.mu.Lock()
	node.Peers[remoteAddr] = conn
	node.mu.Unlock()

	defer func() {
		node.mu.Lock()
		delete(node.Peers, remoteAddr)
		node.mu.Unlock()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg Message
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			continue
		}

		switch msg.Type {
		case "PING":
			response, _ := json.Marshal(Message{Type: "PONG"})
			conn.Write(append(response, '\n'))

		case "NEW_BLOCK":
			if node.NewBlockCb != nil {
				node.NewBlockCb(msg.Payload)
			}
		}
	}
}

// ConnectToPeer mencoba menghubungkan node ini ke node peer lain
func (node *P2PNode) ConnectToPeer(peerAddress string) error {
	conn, err := net.DialTimeout("tcp", peerAddress, 3*time.Second)
	if err != nil {
		return err
	}

	node.mu.Lock()
	node.Peers[peerAddress] = conn
	node.mu.Unlock()

	go node.handleConnection(conn)
	fmt.Printf("[+] Berhasil terhubung ke peer: %s\n", peerAddress)
	return nil
}

// Broadcast mengirimkan pesan ke seluruh peer yang aktif terhubung
func (node *P2PNode) Broadcast(msgType string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	msgBytes, err := json.Marshal(Message{
		Type:    msgType,
		Payload: data,
	})
	if err != nil {
		return
	}

	node.mu.Lock()
	defer node.mu.Unlock()

	for addr, conn := range node.Peers {
		_, err := conn.Write(append(msgBytes, '\n'))
		if err != nil {
			conn.Close()
			delete(node.Peers, addr)
		}
	}
}
