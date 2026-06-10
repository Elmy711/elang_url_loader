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

// Fungsi untuk menampilkan banner "MY EAGLE" dengan warna Cyan selama 30 detik
func printBanner() {
	banner :=             "=== MY EAGLE URL LOADER ==="
	fmt.Println()
	
	// Total durasi target: 10 detik.
	// Kita buat 60 kali perulangan, di mana setiap perulangan memakan waktu total 500ms (0.5 detik).
	// 20 x 0.5 detik = 10 detik.
	totalLoops := 20 
	
	fmt.Println(ColorCyan + " Starting Engine MY EAGLE ......" + ColorReset)
	
	for i := 0; i < totalLoops; i++ {
		fmt.Print(ColorCyan)
		// Efek teks mengetik berjalan (kecepatan disesuaikan agar pas)
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(10 * time.Millisecond) // Total waktu mengetik ~280ms
		}
		
		// Sisa waktu dari 500ms dialokasikan untuk jeda sebelum teks dihapus
		time.Sleep(120 * time.Millisecond) 
		
		// Hapus baris teks di terminal untuk efek animasi berjalan/berkedip
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")
		
		// Jeda mati sebelum siklus berikutnya dimulai
		time.Sleep(100 * time.Millisecond) 
	}
	
	// Setelah 30 detik selesai, cetak banner permanen berwarna Cyan
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

	// Menyiapkan Custom HTTP Client dengan Bypass SSL & Fake User-Agent
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
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
					req, err := http.NewRequest("GET", targetURL, nil)
					if err != nil {
						mu.Lock()
						failCount++
						mu.Unlock()
						continue
					}

					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
					req.Header.Set("Accept", "*/*")

					resp, err := client.Do(req)
					mu.Lock()
					if err != nil {
						failCount++
					} else {
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							successCount++
						} else {
							failCount++
						}
						resp.Body.Close()
					}
					mu.Unlock()

					time.Sleep(10 * time.Millisecond) 
				}
			}
		}()
	}

	// Timer untuk menghentikan pengujian sesuai durasi input
	time.Sleep(time.Duration(durationSeconds) * time.Second)
	close(stopChan) 

	wg.Wait() 

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
