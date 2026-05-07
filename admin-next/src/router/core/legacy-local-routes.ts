import type { LegacyRouteRecord } from './legacy-menu-adapter'

function leaf(
  name: string,
  path: string,
  title: string,
  icon = '',
  authority?: string[],
  component = ''
): LegacyRouteRecord {
  return {
    name,
    path,
    component,
    meta: {
      title,
      icon,
      authority
    }
  }
}

function group(
  name: string,
  path: string,
  title: string,
  icon: string,
  redirect: string,
  children: LegacyRouteRecord[],
  authority?: string[]
): LegacyRouteRecord {
  return {
    name,
    path,
    component: 'BasicLayout',
    redirect,
    children,
    meta: {
      title,
      icon,
      authority
    }
  }
}

export const localLegacyRoutes: LegacyRouteRecord[] = [
  group('Dashboard', '/dashboard', '首页', 'lucide:layout-dashboard', '/dashboard/console', [
    leaf('Home', '/dashboard/console', '仪表盘', 'lucide:area-chart')
  ]),
  group('Order', '/order', '网课订单', 'mdi:book-education-outline', '/order/list', [
    leaf('OrderAdd', '/order/add', '查课交单', 'mdi:file-search-outline'),
    leaf('OrderMobileAdd', '/order/mobile-add', '手机下单', 'lucide:smartphone'),
    leaf('OrderBatchAdd', '/order/batch-add', '批量交单', 'mdi:file-multiple-outline'),
    leaf('OrderList', '/order/list', '订单汇总', 'lucide:clipboard-list'),
    leaf('OrderQuality', '/order/quality', '质量查询', 'mdi:shield-check-outline')
  ]),
  group('Checkin', '/checkin', '打卡业务', 'lucide:clipboard-check', '/yfdk/index', [
    leaf('YFDKIndex', '/yfdk/index', 'YF打卡', 'lucide:clipboard-check'),
    leaf('SXDKIndex', '/sxdk/index', '泰山打卡', 'lucide:mountain'),
    leaf('AppuiIndex', '/appui/index', 'Appui打卡', 'lucide:smartphone'),
    leaf('TuZhiIndex', '/tuzhi/index', '凸知打卡', 'lucide:map-pin-check')
  ]),
  group('InternshipSeal', '/sxgz', '实习盖章', 'lucide:stamp', '/sxgz/index', [
    leaf('SXGZIndex', '/sxgz/index', '迈巴赫平台', 'lucide:stamp'),
    leaf('SXZSIndex', '/sxzs/index', '实习助手', 'lucide:briefcase')
  ]),
  group('Sports', '/sports', '运动业务', 'lucide:activity', '/xm/index', [
    leaf('SDXYIndex', '/sdxy/index', '闪电运动', 'lucide:zap'),
    leaf('YDSJIndex', '/ydsj/index', '运动世界', 'lucide:footprints'),
    leaf('XMIndex', '/xm/index', '小米运动', 'lucide:activity'),
    leaf('WRunIndex', '/w/index', '鲸鱼运动', 'lucide:waves'),
    leaf('YongyeIndex', '/yongye/index', '永夜运动', 'lucide:moon')
  ]),
  group('TutuQG', '/tutuqg', '图图强国', 'mdi:star-shooting-outline', '/tutuqg/index', [
    leaf('TutuQGIndex', '/tutuqg/index', '图图强国')
  ]),
  group('Tuboshu', '/tuboshu', '土拨鼠论文', 'mdi:file-document-edit-outline', '/tuboshu/index', [
    leaf('TuboshuIndex', '/tuboshu/index', '土拨鼠论文')
  ]),
  group('Paper', '/paper', '智文论文', 'mdi:file-document-box', '/paper/index', [
    leaf('PaperIndex', '/paper/index', '智文论文')
  ]),
  group('Chat', '/chat', '在线客服', 'mdi:chat-processing-outline', '/chat/index', [
    leaf('ChatIndex', '/chat/index', '在线客服', 'mdi:headset')
  ]),
  group('User', '/user', '用户中心', 'mdi:account-circle-outline', '/user/profile', [
    leaf('UserProfile', '/user/profile', '我的资料', 'mdi:card-account-details-outline'),
    leaf('UserRecharge', '/user/recharge', '充值', 'mdi:wallet-plus-outline'),
    leaf('UserMoneyLog', '/user/moneylog', '余额流水', 'mdi:cash-multiple'),
    leaf('UserTicket', '/user/ticket', '工单', 'mdi:ticket-confirmation-outline'),
    leaf('UserLogs', '/user/logs', '操作日志', 'mdi:history'),
    leaf('UserCheckin', '/user/checkin', '每日签到', 'mdi:calendar-check'),
    leaf('UserActivities', '/user/activities', '活动中心', 'mdi:gift-outline'),
    leaf('UserPledge', '/user/pledge', '质押折扣', 'mdi:shield-lock-outline')
  ]),
  group('Tenant', '/tenant', '商城管理', 'mdi:store-outline', '/tenant/order-stats', [
    leaf('TenantOrderStats', '/tenant/order-stats', '订单统计', 'mdi:chart-bar'),
    leaf('TenantShop', '/tenant/shop', '店铺设置', 'mdi:storefront-outline'),
    leaf('TenantMallCategories', '/tenant/mall-categories', '商城分类', 'mdi:file-tree-outline'),
    leaf('TenantProducts', '/tenant/products', '选品管理', 'mdi:package-variant'),
    leaf('TenantCUsers', '/tenant/cusers', '会员管理', 'mdi:account-group-outline'),
    leaf('TenantMallOrders', '/tenant/mall-orders', '支付订单', 'mdi:receipt-text-outline'),
    leaf('TenantPayConfig', '/tenant/pay-config', '支付配置', 'mdi:credit-card-settings-outline'),
    leaf('TenantWithdraw', '/tenant/withdraw', '商城提现', 'mdi:cash-fast'),
    leaf('TenantCUserWithdraw', '/tenant/cuser-withdraw', '会员提现', 'mdi:cash-refund')
  ]),
  group('Agent', '/agent', '代理管理', 'mdi:account-supervisor-outline', '/agent/list', [
    leaf('AgentList', '/agent/list', '代理列表', 'mdi:account-group-outline')
  ]),
  group('Dock', '/dock', '接口文档', 'mdi:file-document-outline', '/dock/index', [
    leaf('DockIndex', '/dock/index', '接口文档', 'mdi:file-document-outline')
  ]),
  group(
    'Admin',
    '/admin',
    '后台管理',
    'ic:round-settings-input-composite',
    '/admin/settings',
    [
      leaf('AdminMenus', '/admin/menus', '菜单管理', 'mdi:menu', ['super', 'admin']),
      group(
        'AdminSystem',
        '/admin/system',
        '系统管理',
        'mdi:cog-outline',
        '/admin/settings',
        [
          leaf('AdminSettings', '/admin/settings', '系统设置', 'mdi:tune-variant', [
            'super',
            'admin'
          ]),
          leaf('AdminAnnouncements', '/admin/announcements', '公告管理', 'mdi:bullhorn-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminGrades', '/admin/grades', '等级管理', 'mdi:medal-outline', ['super', 'admin'])
        ],
        ['super', 'admin']
      ),
      group(
        'AdminCourse',
        '/admin/course',
        '网课管理',
        'mdi:book-open-page-variant-outline',
        '/admin/docking',
        [
          leaf('AdminDocking', '/admin/docking', '对接插件', 'mdi:connection', ['super', 'admin']),
          leaf('AdminSuppliers', '/admin/suppliers', '货源管理', 'mdi:store-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminCategories', '/admin/categories', '分类管理', 'mdi:file-tree-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminClass', '/admin/class', '课程管理', 'mdi:bookshelf', ['super', 'admin'])
        ],
        ['super', 'admin']
      ),
      group(
        'AdminOps',
        '/admin/ops-group',
        '运营管理',
        'mdi:headset',
        '/admin/chat',
        [
          leaf('AdminChat', '/admin/chat', '聊天管理', 'mdi:chat-outline', ['super', 'admin']),
          leaf('AdminTickets', '/admin/tickets', '工单管理', 'mdi:ticket-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminCheckin', '/admin/checkin', '签到管理', 'mdi:calendar-check-outline', [
            'super',
            'admin'
          ])
        ],
        ['super', 'admin']
      ),
      group(
        'AdminEmailGroup',
        '/admin/email',
        '邮箱管理',
        'mdi:email-multiple-outline',
        '/admin/email-pool',
        [
          leaf('AdminEmailPool', '/admin/email-pool', '邮箱轮询池', 'mdi:mailbox-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminMail', '/admin/mail', '邮件群发', 'mdi:send-outline', ['super', 'admin']),
          leaf(
            'AdminEmailTemplates',
            '/admin/email-templates',
            '邮件模板',
            'mdi:file-code-outline',
            ['super', 'admin']
          ),
          leaf('AdminEmailLogs', '/admin/email-logs', '发送日志', 'mdi:file-document-outline', [
            'super',
            'admin'
          ])
        ],
        ['super', 'admin']
      ),
      group(
        'AdminUpstream',
        '/admin/upstream',
        '上游对接',
        'mdi:api',
        '/admin/upstream-config',
        [
          leaf(
            'AdminUpstreamConfig',
            '/admin/upstream-config',
            '对接中心',
            'mdi:transit-connection-variant',
            ['super', 'admin']
          ),
          leaf('AdminPlatformConfig', '/admin/platform-config', '平台配置', 'mdi:api', [
            'super',
            'admin'
          ]),
          leaf('AdminMiJia', '/admin/mijia', '密价设置', 'mdi:lock-outline', ['super', 'admin']),
          leaf('AdminModules', '/admin/modules', '模块管理', 'mdi:puzzle-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminSyncMonitor', '/admin/sync-monitor', '商品同步监控', 'mdi:sync-circle', [
            'super',
            'admin'
          ]),
          leaf('AdminDockScheduler', '/admin/queue', '待对接调度', 'mdi:timer-sync-outline', [
            'super',
            'admin'
          ]),
          leaf(
            'AdminOrderProgressSync',
            '/admin/order-progress-sync',
            '主订单同步',
            'mdi:progress-clock',
            ['super', 'admin']
          )
        ],
        ['super', 'admin']
      ),
      group(
        'AdminStats',
        '/admin/stats',
        '数据统计',
        'mdi:chart-bar',
        '/admin/rank-suppliers',
        [
          leaf('AdminRankSuppliers', '/admin/rank-suppliers', '货源排行', 'mdi:trophy-outline', [
            'super',
            'admin'
          ]),
          leaf(
            'AdminRankAgentProducts',
            '/admin/rank-agent-products',
            '代理统计',
            'mdi:chart-timeline-variant',
            ['super', 'admin']
          ),
          leaf('AdminMoneyLog', '/admin/moneylog', '全站流水', 'mdi:cash-register', [
            'super',
            'admin'
          ]),
          leaf('AdminWithdraw', '/admin/withdraw', '商家提现审核', 'mdi:storefront-outline', [
            'super',
            'admin'
          ]),
          leaf(
            'AdminMallCUserWithdraw',
            '/admin/mall-cuser-withdraw',
            '会员佣金提现审核',
            'mdi:cash-refund',
            ['super', 'admin']
          )
        ],
        ['super', 'admin']
      ),
      leaf('AdminTenants', '/admin/tenants', '租户管理', 'mdi:store-cog-outline', [
        'super',
        'admin'
      ]),
      group(
        'AdminAuxGroup',
        '/admin/aux',
        '辅助业务',
        'mdi:puzzle-plus-outline',
        '/admin/cardkeys',
        [
          leaf('AdminCardKeys', '/admin/cardkeys', '卡密管理', 'mdi:credit-card-plus-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminActivities', '/admin/activities', '活动管理', 'mdi:gift-outline', [
            'super',
            'admin'
          ]),
          leaf('AdminPledge', '/admin/pledge', '质押管理', 'mdi:shield-lock-outline', [
            'super',
            'admin'
          ])
        ],
        ['super', 'admin']
      ),
      leaf('AdminOpsDashboard', '/admin/ops', '运维看板', 'mdi:monitor-dashboard', [
        'super',
        'admin'
      ]),
      {
        name: 'AdminDbCompat',
        path: '/admin/db-compat',
        component: '/admin/db-compat/index',
        meta: {
          title: '数据库工具',
          hideInMenu: true,
          authority: ['super', 'admin']
        }
      },
      {
        name: 'AdminPaperConfig',
        path: '/admin/paper-config',
        component: '/admin/paper-config/index',
        meta: {
          title: '智文论文配置',
          hideInMenu: true,
          authority: ['super', 'admin']
        }
      }
    ],
    ['super', 'admin']
  ),
  {
    name: 'ExtPageWrapper',
    path: '/admin/ext',
    component: 'BasicLayout',
    meta: {
      title: '扩展页面',
      hideInMenu: true,
      authority: ['super', 'admin']
    },
    children: [
      {
        name: 'AdminExtPage',
        path: '/admin/ext/:id',
        component: 'IFrameView',
        meta: {
          title: '扩展页面',
          hideInMenu: true,
          authority: ['super', 'admin']
        }
      }
    ]
  },
  {
    name: 'FrontExtPageWrapper',
    path: '/ext',
    component: 'BasicLayout',
    meta: {
      title: '扩展页面',
      hideInMenu: true
    },
    children: [
      {
        name: 'FrontExtPage',
        path: '/ext/:id',
        component: 'IFrameView',
        meta: {
          title: '扩展页面',
          hideInMenu: true
        }
      }
    ]
  }
]
