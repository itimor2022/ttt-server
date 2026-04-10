package user

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/base/event"
	common2 "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkevent"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Manager 用户管理
type Manager struct {
	ctx *config.Context
	log.Log
	db            *managerDB
	userDB        *DB
	userSettingDB *SettingDB
	deviceDB      *deviceDB
	friendDB      *friendDB
	onlineService IOnlineService
	commonService common2.IService
}

// NewManager NewManager
func NewManager(ctx *config.Context) *Manager {
	m := &Manager{
		ctx:           ctx,
		Log:           log.NewTLog("userManager"),
		db:            newManagerDB(ctx),
		deviceDB:      newDeviceDB(ctx),
		friendDB:      newFriendDB(ctx),
		userDB:        NewDB(ctx),
		userSettingDB: NewSettingDB(ctx.DB()),
		onlineService: NewOnlineService(ctx),
		commonService: common2.NewService(ctx),
	}
	m.createManagerAccount()
	return m
}

// Route 配置路由规则
func (m *Manager) Route(r *wkhttp.WKHttp) {
	user := r.Group("/v1/manager")
	{
		user.POST("/login", m.login) // 账号登录
	}
	auth := r.Group("/v1/manager", m.ctx.AuthMiddleware(r))
	{
		auth.POST("/user/admin", m.addAdminUser)              // 添加一个管理员
		auth.GET("/user/admin", m.getAdminUsers)              // 查询管理员用户
		auth.DELETE("/user/admin", m.deleteAdminUsers)        // 删除管理员用户
		auth.POST("/user/internal", m.addInternalUser)        // 添加一个内部用户
		auth.GET("/user/internal", m.getInternalUsers)        // 查询内部用户
		auth.DELETE("/user/internal", m.deleteInternalUser)   // 删除内部用户
		auth.POST("/user/add", m.addUser)                     // 添加一个用户
		auth.POST("/user/resetpassword", m.resetUserPassword) // 重置用户密码
		auth.GET("/user/list", m.list)                        // 用户列表
		auth.GET("/user/friends", m.friends)                  // 某个用户的好友
		auth.GET("/user/blacklist", m.blacklist)              // 用户黑名单列表
		auth.GET("/user/disablelist", m.disableUsers)         // 封禁用户列表
		auth.GET("user/online", m.online)                     // 在线设备信息
		auth.PUT("/user/liftban/:uid/:status", m.liftBanUser) // 解禁或封禁用户
		auth.POST("/user/updatepassword", m.updatePwd)        // 修改用户密码
		auth.GET("/user/devices", m.devices)                  // 查看某用户设备列表
		auth.POST("/user/official", m.addOfficialUser)        // 增加前端官方账号
		auth.PUT("/user/category", m.updateUserCategory)      // 修改用户分类
		auth.PUT("/user/top-setting", m.updateUserTopSetting) // 更新用户置顶设置
		auth.GET("/user/top-setting", m.getUserTopSetting)    // 获取用户置顶设置
	}
}

func (m *Manager) devices(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("请求用户uid不能为空"))
		return
	}
	devices, err := m.deviceDB.queryDeviceWithUID(uid)
	if err != nil {
		m.Error("查询用户设备列表错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户设备列表错误"))
		return
	}
	list := make([]*managerDeviceResp, 0)
	if len(devices) == 0 {
		c.Response(list)
		return
	}
	for _, device := range devices {
		list = append(list, &managerDeviceResp{
			ID:          device.Id,
			DeviceID:    device.DeviceID,
			DeviceName:  device.DeviceName,
			DeviceModel: device.DeviceModel,
			LastLogin:   util.ToyyyyMMddHHmm(time.Unix(device.LastLogin, 0)),
		})
	}
	c.Response(list)
}

func (m *Manager) online(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("请求用户uid不能为空"))
		return
	}
	list, err := m.db.queryUserOnline(uid)
	if err != nil {
		m.Error("查询用户在线设备信息错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户在线设备信息错误"))
		return
	}
	result := make([]*userOnlineResp, 0)
	if len(list) > 0 {
		for _, user := range list {
			result = append(result, &userOnlineResp{
				Online:      user.Online,
				DeviceFlag:  user.DeviceFlag,
				LastOnline:  user.LastOffline,
				LastOffline: user.LastOffline,
				UID:         user.UID,
			})
		}
	}
	c.Response(result)
}

// 用户登录
func (m *Manager) login(c *wkhttp.Context) {
	var req managerLoginReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if err := req.Check(); err != nil {
		c.ResponseError(err)
		return
	}
	userInfo, err := m.db.queryUserInfoWithNameAndPwd(req.Username)
	if err != nil {
		m.Error("登录错误", zap.Error(err))
		c.ResponseError(errors.New("登录错误！"))
		return
	}
	if userInfo == nil || userInfo.UID == "" {
		c.ResponseError(errors.New("登录用户不存在"))
		return
	}
	if userInfo.Password != util.MD5(util.MD5(req.Password)) {
		c.ResponseError(errors.New("用户名或密码错误"))
		return
	}
	if userInfo.Role != string(wkhttp.Admin) && userInfo.Role != string(wkhttp.SuperAdmin) {
		c.ResponseError(errors.New("登录账号未开通管理权限"))
		return
	}
	token := util.GenerUUID()
	// 将token设置到缓存
	err = m.ctx.Cache().SetAndExpire(m.ctx.GetConfig().Cache.TokenCachePrefix+token, fmt.Sprintf("%s@%s@%s", userInfo.UID, userInfo.Name, userInfo.Role), m.ctx.GetConfig().Cache.TokenExpire)
	if err != nil {
		m.Error("设置token缓存失败！", zap.Error(err))
		c.ResponseError(errors.New("设置token缓存失败！"))
		return
	}

	err = m.ctx.Cache().SetAndExpire(fmt.Sprintf("%s%d%s", m.ctx.GetConfig().Cache.UIDTokenCachePrefix, config.Web, userInfo.UID), token, m.ctx.GetConfig().Cache.TokenExpire)
	if err != nil {
		m.Error("设置uidtoken缓存失败！", zap.Error(err))
		c.ResponseError(errors.New("设置token缓存失败！"))
		return
	}

	c.Response(&managerLoginResp{
		UID:   userInfo.UID,
		Token: token,
		Name:  userInfo.Name,
		Role:  userInfo.Role,
	})
}

