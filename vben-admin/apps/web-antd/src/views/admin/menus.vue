<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  InputNumber, Switch, Button, Tabs, TabPane, Tag, Space, message, Spin, Input, Tooltip, Popover, Modal, Select,
} from 'ant-design-vue';
import { ReloadOutlined, SaveOutlined, UpOutlined, DownOutlined, EditOutlined, CheckOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { getMenuConfigs, saveMenuConfigs } from '#/api/menu-config';
import type { MenuConfigItem } from '#/api/menu-config';
import { getExtMenusApi, saveExtMenuApi, deleteExtMenuApi, type ExtMenuItem } from '#/api/ext-menu';

const loading = ref(false);
const saving = ref(false);
const activeTab = ref('frontend');

// 从路由提取的菜单项（含编辑状态）
interface MenuRow {
  menu_key: string;
  parent_key: string;
  title: string;
  icon: string;
  sort_order: number;
  visible: number;
  scope: string;
  level: number;
  children_keys: string[];
}

const frontendMenus = ref<MenuRow[]>([]);
const backendMenus = ref<MenuRow[]>([]);

const router = useRouter();

/** 从 Vue Router 提取菜单树 */
function extractMenuItems() {
  const routes = router.getRoutes();
  const frontItems: MenuRow[] = [];
  const backItems: MenuRow[] = [];

  // 获取顶级路由（有 BasicLayout 组件的）
  const topRoutes = routes.filter(
    (r) => r.meta?.title && r.children?.length && r.path !== '/',
  );

  for (const route of topRoutes) {
    // 跳过无 title 的路由
    if (!route.meta?.title) continue;

    const isAdmin = route.path.startsWith('/admin');
    const target = isAdmin ? backItems : frontItems;
    const scope = isAdmin ? 'backend' : 'frontend';

    if (isAdmin && route.children?.length) {
      // 后台管理的子菜单
      for (const child of route.children) {
        if (!child.meta?.title) continue;
        const hasSubChildren = child.children && child.children.length > 0;

        target.push({
          menu_key: child.name as string,
          parent_key: route.name as string,
          title: child.meta.title as string,
          icon: (child.meta.icon as string) || '',
          sort_order: (child.meta.order as number) || 0,
          visible: child.meta.hideInMenu ? 0 : 1,
          scope,
          level: 1,
          children_keys: [],
        });

        if (hasSubChildren) {
          for (const sub of child.children!) {
            if (!sub.meta?.title) continue;
            target.push({
              menu_key: sub.name as string,
              parent_key: child.name as string,
              title: sub.meta.title as string,
              icon: (sub.meta.icon as string) || '',
              sort_order: (sub.meta.order as number) || 0,
              visible: sub.meta.hideInMenu ? 0 : 1,
              scope,
              level: 2,
              children_keys: [],
            });
          }
        }
      }
    } else {
      // 前台顶级菜单
      target.push({
        menu_key: route.name as string,
        parent_key: '',
        title: route.meta.title as string,
        icon: (route.meta.icon as string) || '',
        sort_order: (route.meta.order as number) || 0,
        visible: route.meta.hideInMenu ? 0 : 1,
        scope,
        level: 0,
        children_keys: [],
      });

      // 前台子菜单
      if (route.children?.length) {
        for (const child of route.children) {
          if (!child.meta?.title) continue;
          // 跳过 hideInMenu 的单页子路由
          if (child.meta.hideInMenu && route.children.length === 1) continue;
          if (child.meta.hideInMenu) continue;
          target.push({
            menu_key: child.name as string,
            parent_key: route.name as string,
            title: child.meta.title as string,
            icon: (child.meta.icon as string) || '',
            sort_order: (child.meta.order as number) || 0,
            visible: 1,
            scope,
            level: 1,
            children_keys: [],
          });
        }
      }
    }
  }

  // 按 sort_order 排序
  frontItems.sort((a, b) => a.sort_order - b.sort_order);
  backItems.sort((a, b) => a.sort_order - b.sort_order);

  return { frontItems, backItems };
}

/** 合并后端已保存的配置到路由数据 */
function mergeConfigs(items: MenuRow[], configs: MenuConfigItem[]) {
  const configMap = new Map(configs.map((c) => [c.menu_key, c]));
  for (const item of items) {
    const saved = configMap.get(item.menu_key);
    if (saved) {
      item.sort_order = saved.sort_order;
      item.visible = saved.visible;
      if (saved.title) item.title = saved.title;
      if (saved.icon) item.icon = saved.icon;
    }
  }
  items.sort((a, b) => a.sort_order - b.sort_order);
}

/** 加载数据 */
async function loadData() {
  loading.value = true;
  try {
    const { frontItems, backItems } = extractMenuItems();
    // 从后端拉取已保存的配置
    const configs = await getMenuConfigs();
    mergeConfigs(frontItems, configs.filter((c) => c.scope === 'frontend'));
    mergeConfigs(backItems, configs.filter((c) => c.scope === 'backend'));
    frontendMenus.value = frontItems;
    backendMenus.value = backItems;
  } catch (e) {
    console.error('加载菜单配置失败', e);
    // 即使后端加载失败也展示路由数据
    const { frontItems, backItems } = extractMenuItems();
    frontendMenus.value = frontItems;
    backendMenus.value = backItems;
  } finally {
    loading.value = false;
  }
}

/** 保存配置 */
async function handleSave() {
  saving.value = true;
  try {
    const allItems = [
      ...frontendMenus.value.map((m) => ({
        menu_key: m.menu_key,
        parent_key: m.parent_key,
        title: m.title,
        icon: m.icon,
        sort_order: m.sort_order,
        visible: m.visible,
        scope: 'frontend',
      })),
      ...backendMenus.value.map((m) => ({
        menu_key: m.menu_key,
        parent_key: m.parent_key,
        title: m.title,
        icon: m.icon,
        sort_order: m.sort_order,
        visible: m.visible,
        scope: 'backend',
      })),
    ];
    await saveMenuConfigs(allItems);
    message.success('菜单配置已保存，刷新页面后生效');
  } catch {
    message.error('保存失败');
  } finally {
    saving.value = false;
  }
}

/** 上移 */
function moveUp(list: MenuRow[], index: number) {
  if (index <= 0) return;
  const current = list[index]!;
  const prev = list[index - 1]!;
  // 同级才能交换
  if (current.level !== prev.level) return;
  const tmpOrder = current.sort_order;
  current.sort_order = prev.sort_order;
  prev.sort_order = tmpOrder;
  list.splice(index - 1, 2, current, prev);
}

/** 下移 */
function moveDown(list: MenuRow[], index: number) {
  if (index >= list.length - 1) return;
  const current = list[index]!;
  const next = list[index + 1]!;
  if (current.level !== next.level) return;
  const tmpOrder = current.sort_order;
  current.sort_order = next.sort_order;
  next.sort_order = tmpOrder;
  list.splice(index, 2, next, current);
}

// 编辑标题
const editingKey = ref<string>('');
const editingTitle = ref('');

function startEditTitle(item: MenuRow) {
  editingKey.value = item.menu_key;
  editingTitle.value = item.title;
}
function confirmEditTitle(item: MenuRow) {
  item.title = editingTitle.value;
  editingKey.value = '';
}
function cancelEditTitle() {
  editingKey.value = '';
}

// 编辑图标
const editingIconKey = ref<string>('');
const editingIconValue = ref('');

function startEditIcon(item: MenuRow) {
  editingIconKey.value = item.menu_key;
  editingIconValue.value = item.icon;
}
function confirmEditIcon(item: MenuRow) {
  item.icon = editingIconValue.value.trim();
  editingIconKey.value = '';
}
function cancelEditIcon() {
  editingIconKey.value = '';
}

// 常用 mdi 图标列表
const commonMdiIcons = [
  'mdi:home', 'mdi:cog-outline', 'mdi:account-circle-outline', 'mdi:chart-bar',
  'mdi:book-open-page-variant-outline', 'mdi:file-document-outline', 'mdi:store-outline',
  'mdi:menu', 'mdi:bell-outline', 'mdi:shield-check-outline', 'mdi:cash-multiple',
  'mdi:account-group-outline', 'mdi:chat-processing-outline', 'mdi:wallet-plus-outline',
  'mdi:history', 'mdi:calendar-check', 'mdi:gift-outline', 'mdi:star-shooting-outline',
  'mdi:connection', 'mdi:bookshelf', 'mdi:file-tree-outline', 'mdi:tune-variant',
  'mdi:bullhorn-outline', 'mdi:medal-outline', 'mdi:headset', 'mdi:package-variant',
  'mdi:receipt-text-outline', 'mdi:credit-card-settings-outline', 'mdi:storefront-outline',
  'mdi:ticket-confirmation-outline', 'mdi:file-search-outline', 'mdi:file-multiple-outline',
  'mdi:account-supervisor-outline', 'mdi:book-education-outline', 'mdi:rocket-launch-outline',
  'mdi:database-outline', 'mdi:hammer-wrench', 'mdi:palette-outline', 'mdi:eye-outline',
  'mdi:lock-outline', 'mdi:key-outline', 'mdi:email-outline', 'mdi:phone-outline',
];

// ===== 扩展菜单管理 =====
const extMenus = ref<ExtMenuItem[]>([]);
const extLoading = ref(false);
const extEditVisible = ref(false);
const extSaving = ref(false);
const extForm = ref<Partial<ExtMenuItem>>({
  id: 0, title: '', icon: 'mdi:puzzle-outline', url: '', sort_order: 0, visible: 1, scope: 'backend',
});

async function loadExtMenus() {
  extLoading.value = true;
  try {
    const raw = await getExtMenusApi();
    extMenus.value = Array.isArray(raw) ? raw : [];
  } catch { extMenus.value = []; }
  finally { extLoading.value = false; }
}

function openExtAdd() {
  extForm.value = { id: 0, title: '', icon: 'mdi:puzzle-outline', url: '', sort_order: 0, visible: 1, scope: 'backend' };
  extEditVisible.value = true;
}

function openExtEdit(item: ExtMenuItem) {
  extForm.value = { ...item };
  extEditVisible.value = true;
}

async function handleExtSave() {
  if (!extForm.value.title || !extForm.value.url) {
    message.warning('请填写标题和URL');
    return;
  }
  extSaving.value = true;
  try {
    await saveExtMenuApi(extForm.value);
    message.success('保存成功');
    extEditVisible.value = false;
    loadExtMenus();
  } catch { message.error('保存失败'); }
  finally { extSaving.value = false; }
}

function handleExtDelete(item: ExtMenuItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定删除扩展菜单「${item.title}」？`,
    onOk: async () => {
      try {
        await deleteExtMenuApi(item.id);
        message.success('已删除');
        loadExtMenus();
      } catch { message.error('删除失败'); }
    },
  });
}

onMounted(() => {
  loadData();
  loadExtMenus();
});
</script>

<template>
  <Page title="菜单管理" description="调整前台和后台菜单的显示顺序与可见性，保存后刷新页面生效。">
    <template #extra>
      <Space>
        <Button @click="loadData" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
        <Button type="primary" @click="handleSave" :loading="saving">
          <template #icon><SaveOutlined /></template>
          保存配置
        </Button>
      </Space>
    </template>

    <Spin :spinning="loading">
      <Tabs v-model:activeKey="activeTab">
        <!-- ===== 前台菜单 ===== -->
        <TabPane key="frontend" tab="前台菜单">
          <div class="space-y-2">
            <div
              v-for="(item, index) in frontendMenus"
              :key="item.menu_key"
              class="group rounded-lg border bg-white p-2 sm:p-3 transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
              :class="[
                item.level === 0 ? 'border-blue-200 dark:border-blue-800' : 'border-gray-200 ml-4 sm:ml-8',
                item.visible === 0 ? 'opacity-50' : '',
              ]"
            >
              <div class="flex flex-wrap sm:flex-nowrap items-center gap-2 sm:gap-3">
                <!-- 图标（点击编辑） -->
                <Popover
                  :open="editingIconKey === item.menu_key"
                  trigger="click"
                  placement="bottomLeft"
                  @openChange="(v: boolean) => { if (v) startEditIcon(item); else cancelEditIcon(); }"
                >
                  <template #content>
                    <div style="width: 280px; max-width: 100vw">
                      <div class="mb-2">
                        <div class="text-xs text-gray-400 mb-1">输入图标名称（如 mdi:home）</div>
                        <div class="flex items-center gap-2">
                          <Input v-model:value="editingIconValue" size="small" placeholder="mdi:icon-name" class="flex-1" @press-enter="confirmEditIcon(item)" />
                          <Button type="primary" size="small" @click="confirmEditIcon(item)"><CheckOutlined /></Button>
                        </div>
                      </div>
                      <div v-if="editingIconValue" class="flex items-center gap-2 mb-2 p-2 bg-gray-50 rounded dark:bg-gray-800">
                        <span class="text-xs text-gray-400">预览：</span>
                        <IconifyIcon :icon="editingIconValue" :style="{ fontSize: '20px' }" />
                      </div>
                      <div class="text-xs text-gray-400 mb-1">常用图标（点击选择）</div>
                      <div class="grid grid-cols-6 sm:grid-cols-8 gap-0.5 max-h-[180px] overflow-y-auto" style="scrollbar-width: thin">
                        <Tooltip v-for="ic in commonMdiIcons" :key="ic" :title="ic">
                          <div
                            class="flex items-center justify-center rounded cursor-pointer h-8 w-8 transition-colors"
                            :class="editingIconValue === ic ? 'bg-blue-50 text-blue-500 ring-1 ring-blue-300' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'"
                            @click="editingIconValue = ic"
                          >
                            <IconifyIcon :icon="ic" :style="{ fontSize: '16px' }" />
                          </div>
                        </Tooltip>
                      </div>
                    </div>
                  </template>
                  <Tooltip title="点击修改图标">
                    <div
                      class="flex h-8 w-8 sm:h-9 sm:w-9 items-center justify-center rounded-lg text-lg cursor-pointer transition-all hover:ring-2 hover:ring-blue-300 flex-shrink-0"
                      :class="item.level === 0 ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'"
                    >
                      <IconifyIcon v-if="item.icon" :icon="item.icon" :style="{ fontSize: '16px' }" class="sm:text-[18px]" />
                      <span v-else class="text-sm">-</span>
                    </div>
                  </Tooltip>
                </Popover>

                <!-- 标题 + key -->
                <div class="min-w-0 flex-1 w-full sm:w-auto flex flex-col justify-center">
                  <div class="flex items-center gap-1 sm:gap-2 flex-wrap">
                    <Tag v-if="item.level === 0" color="blue" class="m-0" style="font-size:10px; line-height: 16px; padding: 0 4px;">顶级</Tag>
                    <Tag v-else color="green" class="m-0" style="font-size:10px; line-height: 16px; padding: 0 4px;">子级</Tag>
                    <!-- 内联编辑标题 -->
                    <template v-if="editingKey === item.menu_key">
                      <Input v-model:value="editingTitle" size="small" class="w-24 sm:w-[140px]" @press-enter="confirmEditTitle(item)" />
                      <Button type="link" size="small" class="px-1" @click="confirmEditTitle(item)"><CheckOutlined /></Button>
                      <Button type="link" size="small" danger class="px-1" @click="cancelEditTitle">取消</Button>
                    </template>
                    <template v-else>
                      <span class="font-medium text-xs sm:text-sm text-gray-800 dark:text-gray-200 truncate max-w-[120px] sm:max-w-none">{{ item.title }}</span>
                      <Tooltip title="编辑标题">
                        <EditOutlined class="cursor-pointer text-gray-300 hover:text-blue-500 transition-colors text-xs" @click="startEditTitle(item)" />
                      </Tooltip>
                    </template>
                  </div>
                  <div class="text-[10px] sm:text-xs text-gray-400 mt-0.5 font-mono truncate max-w-[160px] sm:max-w-none" :title="item.menu_key">{{ item.menu_key }}</div>
                </div>

                <div class="flex items-center gap-2 sm:gap-3 ml-auto w-full sm:w-auto justify-end border-t sm:border-t-0 pt-2 sm:pt-0 mt-1 sm:mt-0 border-gray-100 dark:border-gray-700">
                  <!-- 排序 -->
                  <div class="flex items-center gap-1">
                    <span class="text-[10px] sm:text-xs text-gray-400 mr-1 hidden sm:inline">排序</span>
                    <InputNumber v-model:value="item.sort_order" :min="-99" :max="999" size="small" class="w-14 sm:w-[70px]" />
                  </div>

                  <!-- 显示开关 -->
                  <Tooltip :title="item.visible === 1 ? '点击隐藏' : '点击显示'">
                    <Switch
                      :checked="item.visible === 1"
                      @change="(val: boolean) => (item.visible = val ? 1 : 0)"
                      size="small"
                      :checked-children="'显'"
                      :un-checked-children="'隐'"
                    />
                  </Tooltip>

                  <!-- 排序按钮 -->
                  <Space size="small" :size="4">
                    <Tooltip title="上移">
                      <Button size="small" shape="circle" class="w-6 h-6 min-w-0" :disabled="index === 0" @click="moveUp(frontendMenus, index)">
                        <template #icon><UpOutlined class="text-[10px]" /></template>
                      </Button>
                    </Tooltip>
                    <Tooltip title="下移">
                      <Button size="small" shape="circle" class="w-6 h-6 min-w-0" :disabled="index === frontendMenus.length - 1" @click="moveDown(frontendMenus, index)">
                        <template #icon><DownOutlined class="text-[10px]" /></template>
                      </Button>
                    </Tooltip>
                  </Space>
                </div>
              </div>
            </div>
          </div>
        </TabPane>

        <!-- ===== 后台菜单 ===== -->
        <TabPane key="backend" tab="后台菜单">
          <div class="space-y-2">
            <div
              v-for="(item, index) in backendMenus"
              :key="item.menu_key"
              class="group rounded-lg border bg-white p-2 sm:p-3 transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
              :class="[
                item.level === 1 ? 'border-indigo-200 dark:border-indigo-800' : 'border-gray-200 ml-4 sm:ml-8',
                item.visible === 0 ? 'opacity-50' : '',
              ]"
            >
              <div class="flex flex-wrap sm:flex-nowrap items-center gap-2 sm:gap-3">
                <!-- 图标（点击编辑） -->
                <Popover
                  :open="editingIconKey === item.menu_key"
                  trigger="click"
                  placement="bottomLeft"
                  @openChange="(v: boolean) => { if (v) startEditIcon(item); else cancelEditIcon(); }"
                >
                  <template #content>
                    <div style="width: 280px; max-width: 100vw">
                      <div class="mb-2">
                        <div class="text-xs text-gray-400 mb-1">输入图标名称（如 mdi:home）</div>
                        <div class="flex items-center gap-2">
                          <Input v-model:value="editingIconValue" size="small" placeholder="mdi:icon-name" class="flex-1" @press-enter="confirmEditIcon(item)" />
                          <Button type="primary" size="small" @click="confirmEditIcon(item)"><CheckOutlined /></Button>
                        </div>
                      </div>
                      <div v-if="editingIconValue" class="flex items-center gap-2 mb-2 p-2 bg-gray-50 rounded dark:bg-gray-800">
                        <span class="text-xs text-gray-400">预览：</span>
                        <IconifyIcon :icon="editingIconValue" :style="{ fontSize: '20px' }" />
                      </div>
                      <div class="text-xs text-gray-400 mb-1">常用图标（点击选择）</div>
                      <div class="grid grid-cols-6 sm:grid-cols-8 gap-0.5 max-h-[180px] overflow-y-auto" style="scrollbar-width: thin">
                        <Tooltip v-for="ic in commonMdiIcons" :key="ic" :title="ic">
                          <div
                            class="flex items-center justify-center rounded cursor-pointer h-8 w-8 transition-colors"
                            :class="editingIconValue === ic ? 'bg-blue-50 text-blue-500 ring-1 ring-blue-300' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'"
                            @click="editingIconValue = ic"
                          >
                            <IconifyIcon :icon="ic" :style="{ fontSize: '16px' }" />
                          </div>
                        </Tooltip>
                      </div>
                    </div>
                  </template>
                  <Tooltip title="点击修改图标">
                    <div
                      class="flex h-8 w-8 sm:h-9 sm:w-9 items-center justify-center rounded-lg text-lg cursor-pointer transition-all hover:ring-2 hover:ring-blue-300 flex-shrink-0"
                      :class="item.level === 1 ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400' : 'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'"
                    >
                      <IconifyIcon v-if="item.icon" :icon="item.icon" :style="{ fontSize: '16px' }" class="sm:text-[18px]" />
                      <span v-else class="text-sm">-</span>
                    </div>
                  </Tooltip>
                </Popover>

                <!-- 标题 + key -->
                <div class="min-w-0 flex-1 w-full sm:w-auto flex flex-col justify-center">
                  <div class="flex items-center gap-1 sm:gap-2 flex-wrap">
                    <Tag v-if="item.level === 1" color="blue" class="m-0" style="font-size:10px; line-height: 16px; padding: 0 4px;">分组</Tag>
                    <Tag v-else color="green" class="m-0" style="font-size:10px; line-height: 16px; padding: 0 4px;">子级</Tag>
                    <template v-if="editingKey === item.menu_key">
                      <Input v-model:value="editingTitle" size="small" class="w-24 sm:w-[140px]" @press-enter="confirmEditTitle(item)" />
                      <Button type="link" size="small" class="px-1" @click="confirmEditTitle(item)"><CheckOutlined /></Button>
                      <Button type="link" size="small" danger class="px-1" @click="cancelEditTitle">取消</Button>
                    </template>
                    <template v-else>
                      <span class="font-medium text-xs sm:text-sm text-gray-800 dark:text-gray-200 truncate max-w-[120px] sm:max-w-none">{{ item.title }}</span>
                      <Tooltip title="编辑标题">
                        <EditOutlined class="cursor-pointer text-gray-300 hover:text-blue-500 transition-colors text-xs" @click="startEditTitle(item)" />
                      </Tooltip>
                    </template>
                  </div>
                  <div class="text-[10px] sm:text-xs text-gray-400 mt-0.5 font-mono truncate max-w-[160px] sm:max-w-none" :title="item.menu_key">{{ item.menu_key }}</div>
                </div>

                <div class="flex items-center gap-2 sm:gap-3 ml-auto w-full sm:w-auto justify-end border-t sm:border-t-0 pt-2 sm:pt-0 mt-1 sm:mt-0 border-gray-100 dark:border-gray-700">
                  <!-- 排序 -->
                  <div class="flex items-center gap-1">
                    <span class="text-[10px] sm:text-xs text-gray-400 mr-1 hidden sm:inline">排序</span>
                    <InputNumber v-model:value="item.sort_order" :min="-99" :max="999" size="small" class="w-14 sm:w-[70px]" />
                  </div>

                  <!-- 显示开关 -->
                  <Tooltip :title="item.visible === 1 ? '点击隐藏' : '点击显示'">
                    <Switch
                      :checked="item.visible === 1"
                      @change="(val: boolean) => (item.visible = val ? 1 : 0)"
                      size="small"
                      :checked-children="'显'"
                      :un-checked-children="'隐'"
                    />
                  </Tooltip>

                  <!-- 排序按钮 -->
                  <Space size="small" :size="4">
                    <Tooltip title="上移">
                      <Button size="small" shape="circle" class="w-6 h-6 min-w-0" :disabled="index === 0" @click="moveUp(backendMenus, index)">
                        <template #icon><UpOutlined class="text-[10px]" /></template>
                      </Button>
                    </Tooltip>
                    <Tooltip title="下移">
                      <Button size="small" shape="circle" class="w-6 h-6 min-w-0" :disabled="index === backendMenus.length - 1" @click="moveDown(backendMenus, index)">
                        <template #icon><DownOutlined class="text-[10px]" /></template>
                      </Button>
                    </Tooltip>
                  </Space>
                </div>
              </div>
            </div>
          </div>
        </TabPane>
        <!-- ===== 扩展菜单 ===== -->
        <TabPane key="ext" tab="扩展菜单">
          <Spin :spinning="extLoading">
            <div class="mb-3 flex items-center justify-between">
              <span class="text-gray-500 text-sm">添加 PHP 单页等外部页面到侧边栏菜单（iframe 嵌入显示）</span>
              <Button type="primary" size="small" @click="openExtAdd">
                <template #icon><PlusOutlined /></template>
                添加扩展菜单
              </Button>
            </div>
            <div class="space-y-2">
              <div
                v-for="item in extMenus"
                :key="item.id"
                class="group rounded-lg border bg-white p-3 transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
                :class="item.visible === 0 ? 'opacity-50 border-gray-200' : 'border-purple-200 dark:border-purple-800'"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="flex h-9 w-9 items-center justify-center rounded-lg text-lg bg-purple-50 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400"
                  >
                    <IconifyIcon v-if="item.icon" :icon="item.icon" :style="{ fontSize: '18px' }" />
                    <span v-else class="text-sm">-</span>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <Tag color="purple" class="m-0" style="font-size:11px">扩展</Tag>
                      <Tag :color="item.scope === 'frontend' ? 'blue' : 'orange'" class="m-0" style="font-size:11px">
                        {{ item.scope === 'frontend' ? '前台' : '后台' }}
                      </Tag>
                      <span class="font-medium text-sm text-gray-800 dark:text-gray-200">{{ item.title }}</span>
                    </div>
                    <div class="text-xs text-gray-400 mt-0.5 font-mono truncate max-w-md">{{ item.url }}</div>
                  </div>
                  <div class="flex items-center gap-1">
                    <span class="text-xs text-gray-400 mr-1 hidden sm:inline">排序</span>
                    <Tag>{{ item.sort_order }}</Tag>
                  </div>
                  <Switch
                    :checked="item.visible === 1"
                    size="small"
                    :checked-children="'显'"
                    :un-checked-children="'隐'"
                    disabled
                  />
                  <Space size="small">
                    <Tooltip title="编辑">
                      <Button size="small" @click="openExtEdit(item)">
                        <template #icon><EditOutlined /></template>
                      </Button>
                    </Tooltip>
                    <Tooltip title="删除">
                      <Button size="small" danger @click="handleExtDelete(item)">
                        <template #icon><DeleteOutlined /></template>
                      </Button>
                    </Tooltip>
                  </Space>
                </div>
              </div>
              <div v-if="extMenus.length === 0" class="py-8 text-center text-gray-400">
                暂无扩展菜单，点击上方按钮添加
              </div>
            </div>
          </Spin>
        </TabPane>
      </Tabs>
    </Spin>

    <!-- 扩展菜单编辑弹窗 -->
    <Modal
      v-model:open="extEditVisible"
      :title="extForm.id ? '编辑扩展菜单' : '添加扩展菜单'"
      @ok="handleExtSave"
      :confirm-loading="extSaving"
      width="520px"
    >
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">菜单标题 <span class="text-red-500">*</span></label>
          <Input v-model:value="extForm.title" placeholder="如：旧版管理面板" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">页面地址 <span class="text-red-500">*</span></label>
          <Input v-model:value="extForm.url" placeholder="如：/php-api/ext/index.php 或完整URL" />
          <div class="text-xs text-gray-400 mt-1">
            PHP 文件放在 php-api/public/ 目录下，填写 /php-api/ext/xxx.php 即可通过反向代理访问
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">图标</label>
            <Input v-model:value="extForm.icon" placeholder="mdi:puzzle-outline" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">位置</label>
            <Select v-model:value="extForm.scope" style="width: 100%">
              <Select.Option value="frontend">前台菜单</Select.Option>
              <Select.Option value="backend">后台菜单</Select.Option>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">排序</label>
            <InputNumber v-model:value="extForm.sort_order" :min="-99" :max="999" style="width: 100%" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">可见性</label>
            <Switch
              :checked="extForm.visible === 1"
              @change="(v: any) => (extForm.visible = v ? 1 : 0)"
              checked-children="显示"
              un-checked-children="隐藏"
            />
          </div>
        </div>
        <div v-if="extForm.icon" class="flex items-center gap-2 p-2 bg-gray-50 rounded dark:bg-gray-800">
          <span class="text-xs text-gray-400">图标预览：</span>
          <IconifyIcon :icon="extForm.icon" :style="{ fontSize: '22px' }" />
        </div>
      </div>
    </Modal>
  </Page>
</template>
