type MessageHandler = (data: any) => void

export class WsClient {
  private ws: WebSocket | null = null
  private url: string
  private handlers = new Map<string, Set<MessageHandler>>()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectAttempts = 0
  private maxReconnect = 10
  private closed = false

  constructor(url: string) {
    this.url = url
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) return
    this.closed = false
    const token = localStorage.getItem('kx_token') || ''
    this.ws = new WebSocket(`${this.url}?token=${token}`)

    this.ws.onopen = () => {
      this.reconnectAttempts = 0
      this.startHeartbeat()
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        const type = msg.type as string
        this.handlers.get(type)?.forEach((fn) => fn(msg.data))
        this.handlers.get('*')?.forEach((fn) => fn(msg))
      } catch {
        // ignore non-json
      }
    }

    this.ws.onclose = () => {
      this.stopHeartbeat()
      if (!this.closed) this.reconnect()
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  private reconnect() {
    if (this.reconnectAttempts >= this.maxReconnect) return
    this.reconnectAttempts++
    const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 30000)
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      this.send('ping', {})
    }, 30000)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set())
    }
    this.handlers.get(type)!.add(handler)
  }

  off(type: string, handler: MessageHandler) {
    this.handlers.get(type)?.delete(handler)
  }

  send(type: string, data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }))
    }
  }

  close() {
    this.closed = true
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.stopHeartbeat()
    this.ws?.close()
    this.ws = null
  }
}

let docWs: WsClient | null = null

export function getDocWs(documentId: number): WsClient {
  if (docWs) docWs.close()
  const base = import.meta.env.VITE_WS_URL || `ws://${window.location.host}/ws`
  docWs = new WsClient(`${base}/doc/${documentId}`)
  docWs.connect()
  return docWs
}

export function closeDocWs() {
  docWs?.close()
  docWs = null
}
