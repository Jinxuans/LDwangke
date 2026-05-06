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

export function fetchLegacyModulesByType(type: string) {
  return request.get<LegacyDynamicModule[]>({
    url: '/modules',
    params: { type }
  })
}

export function fetchLegacyModuleFrameUrl(appId: string) {
  return request.get<{
    frame_url?: string
  }>({
    url: `/module/${appId}/frame-url`
  })
}
