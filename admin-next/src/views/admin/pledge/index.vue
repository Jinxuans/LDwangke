<template>
  <div class="admin-pledge-page art-full-height">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap gap-3">
        <ElTag effect="plain">质押管理</ElTag>
        <ElTag effect="plain">方案数 {{ configs.length }}</ElTag>
        <ElTag type="success" effect="plain">启用方案 {{ enabledConfigCount }}</ElTag>
        <ElTag type="warning" effect="plain">记录总数 {{ recordPagination.total }}</ElTag>
      </div>
      <ElButton plain :loading="configLoading || recordLoading" @click="refreshCurrentTab">刷新数据</ElButton>
    </div>

    <ElCard class="art-table-card !pb-0">
      <ElTabs v-model="activeTab" class="px-5 pt-3" @tab-change="handleTabChange">
        <ElTabPane label="质押方案" name="configs">
          <div class="px-0 pb-5">
            <ArtTableHeader
              v-model:columns="configColumnChecks"
              :loading="configLoading"
              @refresh="loadConfigs"
            >
              <template #left>
                <ElSpace wrap>
                  <ElTag effect="plain">方案数 {{ configs.length }}</ElTag>
                  <ElTag type="success" effect="plain">启用方案 {{ enabledConfigCount }}</ElTag>
                  <ElButton type="primary" plain @click="openCreateDialog">新增方案</ElButton>
                </ElSpace>
              </template>
            </ArtTableHeader>

            <ArtTable
              :loading="configLoading"
              :data="configs"
              :columns="configColumns"
              :show-table-header="true"
            />
          </div>
        </ElTabPane>

        <ElTabPane label="质押记录" name="records">
          <div class="px-0 pb-5">
            <ArtTableHeader
              v-model:columns="recordColumnChecks"
              :loading="recordLoading"
              @refresh="loadRecords(recordPagination.current)"
            >
              <template #left>
                <ElTag type="warning" effect="plain">记录总数 {{ recordPagination.total }}</ElTag>
              </template>
            </ArtTableHeader>

            <ArtTable
              :loading="recordLoading"
              :data="records"
              :columns="recordColumns"
              :pagination="recordPagination"
              @pagination:current-change="handleRecordCurrentChange"
              @pagination:size-change="handleRecordSizeChange"
            />
          </div>
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑质押方案' : '新增质押方案'"
      width="720px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.08fr_0.92fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">方案参数</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">分类、金额和折扣率都会直接影响用户下单结算价格。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">课程分类</label>
              <ElSelect
                v-model="editForm.category_id"
                class="w-full"
                clearable
                filterable
                placeholder="请选择分类"
              >
                <ElOption
                  v-for="item in categoryOptions"
                  :key="item.id"
                  :label="`${item.name}（ID ${item.id}）`"
                  :value="item.id"
                />
              </ElSelect>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">质押金额（元）</label>
                <ElInputNumber v-model="editForm.amount" class="w-full" :min="0.01" :precision="2" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">质押天数</label>
                <ElInputNumber v-model="editForm.days" class="w-full" :min="1" :precision="0" />
              </div>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">折扣率</label>
                <ElInputNumber
                  v-model="editForm.discount_rate"
                  class="w-full"
                  :min="0.01"
                  :max="1"
                  :step="0.01"
                  :precision="2"
                />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">提前取消扣费比例</label>
                <ElInputNumber
                  v-model="editForm.cancel_fee"
                  class="w-full"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  :precision="2"
                />
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">方案摘要</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">确认分类、折扣和扣费比例。</p>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">课程分类</span>
              <span class="font-medium text-g-900">{{ selectedCategoryLabel }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">折扣率</span>
              <span class="font-medium text-[var(--el-color-primary)]">{{ formatDiscount(editForm.discount_rate) }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">质押金额</span>
              <span class="font-medium text-g-900">¥{{ formatMoney(editForm.amount) }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">取消扣费</span>
              <span class="font-medium text-g-900">{{ formatPercent(editForm.cancel_fee) }}</span>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存方案</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { fetchLegacyAdminCategoryOptions, type LegacyAdminCategory } from '@/api/legacy/admin-categories'
  import {
    deleteLegacyAdminPledgeConfig,
    fetchLegacyAdminPledgeConfigs,
    fetchLegacyAdminPledgeRecords,
    saveLegacyAdminPledgeConfig,
    toggleLegacyAdminPledgeConfig,
    type LegacyAdminPledgeConfig,
    type LegacyAdminPledgeRecord
  } from '@/api/legacy/admin-auxiliary'
  import { ElMessage, ElMessageBox, ElSwitch, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminPledgePage' })

  const activeTab = ref('configs')
  const configLoading = ref(false)
  const recordLoading = ref(false)
  const saving = ref(false)
  const dialogVisible = ref(false)

  const configs = ref<LegacyAdminPledgeConfig[]>([])
  const records = ref<LegacyAdminPledgeRecord[]>([])
  const categoryOptions = ref<LegacyAdminCategory[]>([])

  const recordPagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const editForm = reactive({
    id: 0,
    category_id: 0,
    amount: 0,
    discount_rate: 0.9,
    days: 30,
    cancel_fee: 0
  })

  const isEditing = computed(() => editForm.id > 0)
  const enabledConfigCount = computed(() => configs.value.filter((item) => Number(item.status) === 1).length)
  const selectedCategoryLabel = computed(() => {
    return categoryOptions.value.find((item) => item.id === editForm.category_id)?.name || '未选择分类'
  })

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)
  const formatPercent = (value?: number | string) => `${(Number(value || 0) * 100).toFixed(0)}%`
  const formatDiscount = (value?: number | string) => `${(Number(value || 0) * 100).toFixed(0)}% 折`

  const { columns: configColumns, columnChecks: configColumnChecks } =
    useTableColumns<LegacyAdminPledgeConfig>(() => [
      {
        type: 'index',
        label: '序号',
        width: 70
      },
      {
        prop: 'category_name',
        label: '分类',
        minWidth: 180,
        formatter: (row) =>
          h('div', { class: 'leading-6' }, [
            h('p', { class: 'font-semibold text-g-900' }, row.category_name || '未命名分类'),
            h('p', { class: 'text-xs text-g-500 mt-1' }, `配置 ID ${row.id}`)
          ])
      },
      {
        prop: 'amount',
        label: '质押金额',
        width: 120,
        formatter: (row) =>
          h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, `¥${formatMoney(row.amount)}`)
      },
      {
        prop: 'discount_rate',
        label: '折扣率',
        width: 100,
        formatter: (row) => formatDiscount(row.discount_rate)
      },
      {
        prop: 'days',
        label: '天数',
        width: 90,
        align: 'center'
      },
      {
        prop: 'cancel_fee',
        label: '取消扣费',
        width: 110,
        formatter: (row) => formatPercent(row.cancel_fee)
      },
      {
        prop: 'status',
        label: '状态',
        width: 120,
        formatter: (row) =>
          h(ElSwitch, {
            modelValue: Number(row.status) === 1,
            activeText: '启用',
            inactiveText: '禁用',
            onChange: (value: string | number | boolean) => handleToggleStatus(row.id, Boolean(value))
          })
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
              onClick: () => handleDeleteConfig(row)
            })
          ])
      }
    ])

  const { columns: recordColumns, columnChecks: recordColumnChecks } =
    useTableColumns<LegacyAdminPledgeRecord>(() => [
      {
        type: 'index',
        label: '序号',
        width: 70
      },
      {
        prop: 'username',
        label: '用户',
        width: 130,
        formatter: (row) => row.username || `UID ${row.uid}`
      },
      {
        prop: 'category_name',
        label: '分类',
        minWidth: 180,
        formatter: (row) => row.category_name || `配置 ${row.config_id}`
      },
      {
        prop: 'amount',
        label: '质押金额',
        width: 120,
        formatter: (row) =>
          h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, `¥${formatMoney(row.amount)}`)
      },
      {
        prop: 'discount_rate',
        label: '折扣率',
        width: 100,
        formatter: (row) => formatDiscount(row.discount_rate)
      },
      {
        prop: 'days',
        label: '天数',
        width: 90,
        align: 'center'
      },
      {
        prop: 'status',
        label: '状态',
        width: 110,
        formatter: (row) =>
          h(ElTag, { type: Number(row.status) === 1 ? 'success' : 'info' }, () =>
            Number(row.status) === 1 ? '生效中' : '已退还'
          )
      },
      {
        prop: 'addtime',
        label: '质押时间',
        width: 170
      },
      {
        prop: 'endtime',
        label: '退还时间',
        width: 170,
        formatter: (row) => row.endtime || '-'
      }
    ])

  const resetEditForm = () => {
    editForm.id = 0
    editForm.category_id = 0
    editForm.amount = 0
    editForm.discount_rate = 0.9
    editForm.days = 30
    editForm.cancel_fee = 0
  }

  const loadCategoryOptions = async () => {
    categoryOptions.value = await fetchLegacyAdminCategoryOptions()
  }

  const loadConfigs = async () => {
    configLoading.value = true
    try {
      const result = await fetchLegacyAdminPledgeConfigs()
      configs.value = Array.isArray(result) ? result : []
    } finally {
      configLoading.value = false
    }
  }

  const loadRecords = async (page = recordPagination.current) => {
    recordLoading.value = true
    recordPagination.current = page
    try {
      const result = await fetchLegacyAdminPledgeRecords({
        page: recordPagination.current,
        limit: recordPagination.size
      })
      records.value = result.list || []
      recordPagination.total = Number(result.pagination?.total || 0)
      recordPagination.current = Number(result.pagination?.page || recordPagination.current)
      recordPagination.size = Number(result.pagination?.limit || recordPagination.size)
    } finally {
      recordLoading.value = false
    }
  }

  const refreshCurrentTab = () => {
    if (activeTab.value === 'records') {
      loadRecords(recordPagination.current)
      return
    }
    loadConfigs()
  }

  const handleTabChange = (name: string | number) => {
    if (String(name) === 'records' && !records.value.length) {
      loadRecords(1)
    }
  }

  const handleRecordCurrentChange = (page: number) => {
    loadRecords(page)
  }

  const handleRecordSizeChange = (size: number) => {
    recordPagination.size = size
    loadRecords(1)
  }

  const openCreateDialog = async () => {
    if (!categoryOptions.value.length) {
      await loadCategoryOptions()
    }
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = async (record: LegacyAdminPledgeConfig) => {
    if (!categoryOptions.value.length) {
      await loadCategoryOptions()
    }
    editForm.id = record.id
    editForm.category_id = record.category_id
    editForm.amount = Number(record.amount || 0)
    editForm.discount_rate = Number(record.discount_rate || 0)
    editForm.days = Number(record.days || 30)
    editForm.cancel_fee = Number(record.cancel_fee || 0)
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.category_id) {
      ElMessage.warning('请先选择分类')
      return
    }
    if (Number(editForm.amount) <= 0) {
      ElMessage.warning('质押金额必须大于 0')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminPledgeConfig({
        id: editForm.id || undefined,
        category_id: editForm.category_id,
        amount: editForm.amount,
        discount_rate: editForm.discount_rate,
        days: editForm.days,
        cancel_fee: editForm.cancel_fee
      })
      ElMessage.success(editForm.id ? '方案已更新' : '方案已创建')
      dialogVisible.value = false
      await loadConfigs()
    } finally {
      saving.value = false
    }
  }

  const handleDeleteConfig = async (record: LegacyAdminPledgeConfig) => {
    await ElMessageBox.confirm(
      `确定删除质押方案「${record.category_name || record.id}」吗？`,
      '删除质押方案',
      { type: 'warning' }
    )
    await deleteLegacyAdminPledgeConfig(record.id)
    ElMessage.success('质押方案已删除')
    await loadConfigs()
  }

  const handleToggleStatus = async (id: number, enabled: boolean) => {
    await toggleLegacyAdminPledgeConfig(id, enabled ? 1 : 0)
    ElMessage.success('状态已更新')
    await loadConfigs()
  }

  onMounted(async () => {
    await Promise.all([loadCategoryOptions(), loadConfigs()])
  })
</script>
