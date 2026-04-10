-- +migrate Up

-- 添加朋友圈开关、AI智投链接、洞察链接配置字段
ALTER TABLE `app_config` ADD COLUMN `moments_off` smallint NOT NULL DEFAULT 0 COMMENT '是否隐藏朋友圈入口 1:隐藏 0:显示';
ALTER TABLE `app_config` ADD COLUMN `ai_invest_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'AI智投链接';
ALTER TABLE `app_config` ADD COLUMN `insight_url` varchar(500) NOT NULL DEFAULT '' COMMENT '洞察链接';

