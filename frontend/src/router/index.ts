import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/AboutView.vue')
    },
    {
      path: '/select-session',
      name: 'select-session',
      component: () => import('../views/SessionSelect.vue')
    },
    {
      path: '/select-source',
      name: 'select-source',
      component: () => import('../views/VideoSelect.vue')
    },
    {
      path: '/video/:id',
      name: 'video',
      component: () => import('../views/Video.vue')
    },
    {
      path: '/clip/:id',
      name: 'clip',
      component: () => import('../views/Clip.vue')
    }
  ]
})

export default router
