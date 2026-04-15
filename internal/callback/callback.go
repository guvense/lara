package callback

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	model "github.com/guvense/lara/internal/model"
	parser "github.com/guvense/lara/internal/parser"

	lara "github.com/guvense/lara/internal"
	auth "github.com/guvense/lara/internal/auth"

	amqp "github.com/rabbitmq/amqp091-go"
	kafka "github.com/segmentio/kafka-go"
)

type CallBack struct {
	TokenServers lara.TokenServers
}

func (s *CallBack) Call(callbacks []model.Callback, parser *parser.Parser) {
	for _, callback := range callbacks {
		if callback.CallbackRest != nil && callback.CallbackRest.Request.Endpoint != "" {

			var token string

			if callback.CallbackRest.AuthorizationKey != "" {
				tokenServers := s.TokenServers
				var err error
				token, err = auth.RetrieveAuthBearerToken(tokenServers[callback.CallbackRest.AuthorizationKey])

				if err != nil {
					log.Print("Error occurred preparing token please check token server credentials: %w", err)
				}
			}
			sendRestCallback(*callback.CallbackRest, token, parser)
		}

		if callback.MessageBroker != nil {
			sendMessageBrokerCallback(*callback.MessageBroker, parser)
		}
	}
}

func sendRestCallback(restCallback model.CallbackRest, token string, parser *parser.Parser) {

	if restCallback.Delay.Delay() > 0 {
		time.Sleep(restCallback.Delay.Delay())
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{Transport: tr}

	var jsonBody []byte

	if len(restCallback.Request.Body) > 0 {
		stringJson := fmt.Sprintf("%v", string(restCallback.Request.Body))
		body := parser.Parse(stringJson)
		jsonBody = []byte(body)
		log.Printf("Callback body : %s", body)
	}

	endPoint := parser.Parse(restCallback.Request.Endpoint)

	req, err := http.NewRequest(restCallback.Request.Method, endPoint, bytes.NewBuffer(jsonBody))

	if err != nil {
		fmt.Print("Error occurred while preparing request: %w", err)
	}

	req.Header = http.Header{}

	for key, val := range *restCallback.Request.Headers {
		req.Header.Add(key, val)
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+ token)
	}

	_, err = client.Do(req)

	if err != nil {
		fmt.Print("Error occurred while sending request server: %w", err)
	}

}

func sendMessageBrokerCallback(messageBroker model.MessageBroker, parser *parser.Parser) {
	if messageBroker.Delay.Delay() > 0 {
		time.Sleep(messageBroker.Delay.Delay())
	}

	typ := strings.ToLower(strings.TrimSpace(messageBroker.Type))
	switch typ {
	case "kafka":
		sendKafkaMessage(messageBroker, parser)
	case "rabbitmq", "rabbit-mq", "rabbit_mq":
		sendRabbitMQMessage(messageBroker, parser)
	default:
		log.Printf("unsupported message-broker callback type: %q", messageBroker.Type)
	}
}

func sendKafkaMessage(messageBroker model.MessageBroker, parser *parser.Parser) {
	if messageBroker.Kafka == nil {
		log.Print("kafka callback is missing kafka configuration")
		return
	}

	topic := parser.Parse(messageBroker.Kafka.Topic)
	if topic == "" {
		log.Print("kafka callback topic cannot be empty")
		return
	}

	brokers := messageBroker.Connection.Brokers
	if len(brokers) == 0 && messageBroker.Connection.URL != "" {
		brokers = []string{messageBroker.Connection.URL}
	}
	if len(brokers) == 0 {
		log.Print("kafka callback requires at least one broker")
		return
	}

	payload := parser.Parse(string(messageBroker.Message))
	key := parser.Parse(messageBroker.Kafka.Key)

	writer := &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: topic,
	}
	defer writer.Close()

	err := writer.WriteMessages(
		context.TODO(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(payload),
			Headers: prepareKafkaHeaders(messageBroker.Headers, parser),
		},
	)
	if err != nil {
		log.Printf("error while writing kafka message: %v", err)
	}
}

func sendRabbitMQMessage(messageBroker model.MessageBroker, parser *parser.Parser) {
	if messageBroker.RabbitMQ == nil {
		log.Print("rabbitmq callback is missing rabbitmq configuration")
		return
	}

	connURL := strings.TrimSpace(messageBroker.Connection.URL)
	if connURL == "" {
		log.Print("rabbitmq callback requires connection.url")
		return
	}

	conn, err := amqp.Dial(connURL)
	if err != nil {
		log.Printf("error while connecting rabbitmq: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("error while creating rabbitmq channel: %v", err)
		return
	}
	defer ch.Close()

	exchange := parser.Parse(messageBroker.RabbitMQ.Exchange)
	routingKey := parser.Parse(messageBroker.RabbitMQ.RoutingKey)
	if routingKey == "" {
		routingKey = parser.Parse(messageBroker.RabbitMQ.Queue)
	}

	body := parser.Parse(string(messageBroker.Message))

	if err := ch.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
			Headers:     prepareAMQPHeaders(messageBroker.Headers, parser),
		},
	); err != nil {
		log.Printf("error while publishing rabbitmq message: %v", err)
	}
}

func prepareKafkaHeaders(headers *map[string]string, parser *parser.Parser) []kafka.Header {
	if headers == nil {
		return nil
	}

	prepared := make([]kafka.Header, 0, len(*headers))
	for key, value := range *headers {
		prepared = append(prepared, kafka.Header{
			Key:   key,
			Value: []byte(parser.Parse(value)),
		})
	}
	return prepared
}

func prepareAMQPHeaders(headers *map[string]string, parser *parser.Parser) amqp.Table {
	if headers == nil {
		return nil
	}

	prepared := amqp.Table{}
	for key, value := range *headers {
		prepared[key] = parser.Parse(value)
	}
	return prepared
}
