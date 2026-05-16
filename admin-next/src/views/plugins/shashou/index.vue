<template>
  <div class="shashou-page art-full-height">
    <section class="rounded-custom-sm border-full-d bg-box px-5 py-4">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-x-4 gap-y-1">
            <h2 class="m-0 text-lg font-semibold text-g-900">{{ remoteName }}</h2>
            <span class="text-sm text-g-500">{{ currentProjectSummary }}</span>
            <span v-if="isMaintenanceMode" class="text-sm font-medium text-danger">上游维护中</span>
          </div>
          <p class="mt-1 text-sm text-g-500">价格按当前项目预估，完成后按上游结果多退少补。</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <ElButton v-if="remoteNoticeHtml" plain @click="noticeDialogVisible = true">查看公告</ElButton>
          <ElButton plain :loading="ordersLoading" @click="loadOrders">刷新订单</ElButton>
          <ElButton type="primary" @click="openOrderDrawer">新建订单</ElButton>
        </div>
      </div>
    </section>

    <section class="mt-5 rounded-custom-sm border-full-d bg-box p-5">
      <div class="mb-4 grid gap-4 md:grid-cols-[1fr_160px_160px_auto]">
        <ElInput v-model="query.account" clearable placeholder="搜索账号" />
        <ElInput v-model="query.order_no" clearable placeholder="订单号" />
        <ElSelect v-model="query.status" clearable placeholder="全部状态">
          <ElOption label="待处理" value="pending" />
          <ElOption label="处理中" value="processing" />
          <ElOption label="已完成" value="completed" />
          <ElOption label="已退款" value="refunded" />
          <ElOption label="失败" value="failed" />
        </ElSelect>
        <ElButton type="primary" :loading="ordersLoading" @click="loadOrders">查询</ElButton>
      </div>

      <ElTabs v-model="activeTab">
        <ElTabPane label="订单列表" name="orders">
          <ElTable :data="orders" border v-loading="ordersLoading">
            <ElTableColumn type="expand">
              <template #default="{ row }">
                <ElTable :data="row.account_details || []" border size="small">
                  <ElTableColumn prop="account" label="账号" min-width="130" />
                  <ElTableColumn prop="password" label="密码" min-width="110" />
                  <ElTableColumn prop="distance" label="公里" width="90" />
                  <ElTableColumn label="时间段" width="150">
                    <template #default="{ row: acc }">
                      {{ formatTime(acc.start_hour, acc.start_minute) }} - {{ formatTime(acc.end_hour, acc.end_minute) }}
                    </template>
                  </ElTableColumn>
                  <ElTableColumn label="周期" width="120">
                    <template #default="{ row: acc }">{{ formatRunDays(acc.run_days) }}</template>
                  </ElTableColumn>
                  <ElTableColumn label="状态" width="110">
                    <template #default="{ row: acc }">
                      <ElTag :type="statusType(acc.status)" effect="plain">{{ statusLabel(acc.status) }}</ElTag>
                    </template>
                  </ElTableColumn>
                  <ElTableColumn label="操作" width="120">
                    <template #default="{ row: acc }">
                      <ElButton v-if="canRefund(acc)" link type="danger" @click="refundAccount(acc)">退单</ElButton>
                    </template>
                  </ElTableColumn>
                </ElTable>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="id" label="ID" width="80" />
            <ElTableColumn prop="order_no" label="订单号" min-width="170" />
            <ElTableColumn label="类型" width="110">
              <template #default="{ row }">{{ orderTypeLabel(row.order_type) }}</template>
            </ElTableColumn>
            <ElTableColumn label="数量/公里" width="130">
              <template #default="{ row }">{{ row.account_count }} / {{ row.total_distance }}</template>
            </ElTableColumn>
            <ElTableColumn prop="pre_deduct" label="预扣" width="100" />
            <ElTableColumn label="结算" width="150">
              <template #default="{ row }">
                <span>{{ settlementLabel(row) }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="110">
              <template #default="{ row }">
                <ElTag :type="statusType(row.status)" effect="plain">{{ statusLabel(row.status) }}</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="created_at" label="创建时间" width="170" />
            <ElTableColumn label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <ElButton link type="primary" @click="syncOrder(row.id)">同步</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
        </ElTabPane>
        <ElTabPane label="查单" name="query">
          <div class="max-w-[720px] rounded-custom-sm border-full-d p-5">
            <div class="grid gap-4 md:grid-cols-3">
              <ElSelect v-model="singleQuery.project_id" placeholder="项目">
                <ElOption v-for="item in projects" :key="item.id" :label="item.name" :value="item.id" />
              </ElSelect>
              <ElSelect v-model="singleQuery.query_type">
                <ElOption label="查询课外跑" :value="3" />
                <ElOption label="查询晨跑" :value="5" />
              </ElSelect>
              <ElInput v-model="singleQuery.account" placeholder="账号" />
            </div>
            <div class="mt-4 flex justify-end">
              <ElButton type="primary" :loading="queryLoading" @click="submitQuery">提交查单</ElButton>
            </div>
          </div>
        </ElTabPane>
      </ElTabs>
    </section>

    <ElDrawer v-model="orderDrawerVisible" title="新建订单" direction="rtl" size="min(900px, 100vw)" class="shashou-order-drawer">
      <div class="drawer-order-body">
        <section class="rounded-custom-sm border-full-d bg-box p-4">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_240px]">
            <div class="grid gap-4 md:grid-cols-3">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">项目</label>
                <ElSelect v-model="orderForm.project_id" class="w-full" filterable placeholder="请选择项目">
                  <ElOption
                    v-for="item in projects"
                    :key="item.id"
                    :label="`${item.name}（课外 ${item.price_normal}/km，晨跑 ${item.price_morning}/km）`"
                    :value="item.id"
                  />
                </ElSelect>
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">订单类型</label>
                <ElSegmented
                  v-model="orderForm.order_type"
                  :options="[
                    { label: '课外跑', value: 1 },
                    { label: '晨跑', value: 2 }
                  ]"
                  class="w-full"
                />
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">抢单模式</label>
                <div class="flex min-h-8 items-center justify-between gap-3 rounded-custom-sm border-full-d px-3 py-2">
                  <span class="text-sm text-g-700">{{ orderForm.is_rush_order ? `加收 ¥${currentProject?.rush_fee || 0}` : '普通队列' }}</span>
                  <ElSwitch v-model="orderForm.is_rush_order" />
                </div>
              </div>

              <p class="m-0 text-sm leading-6 text-g-500 md:col-span-3">{{ modeDesc }}</p>
            </div>

            <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="flex items-center justify-between text-sm">
                <span class="text-g-500">账号数量</span>
                <span class="font-medium text-g-900">{{ orderForm.accounts.length }}</span>
              </div>
              <div class="mt-2 flex items-center justify-between text-sm">
                <span class="text-g-500">总公里数</span>
                <span class="font-medium text-g-900">{{ totalDistance.toFixed(2) }} km</span>
              </div>
              <div class="mt-2 flex items-center justify-between text-sm">
                <span class="text-g-500">预计费用</span>
                <span class="text-lg font-semibold text-primary">¥{{ estimatedAmount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </section>

        <section class="mt-4 rounded-custom-sm border-full-d bg-box p-4">
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="m-0 text-base font-semibold text-g-900">账号明细</h3>
              <p class="mt-1 text-sm text-g-500">同一账号请勿在未完成订单中重复提交。</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <ElButton plain @click="clearAccounts">清空</ElButton>
              <ElButton type="primary" plain @click="addAccount">添加账号</ElButton>
            </div>
          </div>

          <ElTable :data="orderForm.accounts" border>
            <ElTableColumn label="账号" min-width="150">
              <template #default="{ row }">
                <ElInput v-model="row.account" placeholder="账号" />
              </template>
            </ElTableColumn>
            <ElTableColumn label="密码" min-width="130">
              <template #default="{ row }">
                <ElInput v-model="row.password" show-password placeholder="密码" />
              </template>
            </ElTableColumn>
            <ElTableColumn label="公里" width="120">
              <template #default="{ row }">
                <ElInputNumber v-model="row.distance" class="w-full" :min="0.1" :precision="2" :step="0.1" />
              </template>
            </ElTableColumn>
            <ElTableColumn label="开始" width="130">
              <template #default="{ row }">
                <ElTimePicker
                  :model-value="timeValue(row.start_hour, row.start_minute)"
                  class="w-full"
                  format="HH:mm"
                  value-format="HH:mm"
                  @update:model-value="setTime(row, 'start', $event)"
                />
              </template>
            </ElTableColumn>
            <ElTableColumn label="结束" width="130">
              <template #default="{ row }">
                <ElTimePicker
                  :model-value="timeValue(row.end_hour, row.end_minute)"
                  class="w-full"
                  format="HH:mm"
                  value-format="HH:mm"
                  @update:model-value="setTime(row, 'end', $event)"
                />
              </template>
            </ElTableColumn>
            <ElTableColumn label="周期" min-width="180">
              <template #default="{ row }">
                <ElSelect
                  :model-value="runDaysArray(row.run_days)"
                  class="w-full"
                  multiple
                  collapse-tags
                  @update:model-value="setRunDays(row, $event)"
                >
                  <ElOption v-for="item in weekOptions" :key="item.value" :label="item.label" :value="item.value" />
                </ElSelect>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="90" fixed="right">
              <template #default="{ $index }">
                <ElButton link type="danger" @click="removeAccount($index)">删除</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
        </section>
      </div>

      <template #footer>
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm text-g-500">预计费用 ¥{{ estimatedAmount.toFixed(2) }}</span>
          <div class="flex gap-2">
            <ElButton @click="orderDrawerVisible = false">取消</ElButton>
            <ElButton type="primary" :loading="submitLoading" @click="submitOrder">提交订单</ElButton>
          </div>
        </div>
      </template>
    </ElDrawer>

    <ElDialog v-model="noticeDialogVisible" width="760px" class="shashou-notice-dialog" append-to-body>
      <div class="remote-notice" v-html="remoteNoticeHtml"></div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createShashouOrder,
    fetchShashouOrders,
    fetchShashouProjects,
    fetchShashouVersionInfo,
    queryShashouAccount,
    refundShashouAccount,
    syncShashouOrder,
    type ShashouAccount,
    type ShashouAccountForm,
    type ShashouOrder,
    type ShashouProject,
    type ShashouVersionInfo
  } from '@/api/legacy/plugin-shashou'

  defineOptions({ name: 'ShashouIndexPage' })

  const projects = ref<ShashouProject[]>([])
  const orders = ref<ShashouOrder[]>([])
  const versionInfo = ref<ShashouVersionInfo | null>(null)
  const submitLoading = ref(false)
  const ordersLoading = ref(false)
  const queryLoading = ref(false)
  const activeTab = ref('orders')
  const noticeDialogVisible = ref(false)
  const orderDrawerVisible = ref(false)
  const query = reactive({ account: '', order_no: '', status: '', page: 1, limit: 20 })
  const singleQuery = reactive({ project_id: 0, query_type: 3, account: '' })

  const defaultAccount = (): ShashouAccountForm => ({
    account: '',
    password: '',
    distance: 1,
    start_hour: 7,
    start_minute: 0,
    end_hour: 22,
    end_minute: 0,
    run_days: '12345'
  })

  const orderForm = reactive({
    project_id: 0,
    order_type: 1,
    is_rush_order: false,
    accounts: [defaultAccount()]
  })

  const weekOptions = [
    { label: '周一', value: '1' },
    { label: '周二', value: '2' },
    { label: '周三', value: '3' },
    { label: '周四', value: '4' },
    { label: '周五', value: '5' },
    { label: '周六', value: '6' },
    { label: '周日', value: '7' }
  ]

  const currentProject = computed(() => projects.value.find((item) => item.id === orderForm.project_id))
  const remoteName = computed(() => versionInfo.value?.name || '鲨兽运动世界')
  const remoteNoticeHtml = computed(() => sanitizeNoticeHtml(versionInfo.value?.home_notice || ''))
  const isMaintenanceMode = computed(() => Number(versionInfo.value?.status ?? 1) === 0)
  const currentProjectSummary = computed(() => {
    const project = currentProject.value
    if (!project) return '请选择项目'
    return `课外 ¥${Number(project.price_normal || 0).toFixed(2)}/km · 晨跑 ¥${Number(project.price_morning || 0).toFixed(2)}/km`
  })
  const totalDistance = computed(() => orderForm.accounts.reduce((sum, item) => sum + Number(item.distance || 0), 0))
  const estimatedAmount = computed(() => {
    const project = currentProject.value
    if (!project) return 0
    const unit = orderForm.order_type === 2 ? project.price_morning : project.price_normal
    return totalDistance.value * Number(unit || 0) + (orderForm.is_rush_order ? Number(project.rush_fee || 0) : 0)
  })
  const modeDesc = computed(() =>
    orderForm.is_rush_order
      ? `抢单模式会额外加收 ¥${currentProject.value?.rush_fee || 0}，用于优先处理；仍需保证账号、公里数和时间段准确。`
      : '普通模式按队列处理，通常适合不着急的订单；提交后请等待上游处理并及时同步状态。'
  )

  const loadProjects = async () => {
    projects.value = (await fetchShashouProjects()) || []
    if (!orderForm.project_id && projects.value.length) {
      orderForm.project_id = projects.value[0].id
      singleQuery.project_id = projects.value[0].id
    }
  }

  const loadVersionInfo = async () => {
    try {
      versionInfo.value = await fetchShashouVersionInfo()
    } catch {
      versionInfo.value = null
    }
  }

  const loadOrders = async () => {
    ordersLoading.value = true
    try {
      const res = await fetchShashouOrders(query)
      orders.value = res?.list || []
    } finally {
      ordersLoading.value = false
    }
  }

  const addAccount = () => orderForm.accounts.push(defaultAccount())
  const removeAccount = (index: number) => orderForm.accounts.splice(index, 1)
  const clearAccounts = () => {
    orderForm.accounts.splice(0, orderForm.accounts.length, defaultAccount())
  }
  const openOrderDrawer = () => {
    if (!orderForm.accounts.length) clearAccounts()
    orderDrawerVisible.value = true
  }

  const submitOrder = async () => {
    if (!orderForm.project_id) return ElMessage.warning('请选择项目')
    submitLoading.value = true
    try {
      const result = await createShashouOrder({ ...orderForm })
      ElMessage.success(`提交成功，订单号：${result.order_no}`)
      clearAccounts()
      orderDrawerVisible.value = false
      await loadOrders()
    } finally {
      submitLoading.value = false
    }
  }

  const submitQuery = async () => {
    if (!singleQuery.project_id || !singleQuery.account.trim()) return ElMessage.warning('请填写项目和账号')
    queryLoading.value = true
    try {
      await queryShashouAccount({ ...singleQuery })
      ElMessage.success('查单已提交')
      activeTab.value = 'orders'
      await loadOrders()
    } finally {
      queryLoading.value = false
    }
  }

  const syncOrder = async (id: number) => {
    await syncShashouOrder(id)
    ElMessage.success('同步完成')
    await loadOrders()
  }

  const refundAccount = async (account: ShashouAccount) => {
    await ElMessageBox.confirm(`确定提交账号 ${account.account} 的退单请求吗？`, '退单确认', { type: 'warning' })
    await refundShashouAccount({ account_id: account.id })
    ElMessage.success('退单请求已提交')
    await loadOrders()
  }

  const canRefund = (account: ShashouAccount) => ['success', 'completed'].includes(account.status)
  const timeValue = (hour: number, minute: number) => `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
  const setTime = (row: ShashouAccountForm, field: 'start' | 'end', value: string) => {
    const [hour, minute] = String(value || '00:00').split(':').map((item) => Number(item))
    if (field === 'start') {
      row.start_hour = hour
      row.start_minute = minute
    } else {
      row.end_hour = hour
      row.end_minute = minute
    }
  }
  const runDaysArray = (raw: string) => String(raw || '').split('').filter(Boolean)
  const setRunDays = (row: ShashouAccountForm, value: unknown) => {
    row.run_days = Array.isArray(value) ? value.map(String).join('') : ''
  }
  const formatTime = timeValue
  const formatRunDays = (raw: string) =>
    runDaysArray(raw)
      .map((value) => weekOptions.find((item) => item.value === value)?.label || value)
      .join('、')
  const orderTypeLabel = (value: number) => ({ 1: '课外跑', 2: '晨跑', 3: '查询课外跑', 4: '退款', 5: '查询晨跑' })[value] || '-'
  const settlementLabel = (row: ShashouOrder) => {
    if (row.final_charge !== null && row.final_charge !== undefined) return `实付 ¥${Number(row.final_charge).toFixed(2)}`
    if (row.actual_cost !== null && row.actual_cost !== undefined) return `实际 ¥${Number(row.actual_cost).toFixed(2)}`
    return '待结算'
  }
  const statusLabel = (value: string) =>
    ({ pending: '待处理', processing: '处理中', completed: '已完成', success: '成功', refunded: '已退款', refunding: '退款中', failed: '失败' })[
      value
    ] || value
  const statusType = (value: string) => {
    if (['completed', 'success'].includes(value)) return 'success'
    if (['failed', 'refunded'].includes(value)) return 'danger'
    if (['processing', 'refunding'].includes(value)) return 'warning'
    return 'info'
  }

  const sanitizeNoticeHtml = (raw: string) => {
    if (!raw || typeof DOMParser === 'undefined') return ''
    const parser = new DOMParser()
    const doc = parser.parseFromString(raw, 'text/html')
    doc.querySelectorAll('script, style, iframe, object, embed, link, meta, form, input, button').forEach((node) => node.remove())
    doc.body.querySelectorAll('*').forEach((node) => {
      Array.from(node.attributes).forEach((attr) => {
        const name = attr.name.toLowerCase()
        const value = attr.value.trim().toLowerCase()
        if (name.startsWith('on') || value.startsWith('javascript:')) {
          node.removeAttribute(attr.name)
        }
      })
    })
    return doc.body.innerHTML
  }

  onMounted(async () => {
    await loadVersionInfo()
    noticeDialogVisible.value = Boolean(remoteNoticeHtml.value)
    await loadProjects()
    await loadOrders()
  })
</script>

<style scoped>
  .drawer-order-body {
    min-height: 100%;
  }

  .remote-notice {
    max-height: 60vh;
    overflow: auto;
    color: var(--el-text-color-regular);
    font-size: 14px;
    line-height: 1.8;
  }

  .remote-notice :deep(p) {
    margin: 0 0 8px;
  }

  .remote-notice :deep(a) {
    color: var(--el-color-primary);
  }

  :global(.shashou-notice-dialog .el-dialog__header) {
    padding: 0;
  }

  :global(.shashou-notice-dialog .el-dialog__headerbtn) {
    right: 12px;
    top: 10px;
    z-index: 1;
  }

  :global(.shashou-order-drawer .el-drawer__header) {
    margin-bottom: 0;
    padding: 18px 20px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  :global(.shashou-order-drawer .el-drawer__body) {
    padding: 16px;
    background: var(--art-main-bg-color);
  }

  :global(.shashou-order-drawer .el-drawer__footer) {
    padding: 12px 16px;
    border-top: 1px solid var(--el-border-color-lighter);
  }
</style>
