package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"note-system/internal/common"
	"note-system/internal/model"
	"note-system/internal/rag"
	"note-system/internal/service"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// NoteHandler 笔记接口层结构体，依赖 NoteService 接口
type NoteHandler struct {
	svc service.NoteService
	rag *service.RAGService
}

func NewNoteHandler(svc service.NoteService, rag *service.RAGService) *NoteHandler {
	return &NoteHandler{svc: svc, rag: rag}
}

// 1. CreateNote 创建笔记接口（POST /api/note）
func (h *NoteHandler) CreateNote(c *gin.Context) {
	type CreateNoteRequest struct {
		Title   string `json:"title" binding:"required"` // binding:"required" 强制校验参数必传
		Content string `json:"content" binding:"required"`
	}

	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		//参数解析失败，返回统一失败响应
		c.JSON(http.StatusBadRequest, common.Fail("参数错误"+err.Error()))
		return
	}

	//调用service层处理业务
	note, err := h.svc.CreateNote(req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	_ = esIndexNote(note)
	if h.rag != nil {
		_ = h.rag.IndexNote(note)
	}
	// 返回成功响应
	c.JSON(http.StatusOK, common.Success(note))
}

// 2. GetNoteByID 查询笔记接口（GET /api/note/:id）
func (h *NoteHandler) GetNoteByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("笔记ID格式错误:"+err.Error()))
		return
	}

	note, err := h.svc.GetNoteById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, note)
}

// 3. UpdateNote 更新笔记接口（PUT /api/note/:id）
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("笔记ID格式错误:"+err.Error()))
		return
	}

	type UpdateNoteRequest struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("参数错误:"+err.Error()))
		return
	}
	err = h.svc.UpdateNote(id, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	if note, e := h.svc.GetNoteById(id); e == nil {
		_ = esIndexNote(note)
		if h.rag != nil {
			_ = h.rag.IndexNote(note)
		}
	}
	c.JSON(http.StatusOK, common.Success(nil))
}

// 4. DeleteNote 删除笔记接口（DELETE /api/note/:id）
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("笔记ID格式错误:"+err.Error()))
		return
	}
	err = h.svc.DeleteNote(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}

	// 从 ES 删除文档（忽略错误）
	_ = esDeleteNote(id)
	if h.rag != nil {
		_ = h.rag.DeleteVectorsByNoteID(id)
	}
	c.JSON(http.StatusOK, common.Success(nil))
}

// 5. ListNotes 分页查询笔记列表接口（GET /api/note/list）
func (h *NoteHandler) ListNotes(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")  // 默认页码1
	sizeStr := c.DefaultQuery("size", "10") // 默认每页10条
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("页码格式错误："+err.Error()))
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("每页条数格式错误："+err.Error()))
		return
	}

	// 步骤2：调用 Service 层
	list, total, err := h.svc.ListNotes(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}

	// 步骤3：封装列表响应数据（返回列表+总条数，方便前端分页）
	data := map[string]interface{}{
		"list":  list,
		"total": total,
	}
	c.JSON(http.StatusOK, common.Success(data))
}

// 简易搜索：优先尝试 ElasticSearch（http://localhost:9200/notes/_search），失败则回退 MySQL LIKE
func (h *NoteHandler) SearchNotes(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, common.Fail("缺少搜索关键词"))
		return
	}

	esURL := os.Getenv("ES_URL")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}
	index := os.Getenv("ES_INDEX")
	if index == "" {
		index = "notes"
	}

	// 1) 尝试 ES
	type esQuery struct {
		Query map[string]interface{} `json:"query"`
		Size  int                    `json:"size"`
	}
	body := esQuery{Query: map[string]interface{}{"query_string": map[string]interface{}{"query": "*" + q + "*", "fields": []string{"title^2", "content"}}}, Size: 20}
	data, _ := json.Marshal(body)
	resp, err := http.Post(esURL+"/"+index+"/_search", "application/json", bytes.NewReader(data))
	if err == nil && resp.StatusCode == 200 {
		raw, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var parsed map[string]interface{}
		if json.Unmarshal(raw, &parsed) == nil {
			hitsNode, ok := parsed["hits"].(map[string]interface{})
			if ok {
				items, ok2 := hitsNode["hits"].([]interface{})
				if ok2 {
					out := make([]map[string]interface{}, 0, len(items))
					for _, h := range items {
						m, ok3 := h.(map[string]interface{})
						if !ok3 {
							continue
						}
						src, ok4 := m["_source"].(map[string]interface{})
						if !ok4 {
							continue
						}
						out = append(out, map[string]interface{}{"id": src["id"], "title": src["title"], "content": src["content"], "updated_at": src["updated_at"]})
					}
					if len(out) > 0 {
						c.JSON(http.StatusOK, common.Success(map[string]interface{}{"list": out}))
						return
					}
				}
			}
		}
	}

	// 2) 回退 MySQL LIKE
	list, err := h.svc.SearchLike(q, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Success(map[string]interface{}{"list": list}))
}

