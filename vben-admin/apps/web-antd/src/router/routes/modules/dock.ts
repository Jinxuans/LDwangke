import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:file-document-outline',
      order: 16,
      title: '接口文档',
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
          icon: 'mdi:file-document-outline',
          title: '接口文档',
        },
      },
    ],
  },
];

export default routes;