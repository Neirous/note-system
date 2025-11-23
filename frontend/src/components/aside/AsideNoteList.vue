<template>
  <div class="aside-note-list">
    <!-- 顶部添加笔记按钮 -->
    <el-button 
      type="primary" 
      size="small" 
      style="width: 90%; margin: 10px auto; display: block;"
      @click="addNewNote"
      :loading="loading"
    >
      新建笔记
    </el-button>

    <div class="note-list-container">
      <div 
        v-for="(note, index) in noteList" 
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
import { ref, defineEmits,onMounted } from 'vue'
import { getNoteList } from '../../api/note'

const noteList = ref([])
const activeIndex = ref(0)
const loading = ref(false)
const emit = defineEmits(['select-note', 'add-note'])

// 加载笔记列表
const loadNotes = async () => {
  try {
    const res = await getNoteList(1, 100) // 获取前100条记录
    noteList.value = res.data.data.list || []
    
    // 默认选中第一条
    if (noteList.value.length > 0) {
      selectNote(noteList.value[0], 0)
    }
  } catch (error) {
    console.error('加载笔记失败:', error)
  }
}

const selectNote = (note, index) => {
  activeIndex.value = index
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

onMounted(() => {
  loadNotes()
})
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
</style>