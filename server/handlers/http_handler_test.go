package handlers_test

import (
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/loggregatorlib/server/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
)

var _ = Describe("HttpHandler", func() {
	var boundaryRegexp = regexp.MustCompile("boundary=(.*)")
	var handler handlers.WriterHandler
	var fakeResponseWriter *httptest.ResponseRecorder
	var messagesChan chan []byte

	BeforeEach(func() {
		fakeResponseWriter = httptest.NewRecorder()
		messagesChan = make(chan []byte, 10)
		handler = handlers.NewHttpHandler(messagesChan, loggertesthelper.Logger())
	})

    It("keeps track of the total number of messages sent", func() {
        r, _ := http.NewRequest("GET", "ws://loggregator.place/dump/?app=abc-123", nil)
        for i := 0; i < 5; i++ {
            messagesChan <- []byte("message")
        }

        close(messagesChan)
        handler.ServeHTTP(fakeResponseWriter, r)
        totalNumberOfMessages := handler.GetTotalMessagesSent()
        Expect(totalNumberOfMessages).To(Equal(int64(5)))
    })

	It("grabs recent logs and creates a multi-part HTTP response", func(done Done) {
		r, _ := http.NewRequest("GET", "ws://loggregator.place/dump/?app=abc-123", nil)

		for i := 0; i < 5; i++ {
			messagesChan <- []byte("message")
		}

		close(messagesChan)
		handler.ServeHTTP(fakeResponseWriter, r)

		matches := boundaryRegexp.FindStringSubmatch(fakeResponseWriter.Header().Get("Content-Type"))
		Expect(matches).To(HaveLen(2))
		Expect(matches[1]).NotTo(BeEmpty())
		reader := multipart.NewReader(fakeResponseWriter.Body, matches[1])
		partsCount := 0
		var err error
		for err != io.EOF {
			var part *multipart.Part
			part, err = reader.NextPart()
			if err == io.EOF {
				break
			}
			partsCount++

			data := make([]byte, 1024)
			n, _ := part.Read(data)
			Expect(data[0:n]).To(Equal([]byte("message")))
		}

		Expect(partsCount).To(Equal(5))
		close(done)
	})

	It("sets the MIME type correctly", func() {
		close(messagesChan)
		r, err := http.NewRequest("GET", "", nil)
		Expect(err).ToNot(HaveOccurred())

		handler.ServeHTTP(fakeResponseWriter, r)
		Expect(fakeResponseWriter.Header().Get("Content-Type")).To(MatchRegexp(`multipart/x-protobuf; boundary=`))
	})
})
