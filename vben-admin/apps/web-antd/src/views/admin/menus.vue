<script lang="ts" setup>
import { ref, computed, onMounted, nextTick, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  Card, Table, InputNumber, Switch, Button, Tabs, TabPane, Tag, Space, message, Input, Modal, Select, Radio,
} from 'ant-design-vue';
import { ReloadOutlined, SaveOutlined, EditOutlined, PlusOutlined, DeleteOutlined, HolderOutlined } from '@ant-design/icons-vue';
import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { getMenuConfigs, saveMenuConfigs } from '#/api/menu-config';
import type { MenuConfigItem } from '#/api/menu-config';
import { getExtMenusApi, saveExtMenuApi, deleteExtMenuApi, reorderExtMenusApi, type ExtMenuItem } from '#/api/ext-menu';
import Sortable from 'sortablejs';

const loading = ref(false);
const saving = ref(false);
const activeTab = ref('frontend');

// ===== 树形菜单节点 =====
interface MenuTreeNode {
  key: string;
  menu_key: string;
  parent_key: string;
  title: string;
  icon: string;
  sort_order: number;
  visible: number;
  scope: string;
  children: MenuTreeNode[];
}

interface FlatMenuItem {
  menu_key: string;
  parent_key: string;
  title: string;
  icon: string;
  sort_order: number;
  visible: number;
  scope: string;
}

const frontendTree = ref<MenuTreeNode[]>([]);
const backendTree = ref<MenuTreeNode[]>([]);

const router = useRouter();

// ===== 树形工具函数 =====
function loopTree(data: MenuTreeNode[], key: string, callback: (item: MenuTreeNode, index: number, arr: MenuTreeNode[]) => void) {
  for (let i = 0; i < data.length; i++) {
    if (data[i]!.key === key) { callback(data[i]!, i, data); return; }
    if (data[i]!.children?.length) loopTree(data[i]!.children, key, callback);
  }
}

function findInTree(nodes: MenuTreeNode[], key: string): MenuTreeNode | null {
  for (const node of nodes) {
    if (node.key === key) return node;
    const found = node.children?.length ? findInTree(node.children, key) : null;
    if (found) return found;
  }
  return null;
}

function buildTree(items: FlatMenuItem[]): MenuTreeNode[] {
  const map = new Map<string, MenuTreeNode>();
  const roots: MenuTreeNode[] = [];
  for (const item of items) map.set(item.menu_key, { ...item, key: item.menu_key, children: [] });
  for (const item of items) {
    const node = map.get(item.menu_key)!;
    if (item.parent_key && map.has(item.parent_key)) {
      map.get(item.parent_key)!.children.push(node);
    } else {
      roots.push(node);
    }
  }
  const sortFn = (nodes: MenuTreeNode[]) => { nodes.sort((a, b) => a.sort_order - b.sort_order); nodes.forEach(n => sortFn(n.children)); };
  sortFn(roots);
  return roots;
}

function flattenTree(nodes: MenuTreeNode[], parentKey = '', scope = 'frontend'): FlatMenuItem[] {
  const result: FlatMenuItem[] = [];
  nodes.forEach((node, index) => {
    result.push({ menu_key: node.menu_key, parent_key: parentKey, title: node.title, icon: node.icon, sort_order: index, visible: node.visible, scope });
    result.push(...flattenTree(node.children, node.menu_key, scope));
  });
  return result;
}

// ===== 扩展菜单拖拽 =====
const extTableRef = ref<HTMLElement | null>(null);
let extSortable: Sortable | null = null;

function initExtSortable() {
  nextTick(() => {
    extSortable?.destroy();
    const tbody = extTableRef.value?.querySelector('.ant-table-tbody');
    if (tbody) {
      extSortable = Sortable.create(tbody as HTMLElement, {
        handle: '.drag-handle', animation: 200, ghostClass: 'drag-ghost', chosenClass: 'drag-chosen',
        onEnd({ oldIndex, newIndex }) {
          if (oldIndex == null || newIndex == null || oldIndex === newIndex) return;
          const [moved] = extMenus.value.splice(oldIndex, 1);
          extMenus.value.splice(newIndex, 0, moved!);
          extMenus.value.forEach((m, i) => { m.sort_order = i; });
          reorderExtMenusApi(extMenus.value.map(m => ({ id: m.id, sort_order: m.sort_order }))).catch(() => {});
        },
      });
    }
  });
}

