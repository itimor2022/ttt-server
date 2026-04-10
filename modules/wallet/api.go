package wallet

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Wallet 钱包API
type Wallet struct {
	ctx     *config.Context
	service *Service
	log.Log
}

// New 创建钱包API
func New(ctx *config.Context) *Wallet {
	return &Wallet{
		ctx:     ctx,
		service: NewService(ctx),
		Log:     log.NewTLog("walletAPI"),
	}
}

// Route 路由
func (w *Wallet) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/wallet", w.ctx.AuthMiddleware(r))
	{
		auth.GET("/balance", w.getBalance)                     // 获取余额
		auth.POST("/pay_password", w.setPayPassword)           // 设置支付密码
		auth.POST("/verify_pay_password", w.verifyPayPassword) // 验证支付密码
		auth.GET("/records", w.getRecords)                     // 获取流水记录
		auth.GET("/bills", w.getBills)                         // 获取综合账单(含充值提现申请)
		auth.POST("/recharge", w.applyRecharge)                // 申请充值
		auth.GET("/recharge/list", w.getRechargeList)          // 充值记录
		auth.POST("/withdraw", w.applyWithdraw)                // 申请提现
		auth.GET("/withdraw/list", w.getWithdrawList)          // 提现记录
		auth.POST("/real_name/verify", w.submitRealNameVerify) // 提交实名认证
	}
}

// getBalance 获取余额
func (w *Wallet) getBalance(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	wallet, err := w.service.GetOrCreateWallet(uid)
	if err != nil {
		w.Error("获取钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取钱包失败"))
		return
	}
	c.Response(&balanceResp{
		Balance:        wallet.Balance,
		Frozen:         wallet.Frozen,
		TotalRecharge:  wallet.TotalRecharge,
		TotalWithdraw:  wallet.TotalWithdraw,
		HasPayPwd:      wallet.PayPassword != "",
		RealNameStatus: wallet.RealNameStatus,
		RealName:       wallet.RealName,
		IdCard:         wallet.IdCard,
		Interest:       wallet.Interest,
		TodayInterest:  wallet.TodayInterest,
		InterestRate:   wallet.InterestRate,
	})
}

// setPayPassword 设置支付密码
func (w *Wallet) setPayPassword(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req setPayPwdReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		c.ResponseError(errors.New("支付密码不能为空"))
		return
	}
	if len(req.Password) != 6 {
		c.ResponseError(errors.New("支付密码必须为6位数字"))
		return
	}

	wallet, err := w.service.GetOrCreateWallet(uid)
	if err != nil {
		w.Error("获取钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取钱包失败"))
		return
	}

	// 如果已有密码，需要验证旧密码
	if wallet.PayPassword != "" {
		if req.OldPassword == "" {
			c.ResponseError(errors.New("请输入原支付密码"))
			return
		}
		oldPwdHash := w.service.EncryptPassword(req.OldPassword, wallet.PayPasswordSalt)
		if oldPwdHash != wallet.PayPassword {
			c.ResponseError(errors.New("原支付密码错误"))
			return
		}
	}

	salt := w.service.GenerateSalt()
	pwdHash := w.service.EncryptPassword(req.Password, salt)
	err = w.service.walletDB.updatePayPassword(uid, pwdHash, salt)
	if err != nil {
		w.Error("设置支付密码失败", zap.Error(err))
		c.ResponseError(errors.New("设置支付密码失败"))
		return
	}
	c.ResponseOK()
}

// verifyPayPassword 验证支付密码
func (w *Wallet) verifyPayPassword(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req verifyPayPwdReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	wallet, err := w.service.GetOrCreateWallet(uid)
	if err != nil {
		w.Error("获取钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取钱包失败"))
		return
	}

	if wallet.PayPassword == "" {
		c.ResponseError(errors.New("请先设置支付密码"))
		return
	}

	pwdHash := w.service.EncryptPassword(req.Password, wallet.PayPasswordSalt)
	if pwdHash != wallet.PayPassword {
		c.ResponseError(errors.New("支付密码错误"))
		return
	}
	c.ResponseOK()
}

