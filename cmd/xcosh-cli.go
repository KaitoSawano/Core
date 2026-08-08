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

// Struktur untuk membaca rpchost dan rpcport dari xcosh.conf secara otomatis
func getRPCServerURL() string {
	rpcPort := "19332" // Default port RPC
	host := "127.0.0.1" // Default host RPC

	// Cek apakah ada file xcosh.conf di direktori lokal untuk membaca konfigurasi kustom
	if file, err := os.Open("xcosh.conf"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "rpchost=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					host = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "rpcport=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					rpcPort = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return fmt.Sprintf("http://%s:%s", host, rpcPort)
}

func main() {
	rpcServerURL := getRPCServerURL()

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
				runStatus(rpcServerURL)
			case "2", "ping":
				runPing(rpcServerURL)
			case "3", "addnode":
				runAddNode(rpcServerURL, "")
			case "4", "exit":
				fmt.Println("Keluar dari CLI...")
				return
			default:
				fmt.Println("Pilihan tidak valid, silakan coba lagi.")
			}
		}
		return
	}

	// Jika dijalankan langsung dengan argumen
	command := os.Args[1]
	var argVal string
	if len(os.Args) > 2 {
		argVal = os.Args[2]
	}

	switch command {
	case "ping":
		runPing(rpcServerURL)
	case "status":
		runStatus(rpcServerURL)
	case "addnode":
		runAddNode(rpcServerURL, argVal)
	default:
		fmt.Printf("Perintah tidak dikenal: '%s'. Ketik 'xcosh-cli' tanpa argumen untuk menu interaktif.\n", command)
	}
}

func runPing(serverURL string) {
	resp, err := http.Get(serverURL + "/ping")
	if err != nil {
		fmt.Printf("\n[!] Gagal terhubung ke daemon XCOSH: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n[+] Respons Daemon: %s\n", string(body))
}

func runStatus(serverURL string) {
	resp, err := http.Get(serverURL + "/status")
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

func runAddNode(serverURL string, address string) {
	if address == "" {
		fmt.Print("\nMasukkan alamat peer (contoh: 192.168.1.50:19333): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		address = strings.TrimSpace(input)
	}

	if address == "" {
		fmt.Println("[!] Alamat peer tidak boleh kosong.")
		return
	}

	url := fmt.Sprintf("%s/addnode?addr=%s", serverURL, address)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("\n[!] Gagal mengirim permintaan addnode ke daemon: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n[+] Respon Daemon: %s\n", string(body))
}
