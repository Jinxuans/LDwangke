<template>
  <div class="plugin-jiguang-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      class="jiguang-search-bar"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="resetFilters"
    />

    <ElCard
      class="art-table-card jiguang-table-card"
      shadow="never"
      :body-style="{ paddingTop: '14px' }"
    >
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading || syncLoading"
        layout="refresh,size,fullscreen,columns,settings"
        fullClass="art-table-card"
        @refresh="loadOrders(pagination.page)"
      >
        <template #left>
          <div class="jiguang-table-toolbar">
            <h4 class="jiguang-table-title">极光跑步订单</h4>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
            <ElButton plain @click="loadProducts">刷新价格</ElButton>
            <ElButton v-if="isAdmin" plain @click="openConfigDialog">配置接入</ElButton>
            <ElButton v-if="isAdmin" plain :loading="syncLoading" @click="handleSyncOrders"
              >同步订单</ElButton
            >
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="orders"
        :columns="columns"
        :pagination="tablePagination"
        empty-height="360px"
        :show-table-header="true"
        @pagination:current-change="loadOrders"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog
      v-model="addVisible"
      class="jiguang-order-dialog"
      title="新增极光跑步订单"
      width="860px"
      destroy-on-close
    >
      <div class="jiguang-order-form">
        <section class="jiguang-form-section">
          <div class="jiguang-section-title">
            <span>学生信息</span>
            <small>选择商品和学校后填写学生资料</small>
          </div>
          <div class="jiguang-form-grid md:grid-cols-2">
            <div class="jiguang-form-field">
              <label>商品</label>
              <ElSelect v-model="orderForm.product_id" class="w-full">
                <ElOption
                  v-for="item in products"
                  :key="item.product_id"
                  :label="`${item.name} · ${money(item.price)}/km`"
                  :value="item.product_id"
                />
              </ElSelect>
            </div>
            <div class="jiguang-form-field">
              <label>学校</label>
              <ElSelect
                v-model="orderForm.school_name"
                class="w-full"
                filterable
                remote
                reserve-keyword
                :loading="schoolLoading"
                :remote-method="searchSchools"
                placeholder="输入学校关键词搜索"
              >
                <ElOption
                  v-for="item in schools"
                  :key="schoolKey(item)"
                  :label="schoolLabel(item)"
                  :value="schoolLabel(item)"
                />
              </ElSelect>
            </div>
          </div>
          <div class="jiguang-form-grid md:grid-cols-2">
            <div class="jiguang-form-field">
              <label>学生姓名</label>
              <ElInput v-model="orderForm.student_name" placeholder="请输入学生姓名" />
            </div>
            <div class="jiguang-form-field">
              <label>学生学号</label>
              <ElInput v-model="orderForm.student_account" placeholder="请输入学生学号" />
            </div>
          </div>
        </section>

        <section class="jiguang-form-section">
          <div class="jiguang-section-title">
            <span>跑步参数</span>
            <small>按次数和每次公里数计算扣费</small>
          </div>
          <div class="jiguang-form-grid md:grid-cols-2">
            <div class="jiguang-form-field">
              <label>次数</label>
              <ElInputNumber
                v-model="orderForm.times"
                class="w-full"
                :min="1"
                :max="9999"
                :step="1"
              />
            </div>
            <div class="jiguang-form-field">
              <label>每次公里数</label>
              <ElSelect v-model="orderForm.km_per_day" class="w-full">
                <ElOption
                  v-for="item in kmOptions"
                  :key="item"
                  :label="`${item} km`"
                  :value="item"
                />
              </ElSelect>
            </div>
          </div>
        </section>

        <section class="jiguang-form-section">
          <div class="jiguang-section-title">
            <span>备注与确认</span>
            <small>提交前核对商品、次数和金额</small>
          </div>
          <div class="jiguang-form-field">
            <label>备注</label>
            <ElInput
              v-model="orderForm.customer_message"
              type="textarea"
              :rows="3"
              maxlength="500"
              show-word-limit
              placeholder="选填，下单后不可修改"
            />
          </div>

          <div class="jiguang-order-summary">
            <div>
              <span>商品</span>
              <strong>{{ selectedProduct?.name || '-' }}</strong>
            </div>
            <div>
              <span>次数</span>
              <strong>{{ Number(orderForm.times || 0) }} 次</strong>
            </div>
            <div>
              <span>公里</span>
              <strong>{{ Number(orderForm.km_per_day || 0) }} km/次</strong>
            </div>
            <div>
              <span>预估金额</span>
              <strong class="text-[var(--el-color-danger)]">{{ money(estimatedAmount) }}</strong>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="createLoading" @click="handleCreateOrder"
            >确认下单</ElButton
          >
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="configVisible" title="极光跑步配置" width="980px" destroy-on-close>
      <div class="space-y-5">
        <div class="flex flex-wrap justify-end gap-3">
          <ElButton plain :loading="configLoading" @click="loadConfig">重新读取</ElButton>
          <ElButton type="primary" :loading="saveConfigLoading" @click="handleSaveConfig">
            保存配置
          </ElButton>
        </div>

        <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">上游接入</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                选择官方源台、29系统或同系统开放接口，保存后用于学校、下单和订单同步。
              </p>
            </div>

            <div class="mt-5 grid gap-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">对接协议</label>
                <ElSelect v-model="configForm.upstream_protocol" class="w-full">
                  <ElOption label="官方源台" value="source" />
                  <ElOption label="29系统" value="compat29" />
                  <ElOption label="同系统开放接口" value="same_system" />
                </ElSelect>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">上游地址</label>
                <ElInput v-model="configForm.upstream_url" placeholder="例如 https://example.com" />
              </div>
              <div v-if="configForm.upstream_protocol === 'source'">
                <label class="mb-2 block text-sm font-medium text-g-800">官方 API Key</label>
                <ElInput v-model="configForm.api_key" show-password placeholder="源台模式填写" />
              </div>
              <div v-else class="grid gap-4 md:grid-cols-[180px_1fr]">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">上游 UID</label>
                  <ElInputNumber
                    v-model="configForm.upstream_uid"
                    class="w-full"
                    :min="0"
                    :precision="0"
                  />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">上游密钥</label>
                  <ElInput
                    v-model="configForm.upstream_key"
                    show-password
                    placeholder="同系统/29系统填写"
                  />
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">价格与同步</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                极光价格独立维护，用户下单金额按商品单价叠加用户倍率计算。
              </p>
            </div>

            <div class="mt-5 space-y-4">
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">
                    晨跑单价（元/km）
                  </label>
                  <ElInputNumber
                    v-model="configForm.morning_price"
                    class="w-full"
                    :min="0.01"
                    :precision="2"
                    :step="0.1"
                  />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">
                    日常跑单价（元/km）
                  </label>
                  <ElInputNumber
                    v-model="configForm.daily_price"
                    class="w-full"
                    :min="0.01"
                    :precision="2"
                    :step="0.1"
                  />
                </div>
              </div>
              <div class="grid gap-4 md:grid-cols-[1fr_120px]">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">同步间隔(秒)</label>
                  <ElInputNumber
                    v-model="configForm.sync_interval"
                    class="w-full"
                    :min="60"
                    :precision="0"
                    :step="30"
                  />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">自动同步</label>
                  <div class="flex h-8 items-center">
                    <ElSwitch v-model="configForm.auto_sync" />
                  </div>
                </div>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">请求超时(秒)</label>
                <ElInputNumber
                  v-model="configForm.timeout"
                  class="w-full"
                  :min="5"
                  :max="120"
                  :precision="0"
                  :step="5"
                />
              </div>
            </div>
          </section>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="refundVisible" title="退款确认" width="520px">
      <div v-if="refundPreview" class="jiguang-confirm">
        <div>
          <span>订单</span>
          <strong>{{ refundPreview.product_name }} · {{ refundPreview.student_account }}</strong>
        </div>
        <div>
          <span>剩余次数</span>
          <strong>{{ refundPreview.remaining }} 次</strong>
        </div>
        <div>
          <span>计费</span>
          <strong
            >{{ refundPreview.km_per_day }} km × {{ money(refundPreview.price_per_km) }}/km</strong
          >
        </div>
        <div class="is-danger">
          <span>退款金额</span>
          <strong>{{ money(refundPreview.amount) }}</strong>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="refundVisible = false">取消</ElButton>
          <ElButton type="danger" :loading="refundLoading" @click="handleConfirmRefund"
            >确认退款</ElButton
          >
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="addTimesVisible" title="加次数" width="520px">
      <ElForm label-position="top" class="jiguang-form">
        <ElFormItem label="增加次数">
          <ElInputNumber v-model="addTimesDelta" class="w-full" :min="1" :max="9999" :step="1" />
        </ElFormItem>
        <div v-if="addTimesPreview" class="jiguang-confirm">
          <div>
            <span>次数</span>
            <strong
              >{{ addTimesPreview.before_run_times }} →
              {{ addTimesPreview.after_run_times }}</strong
            >
          </div>
          <div>
            <span>计费</span>
            <strong
              >{{ addTimesPreview.delta }} 次 × {{ addTimesPreview.km_per_day }} km ×
              {{ money(addTimesPreview.price_per_km) }}/km</strong
            >
          </div>
          <div class="is-danger">
            <span>扣费</span>
            <strong>{{ money(addTimesPreview.cost) }}</strong>
          </div>
        </div>
      </ElForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addTimesVisible = false">取消</ElButton>
          <ElButton :loading="addTimesLoading" @click="loadAddTimesPreview">预览</ElButton>
          <ElButton
            type="primary"
            :loading="addTimesConfirming"
            :disabled="!addTimesPreview"
            @click="handleConfirmAddTimes"
          >
            确认加次数
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="logsVisible" title="订单日志" width="680px">
      <ElTimeline v-if="logs.length">
        <ElTimelineItem
          v-for="item in logs"
          :key="String(item.id || item.createdAt)"
          :timestamp="formatLogTime(item)"
        >
          <div class="jiguang-log-item">
            <p>{{ item.action || item.type || '日志' }}</p>
            <span v-if="item.type === 'addTimes'"
              >加次数：{{ item.delta }} 次，扣费 {{ item.cost }}</span
            >
            <span v-if="item.type === 'adjustCompleted'"
              >完成次数调整：{{ item.before }} → {{ item.after }}</span
            >
            <span v-if="item.type === 'refund'">退款金额：{{ item.refundAmount }}</span>
            <span v-if="item.operator">操作人：{{ item.operator }}</span>
          </div>
        </ElTimelineItem>
      </ElTimeline>
      <ElEmpty v-else description="暂无日志" />
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElButton,
    ElDropdown,
    ElDropdownItem,
    ElDropdownMenu,
    ElMessage,
    ElMessageBox,
    ElTag
  } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    confirmLegacyJiguangAddTimes,
    confirmLegacyJiguangRefund,
    createLegacyJiguangOrder,
    fetchLegacyJiguangConfig,
    fetchLegacyJiguangLogs,
    fetchLegacyJiguangOrders,
    fetchLegacyJiguangProducts,
    fetchLegacyJiguangSchools,
    previewLegacyJiguangAddTimes,
    previewLegacyJiguangRefund,
    saveLegacyJiguangConfig,
    syncLegacyJiguangOrders,
    type LegacyJiguangAddTimesItem,
    type LegacyJiguangConfig,
    type LegacyJiguangOrder,
    type LegacyJiguangProduct,
    type LegacyJiguangRefundItem
  } from '@/api/legacy/plugin-jiguang'

  defineOptions({ name: 'PluginJiguangPage' })

  const userStore = useUserStore()
  const isAdmin = computed(() => {
    const info = userStore.info as any
    const roles = Array.isArray(info?.roles)
      ? info.roles.map((item: string) => String(item).toLowerCase())
      : []
    const role = String(info?.role || info?.roleCode || '').toLowerCase()
    const uid = Number(info?.uid || info?.userId || 0)
    return (
      uid === 1 ||
      role === 'admin' ||
      role === 'super' ||
      roles.includes('r_admin') ||
      roles.includes('r_super')
    )
  })

  const kmOptions = [1, 1.2, 1.5, 1.6, 2, 3, 5, 10]
  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '待处理', value: 'pending' },
    { label: '进行中', value: 'in_progress' },
    { label: '今日完成', value: 'today_done' },
    { label: '已结束', value: 'finished' },
    { label: '姓名学号不匹配', value: 'name_mismatch' },
    { label: '已退款', value: 'refunded' }
  ]
  const searchTypeOptions = [
    { label: '综合搜索', value: '' },
    { label: '订单ID', value: '1' },
    { label: '学号', value: '2' },
    { label: '用户UID', value: '3' }
  ]

  const loading = ref(false)
  const syncLoading = ref(false)
  const createLoading = ref(false)
  const schoolLoading = ref(false)
  const configLoading = ref(false)
  const saveConfigLoading = ref(false)
  const refundLoading = ref(false)
  const addTimesLoading = ref(false)
  const addTimesConfirming = ref(false)

  const addVisible = ref(false)
  const configVisible = ref(false)
  const refundVisible = ref(false)
  const addTimesVisible = ref(false)
  const logsVisible = ref(false)

  const products = ref<LegacyJiguangProduct[]>([])
  const schools = ref<any[]>([])
  const orders = ref<LegacyJiguangOrder[]>([])
  const logs = ref<any[]>([])
  const activeOrder = ref<LegacyJiguangOrder | null>(null)
  const refundPreview = ref<LegacyJiguangRefundItem | null>(null)
  const addTimesPreview = ref<LegacyJiguangAddTimesItem | null>(null)
  const addTimesDelta = ref(1)

  const pagination = reactive({ limit: 20, page: 1, total: 0 })
  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.limit,
    total: pagination.total
  }))

  const filters = reactive({
    keyword: '',
    school: '',
    searchType: '',
    status: ''
  })

  const searchItems = computed(() => [
    {
      label: '搜索字段',
      key: 'searchType',
      type: 'select',
      props: { clearable: true, options: searchTypeOptions, placeholder: '搜索字段' }
    },
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: { clearable: true, placeholder: '订单号 / 学号 / 姓名 / 学校' }
    },
    {
      label: '学校',
      key: 'school',
      type: 'input',
      props: { clearable: true, placeholder: '学校名称' }
    },
    {
      label: '订单状态',
      key: 'status',
      type: 'select',
      props: { clearable: true, options: statusOptions, placeholder: '订单状态' }
    }
  ])

  const defaultConfig = (): LegacyJiguangConfig => ({
    api_key: '',
    auto_sync: true,
    daily_price: 1,
    morning_price: 1,
    sync_cursor_page: 2,
    sync_interval: 300,
    timeout: 30,
    upstream_key: '',
    upstream_protocol: 'source',
    upstream_uid: 0,
    upstream_url: ''
  })

  const configForm = reactive<LegacyJiguangConfig>(defaultConfig())
  const orderForm = reactive({
    customer_message: '',
    km_per_day: 1.5,
    product_id: 1,
    school_name: '',
    student_account: '',
    student_name: '',
    times: 1
  })

  const selectedProduct = computed(() =>
    products.value.find((item) => item.product_id === orderForm.product_id)
  )
  const estimatedAmount = computed(() =>
    Number(
      ((selectedProduct.value?.price || 0) * orderForm.times * orderForm.km_per_day).toFixed(2)
    )
  )

  const money = (value: number | string | null | undefined) => `¥${Number(value || 0).toFixed(2)}`
  const progressText = (row: LegacyJiguangOrder) => `${row.completed_times}/${row.run_times}`
  const schoolLabel = (item: any) =>
    String(item?.name || item?.schoolName || item?.school_name || item || '')
  const schoolKey = (item: any) => `${schoolLabel(item)}-${item?.id || item?.code || ''}`

  const getStatusText = (status: string) =>
    statusOptions.find((item) => item.value === status)?.label || status || '-'

  const getStatusType = (status: string) =>
    (
      ({
        finished: 'success',
        in_progress: 'primary',
        name_mismatch: 'danger',
        pending: 'warning',
        refunded: 'info',
        today_done: 'success'
      }) as Record<string, 'danger' | 'info' | 'primary' | 'success' | 'warning'>
    )[status] || 'info'

  const canRefund = (row: LegacyJiguangOrder) => row.status !== 'refunded'

  const handleRowCommand = (row: LegacyJiguangOrder, command: string) => {
    if (command === 'logs') {
      openLogs(row)
      return
    }
    if (command === 'addTimes') {
      openAddTimes(row)
      return
    }
    if (command === 'refund') {
      openRefund(row)
    }
  }

  const { columns, columnChecks } = useTableColumns<LegacyJiguangOrder>(() => [
    ...(isAdmin.value
      ? [
          {
            prop: 'uid',
            label: '下单用户',
            width: 130,
            formatter: (row: LegacyJiguangOrder) =>
              h('div', { class: 'leading-6' }, [
                h('p', { class: 'font-mono text-sm font-semibold text-g-900' }, `UID ${row.uid}`),
                h('p', { class: 'truncate text-xs text-g-500' }, row.username || '-')
              ])
          }
        ]
      : []),
    {
      prop: 'student_account',
      label: '学生信息',
      minWidth: 190,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.student_name || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.student_account || '-')
        ])
    },
    {
      prop: 'product_name',
      label: '跑步配置',
      minWidth: 120,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, row.product_name || '-')
        ])
    },
    {
      prop: 'run_times',
      label: '次数',
      width: 110,
      align: 'center',
      formatter: (row) => h('span', { class: 'font-mono text-sm text-g-700' }, progressText(row))
    },
    {
      prop: 'km_per_day',
      label: '公里',
      width: 110,
      align: 'center',
      formatter: (row) =>
        h('span', { class: 'font-mono text-sm text-g-700' }, `${row.km_per_day} km`)
    },
    {
      prop: 'customer_message',
      label: '备注',
      minWidth: 140,
      formatter: (row) =>
        h(
          'span',
          { class: 'block truncate text-sm text-g-600', title: row.customer_message || '' },
          row.customer_message || '-'
        )
    },
    { prop: 'school_name', label: '学校', minWidth: 180 },
    {
      prop: 'status',
      label: '状态',
      width: 130,
      align: 'center',
      formatter: (row) =>
        h(ElTag, { type: getStatusType(row.status), effect: 'plain' }, () =>
          getStatusText(row.status)
        )
    },
    {
      prop: 'fees',
      label: '金额',
      width: 110,
      align: 'right',
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-danger)]' }, money(row.fees))
    },
    {
      prop: 'order_no',
      label: '订单号',
      minWidth: 180,
      formatter: (row) => h('span', { class: 'font-mono text-xs text-g-600' }, row.order_no || '-')
    },
    { prop: 'created_at', label: '下单时间', minWidth: 160 },
    {
      prop: 'operation',
      label: '操作',
      width: 110,
      fixed: 'right' as const,
      formatter: (row) =>
        h(
          ElDropdown,
          {
            trigger: 'click',
            onCommand: (command: string) => handleRowCommand(row, command)
          },
          {
            default: () => h(ElButton, { size: 'small' }, () => '操作'),
            dropdown: () =>
              h(ElDropdownMenu, null, () => [
                h(ElDropdownItem, { command: 'logs' }, () => '查看日志'),
                h(
                  ElDropdownItem,
                  { command: 'addTimes', disabled: row.status === 'refunded' },
                  () => '加次数'
                ),
                h(ElDropdownItem, { command: 'refund', disabled: !canRefund(row) }, () => '退款')
              ])
          }
        )
    }
  ])

  const loadProducts = async () => {
    products.value = (await fetchLegacyJiguangProducts()) || []
  }

  const searchSchools = async (keyword = '') => {
    if (!keyword.trim()) {
      schools.value = []
      return
    }
    schoolLoading.value = true
    try {
      const result = await fetchLegacyJiguangSchools({ keyword, page: 1, pageSize: 50 })
      const list = Array.isArray(result?.list)
        ? result.list
        : Array.isArray(result?.data?.list)
          ? result.data.list
          : Array.isArray(result)
            ? result
            : []
      schools.value = list
    } finally {
      schoolLoading.value = false
    }
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyJiguangOrders({
        keyword: filters.keyword || undefined,
        limit: pagination.limit,
        page: pagination.page,
        school: filters.school || undefined,
        searchType: filters.searchType || undefined,
        status: filters.status || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = () => loadOrders(1)
  const resetFilters = () => {
    filters.keyword = ''
    filters.school = ''
    filters.searchType = ''
    filters.status = ''
    loadOrders(1)
  }
  const handleSizeChange = (size: number) => {
    pagination.limit = size
    loadOrders(1)
  }

  const resetOrderForm = () => {
    Object.assign(orderForm, {
      customer_message: '',
      km_per_day: 1.5,
      product_id: products.value[0]?.product_id || 1,
      school_name: '',
      student_account: '',
      student_name: '',
      times: 1
    })
    schools.value = []
  }

  const openAddDialog = () => {
    resetOrderForm()
    addVisible.value = true
  }

  const handleCreateOrder = async () => {
    if (!orderForm.school_name || !orderForm.student_name || !orderForm.student_account) {
      ElMessage.warning('请填写学校、姓名和学号')
      return
    }
    await ElMessageBox.confirm(`预计扣费 ${money(estimatedAmount.value)}，确认下单？`, '确认下单', {
      type: 'warning'
    })
    createLoading.value = true
    try {
      await createLegacyJiguangOrder({ ...orderForm })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      createLoading.value = false
    }
  }

  const loadConfig = async () => {
    configLoading.value = true
    try {
      Object.assign(configForm, defaultConfig(), await fetchLegacyJiguangConfig())
    } finally {
      configLoading.value = false
    }
  }

  const openConfigDialog = async () => {
    configVisible.value = true
    await loadConfig()
  }

  const handleSaveConfig = async () => {
    saveConfigLoading.value = true
    try {
      await saveLegacyJiguangConfig({ ...configForm })
      ElMessage.success('配置已保存')
      loadProducts()
    } finally {
      saveConfigLoading.value = false
    }
  }

  const handleSyncOrders = async () => {
    syncLoading.value = true
    try {
      const result = await syncLegacyJiguangOrders()
      ElMessage.success(`同步完成，更新 ${result?.updated || 0} 条`)
      loadOrders(pagination.page)
    } finally {
      syncLoading.value = false
    }
  }

  const openRefund = async (row: LegacyJiguangOrder) => {
    activeOrder.value = row
    refundPreview.value = null
    refundLoading.value = true
    refundVisible.value = true
    try {
      const result = await previewLegacyJiguangRefund(row.order_no)
      refundPreview.value = result.item
    } finally {
      refundLoading.value = false
    }
  }

  const handleConfirmRefund = async () => {
    if (!activeOrder.value) return
    refundLoading.value = true
    try {
      await confirmLegacyJiguangRefund(activeOrder.value.order_no)
      ElMessage.success('退款成功')
      refundVisible.value = false
      loadOrders(pagination.page)
    } finally {
      refundLoading.value = false
    }
  }

  const openAddTimes = (row: LegacyJiguangOrder) => {
    activeOrder.value = row
    addTimesDelta.value = 1
    addTimesPreview.value = null
    addTimesVisible.value = true
  }

  const loadAddTimesPreview = async () => {
    if (!activeOrder.value) return
    addTimesLoading.value = true
    try {
      const result = await previewLegacyJiguangAddTimes(
        activeOrder.value.order_no,
        addTimesDelta.value
      )
      addTimesPreview.value = result.item
    } finally {
      addTimesLoading.value = false
    }
  }

  const handleConfirmAddTimes = async () => {
    if (!activeOrder.value || !addTimesPreview.value) return
    addTimesConfirming.value = true
    try {
      await confirmLegacyJiguangAddTimes(activeOrder.value.order_no, addTimesDelta.value)
      ElMessage.success('加次数成功')
      addTimesVisible.value = false
      loadOrders(pagination.page)
    } finally {
      addTimesConfirming.value = false
    }
  }

  const openLogs = async (row: LegacyJiguangOrder) => {
    activeOrder.value = row
    logs.value = []
    logsVisible.value = true
    const result = await fetchLegacyJiguangLogs(row.order_no)
    logs.value = result?.list || []
  }

  const formatLogTime = (item: any) => String(item.createdAt || item.created_at || item.time || '')

  onMounted(() => {
    loadProducts()
    loadOrders(1)
  })
</script>

<style scoped lang="scss">
  .plugin-jiguang-page {
    display: flex;
    flex-direction: column;
    gap: 16px;

    .jiguang-search-bar {
      flex-shrink: 0;
    }

    .jiguang-table-card {
      margin-top: 0;
    }

    .jiguang-table-toolbar {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 10px;
      min-height: 34px;
    }

    .jiguang-table-title {
      margin: 0 8px 0 0;
      color: var(--el-text-color-primary);
      font-size: 16px;
      font-weight: 600;
      line-height: 32px;
      white-space: nowrap;
    }

    .jiguang-form {
      :deep(.el-form-item) {
        margin-bottom: 16px;
      }
    }

    .jiguang-order-form {
      display: grid;
      gap: 16px;
    }

    .jiguang-form-section {
      display: grid;
      gap: 16px;
      padding: 16px;
      border: 1px solid var(--art-card-border);
      border-radius: 8px;
      background: var(--art-main-bg-color);
    }

    .jiguang-section-title {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: space-between;
      gap: 8px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--art-card-border);

      span {
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 600;
      }

      small {
        color: var(--el-text-color-secondary);
        font-size: 12px;
      }
    }

    .jiguang-form-grid {
      display: grid;
      gap: 16px;
    }

    .jiguang-form-field {
      label {
        display: block;
        margin-bottom: 8px;
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 500;
      }
    }

    .jiguang-order-summary {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
      padding: 12px 14px;
      border-radius: 6px;
      background: var(--art-gray-100);
      font-size: 13px;

      div {
        min-width: 0;
      }

      span {
        display: block;
        color: var(--el-text-color-secondary);
      }

      strong {
        display: block;
        overflow: hidden;
        margin-top: 4px;
        color: var(--el-text-color-primary);
        font-weight: 600;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .jiguang-estimate,
    .jiguang-confirm {
      border: 1px solid var(--art-border-color);
      border-radius: 6px;
      background: var(--art-main-bg-color);
    }

    .jiguang-estimate {
      display: grid;
      grid-template-columns: minmax(0, 1fr) repeat(3, auto);
      gap: 12px;
      align-items: center;
      padding: 12px 14px;
      color: var(--el-text-color-secondary);
      font-size: 13px;

      span {
        min-width: 0;
      }

      strong {
        color: var(--el-color-danger);
        font-size: 15px;
      }
    }

    .jiguang-confirm {
      overflow: hidden;

      div {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        padding: 12px 14px;
        border-bottom: 1px solid var(--art-border-color);
        font-size: 14px;

        &:last-child {
          border-bottom: 0;
        }
      }

      span {
        color: var(--el-text-color-secondary);
      }

      strong {
        color: var(--el-text-color-primary);
        font-weight: 600;
        text-align: right;
      }

      .is-danger strong {
        color: var(--el-color-danger);
      }
    }

    .jiguang-log-item {
      display: grid;
      gap: 4px;
      color: var(--el-text-color-secondary);
      font-size: 13px;
      line-height: 1.6;

      p {
        margin: 0;
        color: var(--el-text-color-primary);
        font-weight: 600;
      }
    }
  }

  @media (max-width: 768px) {
    .plugin-jiguang-page {
      .jiguang-order-summary {
        grid-template-columns: 1fr 1fr;
      }

      .jiguang-estimate {
        grid-template-columns: 1fr 1fr;
      }

      .jiguang-confirm div {
        align-items: flex-start;
        flex-direction: column;
        gap: 4px;
      }
    }
  }
</style>
