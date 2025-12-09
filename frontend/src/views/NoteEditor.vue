<template>
  <div class="note-editor-container">
    <div class="editor-toolbar">
      <template v-if="isEditing">
      </template>
      <template v-else>
        <el-button type="primary" size="small" @click="startEdit">编辑</el-button>
        <el-button size="small" @click="onRename">重命名</el-button>
        <el-popconfirm title="确认删除该笔记？" confirm-button-text="删除" cancel-button-text="取消" @confirm="onDelete">
          <template #reference>
            <el-button type="danger" size="small">删除</el-button>
          </template>
        </el-popconfirm>
      </template>
      <el-tag v-if="isEditing && !titleValid" type="danger" effect="plain">第一行作为标题，使用 #</el-tag>
      <el-tag v-if="isEditing && !contentValid" type="warning" effect="plain">内容不能为空</el-tag>
    </div>
    <mavon-editor
      v-if="isEditing"
      v-model="noteContent"
      class="editor"
      :subfield="true"
      :toolbarsFlag="true"
      :editable="true"
      :placeholder="editorPlaceholder"
      @save="onSave"
    />
    <mavon-editor
      v-else
      v-model="noteContent"
      class="editor"
      :subfield="false"
      :toolbarsFlag="false"
      :editable="false"
      defaultOpen="preview"
    />
  </div>
  
</template>

<script setup>
import { mavonEditor } from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getNoteById, updateNote, createNote, deleteNote } from '../api/note'
import { ElMessageBox, ElMessage } from 'element-plus'

const route = useRoute()
const noteContent = ref('')
const currentNote = ref(null)
const isEditing = ref(true)
const editorPlaceholder = '第一行作为标题，示例：# 我的标题\n\n下面书写正文内容'
const TEMPLATE = '# 在此输入标题\n\n在此输入内容…'

const loadNote = async () => {
  const id = route.query.id
  if (id) {
    const res = await getNoteById(id)
    const payload = res.data && res.data.data ? res.data.data : res.data
    currentNote.value = payload || { id }
    noteContent.value = (currentNote.value && currentNote.value.content) || ''
    isEditing.value = false
  } else {
    currentNote.value = null
    noteContent.value = TEMPLATE
    isEditing.value = true
  }
}

onMounted(loadNote)
// 文件夹功能已移除
watch(() => route.fullPath, loadNote)

const titleValid = computed(() => {
  const firstLine = (noteContent.value || '').split('\n')[0].trim()
  return firstLine.startsWith('# ') && firstLine.length > 2
})
const contentValid = computed(() => {
  const lines = (noteContent.value || '').split('\n')
  const body = lines.slice(1).join('\n').trim()
  return body.length > 0
})

const onSave = async () => {
  let value = noteContent.value
  try {
    if (!contentValid.value) { ElMessage.error('内容不能为空'); return }
    if (!titleValid.value) {
      value = ensureTitle(value)
      noteContent.value = value
    }
    if (currentNote.value && currentNote.value.id) {
      const title = generateTitleFromContent(value)
      await updateNote(currentNote.value.id, { title, content: value })
      isEditing.value = false
      window.dispatchEvent(new CustomEvent('note-updated', { detail: { id: currentNote.value.id, title } }))
      ElMessage.success('保存成功')
    } else {
      const title = generateTitleFromContent(value)
      const res = await createNote({ title, content: value })
      const newNote = res.data.data
      currentNote.value = newNote
      history.replaceState({}, '', `${window.location.pathname}?id=${newNote.id}`)
      isEditing.value = false
      window.dispatchEvent(new CustomEvent('note-created', { detail: { id: newNote.id, title } }))
      ElMessage.success('保存成功')
    }
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

const startEdit = () => { isEditing.value = true }

const onRename = async () => {
  try {
    const { value } = await ElMessageBox.prompt('输入新的标题', '重命名', { inputValue: currentNote.value?.title || '' })
    if (!currentNote.value || !currentNote.value.id) return
    const newTitle = value || '未命名笔记'
    await updateNote(currentNote.value.id, { title: newTitle, content: noteContent.value })
    currentNote.value.title = newTitle
    window.dispatchEvent(new CustomEvent('note-updated', { detail: { id: currentNote.value.id, title: newTitle } }))
    ElMessage.success('重命名成功')
  } catch {}
}

const onDelete = async () => {
  if (!currentNote.value || !currentNote.value.id) return
  try {
    await deleteNote(currentNote.value.id)
    window.dispatchEvent(new CustomEvent('note-deleted', { detail: { id: currentNote.value.id } }))
    ElMessage.success('删除成功')
    currentNote.value = null
    noteContent.value = ''
    isEditing.value = true
    history.replaceState({}, '', `${window.location.pathname}`)
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

const generateTitleFromContent = (content) => {
  const lines = content.split('\n')
  for (let line of lines) {
    if (line.startsWith('# ')) return line.substring(2).trim() || '未命名笔记'
  }
  return content.trim().substring(0, 20) || '未命名笔记'
}

const ensureTitle = (content) => {
  const lines = content.split('\n')
  const first = (lines[0] || '').trim()
  if (!first) {
    lines[0] = '# 未命名笔记'
  } else if (first.startsWith('#') && !first.startsWith('# ')) {
    lines[0] = '# ' + first.slice(1)
  } else if (!first.startsWith('# ')) {
    lines[0] = '# ' + first
  }
  return lines.join('\n')
}
</script>

<style scoped>
.note-editor-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.editor-toolbar {
  height: 42px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  border-bottom: 1px solid #e6e6e6;
  gap: 8px;
}
.editor {
  height: calc(100% - 42px);
}

:deep(.v-md-editor) { border: none; }
</style>
 
