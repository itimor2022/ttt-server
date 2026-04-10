package label

import (
	"errors"
	"strconv"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/user"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Label 标签
type Label struct {
	ctx *config.Context
	log.Log
	db          *DB
	userService user.IService
}

// New  New
func New(ctx *config.Context) *Label {
	return &Label{
		ctx:         ctx,
		db:          newDB(ctx),
		Log:         log.NewTLog("label"),
		userService: user.NewService(ctx),
	}
}

// Route 路由配置
func (l *Label) Route(r *wkhttp.WKHttp) {
	labels := r.Group("/api/labels", l.ctx.AuthMiddleware(r))

	{
		labels.POST("", l.add)                            // 添加标签
		labels.DELETE("/:id", l.delete)                   //删除标签
		labels.PUT("/:id", l.update)                      //修改标签
		labels.GET("", l.list)                            //标签列表
		labels.GET("/:id", l.detail)                      //标签详情
		labels.POST("/:id/members", l.addMembers)         //添加标签成员
		labels.DELETE("/:id/members", l.removeMembers)    //移除标签成员
		labels.GET("/internal/users", l.GetInternalUsers) //内部用户列表
	}
}

// add 添加标签
func (l *Label) add(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	var req createLabelReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		c.ResponseError(errors.New("标签名字不能为空"))
		return
	}

	label := &model{
		UID:  loginUID,
		Name: req.Name,
	}
	labelID, err := l.db.insert(label)
	if err != nil {
		l.Error("添加标签失败", zap.Error(err))
		c.ResponseError(errors.New("添加标签失败"))
		return
	}

	// 添加成员
	if len(req.Members) > 0 {
		err = l.db.addMembers(labelID, req.Members)
		if err != nil {
			l.Error("添加标签成员失败", zap.Error(err))
			c.ResponseError(errors.New("添加标签成员失败"))
			return
		}
	}

	count, _ := l.db.countMembers(labelID)
	c.Response(map[string]interface{}{
		"id":          labelID,
		"name":        req.Name,
		"memberCount": count,
	})
}

// addMembers 添加标签成员
func (l *Label) addMembers(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	id := c.Param("id")
	labelID, _ := strconv.ParseInt(id, 10, 64)

	// 检查标签是否存在且属于当前用户
	exists, err := l.db.exists(labelID, loginUID)
	if err != nil || !exists {
		c.ResponseError(errors.New("标签不存在"))
		return
	}

	var req membersReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if len(req.Members) == 0 {
		c.ResponseError(errors.New("成员列表不能为空"))
		return
	}

	err = l.db.addMembers(labelID, req.Members)
	if err != nil {
		l.Error("添加标签成员失败", zap.Error(err))
		c.ResponseError(errors.New("添加标签成员失败"))
		return
	}

	count, _ := l.db.countMembers(labelID)
	c.Response(map[string]interface{}{
		"memberCount": count,
	})
}

// removeMembers 移除标签成员
func (l *Label) removeMembers(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	id := c.Param("id")
	labelID, _ := strconv.ParseInt(id, 10, 64)

	// 检查标签是否存在且属于当前用户
	exists, err := l.db.exists(labelID, loginUID)
	if err != nil || !exists {
		c.ResponseError(errors.New("标签不存在"))
		return
	}

	var req membersReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if len(req.Members) == 0 {
		c.ResponseError(errors.New("成员列表不能为空"))
		return
	}

	err = l.db.removeMembers(labelID, req.Members)
	if err != nil {
		l.Error("移除标签成员失败", zap.Error(err))
		c.ResponseError(errors.New("移除标签成员失败"))
		return
	}

	count, _ := l.db.countMembers(labelID)
	c.Response(map[string]interface{}{
		"memberCount": count,
	})
}

