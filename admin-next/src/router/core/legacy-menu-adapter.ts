import type { AppRouteRecord } from '@/types/router'
import { RoutesAlias } from '@/router/routesAlias'
import type { LegacyMenuConfigItem } from '@/types/legacy-contract'

interface LegacyRouteMeta {
  activePath?: string
  affixTab?: boolean
  authority?: string[]
  hideInMenu?: boolean
  hideInTab?: boolean
  icon?: string
  iframeSrc?: string
  keepAlive?: boolean
  link?: string
  title?: string
}

export interface LegacyRouteRecord {
  children?: LegacyRouteRecord[]
  component?: string
  meta?: LegacyRouteMeta
  name?: string
  path?: string
  redirect?: string
}

const legacyRoleMap: Record<string, string> = {
  admin: 'R_ADMIN',
  super: 'R_SUPER',
  user: 'R_USER'
}

const migratedComponentMap: Record<string, string> = {
  Home: '/dashboard/console',
  DockIndex: '/dock/index',
  OrderAdd: '/order/add/index',
  OrderBatchAdd: '/order/batch-add/index',
  OrderMobileAdd: '/order/mobile-add/index',
  OrderList: '/order/list/index',
  OrderQuality: '/order/quality/index',
  YFDKIndex: '/plugins/yfdk/index',
  XMIndex: '/plugins/xm/index',
  SXDKIndex: '/plugins/sxdk/index',
  SDXYIndex: '/plugins/sdxy/index',
  ShashouIndex: '/plugins/shashou/index',
  WuxinIndex: '/plugins/wuxin/index',
  YDSJIndex: '/plugins/ydsj/index',
  WRunIndex: '/plugins/w/index',
  YongyeIndex: '/plugins/yongye/index',
  SXGZIndex: '/plugins/sxgz/mbh/index',
  TutuQGIndex: '/plugins/tutuqg/index',
  TuboshuIndex: '/plugins/tuboshu/index',
  PaperIndex: '/plugins/paper/index',
  AppuiIndex: '/plugins/appui/index',
  TuZhiIndex: '/plugins/tuzhi/index',
  AdminMenus: '/system/menu',
  AdminSettings: '/admin/settings/index',
  AdminAnnouncements: '/admin/announcements/index',
  AdminGrades: '/admin/grades/index',
  AdminDocking: '/admin/docking/index',
  AdminSuppliers: '/admin/suppliers/index',
  AdminCategories: '/admin/categories/index',
  AdminClass: '/admin/class/index',
  AdminCheckin: '/admin/checkin/index',
  UserProfile: '/user/profile/index',
  UserRecharge: '/user/recharge/index',
  UserMoneyLog: '/user/moneylog/index',
  UserLogs: '/user/logs/index',
  UserTicket: '/user/ticket/index',
  UserCheckin: '/user/checkin/index',
  UserActivities: '/user/activities/index',
  UserPledge: '/user/pledge/index',
  Chat: '/chat/index',
  ChatIndex: '/chat/index',
  AdminChat: '/admin/chat/index',
  AdminTickets: '/admin/tickets/index',
  AdminCardKeys: '/admin/cardkeys/index',
  AdminActivities: '/admin/activities/index',
  AdminPledge: '/admin/pledge/index',
  AdminEmailPool: '/admin/email-pool/index',
  AdminMail: '/admin/mail/index',
  AdminEmailTemplates: '/admin/email-templates/index',
  AdminEmailLogs: '/admin/email-logs/index',
  AdminRankSuppliers: '/admin/rank-suppliers/index',
  AdminRankAgentProducts: '/admin/rank-agent-products/index',
  AdminMoneyLog: '/admin/moneylog/index',
  AdminWithdraw: '/admin/withdraw/index',
  AdminMallCUserWithdraw: '/admin/mall-cuser-withdraw/index',
  AdminDashboard: '/admin/dashboard/index',
  AdminStats: '/admin/stats/index',
  AdminTenants: '/admin/tenants/index',
  AdminUpstreamConfig: '/admin/upstream-config/index',
  AdminShashou: '/admin/shashou/index',
  AdminPlatformConfig: '/admin/platform-config/index',
  AdminPaperConfig: '/admin/paper-config/index',
  AdminSyncMonitor: '/admin/sync-monitor/index',
  AdminDockScheduler: '/admin/queue/index',
  AdminOrderProgressSync: '/admin/order-progress-sync/index',
  AdminModules: '/admin/modules/index',
  AdminMiJia: '/admin/mijia/index',
  AdminOpsDashboard: '/admin/ops/index',
  SportHub: '/module/hub/index',
  SportDetail: '/module/frame/index',
  InternHub: '/module/hub/index',
  InternDetail: '/module/frame/index',
  PaperHub: '/module/hub/index',
  PaperDetail: '/module/frame/index',
  AgentList: '/agent/list/index',
  TenantOrderStats: '/tenant/order-stats/index',
  TenantShop: '/tenant/shop/index',
  TenantMallCategories: '/tenant/mall-categories/index',
  TenantPayConfig: '/tenant/pay-config/index',
  TenantProducts: '/tenant/products/index',
  TenantCUsers: '/tenant/cusers/index',
  TenantMallOrders: '/tenant/mall-orders/index',
  TenantWithdraw: '/tenant/withdraw/index',
  TenantCUserWithdraw: '/tenant/cuser-withdraw/index'
}

