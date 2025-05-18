package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/character-staff-sync/internal/eventing"
)

// serveAnimeCharacterCmd represents the serve-anime-character command
var serveAnimeCharacterCmd = &cobra.Command{
	Use:   "serve-anime-character",
	Short: "Start eventing pipeline for anime characters",
	Long: `Launches the event-driven pipeline responsible for syncing 
anime character data from Pulsar into the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running anime character eventing...")
		return eventing.EventingAnimeCharacter()
	},
}

func init() {
	rootCmd.AddCommand(serveAnimeCharacterCmd)

	// You can define additional flags here if needed in the future.
}
