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
        component: () => import('#/views/order/yfdk.vue'),
        meta: {
          icon: 'lucide:clipboard-check',
          title: 'YF打卡',
        },
      },
      {
        name: 'SXDKIndex',
        path: '/sxdk/index',
        component: () => import('#/views/order/sxdk.vue'),
        meta: {
          icon: 'lucide:mountain',
          title: '泰山打卡',
        },
      },
      {
        name: 'AppuiIndex',
        path: '/appui/index',
        component: () => import('#/views/order/appui.vue'),
        meta: {
          icon: 'lucide:smartphone',
          title: 'Appui打卡',
        },
      },
    ],
  },
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
        component: () => import('#/views/order/sdxy.vue'),
        meta: {
          icon: 'lucide:zap',
          title: '闪电运动',
        },
      },
      {
        name: 'YDSJIndex',
        path: '/ydsj/index',
        component: () => import('#/views/order/ydsj.vue'),
        meta: {
          icon: 'lucide:footprints',
          title: '运动世界',
        },
      },
      {
        name: 'XMIndex',
        path: '/xm/index',
        component: () => import('#/views/order/xm.vue'),
        meta: {
          icon: 'lucide:activity',
          title: '小米运动',
        },
      },
      {
        name: 'WRunIndex',
        path: '/w/index',
        component: () => import('#/views/order/w.vue'),
        meta: {
          icon: 'lucide:waves',
          title: '鲸鱼运动',
        },
      },
    ],
  },
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
        component: () => import('#/views/order/tutuqg.vue'),
        meta: {
          hideInMenu: true,
          title: '图图强国',
        },
      },
    ],
  },
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
        component: () => import('#/views/order/tuboshu.vue'),
        meta: {
          hideInMenu: true,
          title: '土拨鼠论文',
        },
      },
    ],
  },
];

export default routes;
