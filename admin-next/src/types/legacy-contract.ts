export type LegacyAdminConfigMap = Record<string, string>

export interface LegacyMenuConfigItem {
  id?: number
  menu_key: string
  parent_key: string
  title: string
  icon: string
  sort_order: number
  visible: number
  scope: string
}

export interface LegacyExtMenuItem {
  id: number
  title: string
  icon: string
  url: string
  sort_order: number
  visible: number
  scope: string
  created_at: string
}

export interface LegacySiteConfig {
  sitename?: string
  logo?: string
  hlogo?: string
  notice?: string
  qd_notice?: string
  bz?: string
  description?: string
  keywords?: string
  version?: string
  pass2_kg?: string
  sykg?: string
  mall_domain_suffix?: string
  mall_open_price?: string
  onlineStore_trdltz?: string
  top_consumers_open?: string
  fllx?: string
  recommend_channels?: string
  recharge_bonus_rules?: string
  checkin_enabled?: string
  checkin_order_required?: string
  checkin_min_balance?: string
  checkin_max_users?: string
  checkin_min_reward?: string
  checkin_max_reward?: string
  zdpay?: string
  non_direct_recharge_enable?: string
  user_yqzc?: string
  user_htkh?: string
  user_ktmoney?: string
  dl_pkkg?: string
  djfl?: string
  cross_recharge_uids?: string
  [key: string]: string | undefined
}