// 重置用户密码
func (m *Manager) resetUserPassword(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	type reqRUP struct {
		NewPassword              string `json:"new_password"`
		NewPassswordConfirmation string `json:"new_password_confirmation"`
		Uid                      string `json:"uid"`
	}
	var req reqRUP
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if len(req.NewPassword) < 6 {
		c.ResponseError(errors.New("密码长度必须大于6位"))
		return
	}
	if req.NewPassword != req.NewPassswordConfirmation {
		c.ResponseError(errors.New("两次密码不一致！"))
		return
	}
	if req.Uid == "" {
		c.ResponseError(errors.New("用户uid不能为空！"))
		return
	}
	user, err := m.userDB.QueryByUID(req.Uid)
	if err != nil {
		m.Error("查询用户信息错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户信息错误"))
		return
	}
	if user == nil {
		c.ResponseError(errors.New("操作用户不存在"))
		return
	}

	err = m.userDB.UpdateUsersWithField("password", util.MD5(util.MD5(req.NewPassword)), req.Uid)
	if err != nil {
		m.Error("重置用户密码错误", zap.Error(err))
		c.Response("重置用户密码错误")
		return
	}
	c.ResponseOK()
}

// updateUserTopSetting 更新用户置顶设置
func (m *Manager) updateUserTopSetting(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	type req struct {
		UID         string `json:"uid"`          // 用户UID
		IsTop       int    `json:"is_top"`       // 是否置顶：1=置顶，0=取消置顶
		TopPriority int    `json:"top_priority"` // 置顶优先级，数字越大优先级越高
	}

	var request req
	if err := c.BindJSON(&request); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if request.UID == "" {
		c.ResponseError(errors.New("用户UID不能为空"))
		return
	}

	// 查询用户是否存在
	user, err := m.userDB.QueryByUID(request.UID)
	if err != nil {
		m.Error("查询用户失败", zap.Error(err))
		c.ResponseError(errors.New("查询用户失败"))
		return
	}

	if user == nil {
		c.ResponseError(errors.New("用户不存在"))
		return
	}

	// 使用user_top_setting表存储置顶设置
	tx, err := m.ctx.DB().Begin()
	if err != nil {
		c.ResponseError(errors.New("系统繁忙，请重试"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已存在置顶配置
	var count int64
	err = tx.QueryRow("SELECT COUNT(*) FROM user_top_setting WHERE uid=?", request.UID).Scan(&count)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("查询配置失败"))
		return
	}

	if count > 0 {
		// 更新现有配置
		_, err = tx.UpdateBySql("UPDATE user_top_setting SET is_top=?, top_priority=?, updated_at=? WHERE uid=?",
			request.IsTop, request.TopPriority, time.Now(), request.UID).Exec()
	} else {
		// 插入新配置
		_, err = tx.InsertInto("user_top_setting").Columns(
			"uid", "is_top", "top_priority", "created_at", "updated_at",
		).Values(request.UID, request.IsTop, request.TopPriority, time.Now(), time.Now()).Exec()
	}

	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("更新配置失败"))
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		c.ResponseError(errors.New("操作失败，请重试"))
		return
	}

	// 如果用户被置顶，自动添加官方标识标签
	if request.IsTop == 1 {
		// 更新用户分类为官方账号
		err = m.userDB.UpdateUsersWithField("category", "official", request.UID)
		if err != nil {
			m.Error("更新用户分类失败", zap.Error(err))
			// 不返回错误，继续执行
		}

		// 跳过app_config表操作，避免表结构不一致导致的错误
		// 官方账号信息已经通过user表的category字段存储

		// 自动与所有现有用户建立好友关系
		m.autoAddFriendWithAllUsers(request.UID)
	}

	c.ResponseOK()
}

// autoAddFriendWithAllUsers 自动与所有现有用户建立好友关系
func (m *Manager) autoAddFriendWithAllUsers(topUID string) {
	// 查询所有现有用户
	rows, err := m.ctx.DB().Query("SELECT uid FROM user WHERE uid != ? AND status = 1 AND is_destroy = 0", topUID)
	if err != nil {
		m.Error("查询所有用户失败", zap.Error(err))
		return
	}
	defer rows.Close()

	userUIDs := make([]string, 0)
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err == nil && uid != "" {
			userUIDs = append(userUIDs, uid)
		}
	}

	// 为每个用户建立好友关系
	for _, userUID := range userUIDs {
		// 检查是否已经是好友
		isFriend, err := m.friendDB.IsFriend(userUID, topUID)
		if err != nil {
			m.Error("查询是否是好友失败", zap.Error(err), zap.String("userUID", userUID), zap.String("topUID", topUID))
			continue
		}
		if isFriend {
			continue
		}

		// 添加好友到数据库
		tx, err := m.ctx.DB().Begin()
		if err != nil {
			m.Error("开启事务失败", zap.Error(err))
			continue
		}

		version := m.ctx.GenSeq(common.FriendSeqKey)
		err = m.friendDB.InsertTx(&FriendModel{
			UID:     userUID,
			ToUID:   topUID,
			Version: version,
			IsAlone: 0,
			Vercode: fmt.Sprintf("%s@%d", util.GenerUUID(), common.Friend),
		}, tx)
		if err != nil {
			tx.Rollback()
			m.Error("添加好友失败", zap.Error(err), zap.String("userUID", userUID), zap.String("topUID", topUID))
			continue
		}

		// 为置顶账号添加好友关系
		err = m.friendDB.InsertTx(&FriendModel{
			UID:     topUID,
			ToUID:   userUID,
			Version: version,
			IsAlone: 0,
			Vercode: fmt.Sprintf("%s@%d", util.GenerUUID(), common.Friend),
		}, tx)
		if err != nil {
			tx.Rollback()
			m.Error("添加好友失败", zap.Error(err), zap.String("topUID", topUID), zap.String("userUID", userUID))
			continue
		}

		if err := tx.Commit(); err != nil {
			m.Error("提交事务失败", zap.Error(err))
			continue
		}
	}
}

// getUserTopSetting 获取用户置顶设置
func (m *Manager) getUserTopSetting(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("用户UID不能为空"))
		return
	}

	// 查询用户置顶设置
	type topSetting struct {
		IsTop       int `json:"is_top"`
		TopPriority int `json:"top_priority"`
	}

	var setting topSetting
	row := m.ctx.DB().QueryRow("SELECT is_top, top_priority FROM user_top_setting WHERE uid=?", uid)
	err = row.Scan(&setting.IsTop, &setting.TopPriority)
	if err != nil {
		// 如果没有找到设置，返回默认值
		setting.IsTop = 0
		setting.TopPriority = 1
	}

	c.Response(setting)
}

