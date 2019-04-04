package producer

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wlcn/yq-colly/common"

	"github.com/Shopify/sarama"
)

var brokerList = strings.Split(common.BrokerList, ",")

func init() {
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))
}

// SendSync is sync to send data to kafka
func SendSync(syncProducer sarama.SyncProducer, topic string, data interface{}) error {
	byteData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal data %+v", err)
		return err
	}
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	_, _, err = syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(byteData),
	})
	if err != nil {
		fmt.Printf("Failed to store your data:, %s", err)
		return err
	}
	return nil
}

// NewSyncProducer is to New SyncProducer
func NewSyncProducer() sarama.SyncProducer {
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

// CloseSync is to close syncProducer
func CloseSync(syncProducer sarama.SyncProducer) error {
	if err := syncProducer.Close(); err != nil {
		log.Println("Failed to shut down syncProducer cleanly", err)
		return err
	}
	return nil
}

// SendAsync is async to send data to kafka
func SendAsync(asyncProducer sarama.AsyncProducer, topic string, data interface{}) error {
	byteData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal data %+v", err)
		return err
	}
	asyncProducer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(byteData),
	}
	return nil
}

// NewAsyncProducer is to New AsyncProducer
func NewAsyncProducer() sarama.AsyncProducer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write kafka entry:", err)
		}
	}()
	return producer
}

// CloseAsync is to close asyncProducer
func CloseAsync(asyncProducer sarama.AsyncProducer) error {
	if err := asyncProducer.Close(); err != nil {
		log.Println("Failed to shut down asyncProducer cleanly", err)
		return err
	}
	return nil
}
