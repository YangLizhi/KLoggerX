<template>
  <div class="not-found">
    <h1>404</h1>
    <p>页面未找到</p>
    <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
  </div>
</template>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 80vh;
}
.not-found h1 {
  font-size: 72px;
  color: var(--kx-text-placeholder);
  margin-bottom: 8px;
}
.not-found p {
  font-size: 16px;
  color: var(--kx-text-secondary);
  margin-bottom: 24px;
}
</style>
