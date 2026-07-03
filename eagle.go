package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ANSI escape codes untuk pewarnaan di terminal
const (
	ColorReset = "\033[0m"
	ColorCyan = "\033[36m"
	ColorYellow = "\033[33m"
)

var debugMode bool

func printBanner() {
	banner := `███╗ ███╗██╗ ██╗ ███████╗ █████╗ ██████╗ ██╗ ███████╗
████╗ ████║╚██╗ ██╔╝ ██╔════╝██╔══██╗██╔════╝ ██║ ██╔════╝
██╔████╔██║ ╚████╔╝ █████╗ ███████║██║ ███╗██║ █████╗
██║╚██╔╝██║ ╚██╔╝ ██╔══╝ ██╔══██║██║ ██║██║ ██╔══╝
██║ ╚═╝ ██║ ██║ ███████╗██║ ██║╚██████╔╝███████╗███████╗
╚═╝ ╚═╝ ╚═╝ ╚══════╝╚═╝ ╚═╝ ╚═════╝ ╚══════╝╚══════╝`

	totalLoops := 3 // Gue kurangin biar cepet
	charSleepMs := 10
	gapSleepMs := 120
	eraseSleepMs := 100

	for i := 0; i < totalLoops; i++ {
		fmt.Print(ColorCyan)
		for j := 0; j < len(banner); j++ {
			fmt.Print(string(banner[j]))
			time.Sleep(time.Duration(charSleepMs) * time.Millisecond)
		}
		time.Sleep(time.Duration(gapSleepMs) * time.Millisecond)
		fmt.Print("\r" + strings.Repeat(" ", len(banner)) + "\r")
		time.Sleep(time.Duration(eraseSleepMs) * time.Millisecond)
	}

	fmt.Println(ColorCyan + banner + ColorReset)
	fmt.Println()
	fmt.Println(ColorCyan + "💖💜 Starting MY EAGLE script 💜💖" + ColorReset)
	fmt.Println(ColorCyan + "©©" + ColorReset + "\n")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		debugMode = true
		fmt.Println(ColorYellow + "[DEBUG MODE AKTIF]" + ColorReset)
	}
	printBanner()
}
