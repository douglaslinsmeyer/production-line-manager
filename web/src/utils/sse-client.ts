/**
 * SSE Client for managing Server-Sent Events connection
 * Handles connection lifecycle, auto-reconnection, and event dispatching
 */
export class SSEClient {
  private eventSource: EventSource | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 3000 // Start with 3 seconds
  private maxReconnectDelay = 30000 // Max 30 seconds
  private isManuallyDisconnected = false
  private url: string
  private onEvent: (type: string, data: any) => void
  private onError?: (error: Event) => void
  private onConnectionChange?: (connected: boolean) => void

  constructor(
    url: string,
    onEvent: (type: string, data: any) => void,
    onError?: (error: Event) => void,
    onConnectionChange?: (connected: boolean) => void
  ) {
    this.url = url
    this.onEvent = onEvent
    this.onError = onError
    this.onConnectionChange = onConnectionChange
  }

  /**
   * Connect to the SSE endpoint
   */
  connect(): void {
    if (this.eventSource) {
      console.warn('SSE client already connected')
      return
    }

    this.isManuallyDisconnected = false
    console.log('Connecting to SSE endpoint:', this.url)

    try {
      this.eventSource = new EventSource(this.url)

      // Handle successful connection
      this.eventSource.addEventListener('connected', (event) => {
        console.log('SSE connected:', event.data)
        this.resetReconnectDelay()
        this.onConnectionChange?.(true)
      })

      // Listen for specific event types
      this.addEventListener('line.status')
      this.addEventListener('line.updated')
      this.addEventListener('line.created')
      this.addEventListener('line.deleted')

      // Handle errors
      this.eventSource.onerror = (error) => {
        console.error('SSE error:', error)
        this.onConnectionChange?.(false)
        this.onError?.(error)

        // Auto-reconnect unless manually disconnected
        if (!this.isManuallyDisconnected) {
          this.scheduleReconnect()
        }
      }
    } catch (error) {
      console.error('Failed to create EventSource:', error)
      this.onConnectionChange?.(false)
      if (!this.isManuallyDisconnected) {
        this.scheduleReconnect()
      }
    }
  }

  /**
   * Disconnect from the SSE endpoint
   */
  disconnect(): void {
    console.log('Disconnecting SSE client')
    this.isManuallyDisconnected = true

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
      this.onConnectionChange?.(false)
    }
  }

  /**
   * Check if currently connected
   */
  isConnected(): boolean {
    return this.eventSource !== null && this.eventSource.readyState === EventSource.OPEN
  }

  /**
   * Add event listener for a specific event type
   */
  private addEventListener(eventType: string): void {
    if (!this.eventSource) return

    this.eventSource.addEventListener(eventType, (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        this.onEvent(eventType, data)
      } catch (error) {
        console.error(`Failed to parse SSE event data for ${eventType}:`, error)
      }
    })
  }

  /**
   * Schedule automatic reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    // Close existing connection
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }

    // Clear existing timer
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    console.log(`Scheduling reconnect in ${this.reconnectDelay}ms`)

    this.reconnectTimer = setTimeout(() => {
      console.log('Attempting to reconnect...')
      this.connect()

      // Exponential backoff
      this.reconnectDelay = Math.min(
        this.reconnectDelay * 2,
        this.maxReconnectDelay
      )
    }, this.reconnectDelay)
  }

  /**
   * Reset reconnect delay after successful connection
   */
  private resetReconnectDelay(): void {
    this.reconnectDelay = 3000 // Reset to 3 seconds
  }
}
