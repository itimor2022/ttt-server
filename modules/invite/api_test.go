package invite

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/common"
	_ "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/user"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	_ "github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkevent"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGetInvite(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	i := New(ctx)
	// i.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	err = i.db.insertInvite(&InviteModel{
		UID:          testutil.UID,
		Status:       1,
		InviteCode:   "13",
		BeInviteUID:  "",
		BeInviteCode: "",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/v1/invite", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"invite_code":"13"`))
}

func TestReset(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	i := New(ctx)
	// i.Route(s.GetRoute())
	//清除数据
	// err := testutil.CleanAllTables(ctx)
	// assert.NoError(t, err)
	err := i.db.insertInvite(&InviteModel{
		UID:          testutil.UID,
		Status:       1,
		InviteCode:   "13",
		BeInviteUID:  "",
		BeInviteCode: "",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("PUT", "/v1/invite/reset", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateStatus(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	i := New(ctx)
	// i.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	err = i.db.insertInvite(&InviteModel{
		UID:          testutil.UID,
		Status:       0,
		InviteCode:   "13",
		BeInviteUID:  "",
		BeInviteCode: "",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("PUT", "/v1/invite/status", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister(t *testing.T) {
	s, _ := testutil.NewTestServer()
	// i := New(ctx)
	// i.Route(s.GetRoute())
	//清除数据
	// err := testutil.CleanAllTables(ctx)
	// assert.NoError(t, err)
	// err := i.db.insertInvite(&InviteModel{
	// 	UID:          testutil.UID,
	// 	Status:       0,
	// 	InviteCode:   "13",
	// 	BeInviteUID:  "",
	// 	BeInviteCode: "",
	// })
	// assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/user/register", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"code":        "123456",
		"zone":        "0086",
		"phone":       "13600000008",
		"password":    "1234567",
		"invite_code": "9204165",
		"device": map[string]interface{}{
			"device_id":    "device_id2",
			"device_name":  "device_name2",
			"device_model": "device_model2",
		},
	}))))

	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"token":`))
}