// 删除管理员用户
func (m *Manager) deleteAdminUsers(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("删除用户uid不能为空"))
		return
	}
	user, err := m.userDB.QueryByUID(uid)
	if err != nil {
		m.Error("查询管理员用户错误", zap.Error(err))
		c.ResponseError(errors.New("查询管理员用户错误"))
		return
	}
	if user == nil || len(user.UID) == 0 {
		c.ResponseError(errors.New("该用户不存在"))
		return
	}
	if user.Role == "" {
		c.ResponseError(errors.New("该用户不是管理员账号不能删除"))
		return
	}
	if user.Role == string(wkhttp.SuperAdmin) {
		c.ResponseError(errors.New("超级管理员账号不能删除"))
		return
	}
	err = m.db.deleteUserWithUIDAndRole(uid, string(wkhttp.Admin))
	if err != nil {
		m.Error("删除管理员错误", zap.Error(err))
		c.ResponseError(errors.New("删除管理员错误"))
		return
	}
	oldToken, err := m.ctx.Cache().Get(fmt.Sprintf("%s%d%s", m.ctx.GetConfig().Cache.UIDTokenCachePrefix, config.Web, user.UID))
	if err != nil {
		m.Error("获取旧token错误", zap.Error(err))
		c.ResponseError(errors.New("获取旧token错误"))
		return
	}
	if oldToken != "" {
		err = m.ctx.Cache().Delete(m.ctx.GetConfig().Cache.TokenCachePrefix + oldToken)
		if err != nil {
			m.Error("清除旧token数据错误", zap.Error(err))
			c.ResponseError(errors.New("清除旧token数据错误"))
			return
		}
	}
	c.ResponseOK()
}

