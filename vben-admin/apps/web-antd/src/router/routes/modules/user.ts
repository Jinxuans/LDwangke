import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:account-circle-outline',
      order: 10,
      title: '用户中心',
    },
    name: 'User',
    path: '/user',
    redirect: '/user/profile',
    children: [
      {
        name: 'UserProfile',
        path: '/user/profile',
        component: () => import('#/views/user/profile.vue'),
        meta: {
          icon: 'mdi:card-account-details-outline',
          title: '我的资料',
        },
      },
      {
        name: 'UserRecharge',
        path: '/user/recharge',
        component: () => import('#/views/user/recharge.vue'),
        meta: {
          icon: 'mdi:wallet-plus-outline',
          title: '充值',
        },
      },
      {
        name: 'UserMoneyLog',
        path: '/user/moneylog',
        component: () => import('#/views/user/moneylog.vue'),
        meta: {
          icon: 'mdi:cash-multiple',
          title: '余额流水',
        },
      },
      {
        name: 'UserTicket',
        path: '/user/ticket',
        component: () => import('#/views/user/ticket.vue'),
        meta: {
          icon: 'mdi:ticket-confirmation-outline',
          title: '工单',
        },
      },
      {
        name: 'UserLogs',
        path: '/user/logs',
        component: () => import('#/views/user/logs.vue'),
        meta: {
          icon: 'mdi:history',
          title: '操作日志',
        },
      },
      {
        name: 'UserCheckin',
        path: '/user/checkin',
        component: () => import('#/views/user/checkin.vue'),
        meta: {
          icon: 'mdi:calendar-check',
          title: '每日签到',
        },
      },
    ],
  },
];

export default routes;
