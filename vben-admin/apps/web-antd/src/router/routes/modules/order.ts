import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:book-education-outline',
      order: 1,
      title: '网课订单',
    },
    name: 'Order',
    path: '/order',
    redirect: '/order/list',
    children: [
      {
        name: 'OrderAdd',
        path: '/order/add',
        component: () => import('#/views/order/add.vue'),
        meta: {
          icon: 'mdi:file-search-outline',
          title: '查课交单',
        },
      },
      {
        name: 'OrderMobileAdd',
        path: '/order/mobile-add',
        component: () => import('#/views/order/mobile-add.vue'),
        meta: {
          icon: 'lucide:smartphone',
          title: '手机下单',
        },
      },
      {
        name: 'OrderBatchAdd',
        path: '/order/batch-add',
        component: () => import('#/views/order/batch-add.vue'),
        meta: {
          icon: 'mdi:file-multiple-outline',
          title: '批量交单',
        },
      },
      {
        name: 'OrderList',
        path: '/order/list',
        component: () => import('#/views/order/list.vue'),
        meta: {
          icon: 'lucide:clipboard-list',
          title: '订单汇总',
        },
      },
      {
        name: 'OrderQuality',
        path: '/order/quality',
        component: () => import('#/views/order/quality.vue'),
        meta: {
          icon: 'mdi:shield-check-outline',
          title: '质量查询',
        },
      },
    ],
  },
];

export default routes;
