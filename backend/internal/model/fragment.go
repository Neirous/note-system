package model

import "time"

type Fragment struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	NoteID    int64     `json:"note_id" index:"idx_note_id"`
	FragID    string    `gorm:"type:varchar(64);uniqueIndex" json:"frag_id"`
	Content   string    `gorm:"type:longtext" json:"content"`
	IsCode    bool      `json:"is_code"`
	VectorID  string    `gorm:"type:varchar(128)" json:"vector_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
