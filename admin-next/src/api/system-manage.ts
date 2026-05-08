import request from '@/utils/http'
import { adaptLegacyMenus, type LegacyRouteRecord } from '@/router/core/legacy-menu-adapter'
import type { LegacyMenuConfigItem } from '@/types/legacy-contract'
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

export function fetchGetMenuListWithConfigs() {
  const routeRequest = Promise.resolve(localLegacyRoutes as LegacyRouteRecord[])
  const configRequest = request
    .get<LegacyMenuConfigItem[]>({
      url: '/menus',
      showErrorMessage: false
    })
    .catch(() => [])

  return Promise.all([routeRequest, configRequest]).then(([routes, menuConfigs]) => ({
    routes: adaptLegacyMenus(routes, menuConfigs),
    menuConfigs
  }))
}

export function fetchGetMenuList() {
  return fetchGetMenuListWithConfigs().then(({ routes }) => routes)
}
