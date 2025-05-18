package commands

import (
	"github.com/spf13/cobra"
	"github.com/weeb-vip/character-staff-sync/internal/eventing"
	"log"
)

var serveCharacterStaffLinkCmd = &cobra.Command{
	Use:   "serve-character-staff-link",
	Short: "Processes character-staff links from Pulsar and syncs with the database",
	Long: `This command runs a Pulsar consumer that listens for character-staff link events 
and syncs them into the database by resolving character and staff records.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running character-staff link eventing...")
		return eventing.EventingAnimeCharacterStaffLink()
	},
}

func init() {
	rootCmd.AddCommand(serveCharacterStaffLinkCmd)
}
