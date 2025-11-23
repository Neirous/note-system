// src/api/request.js
import axios from 'axios'

const request = axios.create({
    baseURL: 'http://localhost:8080/api', // 后端服务地址
    timeout: 5000,
})

export default request