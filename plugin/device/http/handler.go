package http

import (
	"fmt"
	"io/ioutil"
	"net/http"

	types "github.com/mycontroller-org/2mqtt/pkg/types"
	"go.uber.org/zap"
)

type deviceHandler struct {
	ID             string
	receiveMsgFunc func(rm *types.Message)
}

func (h deviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	address := fmt.Sprintf("%s%s", r.Host, r.URL.Path)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		zap.L().Error("error on reading body", zap.String("adapterName", h.ID), zap.String("url", address), zap.Error(err))
		return
	}
	rawMsg := types.NewMessage(data)

	// update other fields
	rawMsg.Others.Set(types.KeyURL, address, nil)
	// get headers
	headers := map[string][]string{}
	for k, v := range r.Header {
		headers[k] = v
	}
	rawMsg.Others.Set(types.KeyHeaders, headers, nil)
	h.receiveMsgFunc(rawMsg)
}
