
-- +migrate Up

CREATE UNIQUE INDEX  room_uid_id_idx on `rtc_participant` (room_id,uid);

ALTER TABLE `rtc_participant` ADD COLUMN `join_at`  integer  not null default 0  COMMENT '通话加入时间';
ALTER TABLE `rtc_participant` ADD COLUMN `end_at`  integer  not null default 0  COMMENT '通话结束时间';