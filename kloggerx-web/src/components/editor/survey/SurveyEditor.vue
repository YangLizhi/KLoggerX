<template>
  <div class="survey-editor">
    <SurveyToolbar
      :question-count="questions.length"
      :selected-index="selectedIndex"
      @action="handleToolbarAction"
    />
    <div class="survey-main">
      <div class="survey-sidebar">
        <div class="question-list">
          <div
            v-for="(q, idx) in questions"
            :key="q.id"
            class="question-item"
            :class="{ active: selectedIndex === idx }"
            @click="selectedIndex = idx"
          >
            <span class="q-num">{{ idx + 1 }}.</span>
            <span class="q-title">{{ q.title || '未命名题目' }}</span>
            <span class="q-type">{{ getTypeLabel(q.type) }}</span>
          </div>
        </div>
      </div>
      <div class="survey-content" v-if="currentQuestion">
        <div class="question-header">
          <input
            class="question-title"
            v-model="currentQuestion.title"
            placeholder="请输入题目"
            @input="scheduleSave"
          />
        </div>
        <div class="question-options">
          <template v-if="currentQuestion.type === 'single' || currentQuestion.type === 'multiple'">
            <div
              v-for="(opt, oi) in currentQuestion.options"
              :key="oi"
              class="option-item"
            >
              <el-radio v-if="currentQuestion.type === 'single'" :label="opt" disabled>{{ opt }}</el-radio>
              <el-checkbox v-else :label="opt" disabled>{{ opt }}</el-checkbox>
              <input
                class="option-input"
                :value="opt"
                @input="updateOption(oi, ($event.target as HTMLInputElement).value)"
              />
              <el-button
                size="small"
                text
                @click="removeOption(oi)"
                :disabled="currentQuestion.options.length <= 2"
              >
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
            <el-button size="small" text @click="addOption">
              <el-icon><Plus /></el-icon>添加选项
            </el-button>
          </template>
          <template v-else-if="currentQuestion.type === 'text'">
            <el-input
              type="textarea"
              :rows="3"
              placeholder="填空题作答区域"
              disabled
            />
          </template>
          <template v-else-if="currentQuestion.type === 'rating'">
            <div class="rating-preview">
              <el-rate v-model="currentQuestion.ratingMax" :max="currentQuestion.ratingMax || 5" disabled />
              <span class="rating-label">最高 {{ currentQuestion.ratingMax || 5 }} 分</span>
            </div>
            <div class="rating-setting">
              <span>最高分值：</span>
              <el-input-number v-model="currentQuestion.ratingMax" :min="3" :max="10" @change="scheduleSave" />
            </div>
          </template>
          <template v-else-if="currentQuestion.type === 'matrix'">
            <div class="matrix-preview">
              <table class="matrix-table">
                <thead>
                  <tr>
                    <th></th>
                    <th v-for="col in currentQuestion.matrixCols" :key="col">{{ col }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in currentQuestion.matrixRows" :key="row">
                    <td>{{ row }}</td>
                    <td v-for="col in currentQuestion.matrixCols" :key="col">
                      <el-radio :name="row" :label="col" disabled />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>
        </div>
        <div class="question-settings">
          <el-checkbox v-model="currentQuestion.required" @change="scheduleSave">必填</el-checkbox>
        </div>
      </div>
      <div class="survey-empty" v-else>
        <p>点击左侧题目列表或添加新题目</p>
      </div>
    </div>

    <!-- Preview Dialog -->
    <el-dialog v-model="showPreview" title="问卷预览" width="600px">
      <div class="survey-preview">
        <h2>{{ surveyTitle }}</h2>
        <div v-for="(q, idx) in questions" :key="q.id" class="preview-question">
          <p class="preview-q-title">{{ idx + 1 }}. {{ q.title }}{{ q.required ? ' *' : '' }}</p>
          <template v-if="q.type === 'single'">
            <el-radio-group v-model="previewAnswers[idx]">
              <el-radio v-for="opt in q.options" :key="opt" :label="opt">{{ opt }}</el-radio>
            </el-radio-group>
          </template>
          <template v-else-if="q.type === 'multiple'">
            <el-checkbox-group v-model="previewAnswers[idx]">
              <el-checkbox v-for="opt in q.options" :key="opt" :label="opt">{{ opt }}</el-checkbox>
            </el-checkbox-group>
          </template>
          <template v-else-if="q.type === 'text'">
            <el-input type="textarea" :rows="2" v-model="previewAnswers[idx]" />
          </template>
          <template v-else-if="q.type === 'rating'">
            <el-rate v-model="previewAnswers[idx]" :max="q.ratingMax || 5" />
          </template>
        </div>
      </div>
    </el-dialog>

    <!-- Settings Dialog -->
    <el-dialog v-model="showSettingsDialog" title="问卷设置" width="400px">
      <el-form label-width="80px">
        <el-form-item label="问卷标题">
          <el-input v-model="surveyTitle" @input="scheduleSave" />
        </el-form-item>
        <el-form-item label="问卷说明">
          <el-input type="textarea" :rows="3" v-model="surveyDescription" @input="scheduleSave" />
        </el-form-item>
        <el-form-item label="匿名填写">
          <el-switch v-model="anonymousMode" @change="scheduleSave" />
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import SurveyToolbar from './SurveyToolbar.vue'
import { ElMessage } from 'element-plus'

interface Question {
  id: string
  type: 'single' | 'multiple' | 'text' | 'rating' | 'matrix'
  title: string
  options: string[]
  required: boolean
  ratingMax?: number
  matrixRows?: string[]
  matrixCols?: string[]
}

interface SurveyData {
  title: string
  description: string
  anonymous: boolean
  questions: Question[]
}

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

const questions = ref<Question[]>([])
const selectedIndex = ref(-1)
const surveyTitle = ref('未命名问卷')
const surveyDescription = ref('')
const anonymousMode = ref(false)
const showPreview = ref(false)
const showSettingsDialog = ref(false)
const previewAnswers = ref<any[]>([])
let saveTimer: ReturnType<typeof setTimeout> | null = null
let idCounter = 0

function genId(): string {
  return `q_${Date.now()}_${idCounter++}`
}

const currentQuestion = computed(() => {
  if (selectedIndex.value >= 0 && selectedIndex.value < questions.value.length) {
    return questions.value[selectedIndex.value]
  }
  return null
})

function initData() {
  if (props.content) {
    try {
      const parsed: SurveyData = JSON.parse(props.content)
      if (parsed.questions && Array.isArray(parsed.questions)) {
        questions.value = parsed.questions
        surveyTitle.value = parsed.title || '未命名问卷'
        surveyDescription.value = parsed.description || ''
        anonymousMode.value = parsed.anonymous || false
        if (questions.value.length > 0) {
          selectedIndex.value = 0
        }
        return
      }
    } catch {}
  }
  // Default empty survey
  questions.value = []
  selectedIndex.value = -1
}

function getTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    single: '单选',
    multiple: '多选',
    text: '填空',
    rating: '评分',
    matrix: '矩阵',
  }
  return labels[type] || type
}

