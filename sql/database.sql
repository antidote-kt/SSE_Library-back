CREATE TABLE users (
   id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '用户唯一标识ID',
   username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
   password VARCHAR(255) NOT NULL COMMENT '用户密码',
   email VARCHAR(100) UNIQUE NOT NULL COMMENT '用户邮箱地址',
   role varchar(20) DEFAULT 'user' COMMENT '用户角色：user-普通用户，admin-管理员',
   avatar VARCHAR(500) COMMENT '用户头像文件路径，存储头像图片的URL或文件路径',
   status varchar(20) DEFAULT 'active' COMMENT '用户状态：active-正常使用，disabled-被禁用无法登录',
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '用户注册时间',
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '用户信息最后修改时间',
   deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）'
) COMMENT='用户基础信息表';
-- 用户名唯一约束（仅未删除记录）
CREATE UNIQUE INDEX idx_users_username_unique
    ON users (username, (IF(deleted_at IS NULL, 1, NULL)));

-- 邮箱唯一约束（仅未删除记录）
CREATE UNIQUE INDEX idx_users_email_unique
    ON users (email, (IF(deleted_at IS NULL, 1, NULL)));

CREATE TABLE favorites (
      id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '收藏记录ID',
      user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
      source_id BIGINT UNSIGNED NOT NULL COMMENT '被收藏对象ID',
      source_type VARCHAR(20) NOT NULL DEFAULT 'document' COMMENT '被收藏对象类型: document, post' ,
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
      deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
      PRIMARY KEY (id),
      UNIQUE KEY uk_user_favorite (user_id, source_id, source_type),
      KEY idx_source_id (source_id, source_type)
) COMMENT='资源收藏记录表';
-- 用户收藏唯一约束（避免重复收藏）
CREATE UNIQUE INDEX idx_favorites_user_doc_unique
    ON favorites (user_id, source_id, source_type, (IF(deleted_at IS NULL, 1, NULL)));

CREATE TABLE comments (
      id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
      user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
      content TEXT NOT NULL COMMENT '评论内容',
      parent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '父评论ID (用于回复)',
      source_id BIGINT UNSIGNED NOT NULL COMMENT '被评论对象ID (关联文档或帖子)',
      source_type VARCHAR(20) NOT NULL DEFAULT 'document' COMMENT '被评论对象类型: document, post',
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '评论时间',
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
      deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
      PRIMARY KEY (id),
      KEY idx_user_id (user_id),
      KEY idx_source (source_id, source_type),
      KEY idx_parent_id (parent_id)
) COMMENT='通用评论表（书籍/课程）';



CREATE TABLE view_histories (
      id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
      user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
      source_id BIGINT UNSIGNED NOT NULL COMMENT '被浏览对象ID',
      source_type VARCHAR(20) NOT NULL DEFAULT 'document' COMMENT '被浏览对象类型: document, post' ,
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后浏览时间',
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
      deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
      PRIMARY KEY (id),
      KEY idx_user_view (user_id),
      UNIQUE KEY uk_user_view_histories (user_id, source_id, source_type),
      KEY idx_source_id (source_id, source_type)
) COMMENT='通用浏览历史表';
-- 浏览记录唯一约束（避免重复添加）
CREATE UNIQUE INDEX idx_view_histories_user_doc_unique
    ON view_histories (user_id, source_id, source_type, (IF(deleted_at IS NULL, 1, NULL)));



CREATE TABLE documents (
   id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '文档唯一标识ID',
   type varchar(20) NOT NULL COMMENT '资源类型：book-电子书籍，file-其他文档文件，video-视频链接',
   name VARCHAR(200) NOT NULL COMMENT '文档名称',
   book_isbn VARCHAR(20) COMMENT 'ISBN标准书号',
   author VARCHAR(100) DEFAULT '佚名' COMMENT '作者姓名，文档的创作者信息，课程视频主讲老师',
   uploader_id BIGINT UNSIGNED NOT NULL COMMENT '上传者用户ID，关联users表，记录谁上传了此文档/视频链接',
   category_id BIGINT UNSIGNED NOT NULL COMMENT '分类ID',
   cover VARCHAR(500) COMMENT '封面图片路径，存储封面图片的URL或文件路径',
   introduction TEXT COMMENT '文档简介，描述文档内容和特点',
   create_year VARCHAR(10) COMMENT '出版年份，仅书籍类型需要',
   status varchar(20) DEFAULT 'audit' COMMENT '文档状态：open-公开可见，close-关闭, audit-待审核，withdraw-已撤回',
   read_counts INT DEFAULT 0 COMMENT '浏览次数统计',
   collections INT DEFAULT 0 COMMENT '收藏次数统计',
   url VARCHAR(500) NOT NULL COMMENT '下载/预览链接',
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '文档上传时间',
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后更新时间',
   deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
   KEY idx_uploader_id (uploader_id),
   KEY idx_category_id (category_id)
) COMMENT='电子书、文档和视频信息表';

CREATE TABLE tags (
   id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '标签记录唯一标识ID',
   tag_name VARCHAR(50) NOT NULL COMMENT '标签名称',
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '标签创建时间',
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）'
) COMMENT='文档标签表';
-- 标签名称唯一约束
CREATE UNIQUE INDEX idx_tags_name_unique
    ON tags (tag_name, (IF(deleted_at IS NULL, 1, NULL)));

