-- 订单系统核心表创建脚本
-- 数据库: naive_admin
-- Author: Order Security System
-- Date: 2024

USE naive_admin;

-- 1. 创建订单表 (app_order)
CREATE TABLE IF NOT EXISTS `app_order` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `goods_id` int(11) NOT NULL COMMENT '商品ID',
  `num` int(11) NOT NULL DEFAULT '1' COMMENT '购买数量',
  `amount` decimal(10,2) NOT NULL COMMENT '订单金额',
  `tenants_id` int(11) NOT NULL DEFAULT '1' COMMENT '租户ID',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '订单状态: pending-待支付, paid-已支付, completed-已完成, cancelled-已取消, refunded-已退款',
  `no` varchar(50) NOT NULL COMMENT '订单号',
  `remark` text COMMENT '订单备注',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `pay_time` datetime NULL COMMENT '支付时间',
  `complete_time` datetime NULL COMMENT '完成时间',
  `cancel_time` datetime NULL COMMENT '取消时间',
  `refund_time` datetime NULL COMMENT '退款时间',
  `express_no` varchar(100) NULL COMMENT '快递单号',
  `express_company` varchar(50) NULL COMMENT '快递公司',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_order_no` (`no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_goods_id` (`goods_id`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_user_status_time` (`user_id`, `status`, `create_time`),
  KEY `idx_status_time` (`status`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 2. 创建钱包交易记录表 (app_recharge)
CREATE TABLE IF NOT EXISTS `app_recharge` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `order_id` int(11) NULL COMMENT '关联订单ID',
  `order_no` varchar(50) NULL COMMENT '关联订单号',
  `amount` decimal(10,2) NOT NULL COMMENT '交易金额',
  `balance_before` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '交易前余额',
  `balance_after` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '交易后余额',
  `transaction_type` varchar(30) NOT NULL COMMENT '交易类型: recharge-充值, order_payment-订单支付, refund-退款, withdraw-提现',
  `status` varchar(20) NOT NULL DEFAULT 'completed' COMMENT '交易状态: pending-处理中, completed-已完成, failed-失败',
  `payment_method` varchar(20) NULL COMMENT '支付方式: wechat-微信, alipay-支付宝, balance-余额',
  `payment_no` varchar(100) NULL COMMENT '第三方支付流水号',
  `remark` varchar(255) NULL COMMENT '交易备注',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `complete_time` datetime NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_transaction_type` (`transaction_type`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_user_time` (`user_id`, `create_time`),
  KEY `idx_type_time` (`transaction_type`, `create_time`),
  KEY `idx_status_time` (`status`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包交易记录表';

-- 3. 创建商品表 (app_goods) - 如果不存在
CREATE TABLE IF NOT EXISTS `app_goods` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '商品ID',
  `goods_name` varchar(200) NOT NULL COMMENT '商品名称',
  `price` decimal(10,2) NOT NULL COMMENT '商品价格',
  `stock` int(11) NOT NULL DEFAULT '0' COMMENT '库存数量',
  `category_id` int(11) NULL COMMENT '分类ID',
  `tenants_id` int(11) NOT NULL DEFAULT '1' COMMENT '租户ID',
  `status` varchar(10) NOT NULL DEFAULT '1' COMMENT '状态: 1-上架, 0-下架',
  `description` text COMMENT '商品描述',
  `images` text COMMENT '商品图片',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category_id`),
  KEY `idx_status` (`status`),
  KEY `idx_stock` (`stock`),
  KEY `idx_status_stock` (`status`, `stock`),
  KEY `idx_tenants` (`tenants_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- 4. 创建用户钱包表 (app_user_wallet) - 如果不存在
CREATE TABLE IF NOT EXISTS `app_user_wallet` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `balance` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '余额',
  `money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '可用余额',
  `frozen_money` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '冻结金额',
  `total_recharge` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '累计充值',
  `total_consume` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '累计消费',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int(11) NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_id` (`user_id`),
  KEY `idx_balance` (`balance`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户钱包表';

-- 5. 创建用户表 (app_user) - 如果不存在
CREATE TABLE IF NOT EXISTS `app_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NULL COMMENT '用户名',
  `nickname` varchar(100) NULL COMMENT '昵称',
  `phone` varchar(20) NULL COMMENT '手机号',
  `email` varchar(100) NULL COMMENT '邮箱',
  `avatar` varchar(255) NULL COMMENT '头像',
  `status` varchar(10) NOT NULL DEFAULT '1' COMMENT '状态: 1-正常, 0-禁用',
  `openid` varchar(100) NULL COMMENT '微信openid',
  `unionid` varchar(100) NULL COMMENT '微信unionid',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_phone` (`phone`),
  UNIQUE KEY `idx_openid` (`openid`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 6. 插入测试数据
-- 测试商品
INSERT IGNORE INTO `app_goods` (`id`, `goods_name`, `price`, `stock`, `tenants_id`, `status`, `description`) VALUES
(1, '测试商品1', 99.99, 100, 1, '1', '这是一个测试商品'),
(2, '测试商品2', 199.99, 50, 1, '1', '这是第二个测试商品'),
(3, '测试商品3', 299.99, 30, 1, '1', '这是第三个测试商品');

-- 测试用户
INSERT IGNORE INTO `app_user` (`id`, `username`, `nickname`, `phone`, `status`) VALUES
(1, 'test_user', '测试用户', '13800138000', '1'),
(2, 'test_user2', '测试用户2', '13800138001', '1');

-- 用户钱包
INSERT IGNORE INTO `app_user_wallet` (`user_id`, `money`, `balance`, `total_recharge`) VALUES
(1, 1000.00, 1000.00, 1000.00),
(2, 500.00, 500.00, 500.00);

-- 初始化充值记录
INSERT IGNORE INTO `app_recharge` (`user_id`, `amount`, `balance_before`, `balance_after`, `transaction_type`, `status`, `remark`) VALUES
(1, 1000.00, 0.00, 1000.00, 'recharge', 'completed', '初始充值'),
(2, 500.00, 0.00, 500.00, 'recharge', 'completed', '初始充值');

-- 7. 创建视图（用于监控和统计）
CREATE OR REPLACE VIEW `v_order_stats` AS
SELECT 
    DATE(create_time) as order_date,
    status,
    COUNT(*) as order_count,
    SUM(amount) as total_amount,
    AVG(amount) as avg_amount
FROM app_order 
GROUP BY DATE(create_time), status;

CREATE OR REPLACE VIEW `v_user_order_summary` AS
SELECT 
    user_id,
    COUNT(*) as total_orders,
    SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) as completed_amount,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_orders,
    MAX(create_time) as last_order_time
FROM app_order 
GROUP BY user_id;

-- 8. 创建存储过程（用于数据清理）
DELIMITER $$

CREATE PROCEDURE IF NOT EXISTS `CleanExpiredOrders`()
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;
    
    START TRANSACTION;
    
    -- 取消15分钟前的待支付订单
    UPDATE app_order 
    SET status = 'cancelled', 
        cancel_time = NOW(),
        update_time = NOW()
    WHERE status = 'pending' 
    AND create_time < DATE_SUB(NOW(), INTERVAL 15 MINUTE);
    
    COMMIT;
END$$

DELIMITER ;

-- 设置权限和索引优化
OPTIMIZE TABLE app_order;
OPTIMIZE TABLE app_recharge;
OPTIMIZE TABLE app_goods;
OPTIMIZE TABLE app_user_wallet;

-- 完成
SELECT 'Order system tables created successfully!' as message; 