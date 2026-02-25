import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      authority: ['super', 'admin'],
      icon: 'mdi:connection',
      order: 15,
      title: '对接中心',
    },
    name: 'Dock',
    path: '/dock',
    redirect: '/dock/index',
    children: [
      {
        name: 'DockIndex',
        path: '/dock/index',
        component: () => import('#/views/dock/index.vue'),
        meta: {
          icon: 'mdi:connection',
          title: '对接中心',
        },
      },
    ],
  },
];

export default routes;