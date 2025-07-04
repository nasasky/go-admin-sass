-- 为用户表添加索引
CREATE INDEX IF NOT EXISTS idx_user_username ON user(username);
CREATE INDEX IF NOT EXISTS idx_user_phone ON user(phone);
CREATE INDEX IF NOT EXISTS idx_user_username_phone ON user(username, phone);

-- 为字典类型表添加索引
CREATE INDEX IF NOT EXISTS idx_dict_type_code ON sys_dict_type(type_code);
CREATE INDEX IF NOT EXISTS idx_dict_type_del_flag ON sys_dict_type(del_flag);

-- 为字典值表添加索引
CREATE INDEX IF NOT EXISTS idx_dict_type_id ON sys_dict(sys_dict_type_id);
CREATE INDEX IF NOT EXISTS idx_dict_code ON sys_dict(code);
CREATE INDEX IF NOT EXISTS idx_dict_del_flag ON sys_dict(del_flag);
CREATE INDEX IF NOT EXISTS idx_dict_composite ON sys_dict(sys_dict_type_id, del_flag, is_show);

-- 添加外键约束（如果没有的话）
ALTER TABLE sys_dict
ADD CONSTRAINT IF NOT EXISTS fk_dict_type_id
FOREIGN KEY (sys_dict_type_id)
REFERENCES sys_dict_type(id)
ON DELETE CASCADE
ON UPDATE CASCADE; 