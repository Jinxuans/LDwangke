import request from '@/utils/http'
import type { LegacyExtMenuItem, LegacyMenuConfigItem } from '@/types/legacy-contract'

export function fetchLegacyMenuConfigs() {
  return request.get<LegacyMenuConfigItem[]>({
    url: '/menus'
  })
}

export function saveLegacyMenuConfigs(items: LegacyMenuConfigItem[]) {
  return request.post<void>({
    url: '/admin/menus',
    data: { items }
  })
}

export function fetchLegacyExtMenus() {
  return request.get<LegacyExtMenuItem[]>({
    url: '/admin/ext-menus'
  })
}

export function saveLegacyExtMenu(data: Partial<LegacyExtMenuItem>) {
  return request.post<void>({
    url: '/admin/ext-menu/save',
    data
  })
}

export function reorderLegacyExtMenus(items: { id: number; sort_order: number }[]) {
  return request.post<void>({
    url: '/admin/ext-menu/reorder',
    data: { items }
  })
}

export function deleteLegacyExtMenu(id: number) {
  return request.del<void>({
    url: `/admin/ext-menu/${id}`
  })
}
