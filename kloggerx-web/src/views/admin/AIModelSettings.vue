<template>
  <div class="ai-model-settings">
    <div class="page-header">
      <h2>AI模型设置</h2>
    </div>

    <!-- Provider Cards -->
    <div class="provider-section">
      <div class="section-header">
        <span>模型提供商</span>
        <el-button type="primary" @click="showAddProviderDialog = true">
          <el-icon><Plus /></el-icon>添加提供商
        </el-button>
      </div>
      <div class="provider-cards">
        <div v-for="p in providers" :key="p.id" class="provider-card">
          <div class="provider-header">
            <div class="provider-icon" :style="{ background: p.color }">
              {{ p.name.charAt(0) }}
            </div>
            <div class="provider-info">
              <span class="provider-name">{{ p.name }}</span>
              <span class="provider-status" :class="{ active: p.isActive }">
                {{ p.isActive ? '已启用' : '未启用' }}
              </span>
            </div>
            <el-switch v-model="p.isActive" @change="toggleProvider(p)" />
          </div>
          <div class="provider-details">
            <div class="detail-item">
              <label>Base URL</label>
              <span>{{ p.baseUrl }}</span>
            </div>
            <div class="detail-item">
              <label>模型数量</label>
              <span>{{ p.models.length }} 个</span>
            </div>
          </div>
          <div class="provider-actions">
            <el-button link type="primary" @click="editProvider(p)">编辑</el-button>
            <el-button link type="primary" @click="detectModels(p)" :loading="p.detecting">探测模型</el-button>
            <el-button link type="danger" @click="deleteProvider(p)">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Model Classification -->
    <div class="model-section">
      <div class="section-header">
        <span>模型分类</span>
      </div>
      <el-tabs v-model="activeProvider" type="card">
        <el-tab-pane v-for="p in activeProviders" :key="p.id" :label="p.name" :name="String(p.id)">
          <div class="model-tree">
            <div v-for="vendor in getModelVendors(p.id)" :key="vendor.name" class="vendor-group">
              <div class="vendor-header" @click="toggleVendor(p.id, vendor.name)">
                <el-icon class="expand-arrow" :class="{ expanded: vendor.expanded }"><ArrowRight /></el-icon>
                <span class="vendor-name">{{ vendor.name }}</span>
                <span class="vendor-count">{{ vendor.categories.length }} 类</span>
              </div>
              <div v-show="vendor.expanded" class="vendor-body">
                <div v-for="cat in vendor.categories" :key="cat.name" class="category-group">
                  <div class="category-header" @click="toggleCategory(p.id, vendor.name, cat.name)">
                    <el-icon class="expand-arrow" :class="{ expanded: cat.expanded }"><ArrowRight /></el-icon>
                    <span class="category-name">{{ cat.name }}</span>
                    <span class="category-count">{{ cat.models.length }} 个模型</span>
                  </div>
                  <div v-show="cat.expanded" class="category-body">
                    <div v-for="m in cat.models" :key="m.id" class="model-item">
                      <div class="model-info">
                        <span class="model-name">{{ m.name }}</span>
                        <el-tag v-if="m.isDefault" size="small" type="success">默认</el-tag>
                      </div>
                      <div class="model-actions">
                        <el-button link size="small" @click="testModel(m)">测试</el-button>
                        <el-button link size="small" type="primary" @click="setDefaultModel(m)" v-if="!m.isDefault">设为默认</el-button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Knowledge Base AI Settings -->
    <div class="kb-settings-section">
      <div class="section-header">
        <span>知识库AI设置</span>
      </div>
      <el-form :model="kbSettings" label-width="140px" class="settings-form">
        <el-form-item label="嵌入模型">
          <el-select v-model="kbSettings.embeddingModel" style="width: 300px">
            <el-option-group v-for="p in activeProviders" :key="p.id" :label="p.name">
              <el-option v-for="m in getEmbeddingModels(p.id)" :key="m.id" :label="m.name" :value="m.id" />
            </el-option-group>
          </el-select>
          <div class="form-tip">用于将文档内容转换为向量，支持语义搜索</div>
        </el-form-item>
        <el-form-item label="对话模型">
          <el-select v-model="kbSettings.chatModel" style="width: 300px">
            <el-option-group v-for="p in activeProviders" :key="p.id" :label="p.name">
              <el-option v-for="m in getChatModels(p.id)" :key="m.id" :label="m.name" :value="m.id" />
            </el-option-group>
          </el-select>
          <div class="form-tip">用于知识库问答对话</div>
        </el-form-item>
        <el-form-item label="向量维度">
          <el-input-number v-model="kbSettings.vectorDimension" :min="256" :max="4096" :step="256" />
          <div class="form-tip">嵌入向量的维度，需与模型匹配</div>
        </el-form-item>
        <el-form-item label="相似度阈值">
          <el-slider v-model="kbSettings.similarityThreshold" :min="0" :max="1" :step="0.05" show-input style="width: 300px" />
          <div class="form-tip">检索结果的最低相似度要求</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveKbSettings" :loading="savingKb">保存设置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- Add/Edit Provider Dialog -->
    <el-dialog v-model="showAddProviderDialog" :title="editingProvider ? '编辑提供商' : '添加提供商'" width="500px" destroy-on-close>
      <el-form :model="providerForm" :rules="providerRules" ref="providerFormRef" label-width="100px">
        <el-form-item label="提供商名称" prop="name">
          <el-input v-model="providerForm.name" placeholder="如: OpenAI, Anthropic, 阿里云" />
        </el-form-item>
        <el-form-item label="Base URL" prop="baseUrl">
          <el-input v-model="providerForm.baseUrl" placeholder="https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="API Key" prop="apiKey">
          <el-input v-model="providerForm.apiKey" type="password" placeholder="请输入API Key" show-password />
        </el-form-item>
        <el-form-item label="显示颜色">
          <el-color-picker v-model="providerForm.color" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddProviderDialog = false">取消</el-button>
        <el-button type="primary" @click="submitProvider" :loading="submittingProvider">确定</el-button>
      </template>
    </el-dialog>

    <!-- Test Model Dialog -->
    <el-dialog v-model="showTestDialog" title="测试模型" width="500px" destroy-on-close>
      <div class="test-result">
        <div class="test-model-name">{{ testingModel?.name }}</div>
        <el-input
          v-model="testPrompt"
          type="textarea"
          :rows="3"
          placeholder="输入测试提示词"
        />
        <el-button type="primary" @click="runTest" :loading="testing" style="margin-top: 12px">运行测试</el-button>
        <div v-if="testResponse" class="test-response">
          <div class="response-label">响应:</div>
          <div class="response-content">{{ testResponse }}</div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getAIModelSettings, saveAIModelSettings, detectAIModels, testAIModel } from '@/api/modules/admin'
