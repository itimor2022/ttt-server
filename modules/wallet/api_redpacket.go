package wallet

import (
	"errors"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/user"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// RedPacket 红包API
type RedPacket struct {
	ctx         *config.Context
	service     *Service
	userService user.IService
	log.Log
}

// NewRedPacket 创建红包API
func NewRedPacket(ctx *config.Context) *RedPacket {
	return &RedPacket{
		ctx:         ctx,
		service:     NewService(ctx),
		userService: user.NewService(ctx),
		Log:         log.NewTLog("redPacketAPI"),
	}
}

// Route 路由
func (rp *RedPacket) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/wallet/redpacket", rp.ctx.AuthMiddleware(r))
	{
		auth.POST("/send", rp.send)                 // 发红包
		auth.POST("/:packet_no/grab", rp.grab)      // 抢红包
		auth.GET("/:packet_no", rp.detail)          // 红包详情
		auth.GET("/:packet_no/records", rp.records) // 领取记录
		auth.GET("/sent", rp.sentList)              // 发出的红包列表
		auth.GET("/received", rp.receivedList)      // 收到的红包列表
	}
}

// send 发红包
func (rp *RedPacket) send(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req redPacketSendReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if req.TotalAmount <= 0 {
		c.ResponseError(errors.New("红包金额必须大于0"))
		return
	}
	if req.TotalCount <= 0 {
		c.ResponseError(errors.New("红包个数必须大于0"))
		return
	}
	if req.TotalAmount < int64(req.TotalCount) {
		c.ResponseError(errors.New("红包金额不能少于红包个数"))
		return
	}
	if req.Type != RedPacketTypeNormal && req.Type != RedPacketTypeLucky {
		req.Type = RedPacketTypeNormal
	}
	if req.ChannelId == "" {
		c.ResponseError(errors.New("频道ID不能为空"))
		return
	}
	if req.Remark == "" {
		req.Remark = "恭喜发财，大吉大利"
	}

	// 验证支付密码
	wallet, err := rp.service.GetOrCreateWallet(uid)
	if err != nil {
		rp.Error("获取钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取钱包失败"))
		return
	}
	if wallet.PayPassword == "" {
		c.ResponseError(errors.New("请先设置支付密码"))
		return
	}
	pwdHash := rp.service.EncryptPassword(req.PayPassword, wallet.PayPasswordSalt)
	if pwdHash != wallet.PayPassword {
		c.ResponseError(errors.New("支付密码错误"))
		return
	}

	// 检查余额
	if wallet.Balance < req.TotalAmount {
		c.ResponseError(errors.New("余额不足"))
		return
	}

	// 扣除余额
	err = rp.service.walletDB.subBalance(uid, req.TotalAmount)
	if err != nil {
		rp.Error("扣除余额失败", zap.Error(err))
		c.ResponseError(errors.New("扣除余额失败"))
		return
	}

	// 创建红包
	packetNo := rp.service.GenerateOrderNo("RP")
	expireTime := time.Now().Add(24 * time.Hour) // 24小时过期

	packet := &redPacketModel{
		PacketNo:     packetNo,
		UID:          uid,
		ChannelId:    req.ChannelId,
		ChannelType:  req.ChannelType,
		Type:         req.Type,
		TotalAmount:  req.TotalAmount,
		TotalCount:   req.TotalCount,
		RemainAmount: req.TotalAmount,
		RemainCount:  req.TotalCount,
		Remark:       req.Remark,
		Status:       RedPacketStatusActive,
		ExpireTime:   expireTime,
	}

	packetId, err := rp.service.redPacketDB.insert(packet)
	if err != nil {
		// 回滚余额
		rp.service.walletDB.addBalance(uid, req.TotalAmount)
		rp.Error("创建红包失败", zap.Error(err))
		c.ResponseError(errors.New("创建红包失败"))
		return
	}

	// 添加流水记录
	rp.service.AddRecord(uid, RecordTypeRedPacketOut, req.TotalAmount, wallet.Balance, wallet.Balance-req.TotalAmount, "发红包-"+req.Remark, packetNo, "")

	c.Response(&redPacketSendResp{
		PacketNo:    packetNo,
		PacketId:    packetId,
		TotalAmount: req.TotalAmount,
		TotalCount:  req.TotalCount,
		Remark:      req.Remark,
		ExpireTime:  expireTime.Format("2006-01-02 15:04:05"),
	})
}

