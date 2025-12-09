import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'v-calendar/style.css'

// 新增：导入路由实例（类比 Go 导入路由配置）
import router from './router'

const app = createApp(App)
app.use(ElementPlus)
app.use(router) // 新增：挂载路由（类比 Go 的 app.Use(router)）

app.mount('#app')
