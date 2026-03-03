import { requestClient } from '#/api/request';

export interface ExtMenuItem {
  id: number;
  title: string;
  icon: string;
  url: string;
  sort_order: number;
  visible: number;
  scope: string;
  created_at: string;
}

/** 获取可见的扩展菜单（公开接口） */
export function getExtMenusPublicApi(): Promise<ExtMenuItem[]> {
  return requestClient.get('/ext-menus');
}

/** 管理员：获取所有扩展菜单 */
export function getExtMenusApi(): Promise<ExtMenuItem[]> {
  return requestClient.get('/admin/ext-menus');
}

/** 管理员：保存扩展菜单 */
export function saveExtMenuApi(data: Partial<ExtMenuItem>) {
  return requestClient.post('/admin/ext-menu/save', data);
}

/** 管理员：批量更新扩展菜单排序 */
export function reorderExtMenusApi(items: { id: number; sort_order: number }[]) {
  return requestClient.post('/admin/ext-menu/reorder', { items });
}

/** 管理员：删除扩展菜单 */
export function deleteExtMenuApi(id: number) {
  return requestClient.delete(`/admin/ext-menu/${id}`);
}
