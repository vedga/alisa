package alisa

import "github.com/gin-gonic/gin"

// DeviceInfo represent device information
type DeviceInfo struct {
	Manufacturer    string `json:"manufacturer,omitempty"`
	Model           string `json:"model,omitempty"`
	HardwareVersion string `json:"hw_version,omitempty"`
	SoftwareVersion string `json:"sw_version,omitempty"`
}

// Device represent device
type Device struct {
	ID           string        `json:"id,omitempty"`
	Name         string        `json:"name,omitempty"`
	Description  string        `json:"description,omitempty"`
	Room         string        `json:"room,omitempty"`
	Type         string        `json:"type,omitempty"`
	CustomData   string        `json:"custom_data,omitempty"`
	Capabilities []interface{} `json:"capabilities,omitempty"`
	Properties   []interface{} `json:"properties,omitempty"`
	DeviceInfo   DeviceInfo    `json:"device_info,omitempty"`
}

type devicesPayload struct {
	UserID  string   `json:"user_id,omitempty"`
	Devices []Device `json:"devices,omitempty"`
}

// devicesResponse is response for "/v1.0/user/devices" request
type devicesResponse struct {
	RequestID string         `json:"request_id,omitempty"`
	Payload   devicesPayload `json:"payload,omitempty"`
}

func newDevicesResponse(ginCtx *gin.Context) *devicesResponse {
	return &devicesResponse{
		RequestID: ginCtx.GetHeader(headerRequestID),
		Payload: devicesPayload{
			UserID: ginCtx.GetString(contextUserID),
		},
	}
}
