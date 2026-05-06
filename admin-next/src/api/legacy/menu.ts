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
    params: { items },
    showSuccessMessage: true
  })
}

export function fetchLegacyExtMenus() {
  return request.get<LegacyExtMenuItem[]>({
    url: '/ext-menus'
  })
}