import { useNotify } from '@/composables/useNotify'

const $notify = useNotify()

interface Model {
  id: string
  name: string
  vendor: string
  category: string
  type: 'chat' | 'embedding' | 'image'
  isDefault: boolean
}

interface Provider {
  id: number
  name: string
  baseUrl: string
  apiKey: string
  color: string
  isActive: boolean
  models: Model[]
  detecting: boolean
}

interface VendorGroup {
  name: string
  expanded: boolean
  categories: CategoryGroup[]
}

interface CategoryGroup {
  name: string
  expanded: boolean
  models: Model[]
}

const providers = ref<Provider[]>([])
const activeProvider = ref('')
const showAddProviderDialog = ref(false)
const editingProvider = ref<Provider | null>(null)
const providerFormRef = ref<FormInstance>()
const submittingProvider = ref(false)
const providerForm = ref({
  name: '',
  baseUrl: '',
  apiKey: '',
  color: '#3370ff',
})

// 用于存储展开状态的响应式数据
const vendorExpandedMap = ref<Record<string, boolean>>({})
const categoryExpandedMap = ref<Record<string, boolean>>({})

const providerRules: FormRules = {
  name: [{ required: true, message: '请输入提供商名称', trigger: 'blur' }],
  baseUrl: [{ required: true, message: '请输入Base URL', trigger: 'blur' }],
  apiKey: [{ required: true, message: '请输入API Key', trigger: 'blur' }],
}

const activeProviders = computed(() => providers.value.filter(p => p.isActive))

function getVendorKey(providerId: number, vendorName: string) {
  return `${providerId}-${vendorName}`
}

function getCategoryKey(providerId: number, vendorName: string, categoryName: string) {
  return `${providerId}-${vendorName}-${categoryName}`
}

function isVendorExpanded(providerId: number, vendorName: string) {
  return vendorExpandedMap.value[getVendorKey(providerId, vendorName)] ?? false
}

function toggleVendor(providerId: number, vendorName: string) {
  const key = getVendorKey(providerId, vendorName)
  vendorExpandedMap.value[key] = !isVendorExpanded(providerId, vendorName)
}

function isCategoryExpanded(providerId: number, vendorName: string, categoryName: string) {
  return categoryExpandedMap.value[getCategoryKey(providerId, vendorName, categoryName)] ?? false
}

