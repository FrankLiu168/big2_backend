package entrycmd

import (
	"big2backend/client"

	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start the WebAPI server",
	Run: func(cmd *cobra.Command, args []string) {
		StartAI()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}

func StartClient() {
	cli := client.NewClient("ws://127.0.0.1:8080/ws")
	err := cli.Connect()
	if err != nil {
		print("connect error",err.Error())
		return
	}
	err = cli.Start()
	if err != nil {
		print("start error",err.Error())
		return
	}
}
