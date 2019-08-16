package rabbitenv

import (
	"testing"

	"github.com/streadway/amqp"
)

// TestGetConfig tests rabbitenv.GetConfig
func TestGetConfig(t *testing.T) {

	if Config("queue") != "test" {
		t.Error("Config is incorrect")
	}

}

func TestPublish(t *testing.T) {
	body := "test"
	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	}

	err := Publish(msg)
	if err != nil {
		t.Error("Message is not sent")
	}
}
