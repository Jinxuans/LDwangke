import request from '@/utils/http'

export interface LegacyPagination {
  page: number
  limit: number
  total: number
}

export interface LegacyOrderListParams {
  page?: number
  limit?: number
  user?: string
  pass?: string
  school?: string
  oid?: string
  cid?: string
  kcname?: string
  status_text?: string
  dock?: string
  uid?: string
  hid?: string
  search?: string
}

export interface LegacyOrderItem {
  oid: number
  uid: number
  cid: number
  hid: number
  ptname: string
  school: string
  name: string
  user: string
  pass: string
  kcname: string
  kcid: string
  status: string
  fees: string
  process: string
  remarks: string
  dockstatus: string
  yid: string
  addtime: string
  pushUid: string
  pushStatus: string
  pushEmail: string
  pushEmailStatus: string
  showdoc_push_url: string
  pushShowdocStatus: string
  supplier_pt: string
}

export interface LegacyOrderStats {
  total: number
  processing: number
  completed: number
  failed: number
  total_fees: number
}

export interface LegacyOrderLogEntry {
  time: string
  course: string
  status: string
  process: string
  remarks: string
}

export interface LegacyOrderListResult {
  list: LegacyOrderItem[]
  pagination: LegacyPagination
}

export function fetchLegacyOrderList(params: LegacyOrderListParams) {
  return request.post<LegacyOrderListResult>({
    url: '/order/list',
    params
  })
}

export function fetchLegacyOrderDetail(oid: number) {
  return request.get<LegacyOrderItem>({
    url: `/order/${oid}`
  })
}

export function fetchLegacyOrderStats() {
  return request.get<LegacyOrderStats>({
    url: '/order/stats'
  })
}

export function changeLegacyOrderStatus(params: {
  status: string
  oids: number[]
  type: number
}) {
  return request.post<void>({
    url: '/order/status',
    params
  })
}

export function refundLegacyOrders(oids: number[]) {
  return request.post<void>({
    url: '/order/refund',
    params: { oids }
  })
}

export function manualDockLegacyOrders(oids: number[]) {
  return request.post<{ fail: number; msg: string; success: number }>({
    url: '/admin/order/dock',
    params: { oids }
  })
}

export function syncLegacyOrderProgress(oids: number[]) {
  return request.post<{ msg: string; updated: number }>({
    url: '/admin/order/sync',
    params: { oids }
  })
}

export function batchSyncLegacyOrderProgress(oids: number[]) {
  return request.post<{ msg: string; updated: number }>({
    url: '/admin/order/batch-sync',
    params: { oids }
  })
}

export function batchResendLegacyOrders(oids: number[]) {
  return request.post<{ fail: number; msg: string; success: number }>({
    url: '/admin/order/batch-resend',
    params: { oids }
  })
}

export function modifyLegacyOrderRemarks(oids: number[], remarks: string) {
  return request.post<void>({
    url: '/admin/order/remarks',
    params: { oids, remarks }
  })
}

export function cancelLegacyOrder(oid: number) {
  return request.post<void>({
    url: '/order/cancel',
    params: { oid }
  })
}

export function pauseLegacyOrder(oid: number) {
  return request.get<{ message?: string; msg?: string }>({
    url: '/order/pause',
    params: { oid }
  })
}

export function changeLegacyOrderPassword(oid: number, newPwd: string) {
  return request.post<void>({
    url: '/order/changepass',
    params: { oid, newPwd }
  })
}

export function resubmitLegacyOrder(oid: number) {
  return request.get<{ message?: string; msg?: string }>({
    url: '/order/resubmit',
    params: { oid }
  })
}

export function fetchLegacyOrderLogs(oid: number) {
  return request.get<LegacyOrderLogEntry[]>({
    url: '/order/logs',
    params: { oid }
  })
}

export function fetchLegacyPupLogin(oid: number) {
  return request.get<{ url?: string }>({
    url: '/push/puplogin',
    params: { oid }
  })
}

export function resetLegacyPupOrder(
  oid: number,
  type: 'duration' | 'period' | 'score',
  value: number
) {
  return request.post<void>({
    url: '/order/pup-reset',
    params: { oid, type, value }
  })
}
