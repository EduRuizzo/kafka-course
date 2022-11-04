# ZOOKEEPER
$KAFKA_HOME/bin/zookeeper-server-start.sh -daemon  $KAFKA_HOME/config/zookeeper.properties
tail -f $KAFKA_HOME/logs/zookeeper.out
$KAFKA_HOME/bin/zookeeper-server-stop.sh

# KAFKA SERVER
$KAFKA_HOME/bin/kafka-server-start.sh -daemon $KAFKA_HOME/config/server.properties
tail -f $KAFKA_HOME/logs/kafkaServer.out
$KAFKA_HOME/bin/kafka-server-stop.sh

# KAFDROP (UI) ([::1]:9092 si listeners=PLAINTEXT://[::1]:9092 en server.properties)
java -jar kafdrop-3.30.0.jar --kafka.brokerCOnnect=localhost:9092

# CREATE TOPIC
$KAFKA_HOME/bin/kafka-topics.sh --bootstrap-server localhost:9092 \
--create --partitions 3 --replication-factor 1 --topic getting-started

# PUBLISH RECORDS
$KAFKA_HOME/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic getting-started \
--property "parse.key=true" --property "key.separator=:"

# CONSUME RECORDS
$KAFKA_HOME/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic getting-started \
--group cli-consumer --from-beginning --property "print.key=true" --property "key.separator=:"

# LISTING TOPICS
$KAFKA_HOME/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list --exclude-internal

# DESCRIBE TOPIC
$KAFKA_HOME/bin/kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic getting-started

# DELETE TOPIC
$KAFKA_HOME/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic getting

# CONSUMER GROUPS (--delete)
$KAFKA_HOME/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list
$KAFKA_HOME/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-topics --group cli-consumer
$KAFKA_HOME/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-topics --all-groups --state

# RESETTING OFFSETS (also --to-earliest --to-latest --to-datetime --shift-by n)
$KAFKA_HOME/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --reset-offsets \
--topic getting-started --group cli-consumer --to-offset 1 --execute

# take a peek into the broker process to see these threads:
KAFKA_PID=$(jps -l | grep kafka.Kafka | awk '{print $1}') jstack $KAFKA_PID | grep data-plane-kafka-request-handler


# Lets use the --describe switch to list the dynamic value. The property name passed to --describe
# is optional. If omitted, all configuration entries will be shown.
kafka-configs.sh --bootstrap-server localhost:9092 --entity-type brokers --entity-name 0 --describe num.io.threads

# Let’s change the num.io.threads value by invoking the following command: (to broker 0)
kafka-configs.sh --bootstrap-server localhost:9092 --entity-type brokers --entity-name 0 --alter --add-config num.io.threads=4

# Let’s apply the num.io.threads.setting to the entire cluster.
kafka-configs.sh --bootstrap-server localhost:9092 --entity-type brokers --entity-default --alter --add-config num.io.threads=4

# remove the per-broker entry.
kafka-configs.sh --bootstrap-server localhost:9092 --entity-type brokers --entity-name 0 --alter --delete-config num.io.threads

# revert the configuration to its original state by deleting the cluster-wide entry:
kafka-configs.sh --bootstrap-server [::1]:9092 --entity-type brokers --entity-default --alter --delete-config num.io.threads

# apply an override for flush.messages for the test.topic.config topic.
kafka-configs.sh --bootstrap-server [::1]:9092 --entity-type topics --entity-name test.topic.config --alter --add-config flush.messages=100
kafka-configs.sh --bootstrap-server [::1]:9092 --entity-type topics --entity-name test.topic.config --describe flush.messages
# back to default
kafka-configs.sh --bootstrap-server [::1]:9092 --entity-type topics --entity-name test.topic.config --alter --delete-config flush.messages