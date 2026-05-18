<template>
  <div class="admin-shashou-page art-full-height">
    <section class="rounded-custom-sm border-full-d bg-box p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="m-0 text-lg font-semibold text-g-900">鲨兽运动世界</h2>
          <p class="mt-1 text-sm text-g-500">维护上游项目配置，处理订单同步和账号明细。</p>
        </div>
        <div class="flex flex-wrap gap-3">
          <ElButton plain @click="adminGuideVisible = true">对接说明</ElButton>
          <ElButton plain :loading="syncLoading" @click="syncPending">同步待处理</ElButton>
          <ElButton type="primary" @click="openProject()">新增项目</ElButton>
        </div>
      </div>
    </section>

    <ElAlert
      class="mt-5"
      type="info"
      show-icon
      :closable="false"
      title="配置鲨兽项目前，请确认对接类型、API 地址、密钥、用户 ID 与上游项目 ID；上游项目 ID 留 0 时默认使用本地项目 ID。"
    />

    <ElTabs v-model="activeTab" class="mt-5">
      <ElTabPane label="项目配置" name="projects">
        <ElTable :data="projects" border v-loading="projectsLoading">
          <ElTableColumn prop="id" label="ID" width="70" />
          <ElTableColumn prop="name" label="名称" min-width="150" />
          <ElTableColumn label="类型" width="110">
            <template #default="{ row }">
              <ElTag :type="projectTypeTag(row.type)" effect="plain">
                {{ projectTypeLabel(row.type) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="remote_project_id" label="上游项目ID" width="120" />
          <ElTableColumn prop="api_url" label="API 地址" min-width="220" />
          <ElTableColumn label="价格" min-width="180">
            <template #default="{ row }">
              课外 {{ row.price_normal }}/km，晨跑 {{ row.price_morning }}/km
            </template>
          </ElTableColumn>
          <ElTableColumn label="附加" min-width="170">
            <template #default="{ row }">
              抢单 {{ row.rush_fee }}，查单 {{ row.query_fee }}
            </template>
          </ElTableColumn>
          <ElTableColumn label="状态" width="90">
            <template #default="{ row }">
              <ElTag :type="row.status === 1 ? 'success' : 'info'" effect="plain">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <ElButton link type="primary" @click="openProject(row)">编辑</ElButton>
              <ElButton link type="danger" @click="deleteProject(row.id)">删除</ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <ElTabPane label="订单管理" name="orders">
        <div class="mb-4 grid gap-4 md:grid-cols-[1fr_160px_160px_auto]">
          <ElInput v-model="orderQuery.account" clearable placeholder="账号" />
          <ElInput v-model="orderQuery.order_no" clearable placeholder="订单号" />
          <ElSelect v-model="orderQuery.status" clearable placeholder="状态">
            <ElOption label="待处理" value="pending" />
            <ElOption label="处理中" value="processing" />
            <ElOption label="已完成" value="completed" />
            <ElOption label="已退款" value="refunded" />
            <ElOption label="失败" value="failed" />
          </ElSelect>
          <ElButton type="primary" :loading="ordersLoading" @click="loadOrders">查询</ElButton>
        </div>
        <ElTable :data="orders" border v-loading="ordersLoading">
          <ElTableColumn type="expand">
            <template #default="{ row }">
              <ElTable :data="row.account_details || []" border size="small">
                <ElTableColumn prop="account" label="账号" min-width="130" />
                <ElTableColumn prop="password" label="密码" min-width="110" />
                <ElTableColumn prop="distance" label="公里" width="90" />
                <ElTableColumn label="状态" width="110">
                  <template #default="{ row: acc }">
                    <ElTag :type="statusType(acc.status)" effect="plain">{{
                      statusLabel(acc.status)
                    }}</ElTag>
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="error_message" label="错误信息" min-width="180">
                  <template #default="{ row: acc }">{{ emptyText(acc.error_message) }}</template>
                </ElTableColumn>
                <ElTableColumn prop="processed_at" label="处理时间" width="170">
                  <template #default="{ row: acc }">{{ emptyText(acc.processed_at) }}</template>
                </ElTableColumn>
              </ElTable>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="id" label="ID" width="80" />
          <ElTableColumn prop="order_no" label="订单号" min-width="170" />
          <ElTableColumn prop="username" label="所属用户" min-width="130" />
          <ElTableColumn label="类型" width="110">
            <template #default="{ row }">{{ orderTypeLabel(row.order_type) }}</template>
          </ElTableColumn>
          <ElTableColumn label="预扣费" width="100">
            <template #default="{ row }">{{ moneyText(row.pre_deduct) }}</template>
          </ElTableColumn>
          <ElTableColumn label="实际费用" width="100">
            <template #default="{ row }">{{ moneyText(row.actual_cost) }}</template>
          </ElTableColumn>
          <ElTableColumn label="最终收费" width="110">
            <template #default="{ row }">{{ moneyText(row.final_charge) }}</template>
          </ElTableColumn>
          <ElTableColumn label="差价" width="100">
            <template #default="{ row }">
              <span :class="diffClass(row.difference)">{{ signedMoneyText(row.difference) }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="状态" width="110">
            <template #default="{ row }">
              <ElTag :type="statusType(row.status)" effect="plain">{{
                statusLabel(row.status)
              }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="结算状态" width="110">
            <template #default="{ row }">
              <ElTag :type="paymentType(row.payment_status)" effect="plain">{{
                paymentLabel(row.payment_status)
              }}</ElTag>
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

      <ElTabPane label="账号明细" name="accounts">
        <ElTable :data="accounts" border v-loading="accountsLoading">
          <ElTableColumn prop="id" label="ID" width="80" />
          <ElTableColumn prop="order_no" label="订单号" min-width="160" />
          <ElTableColumn prop="username" label="所属用户" min-width="130" />
          <ElTableColumn prop="account" label="账号" min-width="130" />
          <ElTableColumn prop="password" label="密码" min-width="110" />
          <ElTableColumn prop="distance" label="公里" width="90" />
          <ElTableColumn label="状态" width="110">
            <template #default="{ row }">
              <ElTag :type="statusType(row.status)" effect="plain">{{
                statusLabel(row.status)
              }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="error_message" label="错误信息" min-width="180">
            <template #default="{ row }">{{ emptyText(row.error_message) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="processed_at" label="处理时间" width="170">
            <template #default="{ row }">{{ emptyText(row.processed_at) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="created_at" label="创建时间" width="170" />
        </ElTable>
      </ElTabPane>
    </ElTabs>

    <ElDialog
      v-model="projectVisible"
      :title="projectForm.id ? '编辑鲨兽项目' : '新增鲨兽项目'"
      width="860px"
    >
      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">项目名称</label>
          <ElInput v-model="projectForm.name" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">对接类型</label>
          <ElSelect v-model="projectForm.type" class="w-full">
            <ElOption label="源台" :value="0" />
            <ElOption label="29二开" :value="1" />
            <ElOption label="同系统（本系统）" :value="2" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">上游项目ID</label>
          <ElInputNumber
            v-model="projectForm.remote_project_id"
            class="w-full"
            :min="0"
            :step="1"
          />
          <p class="mt-1 text-xs text-g-500"
            >留 0 时使用本地项目 ID；若上游项目编号不同，请填写真实上游 ID。</p
          >
        </div>
        <div class="md:col-span-2">
          <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
          <ElInput v-model="projectForm.api_url" placeholder="https://ssyd.cc" />
          <p class="mt-1 text-xs text-g-500">填写上游站点根地址即可，系统会自动拼接接口路径。</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">{{
            credentialLabel('key')
          }}</label>
          <ElInput v-model="projectForm.api_key" show-password />
          <p class="mt-1 text-xs text-g-500">{{ credentialHelp('key') }}</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">{{
            credentialLabel('uid')
          }}</label>
          <ElInput v-model="projectForm.user_id" />
          <p class="mt-1 text-xs text-g-500">{{ credentialHelp('uid') }}</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">课外跑单价</label>
          <ElInputNumber
            v-model="projectForm.price_normal"
            class="w-full"
            :min="0"
            :precision="2"
            :step="0.1"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">晨跑单价</label>
          <ElInputNumber
            v-model="projectForm.price_morning"
            class="w-full"
            :min="0"
            :precision="2"
            :step="0.1"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">实际费率</label>
          <ElInputNumber
            v-model="projectForm.actual_rate"
            class="w-full"
            :min="0"
            :precision="2"
            :step="0.01"
          />
          <p class="mt-1 text-xs text-g-500">用于记录上游实际成本倍率，不影响用户展示单价。</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">抢单服务费</label>
          <ElInputNumber
            v-model="projectForm.rush_fee"
            class="w-full"
            :min="0"
            :precision="2"
            :step="0.1"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">查单费用</label>
          <ElInputNumber
            v-model="projectForm.query_fee"
            class="w-full"
            :min="0"
            :precision="2"
            :step="0.1"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">最低余额</label>
          <ElInputNumber
            v-model="projectForm.min_balance"
            class="w-full"
            :min="0"
            :precision="2"
            :step="1"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
          <ElSwitch v-model="projectForm.status" :active-value="1" :inactive-value="0" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">自动同步</label>
          <ElSwitch v-model="projectForm.auto_sync" :active-value="1" :inactive-value="0" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">同步间隔(秒)</label>
          <ElInputNumber v-model="projectForm.sync_interval" class="w-full" :min="30" :step="30" />
          <p class="mt-1 text-xs text-g-500">建议 300 秒以上，避免频繁请求上游。</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">超时(秒)</label>
          <ElInputNumber v-model="projectForm.timeout" class="w-full" :min="5" :max="120" />
        </div>
        <div class="md:col-span-2">
          <label class="mb-2 block text-sm font-medium text-g-800">备注</label>
          <ElInput v-model="projectForm.remark" type="textarea" :rows="3" />
        </div>
      </div>
      <template #footer>
        <ElButton @click="projectVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="saving" @click="saveProject">保存</ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="adminGuideVisible" title="鲨兽运动对接说明" width="760px">
      <ElCollapse v-model="adminGuideNames">
        <ElCollapseItem title="对接类型" name="type">
          <ul class="guide-list">
            <li>源台模式：请求鲨兽源接口，使用 `X-API-Key` 和 `X-User-ID` 作为请求头。</li>
            <li>29二开模式：请求上游 `/ss_apis.php`，使用 `uid`、`key`、`act` 参数。</li>
            <li>同系统模式：请求上游 `/api/v1/open/shashou`，使用本系统 API 的 `uid` 和 `key`。</li>
            <li>API 地址填写上游站点根地址即可，不需要手动带接口文件名。</li>
          </ul>
        </ElCollapseItem>
        <ElCollapseItem title="项目与价格" name="project">
          <ul class="guide-list">
            <li>本地价格是用户下单预扣依据，课外跑和晨跑可分别设置。</li>
            <li>上游项目 ID 为 0 时使用本地项目 ID；若上游编号不同，需要填写真实上游 ID。</li>
            <li>抢单服务费和查单费用会参与用户端费用预估。</li>
          </ul>
        </ElCollapseItem>
        <ElCollapseItem title="同步与结算" name="sync">
          <ul class="guide-list">
            <li>自动同步只处理待处理、处理中、退款中的订单。</li>
            <li>同步到上游实际费用后，系统会按预扣差额执行多退少补。</li>
            <li>同步间隔过短可能造成上游限流，建议保持 300 秒或更长。</li>
          </ul>
        </ElCollapseItem>
        <ElCollapseItem title="兼容接口" name="compat">
          <ul class="guide-list">
            <li>已兼容 `/ss_apis.php` 和 `/shashou/api.php`。</li>
            <li>下游可继续按 29 系统方式传 `uid`、`key`、`act` 调用。</li>
            <li>本系统原生页面统一走 `/api/v1/shashou` 接口。</li>
          </ul>
        </ElCollapseItem>
      </ElCollapse>
      <template #footer>
        <ElButton type="primary" @click="adminGuideVisible = false">我知道了</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    deleteAdminShashouProject,
    fetchAdminShashouAccounts,
    fetchAdminShashouOrders,
    fetchAdminShashouProjects,
    saveAdminShashouProject,
    syncAdminShashouOrder,
    syncAdminShashouPending,
    type ShashouAccount,
    type ShashouOrder,
    type ShashouProject
  } from '@/api/legacy/plugin-shashou'

  defineOptions({ name: 'AdminShashouPage' })

  const activeTab = ref('projects')
  const projects = ref<ShashouProject[]>([])
  const orders = ref<ShashouOrder[]>([])
  const accounts = ref<ShashouAccount[]>([])
  const projectsLoading = ref(false)
  const ordersLoading = ref(false)
  const accountsLoading = ref(false)
  const syncLoading = ref(false)
  const saving = ref(false)
  const projectVisible = ref(false)
  const adminGuideVisible = ref(false)
  const adminGuideNames = ref(['type', 'project'])
  const orderQuery = reactive({ account: '', order_no: '', status: '', page: 1, limit: 30 })
  const projectForm = reactive<Partial<ShashouProject>>({})

  const defaultProject = (): Partial<ShashouProject> => ({
    name: '鲨兽运动世界',
    type: 0,
    remote_project_id: 0,
    api_url: 'https://ssyd.cc',
    api_key: '',
    user_id: '',
    price_normal: 9,
    price_morning: 10,
    actual_rate: 1.05,
    rush_fee: 3,
    query_fee: 1,
    min_balance: 0,
    status: 1,
    auto_sync: 1,
    sync_interval: 300,
    timeout: 30,
    remark: ''
  })

  const loadProjects = async () => {
    projectsLoading.value = true
    try {
      projects.value = (await fetchAdminShashouProjects()) || []
    } finally {
      projectsLoading.value = false
    }
  }

  const loadOrders = async () => {
    ordersLoading.value = true
    try {
      const res = await fetchAdminShashouOrders(orderQuery)
      orders.value = res?.list || []
    } finally {
      ordersLoading.value = false
    }
  }

  const loadAccounts = async () => {
    accountsLoading.value = true
    try {
      const res = await fetchAdminShashouAccounts({ page: 1, limit: 100 })
      accounts.value = res?.list || []
    } finally {
      accountsLoading.value = false
    }
  }

  const openProject = (record?: ShashouProject) => {
    Object.assign(projectForm, defaultProject(), record || {})
    projectVisible.value = true
  }

  const saveProject = async () => {
    saving.value = true
    try {
      await saveAdminShashouProject({ ...projectForm })
      ElMessage.success('保存成功')
      projectVisible.value = false
      await loadProjects()
    } finally {
      saving.value = false
    }
  }

  const deleteProject = async (id: number) => {
    await ElMessageBox.confirm('确定删除该项目配置吗？', '删除确认', { type: 'warning' })
    await deleteAdminShashouProject(id)
    ElMessage.success('删除成功')
    await loadProjects()
  }

  const syncOrder = async (id: number) => {
    await syncAdminShashouOrder(id)
    ElMessage.success('同步完成')
    await loadOrders()
  }

  const syncPending = async () => {
    syncLoading.value = true
    try {
      const res = await syncAdminShashouPending(100)
      ElMessage.success(`同步完成，更新 ${res.updated || 0} 条`)
      await loadOrders()
      await loadAccounts()
    } finally {
      syncLoading.value = false
    }
  }

  const orderTypeLabel = (value: number) =>
    ({ 1: '课外跑', 2: '晨跑', 3: '查询课外跑', 4: '退款', 5: '查询晨跑' })[value] || '-'
  const projectTypeLabel = (value: number) =>
    ({ 0: '源台', 1: '29二开', 2: '同系统' })[value] || '未知'
  const projectTypeTagMap: Record<number, 'primary' | 'warning' | 'success' | 'info'> = {
    0: 'primary',
    1: 'warning',
    2: 'success'
  }
  const projectTypeTag = (value: number): 'primary' | 'warning' | 'success' | 'info' =>
    projectTypeTagMap[value] || 'info'
  const credentialLabel = (field: 'key' | 'uid') => {
    if (projectForm.type === 0) return field === 'key' ? 'X-API-Key' : 'X-User-ID'
    return field === 'key' ? 'key' : 'uid'
  }
  const credentialHelp = (field: 'key' | 'uid') => {
    if (projectForm.type === 0) return field === 'key' ? '源台请求头密钥。' : '源台请求头用户 ID。'
    if (projectForm.type === 2)
      return field === 'key' ? '同系统 OpenAPI key 参数。' : '同系统 OpenAPI uid 参数。'
    return field === 'key' ? '29二开接口 key 参数。' : '29二开接口 uid 参数。'
  }
  const normalizeStatus = (value: string) =>
    String(value || '')
      .trim()
      .toLowerCase()
  const emptyText = (value: unknown) => {
    const text = String(value ?? '').trim()
    return text || '-'
  }
  const moneyText = (value: number | null | undefined) =>
    value === null || value === undefined ? '-' : `¥${Number(value).toFixed(2)}`
  const signedMoneyText = (value: number | null | undefined) => {
    if (value === null || value === undefined) return '-'
    const number = Number(value)
    if (number > 0) return `+¥${number.toFixed(2)}`
    if (number < 0) return `-¥${Math.abs(number).toFixed(2)}`
    return '¥0.00'
  }
  const diffClass = (value: number | null | undefined) => {
    const number = Number(value || 0)
    if (number > 0) return 'text-danger'
    if (number < 0) return 'text-success'
    return 'text-g-700'
  }
  const statusLabel = (value: string) =>
    ({
      pending: '待处理',
      wait: '待处理',
      waiting: '待处理',
      '0': '待处理',
      '1': '待处理',
      processing: '处理中',
      running: '处理中',
      '2': '处理中',
      completed: '已完成',
      complete: '已完成',
      '3': '已完成',
      success: '成功',
      refunded: '已退款',
      refund: '已退款',
      '4': '已退款',
      refunding: '退款中',
      failed: '失败',
      fail: '失败',
      error: '失败',
      '-1': '失败',
      '5': '失败'
    })[normalizeStatus(value)] || emptyText(value)
  const statusType = (value: string) => {
    const status = normalizeStatus(value)
    if (['completed', 'complete', 'success', '3'].includes(status)) return 'success'
    if (['failed', 'fail', 'error', 'refunded', 'refund', '-1', '4', '5'].includes(status))
      return 'danger'
    if (['processing', 'running', 'refunding', '2'].includes(status)) return 'warning'
    return 'info'
  }
  const paymentLabel = (value: string) =>
    ({
      pre_deducted: '已预扣',
      settled: '已结算',
      insufficient: '待补款',
      refunded: '已退款',
      paid: '已支付',
      pending: '待支付',
      no_refund: '无需退款',
      partial_refund: '部分退款'
    })[normalizeStatus(value)] || emptyText(value)
  const paymentType = (value: string) => {
    const status = normalizeStatus(value)
    if (['settled', 'paid'].includes(status)) return 'success'
    if (['insufficient', 'partial_refund', 'no_refund'].includes(status)) return 'warning'
    if (status === 'refunded') return 'danger'
    return 'info'
  }

  watch(activeTab, (tab) => {
    if (tab === 'orders') loadOrders()
    if (tab === 'accounts') loadAccounts()
  })

  onMounted(async () => {
    await loadProjects()
    await loadOrders()
  })
</script>

<style scoped>
  .guide-list {
    margin: 0;
    padding-left: 18px;
    color: var(--el-text-color-regular);
    line-height: 1.8;
  }
</style>
