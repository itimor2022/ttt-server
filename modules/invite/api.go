package invite

import (
	"errors"
	"fmt"

	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/base/event"
	common2 "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/source"
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

type Invite struct {
	ctx *config.Context
	log.Log
	db            *db
	commonService common2.IService
}

// New New
func New(ctx *config.Context) *Invite {
	invite := &Invite{
		ctx:           ctx,
		Log:           log.NewTLog("Invite"),
		db:            newDB(ctx),
		commonService: common2.NewService(ctx),
	}
	source.SetInviteCodeProvide(invite)
	return invite
}

// Route 路由配置
func (i *Invite) Route(r *wkhttp.WKHttp) {
	invite := r.Group("/v1/invite", i.ctx.AuthMiddleware(r))
	{
		invite.GET("", i.getInvite)           // 获取邀请码
		invite.PUT("/reset", i.reset)         // 重置邀请码
		invite.PUT("/status", i.updateStatus) // 禁用或启用邀请码
	}
	i.ctx.AddEventListener(event.EventUserRegister, i.handleRegisterUserEvent)
}

func (i *Invite) updateStatus(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	model, err := i.db.getInviteWithUID(loginUID)
	if err != nil {
		i.Error("查询用户邀请码错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户邀请码错误"))
		return
	}
	if model == nil {
		c.ResponseError(errors.New("用户未生成邀请码"))
		return
	}
	if model.Status == 1 {
		model.Status = 0
	} else {
		model.Status = 1
	}
	err = i.db.updateInvite(model)
	if err != nil {
		i.Error("禁用用户邀请码错误", zap.Error(err))
		c.ResponseError(errors.New("禁用用户邀请码错误"))
		return
	}
	c.ResponseOK()
}

func (i *Invite) reset(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	code, err := i.commonService.GetShortno()
	if err != nil {
		i.Error("生成邀请码错误", zap.Error(err))
		c.ResponseError(errors.New("生成邀请码错误"))
		return
	}
	if code == "" {
		c.ResponseError(errors.New("生成邀请码失败"))
		return
	}
	model, err := i.db.getInviteWithUID(loginUID)
	if err != nil {
		i.Error("查询用户邀请码错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户邀请码错误"))
		return
	}
	if model == nil {
		err := i.db.insertInvite(&InviteModel{
			InviteCode:   code,
			UID:          loginUID,
			Status:       1,
			BeInviteUID:  "",
			BeInviteCode: "",
			Vercode:      fmt.Sprintf("%s@%d", util.GenerUUID(), common.InvitationCode),
		})
		if err != nil {
			i.Error("新增用户邀请码错误", zap.Error(err))
			c.ResponseError(errors.New("新增用户邀请码错误"))
			return
		}
	} else {
		model.InviteCode = code
		err := i.db.updateInvite(model)
		if err != nil {
			i.Error("修改用户邀请码错误", zap.Error(err))
			c.ResponseError(errors.New("修改用户邀请码错误"))
			return
		}
	}
	c.ResponseOK()
}

func (i *Invite) getInvite(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	model, err := i.db.getInviteWithUID(loginUID)
	if err != nil {
		i.Error("查询用户邀请码错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户邀请码错误"))
		return
	}
	code := ""
	status := 0
	if model != nil {
		code = model.InviteCode
		status = model.Status
	} else {
		code, err = i.commonService.GetShortno()
		if err != nil {
			i.Error("生成用户邀请码错误", zap.Error(err))
			c.ResponseError(errors.New("生成用户邀请码错误"))
			return
		}
		err = i.db.insertInvite(&InviteModel{
			InviteCode:   code,
			UID:          loginUID,
			Status:       1,
			BeInviteUID:  "",
			BeInviteCode: "",
			Vercode:      fmt.Sprintf("%s@%d", util.GenerUUID(), common.InvitationCode),
		})
		if err != nil {
			i.Error("保存用户邀请码错误", zap.Error(err))
			c.ResponseError(errors.New("保存用户邀请码错误"))
			return
		}
		status = 1
	}
	c.Response(map[string]interface{}{
		"invite_code": code,
		"status":      status,
	})
}
