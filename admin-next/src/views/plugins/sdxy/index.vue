<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <ElAlert
        type="warning"
        :closable="false"
        show-icon
        title="下单前请确认账号密码正确，跑步期间不要登录账号。短信验证码登录仅用于查询学生信息。"
      />

      <div class="mt-4 grid gap-4 xl:grid-cols-[180px_140px_1fr_auto]">
        <ElSelect v-model="filters.status" clearable placeholder="订单状态">
          <ElOption v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </ElSelect>
        <ElSelect v-model="filters.searchType" placeholder="搜索项">
          <ElOption v-for="item in searchFieldOptions" :key="item.value" :label="item.label" :value="item.value" />
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
            <ElTag type="success" effect="plain">进行中 {{ processingCount }}</ElTag>
            <ElTag type="warning" effect="plain">单次 ¥{{ Number(pricePerTask || 0).toFixed(2) }}</ElTag>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="orders" size="large">
        <ElTableColumn label="订单" width="110">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-semibold text-g-900">#{{ row.id }}</p>
              <p class="text-xs text-g-500">{{ row.agg_order_id || '未下发源台ID' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="账号信息" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-semibold text-g-900">{{ row.user }}</p>
              <p class="text-xs text-g-500">{{ row.pass || '无密码' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="学校 / 跑区" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm text-g-800">{{ row.school || '未识别学校' }}</p>
              <p class="text-xs text-g-500">{{ row.run_rule || '未识别规则' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="任务配置" min-width="160">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm text-g-800">{{ getRunTypeLabel(row.run_type) }}</p>
              <p class="text-xs text-g-500">{{ row.num }} 次，{{ row.distance }} km/次</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="状态" width="140" align="center">
          <template #default="{ row }">
            <div class="flex flex-col items-center gap-2">
              <ElTag :type="getStatusType(row.status)" effect="plain">{{ getStatusLabel(row.status) }}</ElTag>
              <ElTag v-if="row.pause === 1" type="warning" effect="plain">已暂停</ElTag>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="费用" width="110" align="right">
          <template #default="{ row }">
            <span class="font-semibold text-[var(--el-color-danger)]">¥{{ Number(row.fees || 0).toFixed(2) }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn prop="created_at" label="创建时间" min-width="160" />

        <ElTableColumn label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <ElButton size="small" @click="openLogsDialog(row)">日志</ElButton>
              <ElButton size="small" :disabled="!row.agg_order_id" @click="handleTogglePause(row)">
                {{ row.pause === 1 ? '恢复' : '暂停' }}
              </ElButton>
              <ElButton size="small" :disabled="!row.agg_order_id" @click="handleDelay(row)">延时</ElButton>
              <ElButton
                size="small"
                type="danger"
                plain
                :disabled="!row.agg_order_id || row.status === '5'"
                @click="handleRefund(row)"
              >
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

    <ElDialog v-model="addVisible" title="新增闪电运动订单" width="780px">
      <div class="space-y-5">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="grid gap-4 md:grid-cols-[160px_1fr]">
              <div>
                <p class="mb-2 text-sm font-medium text-g-800">登录方式</p>
                <ElRadioGroup v-model="loginMode">
                  <ElRadioButton label="password">密码登录</ElRadioButton>
                  <ElRadioButton label="code">验证码登录</ElRadioButton>
                </ElRadioGroup>
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">手机号</p>
                  <ElInput v-model="addForm.phone" placeholder="请输入手机号" />
                </div>
                <div v-if="loginMode === 'password'">
                  <p class="mb-2 text-sm font-medium text-g-800">密码</p>
                  <ElInput v-model="addForm.password" show-password placeholder="密码可选，视学校而定" />
                </div>
                <div v-else>
                  <p class="mb-2 text-sm font-medium text-g-800">验证码</p>
                  <div class="flex gap-3">
                    <ElInput v-model="addForm.code" placeholder="请输入验证码" />
                    <ElButton :disabled="codeCountdown > 0" @click="handleSendCode">
                      {{ codeCountdown > 0 ? `${codeCountdown}s` : '发送验证码' }}
                    </ElButton>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-4 flex justify-end">
              <ElButton plain :loading="queryLoading" @click="handleQueryUserInfo">查询学生信息</ElButton>
            </div>
          </div>

          <ElAlert
            v-if="queryFeedback.message"
            :title="queryFeedback.message"
            :type="queryFeedback.type"
            :closable="false"
            show-icon
          />

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">跑区</p>
              <ElSelect
                v-if="zones.length"
                v-model="addForm.zone_id"
                class="w-full"
                filterable
                placeholder="请选择跑区"
                @change="handleZoneChange"
              >
                <ElOption v-for="item in zones" :key="item.id" :label="item.name" :value="item.id" />
              </ElSelect>
              <ElInput v-else v-model="addForm.zone_name" placeholder="手动输入跑区名称" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">运行规则</p>
              <ElSelect
                v-if="runRules.length"
                v-model="addForm.run_rule_id"
                class="w-full"
                filterable
                placeholder="请选择运行规则"
              >
                <ElOption v-for="item in runRules" :key="item.id" :label="item.label" :value="item.id" />
              </ElSelect>
              <ElInput v-else v-model="addForm.run_rule_id" placeholder="手动填写规则ID" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">跑步类型</p>
              <ElRadioGroup v-model="addForm.run_type">
                <ElRadioButton label="1">有效跑</ElRadioButton>
                <ElRadioButton label="2">自由跑</ElRadioButton>
              </ElRadioGroup>
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">每次公里数</p>
              <ElInputNumber v-model="addForm.dis" class="w-full" :min="0.5" :max="100" :step="0.5" :precision="1" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">开始日期</p>
              <ElDatePicker
                v-model="addForm.start_date"
                class="w-full"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="请选择开始日期"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">任务次数</p>
              <ElInputNumber v-model="addForm.day" class="w-full" :min="1" :max="365" />
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
              <span class="text-g-500">学生信息</span>
              <span class="font-medium text-g-900">{{ studentMeta.name || '未查询' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">学校 / 跑区</span>
              <span class="truncate text-right font-medium text-g-900">{{ studentMeta.school || addForm.zone_name || '未识别' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">学生ID</span>
              <span class="font-medium text-g-900">{{ addForm.student_id || '-' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">任务配置</span>
              <span class="text-right font-medium text-g-900">{{ taskCount }} 次，{{ Number(addForm.dis || 0).toFixed(1) }} km/次</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">单次价格</span>
              <span class="font-medium text-g-900">¥{{ Number(pricePerTask || 0).toFixed(2) }}</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="text-g-500">预估价格</span>
              <span class="font-semibold text-[var(--el-color-danger)]">¥{{ estimatedPrice }}</span>
            </div>
          </div>
          <p class="mt-2 text-xs text-g-500">若未查到跑区或规则，需要手动补齐字段，系统会按开始日期、时间段和周期生成任务。</p>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addLoading" @click="handleCreate">确认下单</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="logsVisible" title="任务日志" width="760px">
      <ElTable v-loading="logsLoading" :data="logs" size="large">
        <ElTableColumn prop="date" label="日期" min-width="120" />
        <ElTableColumn prop="start_time" label="开始时间" min-width="100" />
        <ElTableColumn prop="status" label="状态" min-width="100" />
        <ElTableColumn prop="result" label="结果" min-width="280" />
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
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createLegacySDXYOrder,
    delayLegacySDXYTask,
    fetchLegacySDXYOrders,
    fetchLegacySDXYPrice,
    fetchLegacySDXYRunLogs,
    pauseLegacySDXYOrder,
    queryLegacySDXYUserInfo,
    queryLegacySDXYUserInfoByCode,
    refundLegacySDXYOrder,
    sendLegacySDXYCode,
    type LegacySDXYOrder
  } from '@/api/legacy/plugin-sdxy'

  defineOptions({ name: 'PluginSDXYPage' })

  interface SelectItem {
    id: string
    label: string
    name: string
    [key: string]: any
  }

  const weekOptions = [
    { label: '周一', value: 1 },
    { label: '周二', value: 2 },
    { label: '周三', value: 3 },
    { label: '周四', value: 4 },
    { label: '周五', value: 5 },
    { label: '周六', value: 6 },
    { label: '周日', value: 7 }
  ]

  const searchFieldOptions = [
    { label: '订单ID', value: '1' },
    { label: '下单账号', value: '2' },
    { label: '下单密码', value: '3' },
    { label: '用户UID', value: '4' }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '等待处理', value: '1' },
    { label: '进行中', value: '2' },
    { label: '已完成', value: '3' },
    { label: '失败', value: '4' },
    { label: '已退款', value: '5' }
  ]

  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const queryLoading = ref(false)
  const logsVisible = ref(false)
  const logsLoading = ref(false)

  const loginMode = ref<'code' | 'password'>('password')
  const codeCountdown = ref(0)
  let codeTimer: ReturnType<typeof setInterval> | null = null

  const pricePerTask = ref(0)
  const orders = ref<LegacySDXYOrder[]>([])
  const zones = ref<SelectItem[]>([])
  const runRules = ref<SelectItem[]>([])
  const logs = ref<any[]>([])
  const studentInfo = ref<Record<string, any> | null>(null)

  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const logsPagination = reactive({
    orderId: '',
    page: 1,
    pageSize: 20,
    total: 0
  })

  const filters = reactive({
    keyword: '',
    searchType: '1',
    status: ''
  })

  const addForm = reactive({
    code: '',
    day: 7,
    dis: 2,
    end_time: '08:00',
    password: '',
    phone: '',
    run_rule_id: '',
    run_type: '1',
    run_week: [1, 2, 3, 4, 5] as number[],
    start_date: '',
    start_time: '06:00',
    student_id: '',
    zone_id: '',
    zone_name: ''
  })

  const queryFeedback = reactive({
    message: '',
    type: 'success' as 'error' | 'success' | 'warning'
  })

  const checkAllWeek = ref(false)
  const weekIndeterminate = ref(false)

  const studentMeta = computed(() => {
    const schoolValue = studentInfo.value?.school
    return {
      name: studentInfo.value?.name || studentInfo.value?.student_name || '',
      school:
        typeof schoolValue === 'string'
          ? schoolValue
          : schoolValue?.name || schoolValue?.school_name || studentInfo.value?.zone_name || ''
    }
  })

  const processingCount = computed(() => orders.value.filter((item) => item.status === '2').length)
  const taskList = computed(() => generateTaskList())
  const taskCount = computed(() => taskList.value.length)
  const estimatedPrice = computed(() => (Number(pricePerTask.value || 0) * taskCount.value).toFixed(2))

  const syncWeekState = () => {
    const count = addForm.run_week.length
    checkAllWeek.value = count === weekOptions.length
    weekIndeterminate.value = count > 0 && count < weekOptions.length
  }

  const toggleAllWeek = (value: boolean | string | number) => {
    addForm.run_week = value ? weekOptions.map((item) => item.value) : []
    syncWeekState()
  }

  const pad = (value: number) => String(value).padStart(2, '0')

  const formatDate = (date: Date) =>
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`

  const generateTaskList = () => {
    if (!addForm.start_date || !addForm.start_time || !addForm.end_time || addForm.day <= 0) {
      return []
    }

    const tasks: Array<Record<string, string>> = []
    const startDate = new Date(`${addForm.start_date}T00:00:00`)

    for (let index = 0; index < addForm.day * 3; index += 1) {
      const current = new Date(startDate)
      current.setDate(startDate.getDate() + index)
      const week = current.getDay() === 0 ? 7 : current.getDay()
      if (!addForm.run_week.includes(week)) {
        continue
      }
      tasks.push({
        date: formatDate(current),
        end_time: addForm.end_time,
        start_time: addForm.start_time
      })
      if (tasks.length >= addForm.day) {
        break
      }
    }

    return tasks
  }

  const normalizeSelectItems = (items: any[], idKeys: string[], labelKeys: string[]) =>
    items
      .map((item) => {
        const id = idKeys.map((key) => item?.[key]).find((value) => value !== undefined && value !== null && `${value}` !== '')
        const label = labelKeys
          .map((key) => item?.[key])
          .find((value) => value !== undefined && value !== null && `${value}` !== '')
        return {
          ...item,
          id: String(id || ''),
          label: String(label || ''),
          name: String(label || '')
        }
      })
      .filter((item) => item.id || item.label)

  const parseQueryPayload = (result: any) => {
    const payload = result?.data || result
    const student = payload?.student || payload
    const zonesValue = student?.zones || student?.zone_list || payload?.zones || payload?.zone_list || []
    const rulesValue = student?.run_rules || student?.run_rule_list || payload?.run_rules || payload?.run_rule_list || []
    const queryMsg = payload?.msg || result?.msg || '查询成功'

    studentInfo.value = student && typeof student === 'object' ? student : null
    addForm.student_id = String(student?.id || student?.student_id || '')
    zones.value = normalizeSelectItems(Array.isArray(zonesValue) ? zonesValue : [], ['id', 'zone_id', 'school_id'], ['name', 'zone_name', 'school_name'])
    runRules.value = normalizeSelectItems(
      Array.isArray(rulesValue) ? rulesValue : student?.run_rule ? [student.run_rule] : [],
      ['id', 'run_rule_id'],
      ['label', 'name']
    )

    if (!addForm.zone_name && studentMeta.value.school) {
      addForm.zone_name = studentMeta.value.school
    }
    if (zones.value.length === 1) {
      addForm.zone_id = zones.value[0]!.id
      addForm.zone_name = zones.value[0]!.name
    }
    if (runRules.value.length === 1) {
      addForm.run_rule_id = runRules.value[0]!.id
    }

    queryFeedback.type = 'success'
    queryFeedback.message = queryMsg
  }

  const normalizeLogs = (result: any) => {
    if (Array.isArray(result?.list)) return result.list
    if (Array.isArray(result?.data?.list)) return result.data.list
    if (Array.isArray(result?.data)) return result.data
    if (Array.isArray(result)) return result
    return []
  }

  const loadPrice = async () => {
    try {
      const result = await fetchLegacySDXYPrice()
      pricePerTask.value = Number(result?.price || 0)
    } catch {
      pricePerTask.value = 0
    }
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacySDXYOrders({
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

  const loadLogs = async (page = logsPagination.page) => {
    if (!logsPagination.orderId) return
    logsLoading.value = true
    logsPagination.page = page
    try {
      const result = await fetchLegacySDXYRunLogs(logsPagination.orderId, logsPagination.page, logsPagination.pageSize)
      logs.value = normalizeLogs(result)
      logsPagination.total = Number(result?.total || result?.data?.total || logs.value.length)
    } finally {
      logsLoading.value = false
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
      code: '',
      day: 7,
      dis: 2,
      end_time: '08:00',
      password: '',
      phone: '',
      run_rule_id: '',
      run_type: '1',
      run_week: [1, 2, 3, 4, 5],
      start_date: formatDate(new Date()),
      start_time: '06:00',
      student_id: '',
      zone_id: '',
      zone_name: ''
    })
    queryFeedback.message = ''
    queryFeedback.type = 'success'
    studentInfo.value = null
    zones.value = []
    runRules.value = []
    loginMode.value = 'password'
    syncWeekState()
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
  }

  const handleZoneChange = (value: string) => {
    const current = zones.value.find((item) => item.id === value)
    if (current) {
      addForm.zone_name = current.name
    }
  }

  const startCountdown = () => {
    codeCountdown.value = 60
    if (codeTimer) {
      clearInterval(codeTimer)
    }
    codeTimer = setInterval(() => {
      codeCountdown.value -= 1
      if (codeCountdown.value <= 0 && codeTimer) {
        clearInterval(codeTimer)
        codeTimer = null
      }
    }, 1000)
  }

  const handleSendCode = async () => {
    if (!addForm.phone) {
      ElMessage.warning('请输入手机号')
      return
    }
    await sendLegacySDXYCode({ phone: addForm.phone })
    ElMessage.success('验证码已发送')
    startCountdown()
  }

  const handleQueryUserInfo = async () => {
    if (!addForm.phone) {
      ElMessage.warning('请输入手机号')
      return
    }
    if (loginMode.value === 'code' && !addForm.code) {
      ElMessage.warning('请输入验证码')
      return
    }

    queryLoading.value = true
    queryFeedback.message = ''
    try {
      const result =
        loginMode.value === 'password'
          ? await queryLegacySDXYUserInfo({ password: addForm.password, phone: addForm.phone })
          : await queryLegacySDXYUserInfoByCode({ code: addForm.code, phone: addForm.phone })
      parseQueryPayload(result)
      ElMessage.success('学生信息查询成功')
    } catch (error: any) {
      queryFeedback.type = 'error'
      queryFeedback.message = error?.message || '查询失败，请检查账号信息'
    } finally {
      queryLoading.value = false
    }
  }

  const handleCreate = async () => {
    if (!addForm.phone) {
      ElMessage.warning('请输入手机号')
      return
    }
    if (!addForm.student_id) {
      ElMessage.warning('请先查询学生信息')
      return
    }
    if (!addForm.zone_id && !addForm.zone_name) {
      ElMessage.warning('请选择或填写跑区')
      return
    }
    if (!addForm.run_rule_id) {
      ElMessage.warning('请选择运行规则')
      return
    }
    if (!taskCount.value) {
      ElMessage.warning('请检查开始日期和跑步周期')
      return
    }

    addLoading.value = true
    try {
      await createLegacySDXYOrder({
        day: addForm.day,
        dis: addForm.dis,
        password: addForm.password,
        phone: addForm.phone,
        run_rule_id: addForm.run_rule_id,
        run_type: addForm.run_type,
        student_id: addForm.student_id,
        task_list: taskList.value,
        zone_id: addForm.zone_id,
        zone_name: addForm.zone_name
      })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openLogsDialog = (order: LegacySDXYOrder) => {
    logsPagination.orderId = order.sdxy_order_id
    logsPagination.page = 1
    logsVisible.value = true
    loadLogs(1)
  }

  const handleRefund = async (order: LegacySDXYOrder) => {
    await ElMessageBox.confirm(`确认退款订单 #${order.id}（${order.user}）？`, '退款订单', { type: 'warning' })
    await refundLegacySDXYOrder(order.agg_order_id)
    ElMessage.success('退款成功')
    loadOrders(pagination.page)
  }

  const handleTogglePause = async (order: LegacySDXYOrder) => {
    await pauseLegacySDXYOrder(order.agg_order_id, order.pause === 1 ? 0 : 1)
    ElMessage.success(order.pause === 1 ? '已恢复' : '已暂停')
    loadOrders(pagination.page)
  }

  const handleDelay = async (order: LegacySDXYOrder) => {
    await ElMessageBox.confirm(`确认延迟订单 #${order.id} 的下一条任务？`, '延迟任务', { type: 'warning' })
    await delayLegacySDXYTask(order.agg_order_id)
    ElMessage.success('延时成功')
  }

  const getRunTypeLabel = (value: string) => (value === '2' ? '自由跑' : '有效跑')

  const getStatusLabel = (value: string) =>
    (
      {
        '1': '等待处理',
        '2': '进行中',
        '3': '已完成',
        '4': '失败',
        '5': '已退款'
      } as Record<string, string>
    )[value] || value

  const getStatusType = (value: string) =>
    (
      {
        '1': 'warning',
        '2': 'success',
        '3': 'success',
        '4': 'danger',
        '5': 'info'
      } as Record<string, 'danger' | 'info' | 'success' | 'warning'>
    )[value] || 'info'

  onMounted(async () => {
    resetAddForm()
    await loadPrice()
    await loadOrders(1)
  })

  onBeforeUnmount(() => {
    if (codeTimer) {
      clearInterval(codeTimer)
    }
  })
</script>
