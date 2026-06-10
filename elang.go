package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Fungsi untuk menampilkan banner "MY EAGLE" dengan efek animasi berjalan singkat
func printBanner() {
	banner := "=== MY EAGLE LOAD TESTER ==="
	fmt.Println()
	for i := 0; i < 3; i++ {
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(30 * time.Millisecond)
		}
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println(banner)
	fmt.Println("============================\n")
}

func main() {
	printBanner()

	// Input Parameter dari User
	var targetURL string
	var durationSeconds int
	var concurrency int

	fmt.Print("Masukkan URL Target : ")
	fmt.Scanln(&targetURL)

	// Validasi input URL sederhana
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	fmt.Print("Masukkan Durasi : ")
	fmt.Scanln(&durationSeconds)

	fmt.Print("Masukkan Concurrency: ")
	fmt.Scanln(&concurrency)

	fmt.Printf("\n[!] Memulai pengujian ke %s selama %d detik dengan %d worker...\n\n", targetURL, durationSeconds, concurrency)

	// Channel untuk mengontrol stop signal berdasarkan durasi
	stopChan := make(chan struct{})
	
	// Counter untuk statistik
	var successCount int64
	var failCount int64
	var mu sync.Mutex

	var wg sync.WaitGroup
	client := &http.Client{
		Timeout: 5 * time.Second, // Timeout request agar tidak menggantung
	}

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
					resp, err := client.Get(targetURL)
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
					// Beri sedikit jeda mikro agar CPU tidak 100% overload di lokal
					time.Sleep(10 * time.Millisecond) 
				}
			}
		}()
	}

	// Timer untuk menghentikan pengujian sesuai durasi input
	time.Sleep(time.Duration(durationSeconds) * time.Second)
	close(stopChan) // Mengirim sinyal stop ke semua goroutine

	wg.Wait() // Tunggu semua goroutine selesai merapikan diri

	// Menampilkan Hasil
	fmt.Println("\n=================================")
	fmt.Println("              STATISTIK              ")
	fmt.Println("=================================")
	fmt.Printf("Target URL      : %s\n", targetURL)
	fmt.Printf("Durasi          : %d detik\n", durationSeconds)
	fmt.Printf("Request Sukses  : %d\n", successCount)
	fmt.Printf("Request Gagal   : %d\n", failCount)
	total := successCount + failCount
	fmt.Printf("Total Requests  : %d\n", total)
	if total > 0 {
		fmt.Printf("Avg Requests/sec: %.2f\n", float64(total)/float64(durationSeconds))
	}
	fmt.Println("=================================")
}