export function adaptLegacyMenus(
  routes: LegacyRouteRecord[],
  menuConfigs: LegacyMenuConfigItem[] = []
): AppRouteRecord[] {
  const configMap = new Map(menuConfigs.map((item) => [item.menu_key, item]))

  return routes
    .map((route) => adaptLegacyRoute(route, configMap, 0))
    .filter(Boolean)
    .sort(sortRoutesByOrder) as AppRouteRecord[]
}

function adaptLegacyRoute(
  route: LegacyRouteRecord,
  configMap: Map<string, LegacyMenuConfigItem>,
  depth = 0
): AppRouteRecord {
  const routeName = String(route.name || route.path || 'LegacyRoute')
  const meta = route.meta || {}
  const iframeLink = meta.iframeSrc || meta.link
  const menuConfig = configMap.get(routeName)
  const children =
    route.children
      ?.map((item) => adaptLegacyRoute(item, configMap, depth + 1))
      .filter(Boolean)
      .sort(sortRoutesByOrder) || []

  return {
    path: route.path || '/',
    name: route.name || routeName,
    redirect: route.redirect,
    component: resolveComponent(route, depth),
    children,
    meta: {
      title: menuConfig?.title || meta.title || routeName,
      order: menuConfig?.sort_order,
      icon: menuConfig?.icon || meta.icon,
      isHide: menuConfig ? menuConfig.visible === 0 : meta.hideInMenu,
      isHideTab: meta.hideInTab,
      keepAlive: meta.keepAlive,
      fixedTab: meta.affixTab,
      activePath: meta.activePath,
      isIframe: Boolean(iframeLink),
      link: iframeLink,
      roles: mapLegacyRoles(meta.authority)
    }
  }
}

function sortRoutesByOrder(a: AppRouteRecord, b: AppRouteRecord) {
  return Number(a.meta?.order || 0) - Number(b.meta?.order || 0)
}

function resolveComponent(route: LegacyRouteRecord, depth = 0) {
  const routeName = String(route.name || '')
  const component = route.component || ''
  const iframeLink = route.meta?.iframeSrc || route.meta?.link

  if (iframeLink || component === 'IFrameView') {
    return '/outside/Iframe'
  }

  if (component === 'BasicLayout') {
    return depth === 0 ? RoutesAlias.Layout : ''
  }

  if (migratedComponentMap[routeName]) {
    return migratedComponentMap[routeName]
  }

  const normalized = normalizeLegacyViewPath(component)
  if (normalized) {
    return normalized
  }

  if (route.children?.length) {
    return ''
  }

  return '/compat/legacy-placeholder'
}

function normalizeLegacyViewPath(component: string) {
  const cleaned = component
    .replace(/^(\.\/|\.\.\/)+/, '')
    .replace(/^#\//, '')
    .replace(/^src\//, '')
    .replace(/^views\//, '')
    .replace(/\.vue$/, '')

  if (!cleaned || cleaned === 'BasicLayout' || cleaned === 'IFrameView') {
    return ''
  }

  return cleaned.startsWith('/') ? cleaned : `/${cleaned}`
}

function mapLegacyRoles(authority: string[] = []) {
  if (!authority.length) {
    return undefined
  }

  return authority.map((role) => legacyRoleMap[role] || role)
}
