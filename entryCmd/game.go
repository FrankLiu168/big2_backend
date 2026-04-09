package entrycmd

import (
	"big2backend/game/logic"

	"github.com/spf13/cobra"
)

var gameCmd = &cobra.Command{
	Use:   "game",
	Short: "Start the WebAPI server",
	Run: func(cmd *cobra.Command, args []string) {
		StartGame()
	},
}

func init() {
	rootCmd.AddCommand(gameCmd)
}

func StartGame() {
	logicServer := logic.GetTransferMQ()
	logicServer.Start()
	deck := &logic.Deck{}
	p1 := logic.NewPlayer(1, "Player1", false, logicServer)
	p2 := logic.NewPlayer(2, "Player2", true, logicServer)
	p3 := logic.NewPlayer(3, "Player3", true, logicServer)
	p4 := logic.NewPlayer(4, "Player4", true, logicServer)
	players := []logic.Player{*p1, *p2, *p3, *p4}
	deck.Init(players, logicServer)
	deck.StartGame()
}
