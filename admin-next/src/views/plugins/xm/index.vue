<template>
  <div class="plugin-xm-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="resetFilters"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" layout="refresh" @refresh="loadOrders(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">项目 {{ projects.length }} 个</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag type="warning" effect="plain">待处理 {{ processingCount }}</ElTag>
            <ElButton type="primary" plain :disabled="!projects.length" @click="openAddDialog">新增订单</ElButton>
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

    <ElDialog v-model="addVisible" title="新增小米运动订单" width="860px">
      <div class="grid gap-5 lg:grid-cols-[1fr_300px]">
        <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">选择项目</p>
            <ElSelect v-model="addForm.project_id" class="w-full" filterable placeholder="请选择项目" @change="handleProjectChange">
              <ElOption v-for="item in projects" :key="item.id" :label="item.name" :value="item.id" />
            </ElSelect>
            <p v-if="selectedProject?.description" class="mt-2 text-xs text-g-500">{{ selectedProject.description }}</p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">账号</p>
              <ElInput v-model="addForm.account" placeholder="手机号 / 学号" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput
                v-model="addForm.password"
                :disabled="!needPassword"
                :placeholder="needPassword ? '请输入密码' : '当前项目无需密码'"
                show-password
              />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-[1fr_auto]">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">学校 / 跑区</p>
              <ElInput v-model="addForm.school" placeholder="可手动填写，也可由查询结果回填" />
            </div>
            <div class="flex items-end">
              <ElButton
                v-if="showQueryButton"
                class="w-full"
                plain
                :loading="queryLoading"
                @click="handleQueryAccount"
              >
                查询账号
              </ElButton>
            </div>
          </div>

          <ElAlert
            v-if="queryFeedback"
            :type="queryFeedback.type"
            :title="queryFeedback.message"
            show-icon
            :closable="false"
          />

          <div v-if="runRoleOptions.length" class="space-y-3">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">跑步方案</p>
              <div class="flex flex-wrap gap-2">
                <ElButton
                  v-for="(item, index) in runRoleOptions"
                  :key="`${item.label}-${index}`"
                  :type="selectedRunRoleIndex === index ? 'primary' : 'default'"
                  @click="applyRunRole(item.raw, index)"
                >
                  {{ item.label }}
                </ElButton>
              </div>
            </div>

            <div v-if="currentRunTimes.length">
              <p class="mb-2 text-sm font-medium text-g-800">推荐时间段</p>
              <div class="flex flex-wrap gap-2">
                <ElButton
                  v-for="(item, index) in currentRunTimes"
                  :key="`${item.start_time}-${item.end_time}-${index}`"
                  :type="selectedTimeIndex === index ? 'primary' : 'default'"
                  @click="applyRunTime(item, index)"
                >
                  {{ item.start_time || '--:--' }} - {{ item.end_time || '--:--' }}
                </ElButton>
              </div>
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">{{ orderCountLabel }}</p>
              <ElInputNumber v-model="addForm.total_km" class="w-full" :min="1" :max="500" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">开始日期</p>
              <ElDatePicker
                v-model="addForm.start_day"
                class="w-full"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="请选择开始日期"
              />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">开始时间</p>
              <ElTimePicker
                v-model="addForm.start_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="开始时间"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">结束时间</p>
              <ElTimePicker
                v-model="addForm.end_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="结束时间"
              />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">配速</p>
              <ElInputNumber v-model="addForm.pace" class="w-full" :min="0" :step="0.1" :precision="2" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">单次距离</p>
              <ElInputNumber v-model="addForm.distance" class="w-full" :min="0" :step="0.1" :precision="2" />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">跑步周期</p>
            <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="mb-3">
                <ElCheckbox :model-value="checkAllWeek" :indeterminate="weekIndeterminate" @change="toggleAllWeek">
                  全选
                </ElCheckbox>
              </div>
              <ElCheckboxGroup v-model="addForm.run_date" @change="syncWeekState">
                <ElCheckbox v-for="item in weekDayOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </ElCheckbox>
              </ElCheckboxGroup>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">当前项目</span>
              <span class="max-w-[170px] truncate text-right font-medium text-g-900">
                {{ selectedProject?.name || '未选择项目' }}
              </span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">单价</span>
              <span class="font-medium text-g-900">
                {{ selectedProject ? `¥${Number(selectedProject.price || 0).toFixed(2)}` : '-' }}
              </span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">{{ orderCountLabel }}</span>
              <span class="font-medium text-g-900">{{ addForm.total_km || 0 }}</span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">预估扣费</span>
              <span class="font-semibold text-[var(--el-color-danger)]">¥{{ estimatedPrice }}</span>
            </div>
          </div>

          <div class="rounded-custom-sm border-full-d bg-box px-4 py-4 text-sm leading-6 text-g-600">
            下单前先用“查询账号”回填跑步方案，再核对开始日期、周期和时间段。
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

    <ElDialog v-model="addKMVisible" title="增加次数" width="420px">
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-sm text-g-700">订单号 #{{ addKMForm.order_id }}</p>
          <p class="mt-1 text-sm text-g-500">{{ addKMForm.account }}</p>
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">增加次数</p>
          <ElInputNumber v-model="addKMForm.add_km" class="w-full" :min="1" :max="500" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addKMVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addKMLoading" @click="handleAddKM">确认增加</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="logsVisible" title="订单日志" width="720px">
      <ElTable v-loading="logsLoading" :data="logs" size="large">
        <ElTableColumn prop="updated_at" label="时间" min-width="160" />
        <ElTableColumn prop="log_type" label="类型" min-width="120" />
        <ElTableColumn prop="message" label="详情" min-width="280" />
      </ElTable>

      <div class="mt-4 flex justify-end">
        <ElPagination
          background
          layout="total, prev, pager, next"
          :current-page="logsPagination.page"
          :page-size="logsPagination.pageSize"
          :total="logsPagination.total"
          @current-change="loadLogs"
        />
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    addLegacyXMOrderKM,
    createLegacyXMOrder,
    deleteLegacyXMOrder,
    fetchLegacyXMOrderLogs,
    fetchLegacyXMOrders,
    fetchLegacyXMProjects,
    queryLegacyXMRun,
    refundLegacyXMOrder,
    syncLegacyXMOrder,
    type LegacyXMOrder,
    type LegacyXMProject,
    type LegacyXMRunRole,
    type LegacyXMRunTime
  } from '@/api/legacy/plugin-xm'

  defineOptions({ name: 'PluginXMPage' })

  const weekDayOptions = [
    { label: '周一', value: 1 },
    { label: '周二', value: 2 },
    { label: '周三', value: 3 },
    { label: '周四', value: 4 },
    { label: '周五', value: 5 },
    { label: '周六', value: 6 },
    { label: '周日', value: 7 }
  ]

  const statusOptions = [
    { label: '已下单', value: '已下单' },
    { label: '已提交', value: '已提交' },
    { label: '进行中', value: '进行中' },
    { label: '已完成', value: '已完成' },
    { label: '已退款', value: '已退款' },
    { label: '待退款', value: '待退款' },
    { label: '退款失败', value: '退款失败' }
  ]

  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const queryLoading = ref(false)
  const addKMVisible = ref(false)
  const addKMLoading = ref(false)
  const logsVisible = ref(false)
  const logsLoading = ref(false)

  const projects = ref<LegacyXMProject[]>([])
  const orders = ref<LegacyXMOrder[]>([])
  const logs = ref<any[]>([])
  const queryResult = ref<any>(null)
  const runRoleOptions = ref<Array<{ label: string; raw: LegacyXMRunRole }>>([])
  const selectedRunRoleIndex = ref<number | null>(null)
  const selectedTimeIndex = ref<number | null>(null)

  const pagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.pageSize,
    total: pagination.total
  }))

  const logsPagination = reactive({
    orderId: 0,
    page: 1,
    pageSize: 10,
    total: 0
  })

  const filters = reactive({
    account: '',
    order_id: '',
    project: '',
    status: ''
  })

  const searchItems = computed(() => [
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '状态筛选',
        options: statusOptions
      }
    },
    {
      label: '项目',
      key: 'project',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '项目筛选',
        options: projects.value.map((item) => ({ label: item.name, value: String(item.id) }))
      }
    },
    {
      label: '关键词',
      key: 'account',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索账号、学校或订单号'
      }
    }
  ])

  const addForm = reactive({
    project_id: undefined as number | undefined,
    school: '',
    account: '',
    password: '',
    total_km: 3,
    run_date: [1, 2, 3, 4, 5, 6, 7] as number[],
    start_day: '',
    start_time: '06:00',
    end_time: '08:00',
    type: undefined as number | undefined,
    pace: undefined as number | undefined,
    distance: undefined as number | undefined
  })

  const addKMForm = reactive({
    order_id: 0,
    add_km: 1,
    account: ''
  })

  const checkAllWeek = ref(true)
  const weekIndeterminate = ref(false)

  const selectedProject = computed(() => projects.value.find((item) => item.id === addForm.project_id) || null)
  const showQueryButton = computed(() => Number(selectedProject.value?.query) === 1)
  const needPassword = computed(() => !selectedProject.value || Number(selectedProject.value.password) === 1)
  const processingCount = computed(() =>
    orders.value.filter((item) => ['已下单', '已提交', '进行中'].includes(item.status_name)).length
  )
  const currentRunTimes = computed<LegacyXMRunTime[]>(() =>
    Array.isArray(runRoleOptions.value[selectedRunRoleIndex.value ?? -1]?.raw?.run_times)
      ? (runRoleOptions.value[selectedRunRoleIndex.value ?? -1]?.raw?.run_times as LegacyXMRunTime[])
      : []
  )
  const orderCountLabel = computed(() => '下单次数')
  const estimatedPrice = computed(() =>
    ((Number(selectedProject.value?.price || 0) || 0) * Number(addForm.total_km || 0)).toFixed(2)
  )
  const queryFeedback = computed(() => getQueryFeedback(queryResult.value))

  const syncWeekState = () => {
    const count = addForm.run_date.length
    checkAllWeek.value = count === weekDayOptions.length
    weekIndeterminate.value = count > 0 && count < weekDayOptions.length
  }

  const toggleAllWeek = (value: boolean | string | number) => {
    if (value) {
      addForm.run_date = weekDayOptions.map((item) => item.value)
    } else {
      addForm.run_date = []
    }
    syncWeekState()
  }

  const resetRunRoleState = () => {
    runRoleOptions.value = []
    selectedRunRoleIndex.value = null
    selectedTimeIndex.value = null
  }

  const resetAddForm = () => {
    Object.assign(addForm, {
      project_id: projects.value[0]?.id,
      school: '',
      account: '',
      password: '',
      total_km: 3,
      run_date: [1, 2, 3, 4, 5, 6, 7],
      start_day: '',
      start_time: '06:00',
      end_time: '08:00',
      type: undefined,
      pace: undefined,
      distance: undefined
    })
    queryResult.value = null
    resetRunRoleState()
    syncWeekState()
  }

  const getProjectName = (projectId: number) => projects.value.find((item) => item.id === projectId)?.name || `#${projectId}`

  const getStatusType = (status: string) => {
    const map: Record<string, 'danger' | 'info' | 'success' | 'warning'> = {
      已下单: 'info',
      已提交: 'warning',
      进行中: 'success',
      已完成: 'success',
      已退款: 'info',
      待退款: 'warning',
      退款失败: 'danger'
    }
    return map[status] || 'info'
  }

  const canAddKM = (order: LegacyXMOrder) => !['已退款', '已删除'].includes(order.status_name)
  const canRefund = (order: LegacyXMOrder) => !['已退款', '已删除'].includes(order.status_name)

  const { columns, columnChecks } = useTableColumns<LegacyXMOrder>(() => [
    {
      prop: 'project_id',
      label: '项目 / 类型',
      minWidth: 160,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, getProjectName(row.project_id)),
          h('p', { class: 'text-xs text-g-500' }, row.type || '默认类型')
        ])
    },
    {
      prop: 'account',
      label: '账号信息',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.account || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.password || '无密码展示'),
          h('p', { class: 'text-xs text-g-500' }, row.school || '未填写学校/跑区')
        ])
    },
    {
      prop: 'total_km',
      label: '跑步参数',
      minWidth: 190,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, `总次数 ${row.total_km}，已跑 ${row.run_km ?? 0}`),
          h('p', { class: 'text-xs text-g-500' }, `配速 ${row.pace ?? '-'}，距离 ${row.distance ?? '-'}`),
          h('p', { class: 'text-xs text-g-500' }, `${row.start_time || '--:--'} - ${row.end_time || '--:--'}`)
        ])
    },
    {
      prop: 'status_name',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row) => h(ElTag, { type: getStatusType(row.status_name), effect: 'plain' }, () => row.status_name)
    },
    {
      prop: 'deduction',
      label: '扣费',
      width: 110,
      align: 'right',
      formatter: (row) => h('span', { class: 'font-semibold text-[var(--el-color-danger)]' }, `¥${Number(row.deduction || 0).toFixed(2)}`)
    },
    {
      prop: 'updated_at',
      label: '更新时间',
      minWidth: 160
    },
    {
      prop: 'operation',
      label: '操作',
      width: 340,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { size: 'small', onClick: () => openLogsDialog(row) }, () => '日志'),
          h(ElButton, { size: 'small', disabled: !canAddKM(row), onClick: () => openAddKMDialog(row) }, () => '加次'),
          h(ElButton, { size: 'small', onClick: () => handleSync(row) }, () => '同步'),
          h(ElButton, { size: 'small', type: 'warning', plain: true, disabled: !canRefund(row), onClick: () => handleRefund(row) }, () => '退款'),
          h(ElButton, { size: 'small', type: 'danger', plain: true, disabled: row.is_deleted, onClick: () => handleDelete(row) }, () => '删除')
        ])
    }
  ])

  const normalizeQueryRows = (result: any) => {
    if (Array.isArray(result?.data)) return result.data
    if (Array.isArray(result)) return result
    if (result?.data && typeof result.data === 'object') return [result.data]
    return []
  }

  const getQueryFeedback = (result: any) => {
    if (!result) return null
    const rows = normalizeQueryRows(result)
    const firstRow = rows[0] || null

    if (Number(firstRow?.status) === -1) {
      return {
        type: 'error' as const,
        message: firstRow?.error || firstRow?.msg || result?.msg || '查询失败'
      }
    }

    if (Number(result?.code) === 200 || Number(result?.code) === 0 || result?.msg) {
      return {
        type: 'success' as const,
        message: result?.msg || '查询成功'
      }
    }

    return {
      type: 'error' as const,
      message: result?.msg || '查询失败'
    }
  }

  const applyRunTime = (time: LegacyXMRunTime, index: number) => {
    selectedTimeIndex.value = index
    addForm.start_time = time.start_time || ''
    addForm.end_time = time.end_time || ''
  }

  const applyRunRole = (role: LegacyXMRunRole, index: number) => {
    selectedRunRoleIndex.value = index
    selectedTimeIndex.value = null

    const totalKM = Number(role.total_km)
    if (Number.isFinite(totalKM) && totalKM > 0) {
      addForm.total_km = totalKM
    }
    addForm.run_date = Array.isArray(role.run_date) && role.run_date.length ? [...role.run_date] : []
    addForm.start_day = role.start_day || addForm.start_day
    const typeValue = Number(role.type)
    addForm.type = Number.isFinite(typeValue) ? typeValue : undefined
    syncWeekState()

    if (Array.isArray(role.run_times) && role.run_times.length === 1) {
      applyRunTime(role.run_times[0]!, 0)
    }
  }

  const handleProjectChange = () => {
    addForm.school = ''
    addForm.password = ''
    queryResult.value = null
    resetRunRoleState()
  }

  const loadProjects = async () => {
    projects.value = (await fetchLegacyXMProjects()) || []
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyXMOrders({
        page: pagination.page,
        page_size: pagination.pageSize,
        account: filters.account || undefined,
        order_id: filters.order_id || undefined,
        project: filters.project || undefined,
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
    pagination.pageSize = size
    loadOrders(1)
  }

  const resetFilters = () => {
    filters.account = ''
    filters.order_id = ''
    filters.project = ''
    filters.status = ''
    loadOrders(1)
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
  }

  const handleQueryAccount = async () => {
    if (!addForm.project_id) {
      ElMessage.warning('请先选择项目')
      return
    }
    if (!addForm.account) {
      ElMessage.warning('请输入账号')
      return
    }
    if (needPassword.value && !addForm.password) {
      ElMessage.warning('当前项目需要密码')
      return
    }

    queryLoading.value = true
    try {
      const result = await queryLegacyXMRun({
        project_id: addForm.project_id,
        account: addForm.account,
        password: addForm.password || undefined
      })
      queryResult.value = result
      const rows = normalizeQueryRows(result)
      const first = rows[0] || {}
      if (first?.school) {
        addForm.school = first.school
      }
      if (Array.isArray(first?.run_roles) && first.run_roles.length) {
        runRoleOptions.value = first.run_roles.map((item: LegacyXMRunRole, index: number) => ({
          label: item.run_type || `方案 ${index + 1}`,
          raw: item
        }))
        applyRunRole(runRoleOptions.value[0]!.raw, 0)
      } else {
        resetRunRoleState()
      }
      if (getQueryFeedback(result)?.type === 'success') {
        ElMessage.success('账号信息查询成功')
      }
    } finally {
      queryLoading.value = false
    }
  }

  const handleCreate = async () => {
    if (!addForm.project_id) {
      ElMessage.warning('请选择项目')
      return
    }
    if (!addForm.account) {
      ElMessage.warning('请输入账号')
      return
    }
    if (needPassword.value && !addForm.password) {
      ElMessage.warning('请输入密码')
      return
    }
    if (!addForm.school) {
      ElMessage.warning('请输入学校或跑区')
      return
    }
    if (!addForm.total_km || addForm.total_km < 1) {
      ElMessage.warning(`请输入${orderCountLabel.value}`)
      return
    }
    if (!addForm.run_date.length) {
      ElMessage.warning('请选择跑步周期')
      return
    }
    if (!addForm.start_day || !addForm.start_time || !addForm.end_time) {
      ElMessage.warning('请设置开始日期和时间段')
      return
    }

    addLoading.value = true
    try {
      const payload: Record<string, any> = {
        project_id: addForm.project_id,
        school: addForm.school,
        account: addForm.account,
        password: addForm.password || undefined,
        total_km: addForm.total_km,
        run_date: addForm.run_date,
        start_day: addForm.start_day,
        start_time: addForm.start_time,
        end_time: addForm.end_time
      }
      if (addForm.type !== undefined) payload.type = addForm.type
      if (addForm.pace !== undefined) payload.pace = addForm.pace
      if (addForm.distance !== undefined) payload.distance = addForm.distance

      await createLegacyXMOrder(payload)
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openAddKMDialog = (order: LegacyXMOrder) => {
    addKMForm.order_id = order.id
    addKMForm.account = order.account
    addKMForm.add_km = 1
    addKMVisible.value = true
  }

  const handleAddKM = async () => {
    if (addKMForm.add_km < 1) {
      ElMessage.warning('增加次数至少为 1')
      return
    }
    addKMLoading.value = true
    try {
      await addLegacyXMOrderKM(addKMForm.order_id, addKMForm.add_km)
      ElMessage.success('增加次数成功')
      addKMVisible.value = false
      loadOrders(pagination.page)
    } finally {
      addKMLoading.value = false
    }
  }

  const handleRefund = async (order: LegacyXMOrder) => {
    await ElMessageBox.confirm(`确认退款订单 #${order.id}（${order.account}）？`, '退款订单', { type: 'warning' })
    await refundLegacyXMOrder(order.id)
    ElMessage.success('退款成功')
    loadOrders(pagination.page)
  }

  const handleDelete = async (order: LegacyXMOrder) => {
    await ElMessageBox.confirm(`确认删除订单 #${order.id}（${order.account}）？`, '删除订单', { type: 'warning' })
    await deleteLegacyXMOrder(order.id)
    ElMessage.success('删除成功')
    loadOrders(pagination.page)
  }

  const handleSync = async (order: LegacyXMOrder) => {
    await syncLegacyXMOrder(order.id)
    ElMessage.success('同步成功')
    loadOrders(pagination.page)
  }

  const loadLogs = async (page = logsPagination.page) => {
    logsLoading.value = true
    logsPagination.page = page
    try {
      const result = await fetchLegacyXMOrderLogs(logsPagination.orderId, logsPagination.page, logsPagination.pageSize)
      const inner = result?.data
      if (inner && Array.isArray(inner.data)) {
        logs.value = inner.data
        logsPagination.total = Number(inner.total || inner.data.length)
      } else if (Array.isArray(inner)) {
        logs.value = inner
        logsPagination.total = inner.length
      } else {
        logs.value = []
        logsPagination.total = 0
      }
    } finally {
      logsLoading.value = false
    }
  }

  const openLogsDialog = (order: LegacyXMOrder) => {
    logsPagination.orderId = order.id
    logsPagination.page = 1
    logsVisible.value = true
    loadLogs(1)
  }

  onMounted(async () => {
    await loadProjects()
    resetAddForm()
    await loadOrders(1)
  })
</script>
