package wallet

import (
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

// 转账状态
const (
	TransferStatusPending  = 0 // 待接收
	TransferStatusReceived = 1 // 已接收
	TransferStatusRefunded = 2 // 已退回
	TransferStatusExpired  = 3 // 已过期退回
)

// transferDB 转账数据库操作
type transferDB struct {
	session *dbr.Session
}

func newTransferDB(session *dbr.Session) *transferDB {
	return &transferDB{session: session}
}

// transferModel 转账模型
type transferModel struct {
	Id         int64      `db:"id"`
	TransferNo string     `db:"transfer_no"`
	FromUID    string     `db:"from_uid"`
	ToUID      string     `db:"to_uid"`
	Amount     int64      `db:"amount"`
	Remark     string     `db:"remark"`
	Status     int        `db:"status"`
	ExpireTime time.Time  `db:"expire_time"`
	ReceivedAt *time.Time `db:"received_at"`
	db.BaseModel
}

// insert 创建转账
func (d *transferDB) insert(m *transferModel) (int64, error) {
	result, err := d.session.InsertInto("transfer").Columns(
		"transfer_no", "from_uid", "to_uid", "amount", "remark", "status", "expire_time",
	).Record(m).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// queryByTransferNo 根据转账编号查询
func (d *transferDB) queryByTransferNo(transferNo string) (*transferModel, error) {
	var model transferModel
	_, err := d.session.Select("*").From("transfer").Where("transfer_no=?", transferNo).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.TransferNo == "" {
		return nil, nil
	}
	return &model, nil
}

// queryByID 根据ID查询
func (d *transferDB) queryByID(id int64) (*transferModel, error) {
	var model transferModel
	_, err := d.session.Select("*").From("transfer").Where("id=?", id).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.Id == 0 {
		return nil, nil
	}
	return &model, nil
}

// queryExpired 查询已过期未处理的转账
func (d *transferDB) queryExpired() ([]*transferModel, error) {
	var list []*transferModel
	now := time.Now()
	_, err := d.session.Select("*").From("transfer").
		Where("status=? AND expire_time<?", TransferStatusPending, now).
		Load(&list)
	return list, err
}

// receive 接收转账
func (d *transferDB) receive(transferNo string) error {
	now := time.Now()
	_, err := d.session.Update("transfer").
		Set("status", TransferStatusReceived).
		Set("received_at", now).
		Where("transfer_no=? AND status=?", transferNo, TransferStatusPending).Exec()
	return err
}

// refund 退回转账
func (d *transferDB) refund(transferNo string) error {
	now := time.Now()
	_, err := d.session.Update("transfer").
		Set("status", TransferStatusRefunded).
		Set("received_at", now).
		Where("transfer_no=? AND status=?", transferNo, TransferStatusPending).Exec()
	return err
}

// expire 过期退回
func (d *transferDB) expire(transferNo string) error {
	now := time.Now()
	_, err := d.session.Update("transfer").
		Set("status", TransferStatusExpired).
		Set("received_at", now).
		Where("transfer_no=? AND status=?", transferNo, TransferStatusPending).Exec()
	return err
}

// queryByFromUID 查询用户转出的记录
func (d *transferDB) queryByFromUID(uid string, page, pageSize int) ([]*transferModel, error) {
	var list []*transferModel
	_, err := d.session.Select("*").From("transfer").
		Where("from_uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByToUID 查询用户收到的转账记录
func (d *transferDB) queryByToUID(uid string, page, pageSize int) ([]*transferModel, error) {
	var list []*transferModel
	_, err := d.session.Select("*").From("transfer").
		Where("to_uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByStatus 按状态查询转账
func (d *transferDB) queryByStatus(status int, page, pageSize int) ([]*transferModel, error) {
	var list []*transferModel
	_, err := d.session.Select("*").From("transfer").
		Where("status=?", status).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryAll 查询所有转账
func (d *transferDB) queryAll(page, pageSize int) ([]*transferModel, error) {
	var list []*transferModel
	_, err := d.session.Select("*").From("transfer").
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}
