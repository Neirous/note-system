<template>
  <div class="note-editor-container" style="height: 100%;">
    <mavon-editor
      v-model="noteContent"
      style="height: 100%;"
      placeholder="请输入笔记内容（支持 Markdown 语法）"
      @save="saveNote"
    />
  </div>
</template>

<script setup>
import { mavonEditor } from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import { ref, onMounted, watch } from 'vue'  
import { useRoute } from 'vue-router'
import { getNoteById, updateNote } from '../api/note'

const noteContent = ref('')
const route = useRoute()
const currentNoteId = ref(null)

onMounted(() => {
  updateNoteContent()
})

watch(
  () => route.query.content,  // 监听 content 参数变化
  () => {
    updateNoteContent()
  }
)

const updateNoteContent = async () => {
  // 从URL参数中尝试提取ID
  const urlParams = new URLSearchParams(window.location.search);
  const id = urlParams.get('id');
  const content = urlParams.get('content');
  
  if (id) {
    // 已存在的笔记
    currentNote.value = {
      id: parseInt(id),
      content: decodeURIComponent(content || '')
    }
    noteContent.value = currentNote.value.content
  } else if (content) {
    // 新建但已有内容的笔记（来自其他组件传递）
    noteContent.value = decodeURIComponent(content || '')
    currentNote.value = {
      content: noteContent.value
    }
  } else {
    // 完全新建的笔记
    noteContent.value = ''
    currentNote.value = null
  }
}

const saveNote = async (value, render) => {
  try {
    if (currentNote.value && currentNote.value.id) {
      // 更新已存在的笔记
      await updateNote(currentNote.value.id, {
        title: generateTitleFromContent(value),
        content: value
      })
      alert('保存成功')
    } else {
      // 创建新笔记
      const res = await createNote({
        title: generateTitleFromContent(value),
        content: value
      })
      
      // 保存成功后更新当前笔记信息
      const newNote = res.data.data
      currentNote.value = {
        id: newNote.id,
        content: newNote.content
      }
      
      // 更新URL以包含新笔记ID
      const newUrl = `${window.location.pathname}?id=${newNote.id}&content=${encodeURIComponent(newNote.content)}`
      window.history.replaceState({}, '', newUrl)
      
      alert('保存成功')
    }
  } catch (error) {
    console.error('保存失败:', error)
    alert('保存失败: ' + (error.response?.data?.message || error.message))
  }
}

// 从内容中提取标题的方法
const generateTitleFromContent = (content) => {
  // 查找第一个标题行作为标题
  const lines = content.split('\n')
  for (let line of lines) {
    if (line.startsWith('# ')) {
      return line.substring(2).trim() || '未命名笔记'
    }
  }
  // 如果没有找到标题，则使用内容的前20个字符
  return content.trim().substring(0, 20) || '未命名笔记'
}
</script>

<style scoped>
.note-editor-container :deep(.v-md-editor) {
  height: 100%;
}
</style>