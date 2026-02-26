<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal,
  Popconfirm, message, Pagination, Select, SelectOption, DatePicker,
} from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import {
  getActivityListApi, saveActivityApi, deleteActivityApi,
  type Activity,
} from '#/api/auxiliary';

const loading = ref(false);
const list = ref<Activity[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });

const modalVisible = ref(false);
const saving = ref(false);
const editForm = reactive({
  hid: 0, name: '', yaoqiu: '', type: '1', num: '', money: '',
  addtime: '', endtime: '', status_ok: '1',
});

const columns = [
  { title: 'ID', dataIndex: 'hid', key: 'hid', width: 70, align: 'center' as const },
  { title: '活动名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '类型', key: 'type', width: 110, align: 'center' as const },
  { title: '要求', dataIndex: 'yaoqiu', key: 'yaoqiu', ellipsis: true },
  { title: '数量', dataIndex: 'num', key: 'num', width: 80, align: 'center' as const },
  { title: '奖励(元)', dataIndex: 'money', key: 'money', width: 100, align: 'center' as const },
  { title: '开始时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  { title: '结束时间', dataIndex: 'endtime', key: 'endtime', width: 160 },
  { title: '状态', key: 'status_ok', width: 80, align: 'center' as const },
  { title: '操作', key: 'action', width: 140, align: 'center' as const, fixed: 'right' as const },
];

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const res = await getActivityListApi({ page: pagination.page, limit: pagination.limit });
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function openAdd() {
  Object.assign(editForm, {
    hid: 0, name: '', yaoqiu: '', type: '1', num: '', money: '',
    addtime: '', endtime: '', status_ok: '1',
  });
  modalVisible.value = true;
}

function openEdit(record: Activity) {
  Object.assign(editForm, {
    hid: record.hid, name: record.name, yaoqiu: record.yaoqiu,
    type: record.type, num: record.num, money: record.money,
    addtime: record.addtime, endtime: record.endtime, status_ok: record.status_ok,
  });
  modalVisible.value = true;
}

async function handleSave() {
  if (!editForm.name.trim()) { message.warning('请填写活动名称'); return; }
  if (!editForm.num.trim() || !editForm.money.trim()) { message.warning('请填写要求数量和奖励金额'); return; }
  saving.value = true;
  try {
    await saveActivityApi({ ...editForm });
    message.success('保存成功');
    modalVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '保存失败'); }
  finally { saving.value = false; }
}

async function handleDelete(hid: number) {
  try {
    await deleteActivityApi(hid);
    message.success('删除成功');
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

onMounted(() => loadData());
</script>

<template>
  <Page title="活动管理" content-class="p-4">
    <Card title="活动列表">
      <template #extra>
        <Button type="primary" size="small" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          新建活动
        </Button>
      </template>

      <Table
        :columns="columns" :data-source="list" :loading="loading"
        :pagination="false" row-key="hid" size="small" bordered
        :scroll="{ x: 1100 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <Tag :color="record.type === '1' ? 'blue' : 'green'">
              {{ record.type === '1' ? '邀人活动' : '订单活动' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'status_ok'">
            <Tag :color="record.status_ok === '1' ? 'success' : 'default'">
              {{ record.status_ok === '1' ? '进行中' : '已结束' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small">
              <Button type="primary" size="small" @click="openEdit(record)">
                <template #icon><EditOutlined /></template>
              </Button>
              <Popconfirm title="确定删除此活动？" @confirm="handleDelete(record.hid)">
                <Button danger size="small">
                  <template #icon><DeleteOutlined /></template>
                </Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <div class="flex justify-end mt-4" v-if="pagination.total > pagination.limit">
        <Pagination
          :current="pagination.page" :page-size="pagination.limit" :total="pagination.total"
          @change="(p: number) => loadData(p)"
        />
      </div>
    </Card>

    <!-- 编辑/新建弹窗 -->
    <Modal v-model:open="modalVisible" :title="editForm.hid ? '编辑活动' : '新建活动'"
           @ok="handleSave" :confirm-loading="saving" ok-text="保存" cancel-text="取消"
           :width="560" style="max-width: 95vw">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">活动名称</label>
          <Input v-model:value="editForm.name" placeholder="请输入活动名称" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">活动类型</label>
          <Select v-model:value="editForm.type" style="width: 100%">
            <SelectOption value="1">邀人活动</SelectOption>
            <SelectOption value="2">订单活动</SelectOption>
          </Select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">要求描述</label>
          <Input v-model:value="editForm.yaoqiu" placeholder="例如：邀请10个用户注册" />
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">要求数量</label>
            <Input v-model:value="editForm.num" placeholder="例如 10" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">奖励金额(元)</label>
            <Input v-model:value="editForm.money" placeholder="例如 50" />
          </div>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">开始时间</label>
            <Input v-model:value="editForm.addtime" placeholder="2026-01-01 00:00:00" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">结束时间</label>
            <Input v-model:value="editForm.endtime" placeholder="2026-12-31 23:59:59" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">状态</label>
          <Select v-model:value="editForm.status_ok" style="width: 100%">
            <SelectOption value="1">进行中</SelectOption>
            <SelectOption value="2">已结束</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>
  </Page>
</template>
