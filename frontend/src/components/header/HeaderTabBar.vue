<template>
  <div class="header-tab-bar">
    <div class="tab-list" style="display: flex; align-items: center; height: 100%; overflow-x: auto;">
      <div 
        v-for="(note, index) in openedNotes" 
        :key="note.id || index"  
        class="tab-item"
        :class="{ active: isActiveNote(note) }"
        @click="switchTab(note)"
      >
        <span class="tab-title">{{ note.title || '未命名笔记' }}</span>
        <el-icon 
          class="close-icon"
          @click.stop="closeTab(index, note)"  
        >
          <Close />
        </el-icon>
      </div>

      <div class="add-tab-btn" @click="addNewTab" title="新增笔记">
        <el-icon><Plus /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, defineEmits, watch } from 'vue'
import { Close, Plus } from '@element-plus/icons-vue'

const props = defineProps({
  currentNote: {
    type: Object,
    required: true,
    default: () => ({})
  }
})

const emit = defineEmits(['switch-note', 'close-note', 'add-note'])
const openedNotes = ref([])
const activeNote = ref({})

watch(
  () => props.currentNote,
  (newNote) => {
    if (newNote) {
      // 检查是否已经打开
      const existingIndex = openedNotes.value.findIndex(note => 
        (note.id && note.id === newNote.id) || 
        (note.tempId && note.tempId === newNote.tempId)
      )
      
      if (existingIndex === -1) {
        // 新增到标签列表
        openedNotes.value.push(newNote)
      }
      
      // 设置为活动标签
      activeNote.value = newNote
    }
  },
  { immediate: true }
)

const isActiveNote = (note) => {
  if (activeNote.value.id && note.id) {
    return activeNote.value.id === note.id
  }
  if (activeNote.value.tempId && note.tempId) {
    return activeNote.value.tempId === note.tempId
  }
  return false
}

const switchTab = (note) => {
  activeNote.value = note
  emit('switch-note', note)
}

const closeTab = (index, note) => {
  openedNotes.value.splice(index, 1)
  
  // 如果关闭的是当前活动标签
  if (isActiveNote(note)) {
    const newActiveNote = openedNotes.value[index - 1] || openedNotes.value[0] || {}
    activeNote.value = newActiveNote
    emit('switch-note', newActiveNote)
  }
  
  emit('close-note', note)
}

const addNewTab = () => {
  // 创建临时笔记对象（不立即保存到后端）
  const tempId = 'temp_' + Date.now()
  const newNote = {
    tempId: tempId,
    title: '未命名笔记',
    content: ''
  }
  
  emit('add-note', newNote)
}
</script>

<style scoped>
/* 标签栏容器 */
.header-tab-bar {
  width: 100%;
  height: 100%;
}

/* 标签项样式 */
.tab-item {
  display: flex;
  align-items: center;
  height: 36px;
  line-height: 36px;
  padding: 0 12px;
  margin: 0 4px;
  background-color: #f5f7fa;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
}

/* 激活标签样式 */
.tab-item.active {
  background-color: #e8f4ff;
  border: 1px solid #b3d9ff;
}

/* 标签标题 */
.tab-title {
  max-width: 150px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
}

/* 关闭按钮 */
.close-icon {
  font-size: 12px;
  color: #999;
  cursor: pointer;
  transition: color 0.2s;
}

.close-icon:hover {
  color: #ff4d4f;
}

.add-tab-btn {
  margin-left: 8px;
  height: 36px;
  width: 36px;  /* 宽高一致，配合 circle 变成圆形 */
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
}
  .add-tab-btn:hover {
    background-color: #f5f7fa;
    color: #0080ff;
  }
.tab-list::-webkit-scrollbar {
  display: none;
}
</style>