package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/character-staff-sync/internal/eventing"
)

// serveAnimeCharacterStaffLinkKafkaCmd represents the serve-anime-character-staff-link-kafka command
var serveAnimeCharacterStaffLinkKafkaCmd = &cobra.Command{
	Use:   "serve-anime-character-staff-link-kafka",
	Short: "Start Kafka eventing pipeline for anime character staff links",
	Long: `Launches the Kafka event-driven pipeline responsible for syncing 
anime character staff link data from Kafka into the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running anime character staff link Kafka eventing...")
		return eventing.EventingAnimeCharacterStaffLinkKafka()
	},
}

func init() {
	rootCmd.AddCommand(serveAnimeCharacterStaffLinkKafkaCmd)

	// You can define additional flags here if needed in the future.
}