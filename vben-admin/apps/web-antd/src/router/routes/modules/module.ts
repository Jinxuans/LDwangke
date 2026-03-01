import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

/** 模块路由工厂：每个模块类型生成 hub + frame 两个子路由 */
function makeModuleRoute(
  name: string,
  path: string,
  icon: string,
  order: number,
  hubTitle: string,
  detailTitle: string,
): RouteRecordRaw {
  return {
    component: BasicLayout,
    meta: { icon, order, title: hubTitle },
    name: `${name}Module`,
    path,
    redirect: `${path}/hub`,
    children: [
      {
        name: `${name}Hub`,
        path: `${path}/hub`,
        component: () => import('#/views/module/hub.vue'),
        meta: { icon: 'carbon:workspace', title: hubTitle },
      },
      {
        name: `${name}Detail`,
        path: `${path}/:appId`,
        component: () => import('#/views/module/frame.vue'),
        meta: { title: detailTitle, hideInMenu: true },
      },
    ],
  };
}

const routes: RouteRecordRaw[] = [];

export default routes;
