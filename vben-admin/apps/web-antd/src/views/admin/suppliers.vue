<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message, Select, SelectOption, Tabs, TabPane, Switch, Form, FormItem
} from 'ant-design-vue';
import { PlusOutlined, EditOutlined, ImportOutlined, SyncOutlined, DeleteOutlined, DollarOutlined, ReloadOutlined, ApiOutlined, CloudDownloadOutlined, ControlOutlined, CopyOutlined } from '@ant-design/icons-vue';
import { getSupplierListApi, saveSupplierApi, importSupplierApi, syncSupplierStatusApi, deleteSupplierApi, getPlatformNamesApi, querySupplierBalanceApi, type SupplierItem, cloneFromUpstreamApi, updatePricesApi, autoSyncApi } from '#/api/admin';
import PlatformConfigPanel from './components/PlatformConfigPanel.vue';

const activeTab = ref('suppliers');
const loading = ref(false);
const suppliers = ref<SupplierItem[]>([]);
const editVisible = ref(false);
const form = reactive({ hid: 0, pt: '', name: '', url: '', user: '', pass: '', token: '', status: '1' });
const platformNames = ref<Record<string, string>>({});

const importVisible = ref(false);
const importLoading = ref(false);
const importForm = reactive({ hid: 0, name: '', pricee: 1, category: '999999', fd: 0 });

const cloneVisible = ref(false);
const cloneLoading = ref(false);
const cloneForm = reactive({
  hid: 0,
  name: '',
  price_rate: 5,
  clone_category: false,
  skip_categories: '',
  secret_price_rate: 0
});

