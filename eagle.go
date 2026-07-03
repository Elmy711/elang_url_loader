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

// Fungsi untuk menampilkan banner "MY EAGLE TOOLS" dengan animasi
func printBanner() {
	banner := `███╗     ███╗██╗   ██╗ ███████╗ █████╗ ██████╗ ██╗ ███████╗
              ████╗  ████║╚██╗ ██╔╝ ██╔════╝██╔══██╗██╔════╝ ██║ ██╔════╝
              ██╔████╔██║  ╚████╔╝ █████╗ ███████║██║ ███╗██║ █████╗
              ██║╚██╔╝██║  ╚██╔╝ ██╔══╝ ██╔══██║██║ ██║██║ ██╔══╝
              ██║ ╚═╝  ██║    ██║ ███████╗██║ ██║╚██████╔╝███████╗███████╗
              ╚═╝ ╚═╝  ╚═╝          ╚══════╝╚═╝ ╚═╝ ╚═════╝ ╚══════╝╚══════╝`

	// Animasi sederhana untuk banner
	totalLoops := 1 // 3x aja biar nggak lama. Ganti 1 kalau mau langsung
	charSleepMs := 1 // 1ms biar cepet
	gapSleepMs := 80
	eraseSleepMs := 80

	for i := 0; i < totalLoops; i++ {
		fmt.Print(ColorCyan)
	// Tampilkan karakter satu per satu
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(time.Duration(charSleepMs) * time.Millisecond)
	}
		time.Sleep(time.Duration(gapSleepMs) * time.Millisecond)
		
		// Hapus banner
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")
		time.Sleep(time.Duration(eraseSleepMs) * time.Millisecond)
	}

	// Tampilkan banner final
	fmt.Println(ColorCyan + banner + ColorReset)
	fmt.Println()
	fmt.Println(ColorCyan + "💖💜 Starting MY EAGLE script 💜💖" + ColorReset)
	fmt.Println(ColorCyan + "©©©©©©©©©©©©©©©©©©©©©©©©" + ColorReset + "\n")
}

func main() {
	// Cek argumen command line untuk mode debug
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		debugMode = true
		fmt.Println(ColorYellow + "[DEBUG MODE AKTIF]" + ColorReset)
	}

	printBanner()

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

	if!strings.HasPrefix(targetURL, "http://") &&!strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	fmt.Print("Durasi (detik) : ")
	fmt.Scanln(&durationSeconds)

	fmt.Print("Concurrency : ")
	fmt.Scanln(&concurrency)

	fmt.Print("Metode HTTP (GET/POST/PUT/PATCH/DELETE, default GET): ")
	fmt.Scanln(&httpMethod)
	httpMethod = strings.ToUpper(strings.TrimSpace(httpMethod))
	if httpMethod == "" {
		httpMethod = "GET"
	}

	var bodyReader io.Reader
	if httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH" {
		fmt.Print("Isi Request Body (kosongkan jika tidak ada): ")
		input, _ := io.ReadAll(os.Stdin)
		requestBody = strings.TrimSpace(string(input))

		if requestBody!= "" {
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
		fmt.Print(" Key: ")
		fmt.Scanln(&headerKey)
		headerKey = strings.TrimSpace(headerKey)

		if strings.ToLower(headerKey) == "selesai" {
			break
	}
		if headerKey == "" {
			fmt.Println(ColorYellow + " Key header tidak boleh kosong." + ColorReset)
			continue
		}

		fmt.Print(" Value: ")
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

	var successCount int64
	var failCount int64
	var status2xx, status3xx, status4xx, status5xx, statusErr int64
	var totalLatency time.Duration
	var requestCountForLatency int64
	var mu sync.Mutex

	// HTTP Client Bypass SSL
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	Timeout: time.Duration(clientTimeoutSeconds) * time.Second,
	}

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					req, err := http.NewRequest(httpMethod, targetURL, bodyReader)
					if err!= nil {
						mu.Lock()
						failCount++
						statusErr++
						mu.Unlock()
						time.Sleep(50 * time.Millisecond)
						continue
					}

					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
					req.Header.Set("Accept", "*/*")
					if bodyReader!= nil && contentType!= "" {
						req.Header.Set("Content-Type", contentType)
					}
					for key, value := range customHeaders {
						req.Header.Set(key, value)
					}

					startTime := time.Now()
					resp, err := client.Do(req)
					duration := time.Since(startTime)

					mu.Lock()
					if err!= nil {
						failCount++
						statusErr++
						if debugMode {
							log.Printf("[DEBUG] Worker %d: GAGAL (%v)\n", workerID, err)
						}
					} else {
						code := resp.StatusCode
						if code >= 200 && code < 300 {
							successCount++
							status2xx++
							totalLatency += duration
							requestCountForLatency++
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

	stopTimer := time.NewTimer(time.Duration(durationSeconds) * time.Second)
	<-stopTimer.C
	close(stopChan)
	wg.Wait()

	// Statistik Akhir
	fmt.Println("\n" + ColorCyan + "=================================" + ColorReset)
	fmt.Println(ColorCyan + " STATISTIK " + ColorReset)
	fmt.Println(ColorCyan + "=================================" + ColorReset)
	fmt.Printf("Target URL : %s\n", targetURL)
	fmt.Printf("Metode HTTP : %s\n", httpMethod)
	fmt.Printf("Durasi : %d detik\n", durationSeconds)
	fmt.Printf("Concurrency : %d\n", concurrency)
	fmt.Printf("Request "+ColorGreen+"Sukses"+ColorReset+" : %d\n", successCount)
	fmt.Printf("Request "+ColorRed+"Gagal"+ColorReset+" : %d\n", failCount)
	total := successCount + failCount
	fmt.Printf("Total Requests : %d\n", total)
	if total > 0 && durationSeconds > 0 {
		fmt.Printf("Avg Requests/sec: %.2f\n", float64(total)/float64(durationSeconds))
	}

	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	fmt.Println(ColorCyan + " HTTP STATUS BREAKDOWN " + ColorReset)
	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	fmt.Printf(" 2xx (Sukses) : %d\n", status2xx)
	fmt.Printf(" 3xx (Redirection) : %d\n", status3xx)
	fmt.Printf(" 4xx (Client Error) : %d\n", status4xx)
	fmt.Printf(" 5xx (Server Error) : %d\n", status5xx)
	fmt.Printf(" Network/Timeout Error : %d\n", statusErr)

	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	fmt.Println(ColorCyan + " LATENCY STATISTICS " + ColorReset)
	fmt.Println(ColorCyan + "---------------------------------" + ColorReset)
	if requestCountForLatency > 0 {
		avgLatency := totalLatency / time.Duration(requestCountForLatency)
		fmt.Printf("Rata-rata Latency (Sukses) : %s\n", avgLatency)
	} else {
		fmt.Println("Tidak ada request sukses untuk menghitung latency.")
	}
	fmt.Println(ColorCyan + "=================================" + ColorReset)
}