// getRecords 获取流水记录
func (w *Wallet) getRecords(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := c.Query("page")
	pageSize := c.Query("page_size")
	recordType := c.Query("type")

	w.Info("获取流水记录请求", zap.String("uid", uid), zap.String("page", page), zap.String("pageSize", pageSize))

	pageInt := 1
	pageSizeInt := 20
	if page != "" {
		pageInt = parseInt(page, 1)
	}
	if pageSize != "" {
		pageSizeInt = parseInt(pageSize, 20)
	}

	var records []*recordModel
	var err error
	if recordType != "" {
		typeInt := parseInt(recordType, 0)
		records, err = w.service.recordDB.queryByUIDAndType(uid, typeInt, pageInt, pageSizeInt)
	} else {
		records, err = w.service.recordDB.queryByUID(uid, pageInt, pageSizeInt)
	}

	if err != nil {
		w.Error("获取流水记录失败", zap.Error(err))
		c.ResponseError(errors.New("获取流水记录失败"))
		return
	}

	w.Info("获取流水记录结果", zap.String("uid", uid), zap.Int("count", len(records)))

	list := make([]*recordResp, 0, len(records))
	for _, r := range records {
		list = append(list, &recordResp{
			RecordNo:      r.RecordNo,
			Type:          r.Type,
			TypeName:      w.service.GetRecordTypeName(r.Type),
			Amount:        r.Amount,
			BalanceBefore: r.BalanceBefore,
			BalanceAfter:  r.BalanceAfter,
			Remark:        r.Remark,
			RelatedUID:    r.RelatedUID,
			CreatedAt:     r.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// getBills 获取综合账单(含充值提现申请)
func (w *Wallet) getBills(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	w.Info("获取综合账单请求", zap.String("uid", uid), zap.Int("page", page), zap.Int("pageSize", pageSize))

	// 统一的账单项结构
	type billItem struct {
		Id         int64  `json:"id"`
		BillNo     string `json:"bill_no"`
		BillType   string `json:"bill_type"`   // record/recharge/withdraw
		Type       int    `json:"type"`        // 类型代码
		TypeName   string `json:"type_name"`   // 类型名称
		Amount     int64  `json:"amount"`      // 金额
		Status     int    `json:"status"`      // 状态 0-待处理 1-成功 2-已通过 3-已拒绝
		StatusName string `json:"status_name"` // 状态名称
		Remark     string `json:"remark"`      // 备注
		CreatedAt  string `json:"created_at"`  // 创建时间
	}

	var bills []billItem

	// 1. 获取流水记录（已完成的交易）
	records, err := w.service.recordDB.queryByUID(uid, 1, 100)
	if err != nil {
		w.Error("获取流水记录失败", zap.Error(err))
	} else {
		for _, r := range records {
			bills = append(bills, billItem{
				Id:         r.Id,
				BillNo:     r.RecordNo,
				BillType:   "record",
				Type:       r.Type,
				TypeName:   w.service.GetRecordTypeName(r.Type),
				Amount:     r.Amount,
				Status:     1, // 流水记录都是成功的
				StatusName: "成功",
				Remark:     r.Remark,
				CreatedAt:  r.CreatedAt.String(),
			})
		}
	}

	// 2. 获取充值申请（包括待审核和已拒绝的）
	recharges, err := w.service.rechargeDB.queryByUID(uid, 1, 100)
	if err != nil {
		w.Error("获取充值记录失败", zap.Error(err))
	} else {
		for _, r := range recharges {
			// 已通过的充值在流水中已显示，跳过
			if r.Status == OrderStatusApproved {
				continue
			}
			statusName := "待审核"
			if r.Status == OrderStatusRejected {
				statusName = "已拒绝"
			}
			bills = append(bills, billItem{
				Id:         r.Id,
				BillNo:     r.OrderNo,
				BillType:   "recharge",
				Type:       RecordTypeRecharge,
				TypeName:   "充值申请",
				Amount:     r.Amount,
				Status:     r.Status,
				StatusName: statusName,
				Remark:     r.Remark,
				CreatedAt:  r.CreatedAt.String(),
			})
		}
	}

	// 3. 获取提现申请（包括待审核和已拒绝的）
	withdraws, err := w.service.withdrawDB.queryByUID(uid, 1, 100)
	if err != nil {
		w.Error("获取提现记录失败", zap.Error(err))
	} else {
		for _, r := range withdraws {
			// 已通过的提现在流水中已显示，跳过
			if r.Status == OrderStatusApproved {
				continue
			}
			statusName := "待审核"
			if r.Status == OrderStatusRejected {
				statusName = "已拒绝"
			}
			bills = append(bills, billItem{
				Id:         r.Id,
				BillNo:     r.OrderNo,
				BillType:   "withdraw",
				Type:       RecordTypeWithdraw,
				TypeName:   "提现申请",
				Amount:     r.Amount,
				Status:     r.Status,
				StatusName: statusName,
				Remark:     r.Remark,
				CreatedAt:  r.CreatedAt.String(),
			})
		}
	}

	// 按时间排序（最新的在前）
	sort.Slice(bills, func(i, j int) bool {
		return bills[i].CreatedAt > bills[j].CreatedAt
	})

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(bills) {
		c.Response([]billItem{})
		return
	}
	if end > len(bills) {
		end = len(bills)
	}

	w.Info("获取综合账单结果", zap.String("uid", uid), zap.Int("total", len(bills)), zap.Int("返回", end-start))
	c.Response(bills[start:end])
}

// applyRecharge 申请充值
func (w *Wallet) applyRecharge(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req rechargeReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}
	if req.Amount <= 0 {
		c.ResponseError(errors.New("充值金额必须大于0"))
		return
	}

	orderNo := w.service.GenerateOrderNo("CZ")
	order := &rechargeModel{
		OrderNo: orderNo,
		UID:     uid,
		Amount:  req.Amount,
		Status:  OrderStatusPending,
		Remark:  req.Remark,
	}
	err := w.service.rechargeDB.insert(order)
	if err != nil {
		w.Error("创建充值订单失败", zap.Error(err))
		c.ResponseError(errors.New("创建充值订单失败"))
		return
	}
	c.Response(map[string]string{"order_no": orderNo})
}

// getRechargeList 充值记录
func (w *Wallet) getRechargeList(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	orders, err := w.service.rechargeDB.queryByUID(uid, page, pageSize)
	if err != nil {
		w.Error("获取充值记录失败", zap.Error(err))
		c.ResponseError(errors.New("获取充值记录失败"))
		return
	}

	list := make([]*rechargeResp, 0, len(orders))
	for _, o := range orders {
		list = append(list, &rechargeResp{
			OrderNo:   o.OrderNo,
			Amount:    o.Amount,
			Status:    o.Status,
			Remark:    o.Remark,
			CreatedAt: o.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// applyWithdraw 申请提现
func (w *Wallet) applyWithdraw(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req withdrawReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}
	if req.Amount <= 0 {
		c.ResponseError(errors.New("提现金额必须大于0"))
		return
	}
	if strings.TrimSpace(req.RealName) == "" {
		c.ResponseError(errors.New("真实姓名不能为空"))
		return
	}
	if strings.TrimSpace(req.BankName) == "" {
		c.ResponseError(errors.New("银行名称不能为空"))
		return
	}
	if strings.TrimSpace(req.BankCard) == "" {
		c.ResponseError(errors.New("银行卡号不能为空"))
		return
	}

	// 验证支付密码
	wallet, err := w.service.GetOrCreateWallet(uid)
	if err != nil {
		w.Error("获取钱包失败", zap.Error(err))
		c.ResponseError(errors.New("获取钱包失败"))
		return
	}
	if wallet.PayPassword == "" {
		c.ResponseError(errors.New("请先设置支付密码"))
		return
	}
	pwdHash := w.service.EncryptPassword(req.PayPassword, wallet.PayPasswordSalt)
	if pwdHash != wallet.PayPassword {
		c.ResponseError(errors.New("支付密码错误"))
		return
	}

	// 检查余额
	if wallet.Balance < req.Amount {
		c.ResponseError(errors.New("余额不足"))
		return
	}

	// 使用事务处理提现申请
	orderNo, err := w.service.ApplyWithdraw(uid, req.Amount, req.RealName, req.BankName, req.BankCard, req.Remark, wallet)
	if err != nil {
		w.Error("申请提现失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.Response(map[string]string{"order_no": orderNo})
}

// getWithdrawList 提现记录
func (w *Wallet) getWithdrawList(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)

	orders, err := w.service.withdrawDB.queryByUID(uid, page, pageSize)
	if err != nil {
		w.Error("获取提现记录失败", zap.Error(err))
		c.ResponseError(errors.New("获取提现记录失败"))
		return
	}

	list := make([]*withdrawResp, 0, len(orders))
	for _, o := range orders {
		list = append(list, &withdrawResp{
			OrderNo:   o.OrderNo,
			Amount:    o.Amount,
			RealName:  o.RealName,
			BankName:  o.BankName,
			BankCard:  o.BankCard,
			Status:    o.Status,
			Remark:    o.Remark,
			CreatedAt: o.CreatedAt.String(),
		})
	}
	c.Response(list)
}

// 辅助函数
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return defaultVal
	}
	return val
}

// ========== 请求/响应结构体 ==========

type balanceResp struct {
	Balance        int64   `json:"balance"`          // 可用余额(分)
	Frozen         int64   `json:"frozen"`           // 冻结金额(分)
	TotalRecharge  int64   `json:"total_recharge"`   // 累计充值(分)
	TotalWithdraw  int64   `json:"total_withdraw"`   // 累计提现(分)
	HasPayPwd      bool    `json:"has_pay_pwd"`      // 是否设置了支付密码
	RealNameStatus int     `json:"real_name_status"` // 实名认证状态：0-未认证，1-审核中，2-已认证，3-认证失败
	RealName       string  `json:"real_name"`        // 真实姓名
	IdCard         string  `json:"id_card"`          // 身份证号
	Interest       int64   `json:"interest"`         // 累计利息(分)
	TodayInterest  int64   `json:"today_interest"`   // 今日利息(分)
	InterestRate   float64 `json:"interest_rate"`    // 年化利率(%)
}

type setPayPwdReq struct {
	Password    string `json:"password"`     // 新支付密码
	OldPassword string `json:"old_password"` // 原支付密码(修改时需要)
}

type verifyPayPwdReq struct {
	Password string `json:"password"` // 支付密码
}

type recordResp struct {
	RecordNo      string `json:"record_no"`
	Type          int    `json:"type"`
	TypeName      string `json:"type_name"`
	Amount        int64  `json:"amount"`
	BalanceBefore int64  `json:"balance_before"`
	BalanceAfter  int64  `json:"balance_after"`
	Remark        string `json:"remark"`
	RelatedUID    string `json:"related_uid"`
	CreatedAt     string `json:"created_at"`
}

type rechargeReq struct {
	Amount int64  `json:"amount"` // 充值金额(分)
	Remark string `json:"remark"` // 备注
}

type rechargeResp struct {
	OrderNo   string `json:"order_no"`
	Amount    int64  `json:"amount"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"created_at"`
}

type withdrawReq struct {
	Amount      int64  `json:"amount"`       // 提现金额(分)
	RealName    string `json:"real_name"`    // 真实姓名
	BankName    string `json:"bank_name"`    // 银行名称
	BankCard    string `json:"bank_card"`    // 银行卡号
	PayPassword string `json:"pay_password"` // 支付密码
	Remark      string `json:"remark"`       // 备注
}

type withdrawResp struct {
	OrderNo   string `json:"order_no"`
	Amount    int64  `json:"amount"`
	RealName  string `json:"real_name"`
	BankName  string `json:"bank_name"`
	BankCard  string `json:"bank_card"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"created_at"`
}

// 实名认证请求

type realNameVerifyReq struct {
	RealName string `json:"real_name"` // 真实姓名
	IdCard   string `json:"id_card"`   // 身份证号
	Phone    string `json:"phone"`     // 手机号码
	Code     string `json:"code"`      // 验证码
}

// submitRealNameVerify 提交实名认证
func (w *Wallet) submitRealNameVerify(c *wkhttp.Context) {
	uid := c.GetLoginUID()
	var req realNameVerifyReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}
	if strings.TrimSpace(req.RealName) == "" {
		c.ResponseError(errors.New("真实姓名不能为空"))
		return
	}
	if strings.TrimSpace(req.IdCard) == "" {
		c.ResponseError(errors.New("身份证号不能为空"))
		return
	}
	if strings.TrimSpace(req.Phone) == "" {
		c.ResponseError(errors.New("手机号码不能为空"))
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		c.ResponseError(errors.New("验证码不能为空"))
		return
	}

	err := w.service.SubmitRealNameVerify(uid, req.RealName, req.IdCard, req.Phone, req.Code)
	if err != nil {
		w.Error("提交实名认证失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.ResponseOK()
}
