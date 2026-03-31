import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'lucide:activity',
      order: 3,
      title: '运动业务',
    },
    name: 'Sports',
    path: '/sports',
    redirect: '/sdxy/index',
    children: [
      {
        name: 'SDXYIndex',
        path: '/sdxy/index',
        component: () => import('#/views/plugins/sdxy/index.vue'),
        meta: {
          icon: 'lucide:zap',
          title: '闪电运动',
        },
      },
      {
        name: 'YDSJIndex',
        path: '/ydsj/index',
        component: () => import('#/views/plugins/ydsj/index.vue'),
        meta: {
          icon: 'lucide:footprints',
          title: '运动世界',
        },
      },
      {
        name: 'XMIndex',
        path: '/xm/index',
        component: () => import('#/views/plugins/xm/index.vue'),
        meta: {
          icon: 'lucide:activity',
          title: '小米运动',
        },
      },
      {
        name: 'WRunIndex',
        path: '/w/index',
        component: () => import('#/views/plugins/w/index.vue'),
        meta: {
          icon: 'lucide:waves',
          title: '鲸鱼运动',
        },
      },
      {
        name: 'YongyeIndex',
        path: '/yongye/index',
        component: () => import('#/views/plugins/yongye/index.vue'),
        meta: {
          icon: 'lucide:moon',
          title: '永夜运动',
        },
      },
    ],
  },
];

export default routes;
