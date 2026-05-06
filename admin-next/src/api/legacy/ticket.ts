import request from '@/utils/http'
import type { LegacyPagination } from './order'

export interface LegacyTicket {
  id: number
  uid: number
  oid: number
  type: string
  content: string
  reply: string
  status: number
  addtime: string
  reply_time: string
  supplier_report_id: number
  supplier_status: number
  supplier_answer: string
  order_user?: string
  order_pt?: string
  order_status?: string
  order_yid?: string
  supplier_report_switch: number
  supplier_report_hid_switch: number
}

export interface LegacyTicketListResult {
  list: LegacyTicket[]
  pagination: LegacyPagination
}

export interface LegacyTicketStats {
  total: number
  pending: number
  replied: number
  closed: number
  upstream_pending: number
}

export interface LegacyTicketSyncReportResult {
  supplier_status: number
  supplier_answer: string
}

export function fetchLegacyUserTickets(page = 1, limit = 20) {
  return request.get<LegacyTicketListResult>({
    url: '/user/tickets',
    params: { page, limit }
  })
}

export function createLegacyUserTicket(params: {
  oid?: number
  type?: string
  content: string
}) {
  return request.post<void>({
    url: '/user/ticket/create',
    params
  })
}

export function closeLegacyUserTicket(id: number) {
  return request.post<void>({
    url: `/user/ticket/close/${id}`
  })
}

export function fetchLegacyAdminTickets(params: {
  page?: number
  limit?: number
  status?: number
  uid?: number
  search?: string
}) {
  return request.get<LegacyTicketListResult>({
    url: '/admin/tickets',
    params
  })
}

export function fetchLegacyAdminTicketStats() {
  return request.get<LegacyTicketStats>({
    url: '/admin/ticket/stats'
  })
}

export function replyLegacyAdminTicket(id: number, reply: string) {
  return request.post<void>({
    url: '/admin/ticket/reply',
    params: { id, reply }
  })
}

export function closeLegacyAdminTicket(id: number) {
  return request.post<void>({
    url: `/admin/ticket/close/${id}`
  })
}

export function autoCloseLegacyAdminTickets(days: number) {
  return request.post<{ closed: number }>({
    url: '/admin/ticket/auto-close',
    params: { days }
  })
}

export function reportLegacyAdminTicket(ticketId: number) {
  return request.post<{ report_id?: number; message?: string }>({
    url: '/admin/ticket/report',
    params: { ticket_id: ticketId }
  })
}

export function syncLegacyAdminTicketReport(ticketId: number) {
  return request.post<LegacyTicketSyncReportResult>({
    url: '/admin/ticket/sync-report',
    params: { ticket_id: ticketId }
  })
}