CREATE TABLE document_tag (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
  document_id BIGINT UNSIGNED NOT NULL COMMENT '文档ID',
  tag_id BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）'
) COMMENT='文档与标签关系映射表';
-- 文档标签关系唯一约束
CREATE UNIQUE INDEX idx_document_tag_unique
    ON document_tag (document_id, tag_id, (IF(deleted_at IS NULL, 1, NULL)));


CREATE TABLE categories (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '分类唯一标识ID，一级分类',
    name VARCHAR(50) NOT NULL COMMENT '分类名称/课程名称',
    is_course tinyint NOT NULL COMMENT '1-细分课程或课程大类,0-非课程类别',
    description TEXT COMMENT '分类描述，详细说明分类的用途和包含的内容',
    parent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '父分类ID，关联本表，用于实现层级分类，NULL表示顶级分类，可以是泛课程也可以是非课程大类',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分类创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
    KEY idx_parent_id (parent_id)
) COMMENT='分类表';
-- 分类名称唯一约束（同一父级下）
CREATE UNIQUE INDEX idx_categories_name_unique
    ON categories (name, parent_id, (IF(deleted_at IS NULL, 1, NULL)));


CREATE TABLE sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '会话ID (对应 sessionId)',
  user1_id BIGINT UNSIGNED NOT NULL COMMENT '用户1id',
  user2_id BIGINT UNSIGNED NOT NULL COMMENT '用户2id',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
  PRIMARY KEY (id)
) COMMENT='聊天会话表';
-- 用户会话关系唯一约束（避免重复创建相同用户间的会话）
CREATE UNIQUE INDEX idx_sessions_users_unique
ON sessions (user1_id, user2_id,(IF(deleted_at IS NULL, 1, NULL)));

CREATE TABLE messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息唯一ID',
    session_id BIGINT UNSIGNED NOT NULL COMMENT '所属会话ID (外键, 对应 sessionId)',
    sender_id  BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    content TEXT NOT NULL COMMENT '消息内容 (对应 content)',
    status varchar(20) NOT NULL DEFAULT 'unread' COMMENT '消息状态 (对应 status)，有unread和read两种状态',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
    PRIMARY KEY (id),
    INDEX idx_session_id (session_id), -- 关键索引：快速按会话ID拉取聊天记录
    INDEX idx_sender_id (sender_id)
) COMMENT='聊天消息表';

CREATE TABLE posts (
                       id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '帖子唯一标识ID (对应 postId)',
                       sender_id BIGINT UNSIGNED NOT NULL COMMENT '发帖人ID (对应 senderId)',
                       title VARCHAR(200) NOT NULL COMMENT '帖子标题',
                       content TEXT NOT NULL COMMENT '帖子内容',
                       read_count INT UNSIGNED DEFAULT 0 COMMENT '阅读量',
                       like_count INT UNSIGNED DEFAULT 0 COMMENT '点赞量',
                       collect_count INT UNSIGNED DEFAULT 0 COMMENT '收藏量',
                       comment_count INT UNSIGNED DEFAULT 0 COMMENT '评论量',
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间 (对应 sendTime)',
                       deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记',
                       PRIMARY KEY (id),
                       KEY idx_user_id (sender_id)
) COMMENT='社区帖子表';

CREATE TABLE post_documents (
                                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                post_id BIGINT UNSIGNED NOT NULL COMMENT '帖子ID',
                                document_id BIGINT UNSIGNED NOT NULL COMMENT '文档ID',
                                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记',
                                PRIMARY KEY (id),
                                UNIQUE KEY uk_post_document (post_id, document_id), -- 防止同一个文档在同一个帖子中重复添加
                                KEY idx_document_id (document_id)
) COMMENT='帖子关联文档表';

CREATE TABLE post_likes (
                            id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                            user_id BIGINT UNSIGNED NOT NULL COMMENT '点赞用户ID',
                            post_id BIGINT UNSIGNED NOT NULL COMMENT '帖子ID',
                            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
                            deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '取消点赞时间',
                            PRIMARY KEY (id),
                            UNIQUE KEY uk_user_post_like (user_id, post_id),
                            KEY idx_post_id (post_id)
) COMMENT='帖子点赞记录表';
-- 用户点赞唯一约束（避免重复点赞）
CREATE UNIQUE INDEX idx_likes_user_post_unique
    ON post_likes (user_id, post_id, (IF(deleted_at IS NULL, 1, NULL)));

CREATE TABLE notifications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '通知唯一标识ID (对应 reminderId)',
    receiver_id BIGINT UNSIGNED NOT NULL COMMENT '接收通知的用户ID (对应 recieverId)',
    type VARCHAR(50) NOT NULL COMMENT '通知类型 (对应 type): comment, like, favorite, chat',
    content TEXT COMMENT '通知内容 (对应 content)',    
    -- 状态与时间
    is_read TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读 (对应 isRead): 0-未读, 1-已读',
    source_id BIGINT UNSIGNED NOT NULL COMMENT '通知来源资源标识ID (对应 documentId或者postId)',
    source_type VARCHAR(50) NOT NULL COMMENT '通知来源资源类型 : document，post',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间 (对应 sendTime)',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记',
    PRIMARY KEY (id),
    KEY idx_receiver_status (receiver_id, is_read), -- 常用查询：查某用户的未读消息
    KEY idx_receiver_type (receiver_id, type),      -- 常用查询：查某用户的特定类型消息
    KEY idx_created_at (created_at)                 -- 用于按时间排序
) COMMENT='用户通知表';



