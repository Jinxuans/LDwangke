<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Tag, Space, Input, Select, SelectOption,
  Modal, message, Popconfirm, InputNumber, Switch, Form, FormItem,
  Tooltip, Tabs, TabPane, Badge, Typography,
} from 'ant-design-vue';
import {
  PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined,
  ApiOutlined, FileOutlined,
} from '@ant-design/icons-vue';
import { getAllModulesApi, saveModuleApi, deleteModuleApi, type DynamicModule } from '#/api/module';
import IconSelect, { iconRegistry } from '#/components/IconSelect.vue';

// 图标 key → 组件映射（用于表格渲染）
const iconCompMap: Record<string, any> = Object.fromEntries(
  Object.entries(iconRegistry).map(([k, v]) => [k, v.comp]),
);

const loading = ref(false);
const tableData = ref<DynamicModule[]>([]);
const activeTab = ref('all');

// 编辑弹窗
const editVisible = ref(false);
const editLoading = ref(false);
const editForm = ref<Partial<DynamicModule>>({});
const isEdit = ref(false);

// 类型选项
const typeOptions = [
  { value: 'sport', label: '运动' },
  { value: 'intern', label: '实习' },
  { value: 'paper', label: '论文' },
];

const typeMap: Record<string, { label: string; color: string }> = {
  sport: { label: '运动', color: 'blue' },
  intern: { label: '实习', color: 'green' },
  paper: { label: '论文', color: 'purple' },
};

// 按 Tab 过滤
const filteredData = computed(() => {
  if (activeTab.value === 'all') return tableData.value;
  return tableData.value.filter((m) => m.type === activeTab.value);
});

// 统计
const countByType = computed(() => {
  const map: Record<string, number> = { all: tableData.value.length, sport: 0, intern: 0, paper: 0 };
  for (const m of tableData.value) {
    if (map[m.type] !== undefined) map[m.type]++;
  }
  return map;
});

// 加载列表
async function loadData() {
  loading.value = true;
  try {
    const raw = await getAllModulesApi();
    tableData.value = raw;
    if (!Array.isArray(tableData.value)) tableData.value = [];
  } catch (e) {
    console.error('加载模块失败:', e);
  } finally {
    loading.value = false;
  }
}

// 打开新增
function openAdd() {
  isEdit.value = false;
  editForm.value = {
    app_id: '', type: activeTab.value === 'all' ? 'sport' : activeTab.value,
    name: '', description: '', price: '', icon: '', api_base: '', view_url: '',
    status: 1, sort: 0, config: '{}',
  };
  editVisible.value = true;
}

// 打开编辑
function openEdit(row: DynamicModule) {
  isEdit.value = true;
  editForm.value = { ...row };
  editVisible.value = true;
}

// 保存
async function handleSave() {
  if (!editForm.value.app_id || !editForm.value.name) {
    message.warning('模块标识和名称不能为空');
    return;
  }
  editLoading.value = true;
  try {
    await saveModuleApi(editForm.value);
    message.success('保存成功');
    editVisible.value = false;
    loadData();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    editLoading.value = false;
  }
}

