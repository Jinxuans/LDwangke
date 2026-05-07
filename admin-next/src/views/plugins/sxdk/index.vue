<template>
  <div class="plugin-sxdk-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="resetFilters"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading || syncing" @refresh="loadOrders(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag type="success" effect="plain">运行中 {{ activeCount }}</ElTag>
            <ElTag type="warning" effect="plain">即将到期 {{ expiringCount }}</ElTag>
            <ElButton v-if="isAdmin" plain @click="handleSyncOrders">同步订单</ElButton>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
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

    <ElDialog v-model="addVisible" title="新增泰山打卡订单" width="720px">
      <div class="grid gap-5 lg:grid-cols-[1fr_280px]">
        <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">选择平台</p>
            <ElSelect v-model="addForm.platform" class="w-full" placeholder="请选择平台">
              <ElOption v-for="item in platformOptions" :key="item.value" :label="item.label" :value="item.value" />
            </ElSelect>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">手机号</p>
              <ElInput v-model="addForm.phone" placeholder="请输入手机号" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="addForm.password" placeholder="请输入密码" show-password />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">姓名</p>
              <ElInput v-model="addForm.name" placeholder="选填" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">结束日期</p>
              <ElDatePicker
                v-model="addForm.end_time"
                class="w-full"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="请选择结束日期"
              />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">打卡地址</p>
            <ElInput v-model="addForm.address" placeholder="选填" />
          </div>

          <div class="grid gap-4 md:grid-cols-3">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">上班打卡</p>
              <ElTimePicker
                v-model="addForm.up_check_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="08:30"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">下班打卡</p>
              <ElTimePicker
                v-model="addForm.down_check_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="17:30"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
              <ElInput v-model="addForm.check_week" placeholder="1,2,3,4,5" />
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">平台</span>
              <span class="font-medium text-g-900">{{ getPlatformLabel(addForm.platform) }}</span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">结束日期</span>
              <span class="font-medium text-g-900">{{ addForm.end_time || '未选择' }}</span>
            </div>
          </div>

          <div class="rounded-custom-sm border-full-d bg-box p-4">
            <p class="mb-3 text-sm font-medium text-g-800">附加选项</p>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">日报</span>
                <ElSwitch v-model="addForm.day_paper" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">周报</span>
                <ElSwitch v-model="addForm.week_paper" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">月报</span>
                <ElSwitch v-model="addForm.month_paper" :active-value="1" :inactive-value="0" />
              </div>
            </div>
          </div>

          <div class="rounded-custom-sm border-full-d bg-box px-4 py-4 text-sm leading-6 text-g-600">
            平台、手机号和密码填好后，可以尝试自动获取姓名和地址。
          </div>

          <ElButton plain class="w-full" :loading="fetchingInfo" @click="handleAutoFetch">自动获取信息</ElButton>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addLoading" @click="handleAdd">确认下单</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="editVisible" title="编辑订单" width="640px">
      <div class="space-y-4">
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">手机号</p>
            <ElInput v-model="editForm.phone" disabled />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">密码</p>
            <ElInput v-model="editForm.password" show-password />
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">姓名</p>
            <ElInput v-model="editForm.name" />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">结束日期</p>
            <ElDatePicker
              v-model="editForm.end_time"
              class="w-full"
              type="date"
              value-format="YYYY-MM-DD"
              placeholder="请选择结束日期"
            />
          </div>
        </div>

        <div>
          <p class="mb-2 text-sm font-medium text-g-800">打卡地址</p>
          <ElInput v-model="editForm.address" />
        </div>

        <div class="grid gap-4 md:grid-cols-3">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">上班打卡</p>
            <ElTimePicker
              v-model="editForm.up_check_time"
              class="w-full"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">下班打卡</p>
            <ElTimePicker
              v-model="editForm.down_check_time"
              class="w-full"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
            <ElInput v-model="editForm.check_week" />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="editLoading" @click="handleEdit">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="logVisible" :title="logTitle" width="720px">
      <ElScrollbar max-height="440px">
        <div v-if="logLoading" class="py-10 text-center text-sm text-g-500">日志加载中...</div>
        <div v-else-if="!logData" class="py-10 text-center text-sm text-g-500">暂无日志</div>
        <pre
          v-else
          class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-xs leading-6 text-g-700 whitespace-pre-wrap break-all"
        >{{ typeof logData === 'string' ? logData : JSON.stringify(logData, null, 2) }}</pre>
      </ElScrollbar>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    changeLegacySXDKCheckCode,
    createLegacySXDKOrder,
    deleteLegacySXDKOrder,
    editLegacySXDKOrder,
    fetchLegacySXDKLog,
    fetchLegacySXDKOrders,
    nowCheckLegacySXDKOrder,
    searchLegacySXDKPhoneInfo,
    syncLegacySXDKOrders,
    type LegacySXDKOrder
  } from '@/api/legacy/plugin-sxdk'

  defineOptions({ name: 'PluginSXDKPage' })

  const userStore = useUserStore()

  const platformOptions = [
    { value: 'zxjy', label: '在校教育' },
    { value: 'qzt', label: '签证通' },
    { value: 'xyb', label: '校友邦' },
    { value: 'gxy', label: '工学云' },
    { value: 'xxy', label: '校信' },
    { value: 'xxt', label: '校信通' },
    { value: 'hzj', label: '汇知教' }
  ]

  const searchFieldOptions = [
    { label: '手机号', value: 'phone' },
    { label: '姓名', value: 'name' },
    { label: '平台', value: 'platform' }
  ]

  const loading = ref(false)
  const syncing = ref(false)
  const fetchingInfo = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const editVisible = ref(false)
  const editLoading = ref(false)
  const logVisible = ref(false)
  const logLoading = ref(false)
  const logTitle = ref('订单日志')
  const logData = ref<any>(null)

  const orders = ref<LegacySXDKOrder[]>([])

  const pagination = reactive({
    page: 1,
    size: 10,
    total: 0
  })

  const filters = reactive({
    searchField: 'phone',
    searchValue: ''
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.size,
    total: pagination.total
  }))

  const addForm = reactive({
    platform: '',
    phone: '',
    password: '',
    name: '',
    address: '',
    up_check_time: '08:30',
    down_check_time: '17:30',
    check_week: '1,2,3,4,5',
    end_time: '',
    day_paper: 0,
    week_paper: 0,
    month_paper: 0
  })

  const editForm = reactive<Record<string, any>>({
    id: 0,
    platform: '',
    phone: '',
    password: '',
    name: '',
    address: '',
    up_check_time: '',
    down_check_time: '',
    check_week: '',
    end_time: '',
    day_paper: 0,
    week_paper: 0,
    month_paper: 0
  })

  const isAdmin = computed(() => {
    const roles = userStore.info?.roles || []
    return roles.includes('R_ADMIN') || roles.includes('R_SUPER')
  })

  const activeCount = computed(() => orders.value.filter((item) => item.code === 1).length)
  const expiringCount = computed(
    () => orders.value.filter((item) => getRemainDays(item.end_time) <= 3 && getRemainDays(item.end_time) >= 0).length
  )

  const searchItems = computed(() => [
    {
      label: '搜索项',
      key: 'searchField',
      type: 'select',
      props: {
        placeholder: '搜索项',
        options: searchFieldOptions
      }
    },
    {
      label: '关键词',
      key: 'searchValue',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入手机号、姓名或平台'
      }
    }
  ])

  const getPlatformLabel = (value: string) => platformOptions.find((item) => item.value === value)?.label || value

  const getRemainDays = (endTime?: string) => {
    if (!endTime) return 0
    const target = new Date(endTime).getTime()
    if (!target) return 0
    return Math.ceil((target - Date.now()) / 86400000)
  }

  const formatRemainDays = (endTime?: string) => {
    const days = getRemainDays(endTime)
    if (days < 0) return '已过期'
    if (days === 0) return '今天到期'
    return `剩余 ${days} 天`
  }

  const getStatusText = (order: LegacySXDKOrder) => {
    const days = getRemainDays(order.end_time)
    if (days < 0) return '已过期'
    if (order.code === 0) return '已暂停'
    if (order.code === 1) return '运行中'
    return `状态 ${order.code}`
  }

  const getStatusType = (order: LegacySXDKOrder) => {
    const days = getRemainDays(order.end_time)
    if (days < 0) return 'danger'
    if (order.code === 0) return 'info'
    if (order.code === 1) return days <= 3 ? 'warning' : 'success'
    return 'info'
  }

  const setDefaultEndDate = () => {
    const next = new Date()
    next.setDate(next.getDate() + 30)
    addForm.end_time = next.toISOString().slice(0, 10)
  }

  const { columns, columnChecks } = useTableColumns<LegacySXDKOrder>(() => [
    {
      prop: 'platform',
      label: '平台',
      width: 120,
      align: 'center',
      formatter: (row) => h(ElTag, { effect: 'plain', type: 'primary' }, () => getPlatformLabel(row.platform))
    },
    {
      prop: 'phone',
      label: '账号信息',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.phone || '-'),
          h('p', { class: 'text-xs text-g-500' }, `密码 ${row.password || '-'}`),
          h('p', { class: 'text-xs text-g-500' }, row.name || '未填写姓名')
        ])
    },
    {
      prop: 'up_check_time',
      label: '打卡时间',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, `上班 ${row.up_check_time || '-'}`),
          h('p', { class: 'text-sm text-g-800' }, `下班 ${row.down_check_time || '-'}`),
          h('p', { class: 'text-xs text-g-500' }, `周期 ${row.check_week || '-'}`)
        ])
    },
    {
      prop: 'code',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row) => h(ElTag, { type: getStatusType(row), effect: 'plain' }, () => getStatusText(row))
    },
    {
      prop: 'end_time',
      label: '到期时间',
      minWidth: 160,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-medium text-g-900' }, row.end_time || '-'),
          h('p', { class: 'text-xs text-g-500' }, formatRemainDays(row.end_time))
        ])
    },
    {
      prop: 'address',
      label: '地址 / 备注',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'line-clamp-2 text-sm text-g-700' }, row.address || '未填写地址'),
          h('p', { class: 'line-clamp-2 text-xs text-g-500' }, row.remark || row.wxpush || '暂无备注')
        ])
    },
    {
      prop: 'operation',
      label: '操作',
      width: 360,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { size: 'small', onClick: () => handleToggleStatus(row) }, () => (row.code === 1 ? '暂停' : '启用')),
          h(ElButton, { size: 'small', onClick: () => handleNowCheck(row) }, () => '立即打卡'),
          h(ElButton, { size: 'small', onClick: () => openEditDialog(row) }, () => '编辑'),
          h(ElButton, { size: 'small', onClick: () => openLogDialog(row) }, () => '日志'),
          h(ElButton, { size: 'small', type: 'danger', plain: true, onClick: () => handleDelete(row) }, () => '删除')
        ])
    }
  ])

  const resetFilters = () => {
    filters.searchField = 'phone'
    filters.searchValue = ''
    loadOrders(1)
  }

  const resetAddForm = () => {
    addForm.platform = ''
    addForm.phone = ''
    addForm.password = ''
    addForm.name = ''
    addForm.address = ''
    addForm.up_check_time = '08:30'
    addForm.down_check_time = '17:30'
    addForm.check_week = '1,2,3,4,5'
    addForm.day_paper = 0
    addForm.week_paper = 0
    addForm.month_paper = 0
    setDefaultEndDate()
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacySXDKOrders({
        page: pagination.page,
        size: pagination.size,
        searchField: filters.searchValue ? filters.searchField : undefined,
        searchValue: filters.searchValue || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = () => {
    loadOrders(1)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadOrders(1)
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
  }

  const handleAutoFetch = async () => {
    if (!addForm.platform || !addForm.phone || !addForm.password) {
      ElMessage.warning('请先填写平台、手机号和密码')
      return
    }
    fetchingInfo.value = true
    try {
      const result = await searchLegacySXDKPhoneInfo({
        platform: addForm.platform,
        phone: addForm.phone,
        password: addForm.password
      })
      const data = result?.data || result
      if (data?.name) addForm.name = data.name
      if (data?.address) addForm.address = data.address
      ElMessage.success(result?.msg || '自动获取成功')
    } finally {
      fetchingInfo.value = false
    }
  }

  const handleAdd = async () => {
    if (!addForm.platform || !addForm.phone || !addForm.password || !addForm.end_time) {
      ElMessage.warning('请填写平台、手机号、密码和结束日期')
      return
    }
    addLoading.value = true
    try {
      await createLegacySXDKOrder({ ...addForm })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openEditDialog = (order: LegacySXDKOrder) => {
    Object.assign(editForm, {
      id: order.id,
      platform: order.platform,
      phone: order.phone,
      password: order.password,
      name: order.name,
      address: order.address,
      up_check_time: order.up_check_time,
      down_check_time: order.down_check_time,
      check_week: order.check_week,
      end_time: order.end_time,
      day_paper: order.day_paper,
      week_paper: order.week_paper,
      month_paper: order.month_paper
    })
    editVisible.value = true
  }

  const handleEdit = async () => {
    editLoading.value = true
    try {
      await editLegacySXDKOrder({ ...editForm })
      ElMessage.success('编辑成功')
      editVisible.value = false
      loadOrders(pagination.page)
    } finally {
      editLoading.value = false
    }
  }

  const handleDelete = async (order: LegacySXDKOrder) => {
    await ElMessageBox.confirm(`确认删除 ${order.phone} 的订单？`, '删除订单', {
      type: 'warning',
      distinguishCancelAndClose: true,
      confirmButtonText: '删除并退款',
      cancelButtonText: '删除不退款'
    })
      .then(async () => {
        await deleteLegacySXDKOrder(order.id, true)
        ElMessage.success('删除成功')
        loadOrders(pagination.page)
      })
      .catch(async (action) => {
        if (action === 'cancel') {
          await deleteLegacySXDKOrder(order.id, false)
          ElMessage.success('删除成功')
          loadOrders(pagination.page)
        }
      })
  }

  const handleNowCheck = async (order: LegacySXDKOrder) => {
    await ElMessageBox.confirm(`确认对 ${order.phone} 执行立即打卡？这会触发额外扣费。`, '立即打卡', {
      type: 'warning'
    })
    const result = await nowCheckLegacySXDKOrder(order.id, order.platform)
    ElMessage.success(result?.msg || '打卡请求已发送')
  }

  const handleToggleStatus = async (order: LegacySXDKOrder) => {
    await changeLegacySXDKCheckCode(order.id, order.code === 1 ? 0 : 1)
    ElMessage.success(order.code === 1 ? '已暂停' : '已启用')
    loadOrders(pagination.page)
  }

  const openLogDialog = async (order: LegacySXDKOrder) => {
    logTitle.value = `${order.phone} 的打卡日志`
    logVisible.value = true
    logLoading.value = true
    try {
      logData.value = await fetchLegacySXDKLog(order.id)
    } finally {
      logLoading.value = false
    }
  }

  const handleSyncOrders = async () => {
    syncing.value = true
    try {
      await syncLegacySXDKOrders()
      ElMessage.success('同步完成')
      loadOrders(pagination.page)
    } finally {
      syncing.value = false
    }
  }

  onMounted(() => {
    resetAddForm()
    loadOrders(1)
  })
</script>
