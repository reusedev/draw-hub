package response

import "github.com/gin-gonic/gin"

var (
	ParamError = gin.H{"code": 10001, "message": "param error"}

	InternalError = gin.H{"code": 10002, "message": "internal error"}
)
