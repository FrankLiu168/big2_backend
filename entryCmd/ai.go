package entrycmd

import (
	"big2backend/ai"

	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Start the WebAPI server",
	Run: func(cmd *cobra.Command, args []string) {
		StartAI()
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
}

func StartAI() {
	ai.NewTransferMQ().Start()
}