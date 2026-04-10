package wallet

import (
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

// 流水类型
const (
	RecordTypeRecharge        = 1  // 充值
	RecordTypeWithdraw        = 2  // 提现
	RecordTypeTransferIn      = 3  // 转账收入
	RecordTypeTransferOut     = 4  // 转账支出
	RecordTypeRedPacketOut    = 5  // 红包发出
	RecordTypeRedPacketIn     = 6  // 红包收入
	RecordTypeRedPacketRefund = 7  // 红包退回
	RecordTypeInterest        = 8  // 利息收入
	RecordTypeTransferRefund  = 9  // 转账退回
	RecordTypeBalanceAdjust   = 10 // 余额调整
)

// recordDB 流水记录数据库操作
type recordDB struct {
	session *dbr.Session
}

func newRecordDB(session *dbr.Session) *recordDB {
	return &recordDB{session: session}
}

// recordModel 流水记录模型
type recordModel struct {
	Id            int64  `db:"id"`
	UID           string `db:"uid"`
	RecordNo      string `db:"record_no"`
	Type          int    `db:"type"`
	Amount        int64  `db:"amount"`
	BalanceBefore int64  `db:"balance_before"`
	BalanceAfter  int64  `db:"balance_after"`
	Remark        string `db:"remark"`
	RelatedId     string `db:"related_id"`
	RelatedUID    string `db:"related_uid"`
	db.BaseModel
}

// insert 插入流水记录
func (d *recordDB) insert(m *recordModel) error {
	_, err := d.session.InsertInto("wallet_record").Columns(
		"uid", "record_no", "type", "amount", "balance_before", "balance_after",
		"remark", "related_id", "related_uid",
	).Record(m).Exec()
	return err
}

// queryByUID 查询用户流水记录
func (d *recordDB) queryByUID(uid string, page, pageSize int) ([]*recordModel, error) {
	var list []*recordModel
	_, err := d.session.Select("*").From("wallet_record").
		Where("uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByUIDAndType 根据类型查询用户流水记录
func (d *recordDB) queryByUIDAndType(uid string, recordType int, page, pageSize int) ([]*recordModel, error) {
	var list []*recordModel
	_, err := d.session.Select("*").From("wallet_record").
		Where("uid=? AND type=?", uid, recordType).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// countByUID 统计用户流水数量
func (d *recordDB) countByUID(uid string) (int64, error) {
	var count int64
	_, err := d.session.Select("COUNT(*)").From("wallet_record").Where("uid=?", uid).Load(&count)
	return count, err
}

// queryAll 查询所有流水记录(管理后台)
func (d *recordDB) queryAll(page, pageSize int) ([]*recordModel, error) {
	var list []*recordModel
	_, err := d.session.Select("*").From("wallet_record").
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// countAll 统计所有流水数量
func (d *recordDB) countAll() (int64, error) {
	var count int64
	_, err := d.session.Select("COUNT(*)").From("wallet_record").Load(&count)
	return count, err
}
