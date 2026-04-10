package rtc

import (
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/gocraft/dbr/v2"
)

type roomDB struct {
	ctx     *config.Context
	session *dbr.Session
}

func newRoomDB(ctx *config.Context) *roomDB {
	return &roomDB{
		ctx:     ctx,
		session: ctx.DB(),
	}
}

func (r *roomDB) insert(m *roomModel) error {
	_, err := r.session.InsertInto("rtc_room").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	return err
}
func (r *roomDB) insertTx(m *roomModel, tx *dbr.Tx) error {
	_, err := tx.InsertInto("rtc_room").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	return err
}

func (r *roomDB) queryWithRoomID(roomID string) (*roomModel, error) {
	var m *roomModel
	_, err := r.session.Select("*").From("rtc_room").Where("room_id=?", roomID).Load(&m)
	return m, err
}

// --------------------  participant --------------------

func (r *roomDB) insertOrAddParticipantTx(m *participantModel, tx *dbr.Tx) error {
	_, err := tx.InsertBySql("insert into rtc_participant(room_id,uid,role,status) values(?,?,?,?) ON DUPLICATE KEY UPDATE role=VALUES(role),status=VALUES(status)", m.RoomID, m.UID, m.Role, m.Status).Exec()
	return err
}

func (r *roomDB) existParticipant(roomID string, uid string) (bool, error) {
	var count int
	_, err := r.session.Select("count(*)").From("rtc_participant").Where("room_id=? and uid=?", roomID, uid).Load(&count)
	return count > 0, err
}

func (r *roomDB) queryParticipants(roomID string) ([]*participantModel, error) {
	var models []*participantModel
	_, err := r.session.Select("*").From("rtc_participant").Where("room_id=?", roomID).Load(&models)
	return models, err
}

func (r *roomDB) updateParticipantStatus(status int, uid, roomID string) error {
	_, err := r.session.Update("rtc_participant").Set("status", status).Where("room_id=? and uid=?", roomID, uid).Exec()
	return err
}

func (r *roomDB) updateParticipantToJon(uid, roomID string, joinAt int64) error {
	_, err := r.session.Update("rtc_participant").Set("join_at", joinAt).Set("status", StatusJoined.Int()).Where("uid=? and room_id=?", uid, roomID).Exec()
	return err
}

// 更新参与者到挂断状态
func (r *roomDB) updateParticipantToHangup(uid, roomID string, endAt int64) error {
	_, err := r.session.Update("rtc_participant").Set("end_at", endAt).Set("status", StatusHangup.Int()).Where("uid=? and room_id=?", uid, roomID).Exec()
	return err
}

type roomModel struct {
	RoomID            string // 房间ID
	Owner             string // 房间拥有者的uid
	Name              string // 房间名称
	ParticipantCount  int    // 参与人数
	SourceChannelID   string // 从指定频道发起的房间聊天
	SourceChannelType uint8  // 频道类型
	InviteOn          int    // 是否开启邀请机制
	db.BaseModel
}

type participantModel struct {
	RoomID string // 房间ID
	UID    string // 参与人uid
	Role   string // 参与人角色  presenter 主持人(可语音可视频) viewer：观看者（不可语音不可视频） audio_only_presenter: 只能语音 video_only_viewer:只能视频
	JoinAt int64  // 加入通话时间
	EndAt  int64  // 结束通话
	Status int    // 参与人状态 0.邀请中(邀请了，还没进来) 1.已加入（正在通话中） 2.已拒绝（拒绝了邀请） 3.已挂断（已接受了邀请，结束了通话）
	db.BaseModel
}