func (h *NoteHandler) ListDeleted(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("页码格式错误："+err.Error()))
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("每页条数格式错误："+err.Error()))
		return
	}
	list, total, err := h.svc.ListDeleted(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	data := map[string]interface{}{
		"list":  list,
		"total": total,
	}
	c.JSON(http.StatusOK, common.Success(data))
}

func (h *NoteHandler) Restore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("笔记ID格式错误:"+err.Error()))
		return
	}
	if err := h.svc.Restore(id); err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	if note, e := h.svc.GetNoteById(id); e == nil {
		_ = esIndexNote(note)
		if h.rag != nil {
			_ = h.rag.IndexNote(note)
		}
	}
	c.JSON(http.StatusOK, common.Success(nil))
}

func (h *NoteHandler) HardDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Fail("笔记ID格式错误:"+err.Error()))
		return
	}
	if err := h.svc.HardDelete(id); err != nil {
		c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
		return
	}
	_ = esDeleteNote(id)
	if h.rag != nil {
		_ = h.rag.DeleteVectorsByNoteID(id)
	}
	c.JSON(http.StatusOK, common.Success(nil))
}

// RAG 搜索：问题向量 -> Pinecone TopK -> 返回片段
func (h *NoteHandler) RagSearch(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, common.Fail("缺少搜索关键词"))
		return
	}
	topK := 5
	if v := os.Getenv("RAG_TOPK"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			topK = n
		}
	}
	// 生成向量
	vecs, err := rag.EmbedBatch([]string{q})
	if err != nil || vecs == nil || len(vecs) == 0 {
		c.JSON(http.StatusOK, common.Success(map[string]interface{}{"list": []interface{}{}}))
		return
	}
	// Pinecone 查询
	res, err := rag.PineconeQueryTopK(vecs[0], topK)
	if err != nil || res == nil {
		c.JSON(http.StatusOK, common.Success(map[string]interface{}{"list": []interface{}{}}))
		return
	}
	// 组装输出
	out := make([]map[string]interface{}, 0, len(res.Matches))
	threshold := 0.7
	if s := os.Getenv("SIMILARITY_THRESHOLD"); s != "" {
		if f, e := strconv.ParseFloat(s, 64); e == nil {
			threshold = f
		}
	}
	for _, m := range res.Matches {
		if float64(m.Score) < threshold {
			continue
		}
		noteID, _ := toInt64(m.Metadata["note_id"])
		title, _ := m.Metadata["title"].(string)
		fragID, _ := m.Metadata["frag_id"].(string)
		out = append(out, map[string]interface{}{
			"note_id": noteID,
			"title":   title,
			"frag_id": fragID,
			"score":   m.Score,
			"link":    fmt.Sprintf("/?id=%d", noteID),
		})
	}
	c.JSON(http.StatusOK, common.Success(map[string]interface{}{"list": out}))
}

