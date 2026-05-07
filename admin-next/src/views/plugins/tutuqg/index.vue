<template>
  <div class="plugin-tutuqg-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="resetFilters"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading || batchSyncLoading" @refresh="loadOrders(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag type="success" effect="plain">已上号 {{ onlineCount }}</ElTag>
            <ElTag type="warning" effect="plain">待处理 {{ pendingCount }}</ElTag>
            <ElButton plain :loading="batchSyncLoading" @click="handleBatchSync">批量同步</ElButton>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
            <ElButton v-if="isAdmin" plain @click="openConfigDialog">平台配置</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="orders"
        :columns="columns"
        :pagination="tablePagination"
        @pagination:current-change="loadOrders"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog v-model="addVisible" title="新增图图强国订单" width="520px">
      <div class="space-y-4">
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">账号</p>
          <ElInput v-model="addForm.user" maxlength="11" placeholder="请输入 11 位手机号" />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">密码</p>
          <ElInput v-model="addForm.pass" show-password placeholder="请输入密码" />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">购买天数</p>
          <ElInputNumber v-model="addForm.days" class="w-full" :min="1" :max="365" />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">推送 Token</p>
          <ElInput v-model="addForm.kcname" placeholder="选填" />
        </div>

        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">购买天数</span>
            <span class="font-medium text-g-900">{{ addForm.days || 0 }} 天</span>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">预计扣费</span>
            <span class="font-semibold text-[var(--el-color-danger)]">¥{{ estimatedCost }}</span>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addLoading" @click="handleCreate">确认下单</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="configVisible" title="图图强国配置" width="520px">
      <div class="space-y-4">
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">上游地址</p>
          <ElInput v-model="configForm.base_url" placeholder="例如 http://x.x.x.x:2345" />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">通信 Key</p>
          <ElInput v-model="configForm.key" show-password placeholder="请输入上游分配的 Key" />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">额外加价</p>
          <ElInputNumber v-model="configForm.price_increment" class="w-full" :step="0.1" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="configVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="configLoading" @click="handleSaveConfig">保存</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    batchSyncLegacyTutuQGOrders,
    changeLegacyTutuQGPassword,
    changeLegacyTutuQGToken,
    createLegacyTutuQGOrder,
    deleteLegacyTutuQGOrder,
    fetchLegacyTutuQGConfig,
    fetchLegacyTutuQGOrders,
    fetchLegacyTutuQGPrice,
    refundLegacyTutuQGOrder,
    renewLegacyTutuQGOrder,
    saveLegacyTutuQGConfig,
    syncLegacyTutuQGOrder,
    toggleLegacyTutuQGRenew,
    type LegacyTutuQGConfig,
    type LegacyTutuQGOrder
  } from '@/api/legacy/plugin-tutuqg'

  defineOptions({ name: 'PluginTutuQGPage' })

  const userStore = useUserStore()

  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const configVisible = ref(false)
  const configLoading = ref(false)
  const batchSyncLoading = ref(false)

  const orders = ref<LegacyTutuQGOrder[]>([])
  const estimatedCost = ref('0.00')

  const pagination = reactive({
    limit: 10,
    page: 1,
    total: 0
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.limit,
    total: pagination.total
  }))

  const filters = reactive({
    searchText: ''
  })

  const addForm = reactive({
    days: 30,
    kcname: '',
    pass: '',
    user: ''
  })

  const configForm = reactive<LegacyTutuQGConfig>({
    base_url: '',
    key: '',
    price_increment: 0
  })

  const isAdmin = computed(() => {
    const roles = userStore.info?.roles || []
    return roles.includes('R_ADMIN') || roles.includes('R_SUPER')
  })
  const onlineCount = computed(() => orders.value.filter((item) => item.status === '已上号').length)
  const pendingCount = computed(() => orders.value.filter((item) => item.status === '待处理').length)

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'searchText',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索账号'
      }
    }
  ])

  const { columns, columnChecks } = useTableColumns<LegacyTutuQGOrder>(() => [
    {
      prop: 'user',
      label: '账号信息',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.user || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.pass || '-')
        ])
    },
    {
      prop: 'kcname',
      label: '推送 Token',
      minWidth: 150,
      formatter: (row) => h('span', { class: 'text-sm text-g-700' }, row.kcname || '未填写')
    },
    {
      prop: 'days',
      label: '订单信息',
      minWidth: 150,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, `${row.days} 天`),
          h('p', { class: 'text-xs text-g-500' }, `¥${Number(row.fees || 0).toFixed(2)}`)
        ])
    },
    {
      prop: 'score',
      label: '分数',
      minWidth: 120,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, `今日 ${row.score || '待更新'}`),
          h('p', { class: 'text-xs text-g-500' }, `累计 ${row.scores || '-'}`)
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 140,
      align: 'center',
      formatter: (row) => h(ElTag, { type: getStatusType(row.status), effect: 'plain' }, () => row.status || '未知')
    },
    {
      prop: 'remarks',
      label: '到期时间',
      minWidth: 150,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, row.remarks || '-'),
          h('p', { class: 'text-xs text-g-500' }, getExpireText(row.remarks))
        ])
    },
    {
      prop: 'zdxf',
      label: '自动续费',
      width: 110,
      align: 'center',
      formatter: (row) =>
        h(
          ElTag,
          {
            class: 'cursor-pointer',
            type: row.zdxf === '2' ? 'success' : 'info',
            effect: 'plain',
            onClick: () => handleToggleRenew(row)
          },
          () => (row.zdxf === '2' ? '已开启' : '已关闭')
        )
    },
    {
      prop: 'addtime',
      label: '下单时间',
      minWidth: 160
    },
    {
      prop: 'operation',
      label: '操作',
      width: 380,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { size: 'small', onClick: () => handleSync(row) }, () => '同步'),
          h(ElButton, { size: 'small', onClick: () => handleRenew(row) }, () => '续费'),
          h(ElButton, { size: 'small', onClick: () => handleChangePassword(row) }, () => '改密'),
          h(ElButton, { size: 'small', onClick: () => handleChangeToken(row) }, () => '改 Token'),
          h(ElButton, { size: 'small', type: 'danger', plain: true, onClick: () => handleRefund(row) }, () => '退单'),
          h(ElButton, { size: 'small', type: 'danger', plain: true, onClick: () => handleDelete(row) }, () => '删除')
        ])
    }
  ])

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyTutuQGOrders({
        limit: pagination.limit,
        page: pagination.page,
        search: filters.searchText || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const calculatePrice = async () => {
    if (!addForm.days) {
      estimatedCost.value = '0.00'
      return
    }
    try {
      const result = await fetchLegacyTutuQGPrice(addForm.days)
      estimatedCost.value = Number(result?.total_cost || 0).toFixed(2)
    } catch {
      estimatedCost.value = '0.00'
    }
  }

  const resetFilters = () => {
    filters.searchText = ''
    loadOrders(1)
  }

  const handleSearch = () => {
    loadOrders(1)
  }

  const handleSizeChange = (size: number) => {
    pagination.limit = size
    loadOrders(1)
  }

  const resetAddForm = () => {
    addForm.user = ''
    addForm.pass = ''
    addForm.days = 30
    addForm.kcname = ''
    estimatedCost.value = '0.00'
    calculatePrice()
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
  }

  const handleCreate = async () => {
    if (!addForm.user || !addForm.pass || !addForm.days) {
      ElMessage.warning('请填写完整信息')
      return
    }
    if (addForm.user.length !== 11) {
      ElMessage.warning('账号必须是 11 位手机号')
      return
    }

    await ElMessageBox.confirm(
      `确认给账号 ${addForm.user} 下单 ${addForm.days} 天？预计扣费 ¥${estimatedCost.value}`,
      '确认下单',
      { type: 'warning' }
    )

    addLoading.value = true
    try {
      await createLegacyTutuQGOrder({ ...addForm })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const handleSync = async (order: LegacyTutuQGOrder) => {
    await syncLegacyTutuQGOrder(order.oid)
    ElMessage.success('同步成功')
    loadOrders(pagination.page)
  }

  const handleBatchSync = async () => {
    batchSyncLoading.value = true
    try {
      const result = await batchSyncLegacyTutuQGOrders()
      ElMessage.success(`批量同步完成：成功 ${result?.success || 0}，失败 ${result?.fail || 0}`)
      loadOrders(pagination.page)
    } finally {
      batchSyncLoading.value = false
    }
  }

  const handleRenew = async (order: LegacyTutuQGOrder) => {
    const { value } = await ElMessageBox.prompt(`请输入账号 ${order.user} 的续费天数`, '续费订单', {
      inputPattern: /^[1-9]\d*$/,
      inputValue: '30',
      inputErrorMessage: '请输入正整数'
    })
    if (!value) return
    await renewLegacyTutuQGOrder(order.oid, Number(value))
    ElMessage.success('续费成功')
    loadOrders(pagination.page)
  }

  const handleChangePassword = async (order: LegacyTutuQGOrder) => {
    const { value } = await ElMessageBox.prompt(`请输入账号 ${order.user} 的新密码`, '修改密码', {
      inputValue: order.pass || ''
    })
    if (value === null || value === undefined) return
    await changeLegacyTutuQGPassword(order.oid, value)
    ElMessage.success('密码已更新')
    loadOrders(pagination.page)
  }

  const handleChangeToken = async (order: LegacyTutuQGOrder) => {
    const { value } = await ElMessageBox.prompt(`请输入账号 ${order.user} 的新 Token`, '修改 Token', {
      inputValue: order.kcname || ''
    })
    if (value === null || value === undefined) return
    await changeLegacyTutuQGToken(order.oid, value)
    ElMessage.success('Token 已更新')
    loadOrders(pagination.page)
  }

  const handleRefund = async (order: LegacyTutuQGOrder) => {
    await ElMessageBox.confirm(`确认退单退费 ${order.user}？该操作不可撤销。`, '退单退费', {
      type: 'warning'
    })
    await refundLegacyTutuQGOrder(order.oid)
    ElMessage.success('退单成功')
    loadOrders(pagination.page)
  }

  const handleDelete = async (order: LegacyTutuQGOrder) => {
    await ElMessageBox.confirm(`确认删除订单 ${order.user}？`, '删除订单', {
      type: 'warning'
    })
    await deleteLegacyTutuQGOrder(order.oid)
    ElMessage.success('删除成功')
    loadOrders(pagination.page)
  }

  const handleToggleRenew = async (order: LegacyTutuQGOrder) => {
    await toggleLegacyTutuQGRenew(order.oid)
    ElMessage.success('自动续费状态已更新')
    loadOrders(pagination.page)
  }

  const openConfigDialog = async () => {
    const result = await fetchLegacyTutuQGConfig()
    Object.assign(configForm, result || {})
    configVisible.value = true
  }

  const handleSaveConfig = async () => {
    configLoading.value = true
    try {
      await saveLegacyTutuQGConfig({ ...configForm })
      ElMessage.success('配置已保存')
      configVisible.value = false
    } finally {
      configLoading.value = false
    }
  }

  const getStatusType = (status?: null | string) => {
    if (!status) return 'info'
    if (status === '已上号') return 'success'
    if (status === '待处理') return 'warning'
    if (status === '需要接码' || status === '异常') return 'danger'
    return 'info'
  }

  const getExpireText = (remarks?: null | string) => {
    if (!remarks) return '未同步'
    const target = new Date(remarks).getTime()
    if (!target) return '未同步'
    const diff = Math.ceil((target - Date.now()) / 86400000)
    if (diff < 0) return '已过期'
    if (diff === 0) return '今天到期'
    return `剩余 ${diff} 天`
  }

  watch(
    () => addForm.days,
    () => {
      if (addVisible.value) {
        calculatePrice()
      }
    }
  )

  onMounted(() => {
    loadOrders(1)
  })
</script>
