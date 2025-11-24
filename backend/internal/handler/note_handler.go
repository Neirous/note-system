package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"note-system/internal/common"
	"note-system/internal/service"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NoteHandler 笔记接口层结构体，依赖 NoteService 接口
type NoteHandler struct {
	svc service.NoteService // 依赖接口，不依赖具体实现
}

func NewNoteHandler(svc service.NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
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
	// 步骤4：返回成功响应
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

	// 步骤3：返回成功响应
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
	body := esQuery{Query: map[string]interface{}{"multi_match": map[string]interface{}{"query": q, "fields": []string{"title^2", "content"}}}, Size: 20}
	data, _ := json.Marshal(body)
	resp, err := http.Post(esURL+"/"+index+"/_search", "application/json", bytes.NewReader(data))
	if err == nil && resp.StatusCode == 200 {
		raw, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var parsed map[string]interface{}
		if json.Unmarshal(raw, &parsed) == nil {
			hits := parsed["hits"].(map[string]interface{})["hits"].([]interface{})
			out := make([]map[string]interface{}, 0, len(hits))
			for _, h := range hits {
				src := h.(map[string]interface{})["_source"].(map[string]interface{})
				out = append(out, map[string]interface{}{"id": src["id"], "title": src["title"], "content": src["content"], "updated_at": src["updated_at"]})
			}
			c.JSON(http.StatusOK, common.Success(map[string]interface{}{"list": out}))
			return
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
	c.JSON(http.StatusOK, common.Success(nil))
}
