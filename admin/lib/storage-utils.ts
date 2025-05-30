// lib/storage-utils.ts - Enhanced version with better error handling
export const storage = {
    // Safe localStorage getter
    getItem: (key: string): string | null => {
      if (typeof window === 'undefined') {
        return null
      }
      try {
        return localStorage.getItem(key)
      } catch (error) {
        console.error(`Error reading localStorage key "${key}":`, error)
        return null
      }
    },
  
    // Safe localStorage setter
    setItem: (key: string, value: string): boolean => {
      if (typeof window === 'undefined') {
        return false
      }
      try {
        localStorage.setItem(key, value)
        return true
      } catch (error) {
        console.error(`Error setting localStorage key "${key}":`, error)
        return false
      }
    },
  
    // Safe localStorage remover
    removeItem: (key: string): boolean => {
      if (typeof window === 'undefined') {
        return false
      }
      try {
        localStorage.removeItem(key)
        return true
      } catch (error) {
        console.error(`Error removing localStorage key "${key}":`, error)
        return false
      }
    },
  
    // Safe localStorage clear
    clear: (): boolean => {
      if (typeof window === 'undefined') {
        return false
      }
      try {
        localStorage.clear()
        return true
      } catch (error) {
        console.error('Error clearing localStorage:', error)
        return false
      }
    },
  
    // Check if localStorage is available
    isAvailable: (): boolean => {
      if (typeof window === 'undefined') {
        return false
      }
      try {
        const test = '__localStorage_test__'
        localStorage.setItem(test, test)
        localStorage.removeItem(test)
        return true
      } catch (error) {
        return false
      }
    },
  
    // Get parsed JSON safely
    getJSON: <T>(key: string, defaultValue: T | null = null): T | null => {
      const item = storage.getItem(key)
      if (!item) {
        return defaultValue
      }
      try {
        return JSON.parse(item) as T
      } catch (error) {
        console.error(`Error parsing JSON from localStorage key "${key}":`, error)
        return defaultValue
      }
    },
  
    // Set JSON safely
    setJSON: <T>(key: string, value: T): boolean => {
      try {
        return storage.setItem(key, JSON.stringify(value))
      } catch (error) {
        console.error(`Error stringifying JSON for localStorage key "${key}":`, error)
        return false
      }
    }
  }
  
  // Hook for safe localStorage usage with SSR
  import { useState, useEffect } from 'react'
  
  export function useLocalStorage<T>(
    key: string,
    initialValue: T
  ): [T, (value: T | ((val: T) => T)) => void, boolean] {
    const [storedValue, setStoredValue] = useState<T>(initialValue)
    const [isLoading, setIsLoading] = useState(true)
  
    useEffect(() => {
      // Only run on client side
      if (typeof window === 'undefined') {
        setIsLoading(false)
        return
      }
  
      try {
        const item = storage.getItem(key)
        if (item) {
          setStoredValue(JSON.parse(item))
        }
      } catch (error) {
        console.error(`Error loading localStorage key "${key}":`, error)
      } finally {
        setIsLoading(false)
      }
    }, [key])
  
    const setValue = (value: T | ((val: T) => T)) => {
      try {
        const valueToStore = value instanceof Function ? value(storedValue) : value
        setStoredValue(valueToStore)
        storage.setJSON(key, valueToStore)
      } catch (error) {
        console.error(`Error setting localStorage key "${key}":`, error)
      }
    }
  
    return [storedValue, setValue, isLoading]
  }
  
  // Auth-specific storage helpers with better error handling
  export const authStorage = {
    getToken: (): string | null => {
      try {
        return storage.getItem('admin_token')
      } catch (error) {
        console.error('Error getting admin token:', error)
        return null
      }
    },
    
    setToken: (token: string): boolean => {
      try {
        return storage.setItem('admin_token', token)
      } catch (error) {
        console.error('Error setting admin token:', error)
        return false
      }
    },
    
    removeToken: (): boolean => {
      try {
        return storage.removeItem('admin_token')
      } catch (error) {
        console.error('Error removing admin token:', error)
        return false
      }
    },
  
    getRefreshToken: (): string | null => {
      try {
        return storage.getItem('admin_refresh_token')
      } catch (error) {
        console.error('Error getting refresh token:', error)
        return null
      }
    },
    
    setRefreshToken: (token: string): boolean => {
      try {
        return storage.setItem('admin_refresh_token', token)
      } catch (error) {
        console.error('Error setting refresh token:', error)
        return false
      }
    },
    
    removeRefreshToken: (): boolean => {
      try {
        return storage.removeItem('admin_refresh_token')
      } catch (error) {
        console.error('Error removing refresh token:', error)
        return false
      }
    },
  
    getUser: <T>(): T | null => {
      try {
        return storage.getJSON<T>('admin_user')
      } catch (error) {
        console.error('Error getting user data:', error)
        return null
      }
    },
    
    setUser: <T>(user: T): boolean => {
      try {
        return storage.setJSON('admin_user', user)
      } catch (error) {
        console.error('Error setting user data:', error)
        return false
      }
    },
    
    removeUser: (): boolean => {
      try {
        return storage.removeItem('admin_user')
      } catch (error) {
        console.error('Error removing user data:', error)
        return false
      }
    },
  
    clearAll: (): boolean => {
      try {
        const results = [
          authStorage.removeToken(),
          authStorage.removeRefreshToken(),
          authStorage.removeUser()
        ]
        return results.every(result => result)
      } catch (error) {
        console.error('Error clearing auth data:', error)
        return false
      }
    },
  
    hasValidSession: (): boolean => {
      try {
        const token = authStorage.getToken()
        const user = authStorage.getUser()
        
        if (!token || !user) {
          return false
        }
  
        // Basic token validation - check if it's properly formatted
        const parts = token.split('.')
        if (parts.length !== 3) {
          return false
        }
  
        // Check if token is expired
        try {
          const payload = JSON.parse(atob(parts[1]))
          const currentTime = Math.floor(Date.now() / 1000)
          
          if (payload.exp && payload.exp < currentTime) {
            return false
          }
        } catch (error) {
          console.error('Error validating token:', error)
          return false
        }
  
        return true
      } catch (error) {
        console.error('Error checking session validity:', error)
        return false
      }
    },
  
    // Get token expiration time
    getTokenExpiration: (): number | null => {
      try {
        const token = authStorage.getToken()
        if (!token) {
          return null
        }
  
        const parts = token.split('.')
        if (parts.length !== 3) {
          return null
        }
  
        const payload = JSON.parse(atob(parts[1]))
        return payload.exp || null
      } catch (error) {
        console.error('Error getting token expiration:', error)
        return null
      }
    },
  
    // Check if token will expire soon (within next 5 minutes)
    isTokenExpiringSoon: (): boolean => {
      try {
        const exp = authStorage.getTokenExpiration()
        if (!exp) {
          return true // Assume expiring if we can't determine
        }
  
        const currentTime = Math.floor(Date.now() / 1000)
        const fiveMinutes = 5 * 60
        
        return (exp - currentTime) <= fiveMinutes
      } catch (error) {
        console.error('Error checking token expiration:', error)
        return true
      }
    }
  }