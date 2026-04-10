package wallet

import (
	"errors"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Transfer 转账API
type Transfer struct {
	ctx     *config.Context
	service *Service
	log.Log
}

// NewTransfer 创建转账API
func NewTransfer(ctx *config.Context) *Transfer {
	return &Transfer{
		ctx:     ctx,
		service: NewService(ctx),
		Log:     log.NewTLog("transferAPI"),
	}
}

// Route 路由
func (t *Transfer) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/wallet/transfer", t.ctx.AuthMiddleware(r))
	{
		auth.POST("/send", t.send)                    // 发起转账
		auth.POST("/:transfer_no/receive", t.receive) // 接收转账
		auth.POST("/:transfer_no/refund", t.refund)   // 退回转账
		auth.GET("/:transfer_no", t.detail)           // 转账详情
		auth.GET("/sent", t.sentList)                 // 转出列表
		auth.GET("/received", t.receivedList)         // 收到列表
	}
}

// send 发起转账
func (t *Transfer) send(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req transferSendReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	if req.ToUID == "" {
		c.ResponseError(errors.New("接收人不能为空"))
		return
	}
	if req.ToUID == uid {
		c.ResponseError(errors.New("不能转账给自己"))
		return
	}
	if req.Amount <= 0 {
		c.ResponseError(errors.New("转账金额必须大于0"))
		return
	}

	// 验证支付密码
	wallet, err := t.service.GetOrCreateWallet(uid)
	if err != nil {
		t.Error("获取钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取钱包失败"))
		return
	}
	if wallet.PayPassword == "" {
		c.ResponseError(errors.New("请先设置支付密码"))
		return
	}
	pwdHash := t.service.EncryptPassword(req.PayPassword, wallet.PayPasswordSalt)
	if pwdHash != wallet.PayPassword {
		c.ResponseError(errors.New("支付密码错误"))
		return
	}

	// 检查余额
	if wallet.Balance < req.Amount {
		c.ResponseError(errors.New("余额不足"))
		return
	}

	// 扣除余额
	err = t.service.walletDB.subBalance(uid, req.Amount)
	if err != nil {
		t.Error("扣除余额失败", zap.Error(err))
		c.ResponseError(errors.New("扣除余额失败"))
		return
	}

	// 创建转账记录
	transferNo := t.service.GenerateOrderNo("TF")
	expireTime := time.Now().Add(24 * time.Hour) // 24小时过期

	transfer := &transferModel{
		TransferNo: transferNo,
		FromUID:    uid,
		ToUID:      req.ToUID,
		Amount:     req.Amount,
		Remark:     req.Remark,
		Status:     TransferStatusPending,
		ExpireTime: expireTime,
	}

	_, err = t.service.transferDB.insert(transfer)
	if err != nil {
		// 回滚余额
		t.service.walletDB.addBalance(uid, req.Amount)
		t.Error("创建转账记录失败", zap.Error(err))
		c.ResponseError(errors.New("创建转账记录失败"))
		return
	}

	// 添加流水记录
	t.service.AddRecord(uid, RecordTypeTransferOut, req.Amount, wallet.Balance, wallet.Balance-req.Amount, req.Remark, transferNo, req.ToUID)

	c.Response(&transferSendResp{
		TransferNo: transferNo,
		Amount:     req.Amount,
		ToUID:      req.ToUID,
		Remark:     req.Remark,
		ExpireTime: expireTime.Format("2006-01-02 15:04:05"),
	})
}

// receive 接收转账
func (t *Transfer) receive(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	transferNo := c.Param("transfer_no")
	if transferNo == "" {
		c.ResponseError(errors.New("转账编号不能为空"))
		return
	}

	// 查询转账记录
	transfer, err := t.service.transferDB.queryByTransferNo(transferNo)
	if err != nil {
		t.Error("查询转账记录失败", zap.Error(err))
		c.ResponseError(errors.New("查询转账记录失败"))
		return
	}
	if transfer == nil {
		c.ResponseError(errors.New("转账记录不存在"))
		return
	}

	// 验证接收人
	if transfer.ToUID != uid {
		c.ResponseError(errors.New("您不是该转账的接收人"))
		return
	}

	// 验证状态
	if transfer.Status != TransferStatusPending {
		c.ResponseError(errors.New("转账已被处理"))
		return
	}

	// 检查是否过期
	if time.Now().After(transfer.ExpireTime) {
		c.ResponseError(errors.New("转账已过期"))
		return
	}

	// 更新状态
	err = t.service.transferDB.receive(transferNo)
	if err != nil {
		t.Error("接收转账失败", zap.Error(err))
		c.ResponseError(errors.New("接收转账失败"))
		return
	}

	// 增加余额
	wallet, _ := t.service.GetOrCreateWallet(uid)
	err = t.service.walletDB.addBalance(uid, transfer.Amount)
	if err != nil {
		t.Error("增加余额失败", zap.Error(err))
		c.ResponseError(errors.New("增加余额失败"))
		return
	}

	// 添加流水记录
	t.service.AddRecord(uid, RecordTypeTransferIn, transfer.Amount, wallet.Balance, wallet.Balance+transfer.Amount, transfer.Remark, transferNo, transfer.FromUID)

	c.ResponseOK()
}

