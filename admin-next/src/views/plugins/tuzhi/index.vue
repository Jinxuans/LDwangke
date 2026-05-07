<template>
  <div class="plugin-tuzhi-page art-full-height">
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
            <ElTag effect="plain">商品 {{ goods.length }} 个</ElTag>
            <ElTag type="success" effect="plain">正常 {{ normalCount }}</ElTag>
            <ElTag type="warning" effect="plain">已完成 {{ finishedCount }}</ElTag>
            <ElButton plain @click="handleSync">同步订单</ElButton>
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

    <ElDialog v-model="addVisible" title="新增凸知订单" width="860px">
      <div class="grid gap-5 lg:grid-cols-[1fr_300px]">
        <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">选择商品</p>
            <ElSelect v-model="addForm.goods_id" class="w-full" filterable placeholder="请选择商品">
              <ElOption
                v-for="item in goods"
                :key="item.id"
                :label="item.display_name || item.name || String(item.id)"
                :value="item.id"
              />
            </ElSelect>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">账号</p>
              <ElInput v-model="addForm.username" placeholder="请输入账号" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="addForm.password" placeholder="请输入密码" show-password />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">姓名</p>
              <ElInput v-model="addForm.nickname" placeholder="选填" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">学校</p>
              <ElInput v-model="addForm.school" placeholder="选填" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">岗位名称</p>
              <ElInput v-model="addForm.postname" placeholder="选填" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">截至日期</p>
              <ElDatePicker
                v-model="addForm.work_deadline"
                class="w-full"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="请选择截至日期"
              />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">地址</p>
            <ElInput v-model="addForm.address" placeholder="请输入地址" />
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">纬度</p>
              <ElInput v-model="addForm.address_lat" placeholder="选填" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">经度</p>
              <ElInput v-model="addForm.address_lng" placeholder="选填" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">上班打卡时间</p>
              <ElTimePicker
                v-model="addForm.work_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">下班打卡时间</p>
              <ElTimePicker
                v-model="addForm.off_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
              />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
            <ElCheckboxGroup v-model="addForm.work_days">
              <ElCheckbox v-for="item in weekOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">商品</span>
              <span class="max-w-[170px] truncate text-right font-medium text-g-900">
                {{ getGoodsName(addForm.goods_id) }}
              </span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">截至日期</span>
              <span class="font-medium text-g-900">{{ addForm.work_deadline || '未选择' }}</span>
            </div>
            <div class="mt-4">
              <p class="mb-2 text-xs font-medium text-g-500">息知推送地址</p>
              <ElInput v-model="addForm.xz_push_url" placeholder="可选" />
            </div>
          </div>

          <div class="rounded-custom-sm border-full-d bg-box p-4">
            <p class="mb-3 text-sm font-medium text-g-800">附加选项</p>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">跳过节假日</span>
                <ElSwitch v-model="addForm.holiday_status" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">开启下班打卡</span>
                <ElSwitch v-model="addForm.is_off_time" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">日报</span>
                <ElSwitch v-model="addForm.daily_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">周报</span>
                <ElSwitch v-model="addForm.weekly_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">月报</span>
                <ElSwitch v-model="addForm.monthly_report" :active-value="1" :inactive-value="0" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addLoading" @click="handleAdd">确认下单</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="editVisible" title="编辑订单" width="860px">
      <div class="grid gap-5 lg:grid-cols-[1fr_300px]">
        <div class="space-y-4">
          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">账号</p>
              <ElInput v-model="editForm.username" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="editForm.password" show-password />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">姓名</p>
              <ElInput v-model="editForm.nickname" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">学校</p>
              <ElInput v-model="editForm.school" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">岗位名称</p>
              <ElInput v-model="editForm.postname" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">截至日期</p>
              <ElDatePicker
                v-model="editForm.work_deadline"
                class="w-full"
                type="date"
                value-format="YYYY-MM-DD"
              />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">地址</p>
            <ElInput v-model="editForm.address" />
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">纬度</p>
              <ElInput v-model="editForm.address_lat" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">经度</p>
              <ElInput v-model="editForm.address_lng" />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">上班打卡时间</p>
              <ElTimePicker
                v-model="editForm.work_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">下班打卡时间</p>
              <ElTimePicker
                v-model="editForm.off_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
              />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
            <ElCheckboxGroup v-model="editForm.work_days">
              <ElCheckbox v-for="item in weekOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">商品</span>
              <span class="max-w-[170px] truncate text-right font-medium text-g-900">
                {{ getGoodsName(editForm.goods_id) }}
              </span>
            </div>
            <div class="mt-3 flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">截至日期</span>
              <span class="font-medium text-g-900">{{ editForm.work_deadline || '未选择' }}</span>
            </div>
            <div class="mt-4">
              <p class="mb-2 text-xs font-medium text-g-500">息知推送地址</p>
              <ElInput v-model="editForm.xz_push_url" placeholder="可选" />
            </div>
          </div>

          <div class="rounded-custom-sm border-full-d bg-box p-4">
            <p class="mb-3 text-sm font-medium text-g-800">附加选项</p>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">跳过节假日</span>
                <ElSwitch v-model="editForm.holiday_status" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">开启下班打卡</span>
                <ElSwitch v-model="editForm.is_off_time" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">日报</span>
                <ElSwitch v-model="editForm.daily_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">周报</span>
                <ElSwitch v-model="editForm.weekly_report" :active-value="1" :inactive-value="0" />
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-g-700">月报</span>
                <ElSwitch v-model="editForm.monthly_report" :active-value="1" :inactive-value="0" />
              </div>
            </div>
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

    <ElDialog v-model="signInfoVisible" title="签到信息" width="720px">
      <ElScrollbar max-height="460px">
        <div v-if="signInfoLoading" class="py-10 text-center text-sm text-g-500">签到信息加载中...</div>
        <div v-else-if="!signInfoData" class="py-10 text-center text-sm text-g-500">暂无数据</div>
        <pre
          v-else
          class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-xs leading-6 text-g-700 whitespace-pre-wrap break-all"
        >{{ JSON.stringify(signInfoData, null, 2) }}</pre>
      </ElScrollbar>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    checkInLegacyTuzhiOrder,
    checkOutLegacyTuzhiOrder,
    createLegacyTuzhiOrder,
    deleteLegacyTuzhiOrder,
    editLegacyTuzhiOrder,
    fetchLegacyTuzhiGoods,
    fetchLegacyTuzhiOrders,
    fetchLegacyTuzhiSignInfo,
    syncLegacyTuzhiOrders,
    type LegacyTuzhiGoodsItem
  } from '@/api/legacy/plugin-tuzhi'

  defineOptions({ name: 'PluginTuzhiPage' })

  const weekOptions = [
    { label: '周一', value: '1' },
    { label: '周二', value: '2' },
    { label: '周三', value: '3' },
    { label: '周四', value: '4' },
    { label: '周五', value: '5' },
    { label: '周六', value: '6' },
    { label: '周日', value: '7' }
  ]

  const loading = ref(false)
  const syncing = ref(false)
  const addVisible = ref(false)
  const addLoading = ref(false)
  const editVisible = ref(false)
  const editLoading = ref(false)
  const signInfoVisible = ref(false)
  const signInfoLoading = ref(false)

  const orders = ref<any[]>([])
  const goods = ref<LegacyTuzhiGoodsItem[]>([])
  const signInfoData = ref<any>(null)

  const pagination = reactive({
    page: 1,
    limit: 10,
    total: 0
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.limit,
    total: pagination.total
  }))

  const filters = reactive({
    keyword: ''
  })

  const addForm = reactive<Record<string, any>>({
    goods_id: null,
    username: '',
    password: '',
    nickname: '',
    school: '',
    postname: '',
    address: '',
    address_lat: '',
    address_lng: '',
    work_time: '08:30',
    off_time: '17:30',
    work_days: ['1', '2', '3', '4', '5'],
    work_deadline: '',
    holiday_status: 0,
    daily_report: 0,
    weekly_report: 0,
    monthly_report: 0,
    weekly_report_time: 1,
    monthly_report_time: 0,
    is_off_time: 1,
    xz_push_url: '',
    images: '',
    token: '',
    uuid: '',
    user_school_id: '',
    random_phone: ''
  })

  const editForm = reactive<Record<string, any>>({})

  const normalCount = computed(() => orders.value.filter((item) => Number(item.is_status) === 1).length)
  const finishedCount = computed(() => orders.value.filter((item) => Number(item.status) === 3).length)

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索账号或姓名'
      }
    }
  ])

  const { columns, columnChecks } = useTableColumns<any>(() => [
    {
      prop: 'username',
      label: '账号信息',
      minWidth: 200,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.username || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.nickname || '未填写姓名'),
          h('p', { class: 'text-xs text-g-500' }, row.school || row.postname || '未填写学校/岗位')
        ])
    },
    {
      prop: 'goods_id',
      label: '商品 / 天数',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-medium text-g-900' }, getGoodsName(row.goods_id)),
          h('p', { class: 'text-xs text-g-500' }, `已打 ${row.work_days_ok_num || 0} / ${row.work_days_num || 0}`)
        ])
    },
    {
      prop: 'work_time',
      label: '时间设置',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-800' }, `${row.work_time || '-'} - ${row.off_time || '-'}`),
          h('p', { class: 'text-xs text-g-500' }, `截止 ${row.work_deadline || '-'}`),
          h('p', { class: 'text-xs text-g-500' }, `周期 ${row.work_days || '-'}`)
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row) =>
        h('div', { class: 'space-y-2' }, [
          h(ElTag, { type: getMainStatusType(row.status), effect: 'plain' }, () => getMainStatusText(row.status)),
          h(ElTag, { type: getSignStatusType(row.is_status), effect: 'plain' }, () => getSignStatusText(row.is_status))
        ])
    },
    {
      prop: 'address',
      label: '地址 / 备注',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'line-clamp-2 text-sm text-g-700' }, row.address || '未填写地址'),
          h('p', { class: 'line-clamp-2 text-xs text-g-500' }, row.remark || '暂无备注')
        ])
    },
    {
      prop: 'operation',
      label: '操作',
      width: 360,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { size: 'small', onClick: () => openEditDialog(row) }, () => '编辑'),
          h(ElButton, { size: 'small', onClick: () => openSignInfoDialog(row) }, () => '签到信息'),
          h(ElButton, { size: 'small', type: 'success', plain: true, onClick: () => handleCheckIn(row) }, () => '上班'),
          h(ElButton, { size: 'small', type: 'primary', plain: true, onClick: () => handleCheckOut(row) }, () => '下班'),
          h(ElButton, { size: 'small', type: 'danger', plain: true, onClick: () => handleDelete(row) }, () => '删除')
        ])
    }
  ])

  const getGoodsName = (goodsId: number) =>
    goods.value.find((item) => Number(item.id) === Number(goodsId))?.display_name ||
    goods.value.find((item) => Number(item.id) === Number(goodsId))?.name ||
    `#${goodsId}`

  const getMainStatusText = (status: number | string) => {
    const value = Number(status)
    if (value === 0) return '正常'
    if (value === 1) return '打卡中'
    if (value === 2) return '关闭'
    if (value === 3) return '已完成'
    return '未知'
  }

  const getMainStatusType = (status: number | string) => {
    const value = Number(status)
    if (value === 0) return 'primary'
    if (value === 1) return 'warning'
    if (value === 2) return 'danger'
    if (value === 3) return 'success'
    return 'info'
  }

  const getSignStatusText = (status: number | string) => (Number(status) === 1 ? '打卡正常' : '打卡异常')
  const getSignStatusType = (status: number | string) => (Number(status) === 1 ? 'success' : 'danger')

  const setDefaultDeadline = () => {
    const next = new Date()
    next.setDate(next.getDate() + 30)
    addForm.work_deadline = next.toISOString().slice(0, 10)
  }

  const resetFilters = () => {
    filters.keyword = ''
    loadOrders(1)
  }

  const resetAddForm = () => {
    addForm.goods_id = null
    addForm.username = ''
    addForm.password = ''
    addForm.nickname = ''
    addForm.school = ''
    addForm.postname = ''
    addForm.address = ''
    addForm.address_lat = ''
    addForm.address_lng = ''
    addForm.work_time = '08:30'
    addForm.off_time = '17:30'
    addForm.work_days = ['1', '2', '3', '4', '5']
    addForm.holiday_status = 0
    addForm.daily_report = 0
    addForm.weekly_report = 0
    addForm.monthly_report = 0
    addForm.weekly_report_time = 1
    addForm.monthly_report_time = 0
    addForm.is_off_time = 1
    addForm.xz_push_url = ''
    addForm.images = ''
    addForm.token = ''
    addForm.uuid = ''
    addForm.user_school_id = ''
    addForm.random_phone = ''
    setDefaultDeadline()
  }

  const loadGoods = async () => {
    goods.value = (await fetchLegacyTuzhiGoods()) || []
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyTuzhiOrders({
        page: pagination.page,
        limit: pagination.limit,
        keyword: filters.keyword || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const openAddDialog = async () => {
    if (!goods.value.length) {
      await loadGoods()
    }
    resetAddForm()
    addVisible.value = true
  }

  const handleSearch = () => {
    loadOrders(1)
  }

  const handleSizeChange = (size: number) => {
    pagination.limit = size
    loadOrders(1)
  }

  const handleAdd = async () => {
    if (!addForm.goods_id || !addForm.username || !addForm.password || !addForm.work_deadline) {
      ElMessage.warning('请填写商品、账号、密码和截至日期')
      return
    }
    addLoading.value = true
    try {
      await createLegacyTuzhiOrder({
        ...addForm,
        work_days: Array.isArray(addForm.work_days) ? addForm.work_days.join(',') : addForm.work_days
      })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openEditDialog = (record: any) => {
    Object.assign(editForm, {
      id: record.id,
      goods_id: Number(record.goods_id),
      username: record.username,
      password: record.password,
      nickname: record.nickname || '',
      school: record.school || '',
      postname: record.postname || '',
      address: record.address || '',
      address_lat: record.address_lat || '',
      address_lng: record.address_lng || '',
      work_time: record.work_time || '08:30',
      off_time: record.off_time || '17:30',
      work_days: record.work_days ? String(record.work_days).split(',') : ['1', '2', '3', '4', '5'],
      work_deadline: record.work_deadline || '',
      holiday_status: Number(record.holiday_status) || 0,
      daily_report: Number(record.daily_report) || 0,
      weekly_report: Number(record.weekly_report) || 0,
      monthly_report: Number(record.monthly_report) || 0,
      weekly_report_time: Number(record.weekly_report_time) || 1,
      monthly_report_time: Number(record.monthly_report_time) || 0,
      is_off_time: Number(record.is_off_time) ?? 1,
      xz_push_url: record.xz_push_url || '',
      images: record.images || '',
      token: record.token || '',
      uuid: record.uuid || '',
      user_school_id: record.user_school_id || '',
      random_phone: record.random_phone || ''
    })
    editVisible.value = true
  }

  const handleEdit = async () => {
    if (!editForm.work_deadline) {
      ElMessage.warning('截至日期不能为空')
      return
    }
    editLoading.value = true
    try {
      await editLegacyTuzhiOrder({
        ...editForm,
        work_days: Array.isArray(editForm.work_days) ? editForm.work_days.join(',') : editForm.work_days
      })
      ElMessage.success('修改成功')
      editVisible.value = false
      loadOrders(pagination.page)
    } finally {
      editLoading.value = false
    }
  }

  const handleDelete = async (record: any) => {
    await ElMessageBox.confirm(`确认删除订单 ${record.username}？`, '删除订单', {
      type: 'warning'
    })
    await deleteLegacyTuzhiOrder(record.id)
    ElMessage.success('删除成功')
    loadOrders(pagination.page)
  }

  const handleCheckIn = async (record: any) => {
    await checkInLegacyTuzhiOrder(record.id)
    ElMessage.success('上班打卡成功')
  }

  const handleCheckOut = async (record: any) => {
    await checkOutLegacyTuzhiOrder(record.id)
    ElMessage.success('下班打卡成功')
  }

  const handleSync = async () => {
    syncing.value = true
    try {
      const result = await syncLegacyTuzhiOrders()
      ElMessage.success(result?.msg || '同步完成')
      loadOrders(pagination.page)
    } finally {
      syncing.value = false
    }
  }

  const openSignInfoDialog = async (record: any) => {
    signInfoVisible.value = true
    signInfoLoading.value = true
    signInfoData.value = null
    try {
      signInfoData.value = await fetchLegacyTuzhiSignInfo({
        goods_id: record.goods_id,
        username: record.username,
        password: record.password
      })
    } finally {
      signInfoLoading.value = false
    }
  }

  onMounted(async () => {
    await loadGoods()
    resetAddForm()
    await loadOrders(1)
  })
</script>
