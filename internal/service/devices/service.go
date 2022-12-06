package devices

import (
	"context"
	"errors"
	"sync"

	"github.com/pior/runnable"
	"github.com/vedga/alisa/pkg/api"
)

// Service is device manager service implementation
type Service struct {
	runnable.Runnable
	devices sync.Map
}

// NewService return new service implementation
func NewService() (service *Service, e error) {
	return &Service{}, nil
}

// Run is implementation of runnable.Runnable interface
func (service *Service) Run(ctx context.Context) error {
	// Wait until operation complete
	<-ctx.Done()

	return ctx.Err()
}

// AddDevice is implementation of api.DeviceManager interface
func (service *Service) AddDevice(deviceID string, device api.Device) error {
	// Store device in the internal storage
	if prev, found := service.devices.LoadOrStore(deviceID, device); found {
		// Update existing device
		if e := prev.(api.Device).Update(device); nil != e {
			// Unable to update device
			return e
		}
	}

	return nil
}

// EnumDevices is implementation of api.DeviceManager interface
func (service *Service) EnumDevices() (devices map[string]api.Device, e error) {
	service.devices.Range(func(key, value any) bool {
		var (
			keyValue    string
			deviceValue api.Device
			valid       bool
		)

		if keyValue, valid = key.(string); !valid {
			e = errors.New("invalid device id type")
			return false
		}

		if deviceValue, valid = value.(api.Device); !valid {
			e = errors.New("invalid device")
			return false
		}

		devices[keyValue] = deviceValue

		return true
	})

	return devices, e
}
