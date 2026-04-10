package message

import (
	"fmt"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/gocraft/dbr/v2"
)

// managerDB 管理员代发消息记录
type managerDB struct {
	session *dbr.Session
	db      *DB
	ctx     *config.Context
}

// newManagerDB newManagerDB
func newManagerDB(ctx *config.Context) *managerDB {
	return &managerDB{
		session: ctx.DB(),
		db:      NewDB(ctx),
		ctx:     ctx,
	}
}

// 添加一条发送消息记录
func (m *managerDB) insertMsgHistory(message *managerMsgModel) error {
	_, err := m.session.InsertInto("send_history").Columns(util.AttrToUnderscore(message)...).Record(message).Exec()
	return err
}

// 查询代发消息记录
func (m *managerDB) queryMsgWithPage(pageSize, page uint64) ([]*managerMsgModel, error) {
	var list []*managerMsgModel
	_, err := m.session.Select("*").From("send_history").Offset((page-1)*pageSize).Limit(pageSize).OrderDir("created_at", false).Load(&list)
	return list, err
}

// 查询消息数量
func (m *managerDB) queryMsgCount() (int64, error) {
	var count int64
	_, err := m.session.Select("count(*)").From("send_history").Load(&count)
	return count, err
}

func (m *managerDB) queryWithChannelID(channelID string, page, pageSize uint64) ([]*messageModel, error) {
	var list []*messageModel
	var table = m.db.getTable(channelID)
	_, err := m.session.Select("*").From(table).Where("channel_id=?", channelID).Offset((page-1)*pageSize).Limit(pageSize).OrderDir("created_at", false).Load(&list)
	return list, err
}

func (m *managerDB) queryRecordCount(channelID string) (int64, error) {
	var count int64
	_, err := m.session.Select("count(*)").From(m.db.getTable(channelID)).Where("channel_id=?", channelID).Load(&count)
	return count, err
}

func (m *managerDB) queryMsgExtrWithMsgIds(msgIds []string) ([]*messageExtraModel, error) {
	var list []*messageExtraModel
	_, err := m.session.Select("*").From("message_extra").Where("message_id in ?", msgIds).Load(&list)
	return list, err
}

func (m *managerDB) queryProhibitWordsWithContent(content string) (*prohibitWordsModel, error) {
	var model *prohibitWordsModel
	_, err := m.session.Select("*").From("prohibit_words").Where("content=?", content).Load(&model)
	return model, err
}

func (m *managerDB) queryProhibitWordsWithID(id int) (*prohibitWordsModel, error) {
	var model *prohibitWordsModel
	_, err := m.session.Select("*").From("prohibit_words").Where("id=?", id).Load(&model)
	return model, err
}

func (m *managerDB) updateProhibitWord(word *prohibitWordsModel) error {
	_, err := m.session.Update("prohibit_words").SetMap(map[string]interface{}{
		"version":    word.Version,
		"is_deleted": word.IsDeleted,
	}).Where("content=?", word.Content).Exec()
	return err
}
func (m *managerDB) insertProhibitWord(word *prohibitWordsModel) error {
	_, err := m.session.InsertInto("prohibit_words").Columns(util.AttrToUnderscore(word)...).Record(word).Exec()
	return err
}
func (m *managerDB) queryProhibitWords(pageIndex, pageSize uint64) ([]*prohibitWordsModel, error) {
	var list []*prohibitWordsModel
	_, err := m.session.Select("*").From("prohibit_words").Offset((pageIndex-1)*pageSize).Limit(pageSize).OrderDir("created_at", false).Load(&list)
	return list, err
}

func (m *managerDB) queryProhibitWordsWithContentAndPage(content string, pageIndex, pageSize uint64) ([]*prohibitWordsModel, error) {
	var list []*prohibitWordsModel
	_, err := m.session.Select("*").From("prohibit_words").Where("content like ?", "%"+content+"%").Offset((pageIndex-1)*pageSize).Limit(pageSize).OrderDir("created_at", false).Load(&list)
	return list, err
}

func (m *managerDB) queryProhibitWordsCount() (int64, error) {
	var count int64
	_, err := m.session.Select("count(*)").From("prohibit_words").Load(&count)
	return count, err
}

