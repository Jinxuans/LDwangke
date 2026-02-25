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

const routes: RouteRecordRaw[] = [
  makeModuleRoute('Sport', '/sport', 'mdi:run', 5, '运动跑步', '运动详情'),
  makeModuleRoute('Intern', '/intern', 'mdi:briefcase-clock-outline', 6, '实习打卡', '实习详情'),
  makeModuleRoute('Paper', '/paper', 'mdi:file-document-edit-outline', 7, '论文撰写', '论文详情'),
];

export default routes;
