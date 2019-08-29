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
var err error

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

// Config returns parameters for a queue connection
func Config(key string) interface{} {
	if queueconfig == nil {
		fillConfig()
	}

	return queueconfig[key]
}

// add bool config
func boolConf(key string, value string) {
	var parseErr error
	queueconfig[key], parseErr = strconv.ParseBool(value)
	failOnError(parseErr, `Failed to convert "`+key+`".`)
}

// initialize configuration
func fillConfig() {
	var ok bool
	queueconfig = amqp.Table{}

	queueconfig["port"], ok = os.LookupEnv("RABBITMQ_PORT")
	if ok == false {
		queueconfig["port"] = "5672"
	}

	queueconfig["host"], ok = os.LookupEnv("RABBITMQ_HOST")
	if ok == false {
		queueconfig["host"] = "localhost"
	}

	queueconfig["vhost"], ok = os.LookupEnv("RABBITMQ_VHOST")
	if ok == false {
		queueconfig["vhost"] = "/"
	}

	queueconfig["queue"] = os.Getenv("RABBITMQ_QUEUE_OWN")
	if queueconfig["queue"] == "" {
		queueconfig["queue"] = "signature"
	}

	queueconfig["queue_eng"] = os.Getenv("RABBITMQ_QUEUE_ENG")
	if queueconfig["queue_eng"] == "" {
		queueconfig["queue_eng"] = "write_files"
	}

	queueconfig["user"] = os.Getenv("RABBITMQ_USER")
	if queueconfig["user"] == "" {
		queueconfig["user"] = "guest"
	}

	queueconfig["pass"] = os.Getenv("RABBITMQ_PASS")
	if queueconfig["pass"] == "" {
		queueconfig["pass"] = "guest"
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

	// exchange
	queueconfig["exchange"], ok = os.LookupEnv("RABBITMQ_EXCHANGE")
	if ok == false {
		queueconfig["exchange"] = "exchange1"
	}

	//  exchange type
	queueconfig["exchangeType"] = os.Getenv("RABBITMQ_EXCHANGE_TYPE")
	if queueconfig["exchangeType"] == "" {
		queueconfig["exchangeType"] = "direct"
	}

	// exchange internal
	boolConf("internal", "false")

	// mandatory
	mandatory := os.Getenv("RABBITMQ_MANDATORY")
	if mandatory == "" {
		mandatory = "false"
	}
	boolConf("mandatory", mandatory)

	// immediate
	immediate := os.Getenv("RABBITMQ_IMMEDIATE")
	if immediate == "" {
		immediate = "false"
	}
	boolConf("immediate", immediate)

	queueconfig["consumer"] = ""

}

// Connect provides connection to the queue
// returns error if any
func Connect() error {

	adress := fmt.Sprintf(
		"amqp://%v:%v@%v:%v/",
		Config("user"),
		Config("pass"),
		Config("host"),
		Config("port"),
	)
	log.Println(adress)

	connection, err = amqp.Dial(adress)
	return err
}

// Channel forms channel
func Channel() error {
	if connection == nil {
		err = Connect()
		if err != nil {
			return err
		}
	}

	channel, err = connection.Channel()
	return err
}

// Exchange creation
func Exchange(name string) error {

	if channel == nil {
		err = Channel()
		if err != nil {
			return err
		}
	}

	return err
}

// Queue returnes declared Queue
func Queue(name string) (amqp.Queue, error) {

	if channel == nil {
		err = Channel()
		if err != nil {
			return amqp.Queue{}, err
		}
	}

	return channel.QueueDeclare(
		name,
		Config("durable").(bool),
		Config("autoDelete").(bool),
		Config("exclusive").(bool),
		Config("noWait").(bool),
		nil, // arguments
	)
}

// Listen the Queue
func Listen() (<-chan amqp.Delivery, error) {

	Queue(Config("queue").(string))
	if err != nil {
		return nil, err
	}

	return channel.Consume(
		Config("queue").(string),
		Config("consumer").(string),
		Config("ack").(bool),
		Config("exclusive").(bool),
		Config("noLocal").(bool),
		Config("noWait").(bool),
		nil, // args
	)
}

// Publish to the Queue with the environment configs
func Publish(msg amqp.Publishing) error {

	Exchange(Config("exchange").(string))
	if err != nil {
		return err
	}

	return channel.Publish(
		Config("exchange").(string),
		Config("queue_eng").(string),
		Config("mandatory").(bool),
		Config("immediate").(bool),
		msg,
	)
}

// Close all connections
func Close() {
	channel.Close()
	connection.Close()
}
