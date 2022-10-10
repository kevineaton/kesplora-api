package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIStatusCallRoute(t *testing.T) {
	SetupConfig()
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})

	code, _, err := testEndpoint(http.MethodGet, "/", b, routeApiStatusReady, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
}
