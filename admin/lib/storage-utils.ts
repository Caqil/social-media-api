// lib/storage-utils.ts - Safe localStorage access
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
  
  // Auth-specific storage helpers
  export const authStorage = {
    getToken: (): string | null => storage.getItem('admin_token'),
    setToken: (token: string): boolean => storage.setItem('admin_token', token),
    removeToken: (): boolean => storage.removeItem('admin_token'),
  
    getRefreshToken: (): string | null => storage.getItem('admin_refresh_token'),
    setRefreshToken: (token: string): boolean => storage.setItem('admin_refresh_token', token),
    removeRefreshToken: (): boolean => storage.removeItem('admin_refresh_token'),
  
    getUser: <T>(): T | null => storage.getJSON<T>('admin_user'),
    setUser: <T>(user: T): boolean => storage.setJSON('admin_user', user),
    removeUser: (): boolean => storage.removeItem('admin_user'),
  
    clearAll: (): void => {
      authStorage.removeToken()
      authStorage.removeRefreshToken()
      authStorage.removeUser()
    },
  
    hasValidSession: (): boolean => {
      return !!(authStorage.getToken() && authStorage.getUser())
    }
  }