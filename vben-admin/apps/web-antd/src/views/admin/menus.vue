<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  InputNumber, Switch, Button, Tabs, TabPane, Tag, Space, message, Spin, Input, Tooltip,
} from 'ant-design-vue';
import { ReloadOutlined, SaveOutlined, UpOutlined, DownOutlined, EditOutlined, CheckOutlined } from '@ant-design/icons-vue';
import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { getMenuConfigs, saveMenuConfigs } from '#/api/menu-config';
import type { MenuConfigItem } from '#/api/menu-config';

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

onMounted(loadData);
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
              class="group rounded-lg border bg-white p-3 transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
              :class="[
                item.level === 0 ? 'border-blue-200 dark:border-blue-800' : 'border-gray-200 ml-8',
                item.visible === 0 ? 'opacity-50' : '',
              ]"
            >
              <div class="flex items-center gap-3">
                <!-- 图标 -->
                <div
                  class="flex h-9 w-9 items-center justify-center rounded-lg text-lg"
                  :class="item.level === 0 ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'"
                >
                  <IconifyIcon v-if="item.icon" :icon="item.icon" :style="{ fontSize: '18px' }" />
                  <span v-else class="text-sm">-</span>
                </div>

                <!-- 标题 + key -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <Tag v-if="item.level === 0" color="blue" class="m-0" style="font-size:11px">顶级</Tag>
                    <Tag v-else color="green" class="m-0" style="font-size:11px">子级</Tag>
                    <!-- 内联编辑标题 -->
                    <template v-if="editingKey === item.menu_key">
                      <Input v-model:value="editingTitle" size="small" style="width:140px" @press-enter="confirmEditTitle(item)" />
                      <Button type="link" size="small" @click="confirmEditTitle(item)"><CheckOutlined /></Button>
                      <Button type="link" size="small" danger @click="cancelEditTitle">取消</Button>
                    </template>
                    <template v-else>
                      <span class="font-medium text-sm text-gray-800 dark:text-gray-200">{{ item.title }}</span>
                      <Tooltip title="编辑标题">
                        <EditOutlined class="cursor-pointer text-gray-300 hover:text-blue-500 transition-colors text-xs" @click="startEditTitle(item)" />
                      </Tooltip>
                    </template>
                  </div>
                  <div class="text-xs text-gray-400 mt-0.5 font-mono">{{ item.menu_key }}</div>
                </div>

                <!-- 排序 -->
                <div class="flex items-center gap-1">
                  <span class="text-xs text-gray-400 mr-1 hidden sm:inline">排序</span>
                  <InputNumber v-model:value="item.sort_order" :min="-99" :max="999" size="small" style="width:70px" />
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
                <Space size="small">
                  <Tooltip title="上移">
                    <Button size="small" shape="circle" :disabled="index === 0" @click="moveUp(frontendMenus, index)">
                      <template #icon><UpOutlined /></template>
                    </Button>
                  </Tooltip>
                  <Tooltip title="下移">
                    <Button size="small" shape="circle" :disabled="index === frontendMenus.length - 1" @click="moveDown(frontendMenus, index)">
                      <template #icon><DownOutlined /></template>
                    </Button>
                  </Tooltip>
                </Space>
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
              class="group rounded-lg border bg-white p-3 transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
              :class="[
                item.level === 1 ? 'border-indigo-200 dark:border-indigo-800' : 'border-gray-200 ml-8',
                item.visible === 0 ? 'opacity-50' : '',
              ]"
            >
              <div class="flex items-center gap-3">
                <!-- 图标 -->
                <div
                  class="flex h-9 w-9 items-center justify-center rounded-lg text-lg"
                  :class="item.level === 1 ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400' : 'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'"
                >
                  <IconifyIcon v-if="item.icon" :icon="item.icon" :style="{ fontSize: '18px' }" />
                  <span v-else class="text-sm">-</span>
                </div>

                <!-- 标题 + key -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <Tag v-if="item.level === 1" color="blue" class="m-0" style="font-size:11px">分组</Tag>
                    <Tag v-else color="green" class="m-0" style="font-size:11px">子级</Tag>
                    <template v-if="editingKey === item.menu_key">
                      <Input v-model:value="editingTitle" size="small" style="width:140px" @press-enter="confirmEditTitle(item)" />
                      <Button type="link" size="small" @click="confirmEditTitle(item)"><CheckOutlined /></Button>
                      <Button type="link" size="small" danger @click="cancelEditTitle">取消</Button>
                    </template>
                    <template v-else>
                      <span class="font-medium text-sm text-gray-800 dark:text-gray-200">{{ item.title }}</span>
                      <Tooltip title="编辑标题">
                        <EditOutlined class="cursor-pointer text-gray-300 hover:text-blue-500 transition-colors text-xs" @click="startEditTitle(item)" />
                      </Tooltip>
                    </template>
                  </div>
                  <div class="text-xs text-gray-400 mt-0.5 font-mono">{{ item.menu_key }}</div>
                </div>

                <!-- 排序 -->
                <div class="flex items-center gap-1">
                  <span class="text-xs text-gray-400 mr-1 hidden sm:inline">排序</span>
                  <InputNumber v-model:value="item.sort_order" :min="-99" :max="999" size="small" style="width:70px" />
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
                <Space size="small">
                  <Tooltip title="上移">
                    <Button size="small" shape="circle" :disabled="index === 0" @click="moveUp(backendMenus, index)">
                      <template #icon><UpOutlined /></template>
                    </Button>
                  </Tooltip>
                  <Tooltip title="下移">
                    <Button size="small" shape="circle" :disabled="index === backendMenus.length - 1" @click="moveDown(backendMenus, index)">
                      <template #icon><DownOutlined /></template>
                    </Button>
                  </Tooltip>
                </Space>
              </div>
            </div>
          </div>
        </TabPane>
      </Tabs>
    </Spin>
  </Page>
</template>
