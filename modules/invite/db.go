package invite

import (
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	dba "github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/db"
	"github.com/gocraft/dbr/v2"
)

type db struct {
	session *dbr.Session
}

func newDB(ctx *config.Context) *db {
	return &db{
		session: ctx.DB(),
	}
}
func (d *db) queryWithVercode(vercode string) (*InviteModel, error) {
	var model *InviteModel
	_, err := d.session.Select("*").From("invite").Where("vercode=?", vercode).Load(&model)
	return model, err
}

func (d *db) getInviteWithCode(inviteCode string) (*InviteModel, error) {
	var model *InviteModel
	_, err := d.session.Select("*").From("invite").Where("invite_code=?", inviteCode).Load(&model)
	return model, err
}

func (d *db) getInviteWithUID(uid string) (*InviteModel, error) {
	var model *InviteModel
	_, err := d.session.Select("*").From("invite").Where("uid=?", uid).Load(&model)
	return model, err
}
func (d *db) insertInvite(m *InviteModel) error {
	_, err := d.session.InsertInto("invite").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	return err
}

func (d *db) updateInvite(m *InviteModel) error {
	_, err := d.session.Update("invite").SetMap(map[string]interface{}{
		"status":      m.Status,
		"invite_code": m.InviteCode,
	}).Where("uid=?", m.UID).Exec()
	return err
}

type InviteModel struct {
	UID          string
	InviteCode   string
	BeInviteUID  string
	BeInviteCode string
	Status       int
	Vercode      string
	dba.BaseModel
}
