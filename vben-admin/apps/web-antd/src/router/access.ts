import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
  MenuRoute,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';
import { getMenuConfigs } from '#/api/menu-config';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

async function generateAccess(options: GenerateMenuAndRoutesOptions) {
  const pageMap: ComponentRecordType = import.meta.glob('../views/**/*.vue');

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  return await generateAccessible(preferences.app.accessMode, {
    ...options,
    fetchMenuListAsync: async () => {
      message.loading({
        content: `${$t('common.loadingMenu')}...`,
        duration: 1.5,
      });
      const menus = await getAllMenusApi();
      
      const menuConfigs = await getMenuConfigs();
      const hiddenMenuKeys = new Set(
        menuConfigs.filter((item) => item.visible === 0).map((item) => item.menu_key)
      );
      
      const filterMenus = (menuList: MenuRoute[]): MenuRoute[] => {
        return menuList
          .filter((menu) => {
            if (menu.name && hiddenMenuKeys.has(menu.name)) {
              return false;
            }
            if (menu.children && menu.children.length > 0) {
              menu.children = filterMenus(menu.children);
            }
            return true;
          })
          .map((menu) => ({
            ...menu,
            children: menu.children ? filterMenus(menu.children) : menu.children,
          }));
      };
      
      return filterMenus(menus);
    },
    // 可以指定没有权限跳转403页面
    forbiddenComponent,
    // 如果 route.meta.menuVisibleWithForbidden = true
    layoutMap,
    pageMap,
  });
}

export { generateAccess };
