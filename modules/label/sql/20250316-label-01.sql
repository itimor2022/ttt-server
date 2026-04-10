-- +migrate Up

-- 标签
create table `label`(
           id                  integer         not null primary key AUTO_INCREMENT,
           uid                 varchar(100)     not null default '',
           name                 varchar(100)   not null default '',
           member_uids          text            not null ,
           created_at          timeStamp       not null DEFAULT CURRENT_TIMESTAMP, -- 创建时间
           updated_at          timeStamp       not null DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);
