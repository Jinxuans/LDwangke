import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'lucide:layout-dashboard',
      order: -1,
      title: '首页',
    },
    name: 'Dashboard',
    path: '/',
    redirect: '/dashboard',
    children: [
      {
        name: 'Home',
        path: '/dashboard',
        component: () => import('#/views/dashboard/home/index.vue'),
        meta: {
          affixTab: true,
          icon: 'lucide:area-chart',
          title: '仪表盘',
        },
      },
    ],
  },
];

export default routes;
