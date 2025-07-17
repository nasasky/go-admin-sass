-- 修复用户notice字段脚本
-- 解决推送系统用户接收记录为空的问题

-- 1. 首先查看当前用户notice字段状态
SELECT 
    id, 
    username, 
    user_type, 
    role_id, 
    notice,
    CASE 
        WHEN notice = 1 THEN '可接收推送'
        WHEN notice = 0 THEN '不接收推送'
        WHEN notice IS NULL THEN '未设置'
        ELSE '其他'
    END AS notice_status
FROM user 
ORDER BY id;

-- 2. 为管理员用户（user_type=1）设置notice=1
UPDATE user SET notice = 1 WHERE user_type = 1;

-- 3. 确认已设置的用户
SELECT 
    id, 
    username, 
    user_type, 
    role_id, 
    notice
FROM user 
WHERE notice = 1;

-- 4. 可选：为特定角色的用户设置notice=1
-- 如果需要特定角色也接收推送，取消下面注释并修改role_id
-- UPDATE user SET notice = 1 WHERE role_id IN (1, 4);

-- 5. 可选：为特定用户名设置notice=1
-- UPDATE user SET notice = 1 WHERE username = 'admin';

-- 验证修复结果
SELECT 
    COUNT(*) as total_users,
    SUM(CASE WHEN notice = 1 THEN 1 ELSE 0 END) as notice_enabled_users,
    SUM(CASE WHEN user_type = 1 THEN 1 ELSE 0 END) as admin_users
FROM user; 