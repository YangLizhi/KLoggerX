<template>
  <div class="comment-bubbles" v-if="bubbles.length > 0">
    <div
      v-for="bubble in bubbles"
      :key="bubble.commentId"
      class="comment-bubble"
      :class="{ resolved: bubble.resolved }"
      :style="{ top: bubble.top + 'px' }"
      :title="bubble.resolved ? '已解决评论' : `点击查看评论`"
      @click.stop="handleClick(bubble)"
    >
      <svg viewBox="0 0 24 24" fill="currentColor" class="bubble-icon">
        <path d="M20 2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h14l4 4V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
      </svg>
      <span v-if="bubble.count > 1" class="bubble-count">{{ bubble.count }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import type { Editor } from '@tiptap/vue-3'

interface BubbleInfo {
  commentId: string
  top: number
  count: number
  resolved: boolean
}

import { getComments } from '@/api/modules/collaborate'

interface CommentItem {
  id: number
  resolved: boolean
}

const props = defineProps<{
  editor?: Editor | null
  documentId: number
  editorContainer?: HTMLElement | null
}>()

const emit = defineEmits<{
  (e: 'click-comment', commentId: string): void
}>()

const bubbles = ref<BubbleInfo[]>([])
const comments = ref<CommentItem[]>([])

// 获取评论列表
async function fetchComments() {
  if (!props.documentId) return
  try {
    const res: any = await getComments(props.documentId)
    comments.value = res.data || []
  } catch {
    // handled
  }
}

// 防抖函数
function debounce<T extends (...args: any[]) => any>(fn: T, delay: number): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout> | null = null
  return (...args: Parameters<T>) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}

// 处理点击事件
function handleClick(bubble: BubbleInfo) {
  emit('click-comment', bubble.commentId)
}

// 更新气泡位置（带滚动偏移修正）
function updateBubblePositions() {
  if (!props.editor || !props.editorContainer) {
    bubbles.value = []
    return
  }

  const { view, state } = props.editor
  const { doc } = state
  const editorRect = props.editorContainer.getBoundingClientRect()

  // 用于合并相近位置的评论标记
  const positionMap = new Map<string, { top: number; count: number; resolved: boolean; commentId: string }>()

  // 遍历文档中的所有 commentMark
  doc.descendants((node, pos) => {
    if (!node.marks || node.marks.length === 0) return

    const commentMarks = node.marks.filter(m => m.type.name === 'commentMark')
    
    commentMarks.forEach(mark => {
      const commentId = mark.attrs.commentId as string
      if (!commentId) return

      // 检查评论是否已解决
      const comment = comments.value.find(c => String(c.id) === commentId)
      const isResolved = comment?.resolved || mark.attrs.resolved === true

      try {
        // 使用 coordsAtPos 来处理滚动后的坐标
        const coords = view.coordsAtPos(pos)
        
        // 计算相对于编辑器容器的 top 值（考虑滚动）
        const relativeTop = coords.top - editorRect.top

        // 使用行位置作为合并的键（相近的合并为一个气泡）
        const lineKey = String(Math.round(relativeTop / 24))

        const existing = positionMap.get(lineKey)
        if (!existing || existing.top > relativeTop) {
          positionMap.set(lineKey, {
            top: relativeTop,
            count: existing ? existing.count + 1 : 1,
            resolved: existing ? (existing.resolved && isResolved) : isResolved,
            commentId
          })
        } else if (existing) {
          existing.count++
          if (!isResolved) {
            existing.resolved = false
          }
        }
      } catch {
        // 忽略无效位置
      }
    })
  })

  // 转换为数组并排序
  const result: BubbleInfo[] = []
  positionMap.forEach((info) => {
    result.push({
      commentId: info.commentId,
      top: info.top,
      count: info.count,
      resolved: info.resolved
    })
  })

  result.sort((a, b) => a.top - b.top)
  
  bubbles.value = result
}

// 防抖更新
const debouncedUpdate = debounce(updateBubblePositions, 100)

// 处理滚动事件
function handleScroll() {
  debouncedUpdate()
}

// 处理窗口 resize
function handleResize() {
  debouncedUpdate()
}

// 监听编辑器变化
watch(() => props.editor, (newEditor, oldEditor) => {
  if (oldEditor) {
    oldEditor.off('update', debouncedUpdate)
    oldEditor.off('transaction', debouncedUpdate)
  }
  if (newEditor) {
    newEditor.on('update', debouncedUpdate)
    newEditor.on('transaction', debouncedUpdate)
    nextTick(() => {
      updateBubblePositions()
    })
  }
}, { immediate: true })

// 监听评论列表变化
watch(() => comments.value, () => {
  nextTick(() => {
    updateBubblePositions()
  })
}, { deep: true })

// 监听 documentId 变化，重新获取评论
watch(() => props.documentId, () => {
  fetchComments()
}, { immediate: true })

// 监听编辑器容器变化
watch(() => props.editorContainer, (newContainer, oldContainer) => {
  if (oldContainer) {
    oldContainer.removeEventListener('scroll', handleScroll)
  }
  if (newContainer) {
    newContainer.addEventListener('scroll', handleScroll)
    nextTick(() => {
      updateBubblePositions()
    })
  }
}, { immediate: true })

onMounted(() => {
  window.addEventListener('resize', handleResize)
  // 初始计算
  nextTick(() => {
    updateBubblePositions()
  })
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (props.editorContainer) {
    props.editorContainer.removeEventListener('scroll', handleScroll)
  }
  if (props.editor) {
    props.editor.off('update', debouncedUpdate)
    props.editor.off('transaction', debouncedUpdate)
  }
})
</script>

<style scoped>
.comment-bubbles {
  position: absolute;
  right: -32px;
  top: 0;
  width: 24px;
  z-index: 50;
  pointer-events: none;
}

.comment-bubble {
  position: absolute;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #fef3cd;
  border: 1px solid #f0c040;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
  pointer-events: auto;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.comment-bubble:hover {
  transform: scale(1.15);
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.15);
  background: #fce588;
}

.comment-bubble.resolved {
  background: #f0f0f0;
  border-color: #ccc;
}

.comment-bubble.resolved:hover {
  background: #e8e8e8;
}

.bubble-icon {
  width: 14px;
  height: 14px;
  color: #8b6914;
}

.comment-bubble.resolved .bubble-icon {
  color: #888;
}

.bubble-count {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  background: #ff6b6b;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.comment-bubble.resolved .bubble-count {
  background: #aaa;
}
</style>
