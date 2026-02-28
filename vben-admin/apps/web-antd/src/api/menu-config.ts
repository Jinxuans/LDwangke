import { requestClient } from '#/api/request';

export interface MenuConfigItem {
  id?: number;
  menu_key: string;
  parent_key: string;
  title: string;
  icon: string;
  sort_order: number;
  visible: number;
  scope: string;
}

/** 获取全部菜单配置 */
export function getMenuConfigs(): Promise<MenuConfigItem[]> {
  return requestClient.get('/menus');
}

/** 批量保存菜单配置（管理员） */
export function saveMenuConfigs(items: MenuConfigItem[]) {
  return requestClient.post('/admin/menus', { items });
}