onUnmounted(() => { extSortable?.destroy(); });

function extractMenuItems() {
  const routes = router.getRoutes();
  const frontItems: FlatMenuItem[] = [];
  const backItems: FlatMenuItem[] = [];

  const topRoutes = routes.filter(
    (r) => r.meta?.title && r.children?.length && r.path !== '/',
  );

  for (const route of topRoutes) {
    if (!route.meta?.title) continue;
    const isAdmin = route.path.startsWith('/admin');
    const target = isAdmin ? backItems : frontItems;
    const scope = isAdmin ? 'backend' : 'frontend';

    if (isAdmin && route.children?.length) {
      for (const child of route.children) {
        if (!child.meta?.title) continue;
        target.push({
          menu_key: child.name as string,
          parent_key: route.name as string,
          title: child.meta.title as string,
          icon: (child.meta.icon as string) || '',
          sort_order: (child.meta.order as number) || 0,
          visible: child.meta.hideInMenu ? 0 : 1,
          scope,
        });
        if (child.children && child.children.length > 0) {
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
            });
          }
        }
      }
    } else {
      target.push({
        menu_key: route.name as string, parent_key: '',
        title: route.meta.title as string, icon: (route.meta.icon as string) || '',
        sort_order: (route.meta.order as number) || 0,
        visible: route.meta.hideInMenu ? 0 : 1,
        scope,
      });
      if (route.children?.length) {
        for (const child of route.children) {
          if (!child.meta?.title) continue;
          if (child.meta.hideInMenu && route.children.length === 1) continue;
          if (child.meta.hideInMenu) continue;
          target.push({
            menu_key: child.name as string, parent_key: route.name as string,
            title: child.meta.title as string, icon: (child.meta.icon as string) || '',
            sort_order: (child.meta.order as number) || 0, visible: 1,
            scope,
          });
        }
      }
    }
  }

  frontItems.sort((a, b) => a.sort_order - b.sort_order);
  backItems.sort((a, b) => a.sort_order - b.sort_order);
  return { frontItems, backItems };
}

function mergeConfigs(items: FlatMenuItem[], configs: MenuConfigItem[]) {
  const configMap = new Map(configs.map((c) => [c.menu_key, c]));
  for (const item of items) {
    const saved = configMap.get(item.menu_key);
    if (saved) {
      item.sort_order = saved.sort_order;
      item.visible = saved.visible;
      if (saved.title) item.title = saved.title;
      if (saved.icon) item.icon = saved.icon;
      if (saved.parent_key !== undefined) item.parent_key = saved.parent_key;
    }
  }
  items.sort((a, b) => a.sort_order - b.sort_order);
}

