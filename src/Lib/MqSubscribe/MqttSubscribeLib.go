package MqSubscribeLib

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Mq"
	"github.com/getsentry/sentry-go"
)

const RABBITMQ_HOST = "mq-console.webcoder.io"
const RABBITMQ_PORT = 1883

func MqSubscribe() {
	parameter, err := Helper.GetParameter()
	if err != nil {
		sentry.CaptureException(err)
	}
	
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", RABBITMQ_HOST, RABBITMQ_PORT))
	opts.SetClientID(parameter.Parameters.RabbitMqClient)
	opts.SetUsername(parameter.Parameters.RabbitMqUserName + ":" + parameter.Parameters.RabbitMqVhost)
	opts.SetPassword(os.Getenv("RABBITMQ_PASSWORD"))
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		sentry.CaptureException(token.Error())
	}
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", topic)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var payload = string(msg.Payload())
	var topic = msg.Topic()
	TopicsOperation(topic, payload)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	TopicsToSubscribe(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Mq Connection lost: %v", err)

}

func TopicsToSubscribe(client mqtt.Client) {
	// list out the topics here to subscribe
	topics := []string{
		"/api/tc/product/templates/create",
		"/api/tc/product/user/strategies",
		"/api/tc/product/templates/remove",
		"/api/tc/product/strategies/template/verify",
		"/api/tc/product/instruments/update",
	}

	for _, topic := range topics {
		sub(client, topic)
	}
}

func TopicsOperation(topic string, payload string) {
	// set the file where the operation of the topic has been made
	switch topic {
	case "/api/tc/product/templates/create":
		Mq.CreateExecutionAndPerformance(payload)
	case "/api/tc/product/instruments/update":
		Mq.CreateExecutionAndPerformance(payload)
	case "/api/tc/product/templates/remove":
		Mq.RemoveExecutionAndPerformance(payload)
	case "/api/tc/product/user/strategies":
		Mq.GetUserProfit(payload)
	case "/api/tc/product/strategies/template/verify":
		Mq.CreateStrategyExecution(payload)
	}
}