// grab 抢红包
func (rp *RedPacket) grab(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	packetNo := c.Param("packet_no")
	if packetNo == "" {
		c.ResponseError(errors.New("红包编号不能为空"))
		return
	}

	// 使用事务处理抢红包
	grabResult, err := rp.service.GrabRedPacket(uid, packetNo)
	if err != nil {
		rp.Error("抢红包失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.Response(&redPacketGrabResp{
		Amount:  grabResult.Amount,
		IsBest:  grabResult.IsBest,
		Remark:  grabResult.Remark,
		FromUID: grabResult.FromUID,
	})
}

// detail 红包详情
func (rp *RedPacket) detail(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	packetNo := c.Param("packet_no")

	packet, err := rp.service.redPacketDB.queryByPacketNo(packetNo)
	if err != nil {
		rp.Error("查询红包失败", zap.Error(err))
		c.ResponseError(errors.New("查询红包失败"))
		return
	}
	if packet == nil {
		c.ResponseError(errors.New("红包不存在"))
		return
	}

	// 查询领取记录
	records, _ := rp.service.redPacketRecordDB.queryByPacketId(packet.Id)

	// 获取所有领取用户的UID
	uids := make([]string, 0, len(records))
	for _, r := range records {
		uids = append(uids, r.UID)
	}

	// 批量获取用户信息
	userMap := make(map[string]*user.Resp)
	if len(uids) > 0 {
		users, err := rp.userService.GetUsers(uids)
		if err == nil && len(users) > 0 {
			for _, u := range users {
				userMap[u.UID] = u
			}
		}
	}

	recordList := make([]*redPacketRecordResp, 0, len(records))
	var myGrabAmount int64 = 0
	var myGrabIsBest bool = false
	for _, r := range records {
		name := r.UID
		avatar := ""
		if u, ok := userMap[r.UID]; ok {
			name = u.Name
			avatar = rp.ctx.GetConfig().External.APIBaseURL + "/users/" + r.UID + "/avatar"
		}
		recordList = append(recordList, &redPacketRecordResp{
			UID:       r.UID,
			Name:      name,
			Avatar:    avatar,
			Amount:    r.Amount,
			IsBest:    r.IsBest == 1,
			CreatedAt: r.CreatedAt.String(),
		})
		// 查找当前用户抢到的金额
		if r.UID == uid {
			myGrabAmount = r.Amount
			myGrabIsBest = r.IsBest == 1
		}
	}

	c.Response(&redPacketDetailResp{
		PacketNo:     packet.PacketNo,
		UID:          packet.UID,
		Type:         packet.Type,
		TotalAmount:  packet.TotalAmount,
		TotalCount:   packet.TotalCount,
		RemainAmount: packet.RemainAmount,
		RemainCount:  packet.RemainCount,
		Remark:       packet.Remark,
		Status:       packet.Status,
		ExpireTime:   packet.ExpireTime.Format("2006-01-02 15:04:05"),
		CreatedAt:    packet.CreatedAt.String(),
		Records:      recordList,
		MyGrabAmount: myGrabAmount,
		MyGrabIsBest: myGrabIsBest,
	})
}

