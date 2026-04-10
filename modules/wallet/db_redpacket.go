package wallet

import (
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

// 红包类型
const (
	RedPacketTypeNormal = 1 // 普通红
	RedPacketTypeLucky  = 2 // 拼手气红
)

// 红包状
const (
	RedPacketStatusActive   = 0 // 进行
	RedPacketStatusFinished = 1 // 已抢
	RedPacketStatusExpired  = 2 // 已过期退
)

// redPacketDB 红包数据库操
type redPacketDB struct {
	session *dbr.Session
}

func newRedPacketDB(session *dbr.Session) *redPacketDB {
	return &redPacketDB{session: session}
}

// redPacketModel 红包模型
type redPacketModel struct {
	Id           int64     `db:"id"`
	PacketNo     string    `db:"packet_no"`
	UID          string    `db:"uid"`
	ChannelId    string    `db:"channel_id"`
	ChannelType  int       `db:"channel_type"`
	Type         int       `db:"type"`
	TotalAmount  int64     `db:"total_amount"`
	TotalCount   int       `db:"total_count"`
	RemainAmount int64     `db:"remain_amount"`
	RemainCount  int       `db:"remain_count"`
	Remark       string    `db:"remark"`
	Status       int       `db:"status"`
	ExpireTime   time.Time `db:"expire_time"`
	db.BaseModel
}

// insert 创建红包
func (d *redPacketDB) insert(m *redPacketModel) (int64, error) {
	result, err := d.session.InsertInto("red_packet").Columns(
		"packet_no", "uid", "channel_id", "channel_type", "type",
		"total_amount", "total_count", "remain_amount", "remain_count",
		"remark", "status", "expire_time",
	).Record(m).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// queryByPacketNo 根据红包编号查询
func (d *redPacketDB) queryByPacketNo(packetNo string) (*redPacketModel, error) {
	var model redPacketModel
	_, err := d.session.Select("*").From("red_packet").Where("packet_no=?", packetNo).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.PacketNo == "" {
		return nil, nil
	}
	return &model, nil
}

// queryByID 根据ID查询
func (d *redPacketDB) queryByID(id int64) (*redPacketModel, error) {
	var model redPacketModel
	_, err := d.session.Select("*").From("red_packet").Where("id=?", id).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.Id == 0 {
		return nil, nil
	}
	return &model, nil
}

// queryExpired 查询已过期未处理的红
func (d *redPacketDB) queryExpired() ([]*redPacketModel, error) {
	var list []*redPacketModel
	now := time.Now()
	_, err := d.session.Select("*").From("red_packet").
		Where("status=? AND expire_time<?", RedPacketStatusActive, now).
		Load(&list)
	return list, err
}

// grab 抢红原子操作)
func (d *redPacketDB) grab(packetNo string, amount int64) error {
	_, err := d.session.UpdateBySql(
		"UPDATE red_packet SET remain_amount = remain_amount - ?, remain_count = remain_count - 1 WHERE packet_no = ? AND remain_count > 0 AND remain_amount >= ?",
		amount, packetNo, amount,
	).Exec()
	return err
}

// updateStatus 更新红包状
func (d *redPacketDB) updateStatus(packetNo string, status int) error {
	_, err := d.session.Update("red_packet").Set("status", status).Where("packet_no=?", packetNo).Exec()
	return err
}

// queryByUID 查询用户发出的红
func (d *redPacketDB) queryByUID(uid string, page, pageSize int) ([]*redPacketModel, error) {
	var list []*redPacketModel
	_, err := d.session.Select("*").From("red_packet").
		Where("uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByStatus 按状态查询红
func (d *redPacketDB) queryByStatus(status int, page, pageSize int) ([]*redPacketModel, error) {
	var list []*redPacketModel
	_, err := d.session.Select("*").From("red_packet").
		Where("status=?", status).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryAll 查询所有红
func (d *redPacketDB) queryAll(page, pageSize int) ([]*redPacketModel, error) {
	var list []*redPacketModel
	_, err := d.session.Select("*").From("red_packet").
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// queryByPacketNoForUpdate 根据红包编号查询(带行
func (d *redPacketDB) queryByPacketNoForUpdate(tx *dbr.Tx, packetNo string) (*redPacketModel, error) {
	var model redPacketModel
	_, err := tx.SelectBySql("SELECT * FROM red_packet WHERE packet_no = ? FOR UPDATE", packetNo).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.PacketNo == "" {
		return nil, nil
	}
	return &model, nil
}

// updateRemainTx 更新剩余金额和数事务
func (d *redPacketDB) updateRemainTx(tx *dbr.Tx, packetNo string, remainAmount int64, remainCount int) error {
	_, err := tx.Update("red_packet").
		Set("remain_amount", remainAmount).
		Set("remain_count", remainCount).
		Where("packet_no = ?", packetNo).Exec()
	return err
}

// updateStatusTx 更新红包状事务
func (d *redPacketDB) updateStatusTx(tx *dbr.Tx, packetNo string, status int) error {
	_, err := tx.Update("red_packet").Set("status", status).Where("packet_no=?", packetNo).Exec()
	return err
}

// redPacketRecordDB 红包领取记录数据库操
type redPacketRecordDB struct {
	session *dbr.Session
}

func newRedPacketRecordDB(session *dbr.Session) *redPacketRecordDB {
	return &redPacketRecordDB{session: session}
}

// redPacketRecordModel 红包领取记录模型
type redPacketRecordModel struct {
	Id       int64  `db:"id"`
	PacketId int64  `db:"packet_id"`
	PacketNo string `db:"packet_no"`
	UID      string `db:"uid"`
	Amount   int64  `db:"amount"`
	IsBest   int    `db:"is_best"`
	db.BaseModel
}

// insert 插入领取记录
func (d *redPacketRecordDB) insert(m *redPacketRecordModel) error {
	_, err := d.session.InsertInto("red_packet_record").Columns(
		"packet_id", "packet_no", "uid", "amount", "is_best",
	).Record(m).Exec()
	return err
}

// queryByPacketId 查询红包的领取记
func (d *redPacketRecordDB) queryByPacketId(packetId int64) ([]*redPacketRecordModel, error) {
	var list []*redPacketRecordModel
	_, err := d.session.Select("*").From("red_packet_record").
		Where("packet_id=?", packetId).
		OrderAsc("created_at").
		Load(&list)
	return list, err
}

// queryByPacketIdAndUID 查询用户是否领取过该红包
func (d *redPacketRecordDB) queryByPacketIdAndUID(packetId int64, uid string) (*redPacketRecordModel, error) {
	var model redPacketRecordModel
	_, err := d.session.Select("*").From("red_packet_record").
		Where("packet_id=? AND uid=?", packetId, uid).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.Id == 0 {
		return nil, nil
	}
	return &model, nil
}

// queryByUID 查询用户领取的红包记
func (d *redPacketRecordDB) queryByUID(uid string, page, pageSize int) ([]*redPacketRecordModel, error) {
	var list []*redPacketRecordModel
	_, err := d.session.Select("*").From("red_packet_record").
		Where("uid=?", uid).
		OrderDesc("created_at").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		Load(&list)
	return list, err
}

// updateBest 更新手气最
func (d *redPacketRecordDB) updateBest(packetId int64, uid string) error {
	_, err := d.session.Update("red_packet_record").
		Set("is_best", 1).
		Where("packet_id=? AND uid=?", packetId, uid).Exec()
	return err
}

// queryByPacketIdAndUIDTx 查询用户是否领取过该红包(事务
func (d *redPacketRecordDB) queryByPacketIdAndUIDTx(tx *dbr.Tx, packetId int64, uid string) (*redPacketRecordModel, error) {
	var model redPacketRecordModel
	_, err := tx.Select("*").From("red_packet_record").
		Where("packet_id=? AND uid=?", packetId, uid).Load(&model)
	if err != nil {
		return nil, err
	}
	if model.Id == 0 {
		return nil, nil
	}
	return &model, nil
}

// insertTx 插入领取记录(事务
func (d *redPacketRecordDB) insertTx(tx *dbr.Tx, m *redPacketRecordModel) error {
	_, err := tx.InsertInto("red_packet_record").Columns(
		"packet_id", "packet_no", "uid", "amount", "is_best",
	).Record(m).Exec()
	return err
}
