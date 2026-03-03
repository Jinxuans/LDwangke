import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      authority: ['super', 'admin'],
      hideInMenu: true,
      hideInBreadcrumb: true,
      order: 999,
      title: '扩展页面',
    },
    name: 'ExtPageWrapper',
    path: '/admin/ext',
    children: [
      {
        name: 'AdminExtPage',
        path: '/admin/ext/:id',
        component: () => import('#/views/ext/IframePage.vue'),
        meta: {
          title: '扩展页面',
          hideInMenu: true,
        },
      },
    ],
  },
  {
    component: BasicLayout,
    meta: {
      hideInMenu: true,
      hideInBreadcrumb: true,
      order: 998,
      title: '扩展页面',
    },
    name: 'FrontExtPageWrapper',
    path: '/ext',
    children: [
      {
        name: 'FrontExtPage',
        path: '/ext/:id',
        component: () => import('#/views/ext/IframePage.vue'),
        meta: {
          title: '扩展页面',
          hideInMenu: true,
        },
      },
    ],
  },
];

export default routes;
