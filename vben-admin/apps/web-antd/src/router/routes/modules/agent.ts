import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:account-supervisor-outline',
      order: 14,
      title: '代理管理',
    },
    name: 'Agent',
    path: '/agent',
    redirect: '/agent/list',
    children: [
      {
        name: 'AgentList',
        path: '/agent/list',
        component: () => import('#/views/agent/list.vue'),
        meta: {
          icon: 'mdi:account-group-outline',
          title: '代理列表',
        },
      },
    ],
  },
];

export default routes;
