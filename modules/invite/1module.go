package invite

import (
	"embed"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/model"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/register"
)

//go:embed sql
var sqlFS embed.FS

//go:embed swagger/api.yaml
var swaggerContent string

func init() {
	register.AddModule(func(ctx interface{}) register.Module {
		api := New(ctx.(*config.Context))
		return register.Module{
			Name: "invite",
			SetupAPI: func() register.APIRouter {
				return api
			},
			SQLDir:  register.NewSQLFS(sqlFS),
			Swagger: swaggerContent,
			BussDataSource: register.BussDataSource{
				GetInviteCode: func(inviteCode string) (*model.Invite, error) {
					invite, err := api.db.getInviteWithCode(inviteCode)
					if err != nil {
						return nil, err
					}
					if invite == nil || invite.Status == 0 {
						return nil, nil
					}
					return &model.Invite{
						InviteCode: invite.InviteCode,
						Uid:        invite.UID,
						Vercode:    invite.Vercode,
					}, nil
				},
			},
		}
	})
}
