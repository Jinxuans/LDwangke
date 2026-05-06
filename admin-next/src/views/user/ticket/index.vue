<template>
  <div class="art-full-height">
    <ElCard class="art-table-card">
      <div class="border-b-d px-5 py-4">
        <ArtTableHeader layout="refresh" @refresh="loadTickets(pagination.page)">
          <template #left>
            <ElSpace wrap>
              <ElTag effect="plain">我的工单</ElTag>
              <ElTag type="info" effect="plain">共 {{ pagination.total }} 条</ElTag>
              <ElButton type="primary" plain @click="openCreateDialog">提交工单</ElButton>
            </ElSpace>
          </template>
        </ArtTableHeader>
      </div>

      <ElTable :data="tickets" v-loading="loading" stripe class="w-full">
        <ElTableColumn prop="id" label="ID" width="84" />
        <ElTableColumn label="关联订单" min-width="170">
          <template #default="{ row }">
            <div v-if="row.oid > 0" class="leading-6">
              <p class="font-medium text-g-900">#{{ row.oid }}</p>
              <p class="truncate text-xs text-g-500">{{ row.type || '订单反馈' }}</p>
            </div>
            <span v-else class="text-sm text-g-400">未关联订单</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="内容" min-width="320">
          <template #default="{ row }">
            <p class="line-clamp-2 text-sm leading-6 text-g-700">{{ row.content }}</p>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="110">
          <template #default="{ row }">
            <ElTag :type="getTicketStatusTag(row.status)">{{ getTicketStatusText(row.status) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="回复摘要" min-width="220">
          <template #default="{ row }">
            <p class="line-clamp-2 text-sm leading-6 text-g-500">
              {{ row.reply || '暂无回复' }}
            </p>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="addtime" label="提交时间" width="172" />
        <ElTableColumn label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <ElButton text type="primary" @click="showDetail(row)">查看</ElButton>
              <ElButton text @click="goChat">去聊天</ElButton>
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
    </ElCard>

    <ElDialog v-model="createVisible" title="提交工单" width="720px">
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">问题描述</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">
              提交后会自动建立与客服的会话，并把这次工单内容同步到聊天里。
            </p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">问题类型</label>
              <ElSelect v-model="createForm.type" class="w-full" placeholder="请选择问题类型">
                <ElOption label="订单反馈" value="订单反馈" />
                <ElOption label="账号问题" value="账号问题" />
                <ElOption label="充值结算" value="充值结算" />
                <ElOption label="其他问题" value="其他问题" />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">问题描述</label>
              <ElInput
                v-model="createForm.content"
                type="textarea"
                :rows="7"
                resize="none"
                placeholder="请尽量写清订单号、异常现象、希望如何处理。"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h3 class="text-lg font-semibold text-g-900">关联订单</h3>
              <p class="mt-1 text-sm text-g-500">如果问题来自具体订单，建议关联，处理更快。</p>
            </div>
            <ElButton text @click="loadOrders">刷新订单</ElButton>
          </div>

          <ElSelect
            v-model="createForm.oid"
            class="mt-4 w-full"
            filterable
            clearable
            placeholder="可选，选择一笔订单"
            :loading="orderLoading"
          >
            <ElOption
              v-for="option in orderOptions"
              :key="option.oid"
              :label="option.label"
              :value="option.oid"
            />
          </ElSelect>

          <div class="mt-5 space-y-3">
            <div
              v-for="option in orderOptions.slice(0, 5)"
              :key="option.oid"
              class="rounded-custom-sm border-full-d px-4 py-3 transition hover:border-[var(--el-color-primary)]"
              :class="createForm.oid === option.oid ? 'border-[var(--el-color-primary)] bg-[var(--el-color-primary-light-9)]' : 'bg-box'"
            >
              <button type="button" class="w-full text-left" @click="createForm.oid = option.oid">
                <div class="flex items-center justify-between gap-3">
                  <span class="font-medium text-g-900">#{{ option.oid }}</span>
                  <ElTag size="small" effect="plain">{{ option.status || '待处理' }}</ElTag>
                </div>
                <p class="mt-2 text-sm text-g-600">{{ option.label }}</p>
              </button>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="createVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="createLoading" @click="handleCreate">提交并进入聊天</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="detailVisible" title="工单详情" width="720px">
      <div v-if="currentTicket" class="space-y-5">
        <section class="rounded-custom-sm border-full-d bg-g-100/50 p-5">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-3">
              <ElTag :type="getTicketStatusTag(currentTicket.status)">
                {{ getTicketStatusText(currentTicket.status) }}
              </ElTag>
              <span class="text-base font-semibold text-g-900">工单 #{{ currentTicket.id }}</span>
            </div>
            <span class="text-sm text-g-500">{{ currentTicket.addtime }}</span>
          </div>
          <p class="mt-4 whitespace-pre-wrap text-sm leading-7 text-g-700">{{ currentTicket.content }}</p>
        </section>

        <section
          v-if="currentTicket.oid > 0"
          class="rounded-custom-sm border-full-d bg-box p-5"
        >
          <h3 class="text-sm font-semibold text-g-900">关联订单 #{{ currentTicket.oid }}</h3>
          <div class="mt-3 grid gap-3 sm:grid-cols-2">
            <div class="text-sm text-g-600">平台：{{ currentTicket.order_pt || '-' }}</div>
            <div class="text-sm text-g-600">账号：{{ currentTicket.order_user || '-' }}</div>
            <div class="text-sm text-g-600">状态：{{ currentTicket.order_status || '-' }}</div>
            <div class="text-sm text-g-600">YID：{{ currentTicket.order_yid || '-' }}</div>
          </div>
        </section>

        <section
          v-if="currentTicket.reply"
          class="rounded-custom-sm border-full-d bg-box p-5"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <h3 class="text-sm font-semibold text-g-900">管理员回复</h3>
            <span class="text-sm text-g-500">{{ currentTicket.reply_time || '刚刚' }}</span>
          </div>
          <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-g-700">{{ currentTicket.reply }}</p>
        </section>

        <div class="flex flex-wrap gap-3">
          <ElButton type="primary" @click="goChat">去在线客服</ElButton>
          <ElButton
            v-if="currentTicket.status !== 3"
            type="danger"
            plain
            @click="handleClose(currentTicket)"
          >
            关闭工单
          </ElButton>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { createLegacyChatSession, sendLegacyChatMessage } from '@/api/legacy/chat'
  import { fetchLegacyOrderList, type LegacyOrderItem } from '@/api/legacy/order'
  import {
    closeLegacyUserTicket,
    createLegacyUserTicket,
    fetchLegacyUserTickets,
    type LegacyTicket
  } from '@/api/legacy/ticket'

  defineOptions({ name: 'UserTicketPage' })

  interface TicketOrderOption {
    oid: number
    label: string
    status: string
  }

  const router = useRouter()

  const loading = ref(false)
  const tickets = ref<LegacyTicket[]>([])
  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const createVisible = ref(false)
  const createLoading = ref(false)
  const createForm = reactive({
    oid: undefined as number | undefined,
    type: '订单反馈',
    content: ''
  })

  const orderLoading = ref(false)
  const orderOptions = ref<TicketOrderOption[]>([])

  const detailVisible = ref(false)
  const currentTicket = ref<LegacyTicket | null>(null)

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

  const mapOrderOption = (order: LegacyOrderItem): TicketOrderOption => ({
    oid: order.oid,
    status: order.status,
    label: `${order.ptname || '未知平台'} / ${order.user || '-'} / ${order.kcname || '-'}`.trim()
  })

  const loadTickets = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyUserTickets(pagination.page, pagination.limit)
      tickets.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const loadOrders = async () => {
    orderLoading.value = true
    try {
      const result = await fetchLegacyOrderList({
        page: 1,
        limit: 100
      })
      orderOptions.value = (result.list || []).map(mapOrderOption)
    } finally {
      orderLoading.value = false
    }
  }

  const openCreateDialog = async () => {
    createForm.oid = undefined
    createForm.type = '订单反馈'
    createForm.content = ''
    createVisible.value = true
    await loadOrders()
  }

  const goChat = async () => {
    const session = await createLegacyChatSession(1)
    await router.push({
      path: '/chat',
      query: session.list_id ? { listId: String(session.list_id) } : undefined
    })
  }

  const handleCreate = async () => {
    if (!createForm.content.trim()) {
      ElMessage.warning('请先填写工单内容')
      return
    }

    createLoading.value = true
    try {
      await createLegacyUserTicket({
        oid: createForm.oid,
        type: createForm.type,
        content: createForm.content.trim()
      })

      const session = await createLegacyChatSession(1)
      if (session.list_id) {
        const prefix = createForm.oid ? `【工单反馈 #${createForm.oid}】` : '【工单反馈】'
        await sendLegacyChatMessage({
          list_id: session.list_id,
          to_uid: 1,
          content: `${prefix}${createForm.content.trim()}`
        })
      }

      ElMessage.success('工单已提交，正在进入在线客服')
      createVisible.value = false
      await loadTickets(1)
      await router.push({
        path: '/chat',
        query: session.list_id ? { listId: String(session.list_id) } : undefined
      })
    } finally {
      createLoading.value = false
    }
  }

  const showDetail = (ticket: LegacyTicket) => {
    currentTicket.value = { ...ticket }
    detailVisible.value = true
  }

  const handleClose = async (ticket: LegacyTicket) => {
    await ElMessageBox.confirm(`确定关闭工单 #${ticket.id} 吗？`, '关闭工单', {
      type: 'warning'
    })
    await closeLegacyUserTicket(ticket.id)
    ElMessage.success('工单已关闭')
    if (currentTicket.value?.id === ticket.id) {
      currentTicket.value.status = 3
    }
    await loadTickets(pagination.page)
  }

  onMounted(() => {
    loadTickets(1)
  })
</script>
