import { createRouter, createWebHistory } from 'vue-router'
import CreateNote from '../views/notes/CreateNote.vue'
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: '/notes/create', // 默认跳转到创建笔记页面
    },
    {
      path: '/notes/create',
      name: 'create-note',
      component: CreateNote
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
    },
  ],
})

export default router
