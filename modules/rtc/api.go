package rtc

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Report 举报
type RTC struct {
	ctx *config.Context
	log.Log
	owtClient *http.Client
	roomDB    *roomDB
}

// New 创建一个举报对象
func New(ctx *config.Context) *RTC {
	return &RTC{
		ctx:    ctx,
		Log:    log.NewTLog("RTC"),
		roomDB: newRoomDB(ctx),
		owtClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// Route 配置路由规则
func (r *RTC) Route(l *wkhttp.WKHttp) {
	v := l.Group("/v1/rtc", r.ctx.AuthMiddleware(l))
	{
		v.POST("/rooms", r.roomCreate)              // 创建房间
		v.GET("/rooms/:id/token", r.roomTokenGet)   // 获取房间token (使用prejoin接口)
		v.POST("/rooms/:id/invoke", r.roomInvoke)   // 邀请加入房间
		v.POST("/rooms/:id/prejoin", r.roomPreJoin) // 准备加入房间（主要获取加入房间的之前的一些配置，比如房间token）
		v.POST("/rooms/:id/hangup", r.roomHangup)   // 挂断了
		v.POST("/rooms/:id/refuse", r.roomRefuse)   // 拒绝接听

		v.POST("/rooms/:id/joined", r.roomJoined) // 已加入(前端加入房间成功后调用此接口告知服务器)
		// v.POST("/rooms/:id/leaved", r.roomLeave)  // 已离开

		v.POST("/p2p/invoke", r.p2pInvoke) // p2p 邀请通话
		v.POST("/p2p/accept", r.p2pAccept) // p2p 接受通话
		v.POST("/p2p/refuse", r.p2pRefuse) // p2p 拒绝通话
		v.POST("/p2p/cancel", r.p2pCancel) // p2p 取消通话
		v.POST("/p2p/hangup", r.p2pHangup) // p2p 挂断通话

		// 声网API
		v.POST("/agora/token", r.agoraToken)
	}
}

// 个人邀请通话
func (r *RTC) p2pInvoke(c *wkhttp.Context) {
	var req struct {
		ToUID    string             `json:"to_uid"`
		CallType common.RTCCallType `json:"call_type"`
	}
	if err := c.BindJSON(&req); err != nil {
		r.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if strings.TrimSpace(req.ToUID) == "" {
		c.ResponseError(errors.New("toUID不能为空！"))
		return
	}
	err := r.ctx.SendCMD(config.MsgCMDReq{
		CMD:         CMDRTCP2PInvoke,
		ChannelID:   req.ToUID,
		FromUID:     c.GetLoginUID(),
		ChannelType: common.ChannelTypePerson.Uint8(),
		Param: map[string]interface{}{
			"call_type": req.CallType,
			"from_uid":  c.GetLoginUID(),
		},
	})
	if err != nil {
		r.Error("消息发送失败！", zap.Error(err))
		c.ResponseError(errors.New("消息发送失败！"))
		return
	}
	c.ResponseOK()
}

func (r *RTC) p2pHangup(c *wkhttp.Context) {
	var req struct {
		UID      string             `json:"uid"`       // 对方uid
		Second   int                `json:"second"`    // 通话秒数
		CallType common.RTCCallType `json:"call_type"` // 通话类型
		IsCaller int                `json:"is_caller"` // 当前挂断的人是否是呼叫者 0. 否 1.是
	}
	if err := c.BindJSON(&req); err != nil {
		r.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if strings.TrimSpace(req.UID) == "" {
		c.ResponseError(errors.New("uid不能为空！"))
		return
	}
	err := r.ctx.SendCMD(config.MsgCMDReq{
		CMD:         CMDRTCP2PHangup,
		FromUID:     c.GetLoginUID(),
		ChannelID:   req.UID,
		ChannelType: common.ChannelTypePerson.Uint8(),
		Param: map[string]interface{}{
			"second": req.Second,
			"uid":    c.GetLoginUID(),
		},
	})
	if err != nil {
		r.Error("发送命令失败！", zap.Error(err))
		c.ResponseError(errors.New("发送命令失败！"))
		return
	}

	fromUID := c.GetLoginUID()
	toUID := req.UID
	if req.IsCaller == 0 {
		fromUID = req.UID
		toUID = c.GetLoginUID()
	}

	err = r.ctx.SendRTCCallResult(config.P2pRtcMessageReq{
		FromUID:    fromUID,
		ToUID:      toUID,
		CallType:   req.CallType,
		ResultType: common.RTCResultTypeHangup,
		Second:     req.Second,
	})
	if err != nil {
		r.Error("发送挂断的消息失败！", zap.Error(err))
	}

	c.ResponseOK()
}

// 取消通过
func (r *RTC) p2pCancel(c *wkhttp.Context) {
	var req struct {
		UID      string             `json:"uid"`       // 对方uid
		CallType common.RTCCallType `json:"call_type"` // 通话类型
	}
	if err := c.BindJSON(&req); err != nil {
		r.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if strings.TrimSpace(req.UID) == "" {
		c.ResponseError(errors.New("uid不能为空！"))
		return
	}
	err := r.ctx.SendCMD(config.MsgCMDReq{
		CMD:         CMDRTCP2PCancel,
		FromUID:     c.GetLoginUID(),
		ChannelID:   req.UID,
		ChannelType: common.ChannelTypePerson.Uint8(),
		Param: map[string]interface{}{
			"uid": c.GetLoginUID(),
		},
	})
	if err != nil {
		r.Error("发送命令失败！", zap.Error(err))
		c.ResponseError(errors.New("发送命令失败！"))
		return
	}
	err = r.ctx.SendRTCCallResult(config.P2pRtcMessageReq{
		FromUID:    c.GetLoginUID(),
		ToUID:      req.UID,
		CallType:   req.CallType,
		ResultType: common.RTCResultTypeCancel,
	})
	if err != nil {
		r.Error("发送rtc取消消息失败！", zap.Error(err))
	}

	c.ResponseOK()
}

// 拒绝通话
func (r *RTC) p2pRefuse(c *wkhttp.Context) {
	var req struct {
		UID      string             `json:"uid"`       // 发起用户的uid
		CallType common.RTCCallType `json:"call_type"` // 通话类型
	}
	if err := c.BindJSON(&req); err != nil {
		r.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if strings.TrimSpace(req.UID) == "" {
		c.ResponseError(errors.New("uid不能为空！"))
		return
	}
	err := r.ctx.SendCMD(config.MsgCMDReq{
		CMD:         CMDRTCP2PRefuse,
		FromUID:     c.GetLoginUID(),
		ChannelID:   req.UID,
		ChannelType: common.ChannelTypePerson.Uint8(),
		Param: map[string]interface{}{
			"uid": c.GetLoginUID(),
		},
	})
	if err != nil {
		r.Error("发送命令失败！", zap.Error(err))
		c.ResponseError(errors.New("发送命令失败！"))
		return
	}
	err = r.ctx.SendRTCCallResult(config.P2pRtcMessageReq{
		FromUID:    req.UID,
		ToUID:      c.GetLoginUID(),
		CallType:   req.CallType,
		ResultType: common.RTCResultTypeRefuse,
	})
	if err != nil {
		r.Error("发送rtc拒绝消息失败！", zap.Error(err))
	}

	c.ResponseOK()
}

// 接受通话
func (r *RTC) p2pAccept(c *wkhttp.Context) {
	var req struct {
		FromUID  string             `json:"from_uid"` // 发起通话的用户的uid
		CallType common.RTCCallType `json:"call_type"`
	}
	if err := c.BindJSON(&req); err != nil {
		r.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if strings.TrimSpace(req.FromUID) == "" {
		c.ResponseError(errors.New("FromUID不能为空！"))
		return
	}
	err := r.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		CMD:         CMDRTCP2PAccept,
		FromUID:     c.GetLoginUID(),
		ChannelID:   req.FromUID,
		ChannelType: common.ChannelTypePerson.Uint8(),
		Param: map[string]interface{}{
			"call_type": req.CallType,
			"from_uid":  c.GetLoginUID(),
		},
	})
	if err != nil {
		r.Error("消息发送失败！", zap.Error(err))
		c.ResponseError(errors.New("消息发送失败！"))
		return
	}
	c.ResponseOK()
}

func (r *RTC) agoraToken(c *wkhttp.Context) {
	// var req struct {
	// 	Role   uint16 `json:"role"`
	// 	RoomID string `json:"room_id"`
	// }
	// if err := c.BindJSON(&req); err != nil {
	// 	r.Error("数据格式有误！", zap.Error(err))
	// 	c.ResponseError(errors.New("数据格式有误！"))
	// 	return
	// }
	// expireTimeInSeconds := uint32(60 * 60 * 4)
	// currentTimestamp := uint32(time.Now().UTC().Unix())
	// expireTimestamp := currentTimestamp + expireTimeInSeconds
	// result, err := rtctokenbuilder.BuildTokenWithUID(r.ctx.GetConfig().AgoraAppID, r.ctx.GetConfig().AgoraCertificate, req.RoomID, 0, rtctokenbuilder.Role(req.Role), expireTimestamp)
	// if err != nil {
	// 	c.ResponseErrorf("构建音视频token失败！", err)
	// 	return
	// }
	// c.Response(gin.H{
	// 	"token": result,
	// })
}

// 创建房间
func (r *RTC) roomCreate(c *wkhttp.Context) {
	var req struct {
		Name        string   `json:"name"`         // 房间名称
		Uids        []string `json:"uids"`         // 参与者uids
		InviteOn    int      `json:"invite_on"`    // 是否开启邀请机制 开启后只有收到邀请才能加入房间
		ChannelID   string   `json:"channel_id"`   // 频道id 非必须
		ChannelType uint8    `json:"channel_type"` // 频道类型 非必须
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	if len(req.Uids) > r.ctx.GetConfig().OWT.RoomMaxCount {
		c.ResponseError(errors.New("房间人数不能大于9人！"))
		return
	}
	roomName := req.Name
	if strings.TrimSpace(roomName) == "" {
		roomName = "音视频房间"
	}
	roomCreateResult, err := r.post("/v1/rooms", map[string]interface{}{
		"name": roomName,
	})
	if err != nil {
		c.ResponseErrorf("创建房间失败！", err)
		return
	}
	var roomConfigMap map[string]interface{}
	if err := util.ReadJsonByByte(roomCreateResult, &roomConfigMap); err != nil {
		c.ResponseErrorf("读取创建房间返回数据失败！", err)
		return
	}
	roomID := ""
	if roomConfigMap["id"] != nil {
		roomID = roomConfigMap["id"].(string)
	}
	if strings.TrimSpace(roomID) == "" {
		c.ResponseErrorf("音频服务返回的roomID有误！", err)
		return
	}
	tx, _ := r.ctx.DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	err = r.roomDB.insertTx(&roomModel{
		Name:              roomName,
		RoomID:            roomID,
		InviteOn:          req.InviteOn,
		Owner:             c.GetLoginUID(),
		ParticipantCount:  len(req.Uids),
		SourceChannelID:   req.ChannelID,
		SourceChannelType: req.ChannelType,
	}, tx)
	if err != nil {
		tx.Rollback()
		c.ResponseErrorf("添加房间数据失败！", err)
		return
	}
	if len(req.Uids) > 0 {
		newUIDs := make([]string, 0, len(req.Uids)+1)
		newUIDs = append(newUIDs, req.Uids...)
		newUIDs = append(newUIDs, c.GetLoginUID())
		for _, uid := range newUIDs {
			err = r.roomDB.insertOrAddParticipantTx(&participantModel{
				RoomID: roomID,
				UID:    uid,
				Role:   RolePresenter.String(),
				Status: StatusInviting.Int(),
			}, tx)
			if err != nil {
				tx.Rollback()
				c.ResponseErrorf("添加参与者失败！", err)
				return
			}
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.ResponseErrorf("提交事务失败！", err)
		return
	}
	realUIDs := make([]string, 0, len(req.Uids))
	if len(req.Uids) > 0 {
		for _, uid := range req.Uids {
			if uid == c.GetLoginUID() || strings.TrimSpace(uid) == "" {
				continue
			}
			realUIDs = append(realUIDs, uid)
		}
	}

	if len(req.Uids) > 0 {
		param := map[string]interface{}{
			"room_id":      roomID,
			"inviter":      c.GetLoginUID(),
			"participants": realUIDs,
		}
		if strings.TrimSpace(req.ChannelID) != "" {
			param["channel_id"] = req.ChannelID
			param["channel_type"] = req.ChannelType
		}
		err = r.ctx.SendCMD(config.MsgCMDReq{
			Subscribers: req.Uids,
			CMD:         CMDRoomInvoke,
			Param:       param,
		})
		if err != nil {
			c.ResponseErrorf("发送CMD失败！", err)
			return
		}
	}

	c.Response(gin.H{
		"room_id": roomID,
	})

}

// 准备加入房间
func (r *RTC) roomPreJoin(c *wkhttp.Context) {
	roomID := c.Param("id")
	loginUID := c.GetLoginUID()

	roomM, err := r.roomDB.queryWithRoomID(roomID)
	if err != nil {
		c.ResponseErrorf("查询房间数据失败！", err)
		return
	}
	if roomM == nil {
		c.ResponseError(errors.New("房间不存在！"))
		return
	}

	if roomM.InviteOn == 1 {
		exist, err := r.roomDB.existParticipant(roomID, loginUID)
		if err != nil {
			c.ResponseErrorf("查询是否存在参与者失败！", err)
			return
		}
		if !exist {
			c.ResponseError(errors.New("不在邀请人之内，不允许加入"))
			return
		}
	}
	token, err := r.getRoomPresenterToken(roomID, loginUID)
	if err != nil {
		c.ResponseErrorf("获取房间token失败！", err)
		return
	}

	c.Response(gin.H{
		"token":   token,
		"room_id": roomID,
	})

}

// 挂断
func (r *RTC) roomHangup(c *wkhttp.Context) {
	roomID := c.Param("id")

	roomM, err := r.roomDB.queryWithRoomID(roomID)
	if err != nil {
		c.ResponseErrorf("查询房间失败！", err)
		return
	}
	if roomM == nil {
		c.ResponseError(errors.New("房间不存在！"))
		return
	}

	// 更新为挂断
	err = r.roomDB.updateParticipantToHangup(c.GetLoginUID(), roomID, time.Now().Unix())
	if err != nil {
		r.Error("更新参与者状态失败！", zap.Error(err))
		c.ResponseError(errors.New("更新参与者状态失败！"))
		return
	}

	participants, err := r.roomDB.queryParticipants(roomID)
	if err != nil {
		c.ResponseErrorf("查询参与者失败！", err)
		return
	}
	newParticipants := make([]string, 0, len(participants))
	if len(participants) > 0 {
		for _, participant := range participants {
			if participant.UID != c.GetLoginUID() {
				newParticipants = append(newParticipants, participant.UID)
			}
		}
	}
	if len(newParticipants) > 0 {
		err = r.ctx.SendCMD(config.MsgCMDReq{
			Subscribers: newParticipants,
			CMD:         CMDRoomHangup,
			Param: map[string]interface{}{
				"room_id":     roomID,
				"participant": c.GetLoginUID(),
			},
		})
		if err != nil {
			c.ResponseErrorf("发送CMD失败！", err)
			return
		}
	}
	c.ResponseOK()

}

// 拒绝加入
func (r *RTC) roomRefuse(c *wkhttp.Context) {
	roomID := c.Param("id")
	roomM, err := r.roomDB.queryWithRoomID(roomID)
	if err != nil {
		c.ResponseErrorf("查询房间失败！", err)
		return
	}
	if roomM == nil {
		c.ResponseError(errors.New("房间不存在！"))
		return
	}
	participants, err := r.roomDB.queryParticipants(roomID)
	if err != nil {
		c.ResponseErrorf("查询参与者失败！", err)
		return
	}

	err = r.roomDB.updateParticipantStatus(StatusRefuse.Int(), c.GetLoginUID(), roomID)
	if err != nil {
		r.Error("修改参与者状态失败！", zap.Error(err))
		c.ResponseError(errors.New("修改参与者状态失败！"))
		return
	}

	newParticipants := make([]string, 0, len(participants))
	if len(participants) > 0 {
		for _, participant := range participants {
			if participant.UID != c.GetLoginUID() {
				newParticipants = append(newParticipants, participant.UID)
			}
		}
	}
	if len(newParticipants) > 0 {
		err = r.ctx.SendCMD(config.MsgCMDReq{
			Subscribers: newParticipants,
			CMD:         CMDRoomRefuse,
			Param: map[string]interface{}{
				"room_id":     roomID,
				"participant": c.GetLoginUID(),
			},
		})
		if err != nil {
			c.ResponseErrorf("发送CMD失败！", err)
			return
		}
	}
	c.ResponseOK()
}

// 已加入房间
func (r *RTC) roomJoined(c *wkhttp.Context) {
	roomID := c.Param("id")
	loginUID := c.GetLoginUID()

	err := r.roomDB.updateParticipantToJon(loginUID, roomID, time.Now().Unix())
	if err != nil {
		c.ResponseErrorf("更新参与人状态失败！", err)
		return
	}
	c.ResponseOK()

}

// 邀请加入房间
func (r *RTC) roomInvoke(c *wkhttp.Context) {
	var req struct {
		Uids []string `json:"uids"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	roomID := c.Param("id")

	roomM, err := r.roomDB.queryWithRoomID(roomID)
	if err != nil {
		r.Error("查询房间失败！", zap.Error(err))
		c.ResponseError(errors.New("查询房间失败！"))
		return
	}
	if roomM == nil {
		c.ResponseError(errors.New("房间不存在！"))
		return
	}

	participants, err := r.roomDB.queryParticipants(roomID)
	if err != nil {
		c.ResponseErrorf("查询参与人失败！", err)
		return
	}

	inviterUIDs := util.RemoveRepeatedElement(req.Uids)

	newParticipants := make([]string, 0, len(participants)+len(req.Uids))
	if len(participants) > 0 {
		for _, participant := range participants {
			if participant.UID != c.GetLoginUID() {
				newParticipants = append(newParticipants, participant.UID)
			}
		}
	}
	if len(inviterUIDs) > 0 {
		for _, uid := range inviterUIDs {
			if strings.TrimSpace(uid) != "" {
				newParticipants = append(newParticipants, uid)
			}
		}
	}
	newParticipants = util.RemoveRepeatedElement(newParticipants)

	tx, _ := r.ctx.DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	for _, uid := range inviterUIDs {
		err := r.roomDB.insertOrAddParticipantTx(&participantModel{
			RoomID: roomID,
			UID:    uid,
			Role:   RoleAudioOnlyPresenter.String(),
			Status: StatusInviting.Int(),
		}, tx)
		if err != nil {
			tx.Rollback()
			c.ResponseErrorf("添加参与者失败！", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		tx.RollbackUnlessCommitted()
		c.ResponseErrorf("提交事务失败！", err)
		return
	}

	if len(inviterUIDs) > 0 {
		param := map[string]interface{}{
			"room_id":      roomID,
			"inviter":      c.GetLoginUID(),
			"participants": newParticipants,
		}
		if strings.TrimSpace(roomM.SourceChannelID) != "" {
			param["channel_id"] = roomM.SourceChannelID
			param["channel_type"] = roomM.SourceChannelType
		}
		err := r.ctx.SendCMD(config.MsgCMDReq{
			Subscribers: req.Uids,
			CMD:         CMDRoomInvoke,
			Param:       param,
		})
		if err != nil {
			c.ResponseErrorf("发送CMD失败！", err)
			return
		}
	}
	c.ResponseOK()

}

// TODO： 此接口作废
func (r *RTC) roomTokenGet(c *wkhttp.Context) {
	roomID := c.Param("id")

	token, err := r.getRoomPresenterToken(roomID, c.GetLoginUID())
	if err != nil {
		c.ResponseErrorf("获取room token失败！", err)
		return
	}

	c.Response(gin.H{
		"token": token,
	})
}
