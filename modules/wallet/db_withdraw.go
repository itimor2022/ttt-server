package wallet

import (
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

// withdrawDB 提现订单数据库操作
type withdrawDB struct {
	session *dbr.Session
}

func newWithdrawDB(session *dbr.Session) *withdrawDB {
	return &withdrawDB{session: session}
}

// withdrawModel 提现订单模型
type withdrawModel struct {
	Id          int64      `db:"id"`
	OrderNo     string     `db:"order_no"`
	UID         string     `db:"uid"`
	Amount      int64      `db:"amount"`
	RealName    string     `db:"real_name"`
	BankName    string     `db:"bank_name"`
	BankCard    string     `db:"bank_card"`
	Status      int        `db:"status"`
	Remark      string     `db:"remark"`
	AdminUID    string     `db:"admin_uid"`
	AdminRemark string     `db:"admin_remark"`
	AuditedAt   *time.Time `db:"audited_at"`
	db.BaseModel
}

// insert 创建提现订单
func (d *withdrawDB) insert(m *withdrawModel) error {
	_, err := d.session.InsertInto("withdraw_order").Columns(
		"order_no", "uid", "amount", "real_name", "bank_name", "bank_card", "status", "remark",
	).Record(m).Exec()
	return err
}

// queryByOrderNo 根据订单号查询
func (d *withdrawDB) queryByOrderNo(orderNo string) (*withdrawModel, error) {
	var model withdrawModel
	_, err := d.session.Select("*").From("withdraw_order").Where("order_no=?", orderNo).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.OrderNo == "" {
		return nil, nil
	}
	return &model, nil
}

// queryByID 根据ID查询
func (d *withdrawDB) queryByID(id int64) (*withdrawModel, error) {
	var model withdrawModel
	_, err := d.session.Select("*").From("withdraw_order").Where("id=?", id).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.Id == 0 {
		return nil, nil
	}
	return &model, nil
}

// queryByUID 查询用户提现订单
func (d *withdrawDB) queryByUID(uid string, page, pageSize int) ([]*withdrawModel, error) {
	var list []*withdrawModel
	_, err := d.session.Select("*").From("withdraw_order").
		Where("uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByStatus 根据状态查询(管理后台)
func (d *withdrawDB) queryByStatus(status int, page, pageSize int) ([]*withdrawModel, error) {
	var list []*withdrawModel
	_, err := d.session.Select("*").From("withdraw_order").
		Where("status=?", status).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryAll 查询所有订单(管理后台)
func (d *withdrawDB) queryAll(page, pageSize int) ([]*withdrawModel, error) {
	var list []*withdrawModel
	_, err := d.session.Select("*").From("withdraw_order").
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// countByStatus 统计指定状态订单数量
func (d *withdrawDB) countByStatus(status int) (int64, error) {
	var count int64
	_, err := d.session.Select("COUNT(*)").From("withdraw_order").Where("status=?", status).Load(&count)
	return count, err
}

// approve 审核通过
func (d *withdrawDB) approve(id int64, adminUID string, adminRemark string) error {
	now := time.Now()
	_, err := d.session.Update("withdraw_order").
		Set("status", OrderStatusApproved).
		Set("admin_uid", adminUID).
		Set("admin_remark", adminRemark).
		Set("audited_at", now).
		Where("id=? AND status=?", id, OrderStatusPending).Exec()
	return err
}

// reject 审核拒绝
func (d *withdrawDB) reject(id int64, adminUID string, adminRemark string) error {
	now := time.Now()
	_, err := d.session.Update("withdraw_order").
		Set("status", OrderStatusRejected).
		Set("admin_uid", adminUID).
		Set("admin_remark", adminRemark).
		Set("audited_at", now).
		Where("id=? AND status=?", id, OrderStatusPending).Exec()
	return err
}
