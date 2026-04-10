package label

import (
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/gocraft/dbr/v2"
)

// DB DB
type DB struct {
	ctx     *config.Context
	session *dbr.Session
}

// newDB New
func newDB(ctx *config.Context) *DB {
	return &DB{
		ctx:     ctx,
		session: ctx.DB(),
	}
}

// insert 添加标签
func (d *DB) insert(m *model) (int64, error) {
	result, err := d.session.InsertInto("label").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// addMembers 添加标签成员
func (d *DB) addMembers(labelID int64, userIDs []string) error {
	for _, userID := range userIDs {
		_, err := d.session.InsertInto("label_member").Columns("label_id", "user_id").Values(labelID, userID).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// removeMembers 移除标签成员
func (d *DB) removeMembers(labelID int64, userIDs []string) error {
	_, err := d.session.DeleteFrom("label_member").Where("label_id=? AND user_id IN (?)", labelID, userIDs).Exec()
	return err
}

// getMembers 获取标签成员
func (d *DB) getMembers(labelID int64) ([]string, error) {
	type member struct {
		UserID string `db:"user_id"`
	}
	var members []member
	_, err := d.session.Select("user_id").From("label_member").Where("label_id=?", labelID).Load(&members)
	if err != nil {
		return nil, err
	}
	userIDs := make([]string, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	return userIDs, nil
}

// countMembers 统计标签成员数量
func (d *DB) countMembers(labelID int64) (int, error) {
	type countResult struct {
		Count int `db:"COUNT(*)"`
	}
	var result countResult
	_, err := d.session.Select("COUNT(*)").From("label_member").Where("label_id=?", labelID).Load(&result)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// update 修改标签
func (d *DB) update(name string, uid string, id int64) error {
	_, err := d.session.Update("label").SetMap(map[string]interface{}{
		"name": name,
	}).Where("uid=? and id=?", uid, id).Exec()
	return err
}

// delete 删除
func (d *DB) delete(id int64, uid string) error {
	_, err := d.session.DeleteFrom("label").Where("id=? and uid=?", id, uid).Exec()
	return err
}

// query 标签列表
func (d *DB) query(uid string) ([]*model, error) {
	var labels []*model
	_, err := d.session.Select("*").From("label").Where("uid=?", uid).OrderDir("updated_at", false).Load(&labels)
	return labels, err
}

// queryDetail 查询标签详情
func (d *DB) queryDetail(id int64) (*model, error) {
	var label *model
	_, err := d.session.Select("*").From("label").Where("id=?", id).Load(&label)
	return label, err
}

// exists 检查标签是否存在且属于指定用户
func (d *DB) exists(id int64, uid string) (bool, error) {
	var count int
	_, err := d.session.Select("COUNT(*)").From("label").Where("id=? AND uid=?", id, uid).Load(&count)
	return count > 0, err
}

// model 标签对象
type model struct {
	UID  string //标签所属者
	Name string //标签名字
	db.BaseModel
}
