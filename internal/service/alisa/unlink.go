package alisa

import "github.com/gin-gonic/gin"

type unlinkResponse struct {
	RequestID string `json:"request_id,omitempty"`
}

func newUnlinkResponse(ginCtx *gin.Context) *unlinkResponse {
	return &unlinkResponse{
		RequestID: ginCtx.GetHeader(headerRequestID),
	}
}
