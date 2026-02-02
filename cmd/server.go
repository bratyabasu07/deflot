package cmd

import (
	"fmt"
	"os"

	"github.com/elliot/deflot/internal/server"

	"github.com/spf13/cobra"
)

var (
	serverAddr string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the DEFLOT Web Interface",
	Long:  "Launch localhost web interface for visual recon control",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	fmt.Printf("[*] Starting DEFLOT Web Interface\n")
	fmt.Printf("[*] Open your browser: http://%s\n", serverAddr)
	fmt.Printf("[*] Press Ctrl+C to stop\n\n")

	if err := server.Start(serverAddr); err != nil {
		fmt.Printf("[!] Server error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&serverAddr, "addr", "127.0.0.1:8080", "Server address")
}
