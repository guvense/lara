package callback

import (
	"testing"

	model "github.com/guvense/lara/internal/model"
	parserpkg "github.com/guvense/lara/internal/parser"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestPrepareKafkaHeaders(t *testing.T) {
	headers := map[string]string{"x-user": "${value::request::queryparams::user}"}
	p := &parserpkg.Parser{
		Request: parserpkg.RequestParser{
			Params: &map[string]string{"user": "john"},
		},
	}

	prepared := prepareKafkaHeaders(&headers, p)
	if len(prepared) != 1 {
		t.Fatalf("expected one header, got %d", len(prepared))
	}
	if prepared[0].Key != "x-user" || string(prepared[0].Value) != "john" {
		t.Fatalf("unexpected header content: %+v", prepared[0])
	}
}

func TestPrepareAMQPHeaders(t *testing.T) {
	headers := map[string]string{"trace-id": "abc-123"}
	p := &parserpkg.Parser{}

	prepared := prepareAMQPHeaders(&headers, p)
	val, ok := prepared["trace-id"]
	if !ok {
		t.Fatalf("expected trace-id header to exist")
	}
	if val != "abc-123" {
		t.Fatalf("expected abc-123, got %v", val)
	}
}

func TestPrepareAMQPHeadersNil(t *testing.T) {
	if table := prepareAMQPHeaders(nil, &parserpkg.Parser{}); table != nil {
		t.Fatalf("expected nil headers table")
	}
}

func TestSendMessageBrokerCallbackUnsupportedTypeDoesNotPanic(t *testing.T) {
	p := &parserpkg.Parser{}
	mb := model.MessageBroker{Type: "unknown"}

	sendMessageBrokerCallback(mb, p)
}

func TestSendRabbitMQMessageMissingConfigDoesNotPanic(t *testing.T) {
	p := &parserpkg.Parser{}
	mb := model.MessageBroker{
		Type: "rabbitmq",
	}

	sendRabbitMQMessage(mb, p)
}

func TestAMQPHeadersType(t *testing.T) {
	headers := map[string]string{"k": "v"}
	table := prepareAMQPHeaders(&headers, &parserpkg.Parser{})
	if _, ok := interface{}(table).(amqp.Table); !ok {
		t.Fatalf("expected amqp.Table type")
	}
}