// 添加内部用户
func (m *Manager) addInternalUser(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type reqVO struct {
		Phone    string `json:"phone"`
		Zone     string `json:"zone"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	var req reqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if req.Phone == "" {
		c.ResponseError(errors.New("手机号不能为空"))
		return
	}
	if req.Zone == "" {
		c.ResponseError(errors.New("区号不能为空"))
		return
	}
	if req.Name == "" {
		c.ResponseError(errors.New("用户名不能为空"))
		return
	}
	if req.Password == "" {
		c.ResponseError(errors.New("密码不能为空"))
		return
	}
	username := fmt.Sprintf("%s%s", req.Zone, req.Phone)
	user, err := m.userDB.QueryByUsername(username)
	if err != nil {
		m.Error("查询用户是否存在错误", zap.String("username", username))
		c.ResponseError(errors.New("查询用户是否存在错误"))
		return
	}
	if user != nil && len(user.UID) > 0 {
		c.ResponseError(errors.New("该手机号已存在"))
		return
	}
	userModel := &Model{}
	userModel.UID = util.GenerUUID()
	userModel.Name = req.Name
	userModel.Vercode = fmt.Sprintf("%s@%d", util.GenerUUID(), common.User)
	userModel.QRVercode = fmt.Sprintf("%s@%d", util.GenerUUID(), common.QRCode)
	userModel.Phone = req.Phone
	userModel.Username = username
	userModel.Zone = req.Zone
	userModel.Password = util.MD5(util.MD5(req.Password))
	userModel.ShortNo = util.Ten2Hex(time.Now().UnixNano())
	userModel.IsUploadAvatar = 0
	userModel.Category = "internal"
	userModel.NewMsgNotice = 1
	userModel.MsgShowDetail = 1
	userModel.SearchByPhone = 1
	userModel.SearchByShort = 1
	userModel.VoiceOn = 1
	userModel.ShockOn = 1
	userModel.Sex = 1
	userModel.Status = int(common.UserAvailable)
	err = m.userDB.Insert(userModel)
	if err != nil {
		m.Error("添加内部用户错误", zap.String("username", req.Name))
		c.ResponseError(err)
		return
	}
	c.ResponseOK()
}

// 查询内部用户列表
func (m *Manager) getInternalUsers(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	users, err := m.userDB.QueryByCategory("internal")
	if err != nil {
		m.Error("查询内部用户错误", zap.Error(err))
		c.ResponseError(errors.New("查询内部用户错误"))
		return
	}
	list := make([]*adminUserResp, 0)
	if len(users) > 0 {
		for _, user := range users {
			list = append(list, &adminUserResp{
				UID:          user.UID,
				Name:         user.Name,
				Username:     user.Username,
				RegisterTime: user.CreatedAt.String(),
			})
		}
	}
	c.Response(list)
}

// 删除内部用户
func (m *Manager) deleteInternalUser(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("删除用户uid不能为空"))
		return
	}
	user, err := m.userDB.QueryByUID(uid)
	if err != nil {
		m.Error("查询内部用户错误", zap.Error(err))
		c.ResponseError(errors.New("查询内部用户错误"))
		return
	}
	if user == nil || len(user.UID) == 0 {
		c.ResponseError(errors.New("该用户不存在"))
		return
	}
	if user.Category != "internal" {
		c.ResponseError(errors.New("该用户不是内部用户不能删除"))
		return
	}
	// 修改用户类别为普通用户，而不是删除用户
	_, err = m.db.session.Update("user").Set("category", "normal").Where("uid=?", uid).Exec()
	if err != nil {
		m.Error("修改内部用户为普通用户错误", zap.Error(err))
		c.ResponseError(errors.New("修改内部用户为普通用户错误"))
		return
	}
	c.ResponseOK()
}

// 查询管理员列表
func (m *Manager) getAdminUsers(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	users, err := m.db.queryUsersWithRole(string(wkhttp.Admin))
	if err != nil {
		m.Error("查询管理员用户错误", zap.Error(err))
		c.ResponseError(errors.New("查询管理员用户错误"))
		return
	}
	list := make([]*adminUserResp, 0)
	if len(users) > 0 {
		for _, user := range users {
			list = append(list, &adminUserResp{
				UID:          user.UID,
				Name:         user.Name,
				Username:     user.Username,
				RegisterTime: user.CreatedAt.String(),
			})
		}
	}
	c.Response(list)
}

// 添加一个管理员
func (m *Manager) addAdminUser(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type reqVO struct {
		Phone    string `json:"phone"`
		Zone     string `json:"zone"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	var req reqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if req.Phone == "" {
		c.ResponseError(errors.New("手机号不能为空"))
		return
	}
	if req.Zone == "" {
		c.ResponseError(errors.New("区号不能为空"))
		return
	}
	if req.Name == "" {
		c.ResponseError(errors.New("用户名不能为空"))
		return
	}
	if req.Password == "" {
		c.ResponseError(errors.New("密码不能为空"))
		return
	}
	username := fmt.Sprintf("%s%s", req.Zone, req.Phone)
	user, err := m.db.queryUserWithNameAndRole(username, string(wkhttp.Admin))
	if err != nil {
		m.Error("查询用户是否存在错误", zap.String("username", username))
		c.ResponseError(errors.New("查询用户是否存在错误"))
		return
	}
	if user != nil && len(user.UID) > 0 {
		c.ResponseError(errors.New("该手机号已存在"))
		return
	}
	userModel := &Model{}
	userModel.UID = util.GenerUUID()
	userModel.Name = req.Name
	userModel.Vercode = fmt.Sprintf("%s@%d", util.GenerUUID(), common.User)
	userModel.QRVercode = fmt.Sprintf("%s@%d", util.GenerUUID(), common.QRCode)
	userModel.Phone = req.Phone
	userModel.Username = username
	userModel.Zone = req.Zone
	userModel.Role = string(wkhttp.Admin)
	userModel.Category = "official"
	userModel.Password = util.MD5(util.MD5(req.Password))
	userModel.ShortNo = util.Ten2Hex(time.Now().UnixNano())
	userModel.IsUploadAvatar = 0
	userModel.NewMsgNotice = 0
	userModel.MsgShowDetail = 0
	userModel.SearchByPhone = 0
	userModel.SearchByShort = 0
	userModel.VoiceOn = 0
	userModel.ShockOn = 0
	userModel.Sex = 1
	userModel.Status = int(common.UserAvailable)
	err = m.userDB.Insert(userModel)
	if err != nil {
		m.Error("添加管理员错误", zap.String("username", req.Name))
		c.ResponseError(err)
		return
	}
	c.ResponseOK()
}

// 添加一个用户
func (m *Manager) addUser(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	var req managerAddUserReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if err := req.checkAddUserReq(); err != nil {
		c.ResponseError(err)
		return
	}
	userInfo, err := m.userDB.QueryByUsername(fmt.Sprintf("%s%s", req.Zone, req.Phone))
	if err != nil {
		m.Error("查询用户信息失败！", zap.String("username", req.Phone))
		c.ResponseError(err)
		return
	}
	if userInfo != nil {
		c.ResponseError(errors.New("该用户已存在"))
		return
	}
	uid := util.GenerUUID()
	var shortNo = ""
	var shortNumStatus = 0
	if m.ctx.GetConfig().ShortNo.NumOn {
		shortNo, err = m.commonService.GetShortno()
		if err != nil {
			m.Error("获取短编号失败！", zap.Error(err))
			c.ResponseError(errors.New("获取短编号失败！"))
			return
		}
	} else {
		shortNo = util.Ten2Hex(time.Now().UnixNano())
	}
	if m.ctx.GetConfig().ShortNo.EditOff {
		shortNumStatus = 1
	}
	tx, err := m.db.session.Begin()
	if err != nil {
		m.Error("开启事物错误", zap.Error(err))
		c.ResponseError(errors.New("开启事物错误"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	userModel := &Model{}
	userModel.UID = uid
	userModel.Name = req.Name
	userModel.Vercode = fmt.Sprintf("%s@%d", util.GenerUUID(), common.User)
	userModel.QRVercode = fmt.Sprintf("%s@%d", util.GenerUUID(), common.QRCode)
	userModel.Phone = req.Phone
	userModel.Username = fmt.Sprintf("%s%s", req.Zone, req.Phone)
	userModel.Zone = req.Zone
	userModel.Password = util.MD5(util.MD5(req.Password))
	userModel.ShortNo = shortNo
	userModel.IsUploadAvatar = 0
	userModel.Category = "normal"
	userModel.NewMsgNotice = 1
	userModel.MsgShowDetail = 1
	userModel.SearchByPhone = 1
	userModel.ShortStatus = shortNumStatus
	userModel.SearchByShort = 1
	userModel.VoiceOn = 1
	userModel.ShockOn = 1
	userModel.Sex = req.Sex
	userModel.Status = int(common.UserAvailable)
	err = m.userDB.insertTx(userModel, tx)
	if err != nil {
		tx.Rollback()
		m.Error("添加用户错误", zap.String("username", req.Phone))
		c.ResponseError(err)
		return
	}

	err = m.addSystemFriend(uid)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("添加后台生成用户和系统账号为好友关系失败"))
		return
	}
	err = m.addFileHelperFriend(uid)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("添加后台生成用户和文件助手为好友关系失败"))
		return
	}
	//发送用户注册事件
	eventID, err := m.ctx.EventBegin(&wkevent.Data{
		Event: event.EventUserRegister,
		Type:  wkevent.Message,
		Data: map[string]interface{}{
			"uid": uid,
		},
	}, tx)
	if err != nil {
		tx.RollbackUnlessCommitted()
		m.Error("开启事件失败！", zap.Error(err))
		c.ResponseError(errors.New("开启事件失败！"))
		return
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		m.Error("数据库事物提交失败", zap.Error(err))
		c.ResponseError(errors.New("数据库事物提交失败"))
		return
	}
	m.ctx.EventCommit(eventID)
	c.ResponseOK()
}

