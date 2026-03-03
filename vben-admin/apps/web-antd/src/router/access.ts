import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
} from '@vben/types';

import type { RouteRecordRaw } from 'vue-router';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';
import { getMenuConfigs } from '#/api/menu-config';
import type { MenuConfigItem } from '#/api/menu-config';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

/**
 * 递归地将菜单配置（隐藏、标题、图标、排序）应用到路由的 meta 上，
 * 这样 generateAccessible 内部的 generateMenus 会自动根据 hideInMenu 过滤
 */
function applyMenuConfigsToRoutes(
  routes: RouteRecordRaw[],
  configMap: Map<string, MenuConfigItem>,
) {
  for (const route of routes) {
    const cfg = route.name ? configMap.get(route.name as string) : undefined;
    if (cfg) {
      if (!route.meta) route.meta = {};
      route.meta.hideInMenu = cfg.visible === 0;
      if (cfg.title) {
        route.meta.title = cfg.title;
      }
      if (cfg.icon) {
        route.meta.icon = cfg.icon;
      }
      route.meta.order = cfg.sort_order;
    }
    if (route.children && route.children.length > 0) {
      applyMenuConfigsToRoutes(route.children, configMap);
    }
  }
}

async function generateAccess(options: GenerateMenuAndRoutesOptions) {
  const pageMap: ComponentRecordType = import.meta.glob('../views/**/*.vue');

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  // 在生成路由和菜单之前，先拉取菜单配置并应用到路由 meta 上
  try {
    const menuConfigs = await getMenuConfigs();
    if (menuConfigs && menuConfigs.length > 0) {
      const configMap = new Map(menuConfigs.map((c) => [c.menu_key, c]));
      applyMenuConfigsToRoutes(options.routes, configMap);
    }
  } catch (e) {
    console.warn('加载菜单配置失败，使用默认菜单', e);
  }

  return await generateAccessible(preferences.app.accessMode, {
    ...options,
    fetchMenuListAsync: async () => {
      message.loading({
        content: `${$t('common.loadingMenu')}...`,
        duration: 1.5,
      });
      return await getAllMenusApi();
    },
    forbiddenComponent,
    layoutMap,
    pageMap,
  });
}

export { generateAccess };
