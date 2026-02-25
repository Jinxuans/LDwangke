<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Table, Button, InputNumber, Space, Tag, Modal, message, Switch } from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import { getTenantProductsApi, saveTenantProductApi, deleteTenantProductApi, type TenantProduct } from '#/api/tenant';

const loading = ref(false);
const products = ref<TenantProduct[]>([]);
const editVisible = ref(false);
const form = reactive<Partial<TenantProduct>>({ id: 0, cid: 0, retail_price: 0, sort: 0, status: 1 });

async function load() {
  loading.value = true;
  try {
    const res = await getTenantProductsApi();
    products.value = res;
    if (!Array.isArray(products.value)) products.value = [];
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function openEdit(p?: TenantProduct) {
  if (p) {
    Object.assign(form, { id: p.id, cid: p.cid, retail_price: p.retail_price, sort: p.sort, status: p.status });
  } else {
    Object.assign(form, { id: 0, cid: 0, retail_price: 0, sort: 0, status: 1 });
  }
  editVisible.value = true;
}

async function handleSave() {
  if (!form.cid) { message.warning('请填写课程ID'); return; }
  if (!form.retail_price || form.retail_price <= 0) { message.warning('请填写零售价'); return; }
  try {
    await saveTenantProductApi({ ...form });
    message.success('保存成功');
    editVisible.value = false;
    load();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  }
}

function handleDelete(cid: number) {
  Modal.confirm({
    title: '确认下架',
    content: '下架后C端将无法购买此商品，确定继续？',
    async onOk() {
      try {
        await deleteTenantProductApi(cid);
        message.success('已下架');
        load();
      } catch (e: any) {
        message.error(e?.message || '操作失败');
      }
    },
  });
}

onMounted(load);
</script>

<template>
  <Page title="选品管理" content-class="p-4">
    <Card>
      <div class="flex justify-between items-center mb-4">
        <span class="text-sm text-gray-500">共 {{ products.length }} 个商品</span>
        <Button type="primary" @click="openEdit()">
          <template #icon><PlusOutlined /></template>
          添加商品
        </Button>
      </div>

      <Table :data-source="products" :loading="loading" :pagination="{ pageSize: 20 }" row-key="id" size="small" bordered>
        <Table.Column title="课程ID" data-index="cid" :width="80" />
        <Table.Column title="课程名称" data-index="class_name" ellipsis />
        <Table.Column title="供货价" :width="100">
          <template #default="{ record }">
            ¥{{ record.class_price }}
          </template>
        </Table.Column>
        <Table.Column title="零售价" :width="100">
          <template #default="{ record }">
            <span class="text-red-500 font-medium">¥{{ record.retail_price }}</span>
          </template>
        </Table.Column>
        <Table.Column title="排序" data-index="sort" :width="70" />
        <Table.Column title="状态" :width="80">
          <template #default="{ record }">
            <Tag :color="record.status === 1 ? 'green' : 'default'">{{ record.status === 1 ? '上架' : '下架' }}</Tag>
          </template>
        </Table.Column>
        <Table.Column title="操作" :width="140">
          <template #default="{ record }">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)"><EditOutlined /> 编辑</Button>
              <Button type="link" size="small" danger @click="handleDelete(record.cid)"><DeleteOutlined /> 下架</Button>
            </Space>
          </template>
        </Table.Column>
      </Table>
    </Card>

    <Modal v-model:open="editVisible" :title="form.id ? '编辑商品' : '添加商品'" @ok="handleSave" ok-text="保存">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">课程ID（CID）</label>
          <InputNumber v-model:value="form.cid" :min="1" style="width: 100%" placeholder="填写平台课程ID" :disabled="!!form.id" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">零售价（元）</label>
          <InputNumber v-model:value="form.retail_price" :min="0.01" :precision="2" :step="1" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">排序（数字越小越靠前）</label>
          <InputNumber v-model:value="form.sort" :min="0" style="width: 100%" />
        </div>
        <div class="flex items-center gap-3">
          <label class="text-sm font-medium">状态</label>
          <Switch v-model:checked="form.status" :checked-value="1" :un-checked-value="0" checked-children="上架" un-checked-children="下架" />
        </div>
      </div>
    </Modal>
  </Page>
</template>
