import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:file-document-box',
      order: 6,
      title: '智文论文',
    },
    name: 'Paper',
    path: '/paper',
    redirect: '/paper/index',
    children: [
      {
        name: 'PaperIndex',
        path: '/paper/index',
        component: () => import('#/views/plugins/paper/index.vue'),
        meta: {
          hideInMenu: true,
          title: '智文论文',
        },
      },
    ],
  },
];

export default routes;
