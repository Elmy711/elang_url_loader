package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ANSI escape codes untuk pewarnaan di terminal
const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[36m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

// Fungsi untuk menampilkan banner "MY EAGLE" dengan warna Cyan dan efek animasi berjalan
func printBanner() {
	banner := "=== MY EAGLE LOAD TESTER ==="
	fmt.Println()
	// Efek animasi teks berjalan/mengetik dengan warna Cyan
	for i := 0; i < 2; i++ {
		fmt.Print(ColorCyan)
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(20 * time.Millisecond)
		}
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")
		time.Sleep(150 * time.Millisecond)
	}
	
	// Cetak banner permanen dengan warna Cyan
	fmt.Println(ColorCyan + banner + ColorReset)
	fmt.Println(ColorCyan + "============================" + ColorReset + "\n")
}

func main() {
	printBanner()

	// Input Parameter dari User
	var targetURL string
	var durationSeconds int
	var concurrency int

	fmt.Print("Masukkan URL Target (contoh: https://websiteku.com): ")
	fmt.Scanln(&targetURL)

	// Validasi input URL sederhana
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	fmt.Print("Masukkan Durasi Pengujian (dalam detik): ")
	fmt.Scanln(&durationSeconds)

	fmt.Print("Masukkan Jumlah Concurrency (Jumlah worker/pembuat request): ")
	fmt.Scanln(&concurrency)

	fmt.Printf("\n["+ColorYellow+"!"+ColorReset+"] Memulai pengujian ke %s selama %d detik dengan %d worker...\n\n", targetURL, durationSeconds, concurrency)

	// Channel untuk mengontrol stop signal berdasarkan durasi
	stopChan := make(chan struct{})
	
	// Counter untuk statistik
	var successCount int64
	var failCount int64
	var mu sync.Mutex

	// Menyiapkan Custom HTTP Client dengan:
	// 1. InsecureSkipVerify: true (mengabaikan error sertifikat SSL/HTTPS)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second, // Timeout jika server lambat merespons
	}

	var wg sync.WaitGroup

	// Menjalankan worker pool sebesar jumlah concurrency
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					// Membuat objek request secara manual untuk menyisipkan header User-Agent
					req, err := http.NewRequest("GET", targetURL, nil)
					if err != nil {
						mu.Lock()
						failCount++
						mu.Unlock()
						continue
					}

					// 2. Menyamar sebagai browser Chrome asli untuk menembus proteksi dasar/WAF
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
					req.Header.Set("Accept", "*/*")

					resp, err := client.Do(req)
					mu.Lock()
					if err != nil {
						// Jika Anda ingin melihat pesan error koneksi secara realtime, hilangkan komen baris di bawah:
						// fmt.Printf("["+ColorRed+"ERROR KONEKSI"+ColorReset+"]: %v\n", err)
						failCount++
					} else {
						// Jika response status berkisar di 2xx, dianggap sukses
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							successCount++
						} else {
							// Jika Anda ingin melihat status error server (403/429/502) secara realtime, hilangkan komen baris di bawah:
							// fmt.Printf("["+ColorRed+"ERROR SERVER"+ColorReset+"]: Status %d\n", resp.StatusCode)
							failCount++
						}
						resp.Body.Close()
					}
					mu.Unlock()

					// Jeda mikro 10ms agar resource CPU lokal stabil dan tidak hang
					time.Sleep(10 * time.Millisecond) 
				}
			}
		}()
	}

	// Timer untuk menghentikan pengujian sesuai durasi input
	time.Sleep(time.Duration(durationSeconds) * time.Second)
	close(stopChan) // Mengirim sinyal stop ke semua goroutine worker

	wg.Wait() // Tunggu semua goroutine selesai merapikan diri

	// Menampilkan Hasil Statistik Akhir
	fmt.Println("\n=================================")
	fmt.Println("              STATISTIK          ")
	fmt.Println("=================================")
	fmt.Printf("Target URL      : %s\n", targetURL)
	fmt.Printf("Durasi          : %d detik\n", durationSeconds)
	fmt.Printf("Request "+ColorGreen+"Sukses"+ColorReset+"  : %d\n", successCount)
	fmt.Printf("Request "+ColorRed+"Gagal"+ColorReset+"   : %d\n", failCount)
	total := successCount + failCount
	fmt.Printf("Total Requests  : %d\n", total)
	if total > 0 {
		fmt.Printf("Avg Requests/sec: %.2f\n", float64(total)/float64(durationSeconds))
	}
	fmt.Println("=================================")
}
