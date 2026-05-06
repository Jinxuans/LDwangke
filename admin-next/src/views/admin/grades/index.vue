<template>
  <div class="admin-grades-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">等级 {{ filteredList.length }} 条</ElTag>
            <ElTag type="success" effect="plain">启用 {{ enabledCount }}</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增等级</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="loading" :data="filteredList" :columns="columns" :showTableHeader="true" />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑等级' : '新增等级'"
      width="780px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">等级基础配置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">等级名称和费率会直接影响用户价格体系。</p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
              <ElInput v-model="editForm.sort" placeholder="例如 10" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">等级名称</label>
              <ElInput v-model="editForm.name" maxlength="30" placeholder="例如 VIP1" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">等级费率</label>
              <ElInput v-model="editForm.rate" placeholder="例如 0.95" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">开通价格</label>
              <ElInput v-model="editForm.money" placeholder="例如 10" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <p class="text-lg font-semibold text-g-900">扣费与状态</p>
            <p class="mt-1 text-sm leading-6 text-g-500">维护添加扣费、改价扣费和启用状态。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">添加用户扣费</label>
              <ElSegmented
                v-model="editForm.addkf"
                :options="[
                  { label: '开启', value: '1' },
                  { label: '关闭', value: '0' }
                ]"
                class="w-full"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">改价扣费</label>
              <ElSegmented
                v-model="editForm.gjkf"
                :options="[
                  { label: '开启', value: '1' },
                  { label: '关闭', value: '0' }
                ]"
                class="w-full"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">等级状态</label>
              <ElSegmented
                v-model="editForm.status"
                :options="[
                  { label: '启用', value: '1' },
                  { label: '停用', value: '0' }
                ]"
                class="w-full"
              />
            </div>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">等级名称</span>
              <span class="font-medium text-g-900">{{ editForm.name || '未命名等级' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">费率 / 开通价格</span>
              <span class="font-medium text-g-900">{{ previewRateLabel }} / {{ previewMoneyLabel }}</span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="editForm.status === '1' ? 'success' : 'info'" effect="plain">
                {{ editForm.status === '1' ? '启用中' : '已停用' }}
              </ElTag>
              <ElTag :type="editForm.addkf === '1' ? 'warning' : 'info'" effect="plain">
                {{ editForm.addkf === '1' ? '添加扣费开启' : '添加扣费关闭' }}
              </ElTag>
              <ElTag :type="editForm.gjkf === '1' ? 'primary' : 'info'" effect="plain">
                {{ editForm.gjkf === '1' ? '改价扣费开启' : '改价扣费关闭' }}
              </ElTag>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存等级</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteLegacyAdminGrade,
    fetchLegacyAdminGrades,
    saveLegacyAdminGrade,
    type LegacyAdminGrade
  } from '@/api/legacy/admin-grades'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminGradesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const dialogVisible = ref(false)
  const list = ref<LegacyAdminGrade[]>([])

  const searchForm = ref<{
    keyword?: string
    status?: string
  }>({
    keyword: undefined,
    status: undefined
  })

  const appliedSearch = reactive({
    keyword: undefined as string | undefined,
    status: undefined as string | undefined
  })

  const editForm = reactive({
    id: 0,
    sort: '',
    name: '',
    rate: '1',
    money: '0',
    addkf: '1',
    gjkf: '1',
    status: '1'
  })

  const isEditing = computed(() => editForm.id > 0)

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索等级名称或费率'
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '启用', value: '1' },
          { label: '停用', value: '0' }
        ]
      }
    }
  ])

  const filteredList = computed(() => {
    const keyword = appliedSearch.keyword?.toLowerCase() || ''
    const status = appliedSearch.status

    return [...list.value].filter((item) => {
      const matchesKeyword =
        !keyword ||
        item.name?.toLowerCase().includes(keyword) ||
        String(item.rate || '')
          .toLowerCase()
          .includes(keyword)

      const matchesStatus = !status || item.status === status
      return matchesKeyword && matchesStatus
    })
  })

  const enabledCount = computed(() => filteredList.value.filter((item) => item.status === '1').length)

  const previewRateLabel = computed(() => formatRate(editForm.rate))
  const previewMoneyLabel = computed(() => formatMoney(editForm.money))

  const statusTagType = (status: string): 'info' | 'success' => (status === '1' ? 'success' : 'info')
  const toggleTagType = (value: string): 'info' | 'primary' | 'warning' =>
    value === '1' ? 'warning' : 'info'

  const { columns, columnChecks } = useTableColumns<LegacyAdminGrade>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'name',
      label: '等级名称',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名等级'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, `排序 ${row.sort ?? 0} / ID ${row.id}`)
        ])
    },
    {
      prop: 'rate',
      label: '等级费率',
      width: 120,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, formatRate(row.rate))
    },
    {
      prop: 'money',
      label: '开通价格',
      width: 120,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-warning)]' }, formatMoney(row.money))
    },
    {
      prop: 'addkf',
      label: '添加扣费',
      width: 120,
      formatter: (row) =>
        h(ElTag, { type: toggleTagType(row.addkf) }, () => (row.addkf === '1' ? '开启' : '关闭'))
    },
    {
      prop: 'gjkf',
      label: '改价扣费',
      width: 120,
      formatter: (row) =>
        h(ElTag, { type: row.gjkf === '1' ? 'primary' : 'info' }, () =>
          row.gjkf === '1' ? '开启' : '关闭'
        )
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: statusTagType(row.status) }, () => (row.status === '1' ? '启用' : '停用'))
    },
    {
      prop: 'time',
      label: '添加时间',
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
    editForm.sort = ''
    editForm.name = ''
    editForm.rate = '1'
    editForm.money = '0'
    editForm.addkf = '1'
    editForm.gjkf = '1'
    editForm.status = '1'
  }

  const formatMoney = (value?: string) => `¥${Number(value || 0).toFixed(2)}`
  const formatRate = (value?: string) => {
    const rate = Number(value || 0)
    return Number.isFinite(rate) && rate > 0 ? rate.toFixed(2) : '0.00'
  }

  const loadData = async () => {
    loading.value = true
    try {
      list.value = await fetchLegacyAdminGrades()
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { keyword?: string; status?: string }) => {
    appliedSearch.keyword = params.keyword?.trim() || undefined
    appliedSearch.status = params.status || undefined
  }

  const handleReset = () => {
    appliedSearch.keyword = undefined
    appliedSearch.status = undefined
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyAdminGrade) => {
    editForm.id = record.id
    editForm.sort = String(record.sort ?? '')
    editForm.name = record.name || ''
    editForm.rate = String(record.rate || '1')
    editForm.money = String(record.money || '0')
    editForm.addkf = record.addkf || '1'
    editForm.gjkf = record.gjkf || '1'
    editForm.status = record.status || '1'
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.name.trim()) {
      ElMessage.warning('请先填写等级名称')
      return
    }

    if (!editForm.rate.trim()) {
      ElMessage.warning('请先填写等级费率')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminGrade({
        id: editForm.id || undefined,
        sort: editForm.sort.trim(),
        name: editForm.name.trim(),
        rate: editForm.rate.trim(),
        money: editForm.money.trim(),
        addkf: editForm.addkf,
        gjkf: editForm.gjkf,
        status: editForm.status
      })
      ElMessage.success(editForm.id ? '等级已更新' : '等级已创建')
      dialogVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (record: LegacyAdminGrade) => {
    await ElMessageBox.confirm(`确定删除等级「${record.name || record.id}」吗？`, '删除等级', {
      type: 'warning'
    })
    await deleteLegacyAdminGrade(record.id)
    ElMessage.success('等级已删除')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>
