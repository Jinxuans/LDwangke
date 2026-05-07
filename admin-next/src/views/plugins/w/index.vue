<template>
  <div class="plugin-w-page art-full-height">
    <ElAlert
      type="warning"
      :closable="false"
      show-icon
      title="下单前请确认账号密码正确，跑步期间不要登录账号。先查询跑区，再确认项目规则和扣费方式。"
      class="mb-5"
    />

    <ArtSearchBar
      v-model="filters"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="resetFilters"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadOrders(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">项目 {{ apps.length }} 个</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag type="warning" effect="plain">待下单 {{ waitAddCount }}</ElTag>
            <ElButton type="primary" plain :disabled="!apps.length" @click="openAddDialog">新增订单</ElButton>
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

    <ElDialog v-model="addVisible" title="新增鲸鱼运动订单" width="760px">
      <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">选择项目</p>
            <ElSelect v-model="addForm.app_id" class="w-full" filterable placeholder="请选择项目" @change="handleAppChange">
              <ElOption v-for="item in apps" :key="item.app_id" :label="item.name" :value="item.app_id" />
            </ElSelect>
            <p v-if="selectedApp?.description" class="mt-2 text-xs text-g-500">{{ selectedApp.description }}</p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">{{ accountLabel }}</p>
              <ElInput v-model="addForm.account" :placeholder="`请输入${accountLabel}`" />
            </div>
            <div v-if="needPassword">
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="addForm.password" show-password placeholder="请输入密码" />
            </div>
          </div>

          <div>
            <div class="mb-2 flex items-center justify-between gap-3">
              <p class="text-sm font-medium text-g-800">{{ zoneLabel }}</p>
              <ElButton plain size="small" :loading="queryLoading" :disabled="!selectedApp" @click="handleQueryZones">
                查询{{ zoneLabel }}
              </ElButton>
            </div>

            <ElSelect
              v-if="zones.length"
              v-model="addForm.zone_id"
              class="w-full"
              clearable
              filterable
              :placeholder="`请选择${zoneLabel}`"
              @change="handleZoneChange"
            >
              <ElOption v-for="item in zones" :key="item.id" :label="item.name" :value="item.id" />
            </ElSelect>
            <ElInput v-else v-model="addForm.zone_name" :placeholder="`手动输入${zoneLabel}`" />
            <p v-if="queryMessage" class="mt-2 text-xs" :class="zones.length ? 'text-emerald-600' : 'text-g-500'">
              {{ queryMessage }}
            </p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">跑步类型</p>
              <ElRadioGroup v-model="addForm.run_type">
                <ElRadioButton :label="1">有效跑</ElRadioButton>
                <ElRadioButton :label="2">自由跑</ElRadioButton>
              </ElRadioGroup>
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">次数</p>
              <ElInputNumber v-model="addForm.task_count" class="w-full" :min="1" :max="365" />
            </div>
          </div>

          <div v-if="selectedApp?.cac_type === 'KM'">
            <p class="mb-2 text-sm font-medium text-g-800">每次公里数</p>
            <ElInputNumber v-model="addForm.dis" class="w-full" :min="0.5" :max="100" :step="0.5" :precision="1" />
          </div>

        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <div class="grid gap-3 text-sm md:grid-cols-2">
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">计费方式</span>
              <span class="font-medium text-g-900">{{ selectedApp?.cac_type === 'TS' ? '按次计费' : '按公里计费' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">单价</span>
              <span class="font-medium text-g-900">
                {{ selectedApp ? `¥${Number(selectedApp.price || 0).toFixed(2)}` : '-' }}/{{ selectedApp?.cac_type === 'TS' ? '次' : 'km' }}
              </span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">任务配置</span>
              <span class="text-right font-medium text-g-900">{{ addForm.task_count }} 次{{ selectedApp?.cac_type === 'KM' ? `，${addForm.dis} km/次` : '' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">预估扣费</span>
              <span class="font-semibold text-[var(--el-color-danger)]">¥{{ estimatedPrice }}</span>
            </div>
          </div>
          <p class="mt-2 text-xs text-g-500">先查询{{ zoneLabel }}，再确认项目、次数和扣费方式。若上游查不到，可直接手动填写。</p>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addLoading" @click="handleCreate">确认下单</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    createLegacyWOrder,
    fetchLegacyWApps,
    fetchLegacyWOrders,
    proxyLegacyWAction,
    refundLegacyWOrder,
    resumeLegacyWOrder,
    syncLegacyWOrder,
    type LegacyWApp,
    type LegacyWOrder
  } from '@/api/legacy/plugin-w'

  defineOptions({ name: 'PluginWPage' })

  interface SelectItem {
    id: string
    name: string
  }

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '正常', value: 'NORMAL' },
    { label: '下单中', value: 'ADDING' },
    { label: '待下单', value: 'WAITADD' },
    { label: '已退款', value: 'REFUND' },
    { label: '待退款', value: 'WAITREFUND' },
    { label: '退款失败', value: 'REFUNDFAIL' }
  ]

  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const queryLoading = ref(false)

  const apps = ref<LegacyWApp[]>([])
  const orders = ref<LegacyWOrder[]>([])
  const zones = ref<SelectItem[]>([])
  const queryMessage = ref('')

  const pagination = reactive({
    page: 1,
    page_size: 20,
    total: 0
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.page_size,
    total: pagination.total
  }))

  const filters = reactive({
    account: '',
    app_id: '',
    status: ''
  })

  const searchItems = computed(() => [
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '订单状态',
        options: statusOptions
      }
    },
    {
      label: '项目',
      key: 'app_id',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '项目筛选',
        options: apps.value.map((item) => ({ label: item.name, value: String(item.app_id) }))
      }
    },
    {
      label: '账号',
      key: 'account',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索账号'
      }
    }
  ])

  const addForm = reactive({
    account: '',
    app_id: undefined as number | undefined,
    dis: 2,
    password: '',
    run_type: 1,
    task_count: 7,
    zone_id: '',
    zone_name: ''
  })

  const selectedApp = computed(() => apps.value.find((item) => item.app_id === addForm.app_id) || null)
  const waitAddCount = computed(() => orders.value.filter((item) => item.status === 'WAITADD').length)

  const appFieldConfig = computed(() => {
    switch (selectedApp.value?.code) {
      case 'bdlp':
        return { accountLabel: '学号 / UID', accountField: 'uid', needPassword: false, zoneField: 'school_name', zoneLabel: '跑区' }
      case 'yyd':
        return { accountLabel: '学号', accountField: 'number', needPassword: true, zoneField: 'school_name', zoneLabel: '学校' }
      case 'keep':
      case 'ymty':
        return { accountLabel: '手机号', accountField: 'phone', needPassword: true, zoneField: 'zone_name', zoneLabel: '跑区' }
      default:
        return { accountLabel: '账号', accountField: 'phone', needPassword: true, zoneField: 'zone_name', zoneLabel: '跑区' }
    }
  })

  const accountLabel = computed(() => appFieldConfig.value.accountLabel)
  const needPassword = computed(() => appFieldConfig.value.needPassword)
  const zoneLabel = computed(() => appFieldConfig.value.zoneLabel)

  const estimatedPrice = computed(() => {
    if (!selectedApp.value) return '0.00'
    const unit = Number(selectedApp.value.price || 0)
    const count = Number(addForm.task_count || 0)
    const value = selectedApp.value.cac_type === 'TS' ? unit * count : unit * count * Number(addForm.dis || 0)
    return value.toFixed(2)
  })

  const getAppName = (appId: number) => apps.value.find((item) => item.app_id === appId)?.name || `#${appId}`

  const getStatusText = (value: string) =>
    (
      {
        ADDING: '下单中',
        NORMAL: '正常',
        REFUND: '已退款',
        REFUNDFAIL: '退款失败',
        WAITADD: '待下单',
        WAITREFUND: '待退款'
      } as Record<string, string>
    )[value] || value

  const getStatusType = (value: string) =>
    (
      {
        ADDING: 'warning',
        NORMAL: 'success',
        REFUND: 'info',
        REFUNDFAIL: 'danger',
        WAITADD: 'warning',
        WAITREFUND: 'warning'
      } as Record<string, 'danger' | 'info' | 'success' | 'warning'>
    )[value] || 'info'

  const { columns, columnChecks } = useTableColumns<LegacyWOrder>(() => [
    {
      prop: 'app_id',
      label: '项目',
      minWidth: 170,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, getAppName(row.app_id)),
          h('p', { class: 'text-xs text-g-500' }, row.agg_order_id || '未下发源台ID')
        ])
    },
    {
      prop: 'account',
      label: '账号信息',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.account || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.password || '无密码')
        ])
    },
    {
      prop: 'school',
      label: '学校 / 跑区',
      minWidth: 180,
      formatter: (row) => h('span', { class: 'text-sm text-g-700' }, row.school || '未填写')
    },
    {
      prop: 'num',
      label: '次数',
      width: 80,
      align: 'center'
    },
    {
      prop: 'cost',
      label: '费用',
      width: 110,
      align: 'right',
      formatter: (row) => h('span', { class: 'font-semibold text-[var(--el-color-danger)]' }, `¥${Number(row.cost || 0).toFixed(2)}`)
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row) =>
        h('div', { class: 'flex flex-col items-center gap-2' }, [
          h(ElTag, { type: getStatusType(row.status), effect: 'plain' }, () => getStatusText(row.status)),
          row.pause ? h(ElTag, { type: 'warning', effect: 'plain' }, () => '已暂停') : null
        ])
    },
    {
      prop: 'updated',
      label: '更新时间',
      minWidth: 160
    },
    {
      prop: 'operation',
      label: '操作',
      width: 260,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { size: 'small', disabled: !row.agg_order_id, onClick: () => handleSync(row) }, () => '同步'),
          row.status === 'WAITADD' ? h(ElButton, { size: 'small', onClick: () => handleResume(row) }, () => '重新提交') : null,
          h(ElButton, { size: 'small', type: 'danger', plain: true, disabled: row.status === 'REFUND' || row.deleted, onClick: () => handleRefund(row) }, () => '退款')
        ])
    }
  ])

  const loadApps = async () => {
    apps.value = (await fetchLegacyWApps()) || []
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyWOrders({
        account: filters.account || undefined,
        app_id: filters.app_id || undefined,
        page: pagination.page,
        page_size: pagination.page_size,
        status: filters.status || undefined
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
    pagination.page_size = size
    loadOrders(1)
  }

  const resetFilters = () => {
    filters.account = ''
    filters.app_id = ''
    filters.status = ''
    loadOrders(1)
  }

  const resetAddForm = () => {
    Object.assign(addForm, {
      account: '',
      app_id: apps.value[0]?.app_id,
      dis: 2,
      password: '',
      run_type: 1,
      task_count: 7,
      zone_id: '',
      zone_name: ''
    })
    zones.value = []
    queryMessage.value = ''
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
  }

  const handleAppChange = () => {
    addForm.account = ''
    addForm.password = ''
    addForm.zone_id = ''
    addForm.zone_name = ''
    zones.value = []
    queryMessage.value = ''
  }

  const handleZoneChange = (value: string) => {
    const current = zones.value.find((item) => item.id === value)
    if (current) {
      addForm.zone_name = current.name
    }
  }

  const normalizeZoneItems = (list: any[]) =>
    list
      .map((item) => ({
        id: String(item?.id || item?.zone_id || item?.school_id || ''),
        name: String(item?.name || item?.zone_name || item?.school_name || '')
      }))
      .filter((item) => item.id || item.name)

  const handleQueryZones = async () => {
    if (!selectedApp.value) {
      ElMessage.warning('请先选择项目')
      return
    }
    if (!addForm.account) {
      ElMessage.warning(`请输入${accountLabel.value}`)
      return
    }
    if (needPassword.value && !addForm.password) {
      ElMessage.warning('请输入密码')
      return
    }

    queryLoading.value = true
    queryMessage.value = ''
    zones.value = []
    try {
      const form: Record<string, any> = {
        [appFieldConfig.value.accountField]: addForm.account
      }
      if (addForm.password) {
        form.password = addForm.password
      }
      const result = await proxyLegacyWAction(selectedApp.value.app_id, `get_${selectedApp.value.code}_zone_data`, { form })
      const payload = result?.data || result
      const list = Array.isArray(payload?.data) ? payload.data : Array.isArray(payload) ? payload : []
      zones.value = normalizeZoneItems(list)
      queryMessage.value = payload?.msg || result?.msg || (zones.value.length ? '查询成功' : '未查询到可用数据')
      if (zones.value.length === 1) {
        addForm.zone_id = zones.value[0]!.id
        addForm.zone_name = zones.value[0]!.name
      }
    } catch (error: any) {
      queryMessage.value = error?.message || '查询失败，可手动填写'
    } finally {
      queryLoading.value = false
    }
  }

  const handleCreate = async () => {
    if (!selectedApp.value) {
      ElMessage.warning('请选择项目')
      return
    }
    if (!addForm.account) {
      ElMessage.warning(`请输入${accountLabel.value}`)
      return
    }
    if (needPassword.value && !addForm.password) {
      ElMessage.warning('请输入密码')
      return
    }
    if (!addForm.zone_name && !addForm.zone_id) {
      ElMessage.warning(`请选择或填写${zoneLabel.value}`)
      return
    }
    if (!addForm.task_count || addForm.task_count < 1) {
      ElMessage.warning('请输入次数')
      return
    }

    const form: Record<string, any> = {
      dis: addForm.dis,
      run_type: addForm.run_type,
      task_list: addForm.task_count
    }
    form[appFieldConfig.value.accountField] = addForm.account
    form[appFieldConfig.value.zoneField] = addForm.zone_name
    if (addForm.zone_id) {
      form.zone_id = addForm.zone_id
    }
    if (addForm.password) {
      form.password = addForm.password
    }

    addLoading.value = true
    try {
      await createLegacyWOrder({
        a_account: addForm.account,
        a_password: addForm.password,
        a_school: addForm.zone_name,
        app_id: selectedApp.value.app_id,
        dis: addForm.dis,
        form,
        task_list: Array.from({ length: addForm.task_count }, () => ({}))
      })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const handleRefund = async (order: LegacyWOrder) => {
    await ElMessageBox.confirm(`确认退款订单 #${order.id}（${order.account}）？`, '退款订单', { type: 'warning' })
    await refundLegacyWOrder(order.id)
    ElMessage.success('退款成功')
    loadOrders(pagination.page)
  }

  const handleSync = async (order: LegacyWOrder) => {
    await syncLegacyWOrder(order.id)
    ElMessage.success('同步成功')
    loadOrders(pagination.page)
  }

  const handleResume = async (order: LegacyWOrder) => {
    await resumeLegacyWOrder(order.id)
    ElMessage.success('重新提交成功')
    loadOrders(pagination.page)
  }

  onMounted(async () => {
    await loadApps()
    resetAddForm()
    await loadOrders(1)
  })
</script>
