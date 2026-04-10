package wallet

import (
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"go.uber.org/zap"
)

// ExpireTask 过期退款任务
type ExpireTask struct {
	ctx     *config.Context
	service *Service
	log.Log
}

// NewExpireTask 创建过期退款任务
func NewExpireTask(ctx *config.Context) *ExpireTask {
	return &ExpireTask{
		ctx:     ctx,
		service: NewService(ctx),
		Log:     log.NewTLog("expireTask"),
	}
}

// Start 启动定时任务
func (t *ExpireTask) Start() {
	// 每分钟检查一次过期红包和转账
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			t.processExpiredRedPackets()
			t.processExpiredTransfers()
		}
	}()
	t.Info("过期退款定时任务已启动")

	// 每天凌晨00:00执行一次利息计算
	go t.scheduleInterestCalculation()
}

// processExpiredRedPackets 处理过期红包
func (t *ExpireTask) processExpiredRedPackets() {
	packets, err := t.service.redPacketDB.queryExpired()
	if err != nil {
		t.Error("查询过期红包失败", zap.Error(err))
		return
	}

	for _, packet := range packets {
		if packet.RemainAmount <= 0 {
			// 没有剩余金额，只更新状态
			t.service.redPacketDB.updateStatus(packet.PacketNo, RedPacketStatusExpired)
			continue
		}

		// 退回剩余金额给发红包的人
		err = t.refundRedPacket(packet)
		if err != nil {
			t.Error("退回红包失败", zap.String("packet_no", packet.PacketNo), zap.Error(err))
			continue
		}

		t.Info("红包退款成功", zap.String("packet_no", packet.PacketNo), zap.Int64("amount", packet.RemainAmount))
	}
}

// refundRedPacket 退回红包
func (t *ExpireTask) refundRedPacket(packet *redPacketModel) error {
	// 获取发红包用户的钱包
	wallet, err := t.service.GetOrCreateWallet(packet.UID)
	if err != nil {
		return err
	}

	// 增加余额
	err = t.service.walletDB.addBalance(packet.UID, packet.RemainAmount)
	if err != nil {
		return err
	}

	// 添加流水记录
	t.service.AddRecord(
		packet.UID,
		RecordTypeRedPacketRefund,
		packet.RemainAmount,
		wallet.Balance,
		wallet.Balance+packet.RemainAmount,
		"红包过期退回",
		packet.PacketNo,
		"",
	)

	// 更新红包状态
	return t.service.redPacketDB.updateStatus(packet.PacketNo, RedPacketStatusExpired)
}

// processExpiredTransfers 处理过期转账
func (t *ExpireTask) processExpiredTransfers() {
	transfers, err := t.service.transferDB.queryExpired()
	if err != nil {
		t.Error("查询过期转账失败", zap.Error(err))
		return
	}

	for _, transfer := range transfers {
		err = t.refundTransfer(transfer)
		if err != nil {
			t.Error("退回转账失败", zap.String("transfer_no", transfer.TransferNo), zap.Error(err))
			continue
		}

		t.Info("转账退款成功", zap.String("transfer_no", transfer.TransferNo), zap.Int64("amount", transfer.Amount))
	}
}

// refundTransfer 退回转账
func (t *ExpireTask) refundTransfer(transfer *transferModel) error {
	// 获取发起转账用户的钱包
	wallet, err := t.service.GetOrCreateWallet(transfer.FromUID)
	if err != nil {
		return err
	}

	// 增加余额
	err = t.service.walletDB.addBalance(transfer.FromUID, transfer.Amount)
	if err != nil {
		return err
	}

	// 添加流水记录
	t.service.AddRecord(
		transfer.FromUID,
		RecordTypeTransferRefund,
		transfer.Amount,
		wallet.Balance,
		wallet.Balance+transfer.Amount,
		"转账过期退回",
		transfer.TransferNo,
		transfer.ToUID,
	)

	// 更新转账状态
	return t.service.transferDB.expire(transfer.TransferNo)
}

// scheduleInterestCalculation 调度利息计算任务
func (t *ExpireTask) scheduleInterestCalculation() {
	for {
		// 计算下一次执行时间（当天凌晨00:00）
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := next.Sub(now)

		// 等待到执行时间
		time.Sleep(duration)

		// 执行利息计算
		t.calculateInterest()
	}
}

// calculateInterest 计算利息
func (t *ExpireTask) calculateInterest() {
	// 查询所有活跃的钱包
	rows, err := t.service.ctx.DB().Select("id, uid, balance, interest_rate").From("wallet").Where("status=? AND balance>0 AND interest_rate>0", 1).Rows()
	if err != nil {
		t.Error("查询钱包失败", zap.Error(err))
		return
	}
	defer rows.Close()

	// 遍历钱包计算利息
	for rows.Next() {
		var id int64
		var uid string
		var balance int64
		var interestRate float64

		if err := rows.Scan(&id, &uid, &balance, &interestRate); err != nil {
			t.Error("扫描钱包数据失败", zap.Error(err))
			continue
		}

		// 计算当日利息：余额 * 年化利率 / 365
		// 保留2位小数，转换为分
		todayInterest := int64(float64(balance) * interestRate / 100 / 365 * 100)
		if todayInterest <= 0 {
			continue // 利息为0，跳过
		}

		// 计算更新后的余额
		newBalance := balance + todayInterest

		// 更新钱包利息
		tx, err := t.service.ctx.DB().Begin()
		if err != nil {
			t.Error("开启事务失败", zap.Error(err))
			continue
		}

		// 增加余额和利息
		_, err = tx.UpdateBySql(
			"UPDATE wallet SET balance = ?, interest = interest + ?, today_interest = ? WHERE id = ?",
			newBalance, todayInterest, todayInterest, id,
		).Exec()
		if err != nil {
			tx.Rollback()
			t.Error("更新利息失败", zap.String("uid", uid), zap.Error(err))
			continue
		}

		// 添加利息收益流水记录
		recordNo := t.service.GenerateRecordNo()
		_, err = tx.Exec(
			"INSERT INTO wallet_record (uid, record_no, type, amount, balance_before, balance_after, remark, related_id, related_uid, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			uid, recordNo, RecordTypeInterest, todayInterest, balance, newBalance, "每日利息收益", "", "", time.Now(),
		)
		if err != nil {
			tx.Rollback()
			t.Error("添加利息记录失败", zap.String("uid", uid), zap.Error(err))
			continue
		}

		// 提交事务
		if err = tx.Commit(); err != nil {
			t.Error("提交事务失败", zap.Error(err))
			continue
		}

		t.Info("利息计算成功", zap.String("uid", uid), zap.Int64("interest", todayInterest))
	}

	t.Info("每日利息计算完成")
}
