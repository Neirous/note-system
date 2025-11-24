package repository

import (
	"note-system/internal/model"

	"gorm.io/gorm"
)

type NoteRepository interface {
	Create(note *model.Note) error
	GetByID(id int64) (*model.Note, error)
	Update(note *model.Note) error
	Delete(id int64) error
	List(page, size int) ([]model.Note, int64, error)
	ListDeleted(page, size int) ([]model.Note, int64, error)
	Restore(id int64) error
	HardDelete(id int64) error
	SearchLike(q string, limit int) ([]model.Note, error)
}

type noteRepo struct {
	db *gorm.DB
}

// Create implements NoteRepository.
func (n *noteRepo) Create(note *model.Note) error {
	return n.db.Create(note).Error
}

// Delete implements NoteRepository.
func (n *noteRepo) Delete(id int64) error {
	return n.db.Model(&model.Note{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

// GetByID implements NoteRepository.
func (n *noteRepo) GetByID(id int64) (*model.Note, error) {
	var note model.Note
	err := n.db.Model(&model.Note{}).Where("id = ? AND is_deleted = 0", id).First(&note).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// List implements NoteRepository.
func (n *noteRepo) List(page int, size int) ([]model.Note, int64, error) {
	var (
		noteList []model.Note
		total    int64
	)
	err := n.db.Model(&model.Note{}).Where("is_deleted = 0").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = n.db.Model(&model.Note{}).
		Where("is_deleted = 0").
		Order("updated_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&noteList).Error
	if err != nil {
		return nil, 0, err
	}
	return noteList, total, nil
}

func (n *noteRepo) ListDeleted(page int, size int) ([]model.Note, int64, error) {
	var (
		noteList []model.Note
		total    int64
	)
	err := n.db.Model(&model.Note{}).Where("is_deleted = 1").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = n.db.Model(&model.Note{}).
		Where("is_deleted = 1").
		Order("updated_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&noteList).Error
	if err != nil {
		return nil, 0, err
	}
	return noteList, total, nil
}

func (n *noteRepo) Restore(id int64) error {
	return n.db.Model(&model.Note{}).
		Where("id = ?", id).
		Update("is_deleted", 0).Error
}

func (n *noteRepo) HardDelete(id int64) error {
	return n.db.Delete(&model.Note{}, id).Error
}

func (n *noteRepo) SearchLike(q string, limit int) ([]model.Note, error) {
	var list []model.Note
	like := "%" + q + "%"
	err := n.db.Model(&model.Note{}).
		Where("is_deleted = 0 AND (title LIKE ? OR content LIKE ?)", like, like).
		Order("updated_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Update implements NoteRepository.
func (n *noteRepo) Update(note *model.Note) error {
	return n.db.Model(note).
		Where("id = ? AND is_deleted = 0", note.ID).
		Updates(map[string]interface{}{
			"title":   note.Title,
			"content": note.Content,
		}).Error
}

func NewNoteRepo(db *gorm.DB) NoteRepository {
	return &noteRepo{db: db}
}
