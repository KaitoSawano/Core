package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const rpcServerURL = "http://localhost:19332"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("=============================================================")
		fmt.Println("                XCOSH CLI UTILITY (CLIENT)                   ")
		fmt.Println("=============================================================")
		fmt.Println("Penggunaan:")
		fmt.Println("  go run xcosh-cli.go status    - Mengecek status daemon node")
		fmt.Println("  go run xcosh-cli.go ping      - Cek koneksi ke daemon node")
		fmt.Println("=============================================================")
		return
	}

	command := os.Args[1]

	switch command {
	case "ping":
		resp, err := http.Get(rpcServerURL + "/ping")
		if err != nil {
			fmt.Printf("Gagal terhubung ke daemon XCOSH (Pastikan daemon aktif): %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		var buf [512]byte
		n, _ := resp.Body.Read(buf[:])
		fmt.Printf("[+] Respons Daemon: %s\n", string(buf[:n]))

	case "status":
		resp, err := http.Get(rpcServerURL + "/status")
		if err != nil {
			fmt.Printf("Gagal terhubung ke daemon XCOSH: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Gagal mengurai data JSON: %v\n", err)
			return
		}

		fmt.Println("================ [ STATUS NODE XCOSH ] ================")
		fmt.Printf("Koin          : %v\n", result["coin"])
		fmt.Printf("Versi         : %v\n", result["version"])
		fmt.Printf("Total Blok    : %v\n", result["blocks_count"])
		fmt.Printf("Kesulitan PoW : %v\n", result["difficulty"])
		fmt.Printf("Antrean Tx    : %v\n", result["mempool_size"])
		fmt.Printf("Port P2P      : %v\n", result["p2p_port"])
		fmt.Printf("Peer Aktif    : %v\n", result["active_peers"])
		fmt.Printf("Dompet Miner  : %v\n", result["miner_address"])
		fmt.Println("-------------------------------------------------------")

	default:
		fmt.Printf("Perintah tidak dikenal: '%s'. Ketik tanpa argumen untuk bantuan.\n", command)
	}
}
