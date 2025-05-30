// components/realtime-status.tsx
'use client'

import { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { apiClient } from '@/lib/api-client'
import { 
  IconWifi, 
  IconWifiOff, 
  IconUsers, 
  IconMessages, 
  IconBell,
  IconTrendingUp,
  IconServer,
  IconActivity
} from '@tabler/icons-react'

interface RealtimeData {
  online_users: number
  active_sessions: number
  new_posts: number
  new_comments: number
  new_likes: number
  new_users: number
  new_reports: number
  system_load: number
  timestamp: string
}

export function RealtimeStatus() {
  const [isConnected, setIsConnected] = useState(false)
  const [data, setData] = useState<RealtimeData | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let ws: WebSocket | null = null

    const connectWebSocket = () => {
      try {
        ws = apiClient.connectWebSocket('dashboard', (newData: RealtimeData) => {
          setData(newData)
          setError(null)
        })

        if (ws) {
          ws.onopen = () => {
            setIsConnected(true)
            setError(null)
          }

          ws.onclose = () => {
            setIsConnected(false)
            // Attempt to reconnect after 5 seconds
            setTimeout(connectWebSocket, 5000)
          }

          ws.onerror = () => {
            setError('WebSocket connection failed')
            setIsConnected(false)
          }
        }
      } catch (err) {
        setError('Failed to connect to real-time updates')
        setIsConnected(false)
        // Retry connection after 5 seconds
        setTimeout(connectWebSocket, 5000)
      }
    }

    connectWebSocket()

    return () => {
      if (ws) {
        ws.close()
      }
    }
  }, [])

  if (error) {
    return (
      <Alert variant="destructive">
        <IconWifiOff className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-lg">
          <div className="flex items-center gap-2">
            {isConnected ? (
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <IconWifi className="h-5 w-5 text-green-600" />
              </div>
            ) : (
              <IconWifiOff className="h-5 w-5 text-red-600" />
            )}
            Real-time Activity
          </div>
        </CardTitle>
        <CardDescription>
          {isConnected ? 'Live system metrics' : 'Connecting to real-time updates...'}
          {data && (
            <span className="ml-2 text-xs">
              Last update: {new Date(data.timestamp).toLocaleTimeString()}
            </span>
          )}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {data ? (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconUsers className="h-4 w-4 text-blue-600" />
                <span className="text-2xl font-bold text-blue-600">
                  {data.online_users}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">Online Users</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconMessages className="h-4 w-4 text-green-600" />
                <span className="text-2xl font-bold text-green-600">
                  {data.new_posts}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">New Posts</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconBell className="h-4 w-4 text-orange-600" />
                <span className="text-2xl font-bold text-orange-600">
                  {data.new_reports}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">New Reports</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconServer className="h-4 w-4 text-purple-600" />
                <span className="text-2xl font-bold text-purple-600">
                  {data.system_load.toFixed(1)}%
                </span>
              </div>
              <div className="text-xs text-muted-foreground">System Load</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconActivity className="h-4 w-4 text-indigo-600" />
                <span className="text-2xl font-bold text-indigo-600">
                  {data.active_sessions}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">Active Sessions</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconTrendingUp className="h-4 w-4 text-pink-600" />
                <span className="text-2xl font-bold text-pink-600">
                  {data.new_likes}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">New Likes</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconUsers className="h-4 w-4 text-teal-600" />
                <span className="text-2xl font-bold text-teal-600">
                  {data.new_users}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">New Users</div>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center gap-1 mb-1">
                <IconMessages className="h-4 w-4 text-cyan-600" />
                <span className="text-2xl font-bold text-cyan-600">
                  {data.new_comments}
                </span>
              </div>
              <div className="text-xs text-muted-foreground">New Comments</div>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="text-center animate-pulse">
                <div className="h-8 bg-gray-200 rounded mb-1"></div>
                <div className="h-3 bg-gray-100 rounded"></div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

