import { useEffect, useRef, useState } from 'react'
import { SSEClient } from '@/utils/sse-client'

/**
 * React hook for managing Server-Sent Events connection
 *
 * @param eventType - The SSE event type to listen for (e.g., 'line.status')
 * @param onMessage - Callback function invoked when an event is received
 * @param enabled - Whether the SSE connection should be active (default: true)
 * @returns Object containing connection state and error information
 */
export function useSSE<T>(
  eventType: string,
  onMessage: (data: T) => void,
  enabled = true
): {
  connected: boolean
  error: Error | null
} {
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const clientRef = useRef<SSEClient | null>(null)

  useEffect(() => {
    // Skip if disabled
    if (!enabled) {
      return
    }

    // Get API base URL from environment
    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
    const sseUrl = `${apiBaseUrl}/events/stream`

    // Create SSE client if it doesn't exist
    if (!clientRef.current) {
      clientRef.current = new SSEClient(
        sseUrl,
        (type: string, data: any) => {
          // Only call onMessage for the specific event type we're interested in
          if (type === eventType) {
            onMessage(data as T)
          }
        },
        () => {
          setError(new Error('SSE connection error'))
        },
        (isConnected: boolean) => {
          setConnected(isConnected)
          if (isConnected) {
            setError(null)
          }
        }
      )
    }

    // Connect to SSE endpoint
    clientRef.current.connect()

    // Cleanup on unmount or when enabled changes
    return () => {
      if (clientRef.current) {
        clientRef.current.disconnect()
        clientRef.current = null
      }
    }
  }, [eventType, onMessage, enabled])

  return {
    connected,
    error,
  }
}
