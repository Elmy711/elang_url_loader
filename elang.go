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
	banner := "==== MY EAGLE OVERLOAD TEST ===="
	fmt.Println()
	
	// Total durasi target: 10 detik.
	// 20 kali perulangan x 500ms = 10 detik.
	totalLoops := 10 
	
	fmt.Println(ColorCyan + "💖💜 Starting MY EAGLE script 💜💖" + ColorReset)
	
	for i := 0; i < totalLoops; i++ {
		fmt.Print(ColorCyan)
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(10 * time.Millisecond) 
		}
		
		time.Sleep(120 * time.Millisecond) 
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")
		time.Sleep(100 * time.Millisecond) 
	}
	
	fmt.Println(ColorCyan + banner + ColorReset)
	fmt.Println(ColorCyan + "==================================" + ColorReset + "\n")
}

func main() {
	printBanner()

	// Input Parameter dari User
	var targetURL string
	var durationSeconds int
	var concurrency int

	fmt.Print("URL Target : ")
	fmt.Scanln(&targetURL)

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	fmt.Print("Durasi : ")
	fmt.Scanln(&durationSeconds)

	fmt.Print("Concurrency: ")
	fmt.Scanln(&concurrency)

	fmt.Printf("\n[💜💖] Sending test ke %s selama %d detik dengan %d worker...\n\n", targetURL, durationSeconds, concurrency)

	stopChan := make(chan struct{})
	
	// Counter Statistik Utama
	var successCount int64
	var failCount int64
	
	// Counter Spesifik HTTP Status Code
	var status2xx int64
	var status3xx int64
	var status4xx int64
	var status5xx int64
	var statusErr int64 // Untuk error koneksi jaringan/timeout (tidak ada status code)

	var mu sync.Mutex

	// Menyiapkan Custom HTTP Client dengan Bypass SSL & Fake User-Agent
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
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
						statusErr++
						mu.Unlock()
						continue
					}

					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
					req.Header.Set("Accept", "*/*")

					resp, err := client.Do(req)
					mu.Lock()
					if err != nil {
						failCount++
						statusErr++ // Kegagalan jaringan / TLS / Timeout
					} else {
						code := resp.StatusCode
						
						// Klasifikasi HTTP Status Code
						if code >= 200 && code < 300 {
							successCount++
							status2xx++
						} else {
							failCount++
							if code >= 300 && code < 400 {
								status3xx++
							} else if code >= 400 && code < 500 {
								status4xx++
							} else if code >= 500 && code < 600 {
								status5xx++
							}
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

	// Menampilkan Hasil Statistik Akhir yang Lebih Detail
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
	
	// Bagian Fitur Baru: HTTP Status Breakdown
	fmt.Println("---------------------------------")
	fmt.Println("       HTTP STATUS BREAKDOWN     ")
	fmt.Println("---------------------------------")
	fmt.Printf("  2xx (Sukses/OK)       : %d\n", status2xx)
	fmt.Printf("  3xx (Redirection)     : %d\n", status3xx)
	fmt.Printf("  4xx (Client Error)    : %d\n", status4xx)
	fmt.Printf("  5xx (Server Error)    : %d\n", status5xx)
	fmt.Printf("  Network/Timeout Error : %d\n", statusErr)
	fmt.Println("=================================")
}
