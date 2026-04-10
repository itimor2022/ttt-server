-- +migrate Up


-- 视频房间
create table IF NOT EXISTS  `rtc_room`
(
    id integer PRIMARY KEY AUTO_INCREMENT,
    room_id VARCHAR(40)  not null DEFAULT '' comment '房间id',
    owner VARCHAR(40)  not null DEFAULT '' comment '房间拥有者的uid',
    name VARCHAR(40)  not null DEFAULT '' comment '房间名',
    participant_count  integer not null DEFAULT 0 comment '参与者数量',
    source_channel_id  VARCHAR(40)  not null DEFAULT '' comment '从指定频道发起的房间聊天',
    source_channel_type smallint  not null DEFAULT 0 comment '频道类型',
    invite_on   smallint not null DEFAULT 0 comment '是否开启邀请机制 0.否 1.是 开启后只有收到邀请才能加入房间',
    created_at timeStamp    not null DEFAULT CURRENT_TIMESTAMP,
    updated_at timeStamp    not null DEFAULT CURRENT_TIMESTAMP 
);
CREATE UNIQUE INDEX  room_id on `rtc_room` (room_id);


-- 房间参与者

CREATE TABLE IF NOT EXISTS `rtc_participant`
(
    id integer PRIMARY KEY AUTO_INCREMENT,
    room_id VARCHAR(40)  not null DEFAULT '' comment '房间id',
    uid   VARCHAR(40)  not null DEFAULT '' comment '参与者的uid',
    role   VARCHAR(40)  not null DEFAULT '' comment '参与者角色  presenter 主持人(可语音可视频) viewer：观看者（不可语音不可视频） audio_only_presenter: 只能语音 video_only_viewer:只能视频',
    status  smallint   not null DEFAULT 0 comment '0.邀请中(邀请了，还没进来) 1.已加入（正在通话中） 2.已拒绝（拒绝了邀请） 3.已挂断（已接受了邀请，结束了通话）',
    created_at timeStamp    not null DEFAULT CURRENT_TIMESTAMP,
    updated_at timeStamp    not null DEFAULT CURRENT_TIMESTAMP 

);
CREATE  INDEX  room_id_idx on `rtc_participant` (room_id);