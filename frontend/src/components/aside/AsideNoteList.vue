<template>
  <div class="aside-note-list">
    <div class="search-row">
      <el-input v-model="keyword" size="small" placeholder="搜索笔记（标题/内容）" clearable :prefix-icon="Search" @input="onSearch" />
    </div>
    <div class="top-actions">
      <el-button type="primary" size="small" @click="addNewNote" :loading="loading">新建笔记</el-button>
    </div>

    <!-- 日历与标签 -->
    <div class="calendar-block">
      <DatePicker v-model="calendarDate" is-expanded color="teal" @dayclick="onPickDate($event.date)"/>
      <div class="calendar-actions">
        <el-button text size="small" @click="clearDate">清除日期筛选</el-button>
      </div>
    </div>

    

    <div class="note-list-container">
      <div 
        v-for="(note, index) in filteredNotes" 
        :key="note.id"
        class="note-item"
        :class="{ active: activeIndex === index }"
        @click="selectNote(note, index)"
      >
        <span class="note-title">{{ note.title || '未命名笔记' }}</span>
      </div>
    </div>
  </div>
  
</template>

<script setup>
import { ref, defineEmits,onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { DatePicker } from 'v-calendar'
import { getNoteList } from '../../api/note'
import request from '../../api/request'
import { Search } from '@element-plus/icons-vue'
 

const noteList = ref([])
const filteredNotes = computed(() => {
  let list = [...(noteList.value || [])]
  if (pickedDate.value) {
    const day = fmtDate(pickedDate.value)
    list = list.filter(n => fmtDate(n.updated_at) === day)
  }
  // 关键字搜索优先
  const q = keyword.value.trim()
  if (q) {
    list = list.filter(n => (n.title||'').includes(q) || (n.content||'').includes(q))
  }
  return list
})


const keyword = ref('')
const activeIndex = ref(0)
const loading = ref(false)
const emit = defineEmits(['select-note', 'add-note'])
const route = useRoute()
 

// 加载笔记列表
const loadNotes = async () => {
  try {
    const res = await getNoteList(1, 100) // 获取前100条记录
    noteList.value = res.data.data.list || []
    // 主页希望停留在仓库，不自动跳转到某篇笔记
    activeIndex.value = -1
  } catch (error) {
    console.error('加载笔记失败:', error)
  }
}

const onSearch = async () => {
  const q = keyword.value.trim()
  if (!q) { await loadNotes(); return }
  try {
    const res = await request.get('/note/search', { params: { q } })
    noteList.value = res.data.data.list || []
    activeIndex.value = 0
  } catch (e) { console.error(e) }
}

const selectNote = (note, index) => {
  activeIndex.value = index
  if (note && note.updated_at) {
    try { pickedDate.value = new Date(note.updated_at) } catch {}
    try { calendarDate.value = new Date(note.updated_at) } catch {}
  }
  emit('select-note', note)
}

const addNewNote = () => {
  // 不再直接创建笔记，而是通知父组件打开一个新的空白编辑器
  const newNote = { 
    id: null, 
    title: '未命名笔记', 
    content: '' 
  }
  emit('add-note', newNote)
}

// 回收站入口已搬到左侧图标栏

onMounted(() => {
  loadNotes()
  const refresh = () => loadNotes()
  window.addEventListener('note-updated', refresh)
  window.addEventListener('note-created', refresh)
  window.addEventListener('note-deleted', refresh)
  watch(() => route.query.id, (id) => {
    if (!id) return
    const list = noteList.value || []
    const idx = list.findIndex(n => String(n.id) === String(id))
    if (idx >= 0) {
      selectNote(list[idx], idx)
    }
  }, { immediate: true })
  onUnmounted(() => {
    window.removeEventListener('note-updated', refresh)
    window.removeEventListener('note-created', refresh)
    window.removeEventListener('note-deleted', refresh)
  })
})

// Calendar & tags helpers
const calendarDate = ref(new Date())
const pickedDate = ref(null)
const isToday = (d) => fmtDate(d) === fmtDate(new Date())
const isPicked = (d) => pickedDate.value && fmtDate(d) === fmtDate(pickedDate.value)
const onPickDate = (d) => { pickedDate.value = new Date(d) }
const clearDate = () => { pickedDate.value = null }
function fmtDate(dt) {
  const date = new Date(dt)
  const y = date.getFullYear(); const m = String(date.getMonth()+1).padStart(2,'0'); const da = String(date.getDate()).padStart(2,'0')
  return `${y}-${m}-${da}`
}



</script>

<style scoped>
/* 1. 根容器用 flex 布局，垂直排列，占满 100% 高度 */
.aside-note-list {
  height: 100%;
  width: 100%;
  padding-top: 10px;
  display: flex; /* 关键：flex 布局 */
  flex-direction: column; /* 垂直排列（按钮 + 列表） */
}

.search-row { width: 92%; margin: 6px auto 8px; }
.search-row :deep(.el-input__wrapper) { border-radius: 12px; background: #f5f6f7; box-shadow: none; border: 1px solid #e5e7eb; }
.search-row :deep(.el-input__inner) { height: 34px; }
.top-actions { width: 90%; margin: 6px auto 10px; display: flex; gap: 8px; }
.rag-actions { width: 90%; margin: 0 auto 10px; display: flex; gap: 8px; }
.note-title { font-family: system-ui, -apple-system, 'Segoe UI', Roboto, 'Noto Sans SC', Helvetica, Arial, sans-serif; }

/* 2. 列表容器：flex:1 占满剩余高度，超出滚动 */
.note-list-container {
  width: 90%;
  margin: 0 auto;
  flex: 1; /* 核心：占满 Aside 剩余高度 */
  overflow-y: auto; /* 笔记多的时候滚动，不超出页面 */
}

/* 原有笔记项样式保留 */
.note-item {
  padding: 12px 15px;
  border: 1px solid #ebeef5;
  border-bottom: none;
  cursor: pointer;
  transition: background-color 0.2s;
}

.note-item:first-child {
  border-radius: 4px 4px 0 0;
}

.note-item:last-child {
  border-bottom: 1px solid #ebeef5;
  border-radius: 0 0 4px 4px;
}

.note-item:hover {
  background-color: #f5f7fa;
}

.note-item.active {
  background-color: #e8f4ff;
  border-color: #b3d9ff;
}

.note-title {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

/* Calendar style mimic memos */
.calendar-block { width: 90%; margin: 0 auto 8px; }
.calendar-actions { display:flex; justify-content:flex-end; }
.date-cell { display:flex; justify-content:center; align-items:center; height: 28px; }
.date-cell span { width: 24px; height: 24px; display:flex; align-items:center; justify-content:center; border-radius: 50%; }
.date-cell span.today { border: 1px solid #d1d5db; }
.date-cell span.picked { background:#9f1b1b; color:#fff; }
.calendar-block :deep(.el-calendar) { border: none; }
.calendar-block :deep(.el-calendar__body) { padding: 0 8px 8px; }
.calendar-block :deep(.el-calendar__header) { padding: 8px 8px; border-bottom: none; }

 
</style>