function toggleCategory(providerId: number, vendorName: string, categoryName: string) {
  const key = getCategoryKey(providerId, vendorName, categoryName)
  categoryExpandedMap.value[key] = !isCategoryExpanded(providerId, vendorName, categoryName)
}

function getModelVendors(providerId: number): VendorGroup[] {
  const provider = providers.value.find(p => p.id === providerId)
  if (!provider) return []

  const vendorMap = new Map<string, CategoryGroup[]>()
  provider.models.forEach(m => {
    if (!vendorMap.has(m.vendor)) {
      vendorMap.set(m.vendor, [])
    }
    const categories = vendorMap.get(m.vendor)!
    let cat = categories.find(c => c.name === m.category)
    if (!cat) {
      cat = { name: m.category, expanded: false, models: [] }
      categories.push(cat)
    }
    cat.models.push(m)
  })

  return Array.from(vendorMap.entries()).map(([name, categories]) => ({
    name,
    expanded: isVendorExpanded(providerId, name),
    categories: categories.map(cat => ({
      ...cat,
      expanded: isCategoryExpanded(providerId, name, cat.name),
    })),
  }))
}

function getEmbeddingModels(providerId: number): Model[] {
  const provider = providers.value.find(p => p.id === providerId)
  return provider?.models.filter(m => m.type === 'embedding') || []
}

function getChatModels(providerId: number): Model[] {
  const provider = providers.value.find(p => p.id === providerId)
  return provider?.models.filter(m => m.type === 'chat') || []
}

function editProvider(p: Provider) {
  editingProvider.value = p
  providerForm.value = {
    name: p.name,
    baseUrl: p.baseUrl,
    apiKey: p.apiKey,
    color: p.color,
  }
  showAddProviderDialog.value = true
}

async function submitProvider() {
  const valid = await providerFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!editingProvider.value && providers.value.length >= 15) {
    $notify.warning('最多支持添加15个模型提供商')
    return
  }
  submittingProvider.value = true
  try {
    if (editingProvider.value) {
      Object.assign(editingProvider.value, providerForm.value)
      $notify.success('提供商已更新')
    } else {
      const newProvider: Provider = {
        id: Date.now(),
        ...providerForm.value,
        isActive: true,
        models: [],
        detecting: false,
      }
      providers.value.push(newProvider)
      $notify.success('提供商已添加')
    }
    showAddProviderDialog.value = false
    editingProvider.value = null
    await saveAllProviders()
  } finally {
    submittingProvider.value = false
  }
}

async function saveAllProviders() {
  try {
    await saveAIModelSettings({
      providers: providers.value.map(p => ({
        id: p.id,
        name: p.name,
        baseUrl: p.baseUrl,
        apiKey: p.apiKey,
        color: p.color,
        isActive: p.isActive,
        models: p.models,
      })),
      kbSettings: kbSettings.value,
    })
  } catch {
    // silent save
  }
}

async function toggleProvider(p: Provider) {
  $notify.success(p.isActive ? '已启用' : '已禁用')
  await saveAllProviders()
}

async function deleteProvider(p: Provider) {
  await ElMessageBox.confirm(`确定要删除提供商 "${p.name}" 吗？`, '删除确认')
  providers.value = providers.value.filter(x => x.id !== p.id)
  $notify.success('已删除')
  await saveAllProviders()
}

async function detectModels(p: Provider) {
  if (!p.baseUrl || !p.apiKey) {
    $notify.warning('请先配置 Base URL 和 API Key')
    return
  }
  p.detecting = true
  try {
    const res: any = await detectAIModels({ baseUrl: p.baseUrl, apiKey: p.apiKey })
    if (res.data?.models) {
      p.models = res.data.models.map((m: any) => ({
        id: m.id,
        name: m.name || m.id,
        vendor: m.vendor || 'Other',
        category: m.category || '通用模型',
        type: m.type || 'chat',
        isDefault: m.isDefault || false,
      }))
      // Set first chat model as default if none
      if (!p.models.some((m: Model) => m.isDefault && m.type === 'chat')) {
        const firstChat = p.models.find((m: Model) => m.type === 'chat')
        if (firstChat) firstChat.isDefault = true
      }
      $notify.success(`探测到 ${p.models.length} 个模型`)
      await saveAllProviders()
    }
  } catch (e: any) {
    $notify.error(e.message || '探测失败')
  } finally {
    p.detecting = false
  }
}

function setDefaultModel(m: Model) {
  const provider = providers.value.find(p => p.models.includes(m))
  provider?.models.forEach(x => {
    if (x.type === m.type) x.isDefault = false
  })
  m.isDefault = true
  $notify.success(`已将 ${m.name} 设为默认`)
}

