#1, start the server 
```shell
bin/zookeeper-server-start.sh config/zookeeper.properties bin/kafka-server-start.sh config/server.properties

win zookeeper-server-start.bat ../../config/zookeeper.properties kafka-server-start.bat ../../config/server.properties
```

#2, create a topic
```shell
bin/kafka-topics.sh --create --zookeeper 127.0.0.1:2181 --replication-factor 1 --partitions 1 --topic topic-test

kafka-topics.bat --create --zookeeper 127.0.0.1:2181 --replication-factor 1 --partitions 1 --topic topic-test

#see that topic 
bin/kafka-topics.sh --list --zookeeper 127.0.0.1:2181 

kafka-topics.bat --list --zookeeper 127.0.0.1:2181
```

#3, send some message 
```shell
bin/kafka-console-producer.sh --broker-list 127.0.0.1:9092 --topic topic-test 

kafka-console-producer.bat --broker-list 127.0.0.1:9092 --topic topic-test
```

#4, start a consumer 
```shell
bin/kafka-console-consumer.sh --bootstrap-server 127.0.0.1:9092 --topic topic-test --from-beginning 

kafka-console-consumer.bat --bootstrap-server 127.0.0.1:9092 --topic topic-test --from-beginning
```