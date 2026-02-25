import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
    pass2?: string;
  }

  /** 注册接口参数 */
  export interface RegisterParams {
    user: string;
    pass: string;
    name: string;
    yqm: string;
    email?: string;
    verify_code?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/auth/login', data);
}

/**
 * 注册
 */
export async function registerApi(data: AuthApi.RegisterParams) {
  return baseRequestClient.post('/auth/register', data);
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  return baseRequestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh-token', {}, {
    withCredentials: true,
  });
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.post('/auth/logout', {}, {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
  return requestClient.get<string[]>('/auth/codes');
}

/**
 * 发送验证码（注册用）
 */
export async function sendCodeApi(email: string, purpose: string) {
  return baseRequestClient.post('/auth/send-code', { email, purpose });
}

/**
 * 忘记密码 - 发送验证码
 */
export async function forgotPasswordApi(email: string) {
  return baseRequestClient.post('/auth/forgot-password', { email });
}

/**
 * 忘记密码 - 重置密码
 */
export async function resetPasswordApi(email: string, code: string, password: string) {
  return baseRequestClient.post('/auth/reset-password', { email, code, password });
}
