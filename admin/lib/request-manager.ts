// lib/request-manager.ts - Global Request Deduplication
interface RequestCache {
    [key: string]: {
      promise: Promise<any>
      timestamp: number
      data?: any
    }
  }
  
  class RequestManager {
    private static instance: RequestManager | null = null
    private cache: RequestCache = {}
    private readonly CACHE_DURATION = 30000 // 30 seconds
  
    private constructor() {
      console.log('🔧 Request Manager initialized')
    }
  
    public static getInstance(): RequestManager {
      if (!RequestManager.instance) {
        RequestManager.instance = new RequestManager()
      }
      return RequestManager.instance
    }
  
    private getCacheKey(url: string, params?: any): string {
      return `${url}${params ? JSON.stringify(params) : ''}`
    }
  
    private isValidCache(cacheEntry: { timestamp: number, data?: any }): boolean {
      return Date.now() - cacheEntry.timestamp < this.CACHE_DURATION && cacheEntry.data
    }
  
    async request<T>(
      key: string,
      requestFn: () => Promise<T>,
      options: { 
        cache?: boolean
        cacheDuration?: number
      } = {}
    ): Promise<T> {
      const { cache = true, cacheDuration = this.CACHE_DURATION } = options
  
      // Check if we have a valid cached result
      if (cache && this.cache[key]) {
        const cacheEntry = this.cache[key]
        
        // Return cached data if still valid
        if (this.isValidCache(cacheEntry)) {
          console.log(`📦 Returning cached result for: ${key}`)
          return cacheEntry.data
        }
  
        // If there's an ongoing request, wait for it
        if (cacheEntry.promise) {
          console.log(`⏳ Waiting for ongoing request: ${key}`)
          try {
            const result = await cacheEntry.promise
            return result
          } catch (error) {
            // If the ongoing request fails, remove it and continue
            delete this.cache[key]
            throw error
          }
        }
      }
  
      console.log(`🔄 Making new request: ${key}`)
  
      // Create new request
      const promise = requestFn()
      
      // Store in cache
      this.cache[key] = {
        promise,
        timestamp: Date.now()
      }
  
      try {
        const result = await promise
        
        // Update cache with result
        this.cache[key] = {
          promise,
          timestamp: Date.now(),
          data: result
        }
        
        console.log(`✅ Request completed: ${key}`)
        return result
      } catch (error) {
        // Remove failed request from cache
        delete this.cache[key]
        console.error(`❌ Request failed: ${key}`, error)
        throw error
      }
    }
  
    clearCache(key?: string): void {
      if (key) {
        delete this.cache[key]
        console.log(`🧹 Cleared cache for: ${key}`)
      } else {
        this.cache = {}
        console.log('🧹 Cleared all cache')
      }
    }
  
    getCacheInfo(): { [key: string]: { hasData: boolean, age: number } } {
      const info: { [key: string]: { hasData: boolean, age: number } } = {}
      
      Object.keys(this.cache).forEach(key => {
        const entry = this.cache[key]
        info[key] = {
          hasData: !!entry.data,
          age: Date.now() - entry.timestamp
        }
      })
      
      return info
    }
  
    // Cleanup old cache entries
    cleanup(): void {
      const now = Date.now()
      Object.keys(this.cache).forEach(key => {
        const entry = this.cache[key]
        if (now - entry.timestamp > this.CACHE_DURATION * 2) {
          delete this.cache[key]
        }
      })
    }
  }
  
  export const requestManager = RequestManager.getInstance()
  
  // Cleanup interval
  if (typeof window !== 'undefined') {
    setInterval(() => {
      requestManager.cleanup()
    }, 60000) // Cleanup every minute
  }