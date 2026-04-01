package rabbitmq

import (
	"fmt"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

// Consumer 封装 RabbitMQ 消费者
type Consumer struct {
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	exName string
	queue  amqp091.Queue // 值类型即可，无需指针
}
func getConnStr() string {
	url := os.Getenv("MQ_URL")
	user := os.Getenv("MQ_USER")
	pass := os.Getenv("MQ_PASS")
	return fmt.Sprintf("amqp://%s:%s@%s", user, pass, url)
}
// NewConsumer 创建消费者实例
func NewConsumer(exchangeName string) (*Consumer, error) {
	conn, err := amqp091.Dial(getConnStr())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明 topic exchange（幂等）
	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// 声明临时队列
	q, err := ch.QueueDeclare(
		"",    // name (auto-generated)
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:   conn,
		ch:     ch,
		exName: exchangeName,
		queue:  q,
	}, nil
}

// Close 关闭资源
func (c *Consumer) Close() error {
	var errs []error
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors during consumer close: %v", errs)
	}
	return nil
}

// Listen 绑定 routing keys 并启动消费 goroutine
// handler 用于处理每条消息（注意：若需 ACK，应关闭 autoAck 并在 handler 中手动确认）
func (c *Consumer) Listen(bindingKeys []string, handler func(*amqp091.Delivery)) error {
	// 绑定所有 routing key patterns
	for _, key := range bindingKeys {
		err := c.ch.QueueBind(
			c.queue.Name,
			key,
			c.exName, // ✅ 使用结构体字段，而非硬编码
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue with key %s: %w", key, err)
		}
		log.Printf("Bound queue %s to exchange %s with pattern: %s", c.queue.Name, c.exName, key)
	}

	// 启动消费者（auto-ack = true）
	msgs, err := c.ch.Consume(
		c.queue.Name,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	// 在 goroutine 中处理消息
	go func() {
		for d := range msgs {
			print("\n-------------\n")
			print(d.RoutingKey + "\n")
			print(string(d.Body))
			print("\n-------------\n")
			handler(&d)
		}
		log.Println("Message channel closed. Consumer stopped.")
	}()

	return nil
}
