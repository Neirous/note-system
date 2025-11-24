package service

import (
	"errors"
	"note-system/internal/model"
	"note-system/internal/repository"

	"gorm.io/gorm"
)

type NoteService interface {
	// CreateNote 创建笔记，接收标题和内容，返回创建后的笔记和错误
	CreateNote(title, content string) (*model.Note, error)
	// GetNoteByID 根据ID查询笔记，接收ID，返回笔记和错误
	GetNoteById(id int64) (*model.Note, error)
	// UpdateNote 更新笔记，接收ID、新标题、新内容，返回错误
	UpdateNote(id int64, newTitle, newContent string) error
	// DeleteNote 删除笔记，接收ID，返回错误
	DeleteNote(id int64) error
	// ListNotes 分页查询笔记列表，接收页码和每页大小，返回笔记列表、总条数、错误
	ListNotes(page, size int) ([]model.Note, int64, error)
	ListDeleted(page, size int) ([]model.Note, int64, error)
	Restore(id int64) error
	HardDelete(id int64) error
	SearchLike(q string, limit int) ([]model.Note, error)
}

type noteService struct {
	repo repository.NoteRepository
}

// CreateNote implements NoteService.
func (n *noteService) CreateNote(title string, content string) (*model.Note, error) {
	//业务校验：标题不能为空
	if title == "" {
		return nil, errors.New("笔记标题不能为空")
	}
	//构建Note模型（业务层组装数据，Repository只负责存储）
	note := &model.Note{
		Title:   title,
		Content: content,
	}
	//调用Repository层的Create方法，存储数据
	if err := n.repo.Create(note); err != nil {
		return nil, errors.New("创建笔记失败" + err.Error())
	}
	return note, nil
}

// Delete implements NoteService.
func (n *noteService) DeleteNote(id int64) error {
	// 业务校验：ID 必须大于 0
	if id <= 0 {
		return errors.New("笔记ID不合法(必须大于0)")
	}
	_, err := n.repo.GetByID(id)
	if err != nil {
		return errors.New("未查询到笔记" + err.Error())
	}
	// 调用 Repository 层删除（逻辑删除）
	if err := n.repo.Delete(id); err != nil {
		return errors.New("删除笔记失败：" + err.Error())
	}

	return nil
}

// GetNoteById implements NoteService.
func (n *noteService) GetNoteById(id int64) (*model.Note, error) {
	if id <= 0 {
		return nil, errors.New("笔记ID不合法(必须大于0)")
	}
	// 调用 Repository 层查询
	note, err := n.repo.GetByID(id)
	if err != nil {
		// 区分错误类型：如果是记录不存在，返回明确的业务错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("未找到该笔记(可能已删除或ID不存在)")
		}
		return nil, errors.New("查询笔记失败：" + err.Error())
	}
	return note, nil
}

// ListNotes implements NoteService.
func (n *noteService) ListNotes(page int, size int) ([]model.Note, int64, error) {
	// 业务校验：页码至少为 1，每页大小至少为 1，最多为 100（避免查询过多数据）
	if page < 1 {
		page = 1 // 默认页码 1
	}
	if size < 1 || size > 100 {
		size = 10 // 默认每页 10 条
	}

	list, total, err := n.repo.List(page, size)
	if err != nil {
		return nil, 0, errors.New("查询笔记列表失败:" + err.Error())
	}
	return list, total, nil
}

func (n *noteService) ListDeleted(page int, size int) ([]model.Note, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	list, total, err := n.repo.ListDeleted(page, size)
	if err != nil {
		return nil, 0, errors.New("查询回收站失败:" + err.Error())
	}
	return list, total, nil
}

func (n *noteService) Restore(id int64) error {
	if id <= 0 {
		return errors.New("笔记ID不合法(必须大于0)")
	}
	return n.repo.Restore(id)
}

func (n *noteService) HardDelete(id int64) error {
	if id <= 0 {
		return errors.New("笔记ID不合法(必须大于0)")
	}
	return n.repo.HardDelete(id)
}

func (n *noteService) SearchLike(q string, limit int) ([]model.Note, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return n.repo.SearchLike(q, limit)
}

// UpdateNote implements NoteService.
func (n *noteService) UpdateNote(id int64, newTitle string, newContent string) error {
	// 业务校验 1：ID 必须大于 0
	if id <= 0 {
		return errors.New("笔记ID不合法（必须大于0）")
	}
	// 业务校验 2：新标题不能为空
	if newTitle == "" {
		return errors.New("笔记标题不能为空")
	}

	// 先查询笔记是否存在（避免更新不存在的笔记）
	note, err := n.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未找到该笔记，无法更新")
		}
		return errors.New("查询笔记失败：" + err.Error())
	}

	// 组装更新数据
	note.Title = newTitle
	note.Content = newContent

	// 调用 Repository 层更新
	if err := n.repo.Update(note); err != nil {
		return errors.New("更新笔记失败：" + err.Error())
	}

	return nil
}

func NewNoteService(repo repository.NoteRepository) NoteService {
	return &noteService{repo: repo}
}
