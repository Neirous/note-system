// src/api/request.js
import axios from 'axios'

const request = axios.create({
    baseURL: 'http://localhost:8090/api',
    timeout: 5000,
})

export default request