// 删除
async function handleDelete(id: number) {
  try {
    await deleteModuleApi(id);
    message.success('删除成功');
    loadData();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

// 快速切换状态
async function handleToggle(row: DynamicModule, checked: boolean) {
  try {
    await saveModuleApi({ ...row, status: checked ? 1 : 0 });
    message.success(checked ? '已启用' : '已禁用');
    loadData();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// 表格列
const columns = [
  { title: 'ID', dataIndex: 'id', width: 55, align: 'center' as const },
  { title: '模块信息', key: 'info', width: 240 },
  { title: '类型', key: 'type', width: 80, align: 'center' as const,
    filters: [{ text: '运动', value: 'sport' }, { text: '实习', value: 'intern' }, { text: '论文', value: 'paper' }],
    onFilter: (value: string, record: DynamicModule) => record.type === value,
  },
  { title: 'PHP 路径', key: 'paths' },
  { title: '排序', dataIndex: 'sort', width: 65, align: 'center' as const, sorter: (a: DynamicModule, b: DynamicModule) => a.sort - b.sort },
  { title: '状态', key: 'status', width: 80, align: 'center' as const },
  { title: '操作', key: 'action', width: 120, align: 'center' as const },
];

onMounted(() => loadData());
</script>

<template>
  <Page title="模块管理" description="管理运动/实习/论文等动态功能模块" content-class="p-4">
    <Card>
      <Tabs v-model:activeKey="activeTab" size="small" class="mb-2">
        <TabPane key="all">
          <template #tab>全部 <Badge :count="countByType.all" :number-style="{ backgroundColor: '#8c8c8c' }" :overflow-count="999" /></template>
        </TabPane>
        <TabPane key="sport">
          <template #tab>运动 <Badge :count="countByType.sport" :number-style="{ backgroundColor: '#1677ff' }" :overflow-count="999" /></template>
        </TabPane>
        <TabPane key="intern">
          <template #tab>实习 <Badge :count="countByType.intern" :number-style="{ backgroundColor: '#52c41a' }" :overflow-count="999" /></template>
        </TabPane>
        <TabPane key="paper">
          <template #tab>论文 <Badge :count="countByType.paper" :number-style="{ backgroundColor: '#722ed1' }" :overflow-count="999" /></template>
        </TabPane>
      </Tabs>

      <div class="flex justify-between items-center mb-3">
        <Button type="primary" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          添加模块
        </Button>
        <Button @click="loadData" :loading="loading">
          <template #icon><ReloadOutlined /></template>
        </Button>
      </div>

      <Table
        :columns="columns"
        :data-source="filteredData"
        :loading="loading"
        :pagination="false"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'info'">
            <div class="flex items-center gap-2">
              <component :is="iconCompMap[record.icon]" v-if="record.icon && iconCompMap[record.icon]" class="text-lg text-gray-500" />
              <div>
                <div>
                  <span class="font-medium">{{ record.name }}</span>
                  <Typography.Text type="secondary" class="ml-2 text-xs">{{ record.app_id }}</Typography.Text>
                </div>
                <div v-if="record.description || record.price" class="text-xs text-gray-400 mt-0.5">
                  <span v-if="record.description">{{ record.description }}</span>
                  <Tag v-if="record.price" size="small" class="ml-1" color="orange">{{ record.price }}</Tag>
                </div>
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'type'">
            <Tag :color="typeMap[record.type]?.color ?? 'default'">{{ typeMap[record.type]?.label ?? record.type }}</Tag>
          </template>
          <template v-else-if="column.key === 'paths'">
            <div class="text-xs space-y-0.5">
              <div v-if="record.api_base">
                <Tooltip :title="record.api_base">
                  <Tag color="cyan" size="small"><ApiOutlined class="mr-1" />API</Tag>
                  <Typography.Text code class="text-xs">{{ record.api_base }}</Typography.Text>
                </Tooltip>
              </div>
              <div v-if="record.view_url">
                <Tooltip :title="record.view_url">
                  <Tag color="geekblue" size="small"><FileOutlined class="mr-1" />页面</Tag>
                  <Typography.Text code class="text-xs">{{ record.view_url }}</Typography.Text>
                </Tooltip>
              </div>
              <Typography.Text v-if="!record.api_base && !record.view_url" type="secondary" class="text-xs">未配置</Typography.Text>
            </div>
          </template>
          <template v-else-if="column.key === 'status'">
            <Switch
              :checked="record.status === 1"
              checked-children="启用"
              un-checked-children="禁用"
              size="small"
              @change="(checked: boolean) => handleToggle(record, checked)"
            />
          </template>
          <template v-else-if="column.key === 'action'">
            <Space :size="4">
              <Tooltip title="编辑">
                <Button type="text" size="small" @click="openEdit(record)">
                  <template #icon><EditOutlined /></template>
                </Button>
              </Tooltip>
              <Popconfirm title="确定删除此模块？" @confirm="handleDelete(record.id)" ok-text="删除" ok-type="danger">
                <Tooltip title="删除">
                  <Button type="text" size="small" danger>
                    <template #icon><DeleteOutlined /></template>
                  </Button>
                </Tooltip>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      :title="isEdit ? '编辑模块' : '添加模块'"
      width="600px"
      :confirm-loading="editLoading"
      @ok="handleSave"
      ok-text="保存"
    >
      <Form layout="vertical" class="mt-4">
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="模块标识" required>
            <Input v-model:value="editForm.app_id" placeholder="如 yyd, appui" :disabled="isEdit">
              <template #addonBefore>app_id</template>
            </Input>
          </FormItem>
          <FormItem label="模块类型" required>
            <Select v-model:value="editForm.type" placeholder="选择类型">
              <SelectOption v-for="t in typeOptions" :key="t.value" :value="t.value">{{ t.label }}</SelectOption>
            </Select>
          </FormItem>
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="模块名称" required>
            <Input v-model:value="editForm.name" placeholder="如 云运动、APPUI打卡" />
          </FormItem>
          <FormItem label="展示价格">
            <Input v-model:value="editForm.price" placeholder="如 0.5元/次" />
          </FormItem>
        </div>
        <FormItem label="模块描述">
          <Input v-model:value="editForm.description" placeholder="简短描述，显示在大厅卡片上" />
        </FormItem>
        <FormItem label="图标">
          <IconSelect v-model="editForm.icon" />
        </FormItem>
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="API 路径">
            <Input v-model:value="editForm.api_base" placeholder="/模块名/service/api.php">
              <template #prefix><ApiOutlined /></template>
            </Input>
          </FormItem>
          <FormItem label="前端页面路径">
            <Input v-model:value="editForm.view_url" placeholder="/模块名/view/index.php">
              <template #prefix><FileOutlined /></template>
            </Input>
          </FormItem>
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="排序">
            <InputNumber v-model:value="editForm.sort" :min="0" style="width: 100%" placeholder="数字越小越靠前" />
          </FormItem>
          <FormItem label="状态">
            <Select v-model:value="editForm.status">
              <SelectOption :value="1">启用</SelectOption>
              <SelectOption :value="0">禁用</SelectOption>
            </Select>
          </FormItem>
        </div>
      </Form>
    </Modal>
  </Page>
</template>
