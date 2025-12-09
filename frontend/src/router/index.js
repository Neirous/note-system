// 1. 从 vue-router 库中导入核心函数（类比 Go 导入 gin 的核心结构体）
import { createRouter, createWebHistory } from 'vue-router'

// 2. 导入全局布局组件（后续会创建，先占位，类比 Go 导入自定义 Handler）
import DefaultLayout from '../components/layout/DefaultLayout.vue'

// 3. 定义路由规则（类比 Go 的 router.GET("/", handler)）
const routes = [
    {
        path: '/', // 根路径
        component: DefaultLayout, // 根路径对应全局布局组件
        children: [ // 嵌套路由：子路径渲染到布局的 <router-view /> 中
            {
                path: '', // 默认子路径（访问 / 时渲染）
                redirect: 'notes'
            }
            , {
                path: 'note',
                name: 'NoteEditor',
                component: () => import('../views/NoteEditor.vue')
            }
            , {
                path: 'trash',
                name: 'TrashView',
                component: () => import('../views/TrashView.vue')
            }
            , {
                path: 'rag-qa',
                name: 'RagQAView',
                component: () => import('../views/RagQAView.vue')
            }
            , {
                path: 'notes',
                name: 'NotesListView',
                component: () => import('../views/NotesListView.vue')
            }
        ]
    }
]

// 4. 创建路由实例（类比 Go 的 gin.Default() 创建路由实例）
const router = createRouter({
    history: createWebHistory(), // 路由模式：HTML5 历史模式（无 # 号）
    routes: routes // 绑定路由规则（可简写为 routes）
})

// 5. 导出路由实例（类比 Go 的导出函数/结构体，供 main.js 使用）
export default router
