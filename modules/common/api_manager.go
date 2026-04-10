package common

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Manager 通用后台管理api
type Manager struct {
	ctx *config.Context
	log.Log
	db          *db
	appconfigDB *appConfigDB
}

// NewManager NewManager
func NewManager(ctx *config.Context) *Manager {
	return &Manager{
		ctx:         ctx,
		Log:         log.NewTLog("commonManager"),
		db:          newDB(ctx.DB()),
		appconfigDB: newAppConfigDB(ctx),
	}
}

// Route 配置路由规则
func (m *Manager) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/manager", m.ctx.AuthMiddleware(r))
	{
		auth.GET("/common/appconfig", m.appconfig)               // 获取app配置
		auth.POST("/common/appconfig", m.updateConfig)           // 修改app配置
		auth.GET("/common/appmodule", m.getAppModule)            // 获取app模块
		auth.PUT("/common/appmodule", m.updateAppModule)         // 修改app模块
		auth.POST("/common/appmodule", m.addAppModule)           // 新增app模块
		auth.DELETE("/common/:sid/appmodule", m.deleteAppModule) // 删除app模块
	}
}
func (m *Manager) deleteAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	sid := c.Param("sid")
	if strings.TrimSpace(sid) == "" {
		c.ResponseError(errors.New("sid不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(sid)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module == nil {
		c.ResponseError(errors.New("删除的模块不存在"))
		return
	}
	err = m.db.deleteAppModule(sid)
	if err != nil {
		m.Error("删除app模块错误", zap.Error(err))
		c.ResponseError(errors.New("删除app模块错误"))
		return
	}
	c.ResponseOK()
}

// 新增app模块
func (m *Manager) addAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type ReqVO struct {
		SID    string `json:"sid"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Status int    `json:"status"`
	}
	var req ReqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if strings.TrimSpace(req.SID) == "" || strings.TrimSpace(req.Desc) == "" || strings.TrimSpace(req.Name) == "" {
		c.ResponseError(errors.New("名称/ID/介绍不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(req.SID)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module != nil && module.SID != "" {
		c.ResponseError(errors.New("该sid模块已存在"))
		return
	}
	_, err = m.db.insertAppModule(&appModuleModel{
		SID:    req.SID,
		Name:   req.Name,
		Desc:   req.Desc,
		Status: req.Status,
	})
	if err != nil {
		m.Error("新增app模块错误", zap.Error(err))
		c.ResponseError(errors.New("新增app模块错误"))
		return
	}
	c.ResponseOK()
}
func (m *Manager) updateAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type ReqVO struct {
		SID    string `json:"sid"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Status int    `json:"status"`
	}
	var req ReqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if strings.TrimSpace(req.SID) == "" || strings.TrimSpace(req.Desc) == "" || strings.TrimSpace(req.Name) == "" {
		c.ResponseError(errors.New("名称/ID/介绍不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(req.SID)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module == nil {
		c.ResponseError(errors.New("不存在该模块"))
		return
	}
	module.Name = req.Name
	module.Desc = req.Desc
	module.Status = req.Status
	err = m.db.updateAppModule(module)
	if err != nil {
		m.Error("修改app模块错误", zap.Error(err))
		c.ResponseError(errors.New("修改app模块错误"))
		return
	}
	c.ResponseOK()
}

// 获取app模块
func (m *Manager) getAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	modules, err := m.db.queryAppModule()
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	list := make([]*managerAppModule, 0)
	if len(modules) > 0 {
		for _, module := range modules {
			list = append(list, &managerAppModule{
				SID:    module.SID,
				Name:   module.Name,
				Desc:   module.Desc,
				Status: module.Status,
			})
		}
	}
	c.Response(list)
}
func (m *Manager) updateConfig(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type reqVO struct {
		RevokeSecond                   int    `json:"revoke_second"`
		WelcomeMessage                 string `json:"welcome_message"`
		NewUserJoinSystemGroup         int    `json:"new_user_join_system_group"`
		SearchByPhone                  int    `json:"search_by_phone"`
		RegisterInviteOn               int    `json:"register_invite_on"`                  // 开启注册邀请机制
		SendWelcomeMessageOn           int    `json:"send_welcome_message_on"`             // 开启注册登录发送欢迎语
		InviteSystemAccountJoinGroupOn int    `json:"invite_system_account_join_group_on"` // 开启系统账号加入群聊
		RegisterUserMustCompleteInfoOn int    `json:"register_user_must_complete_info_on"` // 注册用户必须填写完整信息
		ChannelPinnedMessageMaxCount   int    `json:"channel_pinned_message_max_count"`    // 频道置顶消息最大数量
		CanModifyApiUrl                int    `json:"can_modify_api_url"`                  // 是否可以修改api地址
		MomentsOff                     int    `json:"moments_off"`                         // 是否隐藏朋友圈入口
		NormalUserCanAddFriend         int    `json:"normal_user_can_add_friend"`          // 普通用户是否可以添加好友
		AIInvestUrl                    string `json:"ai_invest_url"`                       // AI智投链接
		InsightUrl                     string `json:"insight_url"`                         // 洞察链接
		AIInvestOn                     int    `json:"ai_invest_on"`                        // AI智投开关
		InsightConnectionOn            int    `json:"insight_connection_on"`               // 洞察连接开关
	}
	var req reqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	// 查询应用配置
	appConfigM, err := m.appconfigDB.query()
	if err != nil {
		m.Error("查询应用配置失败！", zap.Error(err))
		c.ResponseError(errors.New("查询应用配置失败！"))
		return
	}

	// 如果配置不存在，创建默认配置
	if appConfigM == nil {
		// 创建默认配置
		appConfigM = &appConfigModel{
			Version:                        1,
			SuperToken:                     util.GenerUUID(),
			SuperTokenOn:                   0,
			RevokeSecond:                   req.RevokeSecond,
			WelcomeMessage:                 req.WelcomeMessage,
			NewUserJoinSystemGroup:         req.NewUserJoinSystemGroup,
			SearchByPhone:                  req.SearchByPhone,
			RegisterInviteOn:               req.RegisterInviteOn,
			SendWelcomeMessageOn:           req.SendWelcomeMessageOn,
			InviteSystemAccountJoinGroupOn: req.InviteSystemAccountJoinGroupOn,
			RegisterUserMustCompleteInfoOn: req.RegisterUserMustCompleteInfoOn,
			ChannelPinnedMessageMaxCount:   req.ChannelPinnedMessageMaxCount,
			CanModifyApiUrl:                req.CanModifyApiUrl,
			MomentsOff:                     req.MomentsOff,
			NormalUserCanAddFriend:         req.NormalUserCanAddFriend,
			AIInvestUrl:                    req.AIInvestUrl,
			InsightUrl:                     req.InsightUrl,
			AIInvestOn:                     req.AIInvestOn,
			InsightConnectionOn:            req.InsightConnectionOn,
		}

		// 生成RSA密钥
		privateKeyBuff := new(bytes.Buffer)
		publicKeyBuff := new(bytes.Buffer)
		bits := 2048
		privateKey, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			m.Error("生成RSA密钥失败！", zap.Error(err))
			c.ResponseError(errors.New("生成RSA密钥失败！"))
			return
		}
		derStream := x509.MarshalPKCS1PrivateKey(privateKey)
		block := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: derStream,
		}
		err = pem.Encode(privateKeyBuff, block)
		if err != nil {
			m.Error("编码RSA私钥失败！", zap.Error(err))
			c.ResponseError(errors.New("编码RSA私钥失败！"))
			return
		}
		publicKey := &privateKey.PublicKey
		derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
		if err != nil {
			m.Error("编码RSA公钥失败！", zap.Error(err))
			c.ResponseError(errors.New("编码RSA公钥失败！"))
			return
		}
		block = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: derPkix,
		}
		err = pem.Encode(publicKeyBuff, block)
		if err != nil {
			m.Error("编码RSA公钥失败！", zap.Error(err))
			c.ResponseError(errors.New("编码RSA公钥失败！"))
			return
		}

		appConfigM.RSAPrivateKey = privateKeyBuff.String()
		appConfigM.RSAPublicKey = publicKeyBuff.String()

		// 插入默认配置
		err = m.appconfigDB.insert(appConfigM)
		if err != nil {
			m.Error("插入默认配置失败！", zap.Error(err))
			c.ResponseError(errors.New("插入默认配置失败！"))
			return
		}

		// 重新查询获取ID
		appConfigM, err = m.appconfigDB.query()
		if err != nil {
			m.Error("查询新创建的配置失败！", zap.Error(err))
			c.ResponseError(errors.New("查询新创建的配置失败！"))
			return
		}
	}

	// 构建更新映射
	configMap := map[string]interface{}{}
	configMap["revoke_second"] = req.RevokeSecond
	configMap["welcome_message"] = req.WelcomeMessage
	configMap["new_user_join_system_group"] = req.NewUserJoinSystemGroup
	configMap["search_by_phone"] = req.SearchByPhone
	configMap["register_invite_on"] = req.RegisterInviteOn
	configMap["send_welcome_message_on"] = req.SendWelcomeMessageOn
	configMap["invite_system_account_join_group_on"] = req.InviteSystemAccountJoinGroupOn
	configMap["register_user_must_complete_info_on"] = req.RegisterUserMustCompleteInfoOn
	configMap["channel_pinned_message_max_count"] = req.ChannelPinnedMessageMaxCount
	configMap["can_modify_api_url"] = req.CanModifyApiUrl
	configMap["moments_off"] = req.MomentsOff
	configMap["normal_user_can_add_friend"] = req.NormalUserCanAddFriend
	// 添加AI相关字段
	configMap["ai_invest_url"] = req.AIInvestUrl
	configMap["insight_url"] = req.InsightUrl
	configMap["ai_invest_on"] = req.AIInvestOn
	configMap["insight_connection_on"] = req.InsightConnectionOn

	// 尝试添加AI相关字段（如果数据库中存在）
	// 这里我们通过直接更新appConfigM对象并重新保存来处理
	appConfigM.RevokeSecond = req.RevokeSecond
	appConfigM.WelcomeMessage = req.WelcomeMessage
	appConfigM.NewUserJoinSystemGroup = req.NewUserJoinSystemGroup
	appConfigM.SearchByPhone = req.SearchByPhone
	appConfigM.RegisterInviteOn = req.RegisterInviteOn
	appConfigM.SendWelcomeMessageOn = req.SendWelcomeMessageOn
	appConfigM.InviteSystemAccountJoinGroupOn = req.InviteSystemAccountJoinGroupOn
	appConfigM.RegisterUserMustCompleteInfoOn = req.RegisterUserMustCompleteInfoOn
	appConfigM.ChannelPinnedMessageMaxCount = req.ChannelPinnedMessageMaxCount
	appConfigM.CanModifyApiUrl = req.CanModifyApiUrl
	appConfigM.MomentsOff = req.MomentsOff
	appConfigM.NormalUserCanAddFriend = req.NormalUserCanAddFriend
	appConfigM.AIInvestUrl = req.AIInvestUrl
	appConfigM.InsightUrl = req.InsightUrl
	appConfigM.AIInvestOn = req.AIInvestOn
	appConfigM.InsightConnectionOn = req.InsightConnectionOn

	// 执行更新
	err = m.appconfigDB.updateWithMap(configMap, appConfigM.Id)
	if err != nil {
		m.Error("修改app配置信息错误", zap.Error(err))
		c.ResponseError(errors.New("修改app配置信息错误"))
		return
	}

	c.ResponseOK()
}
func (m *Manager) appconfig(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	appconfig, err := m.appconfigDB.query()
	if err != nil {
		m.Error("查询应用配置失败！", zap.Error(err))
		c.ResponseError(errors.New("查询应用配置失败！"))
		return
	}
	var revokeSecond = 0
	var newUserJoinSystemGroup = 1
	var welcomeMessage = ""
	var searchByPhone = 1
	var registerInviteOn = 0
	var sendWelcomeMessageOn = 0
	var inviteSystemAccountJoinGroupOn = 0
	var registerUserMustCompleteInfoOn = 0
	var channelPinnedMessageMaxCount = 10
	var canModifyApiUrl = 0
	var momentsOff = 0
	var normalUserCanAddFriend = 0
	var aiInvestUrl = ""
	var insightUrl = ""
	var aiInvestOn = 0
	var insightConnectionOn = 0
	if appconfig != nil {
		revokeSecond = appconfig.RevokeSecond
		welcomeMessage = appconfig.WelcomeMessage
		newUserJoinSystemGroup = appconfig.NewUserJoinSystemGroup
		searchByPhone = appconfig.SearchByPhone
		registerInviteOn = appconfig.RegisterInviteOn
		sendWelcomeMessageOn = appconfig.SendWelcomeMessageOn
		inviteSystemAccountJoinGroupOn = appconfig.InviteSystemAccountJoinGroupOn
		registerUserMustCompleteInfoOn = appconfig.RegisterUserMustCompleteInfoOn
		channelPinnedMessageMaxCount = appconfig.ChannelPinnedMessageMaxCount
		canModifyApiUrl = appconfig.CanModifyApiUrl
		momentsOff = appconfig.MomentsOff
		normalUserCanAddFriend = appconfig.NormalUserCanAddFriend
		aiInvestUrl = appconfig.AIInvestUrl
		insightUrl = appconfig.InsightUrl
		aiInvestOn = appconfig.AIInvestOn
		insightConnectionOn = appconfig.InsightConnectionOn
	}
	if revokeSecond == 0 {
		revokeSecond = 120
	}
	if welcomeMessage == "" {
		welcomeMessage = m.ctx.GetConfig().WelcomeMessage
	}
	c.Response(&managerAppConfigResp{
		RevokeSecond:                   revokeSecond,
		WelcomeMessage:                 welcomeMessage,
		NewUserJoinSystemGroup:         newUserJoinSystemGroup,
		SearchByPhone:                  searchByPhone,
		RegisterInviteOn:               registerInviteOn,
		SendWelcomeMessageOn:           sendWelcomeMessageOn,
		InviteSystemAccountJoinGroupOn: inviteSystemAccountJoinGroupOn,
		RegisterUserMustCompleteInfoOn: registerUserMustCompleteInfoOn,
		ChannelPinnedMessageMaxCount:   channelPinnedMessageMaxCount,
		CanModifyApiUrl:                canModifyApiUrl,
		MomentsOff:                     momentsOff,
		NormalUserCanAddFriend:         normalUserCanAddFriend,
		AIInvestUrl:                    aiInvestUrl,
		InsightUrl:                     insightUrl,
		AIInvestOn:                     aiInvestOn,
		InsightConnectionOn:            insightConnectionOn,
	})
}

