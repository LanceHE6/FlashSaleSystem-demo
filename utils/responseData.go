package utils

import "github.com/gin-gonic/gin"

// Response 构造返回信息
func Response(message string, data gin.H, code int) gin.H {
	return gin.H{
		"msg":  message,
		"data": data,
		"code": code,
	}
}

// ErrorResponse 构造错误返回信息
func ErrorResponse(message string, error string, code int) gin.H {
	return Response(message, gin.H{"error": error}, code)
}
