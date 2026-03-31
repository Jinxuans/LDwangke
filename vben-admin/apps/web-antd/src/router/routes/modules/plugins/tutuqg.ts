import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:star-shooting-outline',
      order: 4,
      title: '图图强国',
    },
    name: 'TutuQG',
    path: '/tutuqg',
    redirect: '/tutuqg/index',
    children: [
      {
        name: 'TutuQGIndex',
        path: '/tutuqg/index',
        component: () => import('#/views/plugins/tutuqg/index.vue'),
        meta: {
          hideInMenu: true,
          title: '图图强国',
        },
      },
    ],
  },
];

export default routes;
