import request from '@/utils/http'

export interface LegacyUserCheckinStatus {
  checked_in: boolean
  reward_money?: number
}

export interface LegacyUserCheckinResult {
  reward_money: number
}

export interface LegacyPublicActivity {
  hid: number
  name: string
  yaoqiu: string
  type: string
  num: string
  money: string
  addtime: string
  endtime: string
  status_ok: string
  status: string
}

export interface LegacyPledgeConfig {
  id: number
  category_id: number
  amount: number
  discount_rate: number
  status: number
  addtime: string
  days: number
  cancel_fee: number
  category_name?: string
}

export interface LegacyPledgeRecord {
  id: number
  uid: number
  config_id: number
  status: number
  addtime: string
  endtime: string | null
  amount?: number
  category_name?: string
  discount_rate?: number
  days?: number
  username?: string
}

export function fetchLegacyUserCheckinStatus() {
  return request.get<LegacyUserCheckinStatus>({
    url: '/user/checkin/status'
  })
}

export function submitLegacyUserCheckin() {
  return request.post<LegacyUserCheckinResult>({
    url: '/user/checkin'
  })
}

export function fetchLegacyPublicActivities() {
  return request.get<LegacyPublicActivity[]>({
    url: '/activities'
  })
}

export function fetchLegacyPledgeConfigs() {
  return request.get<LegacyPledgeConfig[]>({
    url: '/pledge/configs'
  })
}

export function createLegacyPledge(configId: number) {
  return request.post<void>({
    url: '/pledge/create',
    params: {
      config_id: configId
    }
  })
}

export function cancelLegacyPledge(id: number) {
  return request.post<void>({
    url: `/pledge/cancel/${id}`
  })
}

export function fetchLegacyMyPledges() {
  return request.get<LegacyPledgeRecord[]>({
    url: '/pledge/my'
  })
}

export interface LegacyCheckOrderResult {
  oid: number
  ptname: string
  account: string
  school?: string
  kcname: string
  status: string
  process: string
  remarks: string
  addtime: string
  pushUid?: string
  pushStatus?: string
  pushEmail?: string
  pushEmailStatus?: string
  showdoc_push_url?: string
  pushShowdocStatus?: string
}

export function checkLegacyOrder(params: { user?: string; oid?: string }) {
  return request.get<{
    list: LegacyCheckOrderResult[]
    total: number
  }>({
    url: '/query',
    params
  })
}

export function bindLegacyWxPush(data: { account: string; pushUid: string; oids: string }) {
  return request.post<void>({
    url: '/push/bind-wx',
    params: data
  })
}

export function unbindLegacyWxPush(data: { account: string }) {
  return request.post<void>({
    url: '/push/unbind-wx',
    params: data
  })
}

export function bindLegacyEmailPush(data: { account?: string; orderid?: number; pushEmail: string }) {
  return request.post<void>({
    url: '/push/bind-email',
    params: data
  })
}

export function unbindLegacyEmailPush(data: { account?: string; orderid?: number }) {
  return request.post<void>({
    url: '/push/unbind-email',
    params: data
  })
}

export function fetchLegacyWxQrcode(data: { account: string }) {
  return request.post<{
    code: string
    url: string
    shortUrl?: string
  }>({
    url: '/push/wx-qrcode',
    params: data
  })
}

export function fetchLegacyWxScanUid(data: { code: string }) {
  return request.post<{
    uid: string
  }>({
    url: '/push/wx-scan-uid',
    params: data
  })
}

export function fetchLegacyPublicPupLogin(oid: number) {
  return request.get<{
    url?: string
  }>({
    url: '/push/puplogin',
    params: { oid }
  })
}
