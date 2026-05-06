import request from '@/utils/http'

function mapLegacyRole(role: string) {
  switch (role) {
    case 'super':
      return 'R_SUPER'
    case 'admin':
      return 'R_ADMIN'
    default:
      return 'R_USER'
  }
}

function normalizeLegacyUserInfo(raw: Record<string, any>): Api.Auth.UserInfo {
  const username = raw.username || raw.user || ''
  const realName = raw.realName || raw.name || username
  const roles = Array.isArray(raw.roles) ? raw.roles.map(mapLegacyRole) : ['R_USER']

  return {
    buttons: Array.isArray(raw.buttons) ? raw.buttons : [],
    roles,
    userId: Number(raw.userId || raw.uid || 0),
    userName: realName || username,
    username,
    realName,
    email: raw.email || '',
    avatar: raw.avatar || '',
    homePath: raw.homePath || '/',
    desc: raw.desc || ''
  }
}

/**
 * 登录
 * @param params 登录参数
 * @returns 登录响应
 */
export function fetchLogin(params: Api.Auth.LoginParams) {
  return request.post<Api.Auth.LoginResponse>({
    url: '/auth/login',
    params: {
      username: params.username || params.userName,
      password: params.password,
      pass2: params.pass2
    }
    // showSuccessMessage: true // 显示成功消息
    // showErrorMessage: false // 不显示错误消息
  })
}

/**
 * 获取用户信息
 * @returns 用户信息
 */
export function fetchGetUserInfo() {
  return request
    .get<Record<string, any>>({
      url: '/user/info'
    // 自定义请求头
    // headers: {
    //   'X-Custom-Header': 'your-custom-value'
    // }
    })
    .then(normalizeLegacyUserInfo)
}
