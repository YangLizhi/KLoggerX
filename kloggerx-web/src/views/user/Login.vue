<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-left">
        <div class="brand">
          <span class="brand-icon">K</span>
          <h1>KLoggerX</h1>
        </div>
        <p class="brand-desc">{{ $t('auth.brandDesc') }}</p>
        <p class="brand-sub">{{ $t('auth.brandSub') }}</p>
      </div>
      <div class="login-right">
        <div class="login-form-wrap">
          <h2>{{ $t('auth.login') }}</h2>
          <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="handleLogin">
            <el-form-item :label="$t('auth.email')" prop="email">
              <el-input v-model="form.email" :placeholder="$t('auth.emailPlaceholder')" prefix-icon="Message" size="large" />
            </el-form-item>
            <el-form-item :label="$t('auth.password')" prop="password">
              <el-input v-model="form.password" type="password" :placeholder="$t('auth.passwordPlaceholder')" prefix-icon="Lock" show-password size="large" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="handleLogin">{{ $t('auth.loginBtn') }}</el-button>
            </el-form-item>
          </el-form>
          <div class="login-footer">
            <span>{{ $t('auth.noAccount') }}</span>
            <router-link to="/register">{{ $t('auth.registerNow') }}</router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/store/modules/user'
import type { FormInstance } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({ email: '', password: '' })
const rules = {
  email: [
    { required: true, message: t('auth.emailRequired'), trigger: 'blur' },
    { type: 'email' as const, message: t('auth.emailInvalid'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: t('auth.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('auth.passwordMin'), trigger: 'blur' },
  ],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.login(form)
    const redirect = (route.query.redirect as string) || '/documents'
    router.push(redirect)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-container {
  display: flex;
  width: 800px;
  min-height: 480px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}
.login-left {
  width: 320px;
  background: linear-gradient(135deg, #3370ff 0%, #245bdb 100%);
  color: #fff;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 48px 32px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.brand-icon {
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
}
.brand h1 {
  font-size: 24px;
  font-weight: 600;
}
.brand-desc {
  font-size: 18px;
  margin-bottom: 8px;
  opacity: 0.95;
}
.brand-sub {
  font-size: 14px;
  opacity: 0.7;
}
.login-right {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 40px;
}
.login-form-wrap {
  width: 100%;
  max-width: 340px;
}
.login-form-wrap h2 {
  font-size: 22px;
  margin-bottom: 28px;
  color: var(--kx-text-primary);
}
.login-btn {
  width: 100%;
}
.login-footer {
  text-align: center;
  margin-top: 16px;
  font-size: 13px;
  color: var(--kx-text-secondary);
}
.login-footer a {
  color: var(--kx-primary);
  margin-left: 4px;
}
</style>
