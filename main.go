package main

import (
	"big2backend/api"
	"big2backend/connector"
	"big2backend/game/logic"
	"big2backend/infrastructure/rabbitmq"
	"big2backend/ai"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
        log.Println("未找到 .env 檔案，使用系統環境變數")
    }
	startGame()
	
}

func testRabbitmq(){
	con,err := rabbitmq.NewConsumer("ex1")
	if err != nil {
		fmt.Println(err)
	}
	con.Listen([]string{"a01.b01","a01.b02"},handler)

	pro,err := rabbitmq.NewProducer("ex1")
	for i := 0; i < 10; i++ {
		if i % 2 == 0 {
			pro.Publish("a01.b02",fmt.Sprintf("message %d",i),"aaaa")
		} else {
			pro.Publish("a01.b01",fmt.Sprintf("message %d",i),"bbbb")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func handler(data amqp091.Delivery) {
	print(data.RoutingKey)
	print(string(data.Body))
}

func startGame(){
	aiServer := ai.NewTransferServer()
	aiServer.Start()
	logicServer := logic.GetTransferServer()
	logicServer.Start()
	deck := &logic.Deck{}
	p1 := logic.NewPlayer(1, "Player1", true)
	p2 := logic.NewPlayer(2, "Player2", true)
	p3 := logic.NewPlayer(3, "Player3", true)
	p4 := logic.NewPlayer(4, "Player4", true)
	players := []logic.Player{*p1, *p2, *p3, *p4}
	deck.Init(players)
	deck.StartGame()
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