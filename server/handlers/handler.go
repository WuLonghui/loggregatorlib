package handlers

import(
    "net/http"
)

type WriterHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
    GetTotalMessagesSent() int64
}