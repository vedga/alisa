package api

// Device represent device object API
type Device interface {
	GetType() string
	GetFirmwareVersion() string
	Update(newDevice Device) error
}

// DeviceManager is interface for device manager
type DeviceManager interface {
	AddDevice(deviceID string, device Device) error
	EnumDevices() (map[string]Device, error)
}
