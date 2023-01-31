package api

import "net/http"

// routeApiStatusReady is a simple API end point that just returns that the server is listening; it does NOT do any checks
// on connectivity or configuration
func routeApiStatusReady(w http.ResponseWriter, r *http.Request) {
	sendAPIJSONData(w, http.StatusOK, map[string]interface{}{
		"listening": "yes",
		"version":   "0.0.1",
	})
}