// refund 退回转账
func (t *Transfer) refund(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	transferNo := c.Param("transfer_no")
	if transferNo == "" {
		c.ResponseError(errors.New("转账编号不能为空"))
		return
	}

	// 查询转账记录
	transfer, err := t.service.transferDB.queryByTransferNo(transferNo)
	if err != nil {
		t.Error("查询转账记录失败", zap.Error(err))
		c.ResponseError(errors.New("查询转账记录失败"))
		return
	}
	if transfer == nil {
		c.ResponseError(errors.New("转账记录不存在"))
		return
	}

	// 验证接收人
	if transfer.ToUID != uid {
		c.ResponseError(errors.New("您不是该转账的接收人"))
		return
	}

	// 验证状态
	if transfer.Status != TransferStatusPending {
		c.ResponseError(errors.New("转账已被处理"))
		return
	}

	// 更新状态
	err = t.service.transferDB.refund(transferNo)
	if err != nil {
		t.Error("退回转账失败", zap.Error(err))
		c.ResponseError(errors.New("退回转账失败"))
		return
	}

	// 退回给发送人
	senderWallet, _ := t.service.GetOrCreateWallet(transfer.FromUID)
	err = t.service.walletDB.addBalance(transfer.FromUID, transfer.Amount)
	if err != nil {
		t.Error("退回余额失败", zap.Error(err))
	}

	// 添加流水记录
	t.service.AddRecord(transfer.FromUID, RecordTypeTransferRefund, transfer.Amount, senderWallet.Balance, senderWallet.Balance+transfer.Amount, "转账被退回", transferNo, uid)

	c.ResponseOK()
}

// detail 转账详情
func (t *Transfer) detail(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	transferNo := c.Param("transfer_no")

	transfer, err := t.service.transferDB.queryByTransferNo(transferNo)
	if err != nil {
		t.Error("查询转账记录失败", zap.Error(err))
		c.ResponseError(errors.New("查询转账记录失败"))
		return
	}
	if transfer == nil {
		c.ResponseError(errors.New("转账记录不存在"))
		return
	}

	// 只有发送人和接收人可以查看
	if transfer.FromUID != uid && transfer.ToUID != uid {
		c.ResponseError(errors.New("无权查看该转账"))
		return
	}

	receivedAt := ""
	if transfer.ReceivedAt != nil {
		receivedAt = transfer.ReceivedAt.Format("2006-01-02 15:04:05")
	}

	c.Response(&transferDetailResp{
		TransferNo: transfer.TransferNo,
		FromUID:    transfer.FromUID,
		ToUID:      transfer.ToUID,
		Amount:     transfer.Amount,
		Remark:     transfer.Remark,
		Status:     transfer.Status,
		ExpireTime: transfer.ExpireTime.Format("2006-01-02 15:04:05"),
		ReceivedAt: receivedAt,
		CreatedAt:  transfer.CreatedAt.String(),
	})
}

// sentList 转出列表
func (t *Transfer) sentList(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	transfers, err := t.service.transferDB.queryByFromUID(uid, page, pageSize)
	if err != nil {
		t.Error("获取转出列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取转出列表失败"))
		return
	}

	list := make([]*transferListResp, 0, len(transfers))
	for _, tf := range transfers {
		list = append(list, &transferListResp{
			TransferNo: tf.TransferNo,
			ToUID:      tf.ToUID,
			Amount:     tf.Amount,
			Remark:     tf.Remark,
			Status:     tf.Status,
			CreatedAt:  tf.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// receivedList 收到列表
func (t *Transfer) receivedList(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	transfers, err := t.service.transferDB.queryByToUID(uid, page, pageSize)
	if err != nil {
		t.Error("获取收到列表失败", zap.Error(err))
		c.ResponseError(errors.New("获取收到列表失败"))
		return
	}

	list := make([]*transferListResp, 0, len(transfers))
	for _, tf := range transfers {
		list = append(list, &transferListResp{
			TransferNo: tf.TransferNo,
			FromUID:    tf.FromUID,
			Amount:     tf.Amount,
			Remark:     tf.Remark,
			Status:     tf.Status,
			CreatedAt:  tf.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// ========== 请求/响应结构体 ==========

type transferSendReq struct {
	ToUID       string `json:"to_uid"`       // 接收人UID
	Amount      int64  `json:"amount"`       // 金额(分)
	Remark      string `json:"remark"`       // 备注
	PayPassword string `json:"pay_password"` // 支付密码
}

type transferSendResp struct {
	TransferNo string `json:"transfer_no"` // 转账编号
	Amount     int64  `json:"amount"`      // 金额(分)
	ToUID      string `json:"to_uid"`      // 接收人UID
	Remark     string `json:"remark"`      // 备注
	ExpireTime string `json:"expire_time"` // 过期时间
}

type transferReceiveReq struct {
	TransferNo string `json:"transfer_no"` // 转账编号
}

type transferDetailResp struct {
	TransferNo string `json:"transfer_no"`
	FromUID    string `json:"from_uid"`
	ToUID      string `json:"to_uid"`
	Amount     int64  `json:"amount"`
	Remark     string `json:"remark"`
	Status     int    `json:"status"`
	ExpireTime string `json:"expire_time"`
	ReceivedAt string `json:"received_at"`
	CreatedAt  string `json:"created_at"`
}

type transferListResp struct {
	TransferNo string `json:"transfer_no"`
	FromUID    string `json:"from_uid,omitempty"`
	ToUID      string `json:"to_uid,omitempty"`
	Amount     int64  `json:"amount"`
	Remark     string `json:"remark"`
	Status     int    `json:"status"`
	CreatedAt  string `json:"created_at"`
}
