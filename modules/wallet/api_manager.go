package wallet

import (
	"errors"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Manager 钱包管理后台API
type Manager struct {
	ctx     *config.Context
	service *Service
	log.Log
}

// NewManager 创建管理后台API
func NewManager(ctx *config.Context) *Manager {
	return &Manager{
		ctx:     ctx,
		service: NewService(ctx),
		Log:     log.NewTLog("walletManager"),
	}
}

// Route 路由
func (m *Manager) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/manager/wallet", m.ctx.AuthMiddleware(r))
	{
		// 充值管理
		auth.GET("/recharge/list", m.getRechargeList)
		auth.POST("/recharge/:id/approve", m.approveRecharge)
		auth.POST("/recharge/:id/reject", m.rejectRecharge)

		// 提现管理
		auth.GET("/withdraw/list", m.getWithdrawList)
		auth.POST("/withdraw/:id/approve", m.approveWithdraw)
		auth.POST("/withdraw/:id/reject", m.rejectWithdraw)

		// 流水记录
		auth.GET("/records", m.getRecords)

		// 统计数据
		auth.GET("/statistics", m.getStatistics)

		// 余额管理
		auth.POST("/balance/adjust", m.adjustUserBalance)

		// 红包管理
		auth.GET("/redpacket/list", m.getRedPacketList)
		auth.GET("/redpacket/:packet_no", m.getRedPacketDetail)
		auth.GET("/redpacket/:packet_no/records", m.getRedPacketRecords)
		auth.POST("/redpacket/:packet_no/refund", m.refundRedPacket)

		// 转账管理
		auth.GET("/transfer/list", m.getTransferList)
		auth.GET("/transfer/:transfer_no", m.getTransferDetail)

		// 实名认证管理
		auth.GET("/real_name/list", m.getRealNameList)
		auth.POST("/real_name/:id/approve", m.approveRealName)
		auth.POST("/real_name/:id/reject", m.rejectRealName)

		// 用户余额管理
		auth.GET("/balance/:uid", m.getUserBalance)
		auth.GET("/balance/list", m.getBalanceList)

		// 利率管理
		auth.GET("/interest_rate", m.getInterestRate)
		auth.PUT("/interest_rate", m.setInterestRate)

		// 利息统计
		auth.GET("/interest/statistics", m.getInterestStatistics)

		// 利息记录
		auth.GET("/interest/records", m.getInterestRecords)

		// 认证状态管理
		auth.PUT("/real_name/status", m.updateRealNameStatus)

	}
}

