<!-- src/views/notes/CreateNote.vue -->
<template>
  <div class="create-note">
    <h1>创建新笔记</h1>
    <div class="form-group">
      <label for="title">标题:</label>
      <input 
        type="text" 
        id="title" 
        v-model="note.title" 
        placeholder="请输入笔记标题"
        class="form-control"
      />
    </div>
    
    <div class="form-group">
      <label for="content">内容 (Markdown):</label>
      <!-- 使用 mavon-editor 组件 -->
      <mavon-editor 
        v-model="note.content" 
        ref="md"
        @imgAdd="$imgAdd"
        class="markdown-editor"
      />
    </div>
    
    <div class="form-actions">
      <button @click="saveNote" class="btn btn-primary" :disabled="isSaving">
        {{ isSaving ? '保存中...' : '保存笔记' }}
      </button>
      <button @click="cancel" class="btn btn-secondary">取消</button>
    </div>
  </div>
</template>

<script>
// 引入笔记服务
import noteService from '@/services/noteService'

export default {
  name: 'CreateNote',
  data() {
    return {
      // 笔记数据模型
      note: {
        title: '',
        content: ''
      },
      // 保存状态标志
      isSaving: false
    }
  },
  methods: {
    /**
     * 保存笔记方法
     * 调用后端API创建新笔记
     */
    async saveNote() {
      // 基本验证
      if (!this.note.title.trim()) {
        alert('请输入笔记标题')
        return
      }
      
      if (!this.note.content.trim()) {
        alert('请输入笔记内容')
        return
      }
      
      try {
        // 设置保存状态为true，防止重复点击
        this.isSaving = true
        
        // 调用笔记服务创建笔记
        const result = await noteService.createNote(this.note)
        
        // 显示成功消息
        alert('笔记创建成功!')
        console.log('创建笔记成功:', result)
        
        // 清空表单
        this.note = {
          title: '',
          content: ''
        }
      } catch (error) {
        // 处理错误
        console.error('创建笔记失败:', error)
        alert('创建笔记失败: ' + (error.message || '未知错误'))
      } finally {
        // 恢复保存状态
        this.isSaving = false
      }
    },
    
    /**
     * 取消操作，返回上一页
     */
    cancel() {
      this.$router.back()
    },
    
    /**
     * 图片上传处理方法（mavon-editor 的回调）
     * @param {number} pos - 图片位置
     * @param {File} $file - 图片文件
     */
    $imgAdd(pos, $file) {
      // TODO: 实现图片上传逻辑
      console.log('图片上传:', pos, $file)
    }
  }
}
</script>

<style scoped>
.create-note {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-control {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

.markdown-editor {
  min-height: 400px;
}

.form-actions {
  display: flex;
  gap: 10px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background-color: #007bff;
  color: white;
}

.btn-primary:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
}
</style>