package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/Shopify/sarama"
)

var producer sarama.SyncProducer

func init() {
	brokerList := strings.Split(BrokerList, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))
	producer = newDataCollector(brokerList)
}

func send(topic string, data string) error {
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(data),
	})

	if err != nil {
		fmt.Printf("Failed to store your data:, %s", err)
		return err
	}
	// The tuple (topic, partition, offset) can be used as a unique identifier
	// for a message in a Kafka cluster.
	fmt.Printf("Your data is stored with unique identifier important/%d/%d", partition, offset)
	return nil
}

func newDataCollector(brokerList []string) sarama.SyncProducer {

	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	return producer
}
