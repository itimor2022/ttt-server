-- 创建数据库
CREATE DATABASE IF NOT EXISTS test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE test;

-- 应用表
CREATE TABLE IF NOT EXISTS `app` (
  app_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'app id',
  app_key VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'app key',
  status  integer   NOT NULL DEFAULT 0 COMMENT '状态 0.禁用 1.可用',
  created_at timestamp     not null DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp     not null DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS app_id on `app` (app_id);
INSERT INTO `app`(app_id,app_key,status) VALUES('wukongchat',substring(MD5(RAND()),1,20),1);

-- 用户表
CREATE TABLE IF NOT EXISTS `user` (
  id         integer      not null primary key AUTO_INCREMENT,
  uid        VARCHAR(40)  not null default '',                             -- 用户唯一ID
  name       VARCHAR(100) not null default '',                             -- 用户的名字
  short_no   VARCHAR(40)  not null default '',                             -- 短编码
  short_status smallint   not null default 0,                              -- 短编码 0.未修改 1.已修改
  sex        smallint     not null default 0,                              -- 性别 0.女 1.男
  robot      smallint     not null default 0,                              -- 机器人 0.否1.是
  category   VARCHAR(40)  not null default '',                             -- 用户分类  service:客服
  role       VARCHAR(40)  not null default '',                             -- 用户角色  admin:管理员 superAdmin
  username   VARCHAR(40)  not null default '',                             -- 用户名
  password   VARCHAR(40)  not null default '',                             -- 密码
  zone       VARCHAR(40)  not null default '',                             -- 手机区号
  phone      VARCHAR(20)  not null default '',                             -- 手机号
  chat_pwd   VARCHAR(40)  not null default '',                             -- 聊天密码
  lock_screen_pwd varchar(40) not null default '',                         -- 锁屏密码
  lock_after_minute integer  not null default 0,                           -- 在几分钟后锁屏 0 表示立即
  vercode    VARCHAR(100) not null default '',                             -- 验证码 加好友来源
  is_upload_avatar        smallint not null default 0,                     -- 是否上传过头像 1:上传0:未上传
  qr_vercode VARCHAR(100) not null default '',                             -- 二维码验证码 加好友来源
  device_lock           smallint     not null DEFAULT 0,                   -- 是否开启设备锁
  search_by_phone       smallint     not null default 1,                   -- 是否可用通过手机号搜索到本人0.否1.是
  search_by_short       smallint     not null default 1,                   -- 是否可以通过短编号搜索0.否1.是
  new_msg_notice        smallint     not null default 1,                   -- 新消息通知0.否1.是
  msg_show_detail       smallint     not null default 1,                   -- 新消息通知详情0.否1.是
  voice_on              smallint     not null default 1,                   -- 是否开启声音0.否1.是
  shock_on              smallint     not null default 1,                   -- 是否开启震动0.否1.是
  mute_of_app           smallint     not null default 0,                   -- app是否禁音（当pc登录的时候app可以设置禁音，当pc登录后有效）
  offline_protection    smallint     not null default 0,                   -- 离线保护，断网屏保
  `version`    bigint     not null DEFAULT 0,                                -- 数据版本 
  status smallint       not null DEFAULT 1,                                -- 用户状态 0.禁用 1.可用
  bench_no        VARCHAR(40)     not null default '',                     -- 性能测试批次号，性能测试幂等用
  avatar         VARCHAR(255)    not null default '',                     -- 用户头像
  created_at timestamp     not null DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at timestamp     not null DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);
CREATE UNIQUE INDEX IF NOT EXISTS uid on `user` (uid);
CREATE UNIQUE INDEX IF NOT EXISTS short_no_udx on `user` (short_no);

-- 创建系统账号
INSERT INTO `user` (uid,name,short_no,phone,zone,search_by_phone,search_by_short,new_msg_notice,voice_on,shock_on,msg_show_detail,status,is_upload_avatar,category,robot) VALUES ('u_10000','系统账号',10000,'13000000000','0086',0,0,0,0,0,0,1,1,'system',1);
INSERT INTO `user` (uid,name,short_no,phone,zone,search_by_phone,search_by_short,new_msg_notice,voice_on,shock_on,msg_show_detail,status,is_upload_avatar,category) VALUES ('fileHelper','文件传输助手',20000,'13000000001','0086',0,0,0,0,0,0,1,1,'system');

-- 用户设置
CREATE TABLE IF NOT EXISTS `user_setting` (
  id               integer       not null primary key AUTO_INCREMENT,
  uid              VARCHAR(40)   not null default '',                              -- 用户UID
  to_uid           VARCHAR(40)   not null default '',                              -- 对方uid
  mute             smallint      not null DEFAULT 0,                               --  是否免打扰
  top              smallint      not null DEFAULT 0,                               -- 是否置顶
  blacklist        smallint      not null DEFAULT 0,                               -- 是否黑名单 0:正常1:黑名单
  chat_pwd_on      smallint      not null DEFAULT 0,                               -- 是否开启聊天密码
  screenshot       smallint      not null DEFAULT 1,                               -- 截屏通知
  revoke_remind    smallint      not null DEFAULT 1,                               -- 撤回通知
  receipt          smallint      not null default 1,                               -- 消息是否回执
  version          BIGINT        not null DEFAULT 0,                               -- 版本
  created_at       timestamp     not null DEFAULT CURRENT_TIMESTAMP,               -- 创建时间
  updated_at       timestamp     not null DEFAULT CURRENT_TIMESTAMP                -- 更新时间
);
CREATE INDEX IF NOT EXISTS uid_idx on `user_setting` (uid);

-- 用户设备
CREATE TABLE IF NOT EXISTS `device` (
  id                integer             not null primary key AUTO_INCREMENT,
  uid               VARCHAR(40)         not null default '',                      -- 设备所属用户uid                      
  device_id         VARCHAR(40)         not null default '',                      -- 设备唯一ID          
  device_name       VARCHAR(100)        not null default '',                      -- 设备名称                  
  device_model      VARCHAR(100)        not null default '',                      -- 设备型号              
  last_login        integer             not null DEFAULT 0,                       -- 最后一次登录时间(时间戳 10位)
  created_at        timestamp           not null DEFAULT CURRENT_TIMESTAMP,       -- 创建时间
  updated_at        timestamp           not null DEFAULT CURRENT_TIMESTAMP        -- 更新时间
);
CREATE UNIQUE INDEX IF NOT EXISTS device_uid_device_id on `device` (uid, device_id);
CREATE INDEX IF NOT EXISTS device_uid on `device` (uid);
CREATE INDEX IF NOT EXISTS device_device_id on `device` (device_id);

-- 好友表
CREATE TABLE IF NOT EXISTS `friend` (
  id                integer               not null primary key AUTO_INCREMENT,
  uid               VARCHAR(40)           not null default '' comment '用户UID',       
  to_uid            VARCHAR(40)           not null default '' comment '好友uid',                         
  remark            varchar(100)          not null default '' comment '对好友的备注 TODO: 此字段不再使用，已经迁移到user_setting表', 
  flag              smallint              not null default 0 comment '好友标示', 
  `version`           bigint                not null default 0 comment '版本号',
  vercode           VARCHAR(100)          not null default '' comment '验证码 加好友来源',   
  source_vercode    varchar(100)          not null default '' comment '好友来源',      
  is_deleted        smallint              not null default 0 comment '是否已删除', 
  is_alone          smallint              not null default 0 comment  '单项好友',
  initiator         smallint              not null default 0 comment '加好友发起方',
  created_at        timestamp             not null DEFAULT CURRENT_TIMESTAMP comment '创建时间',
  updated_at        timestamp             not null DEFAULT CURRENT_TIMESTAMP comment '更新时间'
);

-- 登录日志
CREATE TABLE IF NOT EXISTS `login_log` (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uid VARCHAR(40) DEFAULT '' NOT NULL  COMMENT '用户OpenId',
  login_ip    VARCHAR(40) DEFAULT '' NOT NULL COMMENT '最后一次登录ip',
  created_at  timestamp     not null DEFAULT CURRENT_TIMESTAMP comment '创建时间',
  updated_at  timestamp     not null DEFAULT CURRENT_TIMESTAMP comment '更新时间'
) CHARACTER SET utf8mb4;

-- 消息表（主表）
CREATE TABLE IF NOT EXISTS `message` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `message_seq` int NOT NULL DEFAULT 0 COMMENT '消息序列号',
  `client_msg_no` varchar(40) NOT NULL DEFAULT '' COMMENT '客户端消息编号',
  `header` text COMMENT '消息头部',
  `setting` tinyint NOT NULL DEFAULT 0 COMMENT '消息设置',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `timestamp` bigint NOT NULL DEFAULT 0 COMMENT '消息时间戳',
  `payload` longblob COMMENT '消息内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `signal` tinyint NOT NULL DEFAULT 0 COMMENT '信号标志',
  `expire` int NOT NULL DEFAULT 0 COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_from_uid` (`from_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

-- 消息分表1
CREATE TABLE IF NOT EXISTS `message1` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `message_seq` int NOT NULL DEFAULT 0 COMMENT '消息序列号',
  `client_msg_no` varchar(40) NOT NULL DEFAULT '' COMMENT '客户端消息编号',
  `header` text COMMENT '消息头部',
  `setting` tinyint NOT NULL DEFAULT 0 COMMENT '消息设置',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `timestamp` bigint NOT NULL DEFAULT 0 COMMENT '消息时间戳',
  `payload` longblob COMMENT '消息内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `signal` tinyint NOT NULL DEFAULT 0 COMMENT '信号标志',
  `expire` int NOT NULL DEFAULT 0 COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_from_uid` (`from_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表1';

-- 消息分表2
CREATE TABLE IF NOT EXISTS `message2` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `message_seq` int NOT NULL DEFAULT 0 COMMENT '消息序列号',
  `client_msg_no` varchar(40) NOT NULL DEFAULT '' COMMENT '客户端消息编号',
  `header` text COMMENT '消息头部',
  `setting` tinyint NOT NULL DEFAULT 0 COMMENT '消息设置',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `timestamp` bigint NOT NULL DEFAULT 0 COMMENT '消息时间戳',
  `payload` longblob COMMENT '消息内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `signal` tinyint NOT NULL DEFAULT 0 COMMENT '信号标志',
  `expire` int NOT NULL DEFAULT 0 COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_from_uid` (`from_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表2';

-- 消息分表3
CREATE TABLE IF NOT EXISTS `message3` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `message_seq` int NOT NULL DEFAULT 0 COMMENT '消息序列号',
  `client_msg_no` varchar(40) NOT NULL DEFAULT '' COMMENT '客户端消息编号',
  `header` text COMMENT '消息头部',
  `setting` tinyint NOT NULL DEFAULT 0 COMMENT '消息设置',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `timestamp` bigint NOT NULL DEFAULT 0 COMMENT '消息时间戳',
  `payload` longblob COMMENT '消息内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `signal` tinyint NOT NULL DEFAULT 0 COMMENT '信号标志',
  `expire` int NOT NULL DEFAULT 0 COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_from_uid` (`from_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表3';

-- 消息分表4
CREATE TABLE IF NOT EXISTS `message4` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `message_seq` int NOT NULL DEFAULT 0 COMMENT '消息序列号',
  `client_msg_no` varchar(40) NOT NULL DEFAULT '' COMMENT '客户端消息编号',
  `header` text COMMENT '消息头部',
  `setting` tinyint NOT NULL DEFAULT 0 COMMENT '消息设置',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `timestamp` bigint NOT NULL DEFAULT 0 COMMENT '消息时间戳',
  `payload` longblob COMMENT '消息内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `signal` tinyint NOT NULL DEFAULT 0 COMMENT '信号标志',
  `expire` int NOT NULL DEFAULT 0 COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`),
  KEY `idx_from_uid` (`from_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表4';

-- 消息扩展表
CREATE TABLE IF NOT EXISTS `message_extra` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `message_seq` bigint NOT NULL DEFAULT 0 COMMENT '消息序列号',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `from_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者UID',
  `revoke` tinyint NOT NULL DEFAULT 0 COMMENT '是否撤回',
  `revoker` varchar(40) NOT NULL DEFAULT '' COMMENT '撤回者UID',
  `clone_no` varchar(40) NOT NULL DEFAULT '' COMMENT '未读编号',
  `version` bigint NOT NULL DEFAULT 0 COMMENT '数据版本',
  `readed_count` int NOT NULL DEFAULT 0 COMMENT '已读数量',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_id` (`message_id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_from_uid` (`from_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息扩展表';

-- 成员已读表
CREATE TABLE IF NOT EXISTS `member_readed` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `clone_no` varchar(40) NOT NULL DEFAULT '' COMMENT '克隆成员唯一编号',
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '已读用户UID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_uid` (`message_id`, `uid`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='成员已读表';

-- 最近会话扩展表
CREATE TABLE IF NOT EXISTS `conversation_extra` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '所属用户',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `browse_to` bigint NOT NULL DEFAULT 0 COMMENT '预览到的位置',
  `keep_message_seq` bigint NOT NULL DEFAULT 0 COMMENT '会话保持的位置',
  `keep_offset_y` int NOT NULL DEFAULT 0 COMMENT '会话保持的位置的偏移量',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_channel` (`uid`, `channel_id`, `channel_type`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='最近会话扩展表';

-- 设备偏移量表
CREATE TABLE IF NOT EXISTS `device_offset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `device_id` varchar(40) NOT NULL DEFAULT '' COMMENT '设备ID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `max_seq` bigint NOT NULL DEFAULT 0 COMMENT '最大序列号',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_device_channel` (`uid`, `device_id`, `channel_id`, `channel_type`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备偏移量表';

-- 频道偏移量表
CREATE TABLE IF NOT EXISTS `channel_offset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `max_seq` bigint NOT NULL DEFAULT 0 COMMENT '最大序列号',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_channel` (`channel_id`, `channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='频道偏移量表';

-- 消息用户扩展表
CREATE TABLE IF NOT EXISTS `message_user_extra` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `is_read` tinyint NOT NULL DEFAULT 0 COMMENT '是否已读',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_message` (`uid`, `message_id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息用户扩展表';

-- 消息用户扩展分表1
CREATE TABLE IF NOT EXISTS `message_user_extra1` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `is_read` tinyint NOT NULL DEFAULT 0 COMMENT '是否已读',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_message` (`uid`, `message_id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息用户扩展表1';

-- 消息用户扩展分表2
CREATE TABLE IF NOT EXISTS `message_user_extra2` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息唯一ID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `is_read` tinyint NOT NULL DEFAULT 0 COMMENT '是否已读',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_message` (`uid`, `message_id`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息用户扩展表2';

-- 违禁词表
CREATE TABLE IF NOT EXISTS `prohibit_words` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `content` varchar(100) NOT NULL DEFAULT '' COMMENT '违禁词内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `version` bigint NOT NULL DEFAULT 0 COMMENT '数据版本',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_content` (`content`),
  KEY `idx_version` (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='违禁词表';

-- 提醒表
CREATE TABLE IF NOT EXISTS `reminders` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息ID',
  `client_msg_no` varchar(40) NOT NULL DEFAULT '' COMMENT '消息client msg no',
  `remind_at` timestamp NOT NULL COMMENT '提醒时间',
  `content` text COMMENT '提醒内容',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否被删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_channel` (`channel_id`, `channel_type`),
  KEY `idx_remind_at` (`remind_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提醒表';

-- 提醒完成表
CREATE TABLE IF NOT EXISTS `reminder_done` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `reminder_id` bigint NOT NULL DEFAULT 0 COMMENT '提醒ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_reminder` (`uid`, `reminder_id`),
  KEY `idx_reminder_id` (`reminder_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提醒完成表';

-- 管理员发送消息记录表
CREATE TABLE IF NOT EXISTS `send_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `receiver` varchar(40) NOT NULL DEFAULT '' COMMENT '接受者uid',
  `receiver_name` varchar(100) NOT NULL DEFAULT '' COMMENT '接受者',
  `receiver_channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '接受者频道类型',
  `sender` varchar(40) NOT NULL DEFAULT '' COMMENT '发送者uid',
  `sender_name` varchar(100) NOT NULL DEFAULT '' COMMENT '发送者名字',
  `handler_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '操作者uid',
  `handler_name` varchar(100) NOT NULL DEFAULT '' COMMENT '操作者名字',
  `content` text COMMENT '发送内容',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_receiver` (`receiver`),
  KEY `idx_sender` (`sender`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员发送消息记录表';

-- 钱包相关表

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
  UNIQUE KEY `wk_uid` (`uid`),
  KEY `idx_balance` (`balance`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包表';

-- 钱包流水记录表
CREATE TABLE IF NOT EXISTS `wallet_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `record_no` varchar(64) NOT NULL DEFAULT '' COMMENT '流水号',
  `type` tinyint NOT NULL DEFAULT 0 COMMENT '类型 1:充值 2:提现 3:转账收入 4:转账支出 5:红包发出 6:红包收入 7:红包退回 8:利息收入',
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

-- 标签表
CREATE TABLE IF NOT EXISTS `label` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '标签名称',
  `color` varchar(20) NOT NULL DEFAULT '' COMMENT '标签颜色',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签表';

-- 用户标签关联表
CREATE TABLE IF NOT EXISTS `user_label` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `label_id` bigint NOT NULL DEFAULT 0 COMMENT '标签ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_label` (`uid`, `label_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户标签关联表';

-- 应用配置表
CREATE TABLE IF NOT EXISTS `app_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `module` varchar(50) NOT NULL DEFAULT '' COMMENT '模块',
  `key` varchar(100) NOT NULL DEFAULT '' COMMENT '键',
  `value` text COMMENT '值',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_module_key` (`module`, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用配置表';

-- 插入默认配置
INSERT INTO `app_config` (module, key, value) VALUES ('system', 'normal_user_can_add_friend', '1');

-- 群组相关表

-- 群表
CREATE TABLE IF NOT EXISTS `group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(40) NOT NULL DEFAULT '' COMMENT '群ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '群名称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '群头像',
  `short_no` varchar(40) NOT NULL DEFAULT '' COMMENT '群短号',
  `category` varchar(40) NOT NULL DEFAULT '' COMMENT '群分类',
  `owner_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '群主UID',
  `member_count` int NOT NULL DEFAULT 0 COMMENT '成员数量',
  `max_member_count` int NOT NULL DEFAULT 500 COMMENT '最大成员数量',
  `is_audit` tinyint NOT NULL DEFAULT 0 COMMENT '是否需要审核',
  `is_public` tinyint NOT NULL DEFAULT 0 COMMENT '是否公开群',
  `join_type` tinyint NOT NULL DEFAULT 0 COMMENT '加入方式 0:邀请 1:扫码 2:搜索',
  `search_by_phone` tinyint NOT NULL DEFAULT 1 COMMENT '是否可以通过手机号搜索',
  `search_by_short` tinyint NOT NULL DEFAULT 1 COMMENT '是否可以通过短号搜索',
  `welcome_message` text COMMENT '欢迎消息',
  `version` bigint NOT NULL DEFAULT 0 COMMENT '数据版本',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_id` (`group_id`),
  KEY `idx_owner_uid` (`owner_uid`),
  KEY `idx_short_no` (`short_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群表';

-- 群成员表
CREATE TABLE IF NOT EXISTS `group_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(40) NOT NULL DEFAULT '' COMMENT '群ID',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '成员UID',
  `role` tinyint NOT NULL DEFAULT 0 COMMENT '角色 0:普通成员 1:管理员 2:群主',
  `nickname` varchar(100) NOT NULL DEFAULT '' COMMENT '群昵称',
  `join_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  `last_read_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后阅读时间',
  `version` bigint NOT NULL DEFAULT 0 COMMENT '数据版本',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_uid` (`group_id`, `uid`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群成员表';

-- 群设置表
CREATE TABLE IF NOT EXISTS `group_setting` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(40) NOT NULL DEFAULT '' COMMENT '群ID',
  `mute` tinyint NOT NULL DEFAULT 0 COMMENT '是否禁言',
  `top` tinyint NOT NULL DEFAULT 0 COMMENT '是否置顶',
  `notification` tinyint NOT NULL DEFAULT 1 COMMENT '是否通知',
  `search_by_phone` tinyint NOT NULL DEFAULT 1 COMMENT '是否可以通过手机号搜索',
  `search_by_short` tinyint NOT NULL DEFAULT 1 COMMENT '是否可以通过短号搜索',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群设置表';

-- 群邀请表
CREATE TABLE IF NOT EXISTS `group_invite` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(40) NOT NULL DEFAULT '' COMMENT '群ID',
  `invite_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '邀请者UID',
  `to_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '被邀请者UID',
  `token` varchar(100) NOT NULL DEFAULT '' COMMENT '邀请令牌',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:待处理 1:已接受 2:已拒绝',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_invite` (`group_id`, `invite_uid`, `to_uid`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_to_uid` (`to_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群邀请表';

-- 频道相关表

-- 频道设置表
CREATE TABLE IF NOT EXISTS `channel_setting` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `channel_id` varchar(100) NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` tinyint NOT NULL DEFAULT 0 COMMENT '频道类型',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '频道名称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '频道头像',
  `description` text COMMENT '频道描述',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_channel` (`channel_id`, `channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='频道设置表';

-- 文件相关表

-- 文件表
CREATE TABLE IF NOT EXISTS `file` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `file_id` varchar(40) NOT NULL DEFAULT '' COMMENT '文件ID',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '上传者UID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `size` bigint NOT NULL DEFAULT 0 COMMENT '文件大小',
  `type` varchar(40) NOT NULL DEFAULT '' COMMENT '文件类型',
  `url` varchar(500) NOT NULL DEFAULT '' COMMENT '文件URL',
  `path` varchar(500) NOT NULL DEFAULT '' COMMENT '文件路径',
  `storage_type` varchar(40) NOT NULL DEFAULT '' COMMENT '存储类型',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_id` (`file_id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件表';

-- 工作区相关表

-- 工作区应用表
CREATE TABLE IF NOT EXISTS `workplace_app` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_id` varchar(40) NOT NULL DEFAULT '' COMMENT '应用ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '应用名称',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '应用图标',
  `url` varchar(500) NOT NULL DEFAULT '' COMMENT '应用URL',
  `description` text COMMENT '应用描述',
  `category` varchar(40) NOT NULL DEFAULT '' COMMENT '应用分类',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_id` (`app_id`),
  KEY `idx_category` (`category`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作区应用表';

-- 工作区应用用户记录表
CREATE TABLE IF NOT EXISTS `workplace_app_user_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `app_id` varchar(40) NOT NULL DEFAULT '' COMMENT '应用ID',
  `open_count` int NOT NULL DEFAULT 0 COMMENT '打开次数',
  `last_open_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后打开时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_app` (`uid`, `app_id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作区应用用户记录表';

-- 工作区用户应用表
CREATE TABLE IF NOT EXISTS `workplace_user_app` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `app_id` varchar(40) NOT NULL DEFAULT '' COMMENT '应用ID',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_app` (`uid`, `app_id`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作区用户应用表';

-- 工作区Banner表
CREATE TABLE IF NOT EXISTS `workplace_banner` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(100) NOT NULL DEFAULT '' COMMENT 'Banner标题',
  `image` varchar(255) NOT NULL DEFAULT '' COMMENT 'Banner图片',
  `url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Banner链接',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作区Banner表';

-- 工作区分类表
CREATE TABLE IF NOT EXISTS `workplace_category` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `category_id` varchar(40) NOT NULL DEFAULT '' COMMENT '分类ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '分类名称',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '分类图标',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_id` (`category_id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作区分类表';

-- 工作区分类应用表
CREATE TABLE IF NOT EXISTS `workplace_category_app` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `category_id` varchar(40) NOT NULL DEFAULT '' COMMENT '分类ID',
  `app_id` varchar(40) NOT NULL DEFAULT '' COMMENT '应用ID',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_app` (`category_id`, `app_id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作区分类应用表';

-- 统计相关表

-- 统计数据表
CREATE TABLE IF NOT EXISTS `statistics` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL COMMENT '统计日期',
  `type` varchar(40) NOT NULL DEFAULT '' COMMENT '统计类型',
  `value` bigint NOT NULL DEFAULT 0 COMMENT '统计值',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_date_type` (`date`, `type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统计数据表';

-- 热点线路相关表

-- 热点线路表
CREATE TABLE IF NOT EXISTS `hotline` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `hotline_id` varchar(40) NOT NULL DEFAULT '' COMMENT '热点线路ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '热点线路名称',
  `description` text COMMENT '热点线路描述',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hotline_id` (`hotline_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热点线路表';

-- 热点线路会话表
CREATE TABLE IF NOT EXISTS `hotline_session` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `session_id` varchar(40) NOT NULL DEFAULT '' COMMENT '会话ID',
  `hotline_id` varchar(40) NOT NULL DEFAULT '' COMMENT '热点线路ID',
  `visitor_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '访客UID',
  `agent_uid` varchar(40) NOT NULL DEFAULT '' COMMENT '客服UID',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:待处理 1:处理中 2:已结束',
  `start_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
  `end_time` timestamp NULL DEFAULT NULL COMMENT '结束时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_session_id` (`session_id`),
  KEY `idx_hotline_id` (`hotline_id`),
  KEY `idx_visitor_uid` (`visitor_uid`),
  KEY `idx_agent_uid` (`agent_uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热点线路会话表';

-- 机器人相关表

-- 机器人表
CREATE TABLE IF NOT EXISTS `robot` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `robot_id` varchar(40) NOT NULL DEFAULT '' COMMENT '机器人ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '机器人名称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '机器人头像',
  `description` text COMMENT '机器人描述',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_robot_id` (`robot_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器人表';

-- 机器人菜单表
CREATE TABLE IF NOT EXISTS `robot_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `robot_id` varchar(40) NOT NULL DEFAULT '' COMMENT '机器人ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '菜单名称',
  `type` tinyint NOT NULL DEFAULT 0 COMMENT '菜单类型 0:文本 1:图片 2:链接 3:子菜单',
  `content` text COMMENT '菜单内容',
  `parent_id` bigint NOT NULL DEFAULT 0 COMMENT '父菜单ID',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器人菜单表';

-- Webhook相关表

-- Webhook表
CREATE TABLE IF NOT EXISTS `webhook` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `webhook_id` varchar(40) NOT NULL DEFAULT '' COMMENT 'Webhook ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT 'Webhook名称',
  `url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Webhook URL',
  `secret` varchar(100) NOT NULL DEFAULT '' COMMENT 'Webhook密钥',
  `event_type` varchar(40) NOT NULL DEFAULT '' COMMENT '事件类型',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态 0:禁用 1:正常',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_webhook_id` (`webhook_id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Webhook表';

-- Webhook消息表
CREATE TABLE IF NOT EXISTS `webhook_message` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `webhook_id` varchar(40) NOT NULL DEFAULT '' COMMENT 'Webhook ID',
  `message_id` varchar(40) NOT NULL DEFAULT '' COMMENT '消息ID',
  `event_type` varchar(40) NOT NULL DEFAULT '' COMMENT '事件类型',
  `payload` text COMMENT '消息内容',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态 0:待发送 1:已发送 2:发送失败',
  `retry_count` int NOT NULL DEFAULT 0 COMMENT '重试次数',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_webhook_id` (`webhook_id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Webhook消息表';

-- 短号表
CREATE TABLE IF NOT EXISTS `shortno` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `shortno` varchar(40) NOT NULL DEFAULT '' COMMENT '短号',
  `business` varchar(40) NOT NULL DEFAULT '' COMMENT '业务类型',
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `is_used` tinyint NOT NULL DEFAULT 0 COMMENT '是否使用',
  `is_locked` tinyint NOT NULL DEFAULT 0 COMMENT '是否锁定',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shortno` (`shortno`),
  KEY `idx_business` (`business`),
  KEY `idx_uid` (`uid`),
  KEY `idx_is_used` (`is_used`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短号表';

-- 用户在线表
CREATE TABLE IF NOT EXISTS `user_online` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户UID',
  `device_flag` smallint NOT NULL DEFAULT 0 COMMENT '设备标记 0.APP 1.web',
  `last_online` int NOT NULL DEFAULT 0 COMMENT '最后一次在线时间',
  `last_offline` int NOT NULL DEFAULT 0 COMMENT '最后一次离线时间',
  `online` tinyint NOT NULL DEFAULT 0 COMMENT '用户是否在线',
  `version` bigint NOT NULL DEFAULT 0 COMMENT '数据版本',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_device` (`uid`, `device_flag`),
  KEY `idx_online` (`online`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户在线表';

-- 初始化系统账号
INSERT INTO `user` (uid,name,short_no,phone,zone,search_by_phone,search_by_short,new_msg_notice,voice_on,shock_on,msg_show_detail,status,is_upload_avatar,category,robot) VALUES ('u_10000','系统账号',10000,'13000000000','0086',0,0,0,0,0,0,1,1,'system',1);
INSERT INTO `user` (uid,name,short_no,phone,zone,search_by_phone,search_by_short,new_msg_notice,voice_on,shock_on,msg_show_detail,status,is_upload_avatar,category) VALUES ('fileHelper','文件传输助手',20000,'13000000001','0086',0,0,0,0,0,0,1,1,'system');

-- 初始化应用配置
INSERT INTO `app_config` (module, key, value) VALUES ('system', 'version', '1.0.0');
INSERT INTO `app_config` (module, key, value) VALUES ('system', 'welcome_message', '欢迎使用悟空聊天！');
INSERT INTO `app_config` (module, key, value) VALUES ('upload', 'max_size', '104857600');
INSERT INTO `app_config` (module, key, value) VALUES ('upload', 'allow_types', 'image/jpeg,image/png,image/gif,video/mp4,audio/mp3,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
INSERT INTO `app_config` (module, key, value) VALUES ('group', 'max_member_count', '500');
INSERT INTO `app_config` (module, key, value) VALUES ('message', 'max_length', '5000');
INSERT INTO `app_config` (module, key, value) VALUES ('message', 'max_image_size', '10485760');
INSERT INTO `app_config` (module, key, value) VALUES ('message', 'max_video_size', '52428800');
INSERT INTO `app_config` (module, key, value) VALUES ('message', 'max_audio_size', '10485760');
INSERT INTO `app_config` (module, key, value) VALUES ('message', 'max_file_size', '104857600');

-- 初始化违禁词
INSERT INTO `prohibit_words` (content, is_deleted, version) VALUES ('违法', 0, 1);
INSERT INTO `prohibit_words` (content, is_deleted, version) VALUES ('赌博', 0, 1);
INSERT INTO `prohibit_words` (content, is_deleted, version) VALUES ('色情', 0, 1);
INSERT INTO `prohibit_words` (content, is_deleted, version) VALUES ('暴力', 0, 1);
INSERT INTO `prohibit_words` (content, is_deleted, version) VALUES ('毒品', 0, 1);

-- 初始化短号
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('10001', 'user', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('10002', 'user', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('10003', 'user', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('10004', 'user', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('10005', 'user', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('20001', 'group', 0, 0);

-- 确保用户表包含所有必要字段
ALTER TABLE `user` ADD COLUMN IF NOT EXISTS `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户头像';

-- 确保钱包表包含所有必要字段
ALTER TABLE `wallet` ADD COLUMN IF NOT EXISTS `real_name_status` tinyint NOT NULL DEFAULT 0 COMMENT '实名认证状态：0-未认证，1-审核中，2-已认证，3-认证失败';
ALTER TABLE `wallet` ADD COLUMN IF NOT EXISTS `real_name` varchar(50) NOT NULL DEFAULT '' COMMENT '真实姓名';
ALTER TABLE `wallet` ADD COLUMN IF NOT EXISTS `id_card` varchar(20) NOT NULL DEFAULT '' COMMENT '身份证号';
ALTER TABLE `wallet` ADD COLUMN IF NOT EXISTS `interest` bigint NOT NULL DEFAULT 0 COMMENT '累计利息(分)';
ALTER TABLE `wallet` ADD COLUMN IF NOT EXISTS `today_interest` bigint NOT NULL DEFAULT 0 COMMENT '今日利息(分)';
ALTER TABLE `wallet` ADD COLUMN IF NOT EXISTS `interest_rate` double NOT NULL DEFAULT 0 COMMENT '年化利率(%)';
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('20002', 'group', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('20003', 'group', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('20004', 'group', 0, 0);
INSERT INTO `shortno` (shortno, business, is_used, is_locked) VALUES ('20005', 'group', 0, 0);
