import { create } from 'zustand'

interface AppState {
  isLoading: boolean
  user: {
    id: number | null
    username: string
    tenantId: number | null
  } | null
  setLoading: (loading: boolean) => void
  setUser: (user: AppState['user']) => void
}

export const useAppStore = create<AppState>((set) => ({
  isLoading: false,
  user: null,
  setLoading: (loading) => set({ isLoading: loading }),
  setUser: (user) => set({ user }),
}))