// 基于笔记的问答：检索片段 -> 构造上下文 -> 调用本地 LLM
func (h *NoteHandler) RagQA(c *gin.Context) {
	var body struct {
		Question string `json:"question"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Question == "" {
		c.JSON(http.StatusBadRequest, common.Fail("缺少问题"))
		return
	}
	// 检索片段
	vecs, err := rag.EmbedBatch([]string{body.Question})
	if err != nil || vecs == nil || len(vecs) == 0 {
		c.JSON(http.StatusOK, common.Success(map[string]interface{}{"answer": ""}))
		return
	}
	res, err := rag.PineconeQueryTopK(vecs[0], 3)
	contexts := make([]string, 0)
	if err == nil && res != nil {
		for _, m := range res.Matches {
			if t, ok := m.Metadata["content"].(string); ok {
				contexts = append(contexts, t)
			} else if t, ok := m.Metadata["title"].(string); ok {
				contexts = append(contexts, t)
			}
		}
	}
	// 构造提示词并调用 LLM
	llmURL := os.Getenv("LLM_URL")
	if llmURL == "" {
		// fallback：直接返回检索到的片段作为参考答案
		c.JSON(http.StatusOK, common.Success(map[string]interface{}{"answer": strings.Join(contexts, "\n\n")}))
		return
	}
	modelName := os.Getenv("LLM_MODEL")
	if modelName == "" {
		modelName = "phi-4"
	}
	payload := map[string]interface{}{
		"model":      modelName,
		"messages":   []map[string]string{{"role": "system", "content": "结合用户个人笔记回答问题，尽量引用原片段。"}, {"role": "user", "content": fmt.Sprintf("问题：%s\n上下文：%s", body.Question, strings.Join(contexts, "\n"))}},
		"max_tokens": 512,
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(llmURL, "application/json", bytes.NewReader(b))
	if err != nil {
		c.JSON(http.StatusOK, common.Success(map[string]interface{}{"answer": strings.Join(contexts, "\n\n")}))
		return
	}
	defer resp.Body.Close()
	var parsed map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&parsed)
	// 兼容 chat/completions 的返回结构
	answer := extractAnswer(parsed)
	c.JSON(http.StatusOK, common.Success(map[string]interface{}{"answer": answer}))
}

func (h *NoteHandler) MockLLM(c *gin.Context) {
	var req map[string]interface{}
	_ = c.BindJSON(&req)
	ans := "虚拟内存通过页表将虚拟地址映射到物理地址。操作系统维护多级页表，TLB 用于加速地址转换，缺页时通过页置换将数据从磁盘载入内存。"
	c.JSON(http.StatusOK, map[string]interface{}{"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": ans}}}})
}

func extractAnswer(p map[string]interface{}) string {
	if p == nil {
		return ""
	}
	if choices, ok := p["choices"].([]interface{}); ok && len(choices) > 0 {
		if m, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := m["message"].(map[string]interface{}); ok {
				if c, ok := msg["content"].(string); ok {
					return c
				}
			}
			if txt, ok := m["text"].(string); ok {
				return txt
			}
		}
	}
	return ""
}

func toInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case float64:
		return int64(t), true
	case string:
		if n, err := strconv.ParseInt(t, 10, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}

// ES 索引与删除（简单 HTTP 实现）
func esIndexNote(note *model.Note) error {
	esURL := os.Getenv("ES_URL")
	index := os.Getenv("ES_INDEX")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}
	if index == "" {
		index = "notes"
	}
	if note == nil {
		return nil
	}
	payload := map[string]interface{}{
		"id":         note.ID,
		"title":      note.Title,
		"content":    note.Content,
		"updated_at": note.UpdatedAt,
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", esURL+"/"+index+"/_doc/"+strconv.FormatInt(note.ID, 10), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func esDeleteNote(id int64) error {
	esURL := os.Getenv("ES_URL")
	index := os.Getenv("ES_INDEX")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}
	if index == "" {
		index = "notes"
	}
	if id == 0 {
		return nil
	}
	req, _ := http.NewRequest("DELETE", esURL+"/"+index+"/_doc/"+strconv.FormatInt(id, 10), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// 管理端：删除全部笔记并清空片段与 ES 索引
func (h *NoteHandler) PurgeAll(c *gin.Context) {
	page, size := 1, 100
	for {
		list, _, err := h.svc.ListNotes(page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Fail(err.Error()))
			return
		}
		if len(list) == 0 {
			break
		}
		for _, n := range list {
			_ = h.svc.HardDelete(n.ID)
			_ = esDeleteNote(n.ID)
		}
		page++
	}
	if h.rag != nil {
		_ = h.rag.PurgeFragments()
	}
	_ = rag.PineconeDeleteAll()
	esURL := os.Getenv("ES_URL")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}
	index := os.Getenv("ES_INDEX")
	if index == "" {
		index = "notes"
	}
	payload := map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}}
	b, _ := json.Marshal(payload)
	_, _ = http.Post(esURL+"/"+index+"/_delete_by_query", "application/json", bytes.NewReader(b))
	c.JSON(http.StatusOK, common.Success(nil))
}
