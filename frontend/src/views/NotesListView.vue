<template>
  <div class="notes-list">
    <div class="toolbar">
      <el-button type="primary" size="small" @click="createBlank">新建笔记</el-button>
    </div>
    <div class="cards">
      <el-card v-for="n in notes" :key="n.id" class="note-card" @click="open(n)">
        <div class="title">{{ n.title || '未命名笔记' }}</div>
        <div class="snippet">{{ snippet(n.content) }}</div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getNoteList, createNote } from '../api/note'
import { ElMessage } from 'element-plus'

const router = useRouter()
const notes = ref([])

const load = async () => {
  try {
    const res = await getNoteList(1, 100)
    notes.value = res.data?.data?.list || []
    
  } catch (e) {}
}
onMounted(load)

const snippet = (s) => {
  const txt = String(s||'').replace(/```[\s\S]*?```/g,'').replace(/[#*>\-`]/g,'')
  return txt.trim().slice(0,120)
}
const open = (n) => {
  router.push({ name: 'NoteEditor', query: { id: n.id } })
}
 
const createBlank = async () => {
  try {
    const title = '未命名笔记'
    const content = '# 未命名笔记\n\n'
    const res = await createNote({ title, content })
    const note = res.data?.data
    if (note?.id) router.push({ name: 'NoteEditor', query: { id: note.id } })
  } catch (e) { ElMessage.error('创建失败') }
}
</script>

<style scoped>
.notes-list { height: 100%; padding: 12px; display: flex; flex-direction: column; }
.toolbar { display:flex; justify-content:flex-end; margin-bottom: 10px; }
.cards { display:grid; grid-template-columns: repeat(3, 1fr); gap: 12px; flex: 1; overflow: auto; min-height: 0; align-content: start; grid-auto-rows: minmax(120px, auto); padding-bottom: 8px; }
.note-card { padding: 8px; }
.note-card { cursor:pointer; }
.title { font-weight: 600; margin-bottom: 6px; }
.snippet { color:#6b7280; font-size: 13px; }
@media (max-width: 1200px) { .cards { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 800px) { .cards { grid-template-columns: 1fr; } }
</style>
