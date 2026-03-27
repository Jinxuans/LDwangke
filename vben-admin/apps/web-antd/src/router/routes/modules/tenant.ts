import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:store-outline',
      order: 13,
      title: '商城管理',
    },
    name: 'Tenant',
    path: '/tenant',
    redirect: '/tenant/order-stats',
    children: [
      {
        name: 'TenantOrderStats',
        path: '/tenant/order-stats',
        component: () => import('#/views/tenant/order-stats.vue'),
        meta: {
          icon: 'mdi:chart-bar',
          order: 1,
          title: '订单统计',
        },
      },
      {
        name: 'TenantShop',
        path: '/tenant/shop',
        component: () => import('#/views/tenant/shop.vue'),
        meta: {
          icon: 'mdi:storefront-outline',
          order: 2,
          title: '店铺设置',
        },
      },
      {
        name: 'TenantMallCategories',
        path: '/tenant/mall-categories',
        component: () => import('#/views/tenant/mall-categories.vue'),
        meta: {
          icon: 'mdi:file-tree-outline',
          order: 3,
          title: '商城分类',
        },
      },
      {
        name: 'TenantProducts',
        path: '/tenant/products',
        component: () => import('#/views/tenant/products.vue'),
        meta: {
          icon: 'mdi:package-variant',
          order: 4,
          title: '选品管理',
        },
      },
      {
        name: 'TenantCUsers',
        path: '/tenant/cusers',
        component: () => import('#/views/tenant/cusers.vue'),
        meta: {
          icon: 'mdi:account-group-outline',
          order: 5,
          title: '会员管理',
        },
      },
      {
        name: 'TenantMallOrders',
        path: '/tenant/mall-orders',
        component: () => import('#/views/tenant/mall-orders.vue'),
        meta: {
          icon: 'mdi:receipt-text-outline',
          order: 6,
          title: '支付订单',
        },
      },
      {
        name: 'TenantPayConfig',
        path: '/tenant/pay-config',
        component: () => import('#/views/tenant/pay-config.vue'),
        meta: {
          icon: 'mdi:credit-card-settings-outline',
          order: 7,
          title: '支付配置',
        },
      },
      {
        name: 'TenantWithdraw',
        path: '/tenant/withdraw',
        component: () => import('#/views/user/withdraw.vue'),
        meta: {
          icon: 'mdi:cash-fast',
          order: 8,
          title: '商城提现',
        },
      },
      {
        name: 'TenantCUserWithdraw',
        path: '/tenant/cuser-withdraw',
        component: () => import('#/views/tenant/cuser-withdraw.vue'),
        meta: {
          icon: 'mdi:cash-refund',
          order: 9,
          title: '会员提现',
        },
      },
    ],
  },
];

export default routes;
