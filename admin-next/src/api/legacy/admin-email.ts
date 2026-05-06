import request from '@/utils/http'

export interface LegacyEmailPoolAccount {
  id: number
  name: string
  host: string
  port: number
  encryption: string
  user: string
  password?: string
  from_email: string
  weight: number
  day_limit: number
  hour_limit: number
  today_sent: number
  hour_sent: number
  total_sent: number
  total_fail: number
  fail_streak: number
  status: number
  last_used: string
  last_error: string
  addtime: string
}

export interface LegacyEmailPoolSaveRequest {
  id?: number
  name: string
  host: string
  port: number
  encryption: string
  user: string
  password?: string
  from_email?: string
  weight?: number
  day_limit?: number
  hour_limit?: number
  status?: number
}

export interface LegacyEmailSendLog {
  id: number
  pool_id: number
  from_email: string
  to_email: string
  subject: string
  mail_type: string
  status: number
  error: string
  addtime: string
}

export interface LegacyEmailPoolStats {
  total_accounts: number
  active_accounts: number
  error_accounts: number
  today_sent: number
  today_fail: number
}

export interface LegacyEmailTemplate {
  id: number
  code: string
  name: string
  subject: string
  content: string
  variables: string
  status: number
  updated_at: string
  created_at: string
}

export interface LegacyMassEmailLog {
  id: number
  target: string
  subject: string
  total: number
  success_count: number
  fail_count: number
  status: string
  addtime: string
}

export interface LegacyMassEmailLogResult {
  list: LegacyMassEmailLog[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export function fetchLegacyEmailPoolAccounts() {
  return request.get<LegacyEmailPoolAccount[]>({
    url: '/admin/email-pool'
  })
}

export function fetchLegacyEmailPoolStats() {
  return request.get<LegacyEmailPoolStats>({
    url: '/admin/email-pool/stats'
  })
}

export function saveLegacyEmailPoolAccount(data: LegacyEmailPoolSaveRequest) {
  return request.post<void>({
    url: '/admin/email-pool/save',
    params: data
  })
}

export function deleteLegacyEmailPoolAccount(id: number) {
  return request.del<void>({
    url: `/admin/email-pool/${id}`
  })
}

export function toggleLegacyEmailPoolAccount(id: number, status: number) {
  return request.post<void>({
    url: '/admin/email-pool/toggle',
    params: { id, status }
  })
}

export function testLegacyEmailPoolAccount(id: number, testTo: string) {
  return request.post<void>({
    url: '/admin/email-pool/test',
    params: { id, test_to: testTo }
  })
}

export function resetLegacyEmailPoolCounters() {
  return request.post<void>({
    url: '/admin/email-pool/reset-counters'
  })
}

export function fetchLegacyEmailSendLogs(params: {
  limit?: number
  mail_type?: string
  page?: number
  status?: number
  to_email?: string
}) {
  return request.get<{
    list: LegacyEmailSendLog[]
    total: number
  }>({
    url: '/admin/email-send-logs',
    params
  })
}

export function fetchLegacyEmailTemplates() {
  return request.get<LegacyEmailTemplate[]>({
    url: '/admin/email-templates'
  })
}

export function saveLegacyEmailTemplate(data: {
  id: number
  subject: string
  content: string
  status: number
}) {
  return request.post<void>({
    url: '/admin/email-templates/save',
    params: data
  })
}

export function previewLegacyEmailTemplate(code: string) {
  return request.get<{
    subject: string
    html: string
  }>({
    url: '/admin/email-templates/preview',
    params: { code }
  })
}

export function testLegacyEmailTemplate(code: string, testTo: string) {
  return request.post<void>({
    url: '/admin/email-templates/test',
    params: { code, test_to: testTo }
  })
}

export function sendLegacyMassEmail(data: {
  target: string
  subject: string
  content: string
}) {
  return request.post<{
    log_id: number
    message: string
  }>({
    url: '/admin/email/send',
    params: data
  })
}

export function fetchLegacyMassEmailLogs(params: {
  page?: number
  limit?: number
}) {
  return request.get<LegacyMassEmailLogResult>({
    url: '/admin/email/logs',
    params
  })
}

export function previewLegacyEmailRecipients(target: string) {
  return request.get<{
    count: number
  }>({
    url: '/admin/email/preview',
    params: { target }
  })
}
