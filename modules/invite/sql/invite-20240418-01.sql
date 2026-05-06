-- +migrate Up

-- 收藏表
create table `invite`
(
  id                integer                not null primary key AUTO_INCREMENT,
  uid               VARCHAR(40)            not null default '',                     -- 用户uid
  invite_code       VARCHAR(20)            not null default '',                     -- 邀请码
  be_invite_uid     VARCHAR(40)            not null default '',                     -- 邀请人uid
  be_invite_code    VARCHAR(40)            not null default '',                     -- 邀请code
  status            smallint               not null default 1,                      -- 1.可用
  created_at        timeStamp              not null DEFAULT CURRENT_TIMESTAMP,      -- 创建时间
  updated_at        timeStamp              not null DEFAULT CURRENT_TIMESTAMP       -- 更新时间
);
CREATE unique INDEX invite_uid_code on `invite` (uid, invite_code);
CREATE INDEX invite_be_invite_uid on `invite` (be_invite_uid);
