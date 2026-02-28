import type { Recordable, UserInfo } from '@vben/types';

import { h, ref } from 'vue';
import { useRouter } from 'vue-router';

import { DEFAULT_HOME_PATH, LOGIN_PATH } from '@vben/constants';
import { updatePreferences } from '@vben/preferences';
import { resetAllStores, useAccessStore, useUserStore } from '@vben/stores';

import { notification, Modal, Input } from 'ant-design-vue';
import { defineStore } from 'pinia';

import { getAccessCodesApi, getUserInfoApi, loginApi, logoutApi, registerApi } from '#/api';
import { getSiteConfigApi } from '#/api/admin';
import { $t } from '#/locales';

function showPass2Modal(): Promise<string | null> {
  return new Promise((resolve) => {
    let pass2Value = '';
    Modal.confirm({
      title: '管理员二级密码验证',
      icon: null,
      width: 'min(90vw, 360px)',
      centered: true,
      content: h('div', { style: 'margin-top:12px' }, [
        h('p', { style: 'color:#666;font-size:13px;margin-bottom:12px' }, '检测到管理员账号，请输入二级密码继续登录'),
        h(Input.Password, {
          placeholder: '请输入二级密码',
          onChange: (e: any) => { pass2Value = e.target.value; },
          onPressEnter: () => {
            if (pass2Value) {
              Modal.destroyAll();
              resolve(pass2Value);
            }
          },
        }),
      ]),
      okText: '确认',
      cancelText: '取消',
      onOk() { resolve(pass2Value || null); },
      onCancel() { resolve(null); },
    });
  });
}

export const useAuthStore = defineStore('auth', () => {
  const accessStore = useAccessStore();
  const userStore = useUserStore();
  const router = useRouter();

  const loginLoading = ref(false);

  /**
   * 异步处理注册操作
   * @param params 注册表单数据
   */
  async function authRegister(params: any) {
    try {
      loginLoading.value = true;
      await registerApi(params);
      notification.success({
        message: '注册成功',
        description: '请使用新账号登录',
        duration: 3,
      });
      await router.push(LOGIN_PATH);
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 异步处理登录操作
   * Asynchronously handle the login process
   * @param params 登录表单数据
   */
  async function authLogin(
    params: Recordable<any>,
    onSuccess?: () => Promise<void> | void,
  ) {
    // 异步处理用户登录操作并获取 accessToken
    let userInfo: null | UserInfo = null;
    try {
      loginLoading.value = true;
      const res = await loginApi(params);

      // 如果返回 code 为 5，表示需要管理员二次验证（后端已根据 pass2_kg 开关判断）
      if ((res as any).code === 5) {
        const pass2 = await showPass2Modal();
        if (!pass2) {
          loginLoading.value = false;
          return { userInfo: null };
        }
        // 再次发起登录请求，带上 pass2
        return await authLogin({ ...params, pass2 }, onSuccess);
      }

      const { accessToken } = res as AuthApi.LoginResult;

      // 如果成功获取到 accessToken
      if (accessToken) {
        accessStore.setAccessToken(accessToken);

        // 获取用户信息并存储到 accessStore 中
        const [fetchUserInfoResult, accessCodesResult] = await Promise.all([
          fetchUserInfo(),
          getAccessCodesApi(),
        ]);

        userInfo = fetchUserInfoResult;
        const accessCodes = accessCodesResult;

        userStore.setUserInfo(userInfo);
        accessStore.setAccessCodes(accessCodes);

        if (accessStore.loginExpired) {
          accessStore.setLoginExpired(false);
        } else {
          onSuccess
            ? await onSuccess?.()
            : await router.push(userInfo.homePath || DEFAULT_HOME_PATH);
        }

        // 动态加载站点配置
        try {
          const siteConfig = await getSiteConfigApi();
          if (siteConfig?.sitename) {
            updatePreferences({ app: { name: siteConfig.sitename } });
            document.title = siteConfig.sitename;
          }
          if (siteConfig?.hlogo) {
            updatePreferences({ logo: { source: siteConfig.hlogo } });
          } else if (siteConfig?.logo) {
            updatePreferences({ logo: { source: siteConfig.logo } });
          }
        } catch { /* ignore */ }

        if (userInfo?.realName) {
          notification.success({
            description: `${$t('authentication.loginSuccessDesc')}:${userInfo?.realName}`,
            duration: 2,
            message: $t('authentication.loginSuccess'),
            style: { width: 'min(90vw, 300px)' },
          });
        }
      }
    } finally {
      loginLoading.value = false;
    }

    return {
      userInfo,
    };
  }

  async function logout(redirect: boolean = true) {
    try {
      await logoutApi();
    } catch {
      // 不做任何处理
    }
    resetAllStores();
    accessStore.setLoginExpired(false);

    // 回登录页带上当前路由地址
    await router.replace({
      path: LOGIN_PATH,
      query: redirect
        ? {
            redirect: encodeURIComponent(router.currentRoute.value.fullPath),
          }
        : {},
    });
  }

  async function fetchUserInfo() {
    let userInfo: null | UserInfo = null;
    const res = await getUserInfoApi();
    userInfo = res;
    userStore.setUserInfo(userInfo);
    return userInfo;
  }

  function $reset() {
    loginLoading.value = false;
  }

  return {
    $reset,
    authLogin,
    authRegister,
    fetchUserInfo,
    loginLoading,
    logout,
  };
});
