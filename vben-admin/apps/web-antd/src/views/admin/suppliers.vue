<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, Switch, message, Select, SelectOption,
} from 'ant-design-vue';
import { PlusOutlined, EditOutlined, ImportOutlined, SyncOutlined, DeleteOutlined, DollarOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import { getSupplierListApi, saveSupplierApi, importSupplierApi, syncSupplierStatusApi, deleteSupplierApi, getPlatformNamesApi, querySupplierBalanceApi, type SupplierItem } from '#/api/admin';

const loading = ref(false);
const suppliers = ref<SupplierItem[]>([]);
const editVisible = ref(false);
const form = reactive({ hid: 0, pt: '', name: '', url: '', user: '', pass: '', token: '', status: '1' });
const platformNames = ref<Record<string, string>>({});

// 一键对接弹窗
const importVisible = ref(false);
const importLoading = ref(false);
const importForm = reactive({ hid: 0, name: '', pricee: 1, category: '999999', fd: 0 });

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
    // 更新本地表格中的余额显示
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
              <Button
                type="link"
                size="small"
                :loading="balanceLoading[record.hid]"
                @click="queryBalance(record.hid)"
                title="刷新余额"
              >
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
              <Button type="link" size="small" @click="handleSyncStatus(record.hid)"><SyncOutlined /> 同步</Button>
              <Button type="link" size="small" danger @click="handleDelete(record.hid)"><DeleteOutlined /> 删除</Button>
            </Space>
          </template>
        </Table.Column>
      </Table>
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
          <label class="block text-sm font-medium mb-1">账号 (user)</label>
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
  </Page>
</template>
