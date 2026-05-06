<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <ElAlert
        type="warning"
        :closable="false"
        show-icon
        title="下单后请提醒学生退出账号，不是关闭 App。跑步时段建议保留足够跨度，避免过窄。"
      />

      <div class="mt-4 grid gap-4 xl:grid-cols-[160px_140px_1fr_auto]">
        <ElSelect v-model="filters.status" clearable placeholder="订单状态">
          <ElOption v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </ElSelect>
        <ElSelect v-model="filters.searchType" placeholder="搜索项">
          <ElOption v-for="item in searchTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </ElSelect>
        <ElInput v-model="filters.keyword" clearable placeholder="搜索订单ID、账号、密码或UID" @keyup.enter="loadOrders(1)" />
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
            <ElTag type="warning" effect="plain">待处理 {{ pendingCount }}</ElTag>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="orders" size="large">
        <ElTableColumn label="账号信息" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-semibold text-g-900">{{ row.user }}</p>
              <p class="text-xs text-g-500">{{ row.pass || '无密码' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="学校" min-width="160">
          <template #default="{ row }">
            <p class="text-sm text-g-800">{{ row.school || '自动识别' }}</p>
          </template>
        </ElTableColumn>

        <ElTableColumn label="跑步参数" min-width="220">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm text-g-800">{{ getRunTypeText(row.run_type) }}</p>
              <p class="text-xs text-g-500">{{ row.distance }} km，{{ getRunWeekText(row.run_week) }}</p>
              <p class="text-xs text-g-500">{{ formatTimeRange(row.start_hour, row.start_minute, row.end_hour, row.end_minute) }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="运行状态" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="row.is_run === 1 ? 'success' : 'info'" effect="plain">
              {{ row.is_run === 1 ? '运行中' : '已暂停' }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="订单状态" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusType(row.status)" effect="plain">{{ getStatusText(row.status) }}</ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="费用" width="120" align="right">
          <template #default="{ row }">
            <div class="leading-5">
              <p class="font-semibold text-[var(--el-color-danger)]">¥{{ Number(row.fees || 0).toFixed(2) }}</p>
              <p v-if="row.refund_money" class="text-xs text-emerald-600">退 ¥{{ Number(row.refund_money || 0).toFixed(2) }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="备注" min-width="180">
          <template #default="{ row }">
            <span class="line-clamp-2 text-sm text-g-500">{{ row.remarks || '暂无备注' }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn prop="addtime" label="下单时间" min-width="160" />

        <ElTableColumn label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <ElButton size="small" @click="openDetailDialog(row)">详情</ElButton>
              <ElButton size="small" @click="openRemarksDialog(row)">备注</ElButton>
              <ElButton size="small" @click="handleToggleRun(row)">
                {{ row.is_run === 1 ? '暂停' : '开启' }}
              </ElButton>
              <ElButton size="small" @click="handleSync(row)">同步</ElButton>
              <ElButton size="small" type="danger" plain :disabled="row.status === 4 || row.status === 5" @click="handleRefund(row)">
                退款
              </ElButton>
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

    <ElDialog v-model="addVisible" title="新增运动世界订单" width="760px">
      <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">跑步类型</p>
            <ElSelect v-model="addForm.run_type" class="w-full">
              <ElOption v-for="item in runTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
            </ElSelect>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">学校</p>
            <ElSelect v-model="addForm.school" class="w-full" clearable filterable placeholder="自动识别或手动选择">
              <ElOption label="自动识别" value="" />
              <ElOption v-for="item in schools" :key="item.school_name" :label="item.school_name" :value="item.school_name" />
            </ElSelect>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">账号</p>
              <ElInput v-model="addForm.user" placeholder="请输入账号" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="addForm.pass" show-password placeholder="请输入密码" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">总公里数</p>
              <ElInputNumber v-model="addForm.distance" class="w-full" :min="0.1" :max="500" :step="0.1" :precision="1" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">备注</p>
              <ElInput v-model="addForm.remarks" placeholder="选填" />
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

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">跑步周期</p>
            <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="mb-3">
                <ElCheckbox :model-value="checkAllWeek" :indeterminate="weekIndeterminate" @change="toggleAllWeek">
                  全选
                </ElCheckbox>
              </div>
              <ElCheckboxGroup v-model="addForm.run_week" @change="syncWeekState">
                <ElCheckbox v-for="item in weekOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </ElCheckbox>
              </ElCheckboxGroup>
            </div>
          </div>

        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <div class="grid gap-3 text-sm md:grid-cols-2">
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">当前配置</span>
              <span class="font-medium text-g-900">{{ getRunTypeText(addForm.run_type) }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">总公里数</span>
              <span class="font-medium text-g-900">{{ Number(addForm.distance || 0).toFixed(1) }} km</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">跑步周期</span>
              <span class="truncate text-right font-medium text-g-900">{{ getRunWeekText(addForm.run_week.join(',')) }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">预估费用</span>
              <span v-if="estimatedPrice !== null" class="font-semibold text-[var(--el-color-danger)]">¥{{ estimatedPrice.toFixed(2) }}</span>
              <span v-else-if="priceLoading" class="text-g-500">计算中...</span>
              <span v-else class="text-g-500">待估算</span>
            </div>
          </div>
          <p class="mt-2 text-xs text-g-500">实际扣费以后端为准，若不选学校则由后端自动识别。</p>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addLoading" @click="handleCreate">确认下单</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="remarksVisible" title="编辑备注" width="460px">
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-sm text-g-700">订单 #{{ remarksForm.id }}</p>
          <p class="mt-1 text-sm text-g-500">{{ remarksForm.user }}</p>
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">备注</p>
          <ElInput v-model="remarksForm.remarks" type="textarea" :rows="4" placeholder="请输入备注" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="remarksVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="remarksLoading" @click="handleSaveRemarks">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="detailVisible" title="订单详情" width="720px">
      <ElTabs v-model="detailTab">
        <ElTabPane label="订单信息" name="info">
          <div
            v-if="detailRecord?.info"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-7 text-g-700"
            v-html="detailRecord.info"
          />
          <div v-else class="py-10 text-center text-sm text-g-500">暂无订单信息</div>
        </ElTabPane>
        <ElTabPane label="退款/操作信息" name="tmp_info">
          <div
            v-if="detailRecord?.tmp_info"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-7 text-g-700"
            v-html="detailRecord.tmp_info"
          />
          <div v-else class="py-10 text-center text-sm text-g-500">暂无操作信息</div>
        </ElTabPane>
      </ElTabs>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createLegacyYDSJOrder,
    editLegacyYDSJRemarks,
    fetchLegacyYDSJOrders,
    fetchLegacyYDSJPrice,
    fetchLegacyYDSJSchools,
    refundLegacyYDSJOrder,
    syncLegacyYDSJOrder,
    toggleLegacyYDSJRun,
    type LegacyYDSJOrder
  } from '@/api/legacy/plugin-ydsj'

  defineOptions({ name: 'PluginYDSJPage' })

  const weekOptions = [
    { label: '周一', value: '1' },
    { label: '周二', value: '2' },
    { label: '周三', value: '3' },
    { label: '周四', value: '4' },
    { label: '周五', value: '5' },
    { label: '周六', value: '6' },
    { label: '周日', value: '7' }
  ]

  const runTypeOptions = [
    { label: '运动世界晨跑', value: 0 },
    { label: '运动世界课外跑', value: 1 },
    { label: '小步点课外跑', value: 2 },
    { label: '小步点晨跑', value: 3 }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '等待处理', value: '1' },
    { label: '处理成功', value: '2' },
    { label: '处理失败', value: '3' },
    { label: '退款成功', value: '4' },
    { label: '申请退款', value: '5' }
  ]

  const searchTypeOptions = [
    { label: '订单ID', value: '1' },
    { label: '下单账号', value: '2' },
    { label: '下单密码', value: '3' },
    { label: '用户UID', value: '4' }
  ]

  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const priceLoading = ref(false)
  const remarksVisible = ref(false)
  const remarksLoading = ref(false)
  const detailVisible = ref(false)

  const schools = ref<any[]>([])
  const orders = ref<LegacyYDSJOrder[]>([])
  const detailRecord = ref<LegacyYDSJOrder | null>(null)
  const detailTab = ref('info')
  const estimatedPrice = ref<null | number>(null)
  const runningCount = computed(() => orders.value.filter((item) => item.is_run === 1).length)
  const pendingCount = computed(() => orders.value.filter((item) => item.status === 1).length)

  const pagination = reactive({
    limit: 20,
    page: 1,
    total: 0
  })

  const filters = reactive({
    keyword: '',
    searchType: '1',
    status: ''
  })

  const addForm = reactive({
    distance: 2,
    end_time: '21:00',
    pass: '',
    remarks: '',
    run_type: 1,
    run_week: ['1', '2', '3', '4', '5', '6', '7'] as string[],
    school: '',
    start_time: '12:00',
    user: ''
  })

  const remarksForm = reactive({
    id: 0,
    remarks: '',
    user: ''
  })

  const checkAllWeek = ref(false)
  const weekIndeterminate = ref(false)

  const syncWeekState = () => {
    const count = addForm.run_week.length
    checkAllWeek.value = count === weekOptions.length
    weekIndeterminate.value = count > 0 && count < weekOptions.length
  }

  const toggleAllWeek = (value: boolean | string | number) => {
    addForm.run_week = value ? weekOptions.map((item) => item.value) : []
    syncWeekState()
  }

  const splitTime = (value: string) => {
    const [hour = '00', minute = '00'] = String(value || '').split(':')
    return { hour, minute }
  }

  const formatTimeRange = (startHour: string, startMinute: string, endHour: string, endMinute: string) =>
    `${String(startHour).padStart(2, '0')}:${String(startMinute).padStart(2, '0')} - ${String(endHour).padStart(2, '0')}:${String(endMinute).padStart(2, '0')}`

  const getRunTypeText = (value: number) => runTypeOptions.find((item) => item.value === value)?.label || `类型 ${value}`

  const getRunWeekText = (value: string) => {
    if (!value) return '未设置周期'
    const map: Record<string, string> = { '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六', '7': '日' }
    return value
      .split(',')
      .filter(Boolean)
      .map((item) => `周${map[item] || item}`)
      .join(' ')
  }

  const getStatusText = (value: number) =>
    (
      {
        1: '等待处理',
        2: '处理成功',
        3: '处理失败',
        4: '退款成功',
        5: '申请退款'
      } as Record<number, string>
    )[value] || `状态 ${value}`

  const getStatusType = (value: number) =>
    (
      {
        1: 'warning',
        2: 'success',
        3: 'danger',
        4: 'info',
        5: 'warning'
      } as Record<number, 'danger' | 'info' | 'success' | 'warning'>
    )[value] || 'info'

  const loadSchools = async () => {
    schools.value = (await fetchLegacyYDSJSchools()) || []
  }

  const refreshPrice = async () => {
    if (!addForm.distance || Number(addForm.distance) <= 0) {
      estimatedPrice.value = null
      return
    }
    priceLoading.value = true
    try {
      const result = await fetchLegacyYDSJPrice(addForm.run_type, Number(addForm.distance))
      estimatedPrice.value = result?.price ?? null
    } catch {
      estimatedPrice.value = null
    } finally {
      priceLoading.value = false
    }
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyYDSJOrders({
        keyword: filters.keyword || undefined,
        limit: pagination.limit,
        page: pagination.page,
        searchType: filters.keyword ? filters.searchType : undefined,
        status: filters.status || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const resetFilters = () => {
    filters.keyword = ''
    filters.searchType = '1'
    filters.status = ''
    loadOrders(1)
  }

  const resetAddForm = () => {
    Object.assign(addForm, {
      distance: 2,
      end_time: '21:00',
      pass: '',
      remarks: '',
      run_type: 1,
      run_week: ['1', '2', '3', '4', '5', '6', '7'],
      school: '',
      start_time: '12:00',
      user: ''
    })
    estimatedPrice.value = null
    syncWeekState()
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
    refreshPrice()
  }

  const handleCreate = async () => {
    if (!addForm.user || !addForm.pass || !addForm.distance) {
      ElMessage.warning('请填写账号、密码和公里数')
      return
    }
    if (!addForm.start_time || !addForm.end_time) {
      ElMessage.warning('请设置跑步时间')
      return
    }
    if (!addForm.run_week.length) {
      ElMessage.warning('请选择跑步周期')
      return
    }

    const start = splitTime(addForm.start_time)
    const end = splitTime(addForm.end_time)

    addLoading.value = true
    try {
      await createLegacyYDSJOrder({
        distance: String(addForm.distance),
        end_hour: end.hour,
        end_minute: end.minute,
        pass: addForm.pass,
        remarks: addForm.remarks,
        run_type: addForm.run_type,
        run_week: addForm.run_week.join(','),
        school: addForm.school || '自动识别',
        start_hour: start.hour,
        start_minute: start.minute,
        user: addForm.user
      })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openRemarksDialog = (order: LegacyYDSJOrder) => {
    remarksForm.id = order.id
    remarksForm.user = order.user
    remarksForm.remarks = order.remarks || ''
    remarksVisible.value = true
  }

  const handleSaveRemarks = async () => {
    remarksLoading.value = true
    try {
      await editLegacyYDSJRemarks(remarksForm.id, remarksForm.remarks)
      ElMessage.success('备注已更新')
      remarksVisible.value = false
      loadOrders(pagination.page)
    } finally {
      remarksLoading.value = false
    }
  }

  const openDetailDialog = (order: LegacyYDSJOrder) => {
    detailRecord.value = order
    detailTab.value = 'info'
    detailVisible.value = true
  }

  const handleRefund = async (order: LegacyYDSJOrder) => {
    await ElMessageBox.confirm(`确认退款订单 #${order.id}（${order.user}）？`, '退款订单', { type: 'warning' })
    await refundLegacyYDSJOrder(order.id)
    ElMessage.success('退款成功')
    loadOrders(pagination.page)
  }

  const handleSync = async (order: LegacyYDSJOrder) => {
    await syncLegacyYDSJOrder(order.id)
    ElMessage.success('同步成功')
    loadOrders(pagination.page)
  }

  const handleToggleRun = async (order: LegacyYDSJOrder) => {
    await toggleLegacyYDSJRun(order.id)
    ElMessage.success(order.is_run === 1 ? '已暂停' : '已开启')
    loadOrders(pagination.page)
  }

  watch(
    () => [addForm.distance, addForm.run_type, addForm.school],
    () => {
      if (addVisible.value) {
        refreshPrice()
      }
    }
  )

  onMounted(async () => {
    syncWeekState()
    await loadSchools()
    await loadOrders(1)
  })
</script>
