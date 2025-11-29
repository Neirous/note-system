import request from './request'

export function ragSearch(q, topK = 5) {
  return request.get('/rag/search', { params: { q, topK } })
}

export function ragQA(question, timeoutMs = 180000) {
  return request.post('/rag/qa', { question }, { timeout: timeoutMs })
}