async function loadData() {
  loading.value = true;
  try {
    const { frontItems, backItems } = extractMenuItems();
    const configs = await getMenuConfigs();
    mergeConfigs(frontItems, configs.filter((c) => c.scope === 'frontend'));
    mergeConfigs(backItems, configs.filter((c) => c.scope === 'backend'));
    frontendTree.value = buildTree(frontItems);
    backendTree.value = buildTree(backItems);
  } catch (e) {
    console.error('加载菜单配置失败', e);
    const { frontItems, backItems } = extractMenuItems();
    frontendTree.value = buildTree(frontItems);
    backendTree.value = buildTree(backItems);
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  saving.value = true;
  try {
    const allItems = [
      ...flattenTree(frontendTree.value, '', 'frontend'),
      ...flattenTree(backendTree.value, '', 'backend'),
    ];
    await saveMenuConfigs(allItems);
    message.success('菜单配置已保存，刷新页面后生效');
  } catch { message.error('保存失败'); }
  finally { saving.value = false; }
}


// ===== 编辑弹窗 =====
const editModalVisible = ref(false);
const editModalScope = ref<'frontend' | 'backend'>('frontend');
const editForm = ref({ menu_key: '', parent_key: '', title: '', icon: '', sort_order: 0, visible: 1 });

function findNodeByKey(key: string): MenuTreeNode | null {
  return findInTree(frontendTree.value, key) || findInTree(backendTree.value, key);
}

function findParentKey(tree: MenuTreeNode[], targetKey: string, pk = ''): string {
  for (const n of tree) {
    if (n.key === targetKey) return pk;
    if (n.children?.length) {
      const found = findParentKey(n.children, targetKey, n.menu_key);
      if (found !== '__NOT_FOUND__') return found;
    }
  }
  return '__NOT_FOUND__';
}

const parentOptions = computed(() => {
  const tree = editModalScope.value === 'frontend' ? frontendTree.value : backendTree.value;
  const opts: { value: string; label: string }[] = [{ value: '', label: '顶级菜单' }];
  const excludeKeys = new Set<string>();
  function collectDescendants(node: MenuTreeNode) {
    excludeKeys.add(node.key);
    node.children?.forEach(collectDescendants);
  }
  const current = findInTree(tree, editForm.value.menu_key);
  if (current) collectDescendants(current);
  function collect(nodes: MenuTreeNode[], depth = 0) {
    for (const n of nodes) {
      if (!excludeKeys.has(n.key)) {
        opts.push({ value: n.menu_key, label: '\u3000'.repeat(depth) + n.title });
        if (n.children?.length) collect(n.children, depth + 1);
      }
    }
  }
  collect(tree);
  return opts;
});

function openEditModal(record: MenuTreeNode, scope: 'frontend' | 'backend') {
  const tree = scope === 'frontend' ? frontendTree.value : backendTree.value;
  const pk = findParentKey(tree, record.menu_key);
  editModalScope.value = scope;
  editForm.value = {
    menu_key: record.menu_key,
    parent_key: pk === '__NOT_FOUND__' ? '' : pk,
    title: record.title,
    icon: record.icon,
    sort_order: record.sort_order,
    visible: record.visible,
  };
  editModalVisible.value = true;
}

function confirmEdit() {
  const scope = editModalScope.value;
  const treeRef = scope === 'frontend' ? frontendTree : backendTree;
  const node = findInTree(treeRef.value, editForm.value.menu_key);
  if (!node) { editModalVisible.value = false; return; }

  node.title = editForm.value.title;
  node.icon = editForm.value.icon;
  node.sort_order = editForm.value.sort_order;
  node.visible = editForm.value.visible;

  // 检查上级菜单是否变更
  const currentPk = findParentKey(treeRef.value, node.key);
  const newPk = editForm.value.parent_key;
  if (currentPk !== '__NOT_FOUND__' && newPk !== currentPk) {
    const data = [...treeRef.value];
    let dragObj: MenuTreeNode | null = null;
    loopTree(data, node.key, (item, index, arr) => { arr.splice(index, 1); dragObj = item; });
    if (dragObj) {
      if (!newPk) {
        data.push(dragObj);
      } else {
        loopTree(data, newPk, (item) => { item.children = item.children || []; item.children.push(dragObj!); });
      }
      treeRef.value = data;
    }
  }
  editModalVisible.value = false;
}

// 菜单表格列定义
const menuColumns = [
  { title: '菜单名称', key: 'title', width: 260 },
  { title: '图标', key: 'icon', width: 60, align: 'center' as const },
  { title: 'Key', dataIndex: 'menu_key', width: 180, ellipsis: true },
  { title: '排序', dataIndex: 'sort_order', width: 80, align: 'center' as const },
  { title: '状态', key: 'visible', width: 80, align: 'center' as const },
  { title: '操作', key: 'action', width: 100, align: 'center' as const },
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

function openExtEdit(item: ExtMenuItem) { extForm.value = { ...item }; extEditVisible.value = true; }

async function handleExtSave() {
  if (!extForm.value.title || !extForm.value.url) { message.warning('请填写标题和URL'); return; }
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
      try { await deleteExtMenuApi(item.id); message.success('已删除'); loadExtMenus(); }
      catch { message.error('删除失败'); }
    },
  });
}

