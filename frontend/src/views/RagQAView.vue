<template>
  <div class="rag-qa-container">
    <div class="qa-input">
      <el-input v-model="question" placeholder="输入你的问题" clearable />
      <el-button type="primary" @click="ask" :loading="loading">提问</el-button>
    </div>
    <el-card class="qa-answer">
      <div v-if="!answer && !loading" class="placeholder">在这里显示回答</div>
      <pre v-else class="answer-text">{{ answer }}</pre>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ragQA } from '../api/rag'

const question = ref('')
const answer = ref('')
const loading = ref(false)

const ask = async () => {
  const q = question.value.trim()
  if (!q) { answer.value = ''; return }
  try {
    loading.value = true
    answer.value = '正在生成…'
    const res = await ragQA(q, 180000)
    answer.value = (res.data?.data?.answer) || ''
  } catch (e) {
    answer.value = '请求失败或超时，请稍后重试。'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.rag-qa-container { padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.qa-input { display: flex; gap: 8px; }
.qa-answer { min-height: 300px; }
.placeholder { color: #999; }
.answer-text { white-space: pre-wrap; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
</style>
