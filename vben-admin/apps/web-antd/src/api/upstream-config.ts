import { requestClient } from '#/api/request';

// ========== YF打卡 ==========
export interface YFDKConfig {
  base_url: string;
  token: string;
}

export async function yfdkConfigGetApi() {
  return requestClient.get<YFDKConfig>('/admin/yfdk/config');
}

export async function yfdkConfigSaveApi(data: YFDKConfig) {
  return requestClient.post('/admin/yfdk/config', data);
}

// ========== 泰山打卡 ==========
export interface SXDKConfig {
  base_url: string;
  token: string;
  admin: string;
}

export async function sxdkConfigGetApi() {
  return requestClient.get<SXDKConfig>('/admin/sxdk/config');
}

export async function sxdkConfigSaveApi(data: SXDKConfig) {
  return requestClient.post('/admin/sxdk/config', data);
}

// ========== 图图强国 ==========
export interface TutuQGUpstreamConfig {
  base_url: string;
  key: string;
  price_increment: number;
}

export async function tutuqgUpstreamConfigGetApi() {
  return requestClient.get<TutuQGUpstreamConfig>('/admin/tutuqg/config');
}

export async function tutuqgUpstreamConfigSaveApi(data: TutuQGUpstreamConfig) {
  return requestClient.post('/admin/tutuqg/config', data);
}

// ========== HZW实时进度Socket ==========
export interface HZWSocketConfig {
  socket_url: string;
}

export async function hzwSocketConfigGetApi() {
  return requestClient.get<HZWSocketConfig>('/admin/hzw-socket/config');
}

export async function hzwSocketConfigSaveApi(data: HZWSocketConfig) {
  return requestClient.post('/admin/hzw-socket/config', data);
}
