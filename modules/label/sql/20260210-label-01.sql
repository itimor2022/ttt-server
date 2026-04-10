-- +migrate Up

-- 删除旧的标签表
DROP TABLE IF EXISTS `label`;

-- 创建新的标签表
CREATE TABLE `label`(
    id                  INTEGER         NOT NULL PRIMARY KEY AUTO_INCREMENT,
    uid                 VARCHAR(100)    NOT NULL DEFAULT '',
    name                VARCHAR(100)    NOT NULL DEFAULT '',
    created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建标签成员表
CREATE TABLE `label_member`(
    id                  INTEGER         NOT NULL PRIMARY KEY AUTO_INCREMENT,
    label_id            INTEGER         NOT NULL,
    user_id             VARCHAR(100)    NOT NULL DEFAULT '',
    added_at            TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (label_id) REFERENCES label(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_label_uid ON `label`(uid);
CREATE INDEX idx_label_member_label_id ON `label_member`(label_id);
CREATE INDEX idx_label_member_user_id ON `label_member`(user_id);