// GetInternalUsers 获取内部用户列表
func (l *Label) GetInternalUsers(c *wkhttp.Context) {
	// 获取内部用户（非普通用户）
	users, err := l.userService.GetUsersWithCategories([]string{"admin", "system", "official"})
	if err != nil {
		l.Error("获取内部用户失败", zap.Error(err))
		c.ResponseError(errors.New("获取内部用户失败"))
		return
	}
	c.Response(users)
}

// delete 删除标签
func (l *Label) delete(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	id := c.Param("id")
	labelID, _ := strconv.ParseInt(id, 10, 64)
	err := l.db.delete(labelID, loginUID)
	if err != nil {
		l.Error("删除标签失败", zap.Error(err))
		c.ResponseError(errors.New("删除标签失败"))
		return
	}
	c.ResponseOK()
}

// list 标签列表
func (l *Label) list(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	labels, err := l.db.query(loginUID)
	if err != nil {
		l.Error("查询用户标签列表失败", zap.Error(err))
		c.ResponseError(errors.New("查询用户标签列表失败"))
		return
	}
	list := make([]map[string]interface{}, 0)
	for _, label := range labels {
		members, _ := l.db.getMembers(label.Id)
		count, _ := l.db.countMembers(label.Id)
		list = append(list, map[string]interface{}{
			"id":          label.Id,
			"name":        label.Name,
			"memberCount": count,
			"members":     members,
		})
	}
	c.Response(list)
}

// detail 标签详情
func (l *Label) detail(c *wkhttp.Context) {
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		c.ResponseError(errors.New("标签ID不能为空"))
		return
	}
	labelID, _ := strconv.ParseInt(id, 10, 64)
	label, err := l.db.queryDetail(labelID)
	if err != nil {
		c.ResponseError(errors.New("标签不存在"))
		return
	}

	members, _ := l.db.getMembers(labelID)
	count, _ := l.db.countMembers(labelID)

	resp := map[string]interface{}{
		"id":          label.Id,
		"name":        label.Name,
		"memberCount": count,
		"members":     members,
	}
	c.Response(resp)
}

// update 修改标签
func (l *Label) update(c *wkhttp.Context) {
	id := c.Param("id")
	loginUID := c.MustGet("uid").(string)
	var req updateLabelReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		c.ResponseError(errors.New("标签名字不能为空"))
		return
	}

	labelID, _ := strconv.ParseInt(id, 10, 64)

	// 检查标签是否存在且属于当前用户
	exists, err := l.db.exists(labelID, loginUID)
	if err != nil || !exists {
		c.ResponseError(errors.New("标签不存在"))
		return
	}

	// 更新标签名称
	err = l.db.update(req.Name, loginUID, labelID)
	if err != nil {
		l.Error("修改标签失败", zap.Error(err))
		c.ResponseError(errors.New("修改标签失败"))
		return
	}

	// 更新成员列表
	if len(req.Members) > 0 {
		// 先清空现有成员
		existingMembers, _ := l.db.getMembers(labelID)
		if len(existingMembers) > 0 {
			_ = l.db.removeMembers(labelID, existingMembers)
		}
		// 添加新成员
		err = l.db.addMembers(labelID, req.Members)
		if err != nil {
			l.Error("更新标签成员失败", zap.Error(err))
			c.ResponseError(errors.New("更新标签成员失败"))
			return
		}
	}

	count, _ := l.db.countMembers(labelID)
	c.Response(map[string]interface{}{
		"id":          labelID,
		"name":        req.Name,
		"memberCount": count,
	})
}

// createLabelReq 创建标签请求
type createLabelReq struct {
	Name    string   `json:"name"`    //标签名称
	Members []string `json:"members"` //成员用户ID
}

// updateLabelReq 更新标签请求
type updateLabelReq struct {
	Name    string   `json:"name"`    //标签名称
	Members []string `json:"members"` //成员用户ID
}

// membersReq 成员操作请求
type membersReq struct {
	Members []string `json:"members"` //成员用户ID
}

