// src/utils/http.js
// HTTP 请求工具封装

import { API_CONFIG } from '@/config/api'

/**
 * 通用请求函数
 * @param {string} url - 请求地址
 * @param {object} options - 请求选项
 * @returns {Promise} 返回Promise对象
 */
async function request(url, options = {}) {
    // 默认请求选项
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers
        }
    }

    // 合并请求选项
    const config = {
        ...defaultOptions,
        ...options
    }

    try {
        // 发起fetch请求
        const response = await fetch(url, config)

        // 检查响应状态
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        // 解析JSON响应
        const data = await response.json()
        return data
    } catch (error) {
        // 处理错误
        console.error('Request failed:', error)
        throw error
    }
}

/**
 * GET 请求封装
 * @param {string} url - 请求地址
 * @param {object} params - 查询参数
 * @returns {Promise} 返回Promise对象
 */
export async function get(url, params = {}) {
    // 构造带查询参数的URL
    const queryString = new URLSearchParams(params).toString()
    const fullUrl = queryString ? `${url}?${queryString}` : url

    return request(fullUrl, {
        method: 'GET'
    })
}

/**
 * POST 请求封装
 * @param {string} url - 请求地址
 * @param {object} data - 请求数据
 * @returns {Promise} 返回Promise对象
 */
export async function post(url, data = {}) {
    return request(url, {
        method: 'POST',
        body: JSON.stringify(data)
    })
}

/**
 * PUT 请求封装
 * @param {string} url - 请求地址
 * @param {object} data - 请求数据
 * @returns {Promise} 返回Promise对象
 */
export async function put(url, data = {}) {
    return request(url, {
        method: 'PUT',
        body: JSON.stringify(data)
    })
}

/**
 * DELETE 请求封装
 * @param {string} url - 请求地址
 * @returns {Promise} 返回Promise对象
 */
export async function del(url) {
    return request(url, {
        method: 'DELETE'
    })
}

// 导出所有方法
export default {
    get,
    post,
    put,
    delete: del
}