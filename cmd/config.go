package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	vtKey string
	usKey string
	avKey string
	ghKey string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure API keys",
	Run: func(cmd *cobra.Command, args []string) {
		changed := false

		if vtKey != "" {
			viper.Set("api_keys.virustotal", vtKey)
			fmt.Println("[+] Set VirusTotal key")
			changed = true
		}
		if usKey != "" {
			viper.Set("api_keys.urlscan", usKey)
			fmt.Println("[+] Set URLScan key")
			changed = true
		}
		if avKey != "" {
			viper.Set("api_keys.alienvault", avKey)
			fmt.Println("[+] Set AlienVault key")
			changed = true
		}
		if ghKey != "" {
			viper.Set("api_keys.github", ghKey)
			fmt.Println("[+] Set GitHub key")
			changed = true
		}

		if changed {
			if err := viper.WriteConfig(); err != nil {
				// If config file doesn't exist, we might need to create it strictly,
				// but WriteConfig usually needs it to exist or SafeWriteConfig.
				// Our InitConfig ensures it exists? No, InitConfig reads it. GenerateDefaultConfig creates it.
				// If strictly failing, try WriteConfigAs.
				err = viper.SafeWriteConfig()
				if err != nil {
					err = viper.WriteConfig()
				}
				if err != nil {
					fmt.Printf("[!] Failed to save config: %v\n", err)
					os.Exit(1)
				}
			}
			fmt.Println("[*] Configuration updated successfully.")
		} else {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVar(&vtKey, "virustotal", "", "Set VirusTotal API Key")
	configCmd.Flags().StringVar(&usKey, "urlscan", "", "Set URLScan API Key")
	configCmd.Flags().StringVar(&avKey, "alienvault", "", "Set AlienVault OTX API Key")
	configCmd.Flags().StringVar(&ghKey, "github", "", "Set GitHub API Token")
}
