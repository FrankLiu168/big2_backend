package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// Producer 封装 RabbitMQ 生产者
type Producer struct {
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	exName string
}

// NewProducer 创建一个新的 Producer 实例
// 注意：不再接收 rKeys（除非你有特殊用途，如预绑定等）
func NewProducer(exchangeName string) (*Producer, error) {
	// TODO: 从配置或参数传入 DSN 更佳
	conn, err := amqp091.Dial(getConnStr())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // 清理连接
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = declareTopicExchange(ch, exchangeName)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare topic exchange: %w", err)
	}

	return &Producer{
		conn:   conn,
		ch:     ch,
		exName: exchangeName,
	}, nil
}

// Close 关闭 channel 和 connection，释放资源
func (p *Producer) Close() error {
	var errs []error
	if p.ch != nil {
		if err := p.ch.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

// declareTopicExchange 声明一个 durable 的 topic exchange
func declareTopicExchange(ch *amqp091.Channel, exchangeName string) error {
	return ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// Publish 发布消息到已声明的 topic exchange
func (p *Producer) Publish(routingKey, message, msgID string) error {
	return p.ch.Publish(
		p.exName, // 使用内部保存的 exchange name
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			MessageId:   msgID,
		},
	)
}
