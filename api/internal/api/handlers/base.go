package handlers

import (
	"errors"
	"reflect"

	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

type HttpContext struct {
	*gin.Context
}

// Data related functions
func BindModel[T any](ctx *gin.Context) *T {
	var model T
	if err := ctx.BindJSON(&model); err != nil {
		BadRequest(ctx, err.Error())
		return nil
	}

	return &model
}

func MapTo[TDestination any](source any) (*TDestination, error) {
	var dest = new(TDestination)
	sourceVal := reflect.ValueOf(source)

	if sourceVal.Kind() != reflect.Struct {
		return nil, errors.New("source is not a struct")
	}

	destVal := reflect.ValueOf(dest).Elem()
	for i := 0; i < sourceVal.NumField(); i++ {
		sourceField := sourceVal.Field(i)
		sourceFieldName := sourceVal.Type().Field(i).Name

		destField := destVal.FieldByName(sourceFieldName)
		if destField.IsValid() && destField.CanSet() {
			destField.Set(sourceField)
		}
	}

	return dest, nil
}

// Return Types for Controllers
func Ok(ctx *gin.Context, data any) {
	ctx.JSON(200, models.ApiResponse[any]{
		Success: true,
		Status:  200,
		Data:    data,
	})
}

func NotFound(ctx *gin.Context, message string) {
	ctx.JSON(200, models.ApiResponse[any]{
		Success: false,
		Status:  404,
		Message: message,
	})
}

func BadRequest(ctx *gin.Context, message string) {
	ctx.JSON(200, models.ApiResponse[any]{
		Success: false,
		Status:  400,
		Message: message,
	})
}

func InternalServerError(ctx *gin.Context, message string) {
	ctx.JSON(200, models.ApiResponse[any]{
		Success: false,
		Status:  500,
		Message: message,
	})
}

func Unauthorized(ctx *gin.Context, message string) {
	ctx.JSON(200, models.ApiResponse[any]{
		Success: false,
		Status:  401,
		Message: message,
	})
}
