import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:file-document-edit-outline',
      order: 5,
      title: '土拨鼠论文',
    },
    name: 'Tuboshu',
    path: '/tuboshu',
    redirect: '/tuboshu/index',
    children: [
      {
        name: 'TuboshuIndex',
        path: '/tuboshu/index',
        component: () => import('#/views/plugins/tuboshu/index.vue'),
        meta: {
          hideInMenu: true,
          title: '土拨鼠论文',
        },
      },
    ],
  },
];

export default routes;
