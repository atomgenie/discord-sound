package requests

import (
	"sync"
)

type request struct {
	Server *Instance
}

var requestMap map[string]request = make(map[string]request)

var requestMux sync.Mutex
