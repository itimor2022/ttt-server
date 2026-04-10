package rtc

type Role string

const (
	// RolePresenter 主持人(可语音可视频)
	RolePresenter Role = "presenter"
	// RoleViewer：观看者（不可语音不可视频）
	RoleViewer Role = "viewer"
	// RoleAudioOnlyPresenter 只能语音
	RoleAudioOnlyPresenter Role = "audio_only_presenter"
	// RoleVideoOnlyViewer 只能视频
	RoleVideoOnlyViewer Role = "video_only_viewer"
)

func (r Role) String() string {
	return string(r)
}

type Status int

const (
	// StatusInviting  邀请中(邀请了，还没进来)
	StatusInviting Status = 0
	// StatusJoined 已加入（正在通话中）
	StatusJoined Status = 1
	// StatusRefuse 已拒绝（拒绝了邀请
	StatusRefuse Status = 2
	// StatusHangup 已挂断（已接受了邀请，结束了通话）
	StatusHangup Status = 3
)

func (s Status) Int() int {
	return int(s)
}

const (
	// CMDRoomInvoke 加入房间邀请
	CMDRoomInvoke = "room.invoke"
	// CMDRoomHangup 挂断
	CMDRoomHangup = "room.hangup"
	// CMDRoomRefuse 拒绝
	CMDRoomRefuse = "room.refuse"
	// CMDRoomLeave 离开
	CMDRoomLeave = "room.leave"
	// rtc p2p通话邀请
	CMDRTCP2PInvoke = "rtc.p2p.invoke"
	// rtc p2p通话接受
	CMDRTCP2PAccept = "rtc.p2p.accept"
	// rtc p2p 拒绝
	CMDRTCP2PRefuse = "rtc.p2p.refuse"
	// rtc p2p 取消
	CMDRTCP2PCancel = "rtc.p2p.cancel"
	// rtc p2p 挂断通话
	CMDRTCP2PHangup = "rtc.p2p.hangup"
)