func (m *managerDB) queryProhibitWordsCountWithContent(content string) (int64, error) {
	var count int64
	_, err := m.session.Select("count(*)").From("prohibit_words").Where("content like ?", "%"+content+"%").Load(&count)
	return count, err
}

func (m *managerDB) updateMsgExtraVersionAndDeletedTx(md *messageExtraModel, tx *dbr.Tx) error {
	_, err := tx.InsertBySql("INSERT INTO message_extra (message_id,message_seq,channel_id,channel_type,is_deleted,version) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE is_deleted=VALUES(is_deleted),version=VALUES(version)", md.MessageID, md.MessageSeq, md.ChannelID, md.ChannelType, md.IsDeleted, md.Version).Exec()
	return err
}

type prohibitWordsModel struct {
	Content   string
	IsDeleted int
	Version   int64
	db.BaseModel
}

// 管理员代发消息记录
type managerMsgModel struct {
	Receiver            string // 接受者uid
	ReceiverName        string // 接受者名字
	ReceiverChannelType int    // 接受者频道类型
	Sender              string // 发送者uid
	SenderName          string // 发送者名字
	HandlerUID          string // 操作者uid
	HandlerName         string // 操作者名字
	Content             string // 发送内容
	db.BaseModel
}

// 构建所有消息表的UNION ALL查询
func (m *managerDB) buildAllTablesUnion() string {
	tableCount := m.ctx.GetConfig().TablePartitionConfig.MessageTableCount
	tables := make([]string, 0, tableCount)
	tables = append(tables, "SELECT channel_id, channel_type, created_at FROM message")
	for i := 1; i < tableCount; i++ {
		tables = append(tables, fmt.Sprintf("SELECT channel_id, channel_type, created_at FROM message%d", i))
	}
	return strings.Join(tables, " UNION ALL ")
}

// 查询有消息记录的频道列表
func (m *managerDB) queryChannelList(pageIndex, pageSize uint64, channelType string, keyword string) ([]*channelStatModel, error) {
	var list []*channelStatModel

	// 构建UNION ALL查询所有消息表
	unionQuery := m.buildAllTablesUnion()

	whereClause := ""
	args := make([]interface{}, 0)

	if channelType != "" || keyword != "" {
		conditions := make([]string, 0)
		if channelType != "" {
			conditions = append(conditions, "channel_type = ?")
			args = append(args, channelType)
		}
		if keyword != "" {
			conditions = append(conditions, "channel_id LIKE ?")
			args = append(args, "%"+keyword+"%")
		}
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 分页参数
	args = append(args, pageSize, (pageIndex-1)*pageSize)

	sql := fmt.Sprintf(`
		SELECT channel_id, channel_type, COUNT(*) as message_count, MAX(created_at) as last_msg_time 
		FROM (%s) as all_messages 
		%s
		GROUP BY channel_id, channel_type 
		ORDER BY last_msg_time DESC 
		LIMIT ? OFFSET ?
	`, unionQuery, whereClause)

	_, err := m.session.SelectBySql(sql, args...).Load(&list)
	return list, err
}

// 查询频道总数
func (m *managerDB) queryChannelCount(channelType string, keyword string) (int64, error) {
	var count int64

	// 构建UNION ALL查询所有消息表
	unionQuery := m.buildAllTablesUnion()

	whereClause := ""
	args := make([]interface{}, 0)

	if channelType != "" || keyword != "" {
		conditions := make([]string, 0)
		if channelType != "" {
			conditions = append(conditions, "channel_type = ?")
			args = append(args, channelType)
		}
		if keyword != "" {
			conditions = append(conditions, "channel_id LIKE ?")
			args = append(args, "%"+keyword+"%")
		}
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	sql := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT DISTINCT channel_id, channel_type 
			FROM (%s) as all_messages 
			%s
		) as t
	`, unionQuery, whereClause)

	err := m.session.SelectBySql(sql, args...).LoadOne(&count)
	return count, err
}

// 频道统计信息
type channelStatModel struct {
	ChannelID    string `db:"channel_id"`
	ChannelType  uint8  `db:"channel_type"`
	MessageCount int64  `db:"message_count"`
	LastMsgTime  string `db:"last_msg_time"`
}