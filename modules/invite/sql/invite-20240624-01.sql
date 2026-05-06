-- +migrate Up

ALTER TABLE `invite` ADD COLUMN vercode  VARCHAR(100)  not null default '' COMMENT '加好友验证码';
