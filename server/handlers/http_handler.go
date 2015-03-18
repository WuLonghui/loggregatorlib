package handlers

import (
	"github.com/cloudfoundry/gosteno"
	"mime/multipart"
	"net/http"
    "sync/atomic"
)

type httpHandler struct {
	messages <-chan []byte
	logger   *gosteno.Logger
    totalMessagesSent int64
}

func NewHttpHandler(m <-chan []byte, logger *gosteno.Logger) *httpHandler {
	return &httpHandler{messages: m, logger: logger}
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("http handler: ServeHTTP entered with request %v", r)
	defer h.logger.Debugf("http handler: ServeHTTP exited")

	mp := multipart.NewWriter(rw)
	defer mp.Close()

	rw.Header().Set("Content-Type", `multipart/x-protobuf; boundary=`+mp.Boundary())

	for message := range h.messages {
        atomic.AddInt64(&h.totalMessagesSent, 1)
        partWriter, _ := mp.CreatePart(nil)
		partWriter.Write(message)
	}
}

func (h *httpHandler) GetTotalMessagesSent() int64{
    return h.totalMessagesSent
}