// records 领取记录
func (rp *RedPacket) records(c *wkhttp.Context) {
	packetNo := c.Param("packet_no")

	packet, err := rp.service.redPacketDB.queryByPacketNo(packetNo)
	if err != nil {
		rp.Error("查询红包失败", zap.Error(err))
		c.ResponseError(errors.New("查询红包失败"))
		return
	}
	if packet == nil {
		c.ResponseError(errors.New("红包不存在"))
		return
	}

	records, err := rp.service.redPacketRecordDB.queryByPacketId(packet.Id)
	if err != nil {
		rp.Error("查询领取记录失败", zap.Error(err))
		c.ResponseError(errors.New("查询领取记录失败"))
		return
	}

	list := make([]*redPacketRecordResp, 0, len(records))
	for _, r := range records {
		list = append(list, &redPacketRecordResp{
			UID:       r.UID,
			Amount:    r.Amount,
			IsBest:    r.IsBest == 1,
			CreatedAt: r.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// sentList 发出的红包列表
func (rp *RedPacket) sentList(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	packets, err := rp.service.redPacketDB.queryByUID(uid, page, pageSize)
	if err != nil {
		rp.Error("获取红包列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取红包列表失败"))
		return
	}

	list := make([]*redPacketListResp, 0, len(packets))
	for _, p := range packets {
		list = append(list, &redPacketListResp{
			PacketNo:      p.PacketNo,
			Type:          p.Type,
			TotalAmount:   p.TotalAmount,
			TotalCount:    p.TotalCount,
			ReceivedCount: p.TotalCount - p.RemainCount,
			Remark:        p.Remark,
			Status:        p.Status,
			CreatedAt:     p.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// receivedList 收到的红包列表
func (rp *RedPacket) receivedList(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	records, err := rp.service.redPacketRecordDB.queryByUID(uid, page, pageSize)
	if err != nil {
		rp.Error("获取领取记录失败", zap.Error(err))
		c.ResponseError(errors.New("获取领取记录失败"))
		return
	}

	list := make([]*redPacketReceivedResp, 0, len(records))
	for _, r := range records {
		list = append(list, &redPacketReceivedResp{
			PacketNo:  r.PacketNo,
			Amount:    r.Amount,
			IsBest:    r.IsBest == 1,
			CreatedAt: r.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// ========== 请求/响应结构体 ==========

type redPacketSendReq struct {
	ChannelId   string `json:"channel_id"`   // 频道ID
	ChannelType int    `json:"channel_type"` // 频道类型 1:个人 2:群
	Type        int    `json:"type"`         // 红包类型 1:普通 2:拼手气
	TotalAmount int64  `json:"total_amount"` // 总金额(分)
	TotalCount  int    `json:"total_count"`  // 红包个数
	Remark      string `json:"remark"`       // 祝福语
	PayPassword string `json:"pay_password"` // 支付密码
}

type redPacketSendResp struct {
	PacketNo    string `json:"packet_no"`
	PacketId    int64  `json:"packet_id"`
	TotalAmount int64  `json:"total_amount"`
	TotalCount  int    `json:"total_count"`
	Remark      string `json:"remark"`
	ExpireTime  string `json:"expire_time"`
}

type redPacketGrabReq struct {
	PacketNo string `json:"packet_no"` // 红包编号
}

type redPacketGrabResp struct {
	Amount  int64  `json:"amount"`   // 领取金额(分)
	IsBest  bool   `json:"is_best"`  // 是否手气最佳
	Remark  string `json:"remark"`   // 祝福语
	FromUID string `json:"from_uid"` // 发红包的人
}

type redPacketDetailResp struct {
	PacketNo     string                 `json:"packet_no"`
	UID          string                 `json:"uid"`
	Type         int                    `json:"type"`
	TotalAmount  int64                  `json:"total_amount"`
	TotalCount   int                    `json:"total_count"`
	RemainAmount int64                  `json:"remain_amount"`
	RemainCount  int                    `json:"remain_count"`
	Remark       string                 `json:"remark"`
	Status       int                    `json:"status"`
	ExpireTime   string                 `json:"expire_time"`
	CreatedAt    string                 `json:"created_at"`
	Records      []*redPacketRecordResp `json:"records"`
	MyGrabAmount int64                  `json:"my_grab_amount"`  // 当前用户抢到的金额
	MyGrabIsBest bool                   `json:"my_grab_is_best"` // 当前用户是否手气最佳
}

type redPacketRecordResp struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Amount    int64  `json:"amount"`
	IsBest    bool   `json:"is_best"`
	CreatedAt string `json:"created_at"`
}

type redPacketListResp struct {
	PacketNo      string `json:"packet_no"`
	Type          int    `json:"type"`
	TotalAmount   int64  `json:"total_amount"`
	TotalCount    int    `json:"total_count"`
	ReceivedCount int    `json:"received_count"`
	Remark        string `json:"remark"`
	Status        int    `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type redPacketReceivedResp struct {
	PacketNo  string `json:"packet_no"`
	Amount    int64  `json:"amount"`
	IsBest    bool   `json:"is_best"`
	CreatedAt string `json:"created_at"`
}
