import request from './request'

// 获取笔记列表
export function getNoteList(page = 1, size = 10) {
    return request.get('/note/list', {
        params: { page, size }
    })
}

// 创建新笔记
export function createNote(data) {
    // 确保传递的数据符合后端要求
    const requestData = {
        title: data.title || '未命名笔记',
        content: data.content || ''
    }
    return request.post('/note', requestData)
}

// 获取指定笔记详情
export function getNoteById(id) {
    return request.get(`/note/${id}`)
}

// 更新笔记
export function updateNote(id, data) {
    return request.put(`/note/${id}`, data)
}

// 删除笔记
export function deleteNote(id) {
    return request.delete(`/note/${id}`)
}