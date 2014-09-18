package handlers

import (
	"mime/multipart"
	"net/http"
	"github.com/cloudfoundry/gosteno"
)

type httpHandler struct {
	messages <-chan []byte
	logger *gosteno.Logger
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
		partWriter, _ := mp.CreatePart(nil)
		partWriter.Write(message)
	}
}
