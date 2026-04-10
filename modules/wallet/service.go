package wallet

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
)

// Service 钱包服务
type Service struct {
	ctx               *config.Context
	walletDB          *walletDB
	recordDB          *recordDB
	rechargeDB        *rechargeDB
	withdrawDB        *withdrawDB
	redPacketDB       *redPacketDB
	redPacketRecordDB *redPacketRecordDB
	transferDB        *transferDB
	log.Log
}

// NewService 创建服务
func NewService(ctx *config.Context) *Service {
	return &Service{
		ctx:               ctx,
		walletDB:          newWalletDB(ctx.DB()),
		recordDB:          newRecordDB(ctx.DB()),
		rechargeDB:        newRechargeDB(ctx.DB()),
		withdrawDB:        newWithdrawDB(ctx.DB()),
		redPacketDB:       newRedPacketDB(ctx.DB()),
		redPacketRecordDB: newRedPacketRecordDB(ctx.DB()),
		transferDB:        newTransferDB(ctx.DB()),
		Log:               log.NewTLog("walletService"),
	}
}

// GetOrCreateWallet 获取或创建钱包
func (s *Service) GetOrCreateWallet(uid string) (*walletModel, error) {
	wallet, err := s.walletDB.queryByUID(uid)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		wallet = &walletModel{
			UID:            uid,
			Balance:        0,
			Frozen:         0,
			TotalRecharge:  0,
			TotalWithdraw:  0,
			Interest:       0,
			TodayInterest:  0,
			InterestRate:   0,
			RealNameStatus: 0,
			Status:         1,
		}
		err = s.walletDB.insert(wallet)
		if err != nil {
			return nil, err
		}
		wallet, err = s.walletDB.queryByUID(uid)
		if err != nil {
			return nil, err
		}
	}
	return wallet, nil
}

// EncryptPassword 加密支付密码
func (s *Service) EncryptPassword(password string, salt string) string {
	hash := md5.Sum([]byte(password + salt))
	return fmt.Sprintf("%x", hash)
}

// GenerateSalt 生成盐值
func (s *Service) GenerateSalt() string {
	return util.GenerUUID()[:16]
}

// GenerateOrderNo 生成订单号
func (s *Service) GenerateOrderNo(prefix string) string {
	// 使用时间戳+UUID前8位，避免高并发重复
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), util.GenerUUID()[:8])
}

// GenerateRecordNo 生成流水号
func (s *Service) GenerateRecordNo() string {
	return s.GenerateOrderNo("R")
}

// AddRecord 添加流水记录
func (s *Service) AddRecord(uid string, recordType int, amount int64, balanceBefore int64, balanceAfter int64, remark string, relatedId string, relatedUID string) error {
	record := &recordModel{
		UID:           uid,
		RecordNo:      s.GenerateRecordNo(),
		Type:          recordType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Remark:        remark,
		RelatedId:     relatedId,
		RelatedUID:    relatedUID,
	}
	return s.recordDB.insert(record)
}

// CalculateLuckyAmount 计算拼手气红包金额
func (s *Service) CalculateLuckyAmount(remainAmount int64, remainCount int) int64 {
	if remainCount == 1 {
		return remainAmount
	}
	// 最小1分钱
	minAmount := int64(1)
	// 最大金额 = 剩余金额 - 剩余人数 * 最小金额，保证每人至少1分
	maxAmount := remainAmount - int64(remainCount-1)*minAmount
	if maxAmount <= minAmount {
		return minAmount
	}
	// 随机金额在 [minAmount, maxAmount*2/remainCount] 范围内，避免前面的人抢太多
	avgMax := maxAmount * 2 / int64(remainCount)
	if avgMax < minAmount {
		avgMax = minAmount
	}
	amount := minAmount + rand.Int63n(avgMax)
	if amount > maxAmount {
		amount = maxAmount
	}
	return amount
}

// GetRecordTypeName 获取流水类型名称
func (s *Service) GetRecordTypeName(recordType int) string {
	switch recordType {
	case RecordTypeRecharge:
		return "充值"
	case RecordTypeWithdraw:
		return "提现"
	case RecordTypeTransferIn:
		return "转账收入"
	case RecordTypeTransferOut:
		return "转账支出"
	case RecordTypeRedPacketOut:
		return "发红包"
	case RecordTypeRedPacketIn:
		return "收红包"
	case RecordTypeRedPacketRefund:
		return "红包退回"
	case RecordTypeTransferRefund:
		return "转账退回"
	case RecordTypeInterest:
		return "利息收益"
	case RecordTypeBalanceAdjust:
		return "余额调整"
	default:
		return "未知"
	}
}

