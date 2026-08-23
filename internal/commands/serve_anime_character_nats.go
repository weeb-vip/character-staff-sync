package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/character-staff-sync/internal/eventing"
)

// serveAnimeCharacterNatsCmd is the NATS counterpart of serve-anime-character-kafka.
//
// A separate command rather than a flag: prod and staging run the same image,
// so the command name is what selects the transport, keeping that choice in the
// deployment values rather than an environment variable.
var serveAnimeCharacterNatsCmd = &cobra.Command{
	Use:   "serve-anime-character-nats",
	Short: "Consume anime character change events from NATS JetStream",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running anime character eventing over NATS...")

		return eventing.EventingAnimeCharacterNats()
	},
}

func init() {
	rootCmd.AddCommand(serveAnimeCharacterNatsCmd)
}
