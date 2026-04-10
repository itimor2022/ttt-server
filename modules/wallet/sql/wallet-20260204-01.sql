-- +migrate Up

-- 实名认证表
CREATE TABLE IF NOT EXISTS `wallet_real_name` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `real_name` varchar(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `id_card` varchar(20) NOT NULL DEFAULT '' COMMENT '身份证号',
  `phone` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号码',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0-待审核，1-已认证，2-认证失败',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '审核备注',
  `admin_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '审核管理员UID',
  `audited_at` timestamp NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='实名认证表';

-- 修改钱包记录类型，添加利息类型
ALTER TABLE `wallet_record` MODIFY COLUMN `type` tinyint NOT NULL DEFAULT 0 COMMENT '类型 1:充值 2:提现 3:转账收入 4:转账支出 5:红包发出 6:红包收入 7:红包退回 8:利息收入';

-- +migrate Down
DROP TABLE IF EXISTS `wallet_real_name`;
ALTER TABLE `wallet_record` MODIFY COLUMN `type` tinyint NOT NULL DEFAULT 0 COMMENT '类型 1:充值 2:提现 3:转账收入 4:转账支出 5:红包发出 6:红包收入 7:红包退回';
