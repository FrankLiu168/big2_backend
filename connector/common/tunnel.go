package common

type ClientTunnel struct {
	HandleMessage       func(client *Client, message []byte)
	BroadcastMessage    func(message []byte)
	SendMessageToClient func(clientID string, message []byte)
}

type TransferTunnel struct {
	HandleMessage func(message []byte)
}
