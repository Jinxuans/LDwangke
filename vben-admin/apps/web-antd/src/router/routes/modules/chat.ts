import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:chat-processing-outline',
      order: 5,
      title: '在线客服',
    },
    name: 'Chat',
    path: '/chat',
    children: [
      {
        name: 'ChatIndex',
        path: '/chat',
        component: () => import('#/views/chat/index.vue'),
        meta: {
          icon: 'mdi:headset',
          title: '在线客服',
        },
      },
    ],
  },
];

export default routes;
