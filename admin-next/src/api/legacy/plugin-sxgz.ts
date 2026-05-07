import request from '@/utils/http'

export interface LegacySXGZPricingRule {
  affected_by_user_rate?: boolean
  price: number
}

export interface LegacySXGZPrintPricing {
  base_free_copies: number
  extra_copy_price: number
  per_copy_price: number
}

export interface LegacySXGZConfig {
  auto_sync: boolean
  delivery_options: Record<string, LegacySXGZPricingRule>
  file_base_url?: string
  price_multiplier: number
  print_options: Record<string, LegacySXGZPricingRule>
  print_pricing: LegacySXGZPrintPricing
  sync_interval: number
  upstream_key: string
  upstream_protocol: 'same_system' | 'source29'
  upstream_uid: number
  upstream_url: string
}

export interface LegacySXGZCompany {
  cid: number
  content: string
  license_price: number
  name: string
  price: number
  source: string
  status: number
  updated_at: string
}

export interface LegacySXGZQuoteRequest {
  business_license: boolean
  company_id: number
  courier_company?: string
  customer_address?: string
  customer_email?: string
  customer_name: string
  customer_phone?: string
  delivery_option?: string
  file_print_options?: LegacySXGZFilePrintOption[]
  material_type: string
  only_business_license: boolean
  paper_size: string
  print_copies: number
  print_options: string[]
  return_tracking_number?: string
  selected_license_companies: number[]
  service_type: string
  special_requirements?: string
  tracking_number?: string
}

export interface LegacySXGZQuoteResult {
  base_price: number
  company_name: string
  extra_options_price: number
  license_price: number
  price_multiplier: number
  print_price: number
  total_price: number
  user_rate: number
}

export interface LegacySXGZCreateResult {
  message: string
  need_refund?: boolean
  order_id: number
  order_no: string
  total_price: number
  upstream_id?: number
}

export interface LegacySXGZFilePrintOption {
  color_mode: string
  name: string
  page_count: number
  paper_size: string
  print_count: number
  print_mode: string
  size: number
  stamp_type: string
}

export interface LegacySXGZFileRecord {
  name: string
  page_count?: number
  print_options?: Omit<LegacySXGZFilePrintOption, 'name' | 'page_count' | 'size'>
  size: number
  storage?: string
  url: string
}

export interface LegacySXGZOrder {
  admin_notes?: string
  agent_order_id?: number
  base_price: number
  business_license: number
  company_id: number
  company_name: string
  completed_at?: string
  courier_company?: string
  created_at: string
  customer_address?: string
  customer_email?: string
  customer_name: string
  customer_phone?: string
  files?: {
    processed: LegacySXGZFileRecord[]
    uploaded: LegacySXGZFileRecord[]
  }
  file_size?: number
  license_price: number
  mail_price: number
  material_type?: string
  only_business_license: number
  order_id: number
  order_no: string
  original_filename?: string
  paper_size: string
  print_copies: number
  print_options?: string
  print_price: number
  processed_file_url?: string
  refund_reason?: string
  return_tracking_number?: string
  service_type: string
  source: string
  special_requirements?: string
  status: string
  total_price: number
  tracking_number?: string
  uid: number
  updated_at?: string
  uploaded_file?: string
}

export interface LegacySXGZAnnouncement {
  AID: number
  Content: string
  Importance: number
  PublishDate: string
  Title: string
}

export interface LegacySXGZAnnouncementListResult {
  data: LegacySXGZAnnouncement[]
  hasMore: boolean
  page: number
  pageSize: number
  total: number
  type: string
}

export function fetchLegacySXGZConfig() {
  return request.get<LegacySXGZConfig>({
    url: '/sxgz/config'
  })
}

export function saveLegacySXGZConfig(data: LegacySXGZConfig) {
  return request.post<void>({
    url: '/sxgz/config',
    params: data
  })
}

export function fetchLegacySXGZCompanies(params?: { search?: string }) {
  return request.get<LegacySXGZCompany[]>({
    url: '/sxgz/companies',
    params
  })
}

export function refreshLegacySXGZCompanies() {
  return request.post<LegacySXGZCompany[]>({
    url: '/sxgz/companies/refresh',
    params: {}
  })
}

export function fetchLegacySXGZLicenseCompanies(params?: { search?: string }) {
  return request.get<LegacySXGZCompany[]>({
    url: '/sxgz/license-companies',
    params
  })
}

export function fetchLegacySXGZAnnouncements(
  params: {
    page?: number
    pageSize?: number
    type?: string
  } = {}
) {
  return request.get<LegacySXGZAnnouncementListResult>({
    url: '/sxgz/announcements',
    params
  })
}

export function fetchLegacySXGZPrintOptions() {
  return request.get<{
    delivery_options: Record<string, LegacySXGZPricingRule>
    print_options: Record<string, LegacySXGZPricingRule>
    print_pricing: LegacySXGZPrintPricing
  }>({
    url: '/sxgz/print-options'
  })
}

export function quoteLegacySXGZOrder(data: LegacySXGZQuoteRequest) {
  return request.post<LegacySXGZQuoteResult>({
    url: '/sxgz/price/quote',
    params: data
  })
}

export function createLegacySXGZOrder(data: LegacySXGZQuoteRequest) {
  return request.post<LegacySXGZCreateResult>({
    url: '/sxgz/orders',
    params: data
  })
}

export function fetchLegacySXGZOrders(params: {
  page: number
  search?: string
  size: number
  status?: string
}) {
  return request.get<{ list: LegacySXGZOrder[]; total: number }>({
    url: '/sxgz/orders',
    params
  })
}

export function fetchLegacySXGZAdminOrders(params: {
  page: number
  search?: string
  size: number
  status?: string
}) {
  return request.get<{ list: LegacySXGZOrder[]; total: number }>({
    url: '/sxgz/admin/orders',
    params
  })
}

export function uploadLegacySXGZOrderFile(
  orderId: number,
  file: File,
  printOptions?: LegacySXGZFilePrintOption
) {
  const form = new FormData()
  form.append('file', file)
  if (printOptions) {
    form.append('file_print_options', JSON.stringify(printOptions))
    form.append('page_count', String(printOptions.page_count || 1))
  }
  return request.request<{
    file_name: string
    file_url: string
    size: number
    total_files: number
  }>({
    data: form,
    method: 'POST',
    url: `/sxgz/orders/${orderId}/files`
  })
}

export function applyLegacySXGZRefund(orderId: number, reason: string) {
  return request.post<void>({
    url: `/sxgz/orders/${orderId}/refund`,
    params: { reason }
  })
}

export function updateLegacySXGZAdminOrder(
  orderId: number,
  data: {
    admin_notes?: string
    refund_reason?: string
    status: string
  }
) {
  return request.request<void>({
    data,
    method: 'PATCH',
    url: `/sxgz/admin/orders/${orderId}`
  })
}

export function syncLegacySXGZOrders() {
  return request.post<{ updated: number }>({
    url: '/sxgz/admin/sync',
    params: {}
  })
}
