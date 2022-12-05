package mqtt

import (
	"context"
	"os"

	mqttclient "github.com/eclipse/paho.mqtt.golang"
	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
)

const (
	// envMQTTBrokerURI is URI for connect to the MQTT server in form "tcp://host:port"
	envMQTTBrokerURI         = "MQTT_BROKER_URI"
	envMQTTUserName          = "MQTT_USER_NAME"
	envMQTTPassword          = "MQTT_PASSWORD"
	clientID                 = "alisa_service"
	waitDisconnectCompleteMS = 1000
	topicDiscoveryTasmota    = "tasmota/discovery/#"
	topicTelemetry           = "tele/#"
	topicCommands            = "cmnd/#"
)

// Service is MQTT client service implementation
type Service struct {
	runnable.Runnable
	client mqttclient.Client
}

// NewService return new service implementation
// example: https://levelup.gitconnected.com/how-to-use-mqtt-with-go-89c617915774
// Official documentation: https://www.emqx.com/en/blog/how-to-use-mqtt-in-golang
// Tasmota MQTT: https://tasmota.github.io/docs/MQTT/#command-flow
func NewService() (service *Service, e error) {
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

	service = &Service{}

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
func (service *Service) onConnected(client mqttclient.Client) {
	service.client.Subscribe(topicDiscoveryTasmota, 1, service.onMessageDiscovery)
	service.client.Subscribe(topicTelemetry, 1, service.onMessageTelemetry)
	service.client.Subscribe(topicCommands, 1, service.onMessageCommand)
}

// onMessageDiscovery called when received discovery message
func (service *Service) onMessageDiscovery(client mqttclient.Client, msg mqttclient.Message) {
	log.Log.Debug("Discovery message", msg)
}

// onMessageTelemetry called when received telemetry message
func (service *Service) onMessageTelemetry(client mqttclient.Client, msg mqttclient.Message) {
	log.Log.Debug("Telemetry message", msg)
}

// onMessageCommand called when received command message
func (service *Service) onMessageCommand(client mqttclient.Client, msg mqttclient.Message) {
	log.Log.Debug("Command message", msg)
}

// onMessage called when received message from MQTT
func (service *Service) onMessage(client mqttclient.Client, msg mqttclient.Message) {
	log.Log.Debug("Publish message", msg)
}
