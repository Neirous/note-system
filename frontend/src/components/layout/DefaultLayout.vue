<template>
  <div class="note-layout">
    <el-container style="height: 100%;">
      <!-- 左侧图标栏 -->
      <el-aside width="64px" style="background-color:#fff;border-right:1px solid #e6e6e6;">
        <IconBar />
      </el-aside>

      <!-- 左侧笔记侧栏（含搜索/日历/标签） -->
      <el-aside width="320px" style="background-color: #f5f7fa; border-right: 1px solid #e6e6e6;">
        <AsideNoteList 
          @select-note="handleSelectNote"
          @add-note="handleAddNoteFromAside"
        />
      </el-aside>

      <el-container style="height: 100%;">
        <el-main style="padding: 0; height: 100%;">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import AsideNoteList from '../aside/AsideNoteList.vue'
// HeaderTabBar 移除，主体直接显示视图
import IconBar from '../aside/IconBar.vue'
import { useRouter } from 'vue-router'
import { ref } from 'vue'

const currentNote = ref(null)
const router = useRouter()

const handleSelectNote = (note) => {
  currentNote.value = note
  // 如果是已有笔记，包含ID
  if (note.id) {
    router.push({ name: 'NoteEditor', query: { id: note.id } })
  } else {
    router.push({ name: 'NoteEditor' })
  }
}

const handleSwitchNote = (note) => {
  currentNote.value = note
  if (note.content !== undefined) {
    if (note.id) {
      router.push({ name: 'NoteEditor', query: { id: note.id } })
    } else {
      router.push({ name: 'NoteEditor' })
    }
  }
}

const handleAddNoteFromAside = (newNote) => {
  currentNote.value = newNote
  router.push({ name: 'NoteEditor' })
}

const handleAddNoteFromHeader = (newNote) => {
  currentNote.value = newNote
  router.push({ name: 'NoteEditor' })
}

const handleCloseNote = (note) => {
  console.log('关闭笔记：', note)
}
</script>

<style scoped>
/* 1. 确保布局根容器占满 100% 高度 */
.note-layout {
  height: 100%;
  width: 100%;
}

/* 2. 让 el-container/el-aside/el-main 强制继承 100% 高度 */
.note-layout :deep(.el-container) {
  --el-container-padding: 0;
  height: 100% !important; /* 强制占满父容器 */
}

.note-layout :deep(.el-aside) {
  height: 100% !important;
}

/* 3. 右侧容器（Header+Main）的高度继承 */
.note-layout :deep(.el-container > .el-container) {
  height: 100% !important;
}

/* 4. Main 区域去掉默认内边距，占满剩余高度 */
.note-layout :deep(.el-main) {
  padding: 0 !important;
  height: calc(100% - 60px) !important; /* 减去 Header 的 60px */
}

/* 5. Header 高度固定 60px */
.note-layout :deep(.el-header) {
  height: 60px !important;
  padding: 0 16px;
  border-bottom: 1px solid #e6e6e6;
}

/* 原有占位符样式保留 */
.aside-placeholder, .header-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
}
</style>
