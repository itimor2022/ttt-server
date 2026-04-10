package rtc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRoomTokenGet(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/rtc/rooms/608b6f66942cf200d028069f/token", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{}))))
	req.Header.Set("token", testutil.Token)
	assert.NoError(t, err)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoomCreate(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/rtc/rooms", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"name": "测试",
		"uids": []string{"dsd", "zzz"},
	}))))
	req.Header.Set("token", testutil.Token)
	assert.NoError(t, err)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
