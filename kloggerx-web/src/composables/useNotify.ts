import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import type { MessageParams, MessageBoxData, NotificationParams } from 'element-plus'
import { createNotification } from '@/api/modules/collaborate'
import i18n from '@/locales'

export type NotifyType = 'success' | 'warning' | 'info' | 'error' | 'system'

interface NotifyOptions {
  message: string
  title?: string
  type?: NotifyType
  duration?: number
  showInCenter?: boolean
}

// 通知类型映射
const notificationTypeMap: Record<NotifyType, string> = {
  success: 'success',
  warning: 'warning',
  info: 'info',
  error: 'error',
  system: 'system',
}

/**
 * 全局通知服务
 * 同时显示页面提示和发送到通知中心
 */
export function useNotify() {
  /**
   * 发送通知（同时显示ElMessage和发送到通知中心）
   */
  const notify = async (options: NotifyOptions | string) => {
    const opts = typeof options === 'string' ? { message: options } : options
    const { message, title, type = 'info', duration = 3000, showInCenter = true } = opts

    // 1. 显示ElMessage
    ElMessage({
      message,
      type: type === 'system' ? 'info' : type,
      duration,
    })

    // 2. 发送到通知中心（默认启用）
    if (showInCenter) {
      try {
        await createNotification({
          type: notificationTypeMap[type],
          title: title || getDefaultTitle(type),
          content: message,
        })
      } catch (e) {
        // 静默失败，不影响用户体验
        console.warn('Failed to send notification to center:', e)
      }
    }
  }

  /**
   * 成功通知
   */
  const success = (message: string, title?: string) => {
    return notify({ message, title, type: 'success' })
  }

  /**
   * 警告通知
   */
  const warning = (message: string, title?: string) => {
    return notify({ message, title, type: 'warning' })
  }

  /**
   * 信息通知
   */
  const info = (message: string, title?: string) => {
    return notify({ message, title, type: 'info' })
  }

  /**
   * 错误通知
   */
  const error = (message: string, title?: string) => {
    return notify({ message, title, type: 'error' })
  }

  /**
   * 系统通知
   */
  const system = (message: string, title?: string) => {
    return notify({ message, title, type: 'system' })
  }

  /**
   * 仅显示ElMessage（不发送到通知中心）
   */
  const toast = (options: MessageParams | string) => {
    return ElMessage(options)
  }

  /**
   * 显示弹窗确认框
   */
  const confirm = (message: string, title?: string): Promise<MessageBoxData> => {
    return ElMessageBox.confirm(message, title || i18n.global.t('notify.confirm'))
  }

  /**
   * 显示桌面通知（ElNotification）
   */
  const notification = (options: NotificationParams) => {
    return ElNotification(options)
  }

  return {
    notify,
    success,
    warning,
    info,
    error,
    system,
    toast,
    confirm,
    notification,
  }
}

// 获取默认标题
function getDefaultTitle(type: NotifyType): string {
  const t = i18n.global.t
  const titles: Record<NotifyType, string> = {
    success: t('notify.defaultTitle.success'),
    warning: t('notify.defaultTitle.warning'),
    info: t('notify.defaultTitle.info'),
    error: t('notify.defaultTitle.error'),
    system: t('notify.defaultTitle.system'),
  }
  return titles[type]
}

// 默认导出一个单例
export const $notify = {
  notify: async (options: NotifyOptions | string) => {
    const { notify } = useNotify()
    return notify(options)
  },
  success: (message: string, title?: string) => {
    const { success } = useNotify()
    return success(message, title)
  },
  warning: (message: string, title?: string) => {
    const { warning } = useNotify()
    return warning(message, title)
  },
  info: (message: string, title?: string) => {
    const { info } = useNotify()
    return info(message, title)
  },
  error: (message: string, title?: string) => {
    const { error } = useNotify()
    return error(message, title)
  },
  system: (message: string, title?: string) => {
    const { system } = useNotify()
    return system(message, title)
  },
  toast: (options: MessageParams | string) => {
    return ElMessage(options)
  },
  confirm: (message: string, title?: string) => {
    return ElMessageBox.confirm(message, title || i18n.global.t('notify.confirm'))
  },
  notification: (options: NotificationParams) => {
    return ElNotification(options)
  },
}
