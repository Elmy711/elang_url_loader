package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI escape codes untuk pewarnaan di terminal
const (
	ColorReset = "\033[0m"
	ColorCyan = "\033[36m"
	ColorRed = "\033[31m"
	ColorGreen = "\033[32m"
	ColorYellow = "\033[33m"
)

// Global flag untuk mode debug
var debugMode bool

// Fungsi untuk menampilkan banner "MY EAGLE TOOLS" dengan warna Cyan dan animasi
func printBanner() {
	banner := `███╗ ███╗██╗ ██╗ ███████╗ █████╗ ██████╗ ██╗ ███████╗
████╗ ████║╚██╗ ██╔╝ ██╔════╝██╔══██╗██╔════╝ ██║ ██╔════╝
██╔████╔██║ ╚████╔╝ █████╗ ███████║██║ ███╗██║ █████╗
██║╚██╔╝██║ ╚██╔╝ ██╔══╝ ██╔══██║██║ ██║██║ ██╔══╝
██║ ╚═╝ ██║ ██║ ███████╗██║ ██║╚██████╔╝███████╗███████╗
╚═╝ ╚═╝ ╚═╝ ╚══════╝╚═╝ ╚═╝ ╚═════╝ ╚══════╝╚══════╝`

	// Animasi sederhana untuk banner
	totalLoops := 10 // Jumlah pengulangan animasi banner
	charSleepMs := 10 // Durasi tampilan per karakter
	gapSleepMs := 120 // Durasi jeda antar karakter saat animasi
	eraseSleepMs := 100 // Durasi jeda setelah banner dihapus sebelum animasi berikutnya

	for i := 0; i < totalLoops; i++ {
		fmt.Print(ColorCyan)
	// Tampilkan karakter satu per satu
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(time.Duration(charSleepMs) * time.Millisecond)
		}

		time.Sleep(time.Duration(gapSleepMs) * time.Millisecond)
		
		// Hapus banner yang tampil dengan menimpanya menggunakan '\r' dllan spasi
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")

		time.Sleep(time.Duration(eraseSleepMs) * time.Millisecond)
	}

	// Tampilkan banner final dengan warna Cyan
	fmt.Println(ColorCyan + banner + ColorReset)
	fmt.Println()
	fmt.Println(ColorCyan + "💖💜 Starting MY EAGLE script 💜💖" + ColorReset)
	fmt.Println(ColorCyan + "©©©" + ColorReset + "\n")
}

func main() {
	// Cek argumen command line untuk mode debug
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		debugMode = true
		fmt.Println(ColorYellow + "[DEBUG MODE AKTIF]" + ColorReset)
	}

	printBanner() // Panggil fungsi banner yang sudah diubah

	//... sisa kode main lu biarin aja dari sini, nggak usah diubah...
	var targetURL string
	var durationSeconds int
	var concurrency int
	var httpMethod string
	var requestBody string
	var contentType string
	var clientTimeoutSeconds int

	fmt.Print("URL Target : ")
	fmt.Scanln(&targetURL)
	// dst...
}
