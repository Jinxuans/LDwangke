import { requestClient } from '#/api/request';

export interface PlatformConfig {
  id: number;
  pt: string;
  name: string;
  auth_type: string;
  api_path_style: string;
  success_codes: string;
  use_json: boolean;
  need_proxy: boolean;
  returns_yid: boolean;
  extra_params: boolean;
  query_act: string;
  query_path: string;
  query_param_style: string;
  query_polling: boolean;
  query_max_attempts: number;
  query_interval: number;
  query_response_map: string;
  order_act: string;
  order_path: string;
  yid_in_data_array: boolean;
  progress_act: string;
  progress_no_yid: string;
  progress_path: string;
  progress_method: string;
  progress_needs_auth: boolean;
  use_id_param: boolean;
  use_uuid_param: boolean;
  always_username: boolean;
  pause_act: string;
  pause_path: string;
  resume_act: string;
  resume_path: string;
  change_pass_act: string;
  change_pass_path: string;
  change_pass_param: string;
  change_pass_id_param: string;
  resubmit_path: string;
  log_act: string;
  log_path: string;
  log_method: string;
  log_id_param: string;
  balance_act: string;
  balance_path: string;
  balance_money_field: string;
  balance_method: string;
  balance_auth_type: string;
  report_param_style: string;
  report_auth_type: string;
  report_path: string;
  get_report_path: string;
  refresh_path: string;
  source_code: string;
  created_at: string;
  updated_at: string;
}

export interface ParsedPHPResult {
  pt: string;
  name: string;
  auth_type: string;
  api_path_style: string;
  success_codes: string;
  use_json: boolean;
  query_act: string;
  query_path: string;
  order_act: string;
  order_path: string;
  progress_act: string;
  progress_path: string;
  progress_method: string;
  pause_act: string;
  pause_path: string;
  change_pass_act: string;
  change_pass_path: string;
  change_pass_param: string;
  change_pass_id_param: string;
  log_act: string;
  log_path: string;
  log_id_param: string;
  returns_yid: boolean;
  balance_act: string;
  balance_path: string;
  balance_money_field: string;
  confidence: number;
  warnings: string[];
}

/** 获取所有平台配置 */
export async function getPlatformConfigsApi() {
  return requestClient.get<PlatformConfig[]>('/admin/platform-configs');
}

/** 保存平台配置 */
export async function savePlatformConfigApi(data: Partial<PlatformConfig>) {
  return requestClient.post('/admin/platform-config/save', data);
}

/** 删除平台配置 */
export async function deletePlatformConfigApi(pt: string) {
  return requestClient.delete(`/admin/platform-config/${pt}`);
}

/** 解析 PHP 代码 */
export async function parsePHPCodeApi(code: string) {
  return requestClient.post<ParsedPHPResult>('/admin/platform-config/parse-php', { code });
}

export interface DetectRequest {
  url: string;
  uid: string;
  key: string;
  token: string;
  cookie: string;
}

export interface ProbeDetail {
  endpoint: string;
  method: string;
  status: string;
  code: string;
  msg: string;
  raw_body: string;
}

export interface DetectResult {
  success: boolean;
  auth_type: string;
  success_code: string;
  api_style: string;
  balance_ok: boolean;
  balance_money: string;
  balance_act: string;
  balance_path: string;
  balance_field: string;
  query_ok: boolean;
  query_act: string;
  class_list_ok: boolean;
  category_ok: boolean;
  use_json: boolean;
  returns_yid: boolean;
  probes: ProbeDetail[];
  suggested_name: string;
  config: Record<string, string>;
}

/** 自动检测平台 */
export async function detectPlatformApi(data: DetectRequest) {
  return requestClient.post<DetectResult>('/admin/platform-config/detect', data);
}
