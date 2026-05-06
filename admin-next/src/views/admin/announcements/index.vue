<template>
  <div class="admin-announcements-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">公告 {{ pagination.total }} 条</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增公告</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑公告' : '新增公告'"
      width="760px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">公告内容</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">维护公告标题和正文内容。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">公告标题</label>
              <ElInput v-model="editForm.title" maxlength="120" placeholder="请输入公告标题" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">公告正文</label>
              <ElInput
                v-model="editForm.content"
                type="textarea"
                :rows="9"
                resize="none"
                placeholder="请输入公告内容"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <p class="text-lg font-semibold text-g-900">发布设置</p>
            <p class="mt-1 text-sm leading-6 text-g-500">控制状态、可见范围和置顶摘要。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">发布状态</label>
              <ElSegmented
                v-model="editForm.status"
                :options="[
                  { label: '已发布', value: '1' },
                  { label: '草稿', value: '0' }
                ]"
                class="w-full"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">可见范围</label>
              <ElSegmented
                v-model="editForm.visibility"
                :options="[
                  { label: '全体用户', value: 0 },
                  { label: '直属下级', value: 1 }
                ]"
                class="w-full"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">公告优先级</label>
              <ElSwitch
                :model-value="editForm.zhiding === '1'"
                active-text="置顶"
                inactive-text="普通"
                @change="(value) => (editForm.zhiding = value ? '1' : '0')"
              />
            </div>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">公告标题</span>
              <span class="truncate font-medium text-g-900">{{ editForm.title || '未填写标题' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">内容长度</span>
              <span class="font-medium text-g-900">{{ editForm.content.trim().length || 0 }} 字</span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="editForm.status === '1' ? 'success' : 'info'" effect="plain">
                {{ editForm.status === '1' ? '已发布' : '草稿' }}
              </ElTag>
              <ElTag :type="editForm.visibility === 1 ? 'warning' : 'primary'" effect="plain">
                {{ editForm.visibility === 1 ? '直属下级' : '全体用户' }}
              </ElTag>
              <ElTag :type="editForm.zhiding === '1' ? 'warning' : 'info'" effect="plain">
                {{ editForm.zhiding === '1' ? '置顶' : '普通' }}
              </ElTag>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存公告</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteLegacyAdminAnnouncement,
    fetchLegacyAdminAnnouncements,
    saveLegacyAdminAnnouncement,
    type LegacyAdminAnnouncement
  } from '@/api/legacy/admin-content'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminAnnouncementsPage' })

  const loading = ref(false)
  const saving = ref(false)
  const list = ref<LegacyAdminAnnouncement[]>([])
  const dialogVisible = ref(false)

  const searchForm = ref<{
    keyword?: string
  }>({
    keyword: undefined
  })

  const appliedSearch = reactive({
    keyword: undefined as string | undefined
  })

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const editForm = reactive({
    id: 0,
    title: '',
    content: '',
    status: '1',
    zhiding: '0',
    visibility: 0
  })

  const isEditing = computed(() => editForm.id > 0)
  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索标题或正文'
      }
    }
  ])

  const visibilityLabel = (value: number) => (value === 1 ? '直属下级' : '全体用户')

  const { columns, columnChecks } = useTableColumns<LegacyAdminAnnouncement>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'title',
      label: '公告标题',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.title || '未命名公告'),
          h(
            'p',
            { class: 'text-xs text-g-500 line-clamp-2 mt-1' },
            row.content || '暂无公告内容'
          )
        ])
    },
    {
      prop: 'author',
      label: '作者',
      width: 120,
      formatter: (row) => row.author || `UID ${row.uid || 0}`
    },
    {
      prop: 'visibility',
      label: '可见范围',
      width: 120,
      formatter: (row) =>
        h(
          ElTag,
          { type: row.visibility === 1 ? 'warning' : 'primary' },
          () => visibilityLabel(row.visibility)
        )
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: row.status === '1' ? 'success' : 'info' }, () =>
          row.status === '1' ? '已发布' : '草稿'
        )
    },
    {
      prop: 'zhiding',
      label: '置顶',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: row.zhiding === '1' ? 'warning' : 'info' }, () =>
          row.zhiding === '1' ? '置顶' : '普通'
        )
    },
    {
      prop: 'time',
      label: '发布时间',
      width: 180
    },
    {
      prop: 'operation',
      label: '操作',
      width: 140,
      fixed: 'right',
      formatter: (row) =>
        h('div', [
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => openEditDialog(row)
          }),
          h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDelete(row)
          })
        ])
    }
  ])

  const resetEditForm = () => {
    editForm.id = 0
    editForm.title = ''
    editForm.content = ''
    editForm.status = '1'
    editForm.zhiding = '0'
    editForm.visibility = 0
  }

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminAnnouncements({
        page: pagination.current,
        limit: pagination.size,
        keyword: appliedSearch.keyword
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const refreshData = async () => {
    await loadData(pagination.current)
  }

  const handleSearch = (params: { keyword?: string }) => {
    appliedSearch.keyword = params.keyword?.trim() || undefined
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.keyword = undefined
    loadData(1)
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyAdminAnnouncement) => {
    editForm.id = record.id
    editForm.title = record.title || ''
    editForm.content = record.content || ''
    editForm.status = record.status || '1'
    editForm.zhiding = record.zhiding || '0'
    editForm.visibility = Number(record.visibility || 0)
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.title.trim() || !editForm.content.trim()) {
      ElMessage.warning('请先填写标题和内容')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminAnnouncement({
        id: editForm.id || undefined,
        title: editForm.title.trim(),
        content: editForm.content.trim(),
        status: editForm.status,
        zhiding: editForm.zhiding,
        visibility: editForm.visibility
      })
      ElMessage.success(editForm.id ? '公告已更新' : '公告已创建')
      dialogVisible.value = false
      await loadData(editForm.id ? pagination.current : 1)
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (record: LegacyAdminAnnouncement) => {
    await ElMessageBox.confirm(`确定删除公告「${record.title || record.id}」吗？`, '删除公告', {
      type: 'warning'
    })
    await deleteLegacyAdminAnnouncement(record.id)
    ElMessage.success('公告已删除')
    const nextPage =
      list.value.length === 1 && pagination.current > 1 ? pagination.current - 1 : pagination.current
    await loadData(nextPage)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
