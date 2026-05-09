<template>
  <div class="plugin-wuxin-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      class="wuxin-search-bar"
      :items="searchItems"
      :showExpand="false"
      @reset="resetFilters"
      @search="handleSearch"
    />

    <ElCard class="art-table-card wuxin-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        layout="refresh,size,fullscreen,columns,settings"
        fullClass="art-table-card"
        @refresh="loadOrders(pagination.page)"
      >
        <template #left>
          <div class="wuxin-table-toolbar">
            <h4 class="wuxin-table-title">无心闪动订单</h4>
            <ElButton type="primary" @click="openAddDialog">提交订单</ElButton>
            <ElButton plain :disabled="!configForm.auth_url" @click="openAuthUrl"
              >授权链接</ElButton
            >
            <ElButton v-if="isAdmin" plain @click="openConfigDialog">配置接入</ElButton>
            <ElButton v-if="isAdmin" plain :loading="syncLoading" @click="handleSyncOrders"
              >同步上游</ElButton
            >
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        rowKey="id"
        :columns="columns"
        :data="orders"
        empty-height="360px"
        :loading="loading"
        :pagination="tablePagination"
        @pagination:current-change="loadOrders"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog
      v-model="addVisible"
      class="wuxin-order-dialog"
      title="提交无心闪动订单"
      width="860px"
      destroy-on-close
    >
      <div class="wuxin-order-form">
        <section class="wuxin-auth-panel">
          <div class="wuxin-form-field">
            <label>
              <span>授权码</span>
              <small>通过授权链接获取</small>
            </label>
            <ElInput v-model="orderForm.auth_code" clearable placeholder="请输入授权码" />
          </div>
          <ElButton
            class="wuxin-auth-button"
            plain
            :loading="schoolLoading"
            @click="handleQuerySchoolInfo"
          >
            获取信息
          </ElButton>
        </section>

        <section class="wuxin-form-section">
          <div class="wuxin-section-title">
            <span>订单参数</span>
            <small>授权信息读取后选择计划和跑区</small>
          </div>
          <div class="wuxin-form-grid md:grid-cols-2">
            <div class="wuxin-form-field">
              <label>开始日期</label>
              <ElDatePicker
                v-model="orderForm.start_date"
                class="w-full"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="请选择开始日期"
              />
            </div>
            <div class="wuxin-form-field">
              <label>下单次数</label>
              <ElInputNumber v-model="orderForm.order_num" class="w-full" :min="1" :max="999" />
            </div>
            <div class="wuxin-form-field">
              <label>跑步计划</label>
              <ElSelect
                v-model="orderForm.run_plan_code"
                class="w-full"
                filterable
                placeholder="请选择跑步计划"
              >
                <ElOption
                  v-for="item in runPlanOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>
            <div class="wuxin-form-field">
              <label>跑步区域</label>
              <ElSelect
                v-model="orderForm.fence_code"
                class="w-full"
                filterable
                placeholder="请选择跑步区域"
              >
                <ElOption label="默认跑区" value="default" />
                <ElOption
                  v-for="item in runZoneOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>
          </div>
        </section>

        <section class="wuxin-form-section">
          <div class="wuxin-section-title">
            <span>跑步设置</span>
            <small>按学校规则填写距离、时间和周期</small>
          </div>
          <div class="wuxin-form-grid md:grid-cols-3">
            <div class="wuxin-form-field">
              <label>跑步类型</label>
              <ElRadioGroup v-model="orderForm.run_type">
                <ElRadioButton :label="1">阳光跑</ElRadioButton>
              </ElRadioGroup>
            </div>
            <div class="wuxin-form-field">
              <label>跑步距离</label>
              <ElInputNumber
                v-model="orderForm.run_meter"
                class="w-full"
                :min="0.5"
                :max="30"
                :precision="1"
                :step="0.1"
              />
            </div>
            <div class="wuxin-form-field">
              <label>配速</label>
              <div class="wuxin-pace-input">
                <ElInputNumber
                  v-model="paceValue"
                  class="w-full"
                  :min="4"
                  :max="15"
                  :precision="1"
                  :step="0.1"
                  controls-position="right"
                />
                <span>分钟/公里</span>
              </div>
            </div>
          </div>
          <div class="wuxin-form-grid md:grid-cols-2">
            <div class="wuxin-form-field">
              <label>开始时间</label>
              <ElTimePicker
                v-model="startTime"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="开始时间"
              />
            </div>
            <div class="wuxin-form-field">
              <label>结束时间</label>
              <ElTimePicker
                v-model="endTime"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="结束时间"
              />
            </div>
          </div>
          <div class="wuxin-form-field">
            <label>跑步周期</label>
            <ElSelect
              v-model="orderForm.run_week"
              class="w-full"
              multiple
              placeholder="请选择跑步周期"
            >
              <ElOption
                v-for="item in weekOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
          </div>
        </section>

        <section class="wuxin-form-section">
          <div class="wuxin-section-title">
            <span>提交确认</span>
            <small>标记便于后续检索订单</small>
          </div>
          <div class="wuxin-form-field">
            <label>客户标记</label>
            <ElInput v-model="orderForm.mark" maxlength="80" placeholder="选填" />
          </div>
          <div class="wuxin-order-summary">
            <div>
              <span>跑步计划</span>
              <strong>{{ selectedRunPlanLabel || '-' }}</strong>
            </div>
            <div>
              <span>跑步区域</span>
              <strong>{{ selectedRunZoneLabel || '默认跑区' }}</strong>
            </div>
            <div>
              <span>预估金额</span>
              <strong class="text-[var(--el-color-danger)]">¥{{ estimatedPrice }}</strong>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="createLoading" @click="handleCreateOrder">
            确认提交
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="configVisible" title="无心闪动配置" width="980px" destroy-on-close>
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
                选择源台原生接口、29系统兼容接口或本系统 OpenAPI。
              </p>
            </div>

            <div class="mt-5 grid gap-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">对接协议</label>
                <ElSelect v-model="configForm.upstream_protocol" class="w-full">
                  <ElOption label="源台" value="source" />
                  <ElOption label="29系统" value="source29" />
                  <ElOption label="同系统（本系统）" value="same_system" />
                </ElSelect>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">上游地址</label>
                <ElInput v-model="configForm.upstream_url" placeholder="例如 https://example.com" />
              </div>
              <div v-if="configForm.upstream_protocol === 'source'">
                <label class="mb-2 block text-sm font-medium text-g-800">源台 API Key</label>
                <ElInput v-model="configForm.api_key" show-password />
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
                  <label class="mb-2 block text-sm font-medium text-g-800">上游 Key</label>
                  <ElInput v-model="configForm.upstream_key" show-password />
                </div>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">授权链接</label>
                <ElInput v-model="configForm.auth_url" placeholder="选填，展示在订单页快捷打开" />
              </div>
            </div>
          </section>

          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">价格与同步</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                用户下单价格按基础单价叠加用户倍率，自动同步按秒执行。
              </p>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">基础单价</label>
                <ElInputNumber
                  v-model="configForm.price"
                  class="w-full"
                  :min="0"
                  :precision="2"
                  :step="0.5"
                />
              </div>
              <div class="grid gap-4 md:grid-cols-[1fr_120px]">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">同步间隔(秒)</label>
                  <ElInputNumber
                    v-model="configForm.sync_interval"
                    class="w-full"
                    :min="30"
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
                />
              </div>
            </div>
          </section>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="logsVisible" title="订单执行日志" width="980px">
      <ArtTable
        :data="logs"
        :loading="logsLoading"
        :columns="logColumns"
        empty-height="260px"
        empty-text="暂无执行日志"
        :show-table-header="false"
        :pagination="logsTablePagination"
        @pagination:current-change="loadLogs"
        @pagination:size-change="handleLogsSizeChange"
      />
    </ElDialog>

    <ElDialog v-model="editVisible" title="编辑订单" width="900px" destroy-on-close>
      <div class="space-y-5">
        <section class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm text-g-600">
          当前订单号：<span class="font-semibold text-g-900">{{ editOrderNumber || '-' }}</span>
        </section>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">跑步计划</label>
            <ElSelect v-model="editForm.run_plan_code" class="w-full" filterable>
              <ElOption
                v-for="item in editRunPlanOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">跑步区域</label>
            <ElSelect v-model="editForm.fence_code" class="w-full" filterable>
              <ElOption label="默认跑区" value="default" />
              <ElOption
                v-for="item in editRunZoneOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
          </div>
        </div>
        <div class="grid gap-4 md:grid-cols-4">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">跑步距离</label>
            <ElInputNumber
              v-model="editForm.run_meter"
              class="w-full"
              :min="0.5"
              :precision="1"
              :step="0.1"
            />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">配速</label>
            <div class="wuxin-pace-input">
              <ElInputNumber
                v-model="editPaceValue"
                class="w-full"
                :min="4"
                :max="15"
                :precision="1"
                :step="0.1"
                controls-position="right"
              />
              <span>分钟/公里</span>
            </div>
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">开始时间</label>
            <ElTimePicker
              v-model="editStartTime"
              class="w-full"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">结束时间</label>
            <ElTimePicker
              v-model="editEndTime"
              class="w-full"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">跑步周期</label>
          <ElSelect
            v-model="editForm.run_week"
            class="w-full"
            multiple
            placeholder="请选择跑步周期"
          >
            <ElOption
              v-for="item in weekOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">客户标记</label>
          <ElInput v-model="editForm.mark" maxlength="80" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="editLoading" @click="handleEditOrder"
            >保存修改</ElButton
          >
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="increaseVisible" title="追加次数" width="460px">
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm text-g-600">
          订单 #{{ increaseForm.id }}，当前将按单次价格扣费。
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">追加次数</label>
          <ElInputNumber v-model="increaseForm.quantity" class="w-full" :min="1" :max="999" />
        </div>
        <div class="text-sm">
          预计扣费
          <span class="font-semibold text-[var(--el-color-danger)]">¥{{ increasePrice }}</span>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="increaseVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="increaseLoading" @click="handleIncreaseOrder">
            确认追加
          </ElButton>
        </div>
      </template>
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
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useUserStore } from '@/store/modules/user'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import {
    createLegacyWuxinOrder,
    editLegacyWuxinOrder,
    fetchLegacyWuxinConfig,
    fetchLegacyWuxinOrderConfig,
    fetchLegacyWuxinOrders,
    fetchLegacyWuxinPrice,
    fetchLegacyWuxinRecords,
    increaseLegacyWuxinOrder,
    queryLegacyWuxinSchoolInfo,
    reassignLegacyWuxinOrder,
    refundLegacyWuxinOrder,
    saveLegacyWuxinConfig,
    syncLegacyWuxinOrders,
    type LegacyWuxinConfig,
    type LegacyWuxinOrder,
    type LegacyWuxinOrderForm
  } from '@/api/legacy/plugin-wuxin'

  defineOptions({ name: 'PluginWuxinPage' })

  interface OptionItem {
    label: string
    value: string
  }

  const userStore = useUserStore()
  const isAdmin = computed(() => {
    const info = userStore.info as any
    const roles = Array.isArray(info?.roles)
      ? info.roles.map((role: string) => String(role).toLowerCase())
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

  const weekOptions = [
    { label: '周一', value: 1 },
    { label: '周二', value: 2 },
    { label: '周三', value: 3 },
    { label: '周四', value: 4 },
    { label: '周五', value: 5 },
    { label: '周六', value: 6 },
    { label: '周日', value: 7 }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '待处理', value: '0' },
    { label: '等待计划', value: '1' },
    { label: '处理中', value: '2' },
    { label: '已完成', value: '3' },
    { label: '已退款', value: '4' },
    { label: '错误', value: '5' }
  ]

  const searchFieldOptions = [
    { label: '订单ID', value: '1' },
    { label: '授权码', value: '2' },
    { label: '用户UID', value: '3' },
    { label: '客户标记', value: '4' },
    { label: '手机号', value: '5' }
  ]

  const configForm = reactive<LegacyWuxinConfig>({
    api_key: '',
    auth_url: '',
    auto_sync: false,
    price: 5,
    sync_interval: 300,
    timeout: 30,
    upstream_key: '',
    upstream_protocol: 'source',
    upstream_uid: 0,
    upstream_url: ''
  })

  const loading = ref(false)
  const priceLoading = ref(false)
  const configLoading = ref(false)
  const saveConfigLoading = ref(false)
  const syncLoading = ref(false)
  const schoolLoading = ref(false)
  const createLoading = ref(false)
  const logsLoading = ref(false)
  const editLoading = ref(false)
  const increaseLoading = ref(false)

  const addVisible = ref(false)
  const configVisible = ref(false)
  const logsVisible = ref(false)
  const editVisible = ref(false)
  const increaseVisible = ref(false)

  const orders = ref<LegacyWuxinOrder[]>([])
  const logs = ref<any[]>([])
  const runPlanOptions = ref<OptionItem[]>([])
  const runZoneOptions = ref<OptionItem[]>([])
  const editRunPlanOptions = ref<OptionItem[]>([])
  const editRunZoneOptions = ref<OptionItem[]>([])
  const editOrderId = ref(0)
  const editOrderNumber = ref('')

  const priceInfo = reactive({
    base_price: 5,
    price: 0,
    user_rate: 1
  })

  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.limit,
    total: pagination.total
  }))

  const logsPagination = reactive({
    id: 0,
    page: 1,
    limit: 20,
    total: 0
  })

  const logsTablePagination = computed(() => ({
    current: logsPagination.page,
    size: logsPagination.limit,
    total: logsPagination.total
  }))

  const filters = reactive({
    keyword: '',
    searchType: '1',
    status: ''
  })

  const searchItems = computed(() => [
    {
      label: '订单状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        options: statusOptions,
        placeholder: '订单状态'
      }
    },
    {
      label: '搜索项',
      key: 'searchType',
      type: 'select',
      props: {
        options: searchFieldOptions,
        placeholder: '搜索项'
      }
    },
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索订单ID、授权码、UID、手机号或标记'
      }
    }
  ])

  function formatDate(date: Date) {
    const pad = (value: number) => String(value).padStart(2, '0')
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
  }

  const defaultOrderForm = () => ({
    auth_code: '',
    fence_code: 'default',
    mark: '',
    order_num: 1,
    run_meter: 1,
    run_plan_code: '',
    run_speed: '6',
    run_time: '06:00-08:00',
    run_type: 1,
    run_week: [1, 2, 3, 4, 5] as number[],
    start_date: formatDate(new Date()),
    zone_name: '默认跑区'
  })

  const orderForm = reactive(defaultOrderForm())
  const editForm = reactive(defaultOrderForm())
  const startTime = ref('06:00')
  const endTime = ref('08:00')
  const editStartTime = ref('06:00')
  const editEndTime = ref('08:00')
  const paceValue = ref(6)
  const editPaceValue = ref(6)
  const increaseForm = reactive({
    id: 0,
    order_number: '',
    quantity: 1
  })

  const selectedRunPlanLabel = computed(
    () => runPlanOptions.value.find((item) => item.value === orderForm.run_plan_code)?.label || ''
  )
  const selectedRunZoneLabel = computed(
    () => runZoneOptions.value.find((item) => item.value === orderForm.fence_code)?.label || ''
  )
  const estimatedPrice = computed(() =>
    (Number(priceInfo.price || 0) * Number(orderForm.order_num || 0)).toFixed(2)
  )
  const increasePrice = computed(() =>
    (Number(priceInfo.price || 0) * Number(increaseForm.quantity || 0)).toFixed(2)
  )
  const { columns, columnChecks } = useTableColumns<LegacyWuxinOrder>(() => [
    {
      prop: 'id',
      label: '订单',
      width: 150,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, `#${row.id}`),
          h('p', { class: 'text-xs text-g-500' }, row.order_number || '未下发源台订单号')
        ])
    },
    {
      prop: 'auth_code',
      label: '授权信息',
      minWidth: 160,
      formatter: (row) =>
        h('div', { class: 'wuxin-table-cell leading-6' }, [
          h('p', { class: 'truncate font-semibold text-g-900' }, row.auth_code || '-'),
          h('p', { class: 'truncate text-xs text-g-500' }, row.phone || '未同步手机号')
        ])
    },
    {
      prop: 'quantity',
      label: '任务进度',
      minWidth: 170,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, `${row.completed_quantity}/${row.quantity} 次`),
          h('p', { class: 'text-xs text-g-500' }, `剩余 ${row.residue_num} 次，${row.run_meter} km`)
        ])
    },
    {
      prop: 'zone_name',
      label: '跑步配置',
      minWidth: 168,
      formatter: (row) =>
        h('div', { class: 'wuxin-table-cell leading-6' }, [
          h('p', { class: 'truncate text-sm text-g-800' }, row.zone_name || '默认跑区'),
          h(
            'p',
            { class: 'truncate text-xs text-g-500' },
            `${row.run_time || '-'} / ${formatWeek(row.run_week)}`
          )
        ])
    },
    {
      prop: 'order_status',
      label: '状态',
      width: 130,
      align: 'center',
      formatter: (row) =>
        h(ElTag, { type: getOrderStatusType(row.order_status), effect: 'plain' }, () =>
          getOrderStatusLabel(row.order_status)
        )
    },
    {
      prop: 'mark',
      label: '标记 / 备注',
      minWidth: 170,
      formatter: (row) =>
        h('div', { class: 'wuxin-table-cell leading-6' }, [
          h('p', { class: 'truncate text-sm text-g-800' }, row.mark || '-'),
          h('p', { class: 'wuxin-two-line text-xs text-g-500' }, row.remarks || '暂无备注')
        ])
    },
    {
      prop: 'fees',
      label: '费用',
      width: 110,
      align: 'right',
      formatter: (row) =>
        h(
          'span',
          { class: 'font-semibold text-[var(--el-color-danger)]' },
          `¥${Number(row.fees || 0).toFixed(2)}`
        )
    },
    {
      prop: 'create_time',
      label: '创建时间',
      width: 150
    },
    {
      prop: 'operation',
      label: '操作',
      width: 178,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'wuxin-row-actions' }, [
          h(ArtButtonTable, {
            icon: 'ri:file-list-3-line',
            iconClass: 'bg-info/12 text-info',
            title: '日志',
            onClick: () => openLogsDialog(row)
          }),
          h(ArtButtonTable, {
            type: 'edit',
            title: '编辑',
            onClick: () => openEditDialog(row)
          }),
          h(
            ElDropdown,
            {
              trigger: 'click',
              onCommand: (command: string) => handleRowCommand(command, row)
            },
            {
              default: () =>
                h(ArtButtonTable, {
                  type: 'more',
                  title: '更多'
                }),
              dropdown: () =>
                h(ElDropdownMenu, null, () => [
                  h(
                    ElDropdownItem,
                    { command: 'increase', disabled: row.status !== 0 },
                    () => '追加'
                  ),
                  h(ElDropdownItem, { command: 'reassign' }, () => '重分配'),
                  h(
                    ElDropdownItem,
                    {
                      command: 'refund',
                      disabled: row.order_status === 4,
                      class: 'wuxin-danger-action'
                    },
                    () => '退款'
                  )
                ])
            }
          )
        ])
    }
  ])

  const { columns: logColumns } = useTableColumns<any>(() => [
    { prop: 'id', label: 'ID', width: 90 },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row) =>
        h(ElTag, { type: getLogStatusType(row.status), effect: 'plain' }, () =>
          getLogStatusLabel(row.status)
        )
    },
    {
      prop: 'progress',
      label: '进度',
      width: 90,
      formatter: (row) => `${Number(row.progress ?? 0)}%`
    },
    { prop: 'scheduled_time', label: '计划时间', minWidth: 160 },
    { prop: 'execute_time', label: '执行时间', minWidth: 160 },
    { prop: 'result', label: '执行结果', minWidth: 260 },
    { prop: 'created_at', label: '记录时间', minWidth: 160 }
  ])

  const normalizeOptions = (items: any[], valueKeys: string[], labelKeys: string[]) =>
    items
      .map((item) => {
        const value = valueKeys
          .map((key) => item?.[key])
          .find((field) => field !== undefined && field !== null && `${field}` !== '')
        const label = labelKeys
          .map((key) => item?.[key])
          .find((field) => field !== undefined && field !== null && `${field}` !== '')
        return { label: String(label || value || ''), value: String(value || '') }
      })
      .filter((item) => item.value)

  const parseSchoolInfo = (
    result: any,
    targetPlans = runPlanOptions,
    targetZones = runZoneOptions
  ) => {
    const data = result?.data?.data || result?.data || result
    const plans = data?.run_plans || data?.runPlans || []
    const zones = data?.areas || data?.zones || data?.run_zones || []
    targetPlans.value = normalizeOptions(
      plans,
      ['runPlanCode', 'run_plan_code', 'code'],
      ['runPlanName', 'run_plan_name', 'name']
    )
    targetZones.value = normalizeOptions(
      zones,
      ['fenceCode', 'fence_code', 'zone_code', 'code'],
      ['fenceName', 'fence_name', 'zone_name', 'name']
    )
    if (targetPlans.value.length === 1) {
      orderForm.run_plan_code = targetPlans.value[0]!.value
    }
    if (targetZones.value.length === 1) {
      orderForm.fence_code = targetZones.value[0]!.value
      orderForm.zone_name = targetZones.value[0]!.label
    }
  }

  const buildFormPayload = (form: ReturnType<typeof defaultOrderForm>): LegacyWuxinOrderForm => {
    const zone = runZoneOptions.value.find((item) => item.value === form.fence_code)
    return {
      auth_code: form.auth_code,
      fence_code: form.fence_code,
      mark: form.mark,
      order_num: Number(form.order_num || 1),
      run_meter: Number(form.run_meter || 0),
      run_plan_code: form.run_plan_code,
      run_speed: String(paceValue.value),
      run_time: `${startTime.value}-${endTime.value}`,
      run_type: Number(form.run_type || 1),
      run_week: JSON.stringify(form.run_week),
      start_date: form.start_date,
      zone_name: zone?.label || form.zone_name || '默认跑区'
    }
  }

  const buildEditPayload = (): LegacyWuxinOrderForm => {
    const zone = editRunZoneOptions.value.find((item) => item.value === editForm.fence_code)
    return {
      auth_code: editForm.auth_code,
      fence_code: editForm.fence_code,
      mark: editForm.mark,
      order_num: Number(editForm.order_num || 1),
      run_meter: Number(editForm.run_meter || 0),
      run_plan_code: editForm.run_plan_code,
      run_speed: String(editPaceValue.value),
      run_time: `${editStartTime.value}-${editEndTime.value}`,
      run_type: Number(editForm.run_type || 1),
      run_week: JSON.stringify(editForm.run_week),
      start_date: editForm.start_date,
      zone_name: zone?.label || editForm.zone_name || '默认跑区'
    }
  }

  const loadConfig = async () => {
    configLoading.value = true
    try {
      Object.assign(configForm, await fetchLegacyWuxinConfig())
    } finally {
      configLoading.value = false
    }
  }

  const loadPrice = async () => {
    priceLoading.value = true
    try {
      Object.assign(priceInfo, await fetchLegacyWuxinPrice())
    } finally {
      priceLoading.value = false
    }
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyWuxinOrders({
        keyword: filters.keyword || undefined,
        limit: pagination.limit,
        page,
        searchType: filters.keyword ? filters.searchType : undefined,
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
    filters.searchType = '1'
    filters.status = ''
    loadOrders(1)
  }

  const handleSizeChange = (size: number) => {
    pagination.limit = size
    loadOrders(1)
  }

  const openAddDialog = () => {
    Object.assign(orderForm, defaultOrderForm())
    startTime.value = '06:00'
    endTime.value = '08:00'
    paceValue.value = 6
    runPlanOptions.value = []
    runZoneOptions.value = []
    addVisible.value = true
  }

  const openAuthUrl = () => {
    if (configForm.auth_url) {
      window.open(configForm.auth_url, '_blank')
    }
  }

  const openConfigDialog = async () => {
    configVisible.value = true
    await loadConfig()
  }

  const handleSaveConfig = async () => {
    saveConfigLoading.value = true
    try {
      await saveLegacyWuxinConfig({ ...configForm })
      ElMessage.success('配置已保存')
      await loadPrice()
    } finally {
      saveConfigLoading.value = false
    }
  }

  const handleQuerySchoolInfo = async () => {
    if (!orderForm.auth_code) {
      ElMessage.warning('请输入授权码')
      return
    }
    schoolLoading.value = true
    try {
      const result = await queryLegacyWuxinSchoolInfo(orderForm.auth_code)
      parseSchoolInfo(result)
      ElMessage.success('信息获取成功')
    } finally {
      schoolLoading.value = false
    }
  }

  const handleCreateOrder = async () => {
    if (!orderForm.auth_code || !orderForm.run_plan_code || !orderForm.fence_code) {
      ElMessage.warning('请补齐授权码、跑步计划和跑步区域')
      return
    }
    createLoading.value = true
    try {
      await createLegacyWuxinOrder(buildFormPayload(orderForm))
      ElMessage.success('下单成功')
      addVisible.value = false
      await loadOrders(1)
      await loadPrice()
    } finally {
      createLoading.value = false
    }
  }

  const openLogsDialog = (order: LegacyWuxinOrder) => {
    logsPagination.id = order.id
    logsPagination.page = 1
    logsVisible.value = true
    loadLogs(1)
  }

  const loadLogs = async (page = logsPagination.page) => {
    logsLoading.value = true
    logsPagination.page = page
    try {
      const result = await fetchLegacyWuxinRecords({
        id: logsPagination.id,
        limit: logsPagination.limit,
        page
      })
      logs.value = Array.isArray(result?.list) ? result.list : []
      logsPagination.total = Number(result?.total || logs.value.length)
    } finally {
      logsLoading.value = false
    }
  }

  const handleLogsSizeChange = (size: number) => {
    logsPagination.limit = size
    loadLogs(1)
  }

  const openEditDialog = async (order: LegacyWuxinOrder) => {
    editVisible.value = true
    editLoading.value = true
    editOrderId.value = order.id
    editOrderNumber.value = order.order_number
    try {
      const result = await fetchLegacyWuxinOrderConfig(order.id, order.order_number)
      const detail = result?.order || order
      Object.assign(editForm, defaultOrderForm(), {
        auth_code: detail.auth_code,
        fence_code: detail.fence_code || 'default',
        mark: detail.mark || '',
        order_num: detail.quantity || 1,
        run_meter: Number(detail.run_meter || 1),
        run_plan_code: detail.run_plan_code || '',
        run_type: Number(detail.run_type || 1),
        run_week: parseWeek(detail.run_week),
        start_date: detail.start_date || formatDate(new Date()),
        zone_name: detail.zone_name || '默认跑区'
      })
      const [start, end] = String(detail.run_time || '06:00-08:00').split('-')
      editStartTime.value = start || '06:00'
      editEndTime.value = end || '08:00'
      editPaceValue.value = Number(String(detail.run_speed || '6').replace(/[^\d]/g, '') || 6)
      parseSchoolInfo(result?.school_info || {}, editRunPlanOptions, editRunZoneOptions)
    } finally {
      editLoading.value = false
    }
  }

  const handleEditOrder = async () => {
    editLoading.value = true
    try {
      await editLegacyWuxinOrder(editOrderId.value, buildEditPayload())
      ElMessage.success('编辑成功')
      editVisible.value = false
      loadOrders(pagination.page)
    } finally {
      editLoading.value = false
    }
  }

  const openIncreaseDialog = (order: LegacyWuxinOrder) => {
    increaseForm.id = order.id
    increaseForm.order_number = order.order_number
    increaseForm.quantity = 1
    increaseVisible.value = true
  }

  const handleRowCommand = (command: string, order: LegacyWuxinOrder) => {
    if (command === 'increase') {
      openIncreaseDialog(order)
      return
    }
    if (command === 'reassign') {
      handleReassign(order)
      return
    }
    if (command === 'refund') {
      handleRefund(order)
    }
  }

  const handleIncreaseOrder = async () => {
    increaseLoading.value = true
    try {
      await increaseLegacyWuxinOrder(
        increaseForm.id,
        increaseForm.quantity,
        increaseForm.order_number
      )
      ElMessage.success('追加成功')
      increaseVisible.value = false
      loadOrders(pagination.page)
    } finally {
      increaseLoading.value = false
    }
  }

  const handleRefund = async (order: LegacyWuxinOrder) => {
    await ElMessageBox.confirm(`确认申请退款订单 #${order.id}？`, '申请退款', { type: 'warning' })
    await refundLegacyWuxinOrder(order.id, order.order_number)
    ElMessage.success('申请退款成功')
    loadOrders(pagination.page)
  }

  const handleReassign = async (order: LegacyWuxinOrder) => {
    await ElMessageBox.confirm(`确认重新分配订单 #${order.id}？`, '重新分配', { type: 'warning' })
    await reassignLegacyWuxinOrder(order.id, order.order_number)
    ElMessage.success('重新分配成功')
    loadOrders(pagination.page)
  }

  const handleSyncOrders = async () => {
    syncLoading.value = true
    try {
      const result = await syncLegacyWuxinOrders()
      ElMessage.success(`同步完成，更新 ${Number(result?.updated || 0)} 条`)
      loadOrders(pagination.page)
    } finally {
      syncLoading.value = false
    }
  }

  const getOrderStatusLabel = (value: number) =>
    (
      ({
        0: '待处理',
        1: '等待计划',
        2: '处理中',
        3: '已完成',
        4: '已退款',
        5: '错误'
      }) as Record<number, string>
    )[value] || '未知'

  const getOrderStatusType = (value: number) =>
    (
      ({
        0: 'info',
        1: 'warning',
        2: 'primary',
        3: 'success',
        4: 'danger',
        5: 'danger'
      }) as Record<number, 'danger' | 'info' | 'primary' | 'success' | 'warning'>
    )[value] || 'info'

  const getLogStatusLabel = (value: number) =>
    (({ 0: '待处理', 1: '执行中', 2: '跑步成功', 3: '跑步失败' }) as Record<number, string>)[
      value
    ] || '未知'

  const getLogStatusType = (value: number) =>
    (
      ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }) as Record<
        number,
        'danger' | 'info' | 'success' | 'warning'
      >
    )[value] || 'info'

  const parseWeek = (raw: string) => {
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed.map((item) => Number(item)) : [1, 2, 3, 4, 5]
    } catch {
      return [1, 2, 3, 4, 5]
    }
  }

  const formatWeek = (raw: string) =>
    parseWeek(raw)
      .map((item) => `周${'一二三四五六日'[item - 1]}`)
      .join('、')

  onMounted(async () => {
    await loadConfig()
    await loadPrice()
    await loadOrders(1)
  })
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .wuxin-search-bar {
    flex-shrink: 0;
  }

  .wuxin-table-card :deep(.el-card__body) {
    padding-top: 14px;
  }

  .wuxin-table-toolbar {
    @apply flex flex-wrap items-center gap-2;
  }

  .wuxin-table-title {
    @apply m-0 mr-2 whitespace-nowrap text-base font-semibold text-g-900;
    line-height: 32px;
  }

  .wuxin-order-form {
    @apply space-y-4;
  }

  .wuxin-auth-panel,
  .wuxin-form-section {
    @apply rounded-lg border border-[var(--art-card-border)] bg-box;
  }

  .wuxin-auth-panel {
    @apply grid gap-3 p-4 md:grid-cols-[1fr_auto] md:items-end;
  }

  .wuxin-auth-button {
    @apply md:min-w-24;
  }

  .wuxin-form-section {
    @apply space-y-4 p-4;
  }

  .wuxin-section-title {
    @apply flex flex-wrap items-center justify-between gap-2 border-b border-[var(--art-card-border)] pb-3;
  }

  .wuxin-section-title span {
    @apply text-sm font-semibold text-g-900;
  }

  .wuxin-section-title small {
    @apply text-xs font-normal text-g-500;
  }

  .wuxin-form-grid {
    @apply grid gap-4;
  }

  .wuxin-form-field label {
    @apply mb-2 flex items-center gap-2 text-sm font-medium text-g-800;
  }

  .wuxin-form-field label small {
    @apply text-xs font-normal text-g-500;
  }

  .wuxin-pace-input {
    @apply flex items-center gap-2;
  }

  .wuxin-pace-input span {
    @apply whitespace-nowrap rounded-md bg-g-100 px-3 text-xs text-g-500;
    line-height: 32px;
  }

  .wuxin-order-summary {
    @apply grid gap-3 rounded-md bg-g-100/60 px-4 py-3 text-sm md:grid-cols-3;
  }

  .wuxin-order-summary div {
    @apply flex min-w-0 items-center justify-between gap-3 md:block;
  }

  .wuxin-order-summary span {
    @apply text-g-500;
  }

  .wuxin-order-summary strong {
    @apply block truncate font-semibold text-g-900 md:mt-1;
  }

  .wuxin-row-actions {
    @apply flex items-center whitespace-nowrap;
  }

  .wuxin-table-cell {
    @apply min-w-0;
  }

  .wuxin-two-line {
    display: -webkit-box;
    overflow: hidden;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  :global(.wuxin-danger-action) {
    color: var(--el-color-danger) !important;
  }
</style>