// SubmitRealNameVerify 提交实名认证
func (s *Service) SubmitRealNameVerify(uid string, realName string, idCard string, phone string, code string) error {
	// 开启事务
	tx, err := s.ctx.DB().Begin()
	if err != nil {
		return errors.New("系统繁忙，请重试")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已经认证
	var existing int
	err = tx.QueryRow("SELECT status FROM wallet_real_name WHERE uid = ?", uid).Scan(&existing)
	if err == nil && existing > 0 {
		return errors.New("您已经提交过实名认证")
	}

	// 创建实名认证记录
	_, err = tx.InsertInto("wallet_real_name").Columns(
		"uid", "real_name", "id_card", "phone", "status",
	).Values(uid, realName, idCard, phone, 0).Exec()
	if err != nil {
		return errors.New("提交认证失败，请重试")
	}

	// 更新钱包表中的实名认证状态为审核中
	_, err = tx.UpdateBySql("UPDATE wallet SET real_name_status = 1, real_name = ?, id_card = ? WHERE uid = ?",
		realName, idCard, uid).Exec()
	if err != nil {
		return errors.New("更新认证状态失败")
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return errors.New("提交认证失败，请重试")
	}

	return nil
}

// GrabResult 抢红包结果
type GrabResult struct {
	Amount  int64
	IsBest  bool
	Remark  string
	FromUID string
}

// GrabRedPacket 抢红包(事务处理，保证并发安全)
func (s *Service) GrabRedPacket(uid string, packetNo string) (*GrabResult, error) {
	// 开启事务
	tx, err := s.ctx.DB().Begin()
	if err != nil {
		return nil, errors.New("系统繁忙，请重试")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 使用 FOR UPDATE 锁定红包行
	packet, err := s.redPacketDB.queryByPacketNoForUpdate(tx, packetNo)
	if err != nil {
		return nil, errors.New("查询红包失败")
	}
	if packet == nil {
		return nil, errors.New("红包不存在")
	}

	// 检查状态
	if packet.Status != RedPacketStatusActive {
		if packet.Status == RedPacketStatusFinished {
			return nil, errors.New("红包已被抢完")
		}
		return nil, errors.New("红包已过期")
	}

	// 检查是否过期
	if time.Now().After(packet.ExpireTime) {
		return nil, errors.New("红包已过期")
	}

	// 检查剩余数量
	if packet.RemainCount <= 0 {
		return nil, errors.New("红包已被抢完")
	}

	// 检查是否已领取(在事务内查询)
	existRecord, _ := s.redPacketRecordDB.queryByPacketIdAndUIDTx(tx, packet.Id, uid)
	if existRecord != nil {
		return nil, errors.New("您已领取过该红包")
	}

	// 计算领取金额
	var grabAmount int64
	if packet.Type == RedPacketTypeLucky {
		grabAmount = s.CalculateLuckyAmount(packet.RemainAmount, packet.RemainCount)
	} else {
		// 普通红包平均分
		grabAmount = packet.TotalAmount / int64(packet.TotalCount)
		if packet.RemainCount == 1 {
			grabAmount = packet.RemainAmount
		}
	}

	// 扣减红包金额(在事务内)
	newRemainAmount := packet.RemainAmount - grabAmount
	newRemainCount := packet.RemainCount - 1
	err = s.redPacketDB.updateRemainTx(tx, packetNo, newRemainAmount, newRemainCount)
	if err != nil {
		return nil, errors.New("抢红包失败，请重试")
	}

	// 创建领取记录(在事务内)
	record := &redPacketRecordModel{
		PacketId: packet.Id,
		PacketNo: packet.PacketNo,
		UID:      uid,
		Amount:   grabAmount,
		IsBest:   0,
	}
	err = s.redPacketRecordDB.insertTx(tx, record)
	if err != nil {
		return nil, errors.New("抢红包失败，请重试")
	}

	// 如果红包抢完了，更新状态并计算手气最佳
	if newRemainCount == 0 {
		err = s.redPacketDB.updateStatusTx(tx, packetNo, RedPacketStatusFinished)
		if err != nil {
			return nil, errors.New("抢红包失败，请重试")
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errors.New("抢红包失败，请重试")
	}

	// 事务提交后，更新用户余额和流水(这些操作失败不影响抢红包结果)
	wallet, _ := s.GetOrCreateWallet(uid)
	s.walletDB.addBalance(uid, grabAmount)
	s.AddRecord(uid, RecordTypeRedPacketIn, grabAmount, wallet.Balance, wallet.Balance+grabAmount, "收红包-"+packet.Remark, packetNo, packet.UID)

	// 如果红包抢完了，计算手气最佳(异步)
	if newRemainCount == 0 {
		go s.calculateBestLuck(packet.Id)
	}

	return &GrabResult{
		Amount:  grabAmount,
		IsBest:  false,
		Remark:  packet.Remark,
		FromUID: packet.UID,
	}, nil
}

// calculateBestLuck 计算手气最佳
func (s *Service) calculateBestLuck(packetId int64) {
	records, err := s.redPacketRecordDB.queryByPacketId(packetId)
	if err != nil || len(records) == 0 {
		return
	}

	var bestUID string
	var bestAmount int64 = 0
	for _, r := range records {
		if r.Amount > bestAmount {
			bestAmount = r.Amount
			bestUID = r.UID
		}
	}
	if bestUID != "" {
		s.redPacketRecordDB.updateBest(packetId, bestUID)
	}
}

// ApplyWithdraw 申请提现(事务处理)
func (s *Service) ApplyWithdraw(uid string, amount int64, realName, bankName, bankCard, remark string, wallet *walletModel) (string, error) {
	tx, err := s.ctx.DB().Begin()
	if err != nil {
		return "", errors.New("系统繁忙，请重试")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 扣减余额
	result, err := tx.UpdateBySql("UPDATE wallet SET balance = balance - ?, frozen = frozen + ? WHERE uid = ? AND balance >= ?",
		amount, amount, uid, amount).Exec()
	if err != nil {
		return "", errors.New("冻结金额失败")
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return "", errors.New("余额不足")
	}

	// 创建提现订单
	orderNo := s.GenerateOrderNo("TX")
	_, err = tx.InsertInto("withdraw_order").Columns(
		"order_no", "uid", "amount", "real_name", "bank_name", "bank_card", "status", "remark",
	).Values(orderNo, uid, amount, realName, bankName, bankCard, OrderStatusPending, remark).Exec()
	if err != nil {
		return "", errors.New("创建提现订单失败")
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return "", errors.New("提现申请失败，请重试")
	}

	return orderNo, nil
}

// RefundRedPacket 退回红包(用于过期处理)
func (s *Service) RefundRedPacket(packetNo string) error {
	packet, err := s.redPacketDB.queryByPacketNo(packetNo)
	if err != nil || packet == nil {
		return errors.New("红包不存在")
	}

	if packet.Status != RedPacketStatusActive {
		return nil // 已处理
	}

	if packet.RemainAmount <= 0 {
		return nil
	}

	// 退回余额
	wallet, _ := s.GetOrCreateWallet(packet.UID)
	err = s.walletDB.addBalance(packet.UID, packet.RemainAmount)
	if err != nil {
		return err
	}

	// 添加流水
	s.AddRecord(packet.UID, RecordTypeRedPacketRefund, packet.RemainAmount, wallet.Balance, wallet.Balance+packet.RemainAmount, "红包过期退回", packetNo, "")

	// 更新状态
	return s.redPacketDB.updateStatus(packetNo, RedPacketStatusExpired)
}

// RefundTransfer 退回转账(用于过期处理)
func (s *Service) RefundTransfer(transferNo string) error {
	transfer, err := s.transferDB.queryByTransferNo(transferNo)
	if err != nil || transfer == nil {
		return errors.New("转账记录不存在")
	}

	if transfer.Status != TransferStatusPending {
		return nil // 已处理
	}

	// 退回余额
	wallet, _ := s.GetOrCreateWallet(transfer.FromUID)
	err = s.walletDB.addBalance(transfer.FromUID, transfer.Amount)
	if err != nil {
		return err
	}

	// 添加流水
	s.AddRecord(transfer.FromUID, RecordTypeTransferRefund, transfer.Amount, wallet.Balance, wallet.Balance+transfer.Amount, "转账过期退回", transferNo, transfer.ToUID)

	// 更新状态
	return s.transferDB.expire(transferNo)
}

// GetInterestRate 获取当前利率设置
func (s *Service) GetInterestRate() (float64, error) {
	// 从数据库中获取利率设置
	// 查询钱包表中的第一个记录，获取当前设置的利率
	var interestRate float64
	err := s.ctx.DB().QueryRow("SELECT interest_rate FROM wallet ORDER BY id LIMIT 1").Scan(&interestRate)
	if err != nil {
		// 如果查询失败或没有记录，返回默认值
		return 0.05, nil // 默认年化利率5%
	}
	return interestRate, nil
}

// SetInterestRate 设置利率
func (s *Service) SetInterestRate(interestRate float64) error {
	if interestRate < 0 {
		return errors.New("利率不能为负数")
	}

	// 这里应该将利率存储到配置或数据库中
	// 实际实现中应该更新配置文件或数据库中的利率设置
	// 暂时先更新所有钱包的利率
	_, err := s.ctx.DB().UpdateBySql("UPDATE wallet SET interest_rate = ?", interestRate).Exec()
	if err != nil {
		return err
	}

	return nil
}
