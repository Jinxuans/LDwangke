import { requestClient } from '#/api/request';

export interface CheckinStatusResult {
  checked_in: boolean;
  reward_money?: number;
}

export interface CheckinResult {
  reward_money: number;
}

export interface CheckinRecord {
  uid: number;
  username: string;
  reward_money: number;
  addtime: string;
}

export interface CheckinStatsResult {
  total_users: number;
  total_reward: number;
  list: CheckinRecord[];
  total: number;
}

export function userCheckinApi() {
  return requestClient.post<CheckinResult>('/user/checkin');
}

export function userCheckinStatusApi() {
  return requestClient.get<CheckinStatusResult>('/user/checkin/status');
}

export function adminCheckinStatsApi(params: { date?: string; page?: number; limit?: number }) {
  return requestClient.get<CheckinStatsResult>('/admin/checkin/stats', { params });
}
