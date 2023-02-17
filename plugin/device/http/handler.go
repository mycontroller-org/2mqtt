package http

import (
	"io"
	"net/http"
	"time"

	types "github.com/mycontroller-org/2mqtt/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/json"
	"go.uber.org/zap"
)

type RequestData struct {
	Method          string              `json:"method"`
	RemoteAddress   string              `json:"remoteAddress"`
	Host            string              `json:"host"`
	Path            string              `json:"path"`
	Body            string              `json:"body"`
	QueryParameters map[string][]string `json:"queryParameters"`
	Headers         map[string][]string `json:"headers"`
	Timestamp       time.Time           `json:"timestamp"`
}

type deviceHandler struct {
	ID             string
	receiveMsgFunc func(rm *types.Message)
}

func (h deviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// update body
	body := ""
	if r.Method == http.MethodPut ||
		r.Method == http.MethodPost ||
		r.Method == http.MethodDelete {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			zap.L().Error("error on reading body", zap.String("adapterName", h.ID), zap.String("path", r.URL.Path), zap.Error(err))
			return
		}
		body = string(data)
	}

	// update headers
	headers := map[string][]string{}
	for k, v := range r.Header {
		headers[k] = v
	}

	// update query parameters
	queryParameters := map[string][]string{}
	for k, v := range r.URL.Query() {
		queryParameters[k] = v
	}

	requestData := RequestData{
		Method:          r.Method,
		RemoteAddress:   r.RemoteAddr,
		Host:            r.Host,
		Path:            r.URL.Path,
		Body:            body,
		QueryParameters: queryParameters,
		Headers:         headers,
		Timestamp:       time.Now(),
	}

	bytes, err := json.Marshal(&requestData)
	if err != nil {
		zap.L().Error("error on converting request data to json bytes", zap.String("adapterName", h.ID), zap.String("path", r.URL.Path), zap.Error(err))
		return
	}

	rawMsg := types.NewMessage(bytes)
	h.receiveMsgFunc(rawMsg)
}
