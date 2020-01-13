package helpers

import (
	"net/http"
	"time"
)

// NetClient is an http client that has a default timenout. If using directly
// http.Client, the default timeout is at 0, which can cause some issues.
var NetClient = &http.Client{
	Timeout: time.Second * 10,
}
