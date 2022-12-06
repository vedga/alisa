package tasmota

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/pior/runnable"
	"github.com/vedga/alisa/internal/pkg/log"
	"github.com/vedga/alisa/internal/service/mqtt"
	"github.com/vedga/alisa/pkg/api"
	"github.com/vedga/alisa/pkg/eventbus"
)

// Service is Tasmota MQTT encoder service implementation
type Service struct {
	runnable.Runnable
	bus           eventbus.Bus
	deviceManager api.DeviceManager
}

// NewService return new service implementation
func NewService(bus eventbus.Bus, deviceManager api.DeviceManager) (service *Service, e error) {
	return &Service{
		bus:           bus,
		deviceManager: deviceManager,
	}, nil
}

// Run is implementation of runnable.Runnable interface
func (service *Service) Run(ctx context.Context) error {
	if e := service.bus.Subscribe(mqtt.RxDiscoveryMQTT, service.rxMessageDiscovery); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(mqtt.RxDiscoveryMQTT, service.rxMessageDiscovery)
	}()

	if e := service.bus.Subscribe(mqtt.RxTelemetryMQTT, service.rxMessageTelemetry); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(mqtt.RxTelemetryMQTT, service.rxMessageTelemetry)
	}()

	if e := service.bus.Subscribe(mqtt.RxCommandsMQTT, service.rxMessageCommand); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(mqtt.RxCommandsMQTT, service.rxMessageCommand)
	}()

	if e := service.bus.Subscribe(mqtt.RxStatusesMQTT, service.rxMessageStatus); nil != e {
		return e
	}
	defer func() {
		_ = service.bus.Unsubscribe(mqtt.RxStatusesMQTT, service.rxMessageStatus)
	}()

	// Wait until operation complete
	<-ctx.Done()

	return ctx.Err()
}

const (
	// mqttDiscoveryIndexPayloadType is index in MQTT topic name part with payload type
	mqttDiscoveryIndexPayloadType = -1
	// mqttDiscoveryIndexMAC is index in MQTT topic name part with MAC address
	mqttDiscoveryIndexMAC = -2
	// tasmotaPayloadConfig is Tasmota-specific config payload
	tasmotaPayloadConfig = "config"
	// tasmotaPayloadSensors is Tasmota-specific sensors payload
	tasmotaPayloadSensors = "sensors"
)

// rxMessageDiscovery called when received discovery message
//
// Example:
// discovery{[tasmota discovery D8F15BB3DB2D config]
// { "ip":"192.168.75.224", "dn":"Tasmota", "fn":["Tasmota","Tasmota2",null,null,null,null,null,null],
// "hn":"tasmota-B3DB2D-6957",
// "mac":"D8F15BB3DB2D","md":"Sonoff T1 2CH","ty":0,"if":0,"ofln":"Offline","onln":"Online",
// "state":["OFF","ON","TOGGLE","HOLD"], "sw":"10.0.0", "t":"tasmota_B3DB2D", "ft":"%prefix%/%topic%/",
// "tp":["cmnd","stat","tele"],"rl":[1,1,0,0,0,0,0,0],"swc":[-1,-1,-1,-1,-1,-1,-1,-1],
// "swn":[null,null,null,null,null,null,null,null],
// "btn":[0,0,0,0,0,0,0,0],
// "so":{"4":0,"11":0,"13":0,"17":0,"20":0,"30":0,"68":0,"73":0,"82":0,"114":0,"117":0},
// "lk":0,"lt_st":0,"sho":[0,0,0,0],"ver":1}
// }
//
// discovery{[tasmota discovery D8F15BB3DB2D sensors]
// {"sn":{"Time":"2022-12-06T18:51:26"},"ver":1}}
func (service *Service) rxMessageDiscovery(event mqtt.EventMQTT) {
	topicParts := len(event.Topic)
	if topicParts < 2 {
		log.Log.Warn("Discovery message don't implemented yet", event)
		return
	}

	// Device MAC address for discovery message
	mac := event.Topic[topicParts+mqttDiscoveryIndexMAC]

	// Payload content
	switch event.Topic[topicParts+mqttDiscoveryIndexPayloadType] {
	case tasmotaPayloadConfig:
		var payload device
		if e := json.Unmarshal(event.Payload, &payload); nil != e {
			log.Log.Error("Discovery message don't implemented yet", event, e)
			return
		}

		if mac != payload.MAC {
			log.Log.Error("Invalid topic name for discovery message", event)
			return
		}

		if e := service.deviceManager.AddDevice(deviceID(payload.MAC), &payload); nil != e {
			log.Log.Error("Invalid topic name for discovery message", event)
			return
		}
	case tasmotaPayloadSensors:
		var payload struct {
			Version int `json:"ver,omitempty"`
		}
		if e := json.Unmarshal(event.Payload, &payload); nil != e {
			log.Log.Error("Discovery message don't implemented yet", event, e)
		}
	default:
		log.Log.Warn("Decoding Discovery message don't implemented yet", event)
		return
	}
}

// rxMessageDiscovery called when received telemetry message
func (service *Service) rxMessageTelemetry(event mqtt.EventMQTT) {
}

// rxMessageDiscovery called when received command message
func (service *Service) rxMessageCommand(event mqtt.EventMQTT) {
}

// rxMessageDiscovery called when received status message
func (service *Service) rxMessageStatus(event mqtt.EventMQTT) {
}

const (
	devicePrefix = "tasmota_"
)

// deviceID return deviceID by internal hardware ID
func deviceID(hardwareID string) string {
	return devicePrefix + hardwareID
}

// device is object which implement api.Device interface
type device struct {
	IP                    string   `json:"ip,omitempty"`
	DN                    string   `json:"dn,omitempty"`
	HardwareCompatibility []string `json:"fn,omitempty"`
	HardwareID            string   `json:"hn,omitempty"`
	MAC                   string   `json:"mac,omitempty"`
	Type                  string   `json:"md,omitempty"`
	SupportedStates       []string `json:"state,omitempty"`
	FirmwareVersion       string   `json:"sw,omitempty"`
	TopicID               string   `json:"t,omitempty"`
}

// GetType is implementation of api.Device interface
func (d *device) GetType() string {
	return d.Type
}

// GetFirmwareVersion is implementation of api.Device interface
func (d *device) GetFirmwareVersion() string {
	return d.FirmwareVersion
}

// Update is implementation of api.Device interface
func (d *device) Update(newDevice api.Device) error {
	if source, valid := newDevice.(*device); valid {
		d.IP = source.IP
		d.DN = source.DN
		d.HardwareCompatibility = source.HardwareCompatibility
		d.HardwareID = source.HardwareID
		d.MAC = source.MAC
		d.Type = source.Type
		d.SupportedStates = source.SupportedStates
		d.FirmwareVersion = source.FirmwareVersion
		d.TopicID = source.TopicID
	}

	return errors.New("invalid device")
}