async function loadSuppliers() {
  loading.value = true;
  try {
    const res = await getSupplierListApi();
    suppliers.value = res;
    if (!Array.isArray(suppliers.value)) suppliers.value = [];
  }
  catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function openEdit(sup?: SupplierItem) {
  if (sup) {
    Object.assign(form, { hid: sup.hid, pt: sup.pt || '', name: sup.name, url: sup.url, user: sup.user, pass: sup.pass, token: sup.token, status: sup.status });
  } else {
    Object.assign(form, { hid: 0, pt: '', name: '', url: '', user: '', pass: '', token: '', status: '1' });
  }
  editVisible.value = true;
}

async function handleSave() {
  if (!form.name.trim()) { message.warning('请填写货源名称'); return; }
  try {
    await saveSupplierApi({ ...form });
    message.success('保存成功');
    editVisible.value = false;
    loadSuppliers();
  } catch (e: any) { message.error(e?.message || '保存失败'); }
}

function openImport(sup: SupplierItem) {
  Object.assign(importForm, { hid: sup.hid, name: sup.name, pricee: 1, category: '999999', fd: 0 });
  importVisible.value = true;
}

async function handleImport() {
  importLoading.value = true;
  try {
    const raw = await importSupplierApi({
      hid: importForm.hid,
      pricee: importForm.pricee,
      category: importForm.category,
      name: importForm.name,
      fd: importForm.fd,
    });
    const res = raw;
    message.success(res.msg || '对接成功');
    importVisible.value = false;
  } catch (e: any) { message.error(e?.message || '对接失败'); }
  finally { importLoading.value = false; }
}

function openClone(sup: SupplierItem) {
  Object.assign(cloneForm, {
    hid: sup.hid,
    name: sup.name,
    price_rate: 5,
    clone_category: false,
    skip_categories: '',
    secret_price_rate: 0
  });
  cloneVisible.value = true;
}

async function handleClone() {
  cloneLoading.value = true;
  try {
    const payload = {
      hid: cloneForm.hid,
      price_rate: cloneForm.price_rate,
      clone_category: cloneForm.clone_category,
      skip_categories: cloneForm.skip_categories ? cloneForm.skip_categories.split(',').map(s => s.trim()) : [],
      secret_price_rate: cloneForm.secret_price_rate
    };
    const res: any = await cloneFromUpstreamApi(payload);
    message.success(res?.message || '克隆完成');
    cloneVisible.value = false;
  } catch (e: any) { message.error(e?.message || '克隆失败'); }
  finally { cloneLoading.value = false; }
}

async function handleUpdatePrices() {
  cloneLoading.value = true;
  try {
    const payload = {
      hid: cloneForm.hid,
      price_rate: cloneForm.price_rate,
      skip_categories: cloneForm.skip_categories ? cloneForm.skip_categories.split(',').map(s => s.trim()) : []
    };
    const res: any = await updatePricesApi(payload);
    message.success(res?.message || '价格更新完成');
    cloneVisible.value = false;
  } catch (e: any) { message.error(e?.message || '更新价格失败'); }
  finally { cloneLoading.value = false; }
}

async function handleAutoSync(hid: number, name: string) {
  Modal.confirm({
    title: `确认自动同步上架: ${name}`,
    content: '此操作将自动识别上游分类，同步新增商品并下架失效商品。价格将按照默认5倍同步（或参考已有商品的比例）。',
    async onOk() {
      const hide = message.loading('正在同步中...', 0);
      try {
        const res: any = await autoSyncApi({ hid, price_rate: 5 });
        hide();
        message.success(res?.message || '同步完成');
      } catch (e: any) {
        hide();
        message.error(e?.message || '同步失败');
      }
    }
  });
}

async function handleSyncStatus(hid: number) {
  try {
    const raw = await syncSupplierStatusApi(hid);
    const res = raw;
    message.success(res.msg || '同步成功');
  } catch (e: any) { message.error(e?.message || '同步失败'); }
}

async function handleDelete(hid: number) {
  Modal.confirm({
    title: '确认删除',
    content: '删除后不可恢复，确定要删除此货源吗？',
    async onOk() {
      try {
        await deleteSupplierApi(hid);
        message.success('删除成功');
        loadSuppliers();
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

function getPtName(pt: string) {
  return platformNames.value[pt] || pt || '-';
}

const balanceLoading = ref<Record<number, boolean>>({});
const batchBalanceLoading = ref(false);

async function queryBalance(hid: number) {
  balanceLoading.value[hid] = true;
  try {
    const raw = await querySupplierBalanceApi(hid);
    const res = raw;
    const money = res?.money ?? '';
    const sup = suppliers.value.find((s) => s.hid === hid);
    if (sup) sup.money = money;
    message.success(`${sup?.name || hid}: 余额 ${money}`);
  } catch (e: any) {
    message.error(`查询余额失败: ${e?.message || '未知错误'}`);
  } finally {
    balanceLoading.value[hid] = false;
  }
}

async function batchQueryBalance() {
  batchBalanceLoading.value = true;
  let ok = 0, fail = 0;
  for (const sup of suppliers.value) {
    if (sup.status !== '1') continue;
    try {
      balanceLoading.value[sup.hid] = true;
      const raw = await querySupplierBalanceApi(sup.hid);
      const res = raw;
      sup.money = res?.money ?? sup.money;
      ok++;
    } catch {
      fail++;
    } finally {
      balanceLoading.value[sup.hid] = false;
    }
  }
  batchBalanceLoading.value = false;
  message.success(`批量查询完成：成功 ${ok}，失败 ${fail}`);
}

onMounted(async () => {
  loadSuppliers();
  try {
    const res = await getPlatformNamesApi();
    platformNames.value = res;
  } catch (e) { /* ignore */ }
});
</script>

<template>
  <Page title="货源管理" content-class="p-4">
    <Card>
      <Tabs v-model:activeKey="activeTab">
        <!-- ===== Tab 1: 货源列表 ===== -->
        <TabPane key="suppliers">
          <template #tab>
            <span class="font-medium">货源列表</span>
          </template>

          <div class="flex justify-between items-center mb-4">
            <span class="text-sm text-gray-500">共 {{ suppliers.length }} 个货源</span>
            <Space>
              <Button :loading="batchBalanceLoading" @click="batchQueryBalance">
                <template #icon><DollarOutlined /></template>
                批量查余额
              </Button>
              <Button type="primary" @click="openEdit()">
                <template #icon><PlusOutlined /></template>
                添加货源
              </Button>
            </Space>
          </div>

          <Table :data-source="suppliers" :loading="loading" :pagination="false" row-key="hid" size="small" bordered :scroll="{ x: 1100 }">
            <Table.Column title="HID" data-index="hid" :width="60" />
            <Table.Column title="平台" :width="100">
              <template #default="{ record }">
                <Tag color="blue">{{ getPtName(record.pt) }}</Tag>
              </template>
            </Table.Column>
            <Table.Column title="名称" data-index="name" :width="150" />
            <Table.Column title="API 地址" data-index="url" ellipsis />
            <Table.Column title="账号" data-index="user" :width="120" />
            <Table.Column title="余额" :width="140">
              <template #default="{ record }">
                <div class="flex items-center gap-1">
                  <span>{{ record.money || '-' }}</span>
                  <Button type="link" size="small" :loading="balanceLoading[record.hid]" @click="queryBalance(record.hid)" title="刷新余额">
                    <template #icon><ReloadOutlined /></template>
                  </Button>
                </div>
              </template>
            </Table.Column>
            <Table.Column title="状态" :width="80">
              <template #default="{ record }">
                <Tag :color="record.status === '1' ? 'green' : 'default'">{{ record.status === '1' ? '启用' : '禁用' }}</Tag>
              </template>
            </Table.Column>
            <Table.Column title="添加时间" data-index="addtime" :width="160" />
            <Table.Column title="操作" :width="220">
              <template #default="{ record }">
                <Space>
                  <Button type="link" size="small" @click="openEdit(record)"><EditOutlined /> 编辑</Button>
                  <Button type="link" size="small" @click="openImport(record)"><ImportOutlined /> 对接</Button>
                  <Button type="link" size="small" @click="openClone(record)"><CopyOutlined /> 克隆</Button>
                  <Button type="link" size="small" @click="handleAutoSync(record.hid, record.name)"><SyncOutlined /> 自动识别上架</Button>
                  <Button type="link" size="small" @click="handleSyncStatus(record.hid)"><SyncOutlined /> 同步上下架</Button>
                  <Button type="link" danger size="small" @click="handleDelete(record.hid)"><DeleteOutlined /> 删除</Button>
                </Space>
              </template>
            </Table.Column>
          </Table>
        </TabPane>

        <!-- ===== Tab 2: 平台接口配置 ===== -->
        <TabPane key="platform-config">
          <template #tab>
            <span class="font-medium"><ApiOutlined class="mr-1" />平台接口配置</span>
          </template>
          <PlatformConfigPanel />
        </TabPane>
      </Tabs>
    </Card>

    <!-- 编辑货源 -->
    <Modal v-model:open="editVisible" :title="form.hid ? '编辑货源' : '添加货源'" @ok="handleSave" ok-text="保存">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">平台类型</label>
          <Select v-model:value="form.pt" style="width: 100%" placeholder="选择平台类型" allow-clear show-search>
            <SelectOption v-for="(label, key) in platformNames" :key="key" :value="key">{{ label }} ({{ key }})</SelectOption>
          </Select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">名称</label>
          <Input v-model:value="form.name" placeholder="货源名称" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">API 地址</label>
          <Input v-model:value="form.url" placeholder="http://..." />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">账号</label>
          <Input v-model:value="form.user" placeholder="上游账号" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">密钥 (pass/key)</label>
          <Input v-model:value="form.pass" placeholder="上游密钥" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Token（可选）</label>
          <Input v-model:value="form.token" placeholder="留空即可" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">状态</label>
          <Select v-model:value="form.status" style="width: 100%">
            <SelectOption value="1">启用</SelectOption>
            <SelectOption value="0">禁用</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>

    <!-- 一键对接弹窗 -->
    <Modal v-model:open="importVisible" title="一键对接" :confirm-loading="importLoading" @ok="handleImport" ok-text="开始对接">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">货源</label>
          <Input :value="importForm.name" disabled />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">价格倍率</label>
          <InputNumber v-model:value="importForm.pricee" :min="0.01" :max="100" :step="0.1" :precision="2" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">分类筛选（留空=全部）</label>
          <Input v-model:value="importForm.category" placeholder="上游分类ID，999999=全部" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">分类名称（新建分类用）</label>
          <Input v-model:value="importForm.name" placeholder="留空则使用货源名称" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">模式</label>
          <Select v-model:value="importForm.fd" style="width: 100%">
            <SelectOption :value="0">全量（新增+更新）</SelectOption>
            <SelectOption :value="1">仅更新已有</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>
    <!-- 一键克隆弹窗 -->
    <Modal v-model:open="cloneVisible" title="克隆 / 更新 / 自动同步" :confirm-loading="cloneLoading">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">货源</label>
          <Input :value="cloneForm.name" disabled />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">价格倍率</label>
          <InputNumber v-model:value="cloneForm.price_rate" :min="0.01" :step="0.1" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">密价比例 (0代表关闭)</label>
          <InputNumber v-model:value="cloneForm.secret_price_rate" :min="0" :step="0.1" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">是否同步分类</label>
          <Switch v-model:checked="cloneForm.clone_category" checked-children="开启" un-checked-children="关闭" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">跳过的分类ID（逗号分隔）</label>
          <Input v-model:value="cloneForm.skip_categories" placeholder="如：1,2,3" />
        </div>
      </div>
      <template #footer>
        <Button @click="cloneVisible = false">取消</Button>
        <Button type="primary" @click="handleUpdatePrices" :loading="cloneLoading">仅更新价格</Button>
        <Button type="primary" danger @click="handleClone" :loading="cloneLoading">一键克隆上架</Button>
      </template>
    </Modal>
  </Page>
</template>
