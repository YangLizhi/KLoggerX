import { Mark, mergeAttributes } from '@tiptap/core'

export interface CommentMarkOptions {
  HTMLAttributes: Record<string, any>
}

export interface CommentMarkAttributes {
  commentId: string
  color?: string
  resolved?: boolean
}

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    commentMark: {
      /**
       * 为选区添加评论标记
       */
      setCommentMark: (commentId: string, color?: string) => ReturnType
      /**
       * 移除指定评论标记
       */
      unsetCommentMark: (commentId: string) => ReturnType
      /**
       * 更新评论标记的解决状态
       */
      updateCommentMarkResolved: (commentId: string, resolved: boolean) => ReturnType
    }
  }
}

export const CommentMark = Mark.create<CommentMarkOptions>({
  name: 'commentMark',

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      commentId: {
        default: null,
        parseHTML: element => element.getAttribute('data-comment-id'),
        renderHTML: attributes => {
          if (!attributes.commentId) {
            return {}
          }
          return {
            'data-comment-id': attributes.commentId,
          }
        },
      },
      color: {
        default: '#fef3cd',
        parseHTML: element => element.getAttribute('data-comment-color'),
        renderHTML: attributes => {
          return {
            'data-comment-color': attributes.color || '#fef3cd',
          }
        },
      },
      resolved: {
        default: false,
        parseHTML: element => element.getAttribute('data-comment-resolved') === 'true',
        renderHTML: attributes => {
          return {
            'data-comment-resolved': String(attributes.resolved === true),
          }
        },
      },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'span[data-comment-id]',
      },
    ]
  },

  renderHTML({ HTMLAttributes }) {
    const color = HTMLAttributes['data-comment-color'] || '#fef3cd'
    const resolved = HTMLAttributes['data-comment-resolved'] === 'true'
    const resolvedClass = resolved ? ' resolved' : ''

    return [
      'span',
      mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, {
        class: `comment-mark${resolvedClass}`,
        style: `background-color: ${resolved ? '#f0f0f0' : color}; border-bottom-color: ${resolved ? '#ccc' : '#f0c040'}`,
      }),
      0,
    ]
  },

  addCommands() {
    return {
      setCommentMark:
        (commentId: string, color = '#fef3cd') =>
        ({ commands }) => {
          return commands.setMark(this.name, { commentId, color, resolved: false })
        },

      unsetCommentMark:
        (commentId: string) =>
        ({ tr, state, dispatch }) => {
          if (!dispatch) return true

          const { doc } = state
          const marksToRemove: { from: number; to: number }[] = []

          // 遍历文档找到所有带此 commentId 的标记
          doc.descendants((node, pos) => {
            if (!node.marks || node.marks.length === 0) return

            const hasCommentMark = node.marks.some(
              mark => mark.type.name === this.name && mark.attrs.commentId === commentId
            )

            if (hasCommentMark) {
              marksToRemove.push({ from: pos, to: pos + node.nodeSize })
            }
          })

          // 移除所有匹配的标记
          marksToRemove.forEach(({ from, to }) => {
            tr.removeMark(from, to, this.type)
          })

          return true
        },

      updateCommentMarkResolved:
        (commentId: string, resolved: boolean) =>
        ({ tr, state, dispatch }) => {
          if (!dispatch) return true

          const { doc } = state
          const marksToUpdate: { from: number; to: number; attrs: CommentMarkAttributes }[] = []

          // 遍历文档找到所有带此 commentId 的标记
          doc.descendants((node, pos) => {
            if (!node.marks || node.marks.length === 0) return

            const commentMark = node.marks.find(
              mark => mark.type.name === this.name && mark.attrs.commentId === commentId
            )

            if (commentMark) {
              marksToUpdate.push({
                from: pos,
                to: pos + node.nodeSize,
                attrs: {
                  commentId,
                  color: commentMark.attrs.color as string,
                  resolved,
                },
              })
            }
          })

          // 更新所有匹配的标记
          marksToUpdate.forEach(({ from, to, attrs }) => {
            tr.removeMark(from, to, this.type)
            tr.addMark(from, to, this.type.create(attrs))
          })

          return true
        },
    }
  },
})
