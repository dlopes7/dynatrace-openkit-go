package communication

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestBuildMonitorURL(t *testing.T) {

	assert.Equal(t, "https://localhost:9999/mbeacon/e/eaa50379?type=m&srvid=1&app=98972aef-02ac-4ecb-be1e-a6698af2de60&va=8.217.20300&pt=1&tt=okgo&resp=json", httpClient.monitorURL)
	assert.Equal(t, 1, httpClient.serverID)

}

func TestSendStatusRequest(t *testing.T) {

	statusResponse := httpClient.SendStatusRequest(ctx)
	assert.True(t, statusResponse.ResponseAttributes.Capture)
	assert.True(t, statusResponse.ResponseCode == http.StatusOK)

}
