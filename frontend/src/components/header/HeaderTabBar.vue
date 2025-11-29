<template>
  <div class="header-bar">
    <div class="nav-btns">
      <el-icon class="ctrl-icon" @click="goBack" title="后退"><ArrowLeft /></el-icon>
      <el-icon class="ctrl-icon" @click="goForward" title="前进"><ArrowRight /></el-icon>
    </div>
    <div class="tab-list">
      <div 
        v-for="(note, index) in openedNotes" 
        :key="note.id || note.tempId || index" 
        class="tab-item"
        :class="{ active: isActiveNote(note) }"
        @click="switchTab(note)"
        :title="note.title || '未命名笔记'"
      >
        <span class="tab-title">{{ note.title || '未命名笔记' }}</span>
        <el-icon class="close-icon" @click.stop="closeTab(index, note)"><Close /></el-icon>
      </div>
    </div>
    <div class="add-btn" @click="addNewNote" title="新建笔记">
      <el-icon><Plus /></el-icon>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, defineEmits, defineProps, onMounted, onUnmounted } from 'vue'
import { ArrowLeft, ArrowRight, Plus, Close } from '@element-plus/icons-vue'

const props = defineProps({ currentNote: { type: Object, default: () => ({}) } })
const emit = defineEmits(['switch-note', 'close-note', 'add-note'])

const openedNotes = ref([])
const activeNote = ref({})

watch(
  () => props.currentNote,
  (newNote) => {
    if (!newNote || (!newNote.id && !newNote.tempId)) return
    const idx = openedNotes.value.findIndex(n => (n.id && n.id === newNote.id) || (n.tempId && n.tempId === newNote.tempId))
    if (idx === -1) openedNotes.value.push(newNote)
    activeNote.value = newNote
  },
  { immediate: true }
)

const isActiveNote = (note) => {
  if (activeNote.value?.id && note.id) return activeNote.value.id === note.id
  if (activeNote.value?.tempId && note.tempId) return activeNote.value.tempId === note.tempId
  return false
}

const switchTab = (note) => {
  activeNote.value = note
  emit('switch-note', note)
}

const closeTab = (index, note) => {
  openedNotes.value.splice(index, 1)
  if (isActiveNote(note)) {
    const newActive = openedNotes.value[index - 1] || openedNotes.value[0] || {}
    activeNote.value = newActive
    emit('switch-note', newActive)
  }
  emit('close-note', note)
}

const goBack = () => history.back()
const goForward = () => history.forward()

const addNewNote = () => {
  const newNote = { tempId: 'temp_' + Date.now(), title: '未命名笔记', content: '' }
  emit('add-note', newNote)
}

const handleNoteUpdated = (e) => {
  const { id, title } = e.detail || {}
  if (!id) return
  const idx = openedNotes.value.findIndex(n => n.id === id)
  if (idx > -1) openedNotes.value[idx].title = title || openedNotes.value[idx].title
}
const handleNoteDeleted = (e) => {
  const { id } = e.detail || {}
  if (!id) return
  const idx = openedNotes.value.findIndex(n => n.id === id)
  if (idx > -1) closeTab(idx, openedNotes.value[idx])
}
onMounted(() => {
  window.addEventListener('note-updated', handleNoteUpdated)
  window.addEventListener('note-deleted', handleNoteDeleted)
})
onUnmounted(() => {
  window.removeEventListener('note-updated', handleNoteUpdated)
  window.removeEventListener('note-deleted', handleNoteDeleted)
})
</script>

<style scoped>
.header-bar {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
}
.nav-btns { display: flex; gap: 8px; }
.ctrl-icon, .add-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  cursor: pointer;
}
.ctrl-icon:hover, .add-btn:hover { color: #0080ff; }

.tab-list {
  flex: 1;
  height: 40px;
  display: flex;
  align-items: flex-end;
  overflow-x: auto;
  border-bottom: 1px solid #d9d9d9;
  padding: 0 6px;
}
.tab-list::-webkit-scrollbar { display: none; }
.tab-item {
  display: flex;
  align-items: center;
  height: 30px;
  padding: 0 12px;
  margin: 0 6px;
  background: linear-gradient(180deg,#fdfdfd,#f1f3f5);
  border: 1px solid #cdd6e4;
  border-bottom: none;
  border-top-left-radius: 10px;
  border-top-right-radius: 10px;
  cursor: pointer;
  flex: 0 1 200px;
  min-width: 120px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
  margin-bottom: -1px;
}
.tab-item.active { background: #ffffff; border-color: #9fb7d6; box-shadow: 0 2px 4px rgba(0,0,0,0.08); }
.tab-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 150px; font-weight: 600; }
.close-icon { font-size: 12px; margin-left: 8px; color: #909399; }
.close-icon:hover { color: #ff4d4f; }
</style>