function handleToolbarAction(event: { action: string; params?: any }) {
  switch (event.action) {
    case 'addQuestion':
      addQuestion(event.params)
      break
    case 'moveUp':
      moveQuestion(-1)
      break
    case 'moveDown':
      moveQuestion(1)
      break
    case 'duplicateQuestion':
      duplicateQuestion()
      break
    case 'deleteQuestion':
      deleteQuestion()
      break
    case 'preview':
      showPreview.value = true
      previewAnswers.value = questions.value.map(q => q.type === 'multiple' ? [] : null)
      break
    case 'settings':
      showSettingsDialog.value = true
      break
  }
}

function addQuestion(type: string) {
  const newQ: Question = {
    id: genId(),
    type: type as any,
    title: '',
    options: type === 'single' || type === 'multiple' ? ['选项1', '选项2'] : [],
    required: false,
    ratingMax: type === 'rating' ? 5 : undefined,
    matrixRows: type === 'matrix' ? ['行1', '行2'] : undefined,
    matrixCols: type === 'matrix' ? ['列1', '列2', '列3'] : undefined,
  }
  questions.value.push(newQ)
  selectedIndex.value = questions.value.length - 1
  scheduleSave()
}

function moveQuestion(direction: number) {
  const newIndex = selectedIndex.value + direction
  if (newIndex < 0 || newIndex >= questions.value.length) return
  const temp = questions.value[selectedIndex.value]
  questions.value[selectedIndex.value] = questions.value[newIndex]
  questions.value[newIndex] = temp
  selectedIndex.value = newIndex
  scheduleSave()
}

