package tests

import (
	"github.com/aicam/game_server/routHandlers"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedisOperations(t *testing.T) {
	req, err := http.NewRequest("GET", "localhost:4500/welcome/ali/2", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
}