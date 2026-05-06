package invite

import (
	"errors"
	"fmt"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"go.uber.org/zap"
)

func (i *Invite) handleRegisterUserEvent(data []byte, commit config.EventCommit) {
	var req map[string]interface{}
	err := util.ReadJsonByByte(data, &req)
	if err != nil {
		i.Error("邀请码处理用户注册加入群聊参数有误")
		commit(err)
		return
	}
	if req == nil || req["invite_vercode"] == nil {
		commit(nil)
		return
	}
	uid := req["uid"].(string)
	if uid == "" {
		i.Error("邀请码处理用户注册uid不能为空")
		commit(errors.New("邀请码处理用户注册uid不能为空"))
		return
	}
	inviteCode := req["invite_code"].(string)
	beInviteUid := ""
	if inviteCode != "" {
		model, err := i.db.getInviteWithCode(inviteCode)
		if err != nil {
			i.Error("处理用户注册查询邀请码错误")
			commit(errors.New("处理用户注册查询邀请码错误"))
			return
		}
		if model != nil {
			beInviteUid = model.UID
		}
	}
	code, err := i.commonService.GetShortno()
	if err != nil {
		i.Error("注册用户生成邀请码错误", zap.Error(err))
		commit(errors.New("注册用户生成邀请码错误"))
		return
	}
	err = i.db.insertInvite(&InviteModel{
		UID:          uid,
		InviteCode:   code,
		BeInviteUID:  beInviteUid,
		BeInviteCode: inviteCode,
		Status:       1,
		Vercode:      fmt.Sprintf("%s@%d", util.GenerUUID(), common.InvitationCode),
	})
	if err != nil {
		i.Error("创建注册用户邀请码错误", zap.Error(err))
		commit(errors.New("创建注册用户邀请码错误"))
		return
	}
	commit(nil)
}
