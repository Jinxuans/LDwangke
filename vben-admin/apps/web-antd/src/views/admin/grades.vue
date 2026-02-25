<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal,
  Popconfirm, message, Select, SelectOption,
} from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import { getGradeListApi, saveGradeApi, deleteGradeApi, type GradeItem } from '#/api/admin';

const loading = ref(false);
const grades = ref<GradeItem[]>([]);

// 添加弹窗
const addVisible = ref(false);
const addForm = reactive({
  id: 0, sort: '', name: '', rate: '', money: '', addkf: '1', gjkf: '1', status: '1',
});

// 编辑弹窗
const editVisible = ref(false);
const editForm = reactive({
  id: 0, sort: '', name: '', rate: '', money: '', addkf: '1', gjkf: '1', status: '1',
});

async function loadGrades() {
  loading.value = true;
  try {
    const res = await getGradeListApi();
    grades.value = res;
    if (!Array.isArray(grades.value)) grades.value = [];
  }
  catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function openAdd() {
  Object.assign(addForm, { id: 0, sort: '', name: '', rate: '', money: '', addkf: '1', gjkf: '1', status: '1' });
  addVisible.value = true;
}

function openEdit(g: GradeItem) {
  Object.assign(editForm, {
    id: g.id,
    sort: String(g.sort ?? ''),
    name: g.name,
    rate: String(g.rate ?? ''),
    money: String(g.money ?? ''),
    addkf: String(g.addkf ?? '1'),
    gjkf: String(g.gjkf ?? '1'),
    status: String(g.status ?? '1'),
  });
  editVisible.value = true;
}

async function handleAdd() {
  if (!addForm.name.trim()) { message.error('等级名称不能为空'); return; }
  if (!addForm.rate.trim()) { message.error('等级费率不能为空'); return; }
  try {
    await saveGradeApi({ ...addForm });
    message.success('添加成功');
    addVisible.value = false;
    loadGrades();
  } catch (e: any) { message.error(e?.message || '添加失败'); }
}

async function handleEdit() {
  if (!editForm.name.trim()) { message.error('等级名称不能为空'); return; }
  if (!editForm.rate.trim()) { message.error('等级费率不能为空'); return; }
  try {
    await saveGradeApi({ ...editForm });
    message.success('修改成功');
    editVisible.value = false;
    loadGrades();
  } catch (e: any) { message.error(e?.message || '修改失败'); }
}

async function handleDelete(id: number) {
  try {
    await deleteGradeApi(id);
    message.success('删除成功');
    loadGrades();
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80, align: 'center' as const },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80, align: 'center' as const },
  { title: '等级名称', dataIndex: 'name', key: 'name', width: 160 },
  { title: '等级费率', dataIndex: 'rate', key: 'rate', width: 120, align: 'center' as const },
  { title: '开通价格', dataIndex: 'money', key: 'money', width: 120, align: 'center' as const },
  { title: '添加扣费', key: 'addkf', width: 120, align: 'center' as const },
  { title: '改价扣费', key: 'gjkf', width: 120, align: 'center' as const },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '添加时间', dataIndex: 'time', key: 'time', width: 160, align: 'center' as const },
  { title: '操作', key: 'action', width: 160, align: 'center' as const },
];

onMounted(loadGrades);
</script>

<template>
  <Page title="等级管理" content-class="p-4">
    <Card title="等级列表">
      <template #extra>
        <Button type="primary" size="small" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          添加
        </Button>
      </template>

      <Table :columns="columns" :data-source="grades" :loading="loading" :pagination="false" row-key="id" size="small" bordered :scroll="{ x: 1100 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'addkf'">
            <Tag :color="String(record.addkf) === '1' ? 'success' : 'error'">
              {{ String(record.addkf) === '1' ? '打开' : '关闭' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'gjkf'">
            <Tag :color="String(record.gjkf) === '1' ? 'success' : 'error'">
              {{ String(record.gjkf) === '1' ? '打开' : '关闭' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="String(record.status) === '1' ? 'success' : 'error'">
              {{ String(record.status) === '1' ? '已启用' : '未启用' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small">
              <Button type="primary" size="small" @click="openEdit(record)">编辑</Button>
              <Popconfirm title="确定删除吗？" @confirm="handleDelete(record.id)">
                <Button danger size="small">删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加弹窗 -->
    <Modal v-model:open="addVisible" title="等级添加" @ok="handleAdd" ok-text="确定" cancel-text="取消" :width="620" style="max-width: 95vw">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">排序</label>
          <Input v-model:value="addForm.sort" placeholder="请输入排序" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">等级名称</label>
          <Input v-model:value="addForm.name" placeholder="请输入等级名称" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">等级费率</label>
          <Input v-model:value="addForm.rate" placeholder="例如 0.95" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">开通价格</label>
          <Input v-model:value="addForm.money" placeholder="例如 10" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">添加用户扣费</label>
          <Select v-model:value="addForm.addkf" style="width: 100%">
            <SelectOption value="1">打开</SelectOption>
            <SelectOption value="0">关闭</SelectOption>
          </Select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">修改费率扣费</label>
          <Select v-model:value="addForm.gjkf" style="width: 100%">
            <SelectOption value="1">打开</SelectOption>
            <SelectOption value="0">关闭</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="等级修改" @ok="handleEdit" ok-text="确定" cancel-text="取消" :width="620" style="max-width: 95vw">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">排序</label>
          <Input v-model:value="editForm.sort" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">等级名称</label>
          <Input v-model:value="editForm.name" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">等级费率</label>
          <Input v-model:value="editForm.rate" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">开通价格</label>
          <Input v-model:value="editForm.money" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">添加用户扣费</label>
          <Select v-model:value="editForm.addkf" style="width: 100%">
            <SelectOption value="1">打开</SelectOption>
            <SelectOption value="0">关闭</SelectOption>
          </Select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">修改费率扣费</label>
          <Select v-model:value="editForm.gjkf" style="width: 100%">
            <SelectOption value="1">打开</SelectOption>
            <SelectOption value="0">关闭</SelectOption>
          </Select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">状态</label>
          <Select v-model:value="editForm.status" style="width: 100%">
            <SelectOption value="1">启用</SelectOption>
            <SelectOption value="0">关闭</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>
  </Page>
</template>
