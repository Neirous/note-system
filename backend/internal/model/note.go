package model

import "time"

// Note 代表一个笔记实体，映射数据库中的笔记表
type Note struct {
	// ID 笔记的唯一标识符，主键且自增
	ID int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// Title 笔记的标题，最大长度200字符，不能为空
	Title string `gorm:"size:200;not null" json:"title"`
	// Content 笔记的内容，长文本类型，不能为空
	Content string `gorm:"type:longtext;not null" json:"content"`
	// CreatedAt 记录笔记创建时间，默认为当前时间戳
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 记录笔记更新时间，默认为当前时间戳并随更新改变
	UpdatedAt time.Time `json:"updated_at"`
	// IsDeleted 软删除标记，0表示未删除，1表示已删除
	IsDeleted int8 `gorm:"not null;default:0" json:"-"`
}

func (Note) TableName() string {
	return "notes"
}
