import request from '@/utils/http'

export interface LegacyDynamicModule {
  id: number
  app_id: string
  type: string
  name: string
  description: string
  price: string
  icon: string
  api_base: string
  view_url: string
  status: number
  sort: number
  config: string
}

export type LegacyDynamicModuleSavePayload = Partial<LegacyDynamicModule>

export function fetchLegacyDynamicModules() {
  return request.get<LegacyDynamicModule[]>({
    url: '/admin/modules'
  })
}

export function saveLegacyDynamicModule(data: LegacyDynamicModuleSavePayload) {
  return request.post<void>({
    url: '/admin/module/save',
    params: data
  })
}

export function deleteLegacyDynamicModule(id: number) {
  return request.del<void>({
    url: `/admin/module/${id}`
  })
}
