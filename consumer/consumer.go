package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/wlcn/yq-colly/common"

	"github.com/Shopify/sarama"
)

func init() {
	// 加载配置文件
}

// NewConsumerGroup 创建消费者
func main() {
	log.Println("Starting a new Sarama consumer")
	version, err := sarama.ParseKafkaVersion(common.KafkaVersion)
	if err != nil {
		log.Fatalf("ParseKafkaVersion err %+v", err)
	}
	brokerList := strings.Split(common.BrokerList, ",")
	config := sarama.NewConfig()
	config.Version = version
	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool, 0),
	}
	ctx := context.Background()
	client, err := sarama.NewConsumerGroup(brokerList, common.GroupID, config)
	if err != nil {
		log.Fatalf("NewConsumerGroup error %v", err)
	}
	go func() {
		for {
			err := client.Consume(ctx, strings.Split(common.Topic, ","), &consumer)
			if err != nil {
				log.Fatalf("consumer err %+v", err)
			}
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm // Await a sigterm signal before safely closing the consumer

	err = client.Close()
	if err != nil {
		log.Fatalf("close error %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		var data Data
		json.Unmarshal(message.Value, data)
		// 入库sql
		// 入库es
		session.MarkMessage(message, "")
	}

	return nil
}
