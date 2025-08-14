package rest

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/glog"
)

func GetMandatoryIdParamValue(context *gin.Context, paramName string) (uuid.UUID, error) {
	paramValue := context.Param(paramName)
	return convertToId(context, paramValue, false)
}

func GetOptionalIdParamValue(context *gin.Context, paramName string) (uuid.UUID, error) {
	paramValue := context.DefaultQuery(paramName, "")
	return convertToId(context, paramValue, true)
}

func convertToId(context *gin.Context, paramValue string, optional bool) (uuid.UUID, error) {
	if paramValue == "" {
		if optional {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("please specify a valid param")
	}
	id, err := uuid.FromString(paramValue)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func GetMandatoryIntParamValue(context *gin.Context, paramName string) (int, error) {
	paramValue := context.Param(paramName)
	return convertToInt(context, paramValue, false)
}

func GetOptionalIntParamValue(context *gin.Context, paramName string) (int, error) {
	paramValue := context.DefaultQuery(paramName, "")
	return convertToInt(context, paramValue, true)
}

func convertToInt(context *gin.Context, paramValue string, optional bool) (int, error) {
	if paramValue == "" {
		if optional {
			return 0, nil
		}
		return 0, fmt.Errorf("please specify a valid param")
	}
	value, err := strconv.Atoi(paramValue)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func LogHandlerError(method string, err error, additionalContext ...string) {
	context := ""
	if len(additionalContext) > 0 {
		context = fmt.Sprintf(" - %s", additionalContext[0])
	}
	glog.Errorf("Handler error in %s: %v%s", method, err, context)
}