// 用户列表
func (m *Manager) list(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	keyword := c.Query("keyword")
	onlineStr := c.Query("online")

	var online int64 = -1
	if strings.TrimSpace(onlineStr) != "" {
		online, _ = strconv.ParseInt(onlineStr, 10, 64)
	}
	pageIndex, pageSize := c.GetPage()
	var userList []*managerUserModel
	var count int64
	if keyword == "" {
		userList, err = m.db.queryUserListWithPage(uint64(pageSize), uint64(pageIndex), int(online))
		if err != nil {
			m.Error("查询用户列表报错", zap.Error(err))
			c.ResponseError(err)
			return
		}

		count, err = m.userDB.queryUserCount()
		if err != nil {
			m.Error("查询用户数量错误", zap.Error(err))
			c.ResponseError(errors.New("查询用户数量错误"))
			return
		}
	} else {
		userList, err = m.db.queryUserListWithPageAndKeyword(keyword, int(online), uint64(pageSize), uint64(pageIndex))
		if err != nil {
			m.Error("查询用户列表报错", zap.Error(err))
			c.ResponseError(err)
			return
		}

		count, err = m.db.queryUserCountWithKeyWord(keyword)
		if err != nil {
			m.Error("查询用户数量错误", zap.Error(err))
			c.ResponseError(errors.New("查询用户数量错误"))
			return
		}
	}

	// 使用原始用户列表，不进行管理员置顶
	combinedUsers := userList

	result := make([]*managerUserResp, 0)
	if len(combinedUsers) > 0 {
		uids := make([]string, 0)
		for _, user := range combinedUsers {
			uids = append(uids, user.UID)
		}
		resps, err := m.onlineService.GetUserLastOnlineStatus(uids)
		respsdata := map[string]*config.OnlinestatusResp{}
		if len(resps) > 0 {
			for _, v := range resps {
				respsdata[v.UID] = v
			}
		}
		if err != nil {
			m.Error("查询用户在线状态失败", zap.Error(err))
			c.ResponseError(errors.New("查询用户在线状态失败"))
			return
		}
		devices, err := m.deviceDB.queryDeviceLastLoginWithUids(uids)
		if err != nil {
			m.Error("查询用户最后一次登录设备信息错误", zap.Error(err))
			c.ResponseError(errors.New("查询用户最后一次登录设备信息错误"))
			return
		}
		var i = 0
		for _, user := range combinedUsers {
			var device *deviceModel
			if len(devices) > 0 {
				for _, model := range devices {
					if model.UID == user.UID {
						device = model
						break
					}
				}
			}
			var lastLoginTime string
			var deviceName string = ""
			var deviceModel string = ""
			var online int
			var lastOnlineTime string = ""
			if device != nil {
				deviceModel = device.DeviceModel
				deviceName = device.DeviceName
				lastLoginTime = util.ToyyyyMMddHHmm(time.Unix(device.LastLogin, 0))
			}
			/* if i < len(resps) {
				online = resps[i].Online
				lastOnlineTime = util.ToyyyyMMddHHmm(time.Unix(int64(resps[i].LastOffline), 0))
			} */
			if respsdata[user.UID] != nil {
				online = respsdata[user.UID].Online
				lastOnlineTime = util.ToyyyyMMddHHmm(time.Unix(int64(respsdata[user.UID].LastOffline), 0))
			}
			showPhone := getShowPhoneNum(user.Phone)
			result = append(result, &managerUserResp{
				UID:            user.UID,
				Username:       user.Username,
				Name:           user.Name,
				Phone:          showPhone,
				Sex:            user.Sex,
				ShortNo:        user.ShortNo,
				LastLoginTime:  lastLoginTime,
				DeviceName:     deviceName,
				DeviceModel:    deviceModel,
				Online:         online,
				LastOnlineTime: lastOnlineTime,
				RegisterTime:   user.CreatedAt.String(),
				Status:         user.Status,
				IsDestroy:      user.IsDestroy,
				Category:       user.Category,
				Role:           user.Role,
				GiteeUID:       user.GiteeUID,
				GithubUID:      user.GithubUID,
				WXOpenid:       user.WXOpenid,
				Avatar:         user.Avatar,
			})
			i++
		}
	}
	c.Response(map[string]interface{}{
		"list":  result,
		"count": count,
	})
}

// 查询某个用户的好友
func (m *Manager) friends(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("查询用户ID不能为空"))
		return
	}
	sortType := c.Query("sort_type")
	if sortType == "" {
		sortType = "1"
	}
	sortTypeInt, err := strconv.Atoi(sortType)
	if err != nil {
		sortTypeInt = 0
	}
	list, err := m.friendDB.QueryFriends(uid)
	if err != nil {
		m.Error("查询用户好友错误", zap.String("uid", uid))
		c.ResponseError(err)
		return
	}
	result := make([]*managerFriendResp, 0)
	if len(list) == 0 {
		c.Response(result)
		return
	}
	if sortTypeInt == 0 {
		for _, friend := range list {
			result = append(result, &managerFriendResp{
				UID:              friend.ToUID,
				Remark:           friend.Remark,
				Name:             friend.ToName,
				RelationshipTime: friend.CreatedAt.String(),
			})
		}
		c.Response(result)
		return
	}
	// 查询最近会话
	conversations, err := m.ctx.IMSyncUserConversation(uid, 0, 1, "", nil)
	if err != nil {
		m.Error("同步离线后的最近会话失败！", zap.Error(err), zap.String("loginUID", uid))
		c.ResponseError(errors.New("同步离线后的最近会话失败！"))
		return
	}
	if len(conversations) == 0 {
		for _, friend := range list {
			result = append(result, &managerFriendResp{
				UID:              friend.ToUID,
				Remark:           friend.Remark,
				Name:             friend.ToName,
				RelationshipTime: friend.CreatedAt.String(),
			})
		}
		c.Response(result)
		return
	}
	sort.SliceStable(conversations, func(i, j int) bool {
		return conversations[i].Timestamp > conversations[j].Timestamp
	})
	for _, conv := range conversations {
		if conv.ChannelType != common.ChannelTypePerson.Uint8() {
			continue
		}
		var f *DetailModel
		for _, friend := range list {
			if friend.ToUID == conv.ChannelID {
				f = friend
				break
			}
		}
		if f != nil {
			result = append(result, &managerFriendResp{
				UID:              f.ToUID,
				Remark:           f.Remark,
				Name:             f.ToName,
				RelationshipTime: f.CreatedAt.String(),
			})
		}
	}
	for _, f := range list {
		isAdd := true
		for _, r := range result {
			if r.UID == f.ToUID {
				isAdd = false
				break
			}
		}
		if isAdd {
			result = append(result, &managerFriendResp{
				UID:              f.ToUID,
				Remark:           f.Remark,
				Name:             f.ToName,
				RelationshipTime: f.CreatedAt.String(),
			})
		}
	}
	c.Response(result)
}

// 查询某个用户的黑名单
func (m *Manager) blacklist(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Query("uid")
	if uid == "" {
		c.ResponseError(errors.New("查询用户ID不能为空"))
		return
	}
	list, err := m.db.queryUserBlacklists(uid)
	if err != nil {
		m.Error("查询黑名单列表失败！", zap.Error(err))
		c.ResponseError(errors.New("查询黑名单列表失败！"))
		return
	}
	blacklists := []*managerBlackUserResp{}
	for _, result := range list {
		blacklists = append(blacklists, &managerBlackUserResp{
			UID:      result.UID,
			Name:     result.Name,
			CreateAt: result.UpdatedAt.String(),
		})
	}
	c.Response(blacklists)
}

