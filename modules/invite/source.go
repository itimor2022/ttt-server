package invite

import "go.uber.org/zap"

// InviteCoceIsExist 邀请码是否存在
func (i *Invite) InviteCoceIsExist(code string) (bool, error) {
	model, err := i.db.queryWithVercode(code)
	if err != nil {
		i.Error("查询邀请码错误", zap.Error(err))
		return false, err
	}
	if model == nil || model.Status == 0 || model.InviteCode == "" {
		return false, nil
	}
	return true, nil
}
