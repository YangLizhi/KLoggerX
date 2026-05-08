<template>
  <el-dialog
    :model-value="visible"
    :title="$t('knowledge.feedback.title')"
    width="480px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
  >
    <el-form label-position="top">
      <el-form-item :label="$t('knowledge.feedback.category')">
        <el-radio-group v-model="form.feedbackType">
          <el-radio value="inaccurate">{{ $t('knowledge.feedback.inaccurate') }}</el-radio>
          <el-radio value="incomplete">{{ $t('knowledge.feedback.incomplete') }}</el-radio>
          <el-radio value="irrelevant">{{ $t('knowledge.feedback.irrelevant') }}</el-radio>
          <el-radio value="outdated">{{ $t('knowledge.feedback.outdated') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item :label="$t('knowledge.feedback.commentLabel')">
        <el-input
          v-model="form.comment"
          type="textarea"
          :rows="3"
          :placeholder="$t('knowledge.feedback.commentPlaceholder')"
        />
      </el-form-item>

      <el-form-item :label="$t('knowledge.feedback.correctAnswerLabel')">
        <el-input
          v-model="form.correctAnswer"
          type="textarea"
          :rows="3"
          :placeholder="$t('knowledge.feedback.correctAnswerPlaceholder')"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('knowledge.feedback.submit') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  visible: boolean
  messageId: number
}>()

const emit = defineEmits<{
  'update:visible': [val: boolean]
  submit: [data: { feedbackType: string; comment: string; correctAnswer: string }]
}>()

const submitting = ref(false)
const { t } = useI18n()

const form = reactive({
  feedbackType: 'inaccurate',
  comment: '',
  correctAnswer: ''
})

watch(() => props.visible, (val) => {
  if (val) {
    form.feedbackType = 'inaccurate'
    form.comment = ''
    form.correctAnswer = ''
  }
})

function handleSubmit() {
  submitting.value = true
  emit('submit', { ...form })
  // Parent will close the dialog after API call
  setTimeout(() => { submitting.value = false }, 500)
}
</script>
