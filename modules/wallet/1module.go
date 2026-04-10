package wallet

import (
	"embed"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/register"
)

//go:embed sql
var sqlFS embed.FS

func init() {
	// ====================== 注册钱包模块 ======================
	register.AddModule(func(ctx interface{}) register.Module {
		x := ctx.(*config.Context)
		api := New(x)
		// 启动过期退款定时任
		expireTask := NewExpireTask(x)
		expireTask.Start()
		return register.Module{
			Name: "wallet",
			SetupAPI: func() register.APIRouter {
				return api
			},
			SQLDir: register.NewSQLFS(sqlFS),
		}
	})

	// ====================== 注册转账模块 ======================
	register.AddModule(func(ctx interface{}) register.Module {
		return register.Module{
			Name: "transfer",
			SetupAPI: func() register.APIRouter {
				return NewTransfer(ctx.(*config.Context))
			},
		}
	})

	// ====================== 注册红包模块 ======================
	register.AddModule(func(ctx interface{}) register.Module {
		return register.Module{
			Name: "redpacket",
			SetupAPI: func() register.APIRouter {
				return NewRedPacket(ctx.(*config.Context))
			},
		}
	})

	// ====================== 注册钱包管理模块 ======================
	register.AddModule(func(ctx interface{}) register.Module {
		return register.Module{
			Name: "wallet_manager",
			SetupAPI: func() register.APIRouter {
				return NewManager(ctx.(*config.Context))
			},
		}
	})
}