const extColumns = [
  { title: '', key: 'drag', width: 40, align: 'center' as const },
  { title: '图标', key: 'icon', width: 60, align: 'center' as const },
  { title: '标题', dataIndex: 'title', ellipsis: true },
  { title: '地址', dataIndex: 'url', ellipsis: true },
  { title: '位置', key: 'scope', width: 80, align: 'center' as const },
  { title: '排序', dataIndex: 'sort_order', width: 80, align: 'center' as const },
  { title: '显示', key: 'visible', width: 80, align: 'center' as const },
  { title: '操作', key: 'action', width: 100, align: 'center' as const },
];

onMounted(async () => { await loadData(); await loadExtMenus(); initExtSortable(); });
</script>

<template>
  <Page title="菜单管理" description="管理系统菜单显示顺序、层级与可见性，保存后刷新页面生效。">
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

    <Card>
      <Tabs v-model:activeKey="activeTab">
        <!-- ===== 前台菜单 ===== -->
        <TabPane key="frontend" tab="前台菜单">
          <Table
            :columns="menuColumns" :data-source="frontendTree" :loading="loading"
            :pagination="false" row-key="menu_key" size="small" bordered
            :default-expand-all-rows="true"
            :row-class-name="(record: any) => record.visible === 0 ? 'row-disabled' : ''"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'title'">
                <span class="font-medium">{{ record.title }}</span>
              </template>
              <template v-else-if="column.key === 'icon'">
                <IconifyIcon v-if="record.icon" :icon="record.icon" :style="{ fontSize: '18px' }" />
                <span v-else class="text-gray-300">-</span>
              </template>
              <template v-else-if="column.key === 'visible'">
                <Tag :color="record.visible === 1 ? 'green' : 'red'" class="m-0">
                  {{ record.visible === 1 ? '启用' : '停用' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <Button type="primary" size="small" shape="circle" @click="openEditModal(record, 'frontend')">
                  <template #icon><EditOutlined /></template>
                </Button>
              </template>
            </template>
          </Table>
        </TabPane>

        <!-- ===== 后台菜单 ===== -->
        <TabPane key="backend" tab="后台菜单">
          <Table
            :columns="menuColumns" :data-source="backendTree" :loading="loading"
            :pagination="false" row-key="menu_key" size="small" bordered
            :default-expand-all-rows="true"
            :row-class-name="(record: any) => record.visible === 0 ? 'row-disabled' : ''"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'title'">
                <span class="font-medium">{{ record.title }}</span>
              </template>
              <template v-else-if="column.key === 'icon'">
                <IconifyIcon v-if="record.icon" :icon="record.icon" :style="{ fontSize: '18px' }" />
                <span v-else class="text-gray-300">-</span>
              </template>
              <template v-else-if="column.key === 'visible'">
                <Tag :color="record.visible === 1 ? 'green' : 'red'" class="m-0">
                  {{ record.visible === 1 ? '启用' : '停用' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <Button type="primary" size="small" shape="circle" @click="openEditModal(record, 'backend')">
                  <template #icon><EditOutlined /></template>
                </Button>
              </template>
            </template>
          </Table>
        </TabPane>

        <!-- ===== 扩展菜单 ===== -->
        <TabPane key="ext" tab="扩展菜单">
          <div class="mb-3 flex items-center justify-between">
            <span class="text-gray-500 text-sm">外部页面（iframe 嵌入侧边栏）</span>
            <Button type="primary" size="small" @click="openExtAdd">
              <template #icon><PlusOutlined /></template>
              添加
            </Button>
          </div>
          <div ref="extTableRef">
          <Table
            :columns="extColumns" :data-source="extMenus" :loading="extLoading"
            :pagination="false" row-key="id" size="small" bordered
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'drag'">
                <HolderOutlined class="drag-handle" />
              </template>
              <template v-else-if="column.key === 'icon'">
                <IconifyIcon v-if="record.icon" :icon="record.icon" :style="{ fontSize: '18px' }" />
                <span v-else>-</span>
              </template>
              <template v-else-if="column.key === 'scope'">
                <Tag :color="record.scope === 'frontend' ? 'blue' : 'orange'" class="m-0">
                  {{ record.scope === 'frontend' ? '前台' : '后台' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'visible'">
                <Tag :color="record.visible === 1 ? 'green' : 'default'">
                  {{ record.visible === 1 ? '显示' : '隐藏' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <Space :size="4">
                  <Button size="small" @click="openExtEdit(record)"><EditOutlined /></Button>
                  <Button size="small" danger @click="handleExtDelete(record)"><DeleteOutlined /></Button>
                </Space>
              </template>
            </template>
          </Table>
          </div>
        </TabPane>
      </Tabs>
    </Card>

    <!-- 修改菜单弹窗 -->
    <Modal v-model:open="editModalVisible" title="修改菜单" @ok="confirmEdit" width="560px">
      <div class="py-3 space-y-5">
        <div class="form-row">
          <label class="form-label">上级菜单</label>
          <Select v-model:value="editForm.parent_key" style="flex:1" :options="parentOptions" />
        </div>
        <div class="form-row">
          <label class="form-label">菜单图标</label>
          <div style="flex:1" class="flex items-center gap-2">
            <Input v-model:value="editForm.icon" placeholder="mdi:icon-name" style="flex:1" />
            <div v-if="editForm.icon" class="w-8 h-8 flex items-center justify-center bg-gray-50 rounded border">
              <IconifyIcon :icon="editForm.icon" :style="{ fontSize: '20px' }" />
            </div>
          </div>
        </div>
        <div class="form-row">
          <label class="form-label"><span class="text-red-500">*</span> 菜单名称</label>
          <Input v-model:value="editForm.title" style="flex:1" />
        </div>
        <div class="form-row">
          <label class="form-label">排序</label>
          <InputNumber v-model:value="editForm.sort_order" :min="-99" :max="999" style="width: 140px" />
        </div>
        <div class="form-row">
          <label class="form-label">状态</label>
          <Radio.Group :value="editForm.visible" @change="(e: any) => editForm.visible = e.target.value">
            <Radio :value="1">启用</Radio>
            <Radio :value="0">停用</Radio>
          </Radio.Group>
        </div>
      </div>
    </Modal>

    <!-- 扩展菜单编辑弹窗 -->
    <Modal v-model:open="extEditVisible" :title="extForm.id ? '编辑扩展菜单' : '添加扩展菜单'" @ok="handleExtSave" :confirm-loading="extSaving" width="520px">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">标题 <span class="text-red-500">*</span></label>
          <Input v-model:value="extForm.title" placeholder="如：旧版管理面板" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">地址 <span class="text-red-500">*</span></label>
          <Input v-model:value="extForm.url" placeholder="/php-api/ext/index.php 或完整URL" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">图标</label>
            <Input v-model:value="extForm.icon" placeholder="mdi:puzzle-outline" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">位置</label>
            <Select v-model:value="extForm.scope" style="width: 100%">
              <Select.Option value="frontend">前台</Select.Option>
              <Select.Option value="backend">后台</Select.Option>
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
            <Switch :checked="extForm.visible === 1" @change="(v: any) => (extForm.visible = v ? 1 : 0)" checked-children="显示" un-checked-children="隐藏" />
          </div>
        </div>
      </div>
    </Modal>
  </Page>
</template>

<style scoped>
.form-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.form-label {
  width: 80px;
  text-align: right;
  font-size: 14px;
  color: #333;
  flex-shrink: 0;
}
:deep(.row-disabled) {
  opacity: 0.45;
}
.drag-handle {
  cursor: grab;
  color: #999;
  font-size: 16px;
  transition: color 0.2s;
}
.drag-handle:hover {
  color: #1890ff;
}
:deep(.drag-ghost) {
  opacity: 0.4;
  background: #e6f7ff;
}
:deep(.drag-chosen) {
  background: #fafafa;
}
</style>
