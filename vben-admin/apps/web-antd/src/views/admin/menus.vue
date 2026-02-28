<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import {
  Card, Table, InputNumber, Switch, Button, Tabs, TabPane, Tag, Space, message, Spin,
} from 'ant-design-vue';
import { ReloadOutlined, SaveOutlined, MenuOutlined } from '@ant-design/icons-vue';
import { Page } from '@vben/common-ui';
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

const frontendColumns = computed(() => [
  {
    title: '菜单名称',
    dataIndex: 'title',
    key: 'title',
    width: 200,
  },
  {
    title: '路由 Key',
    dataIndex: 'menu_key',
    key: 'menu_key',
    width: 180,
  },
  {
    title: '图标',
    dataIndex: 'icon',
    key: 'icon',
    width: 160,
    ellipsis: true,
  },
  {
    title: '排序',
    dataIndex: 'sort_order',
    key: 'sort_order',
    width: 120,
  },
  {
    title: '显示',
    dataIndex: 'visible',
    key: 'visible',
    width: 80,
  },
  {
    title: '操作',
    key: 'action',
    width: 120,
  },
]);

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
        <TabPane key="frontend" tab="前台菜单">
          <Card :bordered="false">
            <Table
              :columns="frontendColumns"
              :dataSource="frontendMenus"
              :pagination="false"
              rowKey="menu_key"
              size="middle"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'title'">
                  <span :style="{ paddingLeft: record.level * 24 + 'px' }">
                    <Tag v-if="record.level === 0" color="blue">顶级</Tag>
                    <Tag v-else-if="record.level === 1" color="green">子级</Tag>
                    {{ record.title }}
                  </span>
                </template>
                <template v-else-if="column.key === 'icon'">
                  <span class="text-xs text-gray-400">{{ record.icon || '-' }}</span>
                </template>
                <template v-else-if="column.key === 'sort_order'">
                  <InputNumber
                    v-model:value="record.sort_order"
                    :min="-99"
                    :max="999"
                    size="small"
                    style="width: 80px"
                  />
                </template>
                <template v-else-if="column.key === 'visible'">
                  <Switch
                    :checked="record.visible === 1"
                    @change="(val: boolean) => (record.visible = val ? 1 : 0)"
                    size="small"
                  />
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space size="small">
                    <Button size="small" @click="moveUp(frontendMenus, index)" :disabled="index === 0">↑</Button>
                    <Button size="small" @click="moveDown(frontendMenus, index)" :disabled="index === frontendMenus.length - 1">↓</Button>
                  </Space>
                </template>
              </template>
            </Table>
          </Card>
        </TabPane>

        <TabPane key="backend" tab="后台菜单">
          <Card :bordered="false">
            <Table
              :columns="frontendColumns"
              :dataSource="backendMenus"
              :pagination="false"
              rowKey="menu_key"
              size="middle"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'title'">
                  <span :style="{ paddingLeft: record.level * 24 + 'px' }">
                    <Tag v-if="record.level === 1" color="blue">分组</Tag>
                    <Tag v-else-if="record.level === 2" color="green">子级</Tag>
                    {{ record.title }}
                  </span>
                </template>
                <template v-else-if="column.key === 'icon'">
                  <span class="text-xs text-gray-400">{{ record.icon || '-' }}</span>
                </template>
                <template v-else-if="column.key === 'sort_order'">
                  <InputNumber
                    v-model:value="record.sort_order"
                    :min="-99"
                    :max="999"
                    size="small"
                    style="width: 80px"
                  />
                </template>
                <template v-else-if="column.key === 'visible'">
                  <Switch
                    :checked="record.visible === 1"
                    @change="(val: boolean) => (record.visible = val ? 1 : 0)"
                    size="small"
                  />
                </template>
                <template v-else-if="column.key === 'action'">
                  <Space size="small">
                    <Button size="small" @click="moveUp(backendMenus, index)" :disabled="index === 0">↑</Button>
                    <Button size="small" @click="moveDown(backendMenus, index)" :disabled="index === backendMenus.length - 1">↓</Button>
                  </Space>
                </template>
              </template>
            </Table>
          </Card>
        </TabPane>
      </Tabs>
    </Spin>
  </Page>
</template>
