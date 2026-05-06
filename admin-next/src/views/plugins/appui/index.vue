<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <div class="grid gap-4 md:grid-cols-[140px_1fr_auto]">
        <ElSelect v-model="search.type" placeholder="搜索项">
          <ElOption v-for="item in searchTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </ElSelect>
        <ElInput v-model="search.keyword" clearable placeholder="输入关键词" @keyup.enter="loadOrders(1)" />
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
            <ElTag effect="plain">平台 {{ courses.length }} 个</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag type="warning" effect="plain">进行中 {{ processingCount }}</ElTag>
            <ElButton type="primary" plain @click="openAddDialog">新增订单</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="orders" size="large">
        <ElTableColumn label="平台" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-semibold text-g-900">{{ getCourseName(row.pid) }}</p>
              <p class="text-xs text-g-500">PID {{ row.pid }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="账号信息" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="font-semibold text-g-900">{{ row.user }}</p>
              <p class="text-xs text-g-500">密码 {{ row.pass || '-' }}</p>
              <p class="text-xs text-g-500">{{ row.name || row.school || '未填写姓名/学校' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="天数 / 状态" min-width="170">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm text-g-800">剩余 {{ row.residue_day }} / 总计 {{ row.total_day }}</p>
              <ElTag :type="getStatusType(row.status)" effect="plain">{{ row.status }}</ElTag>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="工作时间" min-width="180">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm text-g-800">{{ row.shangban_time || '-' }} - {{ row.xiaban_time || '-' }}</p>
              <p class="text-xs text-g-500">周期 {{ row.week || '-' }}</p>
              <p class="text-xs text-g-500">报表 {{ row.report || '-' }}</p>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="签到地址" min-width="220">
          <template #default="{ row }">
            <span class="line-clamp-2 text-sm text-g-500">{{ row.address || '未填写地址' }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="下单时间" min-width="160">
          <template #default="{ row }">
            <span class="text-sm text-g-500">{{ row.addtime || '-' }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <ElButton size="small" @click="openEditDialog(row)">编辑</ElButton>
              <ElButton size="small" @click="openRenewDialog(row)">续费</ElButton>
              <ElButton size="small" type="danger" plain @click="handleDelete(row)">退款</ElButton>
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

    <ElDialog v-model="addVisible" title="新增 Appui 订单" width="760px">
      <div class="grid gap-5 lg:grid-cols-[1fr_280px]">
        <div class="space-y-4">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">选择平台</p>
            <ElSelect v-model="addForm.pid" class="w-full" filterable placeholder="请选择平台" @change="calcAddPrice">
              <ElOption v-for="item in courses" :key="item.pid" :label="item.name" :value="item.pid" />
            </ElSelect>
          </div>

          <ElAlert
            v-if="courseTip"
            :title="courseTip"
            type="warning"
            show-icon
            :closable="false"
          />

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">用户账号</p>
              <ElInput v-model="addForm.user" placeholder="请输入账号" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">用户密码</p>
              <ElInput v-model="addForm.pass" placeholder="请输入密码" show-password />
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">用户姓名</p>
              <ElInput v-model="addForm.userName" placeholder="选填" />
            </div>
            <div v-if="showSchoolInput">
              <p class="mb-2 text-sm font-medium text-g-800">用户学校</p>
              <ElInput v-model="addForm.school" placeholder="请输入学校" />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">签到地址</p>
            <ElInput v-model="addForm.address" type="textarea" :rows="2" placeholder="请输入签到地址" />
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">上班时间</p>
              <ElTimePicker
                v-model="addForm.shangban_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="上班时间"
              />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">下班时间</p>
              <ElTimePicker
                v-model="addForm.xiaban_time"
                class="w-full"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="下班时间"
              />
            </div>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
            <ElCheckboxGroup v-model="addForm.week">
              <ElCheckbox v-for="item in weekOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">报表选择</p>
            <ElCheckboxGroup v-model="addForm.report">
              <ElCheckbox v-for="item in reportOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <div>
            <p class="mb-2 text-sm font-medium text-g-800">下单天数</p>
            <ElInputNumber v-model="addForm.days" class="w-full" :min="1" :max="365" @change="calcAddPrice" />
          </div>
        </div>

        <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">当前平台</span>
            <span class="max-w-[160px] truncate text-right font-medium text-g-900">
              {{ selectedCourse?.name || '未选择平台' }}
            </span>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">单价</span>
            <span class="font-medium text-g-900">
              {{ selectedCourse ? `¥${Number(selectedCourse.price || 0).toFixed(2)}/天` : '-' }}
            </span>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">下单天数</span>
            <span class="font-medium text-g-900">{{ addForm.days || 0 }} 天</span>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">预估总价</span>
            <span class="font-semibold text-[var(--el-color-danger)]">¥{{ addPrice.toFixed(2) }}</span>
          </div>
          <p class="mt-3 text-sm text-g-500">实际金额以后端计算结果为准。</p>
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
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">用户密码</p>
          <ElInput v-model="editForm.pass" placeholder="请输入密码" show-password />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">签到地址</p>
          <ElInput v-model="editForm.address" type="textarea" :rows="2" placeholder="请输入签到地址" />
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">上班时间</p>
            <ElTimePicker
              v-model="editForm.shangban_time"
              class="w-full"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
          <div>
            <p class="mb-2 text-sm font-medium text-g-800">下班时间</p>
            <ElTimePicker
              v-model="editForm.xiaban_time"
              class="w-full"
              format="HH:mm"
              value-format="HH:mm"
            />
          </div>
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">打卡周期</p>
          <ElCheckboxGroup v-model="editForm.week">
            <ElCheckbox v-for="item in weekOptions" :key="item.value" :value="item.value">
              {{ item.label }}
            </ElCheckbox>
          </ElCheckboxGroup>
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">报表选择</p>
          <ElCheckboxGroup v-model="editForm.report">
            <ElCheckbox v-for="item in reportOptions" :key="item.value" :value="item.value">
              {{ item.label }}
            </ElCheckbox>
          </ElCheckboxGroup>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="editLoading" @click="handleEdit">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="renewVisible" title="续费订单" width="420px">
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">订单</span>
            <span class="font-medium text-g-900">#{{ renewForm.id || '-' }}</span>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">当前平台</span>
            <span class="max-w-[180px] truncate text-right font-medium text-g-900">
              {{ getCourseName(renewForm.pid) }}
            </span>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3 text-sm">
            <span class="text-g-500">续费金额</span>
            <span class="font-semibold text-[var(--el-color-danger)]">¥{{ renewPrice.toFixed(2) }}</span>
          </div>
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">续费天数</p>
          <ElInputNumber v-model="renewForm.days" class="w-full" :min="1" :max="365" @change="calcRenewPrice" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="renewVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="renewLoading" @click="handleRenew">确认续费</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createLegacyAppuiOrder,
    deleteLegacyAppuiOrder,
    editLegacyAppuiOrder,
    fetchLegacyAppuiCourses,
    fetchLegacyAppuiOrders,
    fetchLegacyAppuiPrice,
    renewLegacyAppuiOrder,
    type LegacyAppuiCourse,
    type LegacyAppuiOrder
  } from '@/api/legacy/plugin-appui'

  defineOptions({ name: 'PluginAppuiPage' })

  const weekOptions = [
    { label: '周一', value: '1' },
    { label: '周二', value: '2' },
    { label: '周三', value: '3' },
    { label: '周四', value: '4' },
    { label: '周五', value: '5' },
    { label: '周六', value: '6' },
    { label: '周日', value: '7' }
  ]

  const reportOptions = [
    { label: '日报', value: 'daily' },
    { label: '周报', value: 'weekly' },
    { label: '月报', value: 'monthly' }
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
  const editVisible = ref(false)
  const editLoading = ref(false)
  const renewVisible = ref(false)
  const renewLoading = ref(false)

  const orders = ref<LegacyAppuiOrder[]>([])
  const courses = ref<LegacyAppuiCourse[]>([])
  const addPrice = ref(0)
  const renewPrice = ref(0)

  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const search = reactive({
    type: '1',
    keyword: ''
  })

  const addForm = reactive({
    pid: '',
    week: [] as string[],
    report: [] as string[],
    shangban_time: '',
    xiaban_time: '',
    school: '',
    user: '',
    pass: '',
    userName: '',
    address: '',
    days: 30
  })

  const editForm = reactive({
    id: 0,
    week: [] as string[],
    report: [] as string[],
    shangban_time: '',
    xiaban_time: '',
    pass: '',
    address: ''
  })

  const renewForm = reactive({
    id: 0,
    pid: '',
    days: 30
  })

  const selectedCourse = computed(() => courses.value.find((item) => item.pid === addForm.pid) || null)
  const courseTip = computed(() => selectedCourse.value?.content || '')
  const showSchoolInput = computed(() => selectedCourse.value?.yes_school === 1)
  const processingCount = computed(() => orders.value.filter((item) => item.status === '进行中').length)

  const getCourseName = (pid: string) => courses.value.find((item) => item.pid === pid)?.name || pid

  const getStatusType = (status: string) => {
    if (status === '进行中') return 'success'
    if (status === '已完成') return 'primary'
    if (status === '待处理') return 'warning'
    if (status === '已退款') return 'info'
    return 'danger'
  }

  const resetFilters = () => {
    search.type = '1'
    search.keyword = ''
    loadOrders(1)
  }

  const resetAddForm = () => {
    addForm.pid = ''
    addForm.week = []
    addForm.report = []
    addForm.shangban_time = ''
    addForm.xiaban_time = ''
    addForm.school = ''
    addForm.user = ''
    addForm.pass = ''
    addForm.userName = ''
    addForm.address = ''
    addForm.days = 30
    addPrice.value = 0
  }

  const loadCourses = async () => {
    courses.value = (await fetchLegacyAppuiCourses()) || []
  }

  const loadOrders = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyAppuiOrders({
        page: pagination.page,
        limit: pagination.limit,
        searchType: search.keyword ? search.type : undefined,
        keyword: search.keyword || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const calcAddPrice = async () => {
    if (!addForm.pid || addForm.days < 1) {
      addPrice.value = 0
      return
    }
    const result = await fetchLegacyAppuiPrice(addForm.pid, addForm.days)
    addPrice.value = Number(result?.price || 0)
  }

  const calcRenewPrice = async () => {
    if (!renewForm.pid || renewForm.days < 1) {
      renewPrice.value = 0
      return
    }
    const result = await fetchLegacyAppuiPrice(renewForm.pid, renewForm.days)
    renewPrice.value = Number(result?.price || 0)
  }

  const openAddDialog = async () => {
    if (!courses.value.length) {
      await loadCourses()
    }
    resetAddForm()
    addVisible.value = true
  }

  const handleAdd = async () => {
    if (!addForm.pid || !addForm.user || !addForm.pass || addForm.days < 1) {
      ElMessage.warning('请填写平台、账号、密码和天数')
      return
    }
    addLoading.value = true
    try {
      await createLegacyAppuiOrder({
        pid: addForm.pid,
        user: addForm.user,
        pass: addForm.pass,
        userName: addForm.userName,
        address: addForm.address,
        days: addForm.days,
        week: addForm.week.join(','),
        report: addForm.report.join(','),
        shangban_time: addForm.shangban_time,
        xiaban_time: addForm.xiaban_time,
        school: addForm.school
      })
      ElMessage.success('下单成功')
      addVisible.value = false
      loadOrders(1)
    } finally {
      addLoading.value = false
    }
  }

  const openEditDialog = (record: LegacyAppuiOrder) => {
    editForm.id = record.id
    editForm.pass = record.pass
    editForm.address = record.address
    editForm.week = record.week ? record.week.split(',') : []
    editForm.report = record.report ? record.report.split(',') : []
    editForm.shangban_time = record.shangban_time
    editForm.xiaban_time = record.xiaban_time
    editVisible.value = true
  }

  const handleEdit = async () => {
    editLoading.value = true
    try {
      await editLegacyAppuiOrder({
        id: editForm.id,
        pass: editForm.pass,
        address: editForm.address,
        week: editForm.week.join(','),
        report: editForm.report.join(','),
        shangban_time: editForm.shangban_time,
        xiaban_time: editForm.xiaban_time
      })
      ElMessage.success('修改成功')
      editVisible.value = false
      loadOrders(pagination.page)
    } finally {
      editLoading.value = false
    }
  }

  const openRenewDialog = async (record: LegacyAppuiOrder) => {
    renewForm.id = record.id
    renewForm.pid = record.pid
    renewForm.days = 30
    renewPrice.value = 0
    renewVisible.value = true
    await calcRenewPrice()
  }

  const handleRenew = async () => {
    renewLoading.value = true
    try {
      await renewLegacyAppuiOrder(renewForm.id, renewForm.days)
      ElMessage.success('续费成功')
      renewVisible.value = false
      loadOrders(pagination.page)
    } finally {
      renewLoading.value = false
    }
  }

  const handleDelete = async (record: LegacyAppuiOrder) => {
    await ElMessageBox.confirm(`确认退款订单 #${record.id}（${record.user}）？`, '退款订单', {
      type: 'warning'
    })
    await deleteLegacyAppuiOrder(record.id)
    ElMessage.success('退款成功')
    loadOrders(pagination.page)
  }

  onMounted(async () => {
    await loadCourses()
    await loadOrders(1)
  })
</script>
