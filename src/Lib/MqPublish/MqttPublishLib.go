package MqPublishLib

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"data-fetcher-api/src/Helper"
	"time"
	"github.com/getsentry/sentry-go"
)

const RABBITMQ_HOST = "mq-console.webcoder.io"
const RABBITMQ_PORT = 1883

func MqPublish(message string, topic string) {
	parameter, err := Helper.GetParameter()
	if err != nil {
		sentry.CaptureException(err)
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", RABBITMQ_HOST, RABBITMQ_PORT))
	opts.SetClientID(parameter.Parameters.RabbitMqClient)
	opts.SetUsername(parameter.Parameters.RabbitMqUserName + ":" + parameter.Parameters.RabbitMqVhost)
	opts.SetPassword(os.Getenv("RABBITMQ_PASSWORD"))
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		sentry.CaptureException(token.Error())
	}
	publish(client, message, topic)

	client.Disconnect(250)
}

func publish(client mqtt.Client, message string, topic string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
	time.Sleep(time.Second)
}
