import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      authority: ['super', 'admin'],
      icon: 'ic:round-settings-input-composite',
      order: 20,
      title: '后台管理',
    },
    name: 'Admin',
    path: '/admin',
    redirect: '/admin/settings',
    children: [
      // ── 0. 菜单管理 ──
      {
        name: 'AdminMenus',
        path: '/admin/menus',
        component: () => import('#/views/admin/menus.vue'),
        meta: {
          icon: 'mdi:menu',
          order: 0,
          title: '菜单管理',
        },
      },

      // ── 1. 系统管理 ──
      {
        name: 'AdminSystem',
        path: '/admin/system',
        meta: {
          icon: 'mdi:cog-outline',
          order: 1,
          title: '系统管理',
        },
        redirect: '/admin/settings',
        children: [
          {
            name: 'AdminSettings',
            path: '/admin/settings',
            component: () => import('#/views/admin/settings.vue'),
            meta: { icon: 'mdi:tune-variant', order: 1, title: '系统设置' },
          },
          {
            name: 'AdminAnnouncements',
            path: '/admin/announcements',
            component: () => import('#/views/admin/announcements.vue'),
            meta: { icon: 'mdi:bullhorn-outline', order: 2, title: '公告管理' },
          },
          {
            name: 'AdminGrades',
            path: '/admin/grades',
            component: () => import('#/views/admin/grades.vue'),
            meta: { icon: 'mdi:medal-outline', order: 3, title: '等级管理' },
          },
        ],
      },

      // ── 2. 网课管理 ──
      {
        name: 'AdminCourse',
        path: '/admin/course',
        meta: {
          icon: 'mdi:book-open-page-variant-outline',
          order: 2,
          title: '网课管理',
        },
        redirect: '/admin/docking',
        children: [
          {
            name: 'AdminDocking',
            path: '/admin/docking',
            component: () => import('#/views/admin/docking.vue'),
            meta: { icon: 'mdi:connection', order: 1, title: '对接插件' },
          },
          {
            name: 'AdminSuppliers',
            path: '/admin/suppliers',
            component: () => import('#/views/admin/suppliers.vue'),
            meta: { icon: 'mdi:store-outline', order: 2, title: '货源管理' },
          },
          {
            name: 'AdminCategories',
            path: '/admin/categories',
            component: () => import('#/views/admin/categories.vue'),
            meta: { icon: 'mdi:file-tree-outline', order: 3, title: '分类管理' },
          },
          {
            name: 'AdminClass',
            path: '/admin/class',
            component: () => import('#/views/admin/class.vue'),
            meta: { icon: 'mdi:bookshelf', order: 4, title: '课程管理' },
          },
        ],
      },

      // ── 3. 运营管理 ──
      {
        name: 'AdminOps',
        path: '/admin/ops-group',
        meta: {
          icon: 'mdi:headset',
          order: 3,
          title: '运营管理',
        },
        redirect: '/admin/chat',
        children: [
          {
            name: 'AdminChat',
            path: '/admin/chat',
            component: () => import('#/views/admin/chat.vue'),
            meta: { icon: 'mdi:chat-outline', order: 1, title: '聊天管理' },
          },
          {
            name: 'AdminTickets',
            path: '/admin/tickets',
            component: () => import('#/views/admin/tickets.vue'),
            meta: { icon: 'mdi:ticket-outline', order: 2, title: '工单管理' },
          },
          {
            name: 'AdminCheckin',
            path: '/admin/checkin',
            component: () => import('#/views/admin/checkin.vue'),
            meta: { icon: 'mdi:calendar-check-outline', order: 3, title: '签到管理' },
          },
        ],
      },

      // ── 4. 邮箱管理 ──
      {
        name: 'AdminEmailGroup',
        path: '/admin/email',
        meta: {
          icon: 'mdi:email-multiple-outline',
          order: 4,
          title: '邮箱管理',
        },
        redirect: '/admin/email-pool',
        children: [
          {
            name: 'AdminEmailPool',
            path: '/admin/email-pool',
            component: () => import('#/views/admin/email-pool.vue'),
            meta: { icon: 'mdi:mailbox-outline', order: 1, title: '邮箱轮询池' },
          },
          {
            name: 'AdminMail',
            path: '/admin/mail',
            component: () => import('#/views/user/mail.vue'),
            meta: { icon: 'mdi:send-outline', order: 2, title: '邮件群发' },
          },
          {
            name: 'AdminEmailTemplates',
            path: '/admin/email-templates',
            component: () => import('#/views/admin/email-templates.vue'),
            meta: { icon: 'mdi:file-code-outline', order: 3, title: '邮件模板' },
          },
          {
            name: 'AdminEmailLogs',
            path: '/admin/email-logs',
            component: () => import('#/views/admin/email-logs.vue'),
            meta: { icon: 'mdi:file-document-outline', order: 4, title: '发送日志' },
          },
        ],
      },

      // ── 5. 上游对接 ──
      {
        name: 'AdminUpstream',
        path: '/admin/upstream',
        meta: {
          icon: 'mdi:api',
          order: 5,
          title: '上游对接',
        },
        redirect: '/admin/mijia',
        children: [
          {
            name: 'AdminMiJia',
            path: '/admin/mijia',
            component: () => import('#/views/admin/mijia.vue'),
            meta: { icon: 'mdi:lock-outline', order: 2, title: '密价设置' },
          },
          {
            name: 'AdminModules',
            path: '/admin/modules',
            component: () => import('#/views/admin/modules.vue'),
            meta: { icon: 'mdi:puzzle-outline', order: 3, title: '模块管理' },
          },
          {
            name: 'AdminSyncMonitor',
            path: '/admin/sync-monitor',
            component: () => import('#/views/admin/sync-monitor.vue'),
            meta: { icon: 'mdi:sync-circle', order: 4, title: '商品同步监控' },
          },
        ],
      },

      // ── 6. 数据统计 ──
      {
        name: 'AdminStats',
        path: '/admin/stats',
        meta: {
          icon: 'mdi:chart-bar',
          order: 6,
          title: '数据统计',
        },
        redirect: '/admin/rank-suppliers',
        children: [
          {
            name: 'AdminRankSuppliers',
            path: '/admin/rank-suppliers',
            component: () => import('#/views/admin/rank-suppliers.vue'),
            meta: { icon: 'mdi:trophy-outline', order: 1, title: '货源排行' },
          },
          {
            name: 'AdminRankAgentProducts',
            path: '/admin/rank-agent-products',
            component: () => import('#/views/admin/rank-agent-products.vue'),
            meta: { icon: 'mdi:chart-timeline-variant', order: 2, title: '代理统计' },
          },
          {
            name: 'AdminMoneyLog',
            path: '/admin/moneylog',
            component: () => import('#/views/admin/moneylog.vue'),
            meta: { icon: 'mdi:cash-register', order: 3, title: '全站流水' },
          },
        ],
      },

      // ── 7. 租户管理 ──
      {
        name: 'AdminTenants',
        path: '/admin/tenants',
        component: () => import('#/views/admin/tenants.vue'),
        meta: {
          icon: 'mdi:store-cog-outline',
          order: 7,
          title: '租户管理',
        },
      },

      // ── 8. 辅助业务 ──
      {
        name: 'AdminAuxGroup',
        path: '/admin/aux',
        meta: {
          icon: 'mdi:puzzle-plus-outline',
          order: 8,
          title: '辅助业务',
        },
        redirect: '/admin/cardkeys',
        children: [
          {
            name: 'AdminCardKeys',
            path: '/admin/cardkeys',
            component: () => import('#/views/admin/cardkeys.vue'),
            meta: { icon: 'mdi:credit-card-plus-outline', order: 1, title: '卡密管理' },
          },
          {
            name: 'AdminActivities',
            path: '/admin/activities',
            component: () => import('#/views/admin/activities.vue'),
            meta: { icon: 'mdi:gift-outline', order: 2, title: '活动管理' },
          },
          {
            name: 'AdminPledge',
            path: '/admin/pledge',
            component: () => import('#/views/admin/pledge.vue'),
            meta: { icon: 'mdi:shield-lock-outline', order: 3, title: '质押管理' },
          },
        ],
      },

      // ── 9. 运维看板 ──
      {
        name: 'AdminOpsDashboard',
        path: '/admin/ops',
        component: () => import('#/views/admin/ops-dashboard.vue'),
        meta: {
          icon: 'mdi:monitor-dashboard',
          order: 9,
          title: '运维看板',
        },
      },
    ],
  },
];

export default routes;