type managerAppConfigResp struct {
	RevokeSecond                   int    `json:"revoke_second"`
	WelcomeMessage                 string `json:"welcome_message"`
	NewUserJoinSystemGroup         int    `json:"new_user_join_system_group"`
	SearchByPhone                  int    `json:"search_by_phone"`
	RegisterInviteOn               int    `json:"register_invite_on"`                  // 开启注册邀请机制
	SendWelcomeMessageOn           int    `json:"send_welcome_message_on"`             // 开启注册登录发送欢迎语
	InviteSystemAccountJoinGroupOn int    `json:"invite_system_account_join_group_on"` // 开启系统账号加入群聊
	RegisterUserMustCompleteInfoOn int    `json:"register_user_must_complete_info_on"` // 注册用户必须填写完整信息
	ChannelPinnedMessageMaxCount   int    `json:"channel_pinned_message_max_count"`    // 频道置顶消息最大数量
	CanModifyApiUrl                int    `json:"can_modify_api_url"`                  // 是否可以修改api地址
	MomentsOff                     int    `json:"moments_off"`                         // 是否隐藏朋友圈入口
	NormalUserCanAddFriend         int    `json:"normal_user_can_add_friend"`          // 普通用户是否可以添加好友
	AIInvestUrl                    string `json:"ai_invest_url"`                       // AI智投链接
	InsightUrl                     string `json:"insight_url"`                         // 洞察链接
	AIInvestOn                     int    `json:"ai_invest_on"`                        // AI智投开关
	InsightConnectionOn            int    `json:"insight_connection_on"`               // 洞察连接开关
}

type managerAppModule struct {
	SID    string `json:"sid"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Status int    `json:"status"` // 模块状态 1.可选 0.不可选 2.选中不可编辑
}
