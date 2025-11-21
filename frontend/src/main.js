import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
// 引入 mavon-editor 和其样式
import mavonEditor from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'

const app = createApp(App)

app.use(router)
// 使用 mavon-editor 插件
app.use(mavonEditor)

app.mount('#app')
