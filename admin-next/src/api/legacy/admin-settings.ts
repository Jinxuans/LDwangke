import request from '@/utils/http'
import type { LegacyAdminConfigMap } from '@/types/legacy-contract'

export function fetchLegacyAdminConfig() {
  return request.get<LegacyAdminConfigMap>({
    url: '/admin/config'
  })
}

export function saveLegacyAdminConfig(data: LegacyAdminConfigMap) {
  return request.post<void>({
    url: '/admin/config',
    params: data
  })
}
