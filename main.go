package main

import (
	"big2backend/ai"
	"big2backend/api"
	"big2backend/connector"
	transfermq "big2backend/connector/transferMQ"
	entrycmd "big2backend/entryCmd"
	"big2backend/game/logic"
	"big2backend/infrastructure/rabbitmq"
	"big2backend/shared/consts"
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 檔案，使用系統環境變數")
	}
	args := os.Args[1:]
	if args[0] == "-c" {
		startByClientCommand()
	} else {
		startByServerCommand()
	}
	
}

func startByServerCommand() {
	go entrycmd.StartAI()
	go entrycmd.StartConnector()
	entrycmd.StartGame()
}
func startByClientCommand() {
	entrycmd.StartClient()
}

func testAgentAndConnector() {
	agent := ai.NewTransferMQ()
	connector := transfermq.GetTransferMQ()
	agent.Start()
	connector.Start()
	agent.Publish(consts.ROUTING.CONNECTOR.FROM_AGENT, "hello", "1234", "3456")
	print("wait connector handler")
	for {
		time.Sleep(1 * time.Second)
	}
}
func testRabbitmq() {
	con, err := rabbitmq.NewConsumer("ex1")
	if err != nil {
		fmt.Println(err)
	}
	con.Listen([]string{"a01.b01", "a01.b02"}, handler)

	pro, err := rabbitmq.NewProducer("ex1")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			pro.Publish("a01.b02", fmt.Sprintf("message %d", i), "aaaa", "CCCC")
		} else {
			pro.Publish("a01.b01", fmt.Sprintf("message %d", i), "bbbb", "DDDD")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func handler(data *amqp091.Delivery) {
	print(data.RoutingKey)
	print(string(data.Body))
}

func startGame() {
	aiServer := ai.NewTransferMQ()
	aiServer.Start()
	logicServer := logic.GetTransferMQ()
	logicServer.Start()
	//connectorServer := connector.GetTransferMQ()
	//connectorServer.Start()
	http.HandleFunc("/ws", connector.HandleWebSocket)
	deck := &logic.Deck{}
	p1 := logic.NewPlayer(1, "Player1", false, logicServer)
	p2 := logic.NewPlayer(2, "Player2", true, logicServer)
	p3 := logic.NewPlayer(3, "Player3", true, logicServer)
	p4 := logic.NewPlayer(4, "Player4", true, logicServer)
	players := []logic.Player{*p1, *p2, *p3, *p4}
	deck.Init(players, logicServer)
	deck.StartGame()
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startAPIServer() {
	api.StartAPI()
}

func startSocketServer() {
	http.HandleFunc("/ws", connector.HandleWebSocket)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
