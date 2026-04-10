package wallet

import (
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

// walletDB 钱包数据库操作
type walletDB struct {
	session *dbr.Session
}

func newWalletDB(session *dbr.Session) *walletDB {
	return &walletDB{session: session}
}

// walletModel 钱包模型
type walletModel struct {
	Id              int64   `db:"id"`
	UID             string  `db:"uid"`
	Balance         int64   `db:"balance"`
	Frozen          int64   `db:"frozen"`
	TotalRecharge   int64   `db:"total_recharge"`
	TotalWithdraw   int64   `db:"total_withdraw"`
	Interest        int64   `db:"interest"`
	TodayInterest   int64   `db:"today_interest"`
	InterestRate    float64 `db:"interest_rate"`
	RealNameStatus  int     `db:"real_name_status"`
	RealName        string  `db:"real_name"`
	IdCard          string  `db:"id_card"`
	PayPassword     string  `db:"pay_password"`
	PayPasswordSalt string  `db:"pay_password_salt"`
	Status          int     `db:"status"`
	db.BaseModel
}

// queryByUID 根据UID查询钱包
func (d *walletDB) queryByUID(uid string) (*walletModel, error) {
	var model walletModel
	_, err := d.session.Select("*").From("wallet").Where("uid=?", uid).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.UID == "" {
		return nil, nil
	}
	return &model, nil
}

// insert 创建钱包
func (d *walletDB) insert(m *walletModel) error {
	_, err := d.session.InsertInto("wallet").Columns(
		"uid", "balance", "frozen", "total_recharge", "total_withdraw",
		"interest", "today_interest", "interest_rate",
		"real_name_status", "real_name", "id_card",
		"pay_password", "pay_password_salt", "status",
	).Record(m).Exec()
	return err
}

// updateBalance 更新余额
func (d *walletDB) updateBalance(uid string, balance int64) error {
	_, err := d.session.Update("wallet").Set("balance", balance).Where("uid=?", uid).Exec()
	return err
}

// updateFrozen 更新冻结金额
func (d *walletDB) updateFrozen(uid string, frozen int64) error {
	_, err := d.session.Update("wallet").Set("frozen", frozen).Where("uid=?", uid).Exec()
	return err
}

// updatePayPassword 更新支付密码
func (d *walletDB) updatePayPassword(uid string, password string, salt string) error {
	_, err := d.session.Update("wallet").
		Set("pay_password", password).
		Set("pay_password_salt", salt).
		Where("uid=?", uid).Exec()
	return err
}

// addBalance 增加余额(原子操作)
func (d *walletDB) addBalance(uid string, amount int64) error {
	_, err := d.session.UpdateBySql("UPDATE wallet SET balance = balance + ? WHERE uid = ?", amount, uid).Exec()
	return err
}

// subBalance 减少余额(原子操作)
func (d *walletDB) subBalance(uid string, amount int64) error {
	_, err := d.session.UpdateBySql("UPDATE wallet SET balance = balance - ? WHERE uid = ? AND balance >= ?", amount, uid, amount).Exec()
	return err
}

// addTotalRecharge 增加累计充值
func (d *walletDB) addTotalRecharge(uid string, amount int64) error {
	_, err := d.session.UpdateBySql("UPDATE wallet SET total_recharge = total_recharge + ? WHERE uid = ?", amount, uid).Exec()
	return err
}

// addTotalWithdraw 增加累计提现
func (d *walletDB) addTotalWithdraw(uid string, amount int64) error {
	_, err := d.session.UpdateBySql("UPDATE wallet SET total_withdraw = total_withdraw + ? WHERE uid = ?", amount, uid).Exec()
	return err
}

// updateRealNameStatus 更新钱包实名认证状态
func (d *walletDB) updateRealNameStatus(uid string, realNameStatus int, realName, idCard string) error {
	updateStmt := d.session.Update("wallet").Set("real_name_status", realNameStatus)
	if realName != "" {
		updateStmt = updateStmt.Set("real_name", realName)
	}
	if idCard != "" {
		updateStmt = updateStmt.Set("id_card", idCard)
	}
	_, err := updateStmt.Where("uid=?", uid).Exec()
	return err
}

// updateInterestRate 更新钱包利率
func (d *walletDB) updateInterestRate(uid string, interestRate float64) error {
	_, err := d.session.Update("wallet").Set("interest_rate", interestRate).Where("uid=?", uid).Exec()
	return err
}

// updateInterest 更新钱包利息
func (d *walletDB) updateInterest(uid string, interest, todayInterest int64) error {
	_, err := d.session.Update("wallet").Set("interest", interest).Set("today_interest", todayInterest).Where("uid=?", uid).Exec()
	return err
}

// update 更新钱包
func (d *walletDB) update(m *walletModel) error {
	_, err := d.session.Update("wallet").Set("balance", m.Balance).
		Set("frozen", m.Frozen).
		Set("total_recharge", m.TotalRecharge).
		Set("total_withdraw", m.TotalWithdraw).
		Set("interest", m.Interest).
		Set("today_interest", m.TodayInterest).
		Set("interest_rate", m.InterestRate).
		Set("real_name_status", m.RealNameStatus).
		Set("real_name", m.RealName).
		Set("id_card", m.IdCard).
		Set("pay_password", m.PayPassword).
		Set("pay_password_salt", m.PayPasswordSalt).
		Set("status", m.Status).
		Where("uid=?", m.UID).Exec()
	return err
}
