package mqtt

import (
	"context"
	"os"
	"strings"

	mqttclient "github.com/eclipse/paho.mqtt.golang"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
	"github.com/vedga/alisa/pkg/eventbus"
)

const (
	// RxDiscoveryMQTT is events topic where service put received MQTT messages from discovery topic
	RxDiscoveryMQTT = "mqtt:discovery"
	// RxTelemetryMQTT is events topic where service put received MQTT messages from telemetry topic
	RxTelemetryMQTT = "mqtt:telemetry"
	// RxCommandsMQTT is events topic where service put received MQTT messages from command topic
	RxCommandsMQTT = "mqtt:commands"
	// RxStatusesMQTT is events topic where service put received MQTT messages from status topic
	RxStatusesMQTT = "mqtt:statuses"
)

const (
	// envMQTTBrokerURI is URI for connect to the MQTT server in form "tcp://host:port"
	envMQTTBrokerURI         = "MQTT_BROKER_URI"
	envMQTTUserName          = "MQTT_USER_NAME"
	envMQTTPassword          = "MQTT_PASSWORD"
	clientID                 = "alisa_service"
	waitDisconnectCompleteMS = 1000
	topicDiscovery           = "tasmota/discovery/#"
	topicTelemetry           = "tele/#"
	topicCommands            = "cmnd/#"
	topicStatuses            = "stat/#"
	topicPartsDelimiter      = "/"
)

// EventMQTT is MQTT event content
type EventMQTT struct {
	Topic   []string
	Payload string
}

// Service is MQTT client service implementation
type Service struct {
	runnable.Runnable
	bus    eventbus.Bus
	client mqttclient.Client
}

// NewService return new service implementation
// example: https://levelup.gitconnected.com/how-to-use-mqtt-with-go-89c617915774
// Official documentation: https://www.emqx.com/en/blog/how-to-use-mqtt-in-golang
// Tasmota MQTT: https://tasmota.github.io/docs/MQTT/#command-flow
func NewService(bus eventbus.Bus) (service *Service, e error) {
	opts := mqttclient.NewClientOptions()

	opts.SetClientID(clientID)

	if value, found := os.LookupEnv(envMQTTBrokerURI); found {
		opts.AddBroker(value)
	}

	if value, found := os.LookupEnv(envMQTTUserName); found {
		opts.SetUsername(value)
	}

	if value, found := os.LookupEnv(envMQTTPassword); found {
		opts.SetPassword(value)
	}

	service = &Service{
		bus: bus,
	}

	// Set connection established handler
	opts.SetOnConnectHandler(service.onConnected)

	opts.OnConnectionLost = func(client mqttclient.Client, e error) {
		log.Log.Info("Lost connection to the MQTT broker", e)
	}

	opts.SetDefaultPublishHandler(service.onMessage)

	// SetAutoReconnect sets whether the automatic reconnection logic should be used
	// when the connection is lost, even if disabled the ConnectionLostHandler is still
	// called
	opts.SetAutoReconnect(true)

	// SetConnectRetry sets whether the connect function will automatically retry the connection
	// in the event of a failure (when true the token returned by the Connect function will
	// not complete until the connection is up or it is cancelled)
	// If ConnectRetry is true then subscriptions should be requested in OnConnect handler
	// Setting this to TRUE permits messages to be published before the connection is established	opts.SetConnectRetry(true)

	// Create MQTT client
	service.client = mqttclient.NewClient(opts)

	return service, nil
}

// Run is implementation of runnable.Runnable interface
func (service *Service) Run(ctx context.Context) error {
	if e := service.bus.Subscribe(RxDiscoveryMQTT, service.traceMessageDiscovery); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(RxDiscoveryMQTT, service.traceMessageDiscovery)
	}()

	if e := service.bus.Subscribe(RxTelemetryMQTT, service.traceMessageTelemetry); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(RxTelemetryMQTT, service.traceMessageTelemetry)
	}()

	if e := service.bus.Subscribe(RxCommandsMQTT, service.traceMessageCommand); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(RxCommandsMQTT, service.traceMessageCommand)
	}()

	if e := service.bus.Subscribe(RxStatusesMQTT, service.traceMessageStatus); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(RxStatusesMQTT, service.traceMessageStatus)
	}()

	token := service.client.Connect()

	contextDone := ctx.Done()

	for {
		select {
		case <-token.Done():
			if nil != ctx.Err() {
				// Basic context cancelled cause service operation complete
				return token.Error()
			}

			// Basic context still active, attempt to reconnect
			token = service.client.Connect()
		case <-contextDone:
			// Request to disconnect
			service.client.Disconnect(waitDisconnectCompleteMS)

			// Don't use this channel again
			contextDone = nil
		}
	}
}

// onConnected called when connection established
func (service *Service) onConnected(_ mqttclient.Client) {
	service.client.Subscribe(topicDiscovery, 1, service.onMessageDiscovery)
	service.client.Subscribe(topicTelemetry, 1, service.onMessageTelemetry)
	service.client.Subscribe(topicCommands, 1, service.onMessageCommands)
	service.client.Subscribe(topicStatuses, 1, service.onMessageStatuses)
}

// onMessageDiscovery called when received discovery message
func (service *Service) onMessageDiscovery(_ mqttclient.Client, msg mqttclient.Message) {
	event := NewEventMQTT(msg.Topic(), msg.Payload())

	service.bus.Publish(RxDiscoveryMQTT, event)
}

// traceMessageDiscovery trace received discovery message
func (service *Service) traceMessageDiscovery(event EventMQTT) {
	log.Log.Debug("Rx MQTT discovery", event)
}

// onMessageTelemetry called when received telemetry message
func (service *Service) onMessageTelemetry(_ mqttclient.Client, msg mqttclient.Message) {
	event := NewEventMQTT(msg.Topic(), msg.Payload())

	service.bus.Publish(RxTelemetryMQTT, event)
}

// traceMessageTelemetry trace received telemetry message
func (service *Service) traceMessageTelemetry(event EventMQTT) {
	log.Log.Debug("Rx MQTT telemetry", event)
}

// onMessageCommands called when received command message
func (service *Service) onMessageCommands(_ mqttclient.Client, msg mqttclient.Message) {
	event := NewEventMQTT(msg.Topic(), msg.Payload())

	service.bus.Publish(RxCommandsMQTT, event)
}

// traceMessageCommand trace received command message
func (service *Service) traceMessageCommand(event EventMQTT) {
	log.Log.Debug("Rx MQTT command", event)
}

// onMessageStatuses called when received status message
func (service *Service) onMessageStatuses(_ mqttclient.Client, msg mqttclient.Message) {
	event := NewEventMQTT(msg.Topic(), msg.Payload())

	service.bus.Publish(RxStatusesMQTT, event)
}

// traceMessageStatus trace received status message
func (service *Service) traceMessageStatus(event EventMQTT) {
	log.Log.Debug("Rx MQTT status", event)
}

// onMessage called when received message from MQTT
func (service *Service) onMessage(_ mqttclient.Client, msg mqttclient.Message) {
	log.Log.Debug("Publish message", msg)
}

// NewEventMQTT return internal MQTT event representation
func NewEventMQTT(topic string, payload []byte) EventMQTT {
	return EventMQTT{
		Topic:   strings.Split(topic, topicPartsDelimiter),
		Payload: string(payload[:]),
	}
}
