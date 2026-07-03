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
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[36m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

// Global flag untuk mode debug
var debugMode bool

// Fungsi untuk menampilkan banner "MY EAGLE TOOLS" dengan warna Cyan dan animasi
func printBanner() {
	banner := "███╗   ███╗██╗   ██╗    ███████╗ █████╗  ██████╗ ██╗     ███████╗    
████╗ ████║╚██╗ ██╔╝    ██╔════╝██╔══██╗██╔════╝ ██║     ██╔════╝    
██╔████╔██║ ╚████╔╝     █████╗  ███████║██║  ███╗██║     █████╗      
██║╚██╔╝██║  ╚██╔╝      ██╔══╝  ██╔══██║██║   ██║██║     ██╔══╝      
██║ ╚═╝ ██║   ██║       ███████╗██║  ██║╚██████╔╝███████╗███████╗    
╚═╝     ╚═╝   ╚═╝       ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝"   
	fmt.Println()

	fmt.Println(ColorCyan + "💖💜 Starting MY EAGLE script 💜💖" + ColorReset)

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

		// Hapus banner yang tampil dengan menimpanya menggunakan '\r' dan spasi
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")

		time.Sleep(time.Duration(eraseSleepMs) * time.Millisecond)
	}

	// Tampilkan banner final dengan warna Cyan
	fmt.Println(ColorCyan + banner + ColorReset)
	fmt.Println(ColorCyan + "©©©©©©©©©©©©©©©©©©©©©©©©©©©©©" + ColorReset + "\n")
}

