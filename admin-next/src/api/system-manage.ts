import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'
import { adaptLegacyMenus, type LegacyRouteRecord } from '@/router/core/legacy-menu-adapter'
import type { LegacyMenuConfigItem } from '@/types/legacy-contract'
import { isHttpError } from '@/utils/http/error'
import { localLegacyRoutes } from '@/router/core/legacy-local-routes'

// 获取用户列表
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/user/list',
    params
  })
}

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: '/api/role/list',
    params
  })
}

// 获取菜单列表
export function fetchGetMenuList() {
  const routeRequest = request
    .get<LegacyRouteRecord[]>({
      url: '/menu/all',
      showErrorMessage: false
    })
    .catch((error) => {
      if (isHttpError(error) && error.code === 404) {
        console.warn('[fetchGetMenuList] /menu/all 不存在，已切换到本地旧路由清单')
        return localLegacyRoutes
      }

      throw error
    })

  return Promise.all([
    routeRequest,
    request
      .get<LegacyMenuConfigItem[]>({
        url: '/menus',
        showErrorMessage: false
      })
      .catch(() => [])
  ]).then(([routes, menuConfigs]) => adaptLegacyMenus(routes, menuConfigs))
}