// 查看封禁用户列表
func (m *Manager) disableUsers(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	pageIndex, pageSize := c.GetPage()
	list, err := m.db.queryUserListWithStatus(int(common.UserDisable), uint64(pageSize), uint64(pageIndex))
	if err != nil {
		m.Error("通过状态查询用户列表错误", zap.Error(err))
		c.ResponseError(errors.New("通过状态查询用户列表错误"))
		return
	}
	count, err := m.db.queryUserCountWithStatus(int(common.UserDisable))
	if err != nil {
		m.Error("查询用户数量错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户数量错误"))
		return
	}
	result := make([]*managerDisableUserResp, 0)
	if len(list) > 0 {
		for _, user := range list {
			showPhone := getShowPhoneNum(user.Phone)
			result = append(result, &managerDisableUserResp{
				Name:         user.Name,
				ShortNo:      user.ShortNo,
				Phone:        showPhone,
				UID:          user.UID,
				ClosureTime:  user.UpdatedAt.String(),
				RegisterTime: user.CreatedAt.String(),
			})
		}
	}
	c.Response(map[string]interface{}{
		"list":  result,
		"count": count,
	})
}

// 封禁或解禁用户
func (m *Manager) liftBanUser(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	uid := c.Param("uid")
	status := c.Param("status")
	if uid == "" {
		c.ResponseError(errors.New("操作用户id不能为空"))
		return
	}
	if status == "" {
		c.ResponseError(errors.New("修改状态类型不能为空"))
		return
	}
	userStatus, _ := strconv.Atoi(status)
	if userStatus != int(common.UserAvailable) && userStatus != int(common.UserDisable) {
		c.ResponseError(errors.New("修改状态类型不匹配"))
		return
	}
	userInfo, err := m.userDB.QueryByUID(uid)
	if err != nil {
		m.Error("查询用户信息失败！", zap.String("uid", uid))
		c.ResponseError(errors.New("查询用户信息错误"))
		return
	}
	if userInfo == nil {
		c.ResponseError(errors.New("操作用户不存在"))
		return
	}
	if userInfo.Status == userStatus {
		c.ResponseOK()
		return
	}
	err = m.userDB.UpdateUsersWithField("status", status, uid)
	if err != nil {
		m.Error("修改用户状态错误", zap.Error(err))
		c.ResponseError(errors.New("修改用户状态错误"))
		return
	}

	ban := 0
	if userStatus == int(common.UserDisable) {
		ban = 1
	}

	err = m.ctx.IMCreateOrUpdateChannelInfo(&config.ChannelInfoCreateReq{
		ChannelID:   uid,
		ChannelType: common.ChannelTypePerson.Uint8(),
		Ban:         ban,
	})
	if err != nil {
		m.Error("更新WebIM的token失败！", zap.Error(err))
		c.ResponseError(errors.New("更新IM的token失败！"))
		return
	}
	err = m.ctx.QuitUserDevice(userInfo.UID, -1)
	if err != nil {
		m.Error("下线用户所有登录设备错误", zap.Error(err), zap.String("uid", uid))
		c.ResponseError(errors.New("下线用户所有登录设备错误"))
		return
	}
	c.ResponseOK()
}

// 修改登录密码
func (m *Manager) updatePwd(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	loginUID := c.GetLoginUID()
	type updatePwdReq struct {
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}
	var req updatePwdReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if req.Password == "" || req.NewPassword == "" {
		c.ResponseError(errors.New("密码不能为空"))
		return
	}
	user, err := m.userDB.QueryByUID(loginUID)
	if err != nil {
		m.Error("查询用户信息错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户信息错误"))
		return
	}
	if user == nil {
		c.ResponseError(errors.New("操作用户不存在"))
		return
	}
	if util.MD5(util.MD5(req.Password)) != user.Password {
		c.ResponseError(errors.New("原密码错误"))
		return
	}
	if len(req.NewPassword) < 6 {
		c.ResponseError(errors.New("密码长度必须大于6位"))
		return
	}
	if req.Password == req.NewPassword {
		c.ResponseError(errors.New("新密码不能和旧密码一样"))
		return
	}
	err = m.userDB.UpdateUsersWithField("password", util.MD5(util.MD5(req.NewPassword)), loginUID)
	if err != nil {
		m.Error("修改用户密码错误", zap.Error(err))
		c.Response("修改用户密码错误")
		return
	}
	// 清除token缓存
	oldToken, err := m.ctx.Cache().Get(fmt.Sprintf("%s%d%s", m.ctx.GetConfig().Cache.UIDTokenCachePrefix, config.Web, user.UID))
	if err != nil {
		m.Error("获取旧token错误", zap.Error(err))
		c.ResponseError(errors.New("获取旧token错误"))
		return
	}
	if oldToken != "" {
		err = m.ctx.Cache().Delete(m.ctx.GetConfig().Cache.TokenCachePrefix + oldToken)
		if err != nil {
			m.Error("清除旧token数据错误", zap.Error(err))
			c.ResponseError(errors.New("清除旧token数据错误"))
			return
		}
	}
	c.ResponseOK()
}
func (r managerAddUserReq) checkAddUserReq() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("用户名不能为空！")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("密码不能为空！")
	}
	if strings.TrimSpace(r.Phone) == "" {
		return errors.New("手机号不能为空！")
	}

	return nil
}
func (r managerLoginReq) Check() error {
	if strings.TrimSpace(r.Username) == "" {
		return errors.New("用户名不能为空！")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("密码不能为空！")
	}
	return nil
}

// 处理注册用户和文件助手互为好友
func (m *Manager) addFileHelperFriend(uid string) error {
	if uid == "" {
		m.Error("用户ID不能为空")
		return errors.New("用户ID不能为空")
	}
	isFriend, err := m.friendDB.IsFriend(uid, m.ctx.GetConfig().Account.FileHelperUID)
	if err != nil {
		m.Error("查询用户关系失败")
		return err
	}
	if !isFriend {
		version := m.ctx.GenSeq(common.FriendSeqKey)
		err := m.friendDB.Insert(&FriendModel{
			UID:     uid,
			ToUID:   m.ctx.GetConfig().Account.FileHelperUID,
			Version: version,
		})
		if err != nil {
			m.Error("注册用户和文件助手成为好友失败")
			return err
		}
	}
	return nil
}

