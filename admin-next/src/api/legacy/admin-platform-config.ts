import request from '@/utils/http'

export interface LegacyPlatformConfig {
  id: number
  pt: string
  name: string
  auth_type: string
  success_codes: string
  use_json: boolean
  need_proxy: boolean
  returns_yid: boolean
  extra_params: boolean
  query_act: string
  query_path: string
  query_method: string
  query_body_type: string
  query_param_style: string
  query_param_map: string
  query_polling: boolean
  query_max_attempts: number
  query_interval: number
  query_response_map: string
  order_path: string
  order_method: string
  order_body_type: string
  order_param_map: string
  yid_in_data_array: boolean
  progress_path: string
  progress_method: string
  progress_body_type: string
  progress_param_map: string
  batch_progress_path: string
  batch_progress_method: string
  batch_progress_body_type: string
  batch_progress_param_map: string
  category_path: string
  category_method: string
  category_body_type: string
  category_param_map: string
  class_list_path: string
  class_list_method: string
  class_list_body_type: string
  class_list_param_map: string
  pause_path: string
  pause_method: string
  pause_body_type: string
  pause_param_map: string
  pause_id_param: string
  resume_path: string
  resume_method: string
  resume_body_type: string
  resume_param_map: string
  change_pass_path: string
  change_pass_method: string
  change_pass_body_type: string
  change_pass_param_map: string
  change_pass_param: string
  change_pass_id_param: string
  resubmit_path: string
  resubmit_method: string
  resubmit_body_type: string
  resubmit_param_map: string
  resubmit_id_param: string
  log_path: string
  log_method: string
  log_body_type: string
  log_param_map: string
  log_id_param: string
  balance_path: string
  balance_money_field: string
  balance_method: string
  balance_body_type: string
  balance_param_map: string
  balance_auth_type: string
  report_param_style: string
  report_auth_type: string
  report_path: string
  report_method: string
  report_body_type: string
  report_param_map: string
  get_report_path: string
  get_report_method: string
  get_report_body_type: string
  get_report_param_map: string
  refresh_path: string
  source_code: string
  created_at: string
  updated_at: string
}

export interface LegacyParsedPHPResult {
  pt: string
  name: string
  auth_type: string
  api_path_style: string
  success_codes: string
  use_json: boolean
  query_act: string
  query_path: string
  order_act: string
  order_path: string
  progress_act: string
  progress_path: string
  progress_method: string
  pause_act: string
  pause_path: string
  pause_id_param?: string
  change_pass_act: string
  change_pass_path: string
  change_pass_param: string
  change_pass_id_param: string
  log_act: string
  log_path: string
  log_id_param: string
  returns_yid: boolean
  balance_act: string
  balance_path: string
  balance_money_field: string
  confidence: number
  warnings: string[]
}

export interface LegacyPlatformDetectRequest {
  url: string
  uid: string
  key: string
  token: string
  cookie: string
}

export interface LegacyPlatformProbeDetail {
  endpoint: string
  method: string
  status: string
  code: string
  msg: string
  raw_body: string
}

export interface LegacyPlatformDetectResult {
  success: boolean
  auth_type: string
  success_code: string
  api_style: string
  balance_ok: boolean
  balance_money: string
  balance_act: string
  balance_path: string
  balance_field: string
  query_ok: boolean
  query_act: string
  class_list_ok: boolean
  category_ok: boolean
  use_json: boolean
  returns_yid: boolean
  probes: LegacyPlatformProbeDetail[]
  suggested_name: string
  config: Record<string, string>
}

export interface LegacyPlatformAutoDetectSaveRequest extends LegacyPlatformDetectRequest {
  pt: string
  name: string
}

export interface LegacyPlatformAutoDetectSaveResult {
  success: boolean
  msg: string
  detect: LegacyPlatformDetectResult
  config: Partial<LegacyPlatformConfig>
}

export function fetchLegacyPlatformConfigs() {
  return request.get<LegacyPlatformConfig[]>({
    url: '/admin/platform-configs'
  })
}

export function saveLegacyPlatformConfig(data: Partial<LegacyPlatformConfig>) {
  return request.post<void>({
    url: '/admin/platform-config/save',
    params: data
  })
}

export function deleteLegacyPlatformConfig(pt: string) {
  return request.del<void>({
    url: `/admin/platform-config/${pt}`
  })
}

export function parseLegacyPlatformPHPCode(code: string) {
  return request.post<LegacyParsedPHPResult>({
    url: '/admin/platform-config/parse-php',
    params: { code }
  })
}

export function detectLegacyPlatform(data: LegacyPlatformDetectRequest) {
  return request.post<LegacyPlatformDetectResult>({
    url: '/admin/platform-config/detect',
    params: data
  })
}

export function autoDetectSaveLegacyPlatform(data: LegacyPlatformAutoDetectSaveRequest) {
  return request.post<LegacyPlatformAutoDetectSaveResult>({
    url: '/admin/platform-config/auto-detect-save',
    params: data
  })
}
