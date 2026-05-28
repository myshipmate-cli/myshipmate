package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os/exec"
)

// sha1Sum computes SHA1 hex digest of data
func sha1Sum(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// checkDockerInstalled checks if Docker is available on the system
func checkDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// displayBanner prints the Shipmate banner
func displayBanner() {
	fmt.Println()
	fmt.Println("  ╔═══════════════════════════════════════╗")
	fmt.Println("  ║       🚀 SHIPMATE v0.1.0              ║")
	fmt.Println("  ║   The Smart Deployer for Developers   ║")
	fmt.Println("  ╚═══════════════════════════════════════╝")
	fmt.Println()
}
