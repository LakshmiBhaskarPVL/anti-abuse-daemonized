package banner

import (
	"fmt"
	"time"
)

const (
	// AppName is the name of the application
	AppName = "Sentinel"
	// CompanyName is the company/organization name
	CompanyName = "Novel"
	// Version is the application version
	Version = "1.0.0"
	// Website is the official website
	Website = "https://novel.sh"
	// DocsURL is the documentation URL
	DocsURL = "https://docs.novel.sh/sentinel"
)

// ASCII art banner for Sentinel
const Banner = `
   _____            __  _            __
  / ___/___  ____  / /_(_)___  ___  / /
  \__ \/ _ \/ __ \/ __/ / __ \/ _ \/ / 
 ___/ /  __/ / / / /_/ / / / /  __/ /  
/____/\___/_/ /_/\__/_/_/ /_/\___/_/   
`

// PrintBanner prints the application banner with professional formatting
func PrintBanner() {
	fmt.Print(Banner)
	fmt.Printf("\n                        by %s - v%s\n", CompanyName, Version)
	fmt.Printf("                Advanced Abuse Detection & Prevention System\n\n")
	fmt.Printf("  Website:  %s\n", Website)
	fmt.Printf("  Docs:     %s\n", DocsURL)
	fmt.Printf("  License:  MIT License\n")
	fmt.Printf("\nCopyright © 2025 - %d Novel & Contributors\n\n", time.Now().Year())
}
