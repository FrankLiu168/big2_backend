package transfermq

import (

	"github.com/rabbitmq/amqp091-go"
)

type ConnectorTransferGameHelper struct {
	Transfer *ConnectorTransferMQ
	//connectorHandler func(dev *amqp091.Delivery)
}

func NewConnectorTransferGameHelper(transfer *ConnectorTransferMQ) *ConnectorTransferGameHelper {
	return &ConnectorTransferGameHelper{
		Transfer: transfer,
	}
}

func (c *ConnectorTransferGameHelper) HandleConnectMessage(tunnel *ConnectorTunnel, dev *amqp091.Delivery) {
	tunnel.ConnectorHandler(dev)
}


