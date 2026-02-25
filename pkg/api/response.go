package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
	StatusAlreadyExists = "AlreadyExists"
	StatusFileExtNotAllowed = "FileExtNotAllowed"
	StatusBadFileSize = "BadFileSize"
	StatusNotFound = "NotFound"
	StatusBadRequest = "BadRequest"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(err string) Response {
	return Response{
		Status: StatusError,
		Error:  err,
	}
}

func ErrorWithStatus(status string, err string) Response {
	return Response{
		Status: status,
		Error:  err,
	}
}


func ValidateRequest(req interface{}) error {
	validate := validator.New()
	return validate.Struct(req)
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {	
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("Field %s is required", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("Field %s is not a valid URL", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("Field %s is not a valid email", err.Field()))
		case "eqfield":
			errMsgs = append(errMsgs, fmt.Sprintf("Field %s is not a equal %s field", err.Field(), err.Param()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("Field %s is invalid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  fmt.Sprintf("Validation error: %s", strings.Join(errMsgs, ", ")),
	}

}
