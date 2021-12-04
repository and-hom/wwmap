package tile_data_fetcher

import (
	"fmt"
	"net/http"
)

type DataFetchError struct {
	httpStatus int
	cause      error
	message    string
}

func (this *DataFetchError) Error() string {
	return this.message
}

func (this *DataFetchError) Cause() error {
	return this.cause
}

func (this *DataFetchError) HttpStatus() int {
	return this.httpStatus
}

func InternalServerError(cause error, format string, a ...interface{}) *DataFetchError {
	return Err(http.StatusInternalServerError, cause, format, a)
}

func NotFound(format string, a ...interface{}) *DataFetchError {
	return Err(http.StatusInternalServerError, nil, format, a)
}

func Err(httpStatus int, cause error, format string, a []interface{}) *DataFetchError {
	return &DataFetchError{
		httpStatus: httpStatus,
		cause:      cause,
		message:    fmt.Sprintf(format, a...),
	}
}