function duplicateQuestion() {
  if (selectedIndex.value < 0) return
  const copy = { ...questions.value[selectedIndex.value], id: genId() }
  if (copy.options) {
    copy.options = [...copy.options]
  }
  questions.value.splice(selectedIndex.value + 1, 0, copy)
  selectedIndex.value += 1
  scheduleSave()
}

function deleteQuestion() {
  if (selectedIndex.value < 0) return
  questions.value.splice(selectedIndex.value, 1)
  if (selectedIndex.value >= questions.value.length) {
    selectedIndex.value = questions.value.length - 1
  }
  scheduleSave()
}

function addOption() {
  if (!currentQuestion.value || !currentQuestion.value.options) return
  currentQuestion.value.options.push(`选项${currentQuestion.value.options.length + 1}`)
  scheduleSave()
}

function updateOption(index: number, value: string) {
  if (!currentQuestion.value || !currentQuestion.value.options) return
  currentQuestion.value.options[index] = value
  scheduleSave()
}

function removeOption(index: number) {
  if (!currentQuestion.value || !currentQuestion.value.options) return
  if (currentQuestion.value.options.length <= 2) return
  currentQuestion.value.options.splice(index, 1)
  scheduleSave()
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    const data: SurveyData = {
      title: surveyTitle.value,
      description: surveyDescription.value,
      anonymous: anonymousMode.value,
      questions: questions.value,
    }
    emit('save', JSON.stringify(data))
  }, 2000)
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.survey-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.survey-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}
.survey-sidebar {
  width: 240px;
  background: #f5f6f7;
  border-right: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
}
.question-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}
.question-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 4px;
  gap: 8px;
}
.question-item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.question-item.active {
  background: rgba(51, 112, 255, 0.1);
}
.q-num {
  font-size: 12px;
  color: #646a73;
  flex-shrink: 0;
}
.q-title {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.q-type {
  font-size: 11px;
  color: #8f959e;
  background: #e8e9eb;
  padding: 2px 6px;
  border-radius: 4px;
}
.survey-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}
.question-header {
  margin-bottom: 16px;
}
.question-title {
  width: 100%;
  border: none;
  outline: none;
  font-size: 16px;
  font-weight: 500;
  padding: 8px 0;
  border-bottom: 1px solid transparent;
}
.question-title:focus {
  border-bottom-color: #3370ff;
}
.question-options {
  padding: 16px 0;
}
.option-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.option-input {
  border: none;
  outline: none;
  font-size: 13px;
  flex: 1;
  padding: 4px 0;
  border-bottom: 1px solid transparent;
}
.option-input:focus {
  border-bottom-color: #3370ff;
}
.rating-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.rating-label {
  font-size: 12px;
  color: #8f959e;
}
.rating-setting {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}
.matrix-preview {
  overflow-x: auto;
}
.matrix-table {
  width: 100%;
  border-collapse: collapse;
}
.matrix-table th, .matrix-table td {
  border: 1px solid #e5e6eb;
  padding: 8px 12px;
  text-align: center;
}
.matrix-table th {
  background: #f5f6f7;
  font-weight: 500;
}
.question-settings {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e6eb;
}
.survey-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8f959e;
}
.survey-preview h2 {
  margin-bottom: 24px;
  text-align: center;
}
.preview-question {
  margin-bottom: 24px;
}
.preview-q-title {
  font-weight: 500;
  margin-bottom: 12px;
}
</style>
