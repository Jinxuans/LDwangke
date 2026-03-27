<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { Page } from '@vben/common-ui';
import { Button, Card, Input, InputNumber, Modal, Space, Switch, Table, Tag, message } from 'ant-design-vue';
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons-vue';
import Sortable from 'sortablejs';
import {
  deleteTenantMallCategoryApi,
  getTenantMallCategoriesApi,
  saveTenantMallCategoryApi,
  updateTenantMallCategorySortApi,
  type TenantMallCategory,
} from '#/api/tenant';

const loading = ref(false);
const list = ref<TenantMallCategory[]>([]);
const visible = ref(false);
const listRef = ref<HTMLElement | null>(null);
let sortableInstance: Sortable | null = null;

const form = reactive<{
  id: number;
  name: string;
  sort: number;
  status: number;
}>({
  id: 0,
  name: '',
  sort: 10,
  status: 1,
});

function resetForm() {
  Object.assign(form, {
    id: 0,
    name: '',
    sort: 10,
    status: 1,
  });
}

async function load() {
  loading.value = true;
  try {
    const res = await getTenantMallCategoriesApi();
    list.value = Array.isArray(res) ? res : [];
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function initSortable() {
  if (sortableInstance) {
    sortableInstance.destroy();
    sortableInstance = null;
  }
  nextTick(() => {
    const tbody = listRef.value?.querySelector('tbody');
    if (!tbody) return;
    sortableInstance = Sortable.create(tbody as HTMLElement, {
      handle: '.drag-handle',
      animation: 150,
      async onEnd(evt) {
        const { oldIndex, newIndex } = evt;
        if (oldIndex == null || newIndex == null || oldIndex === newIndex) return;
        const moved = list.value.splice(oldIndex, 1)[0];
        if (!moved) return;
        list.value.splice(newIndex, 0, moved);
        try {
          await updateTenantMallCategorySortApi(
            list.value.map((item, index) => ({
              id: item.id,
              sort: index + 1,
            })),
          );
          list.value = list.value.map((item, index) => ({ ...item, sort: index + 1 }));
          message.success('排序更新成功');
        } catch (e: any) {
          message.error(e?.message || '排序更新失败');
          await load();
        }
      },
    });
  });
}

function openAdd() {
  resetForm();
  visible.value = true;
}

function openEdit(item: TenantMallCategory) {
  Object.assign(form, {
    id: item.id,
    name: item.name,
    sort: item.sort,
    status: item.status,
  });
  visible.value = true;
}

async function submit() {
  if (!form.name.trim()) {
    message.warning('请输入分类名称');
    return;
  }
  try {
    await saveTenantMallCategoryApi({ ...form, name: form.name.trim() });
    message.success('保存成功');
    visible.value = false;
    await load();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  }
}

function handleDelete(id: number) {
  Modal.confirm({
    title: '确认删除',
    content: '删除后不可恢复；若分类下仍有商品会被拦截。',
    async onOk() {
      try {
        await deleteTenantMallCategoryApi(id);
        message.success('删除成功');
        await load();
      } catch (e: any) {
        message.error(e?.message || '删除失败');
      }
    },
  });
}

watch(list, () => initSortable());

onMounted(load);

onBeforeUnmount(() => {
  if (sortableInstance) {
    sortableInstance.destroy();
    sortableInstance = null;
  }
});
</script>

<template>
  <Page title="商城分类" content-class="p-4">
    <Card>
      <div class="mb-4 flex items-center justify-between">
        <div class="text-sm text-gray-500">
          共 {{ list.length }} 个分类。选品时会直接复用这里的分类，后续可继续扩展分类排序和展示。
        </div>
        <Button type="primary" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          添加分类
        </Button>
      </div>

      <div ref="listRef">
        <Table
          :data-source="list"
          :loading="loading"
          :pagination="false"
          row-key="id"
          size="small"
          bordered
        >
          <Table.Column title="排序" :width="70">
            <template #default>
              <div class="drag-handle cursor-move text-gray-400 hover:text-gray-600">⋮⋮</div>
            </template>
          </Table.Column>
          <Table.Column title="ID" data-index="id" :width="80" />
          <Table.Column title="分类名称" data-index="name" />
          <Table.Column title="序号" data-index="sort" :width="100" />
          <Table.Column title="状态" :width="100">
            <template #default="{ record }">
              <Tag :color="record.status === 1 ? 'green' : 'default'">
                {{ record.status === 1 ? '启用' : '禁用' }}
              </Tag>
            </template>
          </Table.Column>
          <Table.Column title="添加时间" data-index="addtime" :width="180" />
          <Table.Column title="操作" :width="150">
            <template #default="{ record }">
              <Space>
                <Button type="link" size="small" @click="openEdit(record)">
                  <EditOutlined />
                  编辑
                </Button>
                <Button type="link" size="small" danger @click="handleDelete(record.id)">
                  <DeleteOutlined />
                  删除
                </Button>
              </Space>
            </template>
          </Table.Column>
        </Table>
      </div>
    </Card>

    <Modal
      v-model:open="visible"
      :title="form.id ? '编辑分类' : '添加分类'"
      ok-text="保存"
      @ok="submit"
    >
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium">分类名称</label>
          <Input v-model:value="form.name" placeholder="例如：成人教育、热门推荐" allow-clear />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">排序</label>
          <InputNumber v-model:value="form.sort" :min="0" style="width: 100%" />
        </div>
        <div class="flex items-center gap-3">
          <label class="text-sm font-medium">状态</label>
          <Switch
            v-model:checked="form.status"
            :checked-value="1"
            :un-checked-value="0"
            checked-children="启用"
            un-checked-children="禁用"
          />
        </div>
      </div>
    </Modal>
  </Page>
</template>
