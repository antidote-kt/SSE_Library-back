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


CREATE TABLE favorites (
   id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '收藏记录ID',
   user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
   document_id BIGINT UNSIGNED NOT NULL COMMENT '被收藏对象ID',
   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
   PRIMARY KEY (id),
   UNIQUE KEY uk_user_favorite (user_id, document_id),
   KEY idx_favorite (document_id)
) COMMENT='文档收藏记录表';

CREATE TABLE comments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  content TEXT NOT NULL COMMENT '评论内容',
  parent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '父评论ID (用于回复)',
  document_id BIGINT UNSIGNED NOT NULL COMMENT '被评论对象ID (关联文档或课程)',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '评论时间',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_comment (document_id),
  KEY idx_parent_id (parent_id)
) COMMENT='通用评论表（书籍/课程）';

CREATE TABLE view_histories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    document_id BIGINT UNSIGNED NOT NULL COMMENT '被浏览对象ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后浏览时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
    PRIMARY KEY (id),
    KEY idx_user_view (user_id, document_id)
) COMMENT='通用浏览历史表';


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

CREATE TABLE document_tag (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
  document_id BIGINT UNSIGNED NOT NULL COMMENT '文档ID',
  tag_id BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）'
) COMMENT='文档与标签关系映射表';


CREATE TABLE categories (
     id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '分类唯一标识ID，一级分类',
     name VARCHAR(50) NOT NULL COMMENT '分类名称/课程名称',
     is_course tinyint NOT NULL COMMENT '1-细分课程或课程大类,0-非课程类别',
     description TEXT COMMENT '分类描述，详细说明分类的用途和包含的内容',
     parent_id BIGINT UNSIGNED DEFAULT NULL COMMENT '父分类ID，关联本表，用于实现层级分类，NULL表示顶级分类，可以是泛课程也可以是非课程大类',
     thumbnail VARCHAR(500) COMMENT '课程缩略图文件路径，存储课程封面图片的URL或文件路径',
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分类创建时间',
     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
     deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '软删除标记，（NULL表示未删除）',
     KEY idx_parent_id (parent_id)
) COMMENT='分类表';
