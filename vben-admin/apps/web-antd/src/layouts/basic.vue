<script lang="ts" setup>
import type { NotificationItem } from '@vben/layouts';

import { computed, ref, watch, onMounted, onUnmounted, h } from 'vue';

import { AuthenticationLoginExpiredModal } from '@vben/common-ui';
import { useWatermark } from '@vben/hooks';
import { CircleUserRound, ArrowRightLeft, Settings } from '@vben/icons';
import {
  BasicLayout,
  LockScreen,
  Notification,
  UserDropdown,
} from '@vben/layouts';
import { preferences, updatePreferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { useRoute, useRouter } from 'vue-router';

import { Modal, Input, InputNumber, message } from 'ant-design-vue';
import { $t } from '#/locales';
import { useAuthStore } from '#/store';
import { getSiteConfigApi, getPublicAnnouncementsApi } from '#/api/admin';
import { getMenuConfigs } from '#/api/menu-config';
import { migrateSuperiorApi } from '#/api/user-center';
import { getAccessCodesApi, getUserInfoApi } from '#/api';
import { getChatSessionsApi, markChatReadApi } from '#/api/chat';
import LoginForm from '#/views/_core/authentication/login.vue';

const siteVersion = ref('');
const migrateEnabled = ref(false);

const notifications = ref<NotificationItem[]>([]);
const notifySessionMap = ref<Map<string, number>>(new Map());
let notifyTimer: ReturnType<typeof setInterval> | null = null;

// 已读公告ID集合（存 localStorage）
const readAnnIds = ref<Set<number>>(new Set(
  JSON.parse(localStorage.getItem('__read_ann_ids') || '[]'),
));
function markAnnRead(id: number) {
  readAnnIds.value.add(id);
  localStorage.setItem('__read_ann_ids', JSON.stringify([...readAnnIds.value]));
}

async function loadChatNotifications() {
  try {
    // 1) 聊天消息
    const chatItems: (NotificationItem & { _key: string })[] = [];
    const map = new Map<string, number>();
    try {
      const raw = await getChatSessionsApi();
      const sessions = raw;
      if (Array.isArray(sessions)) {
        for (const s of sessions) {
          if (s.unread_count <= 0) continue;
          const key = `chat_${s.list_id}`;
          map.set(key, s.list_id);
          chatItems.push({
            avatar: `https://q1.qlogo.cn/g?b=qq&nk=${s.uid || s.name}&s=640`,
            date: formatNotifyTime(s.last_time),
            isRead: false,
            message: s.last_msg || '',
            title: `${s.name}（${s.unread_count}条未读）`,
            _key: key,
          });
        }
      }
    } catch { /* ignore */ }
    notifySessionMap.value = map;

    // 2) 公告
    const annItems: (NotificationItem & { _key: string; _content?: string })[] = [];
    try {
      const annRes = await getPublicAnnouncementsApi(1, 10);
      if (annRes?.list?.length) {
        for (const a of annRes.list) {
          annItems.push({
            avatar: 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" rx="32" fill="%23f59e0b"/><text x="32" y="42" text-anchor="middle" font-size="32" fill="white">📢</text></svg>'),
            date: a.time ? a.time.slice(0, 16) : '',
            isRead: readAnnIds.value.has(a.id),
            message: a.title,
            title: (a.zhiding === '1' ? '[置顶] ' : '') + '系统公告',
            _key: `ann_${a.id}`,
            _content: a.content,
          } as any);
        }
      }
    } catch { /* ignore */ }

    // 合并：公告在前，聊天在后
    notifications.value = [...annItems, ...chatItems];
  } catch { /* ignore */ }
}

function formatNotifyTime(t: string) {
  if (!t) return '';
  const d = new Date(t);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return t.slice(11, 16);
  return t.slice(5, 16);
}

const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const { destroyWatermark, updateWatermark } = useWatermark();
const showDot = computed(() =>
  notifications.value.some((item) => !item.isRead),
);

const router = useRouter();
const route = useRoute();

// --- 动态菜单切换：admin 路径下只显示后台管理子菜单，其他路径隐藏后台管理 ---
const fullMenusBackup = ref<any[]>([]);
let menusSynced = false;

watch(
  () => accessStore.accessMenus,
  (menus) => {
    if (!menusSynced && menus.length > 0) {
      fullMenusBackup.value = [...menus];
      menusSynced = true;
      swapMenusForRoute(route.path);
    }
  },
  { immediate: true },
);

function swapMenusForRoute(path: string) {
  if (!fullMenusBackup.value.length) return;
  if (path.startsWith('/admin')) {
    const adminMenu = fullMenusBackup.value.find((m: any) => m.path === '/admin');
    if (adminMenu?.children?.length) {
      accessStore.setAccessMenus(adminMenu.children);
    }
  } else {
    accessStore.setAccessMenus(
      fullMenusBackup.value.filter((m: any) => m.path !== '/admin'),
    );
  }
}

watch(
  () => router.currentRoute.value.path,
  (path) => {
    if (menusSynced) swapMenusForRoute(path);
  },
  { flush: 'sync' },
);

const isAdmin = computed(() => {
  const roles = userStore.userInfo?.roles ?? [];
  return roles.some((r: string) => r === 'super' || r === 'admin');
});

const onAdminPage = computed(() => route.path.startsWith('/admin'));

const menus = computed(() => {
  const list = [
    {
      handler: () => { router.push('/user/profile'); },
      icon: CircleUserRound,
      text: '我的资料',
    },
    {
      handler: () => { openMigrateModal(); },
      icon: ArrowRightLeft,
      text: '上级迁移',
    },
  ];
  if (isAdmin.value) {
    list.push({
      handler: () => {
        const target = onAdminPage.value ? '/' : '/admin';
        swapMenusForRoute(target);
        router.push(target);
      },
      icon: Settings,
      text: onAdminPage.value ? '返回前台' : '后台管理',
    });
  }
  return list;
});

const avatar = computed(() => {
  const username = userStore.userInfo?.username || '';
  if (username) {
    return `https://q1.qlogo.cn/g?b=qq&nk=${username}&s=640`;
  }
  return preferences.app.defaultAvatar;
});

async function handleLogout() {
  await authStore.logout(false);
}

function handleNoticeClear() {
  notifications.value = [];
}

function handleMakeAll() {
  notifications.value.forEach((item) => (item.isRead = true));
  notifySessionMap.value.forEach((listId) => {
    markChatReadApi(listId).catch(() => {});
  });
}

function handleNoticeRead(item: NotificationItem) {
  item.isRead = true;
  const key = (item as any)._key as string;
  if (key?.startsWith('ann_')) {
    // 公告：弹窗显示内容
    const annId = parseInt(key.replace('ann_', ''), 10);
    markAnnRead(annId);
    Modal.info({
      title: item.message || '系统公告',
      content: h('div', {
        style: 'white-space:pre-wrap;max-height:400px;overflow:auto',
        innerHTML: (item as any)._content || item.message || '',
      }),
      okText: '知道了',
      width: 'min(90vw, 500px)',
    });
  } else {
    // 聊天消息
    const listId = key ? notifySessionMap.value.get(key) : undefined;
    if (listId) {
      markChatReadApi(listId).catch(() => {});
    }
    router.push('/chat');
  }
}

function handleViewAll() {
  router.push('/chat');
}

// 上级迁移弹窗
const migrateUidVal = ref<number | null>(null);
const migrateYqmVal = ref('');
function openMigrateModal() {
  migrateUidVal.value = null;
  migrateYqmVal.value = '';
  Modal.confirm({
    title: '上级迁移',
    icon: null,
    content: h('div', { style: 'display:flex;flex-direction:column;gap:12px;margin-top:12px' }, [
      h('div', {}, [
        h('label', { style: 'font-size:13px;font-weight:500;display:block;margin-bottom:4px' }, '新上级UID'),
        h(InputNumber, { min: 1, style: 'width:100%', placeholder: '输入UID', onChange: (v: any) => { migrateUidVal.value = v; } }),
      ]),
      h('div', {}, [
        h('label', { style: 'font-size:13px;font-weight:500;display:block;margin-bottom:4px' }, '新上级邀请码'),
        h(Input, { placeholder: '输入邀请码', onChange: (e: any) => { migrateYqmVal.value = e.target.value; } }),
      ]),
    ]),
    okText: '确认迁移',
    cancelText: '取消',
    async onOk() {
      if (!migrateUidVal.value || !migrateYqmVal.value) {
        message.warning('请填写完整');
        return Promise.reject();
      }
      try {
        const raw = await migrateSuperiorApi(migrateUidVal.value, migrateYqmVal.value);
        const res = raw;
        message.success(res?.message || '迁移成功');
      } catch {
        // 全局拦截器已弹出错误提示，这里只阻止弹窗关闭
        return Promise.reject();
      }
    },
  });
}
const siteName = ref('');
const hasBackupToken = ref(!!localStorage.getItem('admin_backup_token'));

async function handleSwitchBack() {
  const backupToken = localStorage.getItem('admin_backup_token');
  if (!backupToken) return;
  accessStore.setAccessToken(backupToken);
  localStorage.removeItem('admin_backup_token');
  try {
    const [userRes, codesRes] = await Promise.all([
      getUserInfoApi(),
      getAccessCodesApi(),
    ]);
    const userInfo = userRes;
    const codes = codesRes;
    userStore.setUserInfo(userInfo);
    accessStore.setAccessCodes(codes);
  } catch { /* ignore */ }
  window.location.href = '/';
}

onMounted(async () => {
  try {
    const cfg = await getSiteConfigApi();
    siteVersion.value = cfg?.version || '';
    siteName.value = cfg?.sitename || '';
    // 启用 footer 并显示站点名称+版本号
    if (siteVersion.value || siteName.value) {
      updatePreferences({
        footer: { enable: true, fixed: true },
        copyright: {
          enable: true,
          companyName: siteName.value + (siteVersion.value ? ` v${siteVersion.value}` : ''),
          companySiteLink: '',
          date: new Date().getFullYear().toString(),
        },
      });
    }
    // 上级迁移开关
    migrateEnabled.value = cfg?.sjqykg === '1';
    // 水印：sykg 未设置或为 '1' 时默认开启
    if (!cfg?.sykg || cfg.sykg === '1') {
      updatePreferences({ app: { watermark: true } });
    }
    // SEO meta 标签
    if (cfg?.keywords) {
      let el = document.querySelector('meta[name="keywords"]') as HTMLMetaElement;
      if (!el) { el = document.createElement('meta'); el.name = 'keywords'; document.head.appendChild(el); }
      el.content = cfg.keywords;
    }
    if (cfg?.description) {
      let el = document.querySelector('meta[name="description"]') as HTMLMetaElement;
      if (!el) { el = document.createElement('meta'); el.name = 'description'; document.head.appendChild(el); }
      el.content = cfg.description;
    }
    // 反调试保护
    if (cfg?.anti_debug !== '0' && !document.getElementById('__anti_debug')) {
      const s = document.createElement('script');
      s.id = '__anti_debug';
      s.textContent = `(function(){
        // 1. debugger 计时检测
        setInterval(function(){
          var t=performance.now();debugger;
          if(performance.now()-t>100){window.location.href='about:blank';}
        },1000);
        // 2. 屏蔽 F12 / Ctrl+Shift+I / Ctrl+Shift+J / Ctrl+U
        document.addEventListener('keydown',function(e){
          if(e.key==='F12'||(e.ctrlKey&&e.shiftKey&&(e.key==='I'||e.key==='J'||e.key==='C'))||(e.ctrlKey&&e.key==='u')){e.preventDefault();e.stopPropagation();return false;}
        },true);
        // 3. 屏蔽右键菜单
        document.addEventListener('contextmenu',function(e){e.preventDefault();},true);
      })();`;
      document.head.appendChild(s);
    }
    // 自定义特效注入
    if (cfg?.webVfx_open === '1' && cfg?.webVfx) {
      const el = document.createElement('div');
      el.id = '__web_vfx';
      el.innerHTML = cfg.webVfx;
      if (!document.getElementById('__web_vfx')) document.body.appendChild(el);
      // 执行内联 script
      el.querySelectorAll('script').forEach((old) => {
        const ns = document.createElement('script');
        if (old.src) { ns.src = old.src; } else { ns.textContent = old.textContent; }
        old.replaceWith(ns);
      });
    }
  } catch { /* ignore */ }

  // 加载菜单配置并应用排序/可见性
  try {
    const menuConfigs = await getMenuConfigs();
    if (menuConfigs?.length) {
      const configMap = new Map(menuConfigs.map((c: any) => [c.menu_key, c]));
      // 修改路由 meta
      for (const r of router.getRoutes()) {
        const cfg = configMap.get(r.name as string);
        if (cfg) {
          if (typeof cfg.sort_order === 'number') r.meta.order = cfg.sort_order;
          r.meta.hideInMenu = cfg.visible === 0;
        }
      }
      // 重建菜单（触发 accessMenus 重排）
      if (fullMenusBackup.value.length) {
        // 对备份菜单递归应用排序
        const applySort = (items: any[]) => {
          for (const item of items) {
            const cfg = configMap.get(item.name);
            if (cfg) {
              item.meta = { ...item.meta, order: cfg.sort_order };
              if (cfg.visible === 0) item.meta.hideInMenu = true;
              else delete item.meta.hideInMenu;
            }
            if (item.children?.length) applySort(item.children);
          }
          items.sort((a: any, b: any) => (a.meta?.order ?? 0) - (b.meta?.order ?? 0));
        };
        applySort(fullMenusBackup.value);
        swapMenusForRoute(route.path);
      }
    }
  } catch { /* ignore */ }

  loadChatNotifications();
  notifyTimer = setInterval(loadChatNotifications, 30000);
});

onUnmounted(() => {
  if (notifyTimer) {
    clearInterval(notifyTimer);
    notifyTimer = null;
  }
});

watch(
  () => preferences.app.watermark,
  async (enable) => {
    if (enable) {
      const uid = userStore.userInfo?.userId || '';
      const username = userStore.userInfo?.username || '';
      const name = siteName.value || preferences.app.name || '';
      await updateWatermark({
        content: [name, username, uid].filter(Boolean).join('\n'),
        width: 200,
        height: 140,
        globalAlpha: 0.18,
        gridLayoutOptions: {
          cols: 3,
          gap: [10, 10],
          matrix: [
            [1, 0, 1],
            [0, 1, 0],
            [1, 0, 1],
          ],
          rows: 3,
        },
      });
    } else {
      destroyWatermark();
    }
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #user-dropdown>
      <UserDropdown
        :avatar
        :menus
        :text="userStore.userInfo?.realName"
        :description="(userStore.userInfo as any)?.desc || ''"
        @logout="handleLogout"
      />
    </template>
    <template #notification>
      <Notification
        :dot="showDot"
        :notifications="notifications"
        @clear="handleNoticeClear"
        @make-all="handleMakeAll"
        @read="handleNoticeRead"
        @view-all="handleViewAll"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar @to-login="handleLogout" />
    </template>
  </BasicLayout>
  <div
    v-if="hasBackupToken"
    style="position:fixed;bottom:80px;right:24px;z-index:9999;"
  >
    <button
      style="padding:10px 20px;background:#ff4d4f;color:#fff;border:none;border-radius:8px;cursor:pointer;font-size:14px;font-weight:600;box-shadow:0 4px 12px rgba(0,0,0,0.3);"
      @click="handleSwitchBack"
    >
      切回管理员
    </button>
  </div>
</template>
