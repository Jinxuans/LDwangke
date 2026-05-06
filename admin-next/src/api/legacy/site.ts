import request from '@/utils/http'
import type { LegacySiteConfig } from '@/types/legacy-contract'

export function fetchLegacySiteConfig() {
  return request.get<LegacySiteConfig>({
    url: '/site/config'
  })
}
