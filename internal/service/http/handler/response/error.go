package response

import "github.com/gin-gonic/gin"

var (
	ParamError            = gin.H{"code": 10001, "message": "param error"}
	ParamErrorWithMessage = func(message string) gin.H {
		return gin.H{"code": 10001, "message": message}
	}

	InternalError = gin.H{"code": 10002, "message": "internal error"}

	SuccessWithData = func(data interface{}) gin.H {
		return gin.H{"code": 0, "data": data}
	}
)
