package internal

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	RPCPort      string
	P2PPort      string
	RPCUser      string
	RPCPassword  string
}

func LoadConfig(filename string) *Config {
	// Nilai default ala Bitcoin Core jika file conf tidak ditemukan
	cfg := &Config{
		RPCPort:     "19332",
		P2PPort:     "19333",
		RPCUser:     "",
		RPCPassword: "",
	}

	file, err := os.Open(filename)
	if err != nil {
		// Jika file conf belum ada, gunakan default
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Lewati baris kosong atau komentar (#)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "rpcport":
			cfg.RPCPort = val
		case "port":
			cfg.P2PPort = val
		case "rpcuser":
			cfg.RPCUser = val
		case "rpcpassword":
			cfg.RPCPassword = val
		}
	}

	return cfg
}
