<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm overflow-hidden">
      <div class="border-b-d px-5 py-4">
        <div class="space-y-4">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="flex flex-wrap gap-2">
              <ElTag effect="plain">总工单 {{ stats.total }}</ElTag>
              <ElTag type="warning" effect="plain">待回复 {{ stats.pending }}</ElTag>
              <ElTag type="success" effect="plain">已回复 {{ stats.replied }}</ElTag>
              <ElTag type="info" effect="plain">已关闭 {{ stats.closed }}</ElTag>
              <ElTag type="primary" effect="plain">上游处理中 {{ stats.upstream_pending }}</ElTag>
            </div>
            <div class="inline-flex flex-wrap items-center gap-3 rounded-[calc(var(--custom-radius)+6px)] border border-[var(--art-card-border)] bg-[var(--el-fill-color-light)] px-4 py-3">
              <span class="text-sm text-g-500">自动关闭已回复工单</span>
              <ElInputNumber v-model="autoCloseDays" :min="1" :max="90" />
              <ElButton type="primary" plain :loading="autoCloseLoading" @click="handleAutoClose">
                执行
              </ElButton>
            </div>
          </div>

          <div class="flex flex-wrap items-center justify-between gap-4">
            <div class="text-sm text-g-500">支持筛选、回复、关闭、上游反馈和一键跳到聊天页。</div>
            <div class="flex flex-wrap gap-3">
              <ElSelect v-model="filters.status" class="w-[132px]" @change="handleSearch">
                <ElOption label="全部状态" :value="0" />
                <ElOption label="待回复" :value="1" />
                <ElOption label="已回复" :value="2" />
                <ElOption label="已关闭" :value="3" />
              </ElSelect>
              <ElInput
                v-model="filters.uid"
                class="w-[132px]"
                placeholder="用户 UID"
                clearable
                @keyup.enter="handleSearch"
              />
              <ElInput
                v-model="filters.search"
                class="w-[220px]"
                placeholder="搜内容 / 订单号"
                clearable
                @keyup.enter="handleSearch"
              />
              <ElButton type="primary" @click="handleSearch">搜索</ElButton>
              <ElButton plain @click="refreshAll">刷新</ElButton>
            </div>
          </div>
        </div>
      </div>

      <ElTable :data="tickets" v-loading="loading" stripe class="w-full">
        <ElTableColumn prop="id" label="ID" width="82" />
        <ElTableColumn prop="uid" label="UID" width="94" />
        <ElTableColumn label="关联订单" min-width="190">
          <template #default="{ row }">
            <div v-if="row.oid > 0" class="leading-6">
              <p class="font-medium text-g-900">#{{ row.oid }}</p>
              <p class="truncate text-xs text-g-500">
                {{ row.order_pt || '-' }} / {{ row.order_user || '-' }}
              </p>
            </div>
            <span v-else class="text-sm text-g-400">无关联订单</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="内容" min-width="320">
          <template #default="{ row }">
            <p class="line-clamp-2 text-sm leading-6 text-g-700">{{ row.content }}</p>
          </template>
        </ElTableColumn>
        <ElTableColumn label="工单状态" width="110">
          <template #default="{ row }">
            <ElTag :type="getTicketStatusTag(row.status)">{{ getTicketStatusText(row.status) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="上游反馈" width="132">
          <template #default="{ row }">
            <ElTag :type="getSupplierStatusTag(row.supplier_status, row.supplier_report_id)">
              {{ getSupplierStatusText(row.supplier_status, row.supplier_report_id) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="addtime" label="提交时间" width="172" />
        <ElTableColumn label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <ElButton text type="primary" @click="showDetail(row)">查看</ElButton>
              <ElButton text @click="goChat(row.uid)">聊天</ElButton>
              <ElButton
                v-if="row.oid > 0 && row.supplier_report_id === 0 && row.supplier_report_switch"
                text
                @click="handleReport(row)"
              >
                提交上游
              </ElButton>
              <ElButton v-if="row.supplier_report_id > 0" text @click="handleSyncReport(row)">
                同步
              </ElButton>
              <ElButton v-if="row.status !== 3" text type="danger" @click="handleClose(row)">
                关闭
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
          @current-change="loadTickets"
        />
      </div>
    </section>

    <ElDialog v-model="detailVisible" title="工单详情" width="860px">
      <div v-if="currentTicket" class="grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
        <section class="space-y-5">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-3">
                <ElTag :type="getTicketStatusTag(currentTicket.status)">
                  {{ getTicketStatusText(currentTicket.status) }}
                </ElTag>
                <span class="text-base font-semibold text-g-900">工单 #{{ currentTicket.id }}</span>
                <span class="text-sm text-g-500">UID {{ currentTicket.uid }}</span>
              </div>
              <span class="text-sm text-g-500">{{ currentTicket.addtime }}</span>
            </div>
            <p class="mt-4 whitespace-pre-wrap text-sm leading-7 text-g-700">{{ currentTicket.content }}</p>
          </article>

          <article
            v-if="currentTicket.oid > 0"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-5"
          >
            <div class="flex items-center justify-between gap-3">
              <h3 class="text-sm font-semibold text-g-900">关联订单 #{{ currentTicket.oid }}</h3>
              <ElTag effect="plain">{{ currentTicket.order_status || '待处理' }}</ElTag>
            </div>
            <div class="mt-4 grid gap-3 sm:grid-cols-2">
              <div class="text-sm text-g-600">平台：{{ currentTicket.order_pt || '-' }}</div>
              <div class="text-sm text-g-600">账号：{{ currentTicket.order_user || '-' }}</div>
              <div class="text-sm text-g-600">YID：{{ currentTicket.order_yid || '-' }}</div>
              <div class="text-sm text-g-600">
                上游配置：
                {{ currentTicket.supplier_report_hid_switch || '自动识别供应商' }}
              </div>
            </div>
          </article>

          <article
            v-if="currentTicket.reply"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-5"
          >
            <div class="flex items-center justify-between gap-3">
              <h3 class="text-sm font-semibold text-g-900">管理员回复</h3>
              <span class="text-sm text-g-500">{{ currentTicket.reply_time || '刚刚' }}</span>
            </div>
            <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-g-700">{{ currentTicket.reply }}</p>
          </article>

          <article
            v-if="currentTicket.supplier_report_id > 0 || currentTicket.supplier_report_switch"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-5"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-3">
                <h3 class="text-sm font-semibold text-g-900">上游反馈</h3>
                <ElTag
                  :type="getSupplierStatusTag(currentTicket.supplier_status, currentTicket.supplier_report_id)"
                >
                  {{ getSupplierStatusText(currentTicket.supplier_status, currentTicket.supplier_report_id) }}
                </ElTag>
              </div>
              <div class="flex gap-2">
                <ElButton
                  v-if="currentTicket.oid > 0 && currentTicket.supplier_report_id === 0 && currentTicket.supplier_report_switch"
                  type="primary"
                  plain
                  @click="handleReport(currentTicket)"
                >
                  提交上游
                </ElButton>
                <ElButton v-if="currentTicket.supplier_report_id > 0" plain @click="handleSyncReport(currentTicket)">
                  同步状态
                </ElButton>
              </div>
            </div>
            <p class="mt-3 text-sm leading-7 text-g-600">
              {{ currentTicket.supplier_answer || '暂无上游回执。' }}
            </p>
          </article>
        </section>

        <section class="space-y-5">
          <article class="rounded-custom-sm border-full-d bg-box p-5">
            <h3 class="text-lg font-semibold text-g-900">快捷操作</h3>
            <div class="mt-4 flex flex-col gap-3">
              <ElButton type="primary" @click="goChat(currentTicket.uid)">去聊天页</ElButton>
              <ElButton
                v-if="currentTicket.status !== 3"
                type="danger"
                plain
                @click="handleClose(currentTicket)"
              >
                关闭工单
              </ElButton>
            </div>
          </article>

          <article
            v-if="currentTicket.status === 1"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-5"
          >
            <h3 class="text-lg font-semibold text-g-900">回复工单</h3>
            <p class="mt-2 text-sm text-g-500">回复后工单会变为“已回复”，用户也会收到系统推送。</p>
            <ElInput
              v-model="replyContent"
              class="mt-4"
              type="textarea"
              :rows="7"
              resize="none"
              placeholder="输入回复内容"
            />
            <ElButton
              class="mt-4"
              type="primary"
              :loading="replyLoading"
              :disabled="!replyContent.trim()"
              @click="handleReply"
            >
              发送回复
            </ElButton>
          </article>
        </section>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { createLegacyChatSession } from '@/api/legacy/chat'
  import {
    autoCloseLegacyAdminTickets,
    closeLegacyAdminTicket,
    fetchLegacyAdminTicketStats,
    fetchLegacyAdminTickets,
    reportLegacyAdminTicket,
    replyLegacyAdminTicket,
    syncLegacyAdminTicketReport,
    type LegacyTicket,
    type LegacyTicketStats
  } from '@/api/legacy/ticket'

  defineOptions({ name: 'AdminTicketsPage' })

  const router = useRouter()

  const loading = ref(false)
  const tickets = ref<LegacyTicket[]>([])
  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })
  const filters = reactive({
    status: 0,
    uid: '',
    search: ''
  })

  const stats = ref<LegacyTicketStats>({
    total: 0,
    pending: 0,
    replied: 0,
    closed: 0,
    upstream_pending: 0
  })

  const autoCloseDays = ref(7)
  const autoCloseLoading = ref(false)

  const detailVisible = ref(false)
  const currentTicket = ref<LegacyTicket | null>(null)
  const replyContent = ref('')
  const replyLoading = ref(false)

  const getTicketStatusText = (status: number) => {
    if (status === 1) return '待回复'
    if (status === 2) return '已回复'
    if (status === 3) return '已关闭'
    return '未知'
  }

  const getTicketStatusTag = (status: number): 'warning' | 'success' | 'info' | 'danger' => {
    if (status === 1) return 'warning'
    if (status === 2) return 'success'
    if (status === 3) return 'info'
    return 'danger'
  }

  const getSupplierStatusText = (status: number, reportId: number) => {
    if (!reportId) return '未提交'
    const map: Record<number, string> = {
      0: '待处理',
      1: '处理完成',
      3: '暂时搁置',
      4: '处理中',
      6: '已退款'
    }
    return map[status] || '未知状态'
  }

  const getSupplierStatusTag = (
    status: number,
    reportId: number
  ): 'info' | 'success' | 'warning' | 'danger' | 'primary' => {
    if (!reportId) return 'info'
    if (status === 1) return 'success'
    if (status === 0 || status === 4) return 'primary'
    if (status === 3) return 'warning'
    if (status === 6) return 'danger'
    return 'info'
  }

  const loadStats = async () => {
    stats.value = await fetchLegacyAdminTicketStats()
  }

  const loadTickets = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyAdminTickets({
        page: pagination.page,
        limit: pagination.limit,
        status: filters.status || undefined,
        uid: filters.uid ? Number(filters.uid) : undefined,
        search: filters.search || undefined
      })
      tickets.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const refreshAll = async () => {
    await Promise.all([loadStats(), loadTickets(pagination.page)])
  }

  const handleSearch = async () => {
    await loadTickets(1)
  }

  const showDetail = (ticket: LegacyTicket) => {
    currentTicket.value = { ...ticket }
    replyContent.value = ''
    detailVisible.value = true
  }

  const goChat = async (uid: number) => {
    const session = await createLegacyChatSession(uid)
    await router.push({
      path: '/admin/chat',
      query: session.list_id ? { listId: String(session.list_id) } : undefined
    })
  }

  const handleReply = async () => {
    if (!currentTicket.value || !replyContent.value.trim()) {
      return
    }

    replyLoading.value = true
    try {
      await replyLegacyAdminTicket(currentTicket.value.id, replyContent.value.trim())
      ElMessage.success('回复成功')
      currentTicket.value.reply = replyContent.value.trim()
      currentTicket.value.reply_time = new Date().toLocaleString('sv-SE').replace(',', '')
      currentTicket.value.status = 2
      replyContent.value = ''
      await Promise.all([loadStats(), loadTickets(pagination.page)])
    } finally {
      replyLoading.value = false
    }
  }

  const handleClose = async (ticket: LegacyTicket) => {
    await ElMessageBox.confirm(`确定关闭工单 #${ticket.id} 吗？`, '关闭工单', {
      type: 'warning'
    })
    await closeLegacyAdminTicket(ticket.id)
    ElMessage.success('工单已关闭')
    if (currentTicket.value?.id === ticket.id) {
      currentTicket.value.status = 3
    }
    await Promise.all([loadStats(), loadTickets(pagination.page)])
  }

  const handleReport = async (ticket: LegacyTicket) => {
    const result = await reportLegacyAdminTicket(ticket.id)
    ElMessage.success(result.message || '已提交上游反馈')
    if (currentTicket.value?.id === ticket.id) {
      currentTicket.value.supplier_report_id = result.report_id || currentTicket.value.supplier_report_id
      currentTicket.value.supplier_status = 0
      currentTicket.value.supplier_answer = ''
    }
    await Promise.all([loadStats(), loadTickets(pagination.page)])
  }

  const handleSyncReport = async (ticket: LegacyTicket) => {
    const result = await syncLegacyAdminTicketReport(ticket.id)
    ElMessage.success(
      result.supplier_answer
        ? `${getSupplierStatusText(result.supplier_status, 1)}：${result.supplier_answer}`
        : `上游状态已更新为 ${getSupplierStatusText(result.supplier_status, 1)}`
    )
    if (currentTicket.value?.id === ticket.id) {
      currentTicket.value.supplier_status = result.supplier_status
      currentTicket.value.supplier_answer = result.supplier_answer
    }
    await loadTickets(pagination.page)
  }

  const handleAutoClose = async () => {
    await ElMessageBox.confirm(
      `确定关闭已回复超过 ${autoCloseDays.value} 天的工单吗？`,
      '自动关闭工单',
      {
        type: 'warning'
      }
    )
    autoCloseLoading.value = true
    try {
      const result = await autoCloseLegacyAdminTickets(autoCloseDays.value)
      ElMessage.success(`已关闭 ${result.closed || 0} 个超期工单`)
      await Promise.all([loadStats(), loadTickets(pagination.page)])
    } finally {
      autoCloseLoading.value = false
    }
  }

  onMounted(() => {
    loadStats()
    loadTickets(1)
  })
</script>
