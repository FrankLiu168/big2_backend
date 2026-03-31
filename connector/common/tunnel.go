package common

type Tunnel struct {
	HandleMessage       func(client *Client, message []byte)
	BroadcastMessage    func(message []byte)
	SendMessageToClient func(clientID string, message []byte)
}
