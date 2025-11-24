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
  height: 36px;
  display: flex;
  align-items: center;
  overflow-x: auto;
}
.tab-list::-webkit-scrollbar { display: none; }
.tab-item {
  display: flex;
  align-items: center;
  height: 32px;
  padding: 0 10px;
  margin: 0 4px;
  background: linear-gradient(180deg,#fff,#f4f6f8);
  border: 1px solid #e3e7ee;
  border-radius: 8px;
  cursor: pointer;
  flex: 0 1 180px;
  min-width: 110px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.06);
}
.tab-item.active { background: linear-gradient(180deg,#eaf5ff,#e1f0ff); border: 1px solid #b3d9ff; }
.tab-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 140px; font-weight: 600; }
.close-icon { font-size: 12px; margin-left: 6px; color: #999; }
.close-icon:hover { color: #ff4d4f; }
</style>
