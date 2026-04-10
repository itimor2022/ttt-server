package wallet

import (
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

// 订单状态
const (
	OrderStatusPending  = 0 // 待审核
	OrderStatusApproved = 1 // 已通过
	OrderStatusRejected = 2 // 已拒绝
)

// rechargeDB 充值订单数据库操作
type rechargeDB struct {
	session *dbr.Session
}

func newRechargeDB(session *dbr.Session) *rechargeDB {
	return &rechargeDB{session: session}
}

// rechargeModel 充值订单模型
type rechargeModel struct {
	Id          int64      `db:"id"`
	OrderNo     string     `db:"order_no"`
	UID         string     `db:"uid"`
	Amount      int64      `db:"amount"`
	Status      int        `db:"status"`
	Remark      string     `db:"remark"`
	AdminUID    string     `db:"admin_uid"`
	AdminRemark string     `db:"admin_remark"`
	AuditedAt   *time.Time `db:"audited_at"`
	db.BaseModel
}

// insert 创建充值订单
func (d *rechargeDB) insert(m *rechargeModel) error {
	_, err := d.session.InsertInto("recharge_order").Columns(
		"order_no", "uid", "amount", "status", "remark",
	).Record(m).Exec()
	return err
}

// queryByOrderNo 根据订单号查询
func (d *rechargeDB) queryByOrderNo(orderNo string) (*rechargeModel, error) {
	var model rechargeModel
	_, err := d.session.Select("*").From("recharge_order").Where("order_no=?", orderNo).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.OrderNo == "" {
		return nil, nil
	}
	return &model, nil
}

// queryByID 根据ID查询
func (d *rechargeDB) queryByID(id int64) (*rechargeModel, error) {
	var model rechargeModel
	_, err := d.session.Select("*").From("recharge_order").Where("id=?", id).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.Id == 0 {
		return nil, nil
	}
	return &model, nil
}

// queryByUID 查询用户充值订单
func (d *rechargeDB) queryByUID(uid string, page, pageSize int) ([]*rechargeModel, error) {
	var list []*rechargeModel
	_, err := d.session.Select("*").From("recharge_order").
		Where("uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByStatus 根据状态查询(管理后台)
func (d *rechargeDB) queryByStatus(status int, page, pageSize int) ([]*rechargeModel, error) {
	var list []*rechargeModel
	_, err := d.session.Select("*").From("recharge_order").
		Where("status=?", status).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryAll 查询所有订单(管理后台)
func (d *rechargeDB) queryAll(page, pageSize int) ([]*rechargeModel, error) {
	var list []*rechargeModel
	_, err := d.session.Select("*").From("recharge_order").
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// countByStatus 统计指定状态订单数量
func (d *rechargeDB) countByStatus(status int) (int64, error) {
	var count int64
	_, err := d.session.Select("COUNT(*)").From("recharge_order").Where("status=?", status).Load(&count)
	return count, err
}

// approve 审核通过
func (d *rechargeDB) approve(id int64, adminUID string, adminRemark string) error {
	now := time.Now()
	_, err := d.session.Update("recharge_order").
		Set("status", OrderStatusApproved).
		Set("admin_uid", adminUID).
		Set("admin_remark", adminRemark).
		Set("audited_at", now).
		Where("id=? AND status=?", id, OrderStatusPending).Exec()
	return err
}

// reject 审核拒绝
func (d *rechargeDB) reject(id int64, adminUID string, adminRemark string) error {
	now := time.Now()
	_, err := d.session.Update("recharge_order").
		Set("status", OrderStatusRejected).
		Set("admin_uid", adminUID).
		Set("admin_remark", adminRemark).
		Set("audited_at", now).
		Where("id=? AND status=?", id, OrderStatusPending).Exec()
	return err
}
