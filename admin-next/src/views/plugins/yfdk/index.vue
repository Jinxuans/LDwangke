<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-[1.2fr_180px_220px_auto]">
        <ElInput v-model="filters.keyword" clearable placeholder="搜索账号、姓名或学校" @keyup.enter="loadOrders(1)" />
        <ElSelect v-model="filters.status" clearable placeholder="状态筛选">
          <ElOption v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </ElSelect>
        <ElSelect v-model="filters.cid" clearable filterable placeholder="项目筛选">
          <ElOption v-for="item in projects" :key="item.cid" :label="item.name" :value="item.cid" />
        </ElSelect>
        <div class="flex flex-wrap gap-3">
          <ElButton type="primary" @click="loadOrders(1)">查询</ElButton>
          <ElButton plain @click="resetFilters">重置</ElButton>
        </div>
      </div>
    </section>

    <section class="art-card-sm overflow-hidden">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadOrders(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag type="success" effect="plain">运行中 {{ runningCount }}</ElTag>
            <ElTag type="warning" effect="plain">即将到期 {{ expiringCount }}</ElTag>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="orders" class="w-full" size="large">
        <ElTableColumn label="账号信息" min-width="220">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-semibold text-g-900">{{ row.username }}</p>
              <p class="text-xs text-g-500">密码 {{ row.password || '-' }}</p>
              <p class="text-xs text-g-500">{{ row.school || row.name || '未填写学校/姓名' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="项目" min-width="120">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-medium text-g-900">{{ getProjectName(row.cid) }}</p>
              <p class="text-xs text-g-500">CID {{ row.cid }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="订单信息" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-medium text-[var(--el-color-danger)]">¥{{ Number(row.total_fee || 0).toFixed(2) }}</p>
              <p class="text-xs text-g-500">{{ row.day }} 天，单日 ¥{{ Number(row.daily_fee || 0).toFixed(2) }}</p>
              <p class="text-xs text-g-500">创建于 {{ row.create_time || '-' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="状态" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusType(row)" effect="plain">{{ getStatusText(row) }}</ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="到期时间" min-width="150">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-medium text-g-900">{{ row.endtime || '-' }}</p>
              <p class="text-xs text-g-500">{{ formatRemainDays(row.endtime) }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="最近日志" min-width="220">
          <template #default="{ row }">
            <span class="line-clamp-2 text-sm text-g-500">{{ row.mark || '暂无日志' }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <ElButton size="small" @click="handleToggleStatus(row)">
                {{ row.status === 1 ? '暂停' : '启用' }}
              </ElButton>
              <ElButton size="small" @click="handleManualClock(row)">手动打卡</ElButton>
              <ElButton size="small" @click="openRenewDialog(row)">续费</ElButton>
              <ElButton size="small" @click="openLogsDialog(row)">日志</ElButton>
              <ElButton size="small" type="danger" plain @click="handleDelete(row)">删除</ElButton>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="flex justify-end border-t-d px-5 py-4">
        <ElPagination
          background
          layout="total, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.limit"
          :total="pagination.total"
          @current-change="loadOrders"
        />
      </div>
    </section>

    <ElDialog v-model="addVisible" title="新增 YF 打卡订单" width="760px">
      <div class="grid gap-5 lg:grid-cols-[1fr_280px]">
        <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">选择项目</p>
            <ElSelect v-model="addForm.cid" class="w-full" filterable placeholder="请选择项目">
              <ElOption v-for="item in projects" :key="item.cid" :label="item.name" :value="item.cid" />
            </ElSelect>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">账号</p>
              <ElInput v-model="addForm.user" placeholder="请输入账号" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="addForm.pass" placeholder="请输入密码" show-password />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">打卡天数</p>
              <ElInputNumber v-model="addForm.day" class="w-full" :min="1" :max="365" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">学校</p>
              <ElInput v-model="addForm.school" placeholder="选填" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">姓名</p>
              <ElInput v-model="addForm.name" placeholder="选填" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">邮箱</p>
              <ElInput v-model="addForm.email" placeholder="选填" />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">打卡地址</p>
            <ElInput v-model="addForm.address" placeholder="选填" />
          </div>

          <div class="grid gap-4 md:grid-cols-3">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
              <ElInput v-model="addForm.week" placeholder="1,2,3,4,5" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">上班时间</p>
              <ElInput v-model="addForm.worktime" placeholder="08:00" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">下班时间</p>
              <ElInput v-model="addForm.offtime" placeholder="17:30" />
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">项目</span>
              <span class="max-w-[160px] truncate text-right font-medium text-g-900">
                {{ getProjectName(addForm.cid) }}
              </span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">打卡天数</span>
              <span class="font-medium text-g-900">{{ addForm.day || 0 }} 天</span>
            </div>
            <div class="mt-3 text-sm leading-6 text-g-600">
              {{ priceHint || '可先点击“试算价格”确认费用。' }}
            </div>
          </div>

          <div class="rounded-custom-sm border-full-d bg-box p-4">
            <p class="mb-3 text-sm font-medium text-g-800">附加选项</p>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">日报</span>
                <ElSwitch v-model="addForm.day_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">周报</span>
                <ElSwitch v-model="addForm.week_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">月报</span>
                <ElSwitch v-model="addForm.month_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">跳过节假日</span>
                <ElSwitch v-model="addForm.skip_holidays" :active-value="1" :inactive-value="0" />
              </div>
            </div>
          </div>

          <div class="flex gap-3">
            <ElButton plain class="w-full" :loading="priceLoading" @click="handleCalculatePrice">试算价格</ElButton>
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

    <ElDialog v-model="logsVisible" :title="logsTitle" width="640px">
      <ElScrollbar max-height="420px">
        <div v-if="logsLoading" class="py-10 text-center text-sm text-g-500">日志加载中...</div>
        <div v-else-if="!logs.length" class="py-10 text-center text-sm text-g-500">暂无日志</div>
        <div v-else class="space-y-3">
          <article
            v-for="(item, index) in logs"
            :key="`${index}-${item.created_at || item.time || item.message}`"
            class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4"
          >
            <div class="flex items-start justify-between gap-4">
              <p class="text-sm leading-6 text-g-800">{{ item.content || item.message || '-' }}</p>
              <span class="shrink-0 text-xs text-g-500">{{ item.created_at || item.time || '-' }}</span>
            </div>
          </article>
        </div>
      </ElScrollbar>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createLegacyYFDKOrder,
    deleteLegacyYFDKOrder,
    fetchLegacyYFDKOrderLogs,
    fetchLegacyYFDKOrders,
    fetchLegacyYFDKPrice,
    fetchLegacyYFDKProjects,
    manualClockLegacyYFDKOrder,
    renewLegacyYFDKOrder,
    saveLegacyYFDKOrder,
    type LegacyYFDKLogItem,
    type LegacyYFDKOrder,
    type LegacyYFDKProject
  } from '@/api/legacy/plugin-yfdk'

  defineOptions({ name: 'PluginYFDKPage' })

  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const priceLoading = ref(false)
  const logsVisible = ref(false)
  const logsLoading = ref(false)
  const logsTitle = ref('订单日志')
  const priceHint = ref('')

  const projects = ref<LegacyYFDKProject[]>([])
  const orders = ref<LegacyYFDKOrder[]>([])
  const logs = ref<LegacyYFDKLogItem[]>([])

  const pagination = reactive({
    page: 1,
    limit: 10,
    total: 0
  })

  const filters = reactive({
    keyword: '',
    status: '',
    cid: ''
  })

  const addForm = reactive({
    cid: '',
    user: '',
    pass: '',
    day: 30,
    school: '',
    name: '',
    email: '',
    address: '',
    week: '1,2,3,4,5',
    worktime: '08:00',
    offtime: '17:30',
    offwork: 0,
    day_report: 1,
    week_report: 0,
    month_report: 0,
    skip_holidays: 0
  })

  const statusOptions = [
    { label: '运行中', value: '1' },
    { label: '已暂停', value: '0' },
    { label: '已过期', value: '2' },
    { label: '即将到期', value: '3' }
  ]

  const runningCount = computed(() => orders.value.filter((item) => item.status === 1).length)
  const expiringCount = computed(() => orders.value.filter((item) => getRemainDays(item.endtime) <= 3 && getRemainDays(item.endtime) >= 0).length)

  const getRemainDays = (value?: string) => {
    if (!value) return 0
    const target = new Date(value).getTime()
    if (!target) return 0
    return Math.ceil((target - Date.now()) / 86400000)
  }

  const formatRemainDays = (value?: string) => {
    const days = getRemainDays(value)
    if (days < 0) return '已过期'
    if (days === 0) return '今天到期'
    return `剩余 ${days} 天`
  }

  const getProjectName = (cid: string) => projects.value.find((item) => item.cid === cid)?.name || cid

  const getStatusText = (order: LegacyYFDKOrder) => {
    const days = getRemainDays(order.endtime)
    if (days < 0) return '已过期'
    if (order.status === 0) return '已暂停'
    if (order.status === 1) return '运行中'
    return '未知'
  }

  const getStatusType = (order: LegacyYFDKOrder) => {
    const days = getRemainDays(order.endtime)
    if (days < 0) return 'danger'
    if (order.status === 0) return 'info'
    if (order.status === 1) return days <= 3 ? 'warning' : 'success'
    return 'info'
  }

  const resetAddForm = () => {
    Object.assign(addForm, {
      cid: projects.value[0]?.cid || '',
      user: '',
      pass: '',
      day: 30,
      school: '',
      name: '',
      email: '',
      address: '',
      week: '1,2,3,4,5',
      worktime: '08:00',
      offtime: '17:30',
      offwork: 0,
      day_report: 1,
      week_report: 0,
      month_report: 0,
      skip_holidays: 0
    })
    priceHint.value = ''
  }

  const loadProjects = async () => {
    projects.value = (await fetchLegacyYFDKProjects()) || []
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyYFDKOrders({
        page: pagination.page,
        limit: pagination.limit,
        keyword: filters.keyword || undefined,
        status: filters.status || undefined,
        cid: filters.cid || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const resetFilters = () => {
    filters.keyword = ''
    filters.status = ''
    filters.cid = ''
    loadOrders(1)
  }

  const openAddDialog = async () => {
    if (!projects.value.length) {
      await loadProjects()
    }
    resetAddForm()
    addVisible.value = true
  }

  const handleCalculatePrice = async () => {
    if (!addForm.cid) {
      ElMessage.warning('请先选择项目')
      return
    }
    priceLoading.value = true
    try {
      const result = await fetchLegacyYFDKPrice(addForm.cid, Number(addForm.day || 0))
      priceHint.value = result?.msg || `预计费用 ¥${Number(result?.price || 0).toFixed(2)}`
    } finally {
      priceLoading.value = false
    }
  }

  const handleCreate = async () => {
    if (!addForm.cid || !addForm.user || !addForm.pass || !addForm.day) {
      ElMessage.warning('请填写项目、账号、密码和打卡天数')
      return
    }

    if (!priceHint.value) {
      await handleCalculatePrice()
    }

    await ElMessageBox.confirm(
      `确认给账号 ${addForm.user} 提交 ${addForm.day} 天订单？${priceHint.value ? `\n${priceHint.value}` : ''}`,
      '确认下单',
      { type: 'warning' }
    )

    addLoading.value = true
    try {
      await createLegacyYFDKOrder({ ...addForm })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openRenewDialog = async (order: LegacyYFDKOrder) => {
    const { value } = await ElMessageBox.prompt(`为账号 ${order.username} 输入续费天数`, '续费订单', {
      inputPattern: /^[1-9]\d*$/,
      inputErrorMessage: '请输入正整数'
    })
    if (!value) {
      return
    }
    await renewLegacyYFDKOrder(order.id, Number(value))
    ElMessage.success('续费成功')
    loadOrders(pagination.page)
  }

  const handleDelete = async (order: LegacyYFDKOrder) => {
    await ElMessageBox.confirm(`确认删除订单 ${order.username}？未到期部分将自动退款。`, '删除订单', {
      type: 'warning'
    })
    await deleteLegacyYFDKOrder(order.id)
    ElMessage.success('删除成功')
    loadOrders(pagination.page)
  }

  const handleManualClock = async (order: LegacyYFDKOrder) => {
    await manualClockLegacyYFDKOrder(order.id)
    ElMessage.success('手动打卡任务已提交')
    loadOrders(pagination.page)
  }

  const handleToggleStatus = async (order: LegacyYFDKOrder) => {
    await saveLegacyYFDKOrder({
      id: order.id,
      status: order.status === 1 ? 0 : 1
    })
    ElMessage.success(order.status === 1 ? '订单已暂停' : '订单已启用')
    loadOrders(pagination.page)
  }

  const openLogsDialog = async (order: LegacyYFDKOrder) => {
    logsVisible.value = true
    logsLoading.value = true
    logsTitle.value = `${order.username} 的订单日志`
    try {
      const result = await fetchLegacyYFDKOrderLogs(order.id)
      logs.value = Array.isArray(result) ? result : []
    } finally {
      logsLoading.value = false
    }
  }

  onMounted(async () => {
    await loadProjects()
    resetAddForm()
    await loadOrders(1)
  })
</script>
