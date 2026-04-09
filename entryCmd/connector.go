package entrycmd

import (
	"big2backend/connector"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var connectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Start the WebAPI server",
	Run: func(cmd *cobra.Command, args []string) {
		StartConnector()
	},
}

func init() {
	rootCmd.AddCommand(connectorCmd)
}

func StartConnector() {
	connector.Init()
	http.HandleFunc("/ws", connector.HandleWebSocket)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
