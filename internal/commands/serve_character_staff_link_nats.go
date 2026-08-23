package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/character-staff-sync/internal/eventing"
)

// serveCharacterStaffLinkNatsCmd is the NATS counterpart of serve-character-staff-link-kafka.
//
// A separate command rather than a flag: prod and staging run the same image,
// so the command name is what selects the transport, keeping that choice in the
// deployment values rather than an environment variable.
var serveCharacterStaffLinkNatsCmd = &cobra.Command{
	Use:   "serve-character-staff-link-nats",
	Short: "Consume character/staff link change events from NATS JetStream",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running character staff link eventing over NATS...")

		return eventing.EventingAnimeCharacterStaffLinkNats()
	},
}

func init() {
	rootCmd.AddCommand(serveCharacterStaffLinkNatsCmd)
}
