package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const rpcServerURL = "http://localhost:19332"

func main() {
	// Jika dijalankan tanpa argumen, masuk ke menu interaktif
	if len(os.Args) < 2 {
		for {
			fmt.Println("\n=============================================================")
			fmt.Println("                XCOSH CLI UTILITY (CLIENT)                   ")
			fmt.Println("=============================================================")
			fmt.Println("  1. Cek Status Node (status)")
			fmt.Println("  2. Tes Koneksi Node (ping)")
			fmt.Println("  3. Tambah Peer Node (addnode)")
			fmt.Println("  4. Keluar (Exit)")
			fmt.Println("=============================================================")
			fmt.Print("Pilih menu [1-4]: ")

			var choice string
			if _, err := fmt.Scanln(&choice); err != nil {
				break
			}

			switch choice {
			case "1", "status":
				runStatus()
			case "2", "ping":
				runPing()
			case "3", "addnode":
				runAddNode()
			case "4", "exit":
				fmt.Println("Keluar dari CLI...")
				return
			default:
				fmt.Println("Pilihan tidak valid, silakan coba lagi.")
			}
		}
		return
	}

	// Jika dijalankan dengan argumen langsung (misal: xcosh-cli status atau xcosh-cli addnode)
	command := os.Args[1]
	switch command {
	case "ping":
		runPing()
	case "status":
		runStatus()
	case "addnode":
		runAddNode()
	default:
		fmt.Printf("Perintah tidak dikenal: '%s'. Ketik 'xcosh-cli' tanpa argumen untuk menu interaktif.\n", command)
	}
}

func runPing() {
	resp, err := http.Get(rpcServerURL + "/ping")
	if err != nil {
		fmt.Printf("\n[!] Gagal terhubung ke daemon XCOSH (Pastikan xcosh -daemon aktif): %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n[+] Respons Daemon: %s\n", string(body))
}

func runStatus() {
	resp, err := http.Get(rpcServerURL + "/status")
	if err != nil {
		fmt.Printf("\n[!] Gagal terhubung ke daemon XCOSH: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("\n[!] Gagal mengurai data JSON: %v\n", err)
		return
	}

	fmt.Println("\n================ [ STATUS NODE XCOSH ] ================")
	fmt.Printf("Koin          : %v\n", result["coin"])
	fmt.Printf("Versi         : %v\n", result["version"])
	fmt.Printf("Total Blok    : %v\n", result["blocks_count"])
	fmt.Printf("Kesulitan PoW : %v\n", result["difficulty"])
	fmt.Printf("Antrean Tx    : %v\n", result["mempool_size"])
	fmt.Printf("Port P2P      : %v\n", result["p2p_port"])
	fmt.Printf("Port RPC      : %v\n", result["rpc_port"])
	fmt.Printf("Peer Aktif    : %v\n", result["active_peers"])
	fmt.Printf("Dompet Miner  : %v\n", result["miner_address"])
	fmt.Println("-------------------------------------------------------")
}

func runAddNode() {
	var address string

	// Jika addnode dipanggil via argumen CLI (misal: xcosh-cli addnode 127.0.0.1:19333)
	if len(os.Args) > 2 {
		address = os.Args[2]
	} else {
		// Jika dipanggil via menu interaktif, minta input dari pengguna
		fmt.Print("\nMasukkan alamat peer (contoh: 192.168.1.50:19333): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		address = strings.TrimSpace(input)
	}

	if address == "" {
		fmt.Println("[!] Alamat peer tidak boleh kosong.")
		return
	}

	url := fmt.Sprintf("%s/addnode?addr=%s", rpcServerURL, address)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("\n[!] Gagal mengirim permintaan addnode ke daemon: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n[+] Respon Daemon: %s\n", string(body))
}
