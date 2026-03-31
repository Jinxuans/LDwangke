import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'lucide:clipboard-check',
      order: 2,
      title: '打卡业务',
    },
    name: 'Checkin',
    path: '/checkin',
    redirect: '/yfdk/index',
    children: [
      {
        name: 'YFDKIndex',
        path: '/yfdk/index',
        component: () => import('#/views/plugins/yfdk/index.vue'),
        meta: {
          icon: 'lucide:clipboard-check',
          title: 'YF打卡',
        },
      },
      {
        name: 'SXDKIndex',
        path: '/sxdk/index',
        component: () => import('#/views/plugins/sxdk/index.vue'),
        meta: {
          icon: 'lucide:mountain',
          title: '泰山打卡',
        },
      },
      {
        name: 'AppuiIndex',
        path: '/appui/index',
        component: () => import('#/views/plugins/appui/index.vue'),
        meta: {
          icon: 'lucide:smartphone',
          title: 'Appui打卡',
        },
      },
      {
        name: 'TuZhiIndex',
        path: '/tuzhi/index',
        component: () => import('#/views/plugins/tuzhi/index.vue'),
        meta: {
          icon: 'lucide:map-pin-check',
          title: '凸知打卡',
        },
      },
      {
        name: 'SXZSIndex',
        path: '/sxzs/index',
        component: () => import('#/views/plugins/sxzs/index.vue'),
        meta: {
          icon: 'lucide:briefcase',
          title: '实习助手',
        },
      },
    ],
  },
];

export default routes;
