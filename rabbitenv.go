package rabbitenv

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/streadway/amqp"
)

var queueconfig amqp.Table
var connection *amqp.Connection
var channel *amqp.Channel
var failLogger func(e string)

// failOnError handles default errors
func failOnError(err error, msg string) {
	if err != nil {
		errMsg := fmt.Sprintf("%s: %s", msg, err)
		if failLogger == nil {
			log.Fatalf("%s: %s", msg, err)
		} else {
			failLogger(errMsg)
		}
	}
}

// SetFailLogger function should be used to specify fail logger if required
func SetFailLogger(failFunction func(e string)) {
	failLogger = failFunction
}

// GetConfig returns parameters for a queue connection
func GetConfig() amqp.Table {
	if queueconfig == nil {
		fillConfig()
	}

	return queueconfig
}

// add bool config
func boolConf(key string, value string) {
	var parseErr error
	queueconfig[key], parseErr = strconv.ParseBool(value)
	failOnError(parseErr, `Failed to convert "`+key+`".`)
}

// initialize configuration
func fillConfig() {
	queueconfig["port"] = os.Getenv("RABBITMQ_PORT")
	if queueconfig["port"] == "" {
		queueconfig["port"] = "5672"
	}

	queueconfig["host"] = os.Getenv("RABBITMQ_HOST")
	if queueconfig["host"] == "" {
		queueconfig["host"] = "localhost"
	}

	queueconfig["queue"] = os.Getenv("RABBITMQ_QUEUE")
	if queueconfig["queue"] == "" {
		queueconfig["queue"] = "signature"
	}

	queueconfig["user"] = os.Getenv("RABBITMQ_USER")
	if queueconfig["user"] == "" {
		queueconfig["user"] = "guest"
	}

	ack := os.Getenv("RABBITMQ_ACKNOWLADGE")
	if ack == "" {
		ack = "false"
	}
	boolConf("ack", ack)

	noLocal := os.Getenv("RABBITMQ_NOLOCAL")
	if noLocal == "" {
		noLocal = "false"
	}
	boolConf("noLocal", noLocal)

	queueconfig["pass"] = os.Getenv("RABBITMQ_PASS")
	if queueconfig["pass"] == "" {
		queueconfig["pass"] = "guest"
	}

	// durable
	durable := os.Getenv("RABBITMQ_DURABLE")
	if durable == "" {
		durable = "true"
	}
	boolConf("durable", durable)

	// autodelete
	autoDelete := os.Getenv("RABBITMQ_AUTODELETE")
	if autoDelete == "" {
		autoDelete = "false"
	}
	boolConf("autoDelete", autoDelete)

	// no-wait
	noWait := os.Getenv("RABBITMQ_NOWAIT")
	if noWait == "" {
		noWait = "false"
	}
	boolConf("noWait", noWait)

	// exclusive
	exclusive := os.Getenv("RABBITMQ_EXCLUSIVE")
	if exclusive == "" {
		exclusive = "false"
	}
	boolConf("exclusive", exclusive)

	queueconfig["exchange"] = os.Getenv("RABBITMQ_EXCHANGE")
	if queueconfig["exchange"] == "" {
		queueconfig["exchange"] = "exchange1"
	}

	queueconfig["consumer"] = ""

}

// Connect provides connection to the queue
// returns connected amqp Channel
func Connect(qConfig amqp.Table) (*amqp.Connection, error) {
	adress := fmt.Sprintf(
		"amqp://%v:%v@%v:%v/",
		qConfig["user"],
		qConfig["pass"],
		qConfig["host"],
		qConfig["port"],
	)
	log.Println(adress)

	return amqp.Dial(adress)
}

// Channel returned
func Channel(conn *amqp.Connection, qConfig amqp.Table) *amqp.Channel {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	_, err = ch.QueueDeclare(
		qConfig["queue"].(string),
		qConfig["durable"].(bool),
		qConfig["autoDelete"].(bool),
		qConfig["exclusive"].(bool),
		qConfig["noWait"].(bool),
		nil, // arguments
	)
	failOnError(err, "Failed to declare a queue")

	return ch
}

// Listen the Queue
func Listen() (<-chan amqp.Delivery, error) {
	var err error
	conf := GetConfig()
	connection, err = Connect(conf)
	failOnError(err, "Failed to Connect")

	channel = Channel(connection, conf)
	return channel.Consume(
		conf["queue"].(string),
		conf["consumer"].(string),
		conf["ack"].(bool),
		conf["exclusive"].(bool),
		conf["noLocal"].(bool),
		conf["noWait"].(bool),
		nil, // args
	)
}

// Publish to the Queue
func Publish(msg interface{}) error {
	var err error
	conf := GetConfig()
	connection, err = Connect(conf)
	failOnError(err, "Failed to Connect")

	channel = Channel(connection, conf)
	channel.QueueDeclare(
		conf["queue"].(string),
		conf["durable"].(bool),
		conf["autoDelete"].(bool),
		conf["exclusive"].(bool),
		conf["noWait"].(bool),
		nil, // args
	)

	return nil
}

// Close closes all connections
func Close() {
	channel.Close()
	connection.Close()
}
