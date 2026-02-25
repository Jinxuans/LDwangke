<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Tag, Space, Input, Modal, Switch, message,
  Pagination, Popconfirm,
} from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined } from '@ant-design/icons-vue';
import {
  getAnnouncementListApi,
  saveAnnouncementApi,
  deleteAnnouncementApi,
  type AnnouncementItem,
} from '#/api/admin';

const loading = ref(false);
const list = ref<AnnouncementItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const keyword = ref('');

// 编辑弹窗
const modalVisible = ref(false);
const saving = ref(false);
const editForm = reactive({
  id: 0,
  title: '',
  content: '',
  status: '1',
  zhiding: '0',
});

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '作者', dataIndex: 'author', key: 'author', width: 100 },
  { title: '状态', key: 'status', width: 80, align: 'center' as const },
  { title: '置顶', key: 'zhiding', width: 80, align: 'center' as const },
  { title: '时间', dataIndex: 'time', key: 'time', width: 170 },
  { title: '操作', key: 'action', width: 140, align: 'center' as const },
];

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getAnnouncementListApi({
      page: pagination.page,
      limit: pagination.limit,
      keyword: keyword.value || undefined,
    });
    const res = raw;
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error('加载公告失败:', e);
  } finally {
    loading.value = false;
  }
}

function openAdd() {
  Object.assign(editForm, { id: 0, title: '', content: '', status: '1', zhiding: '0' });
  modalVisible.value = true;
}

function openEdit(record: AnnouncementItem) {
  Object.assign(editForm, {
    id: record.id,
    title: record.title,
    content: record.content,
    status: record.status,
    zhiding: record.zhiding,
  });
  modalVisible.value = true;
}

async function handleSave() {
  if (!editForm.title.trim() || !editForm.content.trim()) {
    message.warning('请填写标题和内容');
    return;
  }
  saving.value = true;
  try {
    await saveAnnouncementApi(editForm);
    message.success('保存成功');
    modalVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function handleDelete(id: number) {
  try {
    await deleteAnnouncementApi(id);
    message.success('删除成功');
    loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

onMounted(() => loadData(1));
</script>

<template>
  <Page title="公告管理" content-class="p-4">
    <Card>
      <div class="mb-4 flex flex-wrap items-center gap-3 justify-between">
        <Space wrap>
          <Input
            v-model:value="keyword"
            placeholder="搜索标题/内容"
            allow-clear
            style="max-width: 200px; min-width: 120px"
            @pressEnter="loadData(1)"
          />
          <Button type="primary" @click="loadData(1)">
            <template #icon><SearchOutlined /></template>搜索
          </Button>
        </Space>
        <Button type="primary" @click="openAdd">
          <template #icon><PlusOutlined /></template>新增公告
        </Button>
      </div>

      <Table
        :data-source="list"
        :columns="columns"
        :loading="loading"
        :pagination="false"
        row-key="id"
        size="small"
        bordered
        :scroll="{ x: 700 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="record.status === '1' ? 'green' : 'default'">
              {{ record.status === '1' ? '已发布' : '草稿' }}
            </Tag>
          </template>
          <template v-if="column.key === 'zhiding'">
            <Tag :color="record.zhiding === '1' ? 'orange' : 'default'">
              {{ record.zhiding === '1' ? '置顶' : '普通' }}
            </Tag>
          </template>
          <template v-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">
                <template #icon><EditOutlined /></template>
              </Button>
              <Popconfirm title="确定删除？" @confirm="handleDelete(record.id)">
                <Button type="link" danger size="small">
                  <template #icon><DeleteOutlined /></template>
                </Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <div class="mt-4 flex justify-end" v-if="pagination.total > pagination.limit">
        <Pagination
          :current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          @change="(p: number) => loadData(p)"
        />
      </div>
    </Card>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="modalVisible"
      :title="editForm.id ? '编辑公告' : '新增公告'"
      :confirm-loading="saving"
      @ok="handleSave"
      :width="640"
      style="max-width: 95vw"
    >
      <div class="space-y-4">
        <div>
          <div class="font-medium mb-1">标题</div>
          <Input v-model:value="editForm.title" placeholder="公告标题" />
        </div>
        <div>
          <div class="font-medium mb-1">内容</div>
          <Input.TextArea v-model:value="editForm.content" :rows="6" placeholder="公告内容" />
        </div>
        <div class="flex gap-6">
          <div class="flex items-center gap-2">
            <span class="text-sm">发布状态：</span>
            <Switch
              :checked="editForm.status === '1'"
              checked-children="发布"
              un-checked-children="草稿"
              @change="(v: boolean) => editForm.status = v ? '1' : '0'"
            />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm">置顶：</span>
            <Switch
              :checked="editForm.zhiding === '1'"
              checked-children="是"
              un-checked-children="否"
              @change="(v: boolean) => editForm.zhiding = v ? '1' : '0'"
            />
          </div>
        </div>
      </div>
    </Modal>
  </Page>
</template>