func main() {
	// Cek argumen command line untuk mode debug
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		debugMode = true
		fmt.Println(ColorYellow + "[DEBUG MODE AKTIF]" + ColorReset)
	}

	printBanner() // Panggil fungsi banner yang sudah diubah

	// Input Parameter dari User
	var targetURL string
	var durationSeconds int
	var concurrency int
	var httpMethod string
	var requestBody string
	var contentType string
	var clientTimeoutSeconds int

	fmt.Print("URL Target : ")
	fmt.Scanln(&targetURL)

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	fmt.Print("Durasi (detik) : ")
	fmt.Scanln(&durationSeconds)

	fmt.Print("Concurrency    : ")
	fmt.Scanln(&concurrency)

	fmt.Print("Metode HTTP (GET/POST/PUT/PATCH/DELETE, default GET): ")
	fmt.Scanln(&httpMethod)
	httpMethod = strings.ToUpper(strings.TrimSpace(httpMethod))
	if httpMethod == "" {
		httpMethod = "GET"
	}

	var bodyReader io.Reader
	if httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH" { // Method yang umumnya butuh body
		fmt.Print("Isi Request Body (kosongkan jika tidak ada): ")
		// Menggunakan io.ReadAll untuk menangani input multi-baris dengan spasi
		input, _ := io.ReadAll(os.Stdin)
		requestBody = strings.TrimSpace(string(input))

		if requestBody != "" {
			fmt.Print("Content-Type (default application/json): ")
			fmt.Scanln(&contentType)
			contentType = strings.TrimSpace(contentType)
			if contentType == "" {
				contentType = "application/json"
			}
			bodyReader = strings.NewReader(requestBody)
		}
	}

	// Input untuk Custom Headers
	customHeaders := make(map[string]string)
	fmt.Println("Tambahkan Header Kustom (ketik 'selesai' untuk berhenti):")
	for {
		var headerKey, headerValue string
		fmt.Print("  Key: ")
		fmt.Scanln(&headerKey)
		headerKey = strings.TrimSpace(headerKey)

		if strings.ToLower(headerKey) == "selesai" {
			break
		}
		if headerKey == "" {
			fmt.Println(ColorYellow + "  Key header tidak boleh kosong." + ColorReset)
			continue
		}

		fmt.Print("  Value: ")
		fmt.Scanln(&headerValue)
		customHeaders[headerKey] = strings.TrimSpace(headerValue)
	}

	fmt.Print("Timeout Request Klien (detik, default 10): ")
	fmt.Scanln(&clientTimeoutSeconds)
	if clientTimeoutSeconds <= 0 {
		clientTimeoutSeconds = 10
	}

	fmt.Printf("\n[💜💖] Mengirim request %s ke %s selama %d detik dengan %d worker...\n\n", httpMethod, targetURL, durationSeconds, concurrency)

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

	// Statistik Latency
	var totalLatency time.Duration
	var requestCountForLatency int64 // Hanya hitung request yang punya durasi terukur

	var mu sync.Mutex

	// Menyiapkan Custom HTTP Client dengan Bypass SSL & Fake User-Agent
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// Properti Keep-Alive sudah aktif secara default
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(clientTimeoutSeconds) * time.Second, // Timeout request per request
	}

	var wg sync.WaitGroup

	// Menjalankan worker pool sebesar jumlah concurrency
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		workerID := i // Buat salinan variabel i untuk goroutine
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					if debugMode {
						log.Printf("[DEBUG] Worker %d: Menerima sinyal berhenti.\n", workerID)
					}
					return
				default:
					if debugMode {
						log.Printf("[DEBUG] Worker %d: Membuat request %s ke %s\n", workerID, httpMethod, targetURL)
					}
					req, err := http.NewRequest(httpMethod, targetURL, bodyReader)
					if err != nil {
						mu.Lock()
						failCount++
						statusErr++
						if debugMode {
							log.Printf("[DEBUG] Worker %d: Gagal membuat request: %v\n", workerID, err)
						}
						mu.Unlock()
						time.Sleep(50 * time.Millisecond) // Jeda sedikit setelah error
						continue
					}

					// Set Header Default
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
					req.Header.Set("Accept", "*/*")
					// Tambahkan Content-Type jika request body ada
					if bodyReader != nil && contentType != "" {
						req.Header.Set("Content-Type", contentType)
					}

					// Tambahkan Custom Headers
					for key, value := range customHeaders {
						req.Header.Set(key, value)
					}

					startTime := time.Now()
					resp, err := client.Do(req)
					duration := time.Since(startTime)

					mu.Lock()
					if err != nil {
						failCount++
						statusErr++ // Kegagalan jaringan / TLS / Timeout
						if debugMode {
							log.Printf("[DEBUG] Worker %d: Request GAGAL (%v) dalam %s.\n", workerID, err, duration)
						}
					} else {
						code := resp.StatusCode

						// --- Penanganan Respons Berdasarkan Status Code ---
						if debugMode {
							log.Printf("[DEBUG] Worker %d: Menerima respons %d %s dalam %s.\n", workerID, code, http.StatusText(code), duration)
						}

						// Klasifikasi HTTP Status Code dan Tindakan Spesifik
						if code >= 200 && code < 300 { // 2xx - Sukses
							successCount++
							status2xx++
							totalLatency += duration
							requestCountForLatency++

							// Tindakan Spesifik untuk Sukses (2xx) - Debugging Body
							if debugMode {
								// Membaca sebagian kecil body untuk debug
								bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 100)) // Baca maks 100 byte
								if readErr != nil {
									log.Printf("[DEBUG] Worker %d: Gagal membaca sebagian body respons sukses: %v\n", workerID, readErr)
								} else {
									log.Printf("[DEBUG] Worker %d: Body respons (pertama 100 byte): %s...\n", workerID, string(bodyBytes))
								}
							}

						} else { // Bukan 2xx - Gagal (atau Redirection)
							failCount++

							if code >= 300 && code < 400 { // 3xx - Redirection
								status3xx++
								if debugMode {
									// Log header Location jika ada untuk redirect
									if location, locErr := resp.Location(); locErr == nil {
										log.Printf("[DEBUG] Worker %d: Redirected to %s\n", workerID, location.String())
									}
								}

							} else if code >= 400 && code < 500 { // 4xx - Client Error
								status4xx++
								// Tindakan Spesifik untuk Client Error (4xx) - Debugging Body Error
								if debugMode {
									bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 200)) // Baca maks 200 byte
									if readErr != nil {
										log.Printf("[DEBUG] Worker %d: Client Error (%d %s) - Gagal membaca body: %v\n", workerID, code, http.StatusText(code), readErr)
									} else {
										log.Printf("[DEBUG] Worker %d: Client Error (%d %s) - Body: %s\n", workerID, code, http.StatusText(code), string(bodyBytes))
									}
								}

							} else if code >= 500 && code < 600 { // 5xx - Server Error
								status5xx++
								// Tindakan Spesifik untuk Server Error (5xx) - Debugging Body Error
								if debugMode {
									bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 200)) // Baca maks 200 byte
									if readErr != nil {
										log.Printf("[DEBUG] Worker %d: Server Error (%d %s) - Gagal membaca body: %v\n", workerID, code, http.StatusText(code), readErr)
									} else {
										log.Printf("[DEBUG] Worker %d: Server Error (%d %s) - Body: %s\n", workerID, code, http.StatusText(code), string(bodyBytes))
									}
								}
							} else {
								// Status code yang tidak terduga
								if debugMode {
									log.Printf("[DEBUG] Worker %d: Status code tidak dikenal (%d %s) dalam %s.\n", workerID, code, http.StatusText(code), duration)
								}
							}
						}
						// Penting: Selalu tutup body respons untuk membebaskan koneksi
						resp.Body.Close()
					}
					mu.Unlock()

					// Sedikit jeda antar request per worker
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	// Timer untuk menghentikan pengujian sesuai durasi input
	stopTimer := time.NewTimer(time.Duration(durationSeconds) * time.Second)
	<-stopTimer.C // Tunggu hingga timer habis
	close(stopChan) // Kirim sinyal berhenti ke semua worker

	wg.Wait() // Tunggu semua worker selesai

	// Menampilkan Hasil Statistik Akhir yang Lebih Detail
	fmt.Println("\n" + ColorCyan + "=================================" + ColorReset)
	fmt.Println(ColorCyan + "              STATISTIK          " + ColorReset)
	fmt.Println(ColorCyan + "=================================" + ColorReset)
	fmt.Printf("Target URL      : %s\n", targetURL)
	fmt.Printf("Metode HTTP     : %s\n", httpMethod)
	if requestBody != "" {
		fmt.Printf("Content-Type    : %s\n", contentType)
		// Tampilkan sebagian kecil dari body jika terlalu panjang untuk laporan
		previewBody := requestBody
		if len(previewBody) > 60 { // Sesuaikan panjang preview
			previewBody = previewBody[:60] + "..."
		}
		fmt.Printf("Request Body    : %s\n", previewBody)
	}
	fmt.Printf("Durasi          : %d detik\n", durationSeconds)
	fmt.Printf("Concurrency     : %d\n", concurrency)
	fmt.Printf("Request "+ColorGreen+"Sukses"+ColorReset+"  : %d\n", successCount)
	fmt.Printf("Request "+ColorRed+"Gagal"+ColorReset+"   : %d\n", failCount)
	total := successCount + failCount
	fmt.Printf("Total Requests  : %d\n", total)
	if total > 0 && durationSeconds > 0 {
		fmt.Printf("Avg Requests/sec: %.2f\n", float64(total)/float64(durationSeconds))
	}

	// Bagian Fitur Baru: HTTP Status Breakdown
	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	fmt.Println(ColorCyan + "       HTTP STATUS BREAKDOWN     " + ColorReset)
	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	fmt.Printf("  2xx (Sukses/OK)       : %d\n", status2xx)
	fmt.Printf("  3xx (Redirection)     : %d\n", status3xx)
	fmt.Printf("  4xx (Client Error)    : %d\n", status4xx)
	fmt.Printf("  5xx (Server Error)    : %d\n", status5xx)
	fmt.Printf("  Network/Timeout Error : %d\n", statusErr)

	// Bagian Fitur Baru: Statistik Latency
	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	fmt.Println(ColorCyan + "      LATENCY STATISTICS       " + ColorReset)
	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	if requestCountForLatency > 0 {
		avgLatency := totalLatency / time.Duration(requestCountForLatency)
		fmt.Printf("Rata-rata Latency (Sukses) : %s\n", avgLatency)
	} else {
		fmt.Println("Tidak ada request sukses untuk menghitung latency.")
	}
	fmt.Println(ColorCyan + "=================================" + ColorReset)
}
