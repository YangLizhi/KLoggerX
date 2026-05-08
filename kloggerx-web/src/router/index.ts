import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import i18n from '@/locales'

/**
 * 从 JWT token 中解析用户角色
 * JWT payload 是 base64 编码的 JSON
 */
function getUserRoleFromToken(token: string | null): string {
  if (!token) return ''
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return ''
    // base64url decode
    const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const decoded = JSON.parse(atob(payload))
    return decoded.role || ''
  } catch {
    return ''
  }
}

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
          meta: { titleKey: 'nav.home' },
        },
        {
          path: 'documents',
          name: 'Documents',
          component: () => import('@/views/document/DocumentLibrary.vue'),
          meta: { titleKey: 'nav.cloudDrive' },
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
          meta: { titleKey: 'nav.templateLibrary' },
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/user/Settings.vue'),
        },
        {
          path: 'admin',
          redirect: '/admin/dashboard',
          meta: { requiresAdmin: true, titleKey: 'nav.admin' },
          children: [
            {
              path: 'dashboard',
              name: 'AdminDashboard',
              component: () => import('@/views/admin/AdminDashboard.vue'),
              meta: { titleKey: 'nav.dashboard' },
            },
            {
              path: 'users',
              name: 'UserManagement',
              component: () => import('@/views/admin/UserManagement.vue'),
              meta: { titleKey: 'nav.usersAndPermissions' },
            },
            {
              path: 'departments',
              name: 'DepartmentManagement',
              component: () => import('@/views/admin/DepartmentManagement.vue'),
              meta: { titleKey: 'nav.departmentManagement' },
            },
            {
              path: 'ai-models',
              name: 'AIModelSettings',
              component: () => import('@/views/admin/AIModelSettings.vue'),
              meta: { titleKey: 'nav.aiModelSettings' },
            },
            {
              path: 'storage',
              name: 'StorageSettings',
              component: () => import('@/views/admin/StorageSettings.vue'),
              meta: { titleKey: 'nav.cloudStorage' },
            },
            {
              path: 'templates',
              name: 'TemplateManagement',
              component: () => import('@/views/admin/TemplateManagement.vue'),
              meta: { titleKey: 'nav.templateManagement' },
            },
            {
              path: 'feedback-review',
              name: 'FeedbackReview',
              component: () => import('@/views/admin/FeedbackReview.vue'),
              meta: { titleKey: 'nav.feedbackReview' },
            },
            {
              path: 'operation-logs',
              name: 'OperationLog',
              component: () => import('@/views/admin/OperationLog.vue'),
              meta: { titleKey: 'nav.operationLog' },
            },
          ],
        },
      ],
    },
    {
      path: '/share/:token',
      name: 'SharedConversation',
      component: () => import('@/views/knowledge/SharedConversation.vue'),
      meta: { requiresAuth: false },
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
  } else if (to.meta.requiresAdmin) {
    // 解析token中的role，若非admin则拒绝
    const userRole = getUserRoleFromToken(token)
    if (userRole !== 'admin') {
      next({ path: '/home' })
      ElMessage.warning(i18n.global.t('nav.noAdminPermission'))
    } else {
      next()
    }
  } else {
    next()
  }
})

router.afterEach((to) => {
  const titleKey = to.meta.titleKey as string | undefined
  if (titleKey) {
    document.title = `${i18n.global.t(titleKey)} - KLoggerX`
  } else {
    document.title = 'KLoggerX'
  }
})

export default router
