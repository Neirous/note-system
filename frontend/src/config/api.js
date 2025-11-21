// src/config/api.js
// API 配置文件

// 获取基础URL，如果是生产环境则使用生产环境URL，否则使用开发环境URL
const BASE_URL = import.meta.env.PROD
    ? 'http://your-production-domain.com/api'  // 生产环境URL（部署时替换为实际地址）
    : 'http://localhost:8080/api'              // 开发环境URL（对应后端服务）

// 导出API配置对象
export const API_CONFIG = {
    BASE_URL,
    ENDPOINTS: {
        NOTES: `${BASE_URL}/note`  // 笔记相关API的基础路径
    }
}

export default API_CONFIG