// Test model
const showTestDialog = ref(false)
const testingModel = ref<Model | null>(null)
const testPrompt = ref('你好，请介绍一下自己')
const testResponse = ref('')
const testing = ref(false)

function testModel(m: Model) {
  testingModel.value = m
  testPrompt.value = '你好，请介绍一下自己'
  testResponse.value = ''
  showTestDialog.value = true
}

async function runTest() {
  if (!testingModel.value) return
  const provider = providers.value.find(p => p.models.includes(testingModel.value!))
  if (!provider || !provider.baseUrl || !provider.apiKey) {
    $notify.warning('请先配置提供商的 Base URL 和 API Key')
    return
  }
  testing.value = true
  try {
    const res: any = await testAIModel({
      baseUrl: provider.baseUrl,
      apiKey: provider.apiKey,
      model: testingModel.value.id,
      prompt: testPrompt.value,
    })
    if (res.data?.response) {
      testResponse.value = res.data.response
    } else {
      testResponse.value = '模型返回空响应'
    }
  } catch (e: any) {
    testResponse.value = `测试失败: ${e.message || '未知错误'}`
  } finally {
    testing.value = false
  }
}

// Knowledge Base Settings
const kbSettings = ref({
  embeddingModel: '',
  chatModel: '',
  vectorDimension: 1536,
  similarityThreshold: 0.7,
})
const savingKb = ref(false)

async function saveKbSettings() {
  savingKb.value = true
  try {
    await saveAIModelSettings({
      providers: providers.value,
      kbSettings: kbSettings.value,
    })
    $notify.success('知识库AI设置已保存')
  } catch (e: any) {
    $notify.error(e.message || '保存失败')
  } finally {
    savingKb.value = false
  }
}

async function fetchProviders() {
  try {
    const res: any = await getAIModelSettings()
    if (res.data?.providers) {
      providers.value = res.data.providers
    }
    if (res.data?.kbSettings) {
      kbSettings.value = { ...kbSettings.value, ...res.data.kbSettings }
    }
    if (providers.value.length && providers.value[0].isActive) {
      activeProvider.value = String(providers.value[0].id)
    }
  } catch (e) {
    console.error('Failed to fetch AI settings:', e)
  }
}

onMounted(() => {
  fetchProviders()
})
</script>

<style scoped>
.ai-model-settings {
  padding: 24px;
}
.page-header {
  margin-bottom: 24px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--kx-text-primary);
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-weight: 500;
  font-size: 16px;
}
.provider-section {
  margin-bottom: 32px;
}
.provider-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}
.provider-card {
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 16px;
  background: #fff;
}
.provider-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.provider-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 18px;
}
.provider-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.provider-name {
  font-weight: 500;
}
.provider-status {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.provider-status.active {
  color: #36b37e;
}
.provider-details {
  border-top: 1px solid var(--kx-border);
  border-bottom: 1px solid var(--kx-border);
  padding: 12px 0;
  margin-bottom: 12px;
}
.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 4px;
}
.detail-item label {
  color: var(--kx-text-secondary);
}
.provider-actions {
  display: flex;
  gap: 8px;
}
.model-section {
  margin-bottom: 32px;
}
.model-tree {
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 12px;
}
.vendor-group {
  margin-bottom: 4px;
}
.vendor-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
}
.vendor-header:hover {
  background: rgba(0,0,0,0.04);
}
.vendor-name {
  flex: 1;
  font-weight: 500;
}
.vendor-count {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.vendor-body {
  padding-left: 24px;
}
.category-group {
  margin-bottom: 4px;
}
.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  cursor: pointer;
  border-radius: 6px;
}
.category-header:hover {
  background: rgba(0,0,0,0.04);
}
.category-name {
  flex: 1;
}
.category-count {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.category-body {
  padding-left: 24px;
}
.expand-arrow {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  transition: transform 0.2s;
}
.expand-arrow.expanded {
  transform: rotate(90deg);
}
.model-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  border-radius: 6px;
}
.model-item:hover {
  background: rgba(0,0,0,0.04);
}
.model-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.model-name {
  font-size: 13px;
}
.kb-settings-section {
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 20px;
}
.settings-form {
  max-width: 600px;
}
.form-tip {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 4px;
}
.test-result {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.test-model-name {
  font-weight: 500;
  font-size: 16px;
}
.test-response {
  margin-top: 12px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 12px;
  background: #f7f8fa;
}
.response-label {
  font-weight: 500;
  margin-bottom: 8px;
}
.response-content {
  white-space: pre-wrap;
  font-size: 13px;
}
</style>
