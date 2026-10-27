import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { UserResponse } from '@/types/api'

interface AuthState {
  auth: {
    user: UserResponse | null
    setUser: (user: UserResponse | null) => void
    accessToken: string
    setAccessToken: (accessToken: string) => void
    resetAccessToken: () => void
    reset: () => void
  }
}

const getMemoryStorage = (): Record<string, string> => {
  const g = globalThis as unknown as { __NEWSROOM_STORAGE__?: Record<string, string> }
  if (!g.__NEWSROOM_STORAGE__) {
    g.__NEWSROOM_STORAGE__ = {}
  }
  return g.__NEWSROOM_STORAGE__
}

const customStorage = {
  getItem: (key: string) => {
    if (typeof window !== 'undefined' && window.localStorage) {
      return window.localStorage.getItem(key)
    }
    return getMemoryStorage()[key] ?? null
  },
  setItem: (key: string, value: string) => {
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.setItem(key, value)
    } else {
      getMemoryStorage()[key] = value
    }
  },
  removeItem: (key: string) => {
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.removeItem(key)
    } else {
      delete getMemoryStorage()[key]
    }
  },
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      auth: {
        user: null,
        setUser: (user) =>
          set((state) => ({ ...state, auth: { ...state.auth, user } })),
        accessToken: '',
        setAccessToken: (accessToken) =>
          set((state) => ({ ...state, auth: { ...state.auth, accessToken } })),
        resetAccessToken: () =>
          set((state) => ({ ...state, auth: { ...state.auth, accessToken: '' } })),
        reset: () =>
          set((state) => ({
            ...state,
            auth: { ...state.auth, user: null, accessToken: '' },
          })),
      },
    }),
    {
      name: 'newsroom_auth',
      storage: createJSONStorage(() => customStorage),
      partialize: (state) => ({
        auth: {
          user: state.auth.user,
          accessToken: state.auth.accessToken,
        },
      }),
      merge: (persistedState: unknown, currentState) => ({
        ...currentState,
        auth: {
          ...currentState.auth,
          ...((persistedState as { auth?: { user?: UserResponse | null; accessToken?: string } })?.auth || {}),
        },
      }),
    }
  )
)
