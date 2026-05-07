<template>
  <div class="plugin-yongye-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="resetFilters"
    />

    <section class="art-card-sm overflow-hidden">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="reloadCurrentTab">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">{{ activeTab === 'orders' ? '订单列表' : '学生列表' }}</ElTag>
            <ElTag type="success" effect="plain">订单 {{ orders.length }} 条</ElTag>
            <ElTag type="warning" effect="plain">学生 {{ students.length }} 条</ElTag>
            <ElButton v-if="activeTab === 'orders'" type="primary" plain @click="openAddDialog">新增订单</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTabs v-model="activeTab" @tab-change="handleTabChange">
        <ElTabPane label="订单列表" name="orders">
          <ArtTable
            :loading="loading"
            :data="orders"
            :columns="orderColumns"
            :pagination="tablePagination"
            @pagination:current-change="loadOrders"
            @pagination:size-change="handleSizeChange"
          />
        </ElTabPane>

        <ElTabPane label="学生列表" name="students">
          <ArtTable :loading="loading" :data="students" :columns="studentColumns" :show-table-header="true" />
        </ElTabPane>
      </ElTabs>
    </section>

    <ElDialog v-model="addVisible" title="新增永夜运动订单" width="760px">
      <div class="space-y-4">
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">跑步类型</p>
          <ElSelect v-model="addForm.type" class="w-full">
            <ElOption v-for="item in runTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </ElSelect>
        </div>

        <div>
          <p class="mb-2 text-sm font-medium text-g-800">学校</p>
          <ElSelect v-model="addForm.school" class="w-full" filterable :loading="schoolLoading" placeholder="请选择学校">
            <ElOption label="自动识别" value="自动识别" />
            <ElOption v-for="item in schools" :key="item.name" :label="item.name" :value="item.name" />
          </ElSelect>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">学号</p>
            <ElInput v-model="addForm.user" placeholder="请输入学号" />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">密码</p>
            <ElInput v-model="addForm.pass" show-password placeholder="请输入密码" />
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">公里数</p>
            <ElInputNumber v-model="addForm.zkm" class="w-full" :min="0.1" :max="50" :step="0.5" />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">轮询模式</p>
            <ElSelect v-model="addForm.isPolling" class="w-full">
              <ElOption :value="0" label="关闭" />
              <ElOption :value="1" label="开启" />
            </ElSelect>
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
            <ElCheckboxGroup v-model="selectedWeeks" @change="syncWeeks">
              <ElCheckbox v-for="item in weekOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>
        </div>

        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-600">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <span>{{ getRunTypeText(addForm.type) }}</span>
            <span>{{ Number(addForm.zkm || 0).toFixed(1) }} km</span>
            <span>{{ addForm.start_time }} - {{ addForm.end_time }}</span>
          </div>
          <p class="mt-2 text-xs text-g-500">学校可留给后端自动识别，确认账号、时间段和周期后再提交。</p>
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
    createLegacyYongyeOrder,
    fetchLegacyYongyeOrders,
    fetchLegacyYongyeSchools,
    fetchLegacyYongyeStudents,
    refundLegacyYongyeOrder,
    refundLegacyYongyeStudent,
    toggleLegacyYongyePolling,
    type LegacyYongyeOrder,
    type LegacyYongyeStudent
  } from '@/api/legacy/plugin-yongye'

  defineOptions({ name: 'PluginYongyePage' })

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
    { label: '正常跑(课外)', value: 0 },
    { label: '晨跑', value: 1 }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '未提交', value: '0' },
    { label: '已提交', value: '1' },
    { label: '请求失败', value: '2' },
    { label: '已关闭', value: '3' },
    { label: '轮询中', value: '5' }
  ]

  const activeTab = ref('orders')
  const loading = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const schoolLoading = ref(false)

  const schools = ref<any[]>([])
  const orders = ref<LegacyYongyeOrder[]>([])
  const students = ref<LegacyYongyeStudent[]>([])
  const selectedWeeks = ref<string[]>(['1', '2', '3', '4', '5'])

  const pagination = reactive({
    limit: 20,
    page: 1,
    total: 0
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.limit,
    total: pagination.total
  }))

  const filters = reactive({
    keyword: '',
    status: ''
  })

  const searchItems = computed(() => [
    {
      label: '订单状态',
      key: 'status',
      type: 'select',
      hidden: activeTab.value !== 'orders',
      props: {
        clearable: true,
        placeholder: '订单状态',
        options: statusOptions
      }
    },
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: activeTab.value === 'orders' ? '搜索账号或订单ID' : '搜索学生账号'
      }
    }
  ])

  const addForm = reactive({
    end_time: '21:00',
    isPolling: 0,
    pass: '',
    school: '自动识别',
    start_time: '09:00',
    type: 0,
    user: '',
    weeks: '12345',
    zkm: 2
  })

  const splitTime = (value: string) => {
    const [hour = '00', minute = '00'] = String(value || '').split(':')
    return { hour: Number(hour), minute: Number(minute) }
  }

  const formatTimeRange = (startHour: number, startMinute: number, endHour: number, endMinute: number) =>
    `${String(startHour).padStart(2, '0')}:${String(startMinute).padStart(2, '0')} - ${String(endHour).padStart(2, '0')}:${String(endMinute).padStart(2, '0')}`

  const getRunTypeText = (value: number) => runTypeOptions.find((item) => item.value === value)?.label || `类型 ${value}`

  const getWeeksText = (value: string) => {
    if (!value) return '未设置周期'
    const map: Record<string, string> = { '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六', '7': '日' }
    return value
      .split('')
      .filter(Boolean)
      .map((item) => `周${map[item] || item}`)
      .join(' ')
  }

  const getStatusText = (value: number) =>
    (
      {
        0: '未提交',
        1: '已提交',
        2: '请求失败',
        3: '已关闭',
        5: '轮询中'
      } as Record<number, string>
    )[value] || `状态 ${value}`

  const getStatusType = (value: number) =>
    (
      {
        0: 'info',
        1: 'warning',
        2: 'danger',
        3: 'warning',
        5: 'success'
      } as Record<number, 'danger' | 'info' | 'success' | 'warning'>
    )[value] || 'info'

  const { columns: orderColumns } = useTableColumns<LegacyYongyeOrder>(() => [
    {
      prop: 'user',
      label: '账号信息',
      minWidth: 180,
      formatter: (row: LegacyYongyeOrder) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.user || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.pass || '无密码')
        ])
    },
    {
      prop: 'school',
      label: '学校',
      minWidth: 150,
      formatter: (row: LegacyYongyeOrder) => h('span', { class: 'text-sm text-g-700' }, row.school || '自动识别')
    },
    {
      prop: 'type',
      label: '跑步配置',
      minWidth: 220,
      formatter: (row: LegacyYongyeOrder) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, getRunTypeText(row.type)),
          h('p', { class: 'text-xs text-g-500' }, `${row.zkm} km，${getWeeksText(row.weeks)}`),
          h('p', { class: 'text-xs text-g-500' }, formatTimeRange(row.ks_h, row.ks_m, row.js_h, row.js_m))
        ])
    },
    {
      prop: 'yfees',
      label: '预扣',
      width: 100,
      align: 'right',
      formatter: (row: LegacyYongyeOrder) => h('span', { class: 'font-semibold text-[var(--el-color-danger)]' }, `¥${Number(row.yfees || 0).toFixed(2)}`)
    },
    {
      prop: 'dockstatus',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row: LegacyYongyeOrder) => h(ElTag, { type: getStatusType(row.dockstatus), effect: 'plain' }, () => getStatusText(row.dockstatus))
    },
    {
      prop: 'pol',
      label: '轮询',
      width: 100,
      align: 'center',
      formatter: (row: LegacyYongyeOrder) => h(ElTag, { type: row.pol === 1 ? 'success' : 'info', effect: 'plain' }, () => (row.pol === 1 ? '开启' : '关闭'))
    },
    {
      prop: 'tktext',
      label: '日志',
      minWidth: 180,
      formatter: (row: LegacyYongyeOrder) => h('span', { class: 'line-clamp-2 text-sm text-g-500' }, row.tktext || '暂无日志')
    },
    {
      prop: 'addtime',
      label: '下单时间',
      minWidth: 160
    },
    {
      prop: 'operation',
      label: '操作',
      width: 220,
      fixed: 'right' as const,
      formatter: (row: LegacyYongyeOrder) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { size: 'small', onClick: () => handleTogglePolling(row) }, () => (row.pol === 1 ? '关轮询' : '开轮询')),
          h(ElButton, { size: 'small', type: 'danger', plain: true, disabled: row.dockstatus === 3, onClick: () => handleRefundOrder(row) }, () => '退款')
        ])
    }
  ])

  const { columns: studentColumns } = useTableColumns<LegacyYongyeStudent>(() => [
    {
      prop: 'user',
      label: '账号信息',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.user || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.pass || '无密码')
        ])
    },
    {
      prop: 'type',
      label: '跑步配置',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, getRunTypeText(row.type)),
          h('p', { class: 'text-xs text-g-500' }, `${row.zkm} km，${getWeeksText(row.weeks)}`)
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 110,
      align: 'center',
      formatter: (row) => h(ElTag, { type: getStudentStatusType(row.status), effect: 'plain' }, () => getStudentStatusText(row.status))
    },
    {
      prop: 'tdkm',
      label: '退单信息',
      minWidth: 160,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-700' }, `退单公里 ${Number(row.tdkm || 0).toFixed(2)}`),
          h('p', { class: 'text-xs text-g-500' }, `退单金额 ¥${Number(row.tdmoney || 0).toFixed(2)}`)
        ])
    },
    {
      prop: 'stulog',
      label: '日志',
      minWidth: 180,
      formatter: (row) => h('span', { class: 'line-clamp-2 text-sm text-g-500' }, row.stulog || '暂无日志')
    },
    {
      prop: 'last_time',
      label: '最后更新',
      minWidth: 160
    },
    {
      prop: 'operation',
      label: '操作',
      width: 120,
      fixed: 'right',
      formatter: (row) =>
        h(ElButton, { size: 'small', type: 'danger', plain: true, disabled: row.status === 3, onClick: () => handleRefundStudent(row) }, () => '退单')
    }
  ])

  const getStudentStatusText = (value: number) =>
    (
      {
        0: '正常',
        1: '暂停',
        2: '完成',
        3: '退单'
      } as Record<number, string>
    )[value] || `状态 ${value}`

  const getStudentStatusType = (value: number) =>
    (
      {
        0: 'success',
        1: 'warning',
        2: 'info',
        3: 'danger'
      } as Record<number, 'danger' | 'info' | 'success' | 'warning'>
    )[value] || 'info'

  const syncWeeks = () => {
    selectedWeeks.value = [...selectedWeeks.value].sort()
    addForm.weeks = selectedWeeks.value.join('')
  }

  const loadSchools = async () => {
    schoolLoading.value = true
    try {
      const result = await fetchLegacyYongyeSchools()
      schools.value = Array.isArray(result?.data) ? result.data : Array.isArray(result) ? result : []
    } finally {
      schoolLoading.value = false
    }
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyYongyeOrders({
        keyword: filters.keyword || undefined,
        limit: pagination.limit,
        page: pagination.page,
        status: filters.status || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const loadStudents = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyYongyeStudents({
        keyword: filters.keyword || undefined
      })
      students.value = Array.isArray(result) ? result : []
    } finally {
      loading.value = false
    }
  }

  const handleSearch = () => {
    if (activeTab.value === 'orders') {
      loadOrders(1)
    } else {
      loadStudents()
    }
  }

  const reloadCurrentTab = () => {
    handleSearch()
  }

  const handleSizeChange = (size: number) => {
    pagination.limit = size
    loadOrders(1)
  }

  const resetFilters = () => {
    filters.keyword = ''
    filters.status = ''
    handleSearch()
  }

  const resetAddForm = () => {
    Object.assign(addForm, {
      end_time: '21:00',
      isPolling: 0,
      pass: '',
      school: '自动识别',
      start_time: '09:00',
      type: 0,
      user: '',
      weeks: '12345',
      zkm: 2
    })
    selectedWeeks.value = ['1', '2', '3', '4', '5']
    syncWeeks()
  }

  const openAddDialog = () => {
    resetAddForm()
    addVisible.value = true
    if (!schools.value.length) {
      loadSchools()
    }
  }

  const handleCreate = async () => {
    if (!addForm.user || !addForm.pass || !addForm.zkm) {
      ElMessage.warning('请填写账号、密码和公里数')
      return
    }

    const start = splitTime(addForm.start_time)
    const end = splitTime(addForm.end_time)
    syncWeeks()

    addLoading.value = true
    try {
      await createLegacyYongyeOrder({
        isPolling: addForm.isPolling,
        js_h: end.hour,
        js_m: end.minute,
        ks_h: start.hour,
        ks_m: start.minute,
        pass: addForm.pass,
        school: addForm.school,
        type: addForm.type,
        user: addForm.user,
        weeks: addForm.weeks,
        zkm: addForm.zkm
      })
      ElMessage.success('下单成功')
      addVisible.value = false
      if (activeTab.value !== 'orders') {
        activeTab.value = 'orders'
      }
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const handleRefundOrder = async (order: LegacyYongyeOrder) => {
    await ElMessageBox.confirm(`确认退款订单 #${order.id}（${order.user}）？`, '退款订单', { type: 'warning' })
    await refundLegacyYongyeOrder(order.id)
    ElMessage.success('退款成功')
    loadOrders(pagination.page)
  }

  const handleTogglePolling = async (order: LegacyYongyeOrder) => {
    await toggleLegacyYongyePolling(order.id)
    ElMessage.success(order.pol === 1 ? '已关闭轮询' : '已开启轮询')
    loadOrders(pagination.page)
  }

  const handleRefundStudent = async (student: LegacyYongyeStudent) => {
    await ElMessageBox.confirm(`确认退单学生 ${student.user}？`, '退单学生', { type: 'warning' })
    await refundLegacyYongyeStudent(student.user, student.type)
    ElMessage.success('退单请求已发送')
    loadStudents()
  }

  const handleTabChange = (name: string | number) => {
    activeTab.value = String(name)
    if (activeTab.value === 'orders') {
      loadOrders(1)
    } else {
      loadStudents()
    }
  }

  onMounted(() => {
    syncWeeks()
    loadOrders(1)
  })
</script>
