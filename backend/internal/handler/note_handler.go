package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"note-system/internal/common"
	"note-system/internal/model"
	"note-system/internal/rag"
	"note-system/internal/service"
	"os"
	"strconv"
	"strings"
	"time"

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

// 文件夹功能已移除

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
	maxTokens := 2048
	if s := os.Getenv("LLM_MAX_TOKENS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxTokens = n
		}
	}
	payload := map[string]interface{}{
		"model":      modelName,
		"messages":   []map[string]string{{"role": "system", "content": "结合用户个人笔记回答问题，尽量引用原片段。"}, {"role": "user", "content": fmt.Sprintf("问题：%s\n上下文：%s", body.Question, strings.Join(contexts, "\n"))}},
		"max_tokens": maxTokens,
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

// 批量生成中文 IT 笔记（使用 Go 原生字符串直接写入）
func (h *NoteHandler) SeedCNNotes(c *gin.Context) {
	type Item struct{ Title, Content string }
	items := []Item{
		{Title: "操作系统：进程与线程", Content: "# 操作系统：进程与线程\n\n进程负责资源管理，线程负责调度。现代内核通过多级页表与 TLB 加速地址转换，调度器结合优先级与时间片实现公平竞争。上下文切换保存与恢复寄存器、内核栈与页表指针；频繁切换会带来缓存失效与额外开销。\n\n示例：\n```c\npthread_create(&tid, NULL, worker, NULL);\n```\n\n设计要点：减少共享可变状态，以消息或事件驱动合并竞争；避免巨锁，必要时用读写锁分离；I/O 密集任务配合线程池与异步机制，减小调度压力。"},
		{Title: "网络：TCP 握手/挥手与拥塞控制", Content: "# 网络：TCP 握手/挥手与拥塞控制\n\n连接建立采用三次握手以确认双方收发能力并同步初始序号；断开则四次挥手确保半关闭后缓冲区数据完成发送。可靠性通过滑动窗口、重传与累计确认保障。拥塞控制阶段包含慢启动、拥塞避免、快速重传与快速恢复，不同实现细节在 Reno/NewReno/CUBIC 上有所差异。生产环境中需观察 RTT、重传率与队列时延，结合 BBR 或 ECN 减缓排队延迟。"},
		{Title: "HTTP/HTTPS 与 TLS", Content: "# HTTP/HTTPS 与 TLS\n\nHTTP 是无状态的请求-响应协议，语义清晰但明文传输。HTTPS 在其上叠加 TLS，利用握手阶段协商套件并完成身份认证与密钥交换，后续通信以对称加密保障机密性与完整性。部署上应启用 HSTS 防止降级，中间件需正确处理 SNI 与证书链；客户端侧要校验主机名匹配，后端轮换证书时兼顾 OCSP 与缓存。"},
		{Title: "Go 并发：goroutine/channel 深入", Content: "# Go 并发：goroutine/channel 深入\n\n调度器以 M-P-G 模型运行，goroutine 切换开销远低于线程。channel 适合表达拥有者转移与背压，缓冲区用于削峰但过大可能掩盖阻塞。实践中以 context 控制取消与超时，模块边界以不可变数据传递，避免共享内存。\n\n示例：\n```go\nfunc main(){\n  ch := make(chan int, 8)\n  go func(){ for i:=0;i<100;i++{ ch<-i } close(ch) }()\n  for v := range ch { fmt.Println(v) }\n}\n```"},
		{Title: "Go 内存与 GC 调优", Content: "# Go 内存与 GC 调优\n\nGo 的 GC 采用并发标记清除，触发与堆增长相关。逃逸分析决定对象分配位置；栈上分配可减少 GC 压力。优化策略包括减少临时对象、重用大缓冲、避免在热路径上频繁分配；使用 `sync.Pool` 需衡量一致性与可见性。压测时结合 GOMAXPROCS 与 `GODEBUG=gctrace=1` 观察暂停时间与周期。"},
		{Title: "MySQL 索引与事务", Content: "# MySQL 索引与事务\n\nInnoDB 以聚簇索引存储主键，二级索引指向主键形成回表。合理设计前缀与覆盖索引可显著降低 I/O。事务隔离以 RR 常见，MVCC 通过 undo log 与快照读取实现可重复读。热点更新可拆分批次并控制锁粒度，长事务需避免以免阻塞 purge 与增长历史版本。"},
		{Title: "PostgreSQL 特性与查询优化", Content: "# PostgreSQL 特性与查询优化\n\nJSONB 支持索引与高效操作，窗口函数在统计与分页中极其强大。计划器基于代价估算选择 Hash/Sort Merge 等策略；合理的统计与 `ANALYZE` 能显著提升性能。CTE 在新版本可 inline，过度使用可能限制优化。并行查询需评估工作进程数量与数据倾斜。"},
		{Title: "Redis 机制与雪崩防护", Content: "# Redis 机制与雪崩防护\n\n数据结构丰富：String/Hash/List/Set/ZSet；过期与淘汰策略影响命中与内存占用。防止雪崩可采用随机过期、分片锁与二级缓存；击穿用互斥锁或逻辑过期；穿透通过布隆过滤器或参数校验。持久化 RDB/AOF 结合使用，主从与哨兵保障高可用。"},
		{Title: "Kafka 架构与一致性", Content: "# Kafka 架构与一致性\n\n主题分区与副本形成高吞吐日志系统，生产者可选择幂等与事务写入以实现精确一次。消费者组以位移管理并发处理，重平衡需快速恢复。跨区多活场景下要关注延迟与顺序保证；消息模式建议事件化，避免过度耦合。"},
		{Title: "Docker 镜像与多阶段构建", Content: "# Docker 镜像与多阶段构建\n\n分层镜像适合缓存复用但层数过多会增加拉取时间。多阶段构建能显著减小最终镜像，尽量使用静态编译与最小基础镜像；限制 `RUN` 合并命令减少层。资源控制以 cgroup 为基础，生产部署结合安全基线与镜像签名。"},
		{Title: "Kubernetes 调度与弹性", Content: "# Kubernetes 调度与弹性\n\n核心对象包括 Pod/Deployment/Service/Ingress；调度器考虑节点亲和与资源请求，HPA 基于指标进行自动扩缩容。正确设置 Requests/Limits 可提升稳定性；就绪与存活探针确保滚动更新。网络策略用于隔离流量，配合 ServiceMesh 完成细粒度治理。"},
		{Title: "REST 与 gRPC 设计抉择", Content: "# REST 与 gRPC 设计抉择\n\nREST 易于调试与跨语言互通，适合公开 API；gRPC 以 Proto 定义强类型契约，HTTP/2 与流式能力在内网高效。统一错误模型与版本化策略是长期演进基础；速率限制与幂等写入避免故障放大。"},
		{Title: "JWT/OAuth2 实战要点", Content: "# JWT/OAuth2 实战要点\n\nJWT 自包含但需控制大小与过期；签名算法与密钥轮换要到位。OAuth2 授权码模式结合 PKCE 提升安全，刷新令牌须具备撤销与黑名单管理。服务端保存会话快照以便风险控制与审计。"},
		{Title: "Web 安全：XSS/CSRF/SQL", Content: "# Web 安全：XSS/CSRF/SQL\n\n前端输出严格转义与 CSP 白名单，表单使用 SameSite 与 CSRF Token 防伪造；数据库操作采用参数化与权限最小化，审计日志记录关键行为。漏洞响应流程要包含回滚、封禁与通报。"},
		{Title: "可观测性：日志/指标/追踪", Content: "# 可观测性：日志/指标/追踪\n\n结构化日志便于检索与聚合；指标体系以 RED/USE 为指导划分服务层关键指标；分布式追踪帮助定位跨服务瓶颈。采样策略应动态调整，避免高峰期 IO 压力。"},
		{Title: "Nginx 反代与限流", Content: "# Nginx 反代与限流\n\n示例：\n```nginx\nhttp { limit_req_zone $binary_remote_addr zone=api:10m rate=5r/s; }\nserver { location /api { limit_req zone=api burst=10 nodelay; } }\n```\n\n结合缓存与动态上游权重可提升整体韧性；在链路尾部对超时与连接数实施硬限制，避免服务被压垮。"},
		{Title: "Git 工作流与提交规范", Content: "# Git 工作流与提交规范\n\n在团队内选择 Trunk-based 或 GitFlow，保证发布节奏与分支策略一致。采用语义化提交（feat/fix/docs）与规范化 PR 模板，自动化检查风格与冲突。"},
		{Title: "CI/CD 实践", Content: "# CI/CD 实践\n\n流水线包含构建、测试、审查与发布；部署策略以蓝绿或金丝雀降低风险。失败回滚需脚本化并保留工件，版本标记与变更日志可追溯。"},
		{Title: "性能优化：CPU/IO/内存", Content: "# 性能优化：CPU/IO/内存\n\n结合 pprof/trace 精确定位热点；减少系统调用与锁竞争；网络层采用批处理与零拷贝。内存抖动可通过对象复用与池化缓解。"},
		{Title: "Gin 最佳实践", Content: "# Gin 最佳实践\n\n中间件统一日志与错误响应，参数校验与绑定保证入口可靠；为跨域与安全头设置合理策略。"},
		{Title: "并发控制：锁/原子/无锁", Content: "# 并发控制：锁/原子/无锁\n\n选择合适的数据结构与粒度；热点路径采用原子与 ring buffer；避免长持锁阻塞 GC 与调度。"},
		{Title: "算法：排序与搜索", Content: "# 算法：排序与搜索\n\n对数据规模与稳定性进行权衡；在工程场景下配合缓存与批量接口减少复杂度。"},
		{Title: "设计模式精要", Content: "# 设计模式精要\n\n工厂/策略/观察者在解耦与扩展性上效果明显；避免过度设计，保持语义清晰。"},
		{Title: "Linux 工具箱", Content: "# Linux 工具箱\n\ntop/iostat/vmstat 观察资源，ss/tcpdump 分析网络；systemctl/journalctl 管理服务与日志。"},
		{Title: "存储与文件系统", Content: "# 存储与文件系统\n\n理解 EXT4/XFS 特性与写放大；合理选择 RAID/快照与备份策略。"},
		{Title: "微服务治理", Content: "# 微服务治理\n\n以领域划分服务，注册发现与配置中心保障弹性；熔断/限流/重试是稳定性三板斧。"},
		{Title: "接口稳定与兼容", Content: "# 接口稳定与兼容\n\n版本策略与幂等语义配合重试，灰度发布与回滚提升安全性。"},
		{Title: "测试金字塔", Content: "# 测试金字塔\n\n单测覆盖核心逻辑，集成测校验协作，端到端保证真实场景；关注可维护与执行时间。"},
		{Title: "Service Mesh 与边车", Content: "# Service Mesh 与边车\n\n通过边车代理实现统一的流量管理、可观测与安全策略；Istio 在路由、熔断、限流及 mTLS 上提供丰富能力。合理配置 sidecar 资源与过滤器链避免性能下降。"},
		{Title: "零信任架构", Content: "# 零信任架构\n\n核心是持续身份验证与最小权限；结合设备与上下文进行动态评估。网络分段与细粒度策略配合审计形成闭环。"},
		{Title: "SRE 指标体系", Content: "# SRE 指标体系\n\nSLI/SLO/SLA 的协同定义是可靠性治理的核心；误差预算指导变更速率。事件响应流程需覆盖预案、演练与复盘。"},
		{Title: "RTO/RPO 与容灾", Content: "# RTO/RPO 与容灾\n\n恢复时间目标与数据丢失目标决定技术选型；冷/温/热备架构的成本与恢复速度差异显著。"},
		{Title: "混沌工程", Content: "# 混沌工程\n\n通过受控实验验证系统在故障下的恢复与隔离能力；设计指标与回滚阈值，避免引入级联风险。"},
		{Title: "分布式一致性：CAP/BASE", Content: "# 分布式一致性：CAP/BASE\n\n理解一致性、可用性与分区容错的权衡；BASE 倡导最终一致性与柔性事务，工程中以补偿与重试确保业务正确。"},
		{Title: "共识算法：Raft", Content: "# 共识算法：Raft\n\n领导者选举、日志复制与安全性保证了易理解与工程可落地；快照与日志截断控制存储膨胀。"},
		{Title: "分布式事务：Saga/TCC", Content: "# 分布式事务：Saga/TCC\n\nSaga 以长事务拆分为本地事务与补偿；TCC 明确 Try/Confirm/Cancel 接口。选择受业务一致性强弱与性能影响。"},
		{Title: "事件驱动与溯源", Content: "# 事件驱动与溯源\n\n采用事件作为系统状态变化的唯一事实来源；通过重放还原对象状态，适合审计与回滚场景。"},
		{Title: "DDD 与六边形架构", Content: "# DDD 与六边形架构\n\n以领域模型划分限界上下文，适配器隔离外部系统；保持核心域与应用服务纯净。"},
		{Title: "网络 I/O：epoll/多路复用", Content: "# 网络 I/O：epoll/多路复用\n\n在大并发场景下以边缘触发与批量收发提升吞吐；注意环形缓冲与半包处理。"},
		{Title: "C10K 到 C10M", Content: "# C10K 到 C10M\n\n从多进程到事件驱动与用户态网络栈的演进；减少拷贝与锁争用是突破瓶颈的关键。"},
		{Title: "Rust 安全与所有权", Content: "# Rust 安全与所有权\n\n所有权、借用与生命周期通过编译期保证内存安全，无需 GC。零成本抽象让泛型与 trait 在性能上可与 C/C++ 比肩。并发以 Send/Sync 限定跨线程共享，避免数据竞争。工程上结合 `cargo` 工作区、`clippy` 与 `rustfmt` 保持质量；FFI 需注意 ABI 与不安全块的边界。"},
		{Title: "WebAssembly 应用场景", Content: "# WebAssembly 应用场景\n\nWasm 提供接近原生的沙箱执行环境，适用于前端重计算、插件体系与边缘计算。通过 WASI 可访问文件与网络等系统接口；运行时如 Wasmtime/Wasmer 便于在服务端托管。将计算逻辑以 Wasm 分发可降低语言绑定成本，版本管理依赖模块签名与能力声明。"},
		{Title: "前端性能优化实践", Content: "# 前端性能优化实践\n\n关键路径资源内联与延迟加载减少首次渲染时间；图片采用现代格式与按需加载；减少重排与回流，合理使用虚拟列表。监控以 FCP/LCP/CLS/TBT 指标度量，结合 Web Vitals 收集数据。构建层面启用代码分割与缓存哈希，避免大包阻塞。"},
		{Title: "移动端网络优化", Content: "# 移动端网络优化\n\n弱网下采用连接复用与请求合并，压缩与差量同步减少带宽消耗。链路采用超时与重试回退策略，避免雪崩。CDN 边缘缓存与预取提升体验；QoS 限制后台流量防止系统杀进程。统计上采集 RTT、丢包与首包时间做画像。"},
		{Title: "API 可用性设计", Content: "# API 可用性设计\n\n统一错误模型与返回码，语义清晰且可扩展；分页与过滤约定一致，避免歧义。速率限制与幂等键配合重试，保障在故障下的可恢复性。为客户端提供稳定契约与版本兼容策略；文档自动化生成并与测试集成。"},
		{Title: "缓存一致性策略", Content: "# 缓存一致性策略\n\n写入路径采用先删后写或延迟双删，结合逻辑过期防止穿透。多副本一致性以订阅通知或 CDC 事件驱动更新；热点键采用分片锁与局部失效。监控命中率与回源延迟，控制内存占用与淘汰策略。"},
		{Title: "向量数据库入门", Content: "# 向量数据库入门\n\n基于 ANN 的近似最近邻检索，如 HNSW/IVF/PQ；索引构建在召回与存储之间取舍。向量维度与归一化影响距离度量；融合元数据过滤实现语义检索。生产中评估吞吐、延迟与召回率，分片与副本策略保障扩展性。"},
		{Title: "日志采集与清洗", Content: "# 日志采集与清洗\n\nAgent 采集后以缓冲与批量压缩传输，避免高峰期阻塞。清洗阶段进行结构化、脱敏与落盘归档；管道故障回退到本地队列。检索层结合索引模板与冷热分层，控制成本并保障可用性。"},
		{Title: "数据建模与分区", Content: "# 数据建模与分区\n\n事实表与维表基于业务查询路径设计，避免雪花模型下过度关联。分区策略按时间或范围划分，减少扫描与提升维护效率；冷热分层与归档策略控制数据生命周期。"},
		{Title: "灰度发布与回滚", Content: "# 灰度发布与回滚\n\n以少量流量逐步导入新版本，观测关键指标与错误率决定推进或回滚。配合特性开关减少风险面；回滚脚本与数据兼容策略必须可重复验证。"},
		{Title: "消息顺序与幂等", Content: "# 消息顺序与幂等\n\n同键顺序依赖分区与单并发处理；跨分区需要局部有序或重排机制。幂等以业务键与去重窗口实现，避免重复消费带来的副作用。"},
		{Title: "数据库分库分表", Content: "# 数据库分库分表\n\n按用户或业务键做水平拆分，路由层负责分发与聚合。跨分片事务需要补偿或两阶段协议；统计与报表通过离线汇总。迁移过程保证双写与校验对账。"},
		{Title: "Snowflake 与 ID 生成", Content: "# Snowflake 与 ID 生成\n\n时间戳 + 机器号 + 序列构成趋势递增 ID，便于索引插入与日志关联。时钟漂移与回拨需要防护；多机部署依赖号段分配或中心协调。"},
		{Title: "时序数据库实践", Content: "# 时序数据库实践\n\n写入高吞吐与查询按时间窗口优化；标签维度控制基数。压缩编码如 Gorilla 提升存储效率；下采样与保留策略管理历史数据。"},
		{Title: "边缘计算与 CDN", Content: "# 边缘计算与 CDN\n\n在靠近用户的节点执行计算与缓存，降低时延与带宽成本。函数计算与边缘 KV 组合实现动态路由与 A/B 测试。观测与发布管控确保一致性。"},
		{Title: "负载均衡算法", Content: "# 负载均衡算法\n\n常见有轮询、最小连接、加权与一致性哈希；健康检查与熔断策略保障稳定。对长连接与会话粘性做特殊处理。"},
		{Title: "存储压缩与编码", Content: "# 存储压缩与编码\n\n列式存储在分析型场景表现优异，结合字典编码与位图压缩降低空间。日志采用结构化与分块索引提升检索速度。"},
		{Title: "云原生成本优化", Content: "# 云原生成本优化\n\n通过自动扩缩容与预留实例控制计算成本；对象存储分层与生命周期策略降低存储费用。链路优化减少出口流量。监控与告警围绕单位成本指标。"},
		{Title: "安全扫描与合规", Content: "# 安全扫描与合规\n\n依赖漏洞扫描与镜像签名构建可信供应链；合规上遵循数据保护与访问控制要求。对密钥与证书的生命周期实施审计。"},
		{Title: "密钥管理与轮换", Content: "# 密钥管理与轮换\n\n集中式 KMS 提供加密材料托管与审计；密钥轮换需无感并支持灰度。最小权限与分层访问控制降低泄漏风险。"},
	}

	// 标题去重
	existing := map[string]struct{}{}
	page, size := 1, 100
	for {
		list, _, err := h.svc.ListNotes(page, size)
		if err != nil || len(list) == 0 {
			break
		}
		for _, n := range list {
			existing[n.Title] = struct{}{}
		}
		page++
		if page > 50 { // 防止过多循环
			break
		}
	}

	rand.Seed(time.Now().UnixNano())
	created := 0
	for _, it := range items {
		if _, ok := existing[it.Title]; ok {
			continue
		}
		note, err := h.svc.CreateNote(it.Title, it.Content)
		if err != nil {
			continue
		}
		days := rand.Intn(120) + 1
		hour := rand.Intn(24)
		min := rand.Intn(60)
		sec := rand.Intn(60)
		t := time.Now().AddDate(0, 0, -days).Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second)
		_ = h.svc.SetNoteTimes(note.ID, t, t)
		_ = esIndexNote(note)
		if h.rag != nil {
			_ = h.rag.IndexNote(note)
		}
		created++
	}
	c.JSON(http.StatusOK, common.Success(map[string]interface{}{"created": created}))
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