// getRechargeList 充值订单列表
func (m *Manager) getRechargeList(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	status := c.Query("status")

	var orders []*rechargeModel
	if status != "" {
		statusInt := parseInt(status, -1)
		orders, err = m.service.rechargeDB.queryByStatus(statusInt, page, pageSize)
	} else {
		orders, err = m.service.rechargeDB.queryAll(page, pageSize)
	}

	if err != nil {
		m.Error("获取充值订单列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取充值订单列表失败"))
		return
	}

	list := make([]*managerRechargeResp, 0, len(orders))
	for _, o := range orders {
		auditedAt := ""
		if o.AuditedAt != nil {
			auditedAt = o.AuditedAt.Format("2006-01-02 15:04:05")
		}
		list = append(list, &managerRechargeResp{
			Id:          o.Id,
			OrderNo:     o.OrderNo,
			UID:         o.UID,
			Amount:      o.Amount,
			Status:      o.Status,
			Remark:      o.Remark,
			AdminUID:    o.AdminUID,
			AdminRemark: o.AdminRemark,
			AuditedAt:   auditedAt,
			CreatedAt:   o.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// approveRecharge 审核通过充值
func (m *Manager) approveRecharge(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	idStr := c.Param("id")
	id := int64(parseInt(idStr, 0))
	if id == 0 {
		c.ResponseError(errors.New("订单ID不能为空"))
		return
	}

	var req auditReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	adminUID := c.GetLoginUID()

	// 查询订单
	order, err := m.service.rechargeDB.queryByID(id)
	if err != nil {
		m.Error("查询订单失败", zap.Error(err))
		c.ResponseError(errors.New("查询订单失败"))
		return
	}
	if order == nil {
		c.ResponseError(errors.New("订单不存在"))
		return
	}
	if order.Status != OrderStatusPending {
		c.ResponseError(errors.New("订单已处理"))
		return
	}

	// 审核通过
	err = m.service.rechargeDB.approve(id, adminUID, req.Remark)
	if err != nil {
		m.Error("审核失败", zap.Error(err))
		c.ResponseError(errors.New("审核失败"))
		return
	}

	// 增加用户余额
	wallet, err := m.service.GetOrCreateWallet(order.UID)
	if err != nil {
		m.Error("获取用户钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取用户钱包失败"))
		return
	}

	err = m.service.walletDB.addBalance(order.UID, order.Amount)
	if err != nil {
		m.Error("增加余额失败", zap.Error(err))
		c.ResponseError(errors.New("增加余额失败"))
		return
	}

	err = m.service.walletDB.addTotalRecharge(order.UID, order.Amount)
	if err != nil {
		m.Error("更新累计充值失败", zap.Error(err))
	}

	// 添加流水记录
	m.service.AddRecord(order.UID, RecordTypeRecharge, order.Amount, wallet.Balance, wallet.Balance+order.Amount, "充值", order.OrderNo, "")

	c.ResponseOK()
}

// rejectRecharge 拒绝充值
func (m *Manager) rejectRecharge(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	idStr := c.Param("id")
	id := int64(parseInt(idStr, 0))
	if id == 0 {
		c.ResponseError(errors.New("订单ID不能为空"))
		return
	}

	var req auditReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	adminUID := c.GetLoginUID()

	err = m.service.rechargeDB.reject(id, adminUID, req.Remark)
	if err != nil {
		m.Error("拒绝失败", zap.Error(err))
		c.ResponseError(errors.New("拒绝失败"))
		return
	}
	c.ResponseOK()
}

// getWithdrawList 提现订单列表
func (m *Manager) getWithdrawList(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	status := c.Query("status")

	var orders []*withdrawModel
	if status != "" {
		statusInt := parseInt(status, -1)
		orders, err = m.service.withdrawDB.queryByStatus(statusInt, page, pageSize)
	} else {
		orders, err = m.service.withdrawDB.queryAll(page, pageSize)
	}

	if err != nil {
		m.Error("获取提现订单列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取提现订单列表失败"))
		return
	}

	list := make([]*managerWithdrawResp, 0, len(orders))
	for _, o := range orders {
		auditedAt := ""
		if o.AuditedAt != nil {
			auditedAt = o.AuditedAt.Format("2006-01-02 15:04:05")
		}
		list = append(list, &managerWithdrawResp{
			Id:          o.Id,
			OrderNo:     o.OrderNo,
			UID:         o.UID,
			Amount:      o.Amount,
			RealName:    o.RealName,
			BankName:    o.BankName,
			BankCard:    o.BankCard,
			Status:      o.Status,
			Remark:      o.Remark,
			AdminUID:    o.AdminUID,
			AdminRemark: o.AdminRemark,
			AuditedAt:   auditedAt,
			CreatedAt:   o.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// approveWithdraw 审核通过提现
func (m *Manager) approveWithdraw(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	idStr := c.Param("id")
	id := int64(parseInt(idStr, 0))
	if id == 0 {
		c.ResponseError(errors.New("订单ID不能为空"))
		return
	}

	var req auditReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	adminUID := c.GetLoginUID()

	// 查询订单
	order, err := m.service.withdrawDB.queryByID(id)
	if err != nil {
		m.Error("查询订单失败", zap.Error(err))
		c.ResponseError(errors.New("查询订单失败"))
		return
	}
	if order == nil {
		c.ResponseError(errors.New("订单不存在"))
		return
	}
	if order.Status != OrderStatusPending {
		c.ResponseError(errors.New("订单已处理"))
		return
	}

	// 审核通过
	err = m.service.withdrawDB.approve(id, adminUID, req.Remark)
	if err != nil {
		m.Error("审核失败", zap.Error(err))
		c.ResponseError(errors.New("审核失败"))
		return
	}

	// 获取钱包并更新
	wallet, err := m.service.GetOrCreateWallet(order.UID)
	if err != nil {
		m.Error("获取用户钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取用户钱包失败"))
		return
	}

	// 解冻并扣除冻结金额
	newFrozen := wallet.Frozen - order.Amount
	if newFrozen < 0 {
		newFrozen = 0
	}
	err = m.service.walletDB.updateFrozen(order.UID, newFrozen)
	if err != nil {
		m.Error("更新冻结金额失败", zap.Error(err))
	}

	err = m.service.walletDB.addTotalWithdraw(order.UID, order.Amount)
	if err != nil {
		m.Error("更新累计提现失败", zap.Error(err))
	}

	// 添加流水记录
	m.service.AddRecord(order.UID, RecordTypeWithdraw, order.Amount, wallet.Balance+order.Amount, wallet.Balance, "提现", order.OrderNo, "")

	c.ResponseOK()
}

// rejectWithdraw 拒绝提现
func (m *Manager) rejectWithdraw(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	idStr := c.Param("id")
	id := int64(parseInt(idStr, 0))
	if id == 0 {
		c.ResponseError(errors.New("订单ID不能为空"))
		return
	}

	var req auditReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	adminUID := c.GetLoginUID()

	// 查询订单
	order, err := m.service.withdrawDB.queryByID(id)
	if err != nil {
		m.Error("查询订单失败", zap.Error(err))
		c.ResponseError(errors.New("查询订单失败"))
		return
	}
	if order == nil {
		c.ResponseError(errors.New("订单不存在"))
		return
	}
	if order.Status != OrderStatusPending {
		c.ResponseError(errors.New("订单已处理"))
		return
	}

	// 拒绝
	err = m.service.withdrawDB.reject(id, adminUID, req.Remark)
	if err != nil {
		m.Error("拒绝失败", zap.Error(err))
		c.ResponseError(errors.New("拒绝失败"))
		return
	}

	// 获取钱包并退回冻结金额
	wallet, err := m.service.GetOrCreateWallet(order.UID)
	if err != nil {
		m.Error("获取用户钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取用户钱包失败"))
		return
	}

	// 退回余额
	err = m.service.walletDB.addBalance(order.UID, order.Amount)
	if err != nil {
		m.Error("退回余额失败", zap.Error(err))
	}

	// 减少冻结金额
	newFrozen := wallet.Frozen - order.Amount
	if newFrozen < 0 {
		newFrozen = 0
	}
	err = m.service.walletDB.updateFrozen(order.UID, newFrozen)
	if err != nil {
		m.Error("更新冻结金额失败", zap.Error(err))
	}

	c.ResponseOK()
}

// getRecords 获取所有流水记录
func (m *Manager) getRecords(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	records, err := m.service.recordDB.queryAll(page, pageSize)
	if err != nil {
		m.Error("获取流水记录失败", zap.Error(err))
		c.ResponseError(errors.New("获取流水记录失败"))
		return
	}

	list := make([]*managerRecordResp, 0, len(records))
	for _, r := range records {
		list = append(list, &managerRecordResp{
			Id:            r.Id,
			UID:           r.UID,
			RecordNo:      r.RecordNo,
			Type:          r.Type,
			TypeName:      m.service.GetRecordTypeName(r.Type),
			Amount:        r.Amount,
			BalanceBefore: r.BalanceBefore,
			BalanceAfter:  r.BalanceAfter,
			Remark:        r.Remark,
			RelatedId:     r.RelatedId,
			RelatedUID:    r.RelatedUID,
			CreatedAt:     r.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// getStatistics 获取统计数据
func (m *Manager) getStatistics(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	pendingRecharge, _ := m.service.rechargeDB.countByStatus(OrderStatusPending)
	pendingWithdraw, _ := m.service.withdrawDB.countByStatus(OrderStatusPending)
	totalRecords, _ := m.service.recordDB.countAll()

	c.Response(&statisticsResp{
		PendingRecharge: pendingRecharge,
		PendingWithdraw: pendingWithdraw,
		TotalRecords:    totalRecords,
	})
}

// ========== 红包管理 ==========

// getRedPacketList 红包列表
func (m *Manager) getRedPacketList(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	status := c.Query("status")
	uid := c.Query("uid")

	var packets []*redPacketModel
	if uid != "" {
		packets, err = m.service.redPacketDB.queryByUID(uid, page, pageSize)
	} else if status != "" {
		statusInt := parseInt(status, -1)
		packets, err = m.service.redPacketDB.queryByStatus(statusInt, page, pageSize)
	} else {
		packets, err = m.service.redPacketDB.queryAll(page, pageSize)
	}

	if err != nil {
		m.Error("获取红包列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取红包列表失败"))
		return
	}

	list := make([]*managerRedPacketResp, 0, len(packets))
	for _, p := range packets {
		list = append(list, &managerRedPacketResp{
			Id:           p.Id,
			PacketNo:     p.PacketNo,
			UID:          p.UID,
			ChannelId:    p.ChannelId,
			ChannelType:  p.ChannelType,
			Type:         p.Type,
			TypeName:     m.getRedPacketTypeName(p.Type),
			TotalAmount:  p.TotalAmount,
			TotalCount:   p.TotalCount,
			RemainAmount: p.RemainAmount,
			RemainCount:  p.RemainCount,
			Remark:       p.Remark,
			Status:       p.Status,
			StatusName:   m.getRedPacketStatusName(p.Status),
			ExpireTime:   p.ExpireTime.Format("2006-01-02 15:04:05"),
			CreatedAt:    p.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// getRedPacketDetail 红包详情
func (m *Manager) getRedPacketDetail(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	packetNo := c.Param("packet_no")
	packet, err := m.service.redPacketDB.queryByPacketNo(packetNo)
	if err != nil {
		m.Error("查询红包失败", zap.Error(err))
		c.ResponseError(errors.New("查询红包失败"))
		return
	}
	if packet == nil {
		c.ResponseError(errors.New("红包不存在"))
		return
	}

	// 查询领取记录
	records, _ := m.service.redPacketRecordDB.queryByPacketId(packet.Id)
	recordList := make([]*managerRedPacketRecordResp, 0, len(records))
	for _, r := range records {
		recordList = append(recordList, &managerRedPacketRecordResp{
			Id:        r.Id,
			UID:       r.UID,
			Amount:    r.Amount,
			IsBest:    r.IsBest == 1,
			CreatedAt: r.CreatedAt.String(),
		})
	}

	c.Response(&managerRedPacketDetailResp{
		Id:           packet.Id,
		PacketNo:     packet.PacketNo,
		UID:          packet.UID,
		ChannelId:    packet.ChannelId,
		ChannelType:  packet.ChannelType,
		Type:         packet.Type,
		TypeName:     m.getRedPacketTypeName(packet.Type),
		TotalAmount:  packet.TotalAmount,
		TotalCount:   packet.TotalCount,
		RemainAmount: packet.RemainAmount,
		RemainCount:  packet.RemainCount,
		Remark:       packet.Remark,
		Status:       packet.Status,
		StatusName:   m.getRedPacketStatusName(packet.Status),
		ExpireTime:   packet.ExpireTime.Format("2006-01-02 15:04:05"),
		CreatedAt:    packet.CreatedAt.String(),
		Records:      recordList,
	})
}

// getRedPacketRecords 红包领取记录
func (m *Manager) getRedPacketRecords(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	packetNo := c.Param("packet_no")
	packet, err := m.service.redPacketDB.queryByPacketNo(packetNo)
	if err != nil || packet == nil {
		c.ResponseError(errors.New("红包不存在"))
		return
	}

	records, err := m.service.redPacketRecordDB.queryByPacketId(packet.Id)
	if err != nil {
		m.Error("查询领取记录失败", zap.Error(err))
		c.ResponseError(errors.New("查询领取记录失败"))
		return
	}

	list := make([]*managerRedPacketRecordResp, 0, len(records))
	for _, r := range records {
		list = append(list, &managerRedPacketRecordResp{
			Id:        r.Id,
			UID:       r.UID,
			Amount:    r.Amount,
			IsBest:    r.IsBest == 1,
			CreatedAt: r.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// refundRedPacket 手动退回红包
func (m *Manager) refundRedPacket(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	packetNo := c.Param("packet_no")
	packet, err := m.service.redPacketDB.queryByPacketNo(packetNo)
	if err != nil || packet == nil {
		c.ResponseError(errors.New("红包不存在"))
		return
	}

	if packet.Status != RedPacketStatusActive {
		c.ResponseError(errors.New("红包已处理，无法退回"))
		return
	}

	if packet.RemainAmount <= 0 {
		c.ResponseError(errors.New("红包无剩余金额"))
		return
	}

	// 退回剩余金额
	wallet, err := m.service.GetOrCreateWallet(packet.UID)
	if err != nil {
		c.ResponseError(errors.New("获取用户钱包失败"))
		return
	}

	err = m.service.walletDB.addBalance(packet.UID, packet.RemainAmount)
	if err != nil {
		c.ResponseError(errors.New("退回余额失败"))
		return
	}

	m.service.AddRecord(packet.UID, RecordTypeRedPacketRefund, packet.RemainAmount, wallet.Balance, wallet.Balance+packet.RemainAmount, "管理员手动退回红包", packet.PacketNo, "")
	m.service.redPacketDB.updateStatus(packetNo, RedPacketStatusExpired)

	c.ResponseOK()
}

// ========== 转账管理 ==========

// getTransferList 转账列表
func (m *Manager) getTransferList(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	status := c.Query("status")
	uid := c.Query("uid")

	var transfers []*transferModel
	if uid != "" {
		transfers, err = m.service.transferDB.queryByFromUID(uid, page, pageSize)
	} else if status != "" {
		statusInt := parseInt(status, -1)
		transfers, err = m.service.transferDB.queryByStatus(statusInt, page, pageSize)
	} else {
		transfers, err = m.service.transferDB.queryAll(page, pageSize)
	}

	if err != nil {
		m.Error("获取转账列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取转账列表失败"))
		return
	}

	list := make([]*managerTransferResp, 0, len(transfers))
	for _, t := range transfers {
		receivedAt := ""
		if t.ReceivedAt != nil {
			receivedAt = t.ReceivedAt.Format("2006-01-02 15:04:05")
		}
		list = append(list, &managerTransferResp{
			Id:         t.Id,
			TransferNo: t.TransferNo,
			FromUID:    t.FromUID,
			ToUID:      t.ToUID,
			Amount:     t.Amount,
			Remark:     t.Remark,
			Status:     t.Status,
			StatusName: m.getTransferStatusName(t.Status),
			ExpireTime: t.ExpireTime.Format("2006-01-02 15:04:05"),
			ReceivedAt: receivedAt,
			CreatedAt:  t.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// getTransferDetail 转账详情
func (m *Manager) getTransferDetail(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	transferNo := c.Param("transfer_no")
	transfer, err := m.service.transferDB.queryByTransferNo(transferNo)
	if err != nil || transfer == nil {
		c.ResponseError(errors.New("转账记录不存在"))
		return
	}

	receivedAt := ""
	if transfer.ReceivedAt != nil {
		receivedAt = transfer.ReceivedAt.Format("2006-01-02 15:04:05")
	}

	c.Response(&managerTransferResp{
		Id:         transfer.Id,
		TransferNo: transfer.TransferNo,
		FromUID:    transfer.FromUID,
		ToUID:      transfer.ToUID,
		Amount:     transfer.Amount,
		Remark:     transfer.Remark,
		Status:     transfer.Status,
		StatusName: m.getTransferStatusName(transfer.Status),
		ExpireTime: transfer.ExpireTime.Format("2006-01-02 15:04:05"),
		ReceivedAt: receivedAt,
		CreatedAt:  transfer.CreatedAt.String(),
	})
}

// ========== 实名认证管理 ==========

// getRealNameList 实名认证列表
func (m *Manager) getRealNameList(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	status := c.Query("status")

	rows := m.service.ctx.DB().Select("id, uid, real_name, id_card, phone, status, remark, admin_uid, audited_at, created_at").From("wallet_real_name")
	if status != "" {
		statusInt := parseInt(status, -1)
		rows = rows.Where("status=?", statusInt)
	}
	rows = rows.OrderBy("created_at DESC").Limit(uint64(pageSize)).Offset(uint64((page - 1) * pageSize))

	result, err := rows.Rows()
	if err != nil {
		m.Error("查询实名认证列表失败", zap.Error(err))
		c.ResponseError(errors.New("查询实名认证列表失败"))
		return
	}
	defer result.Close()

	var list []*managerRealNameResp
	for result.Next() {
		var id int64
		var uid, realName, idCard, phone, remark, adminUID string
		var status int
		var auditedAt *time.Time
		var createdAt time.Time

		err := result.Scan(&id, &uid, &realName, &idCard, &phone, &status, &remark, &adminUID, &auditedAt, &createdAt)
		if err != nil {
			m.Error("扫描实名认证数据失败", zap.Error(err))
			continue
		}

		auditedAtStr := ""
		if auditedAt != nil {
			auditedAtStr = auditedAt.Format("2006-01-02 15:04:05")
		}

		list = append(list, &managerRealNameResp{
			Id:         id,
			UID:        uid,
			RealName:   realName,
			IdCard:     idCard,
			Phone:      phone,
			Status:     status,
			StatusName: m.getRealNameStatusName(status),
			Remark:     remark,
			AdminUID:   adminUID,
			AuditedAt:  auditedAtStr,
			CreatedAt:  createdAt.String(),
		})
	}

	c.Response(list)
}

// approveRealName 审核通过实名认证
func (m *Manager) approveRealName(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	idStr := c.Param("id")
	id := int64(parseInt(idStr, 0))
	if id == 0 {
		c.ResponseError(errors.New("认证ID不能为空"))
		return
	}

	var req auditReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	adminUID := c.GetLoginUID()

	// 开启事务
	tx, err := m.service.ctx.DB().Begin()
	if err != nil {
		c.ResponseError(errors.New("系统繁忙，请重试"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 查询认证记录
	var uid, realName, idCard string
	err = tx.QueryRow("SELECT uid, real_name, id_card FROM wallet_real_name WHERE id=? AND status=?", id, 0).Scan(&uid, &realName, &idCard)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("认证记录不存在或已处理"))
		return
	}

	// 更新实名认证状态
	_, err = tx.UpdateBySql("UPDATE wallet_real_name SET status=1, admin_uid=?, remark=?, audited_at=? WHERE id=?", adminUID, req.Remark, time.Now(), id).Exec()
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("更新认证状态失败"))
		return
	}

	// 更新钱包表中的实名认证状态
	_, err = tx.UpdateBySql("UPDATE wallet SET real_name_status=2, real_name=?, id_card=? WHERE uid=?", realName, idCard, uid).Exec()
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("更新钱包认证状态失败"))
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		c.ResponseError(errors.New("审核失败，请重试"))
		return
	}

	c.ResponseOK()
}

// rejectRealName 拒绝实名认证
func (m *Manager) rejectRealName(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	idStr := c.Param("id")
	id := int64(parseInt(idStr, 0))
	if id == 0 {
		c.ResponseError(errors.New("认证ID不能为空"))
		return
	}

	var req auditReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	adminUID := c.GetLoginUID()

	// 开启事务
	tx, err := m.service.ctx.DB().Begin()
	if err != nil {
		c.ResponseError(errors.New("系统繁忙，请重试"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 查询认证记录
	var uid string
	err = tx.QueryRow("SELECT uid FROM wallet_real_name WHERE id=? AND status=?", id, 0).Scan(&uid)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("认证记录不存在或已处理"))
		return
	}

	// 更新实名认证状态
	_, err = tx.UpdateBySql("UPDATE wallet_real_name SET status=2, admin_uid=?, remark=?, audited_at=? WHERE id=?", adminUID, req.Remark, time.Now(), id).Exec()
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("更新认证状态失败"))
		return
	}

	// 更新钱包表中的实名认证状态
	_, err = tx.UpdateBySql("UPDATE wallet SET real_name_status=3 WHERE uid=?", uid).Exec()
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("更新钱包认证状态失败"))
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		c.ResponseError(errors.New("审核失败，请重试"))
		return
	}

	c.ResponseOK()
}

// getRealNameStatusName 获取实名认证状态名称
func (m *Manager) getRealNameStatusName(status int) string {
	switch status {
	case 0:
		return "待审核"
	case 1:
		return "已认证"
	case 2:
		return "认证失败"
	default:
		return "未知"
	}
}

// getUserBalance 获取用户实时余额
func (m *Manager) getUserBalance(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.ResponseError(errors.New("用户UID不能为空"))
		return
	}

	// 查询用户钱包信息
	wallet, err := m.service.GetOrCreateWallet(uid)
	if err != nil {
		m.Error("获取用户钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取用户钱包失败"))
		return
	}

	// 返回用户余额信息
	c.Response(&managerUserBalanceResp{
		UID:            wallet.UID,
		Balance:        wallet.Balance,
		Frozen:         wallet.Frozen,
		TotalRecharge:  wallet.TotalRecharge,
		TotalWithdraw:  wallet.TotalWithdraw,
		Interest:       wallet.Interest,
		TodayInterest:  wallet.TodayInterest,
		InterestRate:   wallet.InterestRate,
		RealNameStatus: wallet.RealNameStatus,
		RealName:       wallet.RealName,
		IdCard:         wallet.IdCard,
		HasPayPwd:      wallet.PayPassword != "",
		Status:         wallet.Status,
		CreatedAt:      wallet.CreatedAt.String(),
		UpdatedAt:      wallet.UpdatedAt.String(),
	})
}

// getBalanceList 获取有钱的用户余额列表
func (m *Manager) getBalanceList(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	// 查询有钱的用户钱包
	rows := m.service.ctx.DB().Select("id, uid, balance, frozen, total_recharge, total_withdraw, interest, today_interest, interest_rate, real_name_status, real_name, id_card, pay_password, status, created_at, updated_at").From("wallet").Where("balance > 0").OrderBy("balance DESC").Limit(uint64(pageSize)).Offset(uint64((page - 1) * pageSize))

	result, err := rows.Rows()
	if err != nil {
		m.Error("查询钱包列表失败", zap.Error(err))
		c.ResponseError(errors.New("查询钱包列表失败"))
		return
	}
	defer result.Close()

	var list []*managerUserBalanceResp
	for result.Next() {
		var id int64
		var uid, realName, idCard, payPassword string
		var balance, frozen, totalRecharge, totalWithdraw, interest, todayInterest int64
		var interestRate float64
		var realNameStatus, status int
		var createdAt, updatedAt time.Time

		err := result.Scan(&id, &uid, &balance, &frozen, &totalRecharge, &totalWithdraw, &interest, &todayInterest, &interestRate, &realNameStatus, &realName, &idCard, &payPassword, &status, &createdAt, &updatedAt)
		if err != nil {
			m.Error("扫描钱包数据失败", zap.Error(err))
			continue
		}

		list = append(list, &managerUserBalanceResp{
			UID:            uid,
			Balance:        balance,
			Frozen:         frozen,
			TotalRecharge:  totalRecharge,
			TotalWithdraw:  totalWithdraw,
			Interest:       interest,
			TodayInterest:  todayInterest,
			InterestRate:   interestRate,
			RealNameStatus: realNameStatus,
			RealName:       realName,
			IdCard:         idCard,
			HasPayPwd:      payPassword != "",
			Status:         status,
			CreatedAt:      createdAt.String(),
			UpdatedAt:      updatedAt.String(),
		})
	}

	c.Response(list)
}

// getInterestRate 获取当前利率设置
func (m *Manager) getInterestRate(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	// 调用服务方法获取利率
	interestRate, err := m.service.GetInterestRate()
	if err != nil {
		m.Error("获取利率设置失败", zap.Error(err))
		c.ResponseError(errors.New("获取利率设置失败"))
		return
	}

	c.Response(map[string]interface{}{
		"interest_rate": interestRate,
	})
}

// setInterestRate 设置利率
func (m *Manager) setInterestRate(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	type interestRateReq struct {
		InterestRate float64 `json:"interest_rate"` // 年化利率(%)
	}

	var req interestRateReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if req.InterestRate < 0 {
		c.ResponseError(errors.New("利率不能为负数"))
		return
	}

	// 调用服务方法设置利率
	err = m.service.SetInterestRate(req.InterestRate)
	if err != nil {
		m.Error("设置利率失败", zap.Error(err))
		c.ResponseError(errors.New("设置利率失败"))
		return
	}

	c.Response(map[string]interface{}{
		"interest_rate": req.InterestRate,
		"message":       "利率设置成功",
	})
}

// getInterestStatistics 获取利息统计数据
func (m *Manager) getInterestStatistics(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	// 查询总发放次数和总发放金额
	var totalCount int64
	var totalAmount int64
	err = m.service.ctx.DB().QueryRow("SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM wallet_record WHERE type=?", RecordTypeInterest).Scan(&totalCount, &totalAmount)
	if err != nil {
		m.Error("查询总利息发放统计失败", zap.Error(err))
		totalCount = 0
		totalAmount = 0
	}

	// 查询今日发放次数和今日发放金额
	var todayCount int64
	var todayAmount int64
	today := time.Now().Format("2006-01-02")
	err = m.service.ctx.DB().QueryRow("SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM wallet_record WHERE type=? AND DATE(created_at)=?", RecordTypeInterest, today).Scan(&todayCount, &todayAmount)
	if err != nil {
		m.Error("查询今日利息发放统计失败", zap.Error(err))
		todayCount = 0
		todayAmount = 0
	}

	c.Response(map[string]interface{}{
		"total_count":  totalCount,
		"total_amount": totalAmount,
		"today_count":  todayCount,
		"today_amount": todayAmount,
	})
}

// adjustUserBalance 调整用户钱包余额
func (m *Manager) adjustUserBalance(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	type adjustBalanceReq struct {
		UID    string `json:"uid"`    // 用户UID
		Amount int64  `json:"amount"` // 调整金额(分)，正数增加，负数减少
		Remark string `json:"remark"` // 调整备注
	}

	var req adjustBalanceReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if req.UID == "" {
		c.ResponseError(errors.New("用户UID不能为空"))
		return
	}

	if req.Amount == 0 {
		c.ResponseError(errors.New("调整金额不能为0"))
		return
	}

	// 获取用户钱包
	wallet, err := m.service.GetOrCreateWallet(req.UID)
	if err != nil {
		m.Error("获取用户钱包失败", zap.String("uid", req.UID), zap.Error(err))
		c.ResponseError(errors.New("获取用户钱包失败"))
		return
	}

	// 检查余额是否足够（如果是减少余额）
	if req.Amount < 0 && wallet.Balance < -req.Amount {
		c.ResponseError(errors.New("用户余额不足"))
		return
	}

	// 计算新余额
	newBalance := wallet.Balance + req.Amount

	// 开始事务
	tx, err := m.service.ctx.DB().Begin()
	if err != nil {
		m.Error("开启事务失败", zap.Error(err))
		c.ResponseError(errors.New("开启事务失败"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 更新钱包余额
	_, err = tx.UpdateBySql("UPDATE wallet SET balance = ? WHERE uid = ?", newBalance, req.UID).Exec()
	if err != nil {
		m.Error("更新钱包余额失败", zap.String("uid", req.UID), zap.Error(err))
		c.ResponseError(errors.New("更新钱包余额失败"))
		return
	}

	// 添加余额调整流水记录
	recordNo := m.service.GenerateRecordNo()
	recordType := RecordTypeBalanceAdjust // 使用专门的余额调整类型

	_, err = tx.InsertInto("wallet_record").Columns(
		"uid", "record_no", "type", "amount", "balance_before", "balance_after", "remark", "related_id", "related_uid",
	).Values(
		req.UID, recordNo, recordType, req.Amount, wallet.Balance, newBalance, req.Remark, "", "",
	).Exec()
	if err != nil {
		m.Error("添加余额调整记录失败", zap.String("uid", req.UID), zap.Error(err))
		c.ResponseError(errors.New("添加余额调整记录失败"))
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		m.Error("提交事务失败", zap.Error(err))
		c.ResponseError(errors.New("提交事务失败"))
		return
	}

	m.Info("调整用户余额成功", zap.String("uid", req.UID), zap.Int64("amount", req.Amount), zap.Int64("new_balance", newBalance))

	c.Response(map[string]interface{}{
		"uid":         req.UID,
		"amount":      req.Amount,
		"old_balance": wallet.Balance,
		"new_balance": newBalance,
		"remark":      req.Remark,
	})
}

// abs 获取绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// getInterestRecords 获取利息记录
func (m *Manager) getInterestRecords(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}

	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	// 查询利息记录
	rows := m.service.ctx.DB().Select("id, uid, record_no, amount, balance_before, balance_after, remark, related_id, related_uid, created_at").
		From("wallet_record").
		Where("type=?", RecordTypeInterest).
		OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize))

	result, err := rows.Rows()
	if err != nil {
		m.Error("查询利息记录失败", zap.Error(err))
		c.ResponseError(errors.New("查询利息记录失败"))
		return
	}
	defer result.Close()

	var list []*managerInterestRecordResp
	for result.Next() {
		var id int64
		var uid, recordNo, remark, relatedId, relatedUID string
		var amount, balanceBefore, balanceAfter int64
		var createdAt time.Time

		err := result.Scan(&id, &uid, &recordNo, &amount, &balanceBefore, &balanceAfter, &remark, &relatedId, &relatedUID, &createdAt)
		if err != nil {
			m.Error("扫描利息记录数据失败", zap.Error(err))
			continue
		}

		list = append(list, &managerInterestRecordResp{
			Id:            id,
			UID:           uid,
			RecordNo:      recordNo,
			Amount:        amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Remark:        remark,
			RelatedId:     relatedId,
			RelatedUID:    relatedUID,
			CreatedAt:     createdAt.String(),
		})
	}

	// 查询总记录数
	var total int64
	err = m.service.ctx.DB().QueryRow("SELECT COUNT(*) FROM wallet_record WHERE type=?", RecordTypeInterest).Scan(&total)
	if err != nil {
		m.Error("查询利息记录总数失败", zap.Error(err))
		total = 0
	}

	c.Response(map[string]interface{}{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// updateRealNameStatus 调整用户钱包认证状态
func (m *Manager) updateRealNameStatus(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	type updateRealNameStatusReq struct {
		UID            string `json:"uid"`              // 用户UID
		RealNameStatus int    `json:"real_name_status"` // 新的认证状态: 1-未认证, 2-已认证, 3-认证失败
		RealName       string `json:"real_name"`        // 真实姓名
		IdCard         string `json:"id_card"`          // 身份证号
	}

	var req updateRealNameStatusReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if req.UID == "" {
		c.ResponseError(errors.New("用户UID不能为空"))
		return
	}

	if req.RealNameStatus < 1 || req.RealNameStatus > 3 {
		c.ResponseError(errors.New("认证状态值无效，应为1-未认证, 2-已认证, 3-认证失败"))
		return
	}

	// 开启事务
	tx, err := m.service.ctx.DB().Begin()
	if err != nil {
		c.ResponseError(errors.New("系统繁忙，请重试"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 检查钱包记录是否存在
	var walletExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM wallet WHERE uid=?)", req.UID).Scan(&walletExists)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("查询钱包记录失败"))
		return
	}

	// 如果钱包记录不存在，创建一个
	if !walletExists {
		_, err = tx.Exec(
			"INSERT INTO wallet (uid, real_name_status, real_name, id_card, balance, frozen, interest, today_interest, total_recharge, total_withdraw, interest_rate, status, created_at, updated_at) VALUES (?, ?, ?, ?, 0, 0, 0, 0, 0, 0, 0, 1, NOW(), NOW())",
			req.UID, req.RealNameStatus, req.RealName, req.IdCard,
		)
		if err != nil {
			tx.Rollback()
			c.ResponseError(errors.New("创建钱包记录失败"))
			return
		}
	} else {
		// 查询当前认证状态
		var currentStatus int
		err = tx.QueryRow("SELECT real_name_status FROM wallet WHERE uid=?", req.UID).Scan(&currentStatus)
		if err != nil {
			tx.Rollback()
			c.ResponseError(errors.New("查询当前认证状态失败"))
			return
		}

		// 只有当有字段需要更新时才执行更新操作
		needUpdate := false
		updateFields := make(map[string]interface{})

		// 只有当认证状态发生变化时才更新
		if currentStatus != req.RealNameStatus {
			updateFields["real_name_status"] = req.RealNameStatus
			needUpdate = true
		}

		// 只有当提供了真实姓名时才更新
		if req.RealName != "" {
			updateFields["real_name"] = req.RealName
			needUpdate = true
		}

		// 只有当提供了身份证号时才更新
		if req.IdCard != "" {
			updateFields["id_card"] = req.IdCard
			needUpdate = true
		}

		// 如果需要更新，执行更新操作
		if needUpdate {
			// 构建更新语句
			updateStmt := tx.Update("wallet")

			// 添加更新字段
			for field, value := range updateFields {
				updateStmt = updateStmt.Set(field, value)
			}

			// 执行更新
			_, err = updateStmt.Where("uid=?", req.UID).Exec()
			if err != nil {
				tx.Rollback()
				c.ResponseError(errors.New("更新钱包认证状态失败"))
				return
			}
		}
	}

	// 检查是否存在实名认证记录
	var count int64
	err = tx.QueryRow("SELECT COUNT(*) FROM wallet_real_name WHERE uid=?", req.UID).Scan(&count)
	if err != nil {
		tx.Rollback()
		c.ResponseError(errors.New("查询实名认证记录失败"))
		return
	}

	// 如果存在实名认证记录，更新其状态
	if count > 0 {
		authStatus := 0
		if req.RealNameStatus == 2 {
			authStatus = 1 // 已认证
		} else if req.RealNameStatus == 3 {
			authStatus = 2 // 认证失败
		}
		if authStatus > 0 {
			_, err = tx.UpdateBySql("UPDATE wallet_real_name SET status=?, updated_at=NOW() WHERE uid=?", authStatus, req.UID).Exec()
			if err != nil {
				tx.Rollback()
				c.ResponseError(errors.New("更新实名认证记录失败"))
				return
			}
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		c.ResponseError(errors.New("操作失败，请重试"))
		return
	}

	c.Response(map[string]interface{}{
		"uid":              req.UID,
		"real_name_status": req.RealNameStatus,
		"message":          "认证状态调整成功",
	})
}

// ========== 辅助方法 ==========

func (m *Manager) getRedPacketTypeName(t int) string {
	if t == RedPacketTypeLucky {
		return "拼手气红包"
	}
	return "普通红包"
}

func (m *Manager) getRedPacketStatusName(s int) string {
	switch s {
	case RedPacketStatusActive:
		return "进行中"
	case RedPacketStatusFinished:
		return "已抢完"
	case RedPacketStatusExpired:
		return "已过期退回"
	default:
		return "未知"
	}
}

func (m *Manager) getTransferStatusName(s int) string {
	switch s {
	case TransferStatusPending:
		return "待领取"
	case TransferStatusReceived:
		return "已领取"
	case TransferStatusRefunded:
		return "已退回"
	case TransferStatusExpired:
		return "已过期退回"
	default:
		return "未知"
	}
}

// ========== 请求/响应结构体 ==========

type auditReq struct {
	Remark string `json:"remark"` // 审核备注
}

type managerRechargeResp struct {
	Id          int64  `json:"id"`
	OrderNo     string `json:"order_no"`
	UID         string `json:"uid"`
	Amount      int64  `json:"amount"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	AdminUID    string `json:"admin_uid"`
	AdminRemark string `json:"admin_remark"`
	AuditedAt   string `json:"audited_at"`
	CreatedAt   string `json:"created_at"`
}

type managerWithdrawResp struct {
	Id          int64  `json:"id"`
	OrderNo     string `json:"order_no"`
	UID         string `json:"uid"`
	Amount      int64  `json:"amount"`
	RealName    string `json:"real_name"`
	BankName    string `json:"bank_name"`
	BankCard    string `json:"bank_card"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	AdminUID    string `json:"admin_uid"`
	AdminRemark string `json:"admin_remark"`
	AuditedAt   string `json:"audited_at"`
	CreatedAt   string `json:"created_at"`
}

type managerRecordResp struct {
	Id            int64  `json:"id"`
	UID           string `json:"uid"`
	RecordNo      string `json:"record_no"`
	Type          int    `json:"type"`
	TypeName      string `json:"type_name"`
	Amount        int64  `json:"amount"`
	BalanceBefore int64  `json:"balance_before"`
	BalanceAfter  int64  `json:"balance_after"`
	Remark        string `json:"remark"`
	RelatedId     string `json:"related_id"`
	RelatedUID    string `json:"related_uid"`
	CreatedAt     string `json:"created_at"`
}

type statisticsResp struct {
	PendingRecharge int64 `json:"pending_recharge"` // 待审核充值订单数
	PendingWithdraw int64 `json:"pending_withdraw"` // 待审核提现订单数
	TotalRecords    int64 `json:"total_records"`    // 总流水数
}

type managerRedPacketResp struct {
	Id           int64  `json:"id"`
	PacketNo     string `json:"packet_no"`
	UID          string `json:"uid"`
	ChannelId    string `json:"channel_id"`
	ChannelType  int    `json:"channel_type"`
	Type         int    `json:"type"`
	TypeName     string `json:"type_name"`
	TotalAmount  int64  `json:"total_amount"`
	TotalCount   int    `json:"total_count"`
	RemainAmount int64  `json:"remain_amount"`
	RemainCount  int    `json:"remain_count"`
	Remark       string `json:"remark"`
	Status       int    `json:"status"`
	StatusName   string `json:"status_name"`
	ExpireTime   string `json:"expire_time"`
	CreatedAt    string `json:"created_at"`
}

type managerRedPacketDetailResp struct {
	Id           int64                         `json:"id"`
	PacketNo     string                        `json:"packet_no"`
	UID          string                        `json:"uid"`
	ChannelId    string                        `json:"channel_id"`
	ChannelType  int                           `json:"channel_type"`
	Type         int                           `json:"type"`
	TypeName     string                        `json:"type_name"`
	TotalAmount  int64                         `json:"total_amount"`
	TotalCount   int                           `json:"total_count"`
	RemainAmount int64                         `json:"remain_amount"`
	RemainCount  int                           `json:"remain_count"`
	Remark       string                        `json:"remark"`
	Status       int                           `json:"status"`
	StatusName   string                        `json:"status_name"`
	ExpireTime   string                        `json:"expire_time"`
	CreatedAt    string                        `json:"created_at"`
	Records      []*managerRedPacketRecordResp `json:"records"`
}

type managerRedPacketRecordResp struct {
	Id        int64  `json:"id"`
	UID       string `json:"uid"`
	Amount    int64  `json:"amount"`
	IsBest    bool   `json:"is_best"`
	CreatedAt string `json:"created_at"`
}

type managerTransferResp struct {
	Id         int64  `json:"id"`
	TransferNo string `json:"transfer_no"`
	FromUID    string `json:"from_uid"`
	ToUID      string `json:"to_uid"`
	Amount     int64  `json:"amount"`
	Remark     string `json:"remark"`
	Status     int    `json:"status"`
	StatusName string `json:"status_name"`
	ExpireTime string `json:"expire_time"`
	ReceivedAt string `json:"received_at"`
	CreatedAt  string `json:"created_at"`
}

type managerRealNameResp struct {
	Id         int64  `json:"id"`
	UID        string `json:"uid"`
	RealName   string `json:"real_name"`
	IdCard     string `json:"id_card"`
	Phone      string `json:"phone"`
	Status     int    `json:"status"`
	StatusName string `json:"status_name"`
	Remark     string `json:"remark"`
	AdminUID   string `json:"admin_uid"`
	AuditedAt  string `json:"audited_at"`
	CreatedAt  string `json:"created_at"`
}

type managerUserBalanceResp struct {
	UID            string  `json:"uid"`              // 用户UID
	Balance        int64   `json:"balance"`          // 可用余额(分)
	Frozen         int64   `json:"frozen"`           // 冻结金额(分)
	TotalRecharge  int64   `json:"total_recharge"`   // 累计充值(分)
	TotalWithdraw  int64   `json:"total_withdraw"`   // 累计提现(分)
	Interest       int64   `json:"interest"`         // 累计利息(分)
	TodayInterest  int64   `json:"today_interest"`   // 今日利息(分)
	InterestRate   float64 `json:"interest_rate"`    // 年化利率(%)
	RealNameStatus int     `json:"real_name_status"` // 实名认证状态
	RealName       string  `json:"real_name"`        // 真实姓名
	IdCard         string  `json:"id_card"`          // 身份证号
	HasPayPwd      bool    `json:"has_pay_pwd"`      // 是否设置支付密码
	Status         int     `json:"status"`           // 钱包状态
	CreatedAt      string  `json:"created_at"`       // 创建时间
	UpdatedAt      string  `json:"updated_at"`       // 更新时间
}

type managerInterestRecordResp struct {
	Id            int64  `json:"id"`             // 记录ID
	UID           string `json:"uid"`            // 用户UID
	RecordNo      string `json:"record_no"`      // 流水号
	Amount        int64  `json:"amount"`         // 利息金额(分)
	BalanceBefore int64  `json:"balance_before"` // 变动前余额(分)
	BalanceAfter  int64  `json:"balance_after"`  // 变动后余额(分)
	Remark        string `json:"remark"`         // 备注
	RelatedId     string `json:"related_id"`     // 关联ID
	RelatedUID    string `json:"related_uid"`    // 关联用户UID
	CreatedAt     string `json:"created_at"`     // 创建时间
}
