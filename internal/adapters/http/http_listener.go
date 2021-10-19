package http

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"reflect"
)

const (
	AdapterName       = "http"
	configKeywordPath = "path"
	configKeywordAddr = "addr"
)

type HttpListener struct {
	panicChannel chan string
	addr         string
	path         string
}

func HttpListenerFromConfigMap(args map[string]interface{}) (*HttpListener, error) {
	if args == nil {
		return nil, errors.New("empty args provided")
	}

	expectedTypes := map[string]reflect.Kind{
		configKeywordPath: reflect.String,
		configKeywordAddr: reflect.String,
	}
	for keyword, valueType := range expectedTypes {
		t, ok := args[keyword]
		if !ok {
			return nil, fmt.Errorf("no '%s' in args", keyword)
		}

		switch v := reflect.ValueOf(t); v.Kind() {
		case valueType:
		default:
			return nil, fmt.Errorf("expected type %v for keyword '%s'", valueType, keyword)
		}
	}

	return NewHttpListener(args[configKeywordPath].(string), args[configKeywordAddr].(string))
}

func NewHttpListener(path, addr string) (*HttpListener, error) {
	if len(path) == 0 {
		return nil, errors.New("empty path given")
	}

	if len(addr) == 0 {
		return nil, errors.New("no address to listen on given")
	}

	return &HttpListener{addr: addr, path: path}, nil
}

func (listener *HttpListener) handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error().Msgf("error reading body: %v", err)
	}

	listener.panicChannel <- string(body)
}

func (listener *HttpListener) Name() string {
	return AdapterName
}

func (listener *HttpListener) Start(panicChannel chan string) error {
	listener.panicChannel = panicChannel

	http.HandleFunc(listener.path, listener.handler)

	var err error
	go func() {
		err = http.ListenAndServe(listener.addr, nil)
	}()

	return err
}

func (listener *HttpListener) Stop() error {
	return nil
}
