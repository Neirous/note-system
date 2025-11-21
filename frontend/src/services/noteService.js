// src/services/noteService.js
// 笔记服务API封装

import http from '@/utils/http'
import { API_CONFIG } from '@/config/api'

/**
 * 笔记服务类
 * 封装所有与笔记相关的API调用
 */
class NoteService {
    /**
     * 构造函数
     */
    constructor() {
        // 设置笔记API的基础路径
        this.baseUrl = API_CONFIG.ENDPOINTS.NOTES
    }

    /**
     * 创建新笔记
     * @param {Object} noteData - 笔记数据
     * @param {string} noteData.title - 笔记标题
     * @param {string} noteData.content - 笔记内容
     * @returns {Promise} 返回创建的笔记对象
     */
    async createNote(noteData) {
        try {
            // 调用POST请求创建笔记
            const response = await http.post(this.baseUrl, noteData)
            return response
        } catch (error) {
            // 错误处理
            console.error('创建笔记失败:', error)
            throw error
        }
    }

    /**
     * 根据ID获取笔记
     * @param {number|string} id - 笔记ID
     * @returns {Promise} 返回指定ID的笔记对象
     */
    async getNoteById(id) {
        try {
            // 构造获取单个笔记的URL
            const url = `${this.baseUrl}/${id}`
            // 调用GET请求获取笔记
            const response = await http.get(url)
            return response
        } catch (error) {
            // 错误处理
            console.error('获取笔记失败:', error)
            throw error
        }
    }

    /**
     * 更新笔记
     * @param {number|string} id - 笔记ID
     * @param {Object} noteData - 更新的笔记数据
     * @returns {Promise} 返回更新后的笔记对象
     */
    async updateNote(id, noteData) {
        try {
            // 构造更新笔记的URL
            const url = `${this.baseUrl}/${id}`
            // 调用PUT请求更新笔记
            const response = await http.put(url, noteData)
            return response
        } catch (error) {
            // 错误处理
            console.error('更新笔记失败:', error)
            throw error
        }
    }

    /**
     * 删除笔记
     * @param {number|string} id - 笔记ID
     * @returns {Promise} 返回删除结果
     */
    async deleteNote(id) {
        try {
            // 构造删除笔记的URL
            const url = `${this.baseUrl}/${id}`
            // 调用DELETE请求删除笔记
            const response = await http.delete(url)
            return response
        } catch (error) {
            // 错误处理
            console.error('删除笔记失败:', error)
            throw error
        }
    }

    /**
     * 获取笔记列表（分页）
     * @param {Object} params - 查询参数
     * @param {number} params.page - 页码
     * @param {number} params.pageSize - 每页数量
     * @returns {Promise} 返回笔记列表和分页信息
     */
    async listNotes(params = {}) {
        try {
            // 调用GET请求获取笔记列表
            const response = await http.get(`${this.baseUrl}/list`, params)
            return response
        } catch (error) {
            // 错误处理
            console.error('获取笔记列表失败:', error)
            throw error
        }
    }
}

// 创建并导出笔记服务实例
export default new NoteService()