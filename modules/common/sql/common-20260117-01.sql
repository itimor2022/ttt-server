-- +migrate Up

-- 添加朋友圈开关、AI智投链接、洞察链接配置字段
ALTER TABLE `app_config` ADD COLUMN `moments_off` smallint NOT NULL DEFAULT 0 COMMENT '是否隐藏朋友圈入口 1:隐藏 0:显示';
ALTER TABLE `app_config` ADD COLUMN `ai_invest_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'AI智投链接';
ALTER TABLE `app_config` ADD COLUMN `insight_url` varchar(500) NOT NULL DEFAULT '' COMMENT '洞察链接';
ALTER TABLE `app_config` ADD COLUMN `ai_invest_on` INT DEFAULT 0 COMMENT 'AI智投开关 1:开启 0:关闭';
ALTER TABLE `app_config` ADD COLUMN `insight_connection_on` INT DEFAULT 0 COMMENT '洞察连接开关 1:开启 0:关闭';
ALTER TABLE `app_config` ADD COLUMN `normal_user_can_add_friend` INT DEFAULT 1 COMMENT '普通用户是否可以添加好友 1:可以 0:不可以';