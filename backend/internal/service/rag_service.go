package service

import (
	"crypto/sha1"
	"encoding/hex"
	"note-system/internal/model"
	"note-system/internal/rag"
	"os"

	"gorm.io/gorm"
)

type RAGService struct {
	db *gorm.DB
}

func NewRAGService(db *gorm.DB) *RAGService { return &RAGService{db: db} }

func (r *RAGService) IndexNote(note *model.Note) error {
	cands := rag.SplitMarkdown(note.Content)
	texts := make([]string, 0, len(cands))
	ids := make([]string, 0, len(cands))
	metas := make(map[string]map[string]interface{})
	for i, c := range cands {
		fid := fragID(note.ID, i, c.Content)
		f := &model.Fragment{NoteID: note.ID, FragID: fid, Content: c.Content, IsCode: c.IsCode}
		_ = r.db.Where("frag_id = ?", fid).Delete(&model.Fragment{}).Error
		if err := r.db.Create(f).Error; err != nil {
			continue
		}
		texts = append(texts, c.Content)
		ids = append(ids, fid)
		metas[fid] = map[string]interface{}{"note_id": note.ID, "frag_id": fid, "title": note.Title, "content": c.Content}
	}
	if len(texts) == 0 {
		return nil
	}
	vecs, err := rag.EmbedBatch(texts)
	if err != nil || vecs == nil {
		return nil
	}
	vmap := make(map[string][]float32)
	for i, id := range ids {
		if i < len(vecs) {
			vmap[id] = vecs[i]
		}
	}
	_ = rag.PineconeUpsert(getEnv("PINECONE_INDEX"), vmap, metas)
	return nil
}

func (r *RAGService) PurgeFragments() error {
	if r.db == nil {
		return nil
	}
	return r.db.Exec("DELETE FROM fragments").Error
}

func (r *RAGService) DeleteVectorsByNoteID(noteID int64) error {
	if r.db == nil || noteID <= 0 {
		return nil
	}
	var frags []model.Fragment
	if err := r.db.Model(&model.Fragment{}).Where("note_id = ?", noteID).Find(&frags).Error; err != nil {
		return err
	}
	ids := make([]string, 0, len(frags))
	for _, f := range frags {
		if f.FragID != "" {
			ids = append(ids, f.FragID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return rag.PineconeDeleteByIDs(ids)
}

func fragID(noteID int64, i int, content string) string {
	h := sha1.Sum([]byte(content))
	return hex.EncodeToString(h[:])
}

func getEnv(key string) string { return os.Getenv(key) }
