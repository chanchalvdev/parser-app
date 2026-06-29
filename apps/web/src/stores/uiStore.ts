import { create } from 'zustand'

const DEFAULT_TENANT = '11111111-1111-1111-1111-111111111001'

type UiState = {
  tenantId: string
  sidebarCollapsed: boolean
  setTenantId: (tenantId: string) => void
  setSidebarCollapsed: (value: boolean) => void
  toggleSidebar: () => void
}

export const useUiStore = create<UiState>((set) => ({
  tenantId: DEFAULT_TENANT,
  sidebarCollapsed: false,
  setTenantId: (tenantId) => set({ tenantId }),
  setSidebarCollapsed: (value) => set({ sidebarCollapsed: value }),
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
}))

export const DEFAULT_TENANT_ID = DEFAULT_TENANT
