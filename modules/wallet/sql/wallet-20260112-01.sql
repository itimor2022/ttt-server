
-- +migrate Up

-- 用户钱包表
CREATE TABLE IF NOT EXISTS `wallet` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `balance` bigint NOT NULL DEFAULT 0 COMMENT '可用余额(分)',
  `frozen` bigint NOT NULL DEFAULT 0 COMMENT '冻结金额(分)',
  `total_recharge` bigint NOT NULL DEFAULT 0 COMMENT '累计充值(分)',
  `total_withdraw` bigint NOT NULL DEFAULT 0 COMMENT '累计提现(分)',
  `interest` bigint NOT NULL DEFAULT 0 COMMENT '累计利息(分)',
  `today_interest` bigint NOT NULL DEFAULT 0 COMMENT '今日利息(分)',
  `interest_rate` double NOT NULL DEFAULT 0 COMMENT '年化利率(%)',
  `real_name_status` tinyint NOT NULL DEFAULT 0 COMMENT '实名认证状态：0-未认证，1-审核中，2-已认证，3-认证失败',
  `real_name` varchar(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `id_card` varchar(20) NOT NULL DEFAULT '' COMMENT '身份证号',
  `pay_password` varchar(100) NOT NULL DEFAULT '' COMMENT '支付密码(加密)',
  `pay_password_salt` varchar(40) NOT NULL DEFAULT '' COMMENT '支付密码盐',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包表';

-- 钱包流水记录表
CREATE TABLE IF NOT EXISTS `wallet_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `record_no` varchar(64) NOT NULL DEFAULT '' COMMENT '流水号',
  `type` tinyint NOT NULL DEFAULT 0 COMMENT '类型 1:充值 2:提现 3:转账收入 4:转账支出 5:红包发出 6:红包收入 7:红包退回',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT '金额(分)',
  `balance_before` bigint NOT NULL DEFAULT 0 COMMENT '变动前余额(分)',
  `balance_after` bigint NOT NULL DEFAULT 0 COMMENT '变动后余额(分)',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `related_id` varchar(64) NOT NULL DEFAULT '' COMMENT '关联ID(订单号/红包ID/转账ID)',
  `related_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '关联用户UID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_uid_type` (`uid`, `type`),
  KEY `idx_record_no` (`record_no`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='钱包流水记录表';

-- 充值订单表
CREATE TABLE IF NOT EXISTS `recharge_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `order_no` varchar(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT '充值金额(分)',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:待审核 1:已通过 2:已拒绝',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '用户备注',
  `admin_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '审核管理员UID',
  `admin_remark` varchar(255) NOT NULL DEFAULT '' COMMENT '管理员备注',
  `audited_at` timestamp NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值订单表';

-- 提现订单表
CREATE TABLE IF NOT EXISTS `withdraw_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `order_no` varchar(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT '提现金额(分)',
  `real_name` varchar(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `bank_name` varchar(100) NOT NULL DEFAULT '' COMMENT '银行名称',
  `bank_card` varchar(50) NOT NULL DEFAULT '' COMMENT '银行卡号',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:待审核 1:已通过 2:已拒绝',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '用户备注',
  `admin_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '审核管理员UID',
  `admin_remark` varchar(255) NOT NULL DEFAULT '' COMMENT '管理员备注',
  `audited_at` timestamp NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现订单表';

-- 红包表
CREATE TABLE IF NOT EXISTS `red_packet` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `packet_no` varchar(64) NOT NULL DEFAULT '' COMMENT '红包编号',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID(群ID或个人UID)',
  `channel_type` tinyint NOT NULL DEFAULT 1 COMMENT '频道类型 1:个人 2:群',
  `type` tinyint NOT NULL DEFAULT 1 COMMENT '红包类型 1:普通红包 2:拼手气红包',
  `total_amount` bigint NOT NULL DEFAULT 0 COMMENT '总金额(分)',
  `total_count` int NOT NULL DEFAULT 0 COMMENT '总个数',
  `remain_amount` bigint NOT NULL DEFAULT 0 COMMENT '剩余金额(分)',
  `remain_count` int NOT NULL DEFAULT 0 COMMENT '剩余个数',
  `remark` varchar(100) NOT NULL DEFAULT '恭喜发财，大吉大利' COMMENT '祝福语',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:进行中 1:已抢完 2:已过期退回',
  `expire_time` timestamp NOT NULL COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_packet_no` (`packet_no`),
  KEY `idx_uid` (`uid`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_time` (`expire_time`),
  KEY `idx_status_expire` (`status`, `expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='红包表';

-- 红包领取记录表
CREATE TABLE IF NOT EXISTS `red_packet_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `packet_id` bigint NOT NULL DEFAULT 0 COMMENT '红包ID',
  `packet_no` varchar(64) NOT NULL DEFAULT '' COMMENT '红包编号',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '领取者UID',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT '领取金额(分)',
  `is_best` tinyint NOT NULL DEFAULT 0 COMMENT '是否手气最佳 0:否 1:是',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  PRIMARY KEY (`id`),
  KEY `idx_packet_id` (`packet_id`),
  KEY `idx_uid` (`uid`),
  UNIQUE KEY `uk_packet_uid` (`packet_id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='红包领取记录表';

-- 转账表
CREATE TABLE IF NOT EXISTS `transfer` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `transfer_no` varchar(64) NOT NULL DEFAULT '' COMMENT '转账编号',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '转出者UID',
  `to_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '接收者UID',
  `amount` bigint NOT NULL DEFAULT 0 COMMENT '转账金额(分)',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '转账备注',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:待接收 1:已接收 2:已退回 3:已过期退回',
  `expire_time` timestamp NOT NULL COMMENT '过期时间',
  `received_at` timestamp NULL DEFAULT NULL COMMENT '接收/退回时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_transfer_no` (`transfer_no`),
  KEY `idx_from_uid` (`from_uid`),
  KEY `idx_to_uid` (`to_uid`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_time` (`expire_time`),
  KEY `idx_status_expire` (`status`, `expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='转账表';

-- +migrate Down
DROP TABLE IF EXISTS `transfer`;
DROP TABLE IF EXISTS `red_packet_record`;
DROP TABLE IF EXISTS `red_packet`;
DROP TABLE IF EXISTS `withdraw_order`;
DROP TABLE IF EXISTS `recharge_order`;
DROP TABLE IF EXISTS `wallet_record`;
DROP TABLE IF EXISTS `wallet`;

