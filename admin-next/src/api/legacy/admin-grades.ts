import request from '@/utils/http'

export interface LegacyAdminGrade {
  id: number
  sort: number
  name: string
  rate: string
  money: string
  addkf: string
  gjkf: string
  status: string
  time: string
}

export interface LegacyAdminGradeSavePayload {
  id?: number
  sort?: string
  name: string
  rate: string
  money?: string
  addkf?: string
  gjkf?: string
  status?: string
}

export function fetchLegacyAdminGrades() {
  return request.get<LegacyAdminGrade[]>({
    url: '/admin/grades'
  })
}

export function saveLegacyAdminGrade(data: LegacyAdminGradeSavePayload) {
  return request.post<void>({
    url: '/admin/grade/save',
    params: data
  })
}

export function deleteLegacyAdminGrade(id: number) {
  return request.del<void>({
    url: `/admin/grade/${id}`
  })
}
