import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/user/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/user/Register.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/views/layout/MainLayout.vue'),
      redirect: '/home',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'home',
          name: 'Home',
          component: () => import('@/views/home/HomePage.vue'),
          meta: { title: '主页' },
        },
        {
          path: 'documents',
          name: 'Documents',
          component: () => import('@/views/document/DocumentLibrary.vue'),
          meta: { title: '云盘' },
        },
        {
          path: 'recent',
          name: 'Recent',
          component: () => import('@/views/document/RecentDocuments.vue'),
        },
        {
          path: 'favorites',
          name: 'Favorites',
          component: () => import('@/views/document/Favorites.vue'),
        },
        {
          path: 'recycle-bin',
          name: 'RecycleBin',
          component: () => import('@/views/document/RecycleBin.vue'),
        },
        {
          path: 'knowledge',
          name: 'KnowledgeHome',
          component: () => import('@/views/knowledge/KnowledgeHome.vue'),
        },
        {
          path: 'knowledge/list',
          name: 'KnowledgeList',
          component: () => import('@/views/knowledge/KnowledgeList.vue'),
        },
        {
          path: 'knowledge/:id',
          name: 'KnowledgeDetail',
          component: () => import('@/views/knowledge/KnowledgeDetail.vue'),
          props: true,
        },
        {
          path: 'templates',
          name: 'TemplateCenter',
          component: () => import('@/views/template/TemplateCenter.vue'),
          meta: { title: '模板库' },
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/user/Settings.vue'),
        },
        {
          path: 'admin',
          redirect: '/admin/users',
          meta: { requiresAdmin: true, title: '系统管理' },
          children: [
            {
              path: 'users',
              name: 'UserManagement',
              component: () => import('@/views/admin/UserManagement.vue'),
              meta: { title: '用户和权限' },
            },
            {
              path: 'departments',
              name: 'DepartmentManagement',
              component: () => import('@/views/admin/DepartmentManagement.vue'),
              meta: { title: '部门管理' },
            },
            {
              path: 'ai-models',
              name: 'AIModelSettings',
              component: () => import('@/views/admin/AIModelSettings.vue'),
              meta: { title: 'AI模型设置' },
            },
            {
              path: 'storage',
              name: 'StorageSettings',
              component: () => import('@/views/admin/StorageSettings.vue'),
              meta: { title: '云盘存储' },
            },
            {
              path: 'templates',
              name: 'TemplateManagement',
              component: () => import('@/views/admin/TemplateManagement.vue'),
              meta: { title: '模板管理' },
            },
          ],
        },
      ],
    },
    {
      path: '/doc/:id',
      name: 'DocEditor',
      component: () => import('@/views/document/DocEditor.vue'),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/common/NotFound.vue'),
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('kx_token')
  if (to.meta.requiresAuth !== false && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
