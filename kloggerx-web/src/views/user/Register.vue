<template>
  <div class="login-page">
    <div class="register-container">
      <div class="register-form-wrap">
        <h2>{{ $t('auth.registerKLoggerX') }}</h2>
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="handleRegister">
          <el-form-item :label="$t('auth.username')" prop="username">
            <el-input v-model="form.username" :placeholder="$t('auth.usernamePlaceholder')" prefix-icon="User" size="large" />
          </el-form-item>
          <el-form-item :label="$t('auth.email')" prop="email">
            <el-input v-model="form.email" :placeholder="$t('auth.emailPlaceholder')" prefix-icon="Message" size="large" />
          </el-form-item>
          <el-form-item :label="$t('auth.password')" prop="password">
            <el-input v-model="form.password" type="password" :placeholder="$t('auth.passwordPlaceholder')" prefix-icon="Lock" show-password size="large" />
          </el-form-item>
          <el-form-item :label="$t('auth.confirmPassword')" prop="confirmPassword">
            <el-input v-model="form.confirmPassword" type="password" :placeholder="$t('auth.confirmPasswordPlaceholder')" prefix-icon="Lock" show-password size="large" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" size="large" class="register-btn" :loading="loading" @click="handleRegister">{{ $t('auth.registerBtn') }}</el-button>
          </el-form-item>
        </el-form>
        <div class="login-footer">
          <span>{{ $t('auth.hasAccount') }}</span>
          <router-link to="/login">{{ $t('auth.goLogin') }}</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '@/api/modules/user'
import { ElMessage, type FormInstance } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({ username: '', email: '', password: '', confirmPassword: '' })
const rules = {
  username: [
    { required: true, message: t('auth.usernameRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: t('auth.usernameLength'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: t('auth.emailRequired'), trigger: 'blur' },
    { type: 'email' as const, message: t('auth.emailInvalid'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: t('auth.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('auth.passwordMin'), trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: t('auth.confirmPasswordRequired'), trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: (e?: Error) => void) => {
        if (value !== form.password) callback(new Error(t('auth.passwordMismatch')))
        else callback()
      },
      trigger: 'blur',
    },
  ],
}

async function handleRegister() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await register(form)
    ElMessage.success(t('auth.registerSuccess'))
    router.push('/login')
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
.register-container {
  width: 440px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  padding: 48px 40px;
}
.register-form-wrap h2 {
  font-size: 22px;
  margin-bottom: 28px;
  text-align: center;
  color: var(--kx-text-primary);
}
.register-btn {
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