// addSystemFriend 处理注册用户和系统账号互为好友
func (m *Manager) addSystemFriend(uid string) error {

	if uid == "" {
		m.Error("用户ID不能为空")
		return errors.New("用户ID不能为空")
	}
	isFriend, err := m.friendDB.IsFriend(uid, m.ctx.GetConfig().Account.SystemUID)
	if err != nil {
		m.Error("查询用户关系失败")
		return err
	}
	tx, err := m.friendDB.session.Begin()
	if err != nil {
		m.Error("开启事物错误", zap.Error(err))
		return errors.New("开启事物错误")
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	if !isFriend {
		version := m.ctx.GenSeq(common.FriendSeqKey)
		err := m.friendDB.InsertTx(&FriendModel{
			UID:     uid,
			ToUID:   m.ctx.GetConfig().Account.SystemUID,
			Version: version,
		}, tx)
		if err != nil {
			m.Error("注册用户和系统账号成为好友失败")
			tx.Rollback()
			return err
		}
	}
	// systemIsFriend, err := u.friendDB.IsFriend(u.ctx.GetConfig().SystemUID, uid)
	// if err != nil {
	// 	u.Error("查询系统账号和注册用户关系失败")
	// 	tx.Rollback()
	// 	return err
	// }
	// if !systemIsFriend {
	// 	version := u.ctx.GenSeq(common.FriendSeqKey)
	// 	err := u.friendDB.InsertTx(&FriendModel{
	// 		UID:     u.ctx.GetConfig().SystemUID,
	// 		ToUID:   uid,
	// 		Version: version,
	// 	}, tx)
	// 	if err != nil {
	// 		u.Error("系统账号和注册用户成为好友失败")
	// 		tx.Rollback()
	// 		return err
	// 	}
	// }
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		m.Error("用户注册数据库事物提交失败", zap.Error(err))
		return err
	}
	return nil
}

// 创建一个系统管理账户
func (m *Manager) createManagerAccount() {
	// 获取管理员UID，如果配置中没有设置则使用默认值
	adminUID := m.ctx.GetConfig().Account.AdminUID
	if adminUID == "" {
		adminUID = "admin"
	}

	user, err := m.userDB.QueryByUID(adminUID)
	if err != nil {
		m.Error("查询系统管理账号错误", zap.Error(err))
		return
	}
	if user != nil && user.UID != "" {
		return
	}

	username := "admin"
	role := string(wkhttp.SuperAdmin)
	var pwd = m.ctx.GetConfig().AdminPwd
	// 如果 AdminPwd 为空，使用默认密码
	if pwd == "" {
		pwd = "123456"
	}
	err = m.userDB.Insert(&Model{
		UID:      adminUID,
		Name:     "超级管理员",
		ShortNo:  "30000",
		Category: "system",
		Role:     role,
		Username: username,
		Zone:     "0086",
		Phone:    "13000000002",
		Status:   1,
		Password: util.MD5(util.MD5(pwd)),
	})
	if err != nil {
		m.Error("新增系统管理员错误", zap.Error(err))
		return
	}
}
func getShowPhoneNum(mobile string) string {
	if len(mobile) <= 3 {
		return mobile
	}
	phone := mobile[:3]
	var length = len(mobile) - 3
	if length > 4 {
		length = 4
	}
	for i := 0; i < length; i++ {
		phone = fmt.Sprintf("%s*", phone)
	}
	var index = 3 + length
	if index > 0 && index < len(mobile) {
		return phone + mobile[index:]
	}
	return phone
}

type managerLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type managerLoginResp struct {
	UID   string `json:"uid"`
	Token string `json:"token"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
type managerAddUserReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Zone     string `json:"zone"`
	Sex      int    `json:"sex"`
}
type managerBlackUserResp struct {
	Name     string `json:"name"`
	UID      string `json:"uid"`
	CreateAt string `json:"create_at"`
}
type adminUserResp struct {
	Name         string `json:"name"`
	UID          string `json:"uid"`
	Username     string `json:"username"`
	RegisterTime string `json:"register_time"`
}
type managerUserResp struct {
	Name           string `json:"name"`
	UID            string `json:"uid"`
	Phone          string `json:"phone"`
	Username       string `json:"username"`
	ShortNo        string `json:"short_no"`
	Sex            int    `json:"sex"`
	RegisterTime   string `json:"register_time"`
	LastLoginTime  string `json:"last_login_time"`
	DeviceName     string `json:"device_name"`
	DeviceModel    string `json:"device_model"`
	Online         int    `json:"online"`
	LastOnlineTime string `json:"last_online_time"`
	Status         int    `json:"status"`
	IsDestroy      int    `json:"is_destroy"`
	Category       string `json:"category"`   // 用户分类
	Role           string `json:"role"`       // 角色 admin/superAdmin
	WXOpenid       string `json:"wx_openid"`  // 微信openid
	GiteeUID       string `json:"gitee_uid"`  // gitee uid
	GithubUID      string `json:"github_uid"` // github uid
	Avatar         string `json:"avatar"`     // 用户头像
}

type managerFriendResp struct {
	Name             string `json:"name"`
	UID              string `json:"uid"`
	Remark           string `json:"remark"`
	RelationshipTime string `json:"relationship_time"`
}

type managerDisableUserResp struct {
	Name         string `json:"name"`
	UID          string `json:"uid"`
	ShortNo      string `json:"short_no"`
	Sex          int    `json:"sex"`
	RegisterTime string `json:"register_time"`
	Phone        string `json:"phone"`
	ClosureTime  string `json:"closure_time"`
}

type managerDeviceResp struct {
	ID          int64  `json:"id"`
	DeviceID    string `json:"device_id"`    // 设备ID
	DeviceName  string `json:"device_name"`  // 设备名称
	DeviceModel string `json:"device_model"` // 设备型号
	LastLogin   string `json:"last_login"`   // 设备最后一次登录时间
	Self        int    `json:"self"`         // 是否是本机
}

type userOnlineResp struct {
	UID         string `json:"uid"`
	DeviceFlag  uint8  `json:"device_flag"`
	LastOnline  int    `json:"last_online"`
	LastOffline int    `json:"last_offline"`
	Online      int    `json:"online"`
}

type addOfficialUserReq struct {
	UID      string `json:"uid"`      // 官方账号UID
	Nickname string `json:"nickname"` // 官方账号昵称
	Avatar   string `json:"avatar"`   // 官方账号头像
	Phone    string `json:"phone"`    // 官方账号手机号
	Email    string `json:"email"`    // 官方账号邮箱
}

func newUserOnlineResp(m *onlineStatusWeightModel) *userOnlineResp {

	return &userOnlineResp{
		UID:         m.UID,
		DeviceFlag:  m.DeviceFlag,
		LastOnline:  m.LastOnline,
		LastOffline: m.LastOffline,
		Online:      m.Online,
	}
}

// updateUserCategory 修改用户分类
func (m *Manager) updateUserCategory(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	type req struct {
		UID      string `json:"uid"`      // 用户UID
		Category string `json:"category"` // 新的用户分类
	}

	var request req
	if err := c.BindJSON(&request); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if request.UID == "" {
		c.ResponseError(errors.New("用户UID不能为空"))
		return
	}

	if request.Category == "" {
		c.ResponseError(errors.New("用户分类不能为空"))
		return
	}

	// 查询用户是否存在
	user, err := m.userDB.QueryByUID(request.UID)
	if err != nil {
		m.Error("查询用户失败", zap.Error(err))
		c.ResponseError(errors.New("查询用户失败"))
		return
	}

	if user == nil {
		c.ResponseError(errors.New("用户不存在"))
		return
	}

	// 更新用户分类
	err = m.userDB.UpdateUsersWithField("category", request.Category, request.UID)
	if err != nil {
		m.Error("更新用户分类失败", zap.Error(err))
		c.ResponseError(errors.New("更新用户分类失败"))
		return
	}

	// 如果用户分类被设置为官方账号，在app_config表中存储官方账号信息
	if request.Category == "official" {
		// 在app_config表中存储官方账号信息
		tx, err := m.ctx.DB().Begin()
		if err != nil {
			m.Error("开启事务失败", zap.Error(err))
			// 不返回错误，继续执行
		} else {
			defer func() {
				if err != nil {
					tx.Rollback()
				}
			}()

			// 检查是否已存在官方账号配置
			var count int64
			err = tx.QueryRow("SELECT COUNT(*) FROM app_config WHERE module=? AND `key`=?", "official", request.UID).Scan(&count)
			if err != nil {
				m.Error("查询配置失败", zap.Error(err))
				// 不返回错误，继续执行
			} else {
				if count > 0 {
					// 更新现有配置
					_, err = tx.UpdateBySql("UPDATE app_config SET value=?, updated_at=? WHERE module=? AND `key`=?", user.Name, time.Now(), "official", request.UID).Exec()
				} else {
					// 插入新配置
					_, err = tx.InsertInto("app_config").Columns(
						"module", "`key`", "value", "created_at", "updated_at",
					).Values("official", request.UID, user.Name, time.Now(), time.Now()).Exec()
				}
				if err != nil {
					m.Error("更新官方账号配置失败", zap.Error(err))
					// 不返回错误，继续执行
				} else {
					// 提交事务
					err = tx.Commit()
					if err != nil {
						m.Error("提交事务失败", zap.Error(err))
						// 不返回错误，继续执行
					}
				}
			}
		}
	}

	c.ResponseOK()
}

// addOfficialUser 增加前端官方账号
func (m *Manager) addOfficialUser(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	var req addOfficialUserReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if req.UID == "" {
		c.ResponseError(errors.New("官方账号UID不能为空"))
		return
	}

	if req.Nickname == "" {
		c.ResponseError(errors.New("官方账号昵称不能为空"))
		return
	}

	// 检查用户是否已存在
	user, err := m.userDB.QueryByUID(req.UID)
	if err != nil {
		m.Error("查询用户失败", zap.Error(err))
		c.ResponseError(errors.New("查询用户失败"))
		return
	}

	if user == nil {
		// 创建新用户
		user = &Model{
			UID:      req.UID,
			Name:     req.Nickname,
			Avatar:   req.Avatar,
			Phone:    req.Phone,
			Email:    req.Email,
			Status:   1,
			Category: "official",
		}

		// 插入用户
		err = m.userDB.Insert(user)
		if err != nil {
			m.Error("创建官方账号失败", zap.Error(err))
			c.ResponseError(errors.New("创建官方账号失败"))
			return
		}
	} else {
		// 更新现有用户信息
		user.Name = req.Nickname
		if req.Avatar != "" {
			user.Avatar = req.Avatar
		}
		if req.Phone != "" {
			user.Phone = req.Phone
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		user.Status = 1
		user.Category = "official"

		// 更新用户
		err = m.userDB.UpdateUsersWithField("name", req.Nickname, user.UID)
		if err != nil {
			m.Error("更新官方账号失败", zap.Error(err))
			c.ResponseError(errors.New("更新官方账号失败"))
			return
		}

		if req.Avatar != "" {
			err = m.userDB.UpdateUsersWithField("avatar", req.Avatar, user.UID)
			if err != nil {
				m.Error("更新官方账号失败", zap.Error(err))
				c.ResponseError(errors.New("更新官方账号失败"))
				return
			}
		}

		if req.Phone != "" {
			err = m.userDB.UpdateUsersWithField("phone", req.Phone, user.UID)
			if err != nil {
				m.Error("更新官方账号失败", zap.Error(err))
				c.ResponseError(errors.New("更新官方账号失败"))
				return
			}
		}

		if req.Email != "" {
			err = m.userDB.UpdateUsersWithField("email", req.Email, user.UID)
			if err != nil {
				m.Error("更新官方账号失败", zap.Error(err))
				c.ResponseError(errors.New("更新官方账号失败"))
				return
			}
		}

		err = m.userDB.UpdateUsersWithField("status", "1", user.UID)
		if err != nil {
			m.Error("更新官方账号失败", zap.Error(err))
			c.ResponseError(errors.New("更新官方账号失败"))
			return
		}

		err = m.userDB.UpdateUsersWithField("category", "official", user.UID)
		if err != nil {
			m.Error("更新官方账号失败", zap.Error(err))
			c.ResponseError(errors.New("更新官方账号失败"))
			return
		}
	}

	// 标记为官方账号（可以通过设置用户扩展信息或特殊字段来实现）
	// 这里假设我们使用app_config表来存储官方账号信息
	tx, err := m.ctx.DB().Begin()
	if err != nil {
		c.ResponseError(errors.New("系统繁忙，请重试"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已存在官方账号配置
	var count int64
	err = tx.QueryRow("SELECT COUNT(*) FROM app_config WHERE module=? AND `key`=?", "official", req.UID).Scan(&count)
	if err != nil {
		m.Error("查询官方账号配置失败", zap.Error(err))
		return
	}
	if count > 0 {
		// 更新现有配置
		_, err = tx.UpdateBySql("UPDATE app_config SET value=?, updated_at=? WHERE module=? AND `key`=?", req.Nickname, time.Now(), "official", req.UID).Exec()
	} else {
		// 插入新配置
		_, err = tx.InsertInto("app_config").Columns(
			"module", "`key`", "value", "created_at", "updated_at",
		).Values("official", req.UID, req.Nickname, time.Now(), time.Now()).Exec()
	}

	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("更新配置失败"))
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		c.ResponseError(errors.New("操作失败，请重试"))
		return
	}

	c.ResponseOK()
}
