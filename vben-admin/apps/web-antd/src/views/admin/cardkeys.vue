<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, InputNumber, Space, Tag, Modal,
  Popconfirm, message, Pagination, Textarea,
} from 'ant-design-vue';
import { PlusOutlined, DeleteOutlined, CopyOutlined } from '@ant-design/icons-vue';
import {
  getCardKeyListApi, generateCardKeysApi, deleteCardKeysApi,
  type CardKey,
} from '#/api/auxiliary';

const loading = ref(false);
const list = ref<CardKey[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const selectedRowKeys = ref<number[]>([]);

// 生成弹窗
const genVisible = ref(false);
const genForm = reactive({ money: 10, count: 10 });
const generating = ref(false);
const generatedCodes = ref<string[]>([]);
const showCodesVisible = ref(false);

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70, align: 'center' as const },
  { title: '卡密内容', dataIndex: 'content', key: 'content', ellipsis: true },
  { title: '面额(元)', dataIndex: 'money', key: 'money', width: 100, align: 'center' as const },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '使用者', key: 'uid', width: 100, align: 'center' as const },
  { title: '创建时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
  { title: '使用时间', dataIndex: 'usedtime', key: 'usedtime', width: 170 },
];

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const res = await getCardKeyListApi({ page: pagination.page, limit: pagination.limit });
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function openGen() {
  Object.assign(genForm, { money: 10, count: 10 });
  genVisible.value = true;
}

async function handleGenerate() {
  if (genForm.money < 1) { message.error('面额至少1元'); return; }
  if (genForm.count < 1 || genForm.count > 100) { message.error('数量1-100张'); return; }
  generating.value = true;
  try {
    const res = await generateCardKeysApi(genForm.money, genForm.count);
    generatedCodes.value = res.codes || [];
    message.success(`成功生成 ${res.count} 张卡密`);
    genVisible.value = false;
    showCodesVisible.value = true;
    loadData(1);
  } catch (e: any) { message.error(e?.message || '生成失败'); }
  finally { generating.value = false; }
}

async function handleBatchDelete() {
  if (selectedRowKeys.value.length === 0) { message.warning('请选择要删除的卡密'); return; }
  try {
    const res = await deleteCardKeysApi(selectedRowKeys.value);
    message.success(`成功删除 ${res.deleted} 张未使用的卡密`);
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

function copyCodes() {
  const text = generatedCodes.value.join('\n');
  navigator.clipboard.writeText(text).then(() => message.success('已复制到剪贴板'));
}

const rowSelection = {
  selectedRowKeys,
  onChange: (keys: number[]) => { selectedRowKeys.value = keys; },
  getCheckboxProps: (record: CardKey) => ({ disabled: record.status === 1 }),
};

onMounted(() => loadData());
</script>

<template>
  <Page title="卡密管理" content-class="p-4">
    <Card title="卡密列表">
      <template #extra>
        <Space>
          <Popconfirm title="确定删除选中的未使用卡密？" @confirm="handleBatchDelete" :disabled="selectedRowKeys.length === 0">
            <Button danger size="small" :disabled="selectedRowKeys.length === 0">
              <template #icon><DeleteOutlined /></template>
              批量删除 ({{ selectedRowKeys.length }})
            </Button>
          </Popconfirm>
          <Button type="primary" size="small" @click="openGen">
            <template #icon><PlusOutlined /></template>
            生成卡密
          </Button>
        </Space>
      </template>

      <Table
        :columns="columns" :data-source="list" :loading="loading"
        :pagination="false" row-key="id" size="small" bordered
        :row-selection="rowSelection"
        :scroll="{ x: 900 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="record.status === 1 ? 'default' : 'success'">
              {{ record.status === 1 ? '已使用' : '未使用' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'uid'">
            {{ record.uid ? `UID: ${record.uid}` : '-' }}
          </template>
        </template>
      </Table>

      <div class="flex justify-end mt-4" v-if="pagination.total > pagination.limit">
        <Pagination
          :current="pagination.page" :page-size="pagination.limit" :total="pagination.total"
          show-size-changer :page-size-options="['20','50','100']"
          @change="(p: number) => loadData(p)"
          @showSizeChange="(_c: number, s: number) => { pagination.limit = s; loadData(1); }"
        />
      </div>
    </Card>

    <!-- 生成弹窗 -->
    <Modal v-model:open="genVisible" title="生成卡密" @ok="handleGenerate" :confirm-loading="generating"
           ok-text="生成" cancel-text="取消" :width="420" style="max-width: 95vw">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">面额（元）</label>
          <InputNumber v-model:value="genForm.money" :min="1" :max="10000" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">数量（张）</label>
          <InputNumber v-model:value="genForm.count" :min="1" :max="100" style="width: 100%" />
        </div>
      </div>
    </Modal>

    <!-- 生成结果 -->
    <Modal v-model:open="showCodesVisible" title="生成成功" :footer="null" :width="520" style="max-width: 95vw">
      <div class="space-y-3">
        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-500">共 {{ generatedCodes.length }} 张卡密</span>
          <Button size="small" @click="copyCodes">
            <template #icon><CopyOutlined /></template>
            复制全部
          </Button>
        </div>
        <Textarea :value="generatedCodes.join('\n')" :rows="8" readonly style="font-family: monospace" />
      </div>
    </Modal>
  </Page>
</template>
