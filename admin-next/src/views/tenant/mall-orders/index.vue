<template>
  <div class="tenant-mall-orders-page art-full-height">
    <ElCard class="art-table-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d px-5 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">支付订单</h2>
          <ElTag effect="plain">订单总数 {{ pagination.total }}</ElTag>
          <ElTag type="success" effect="plain">已下单 {{ linkedCount }}</ElTag>
          <ElTag type="warning" effect="plain">待支付 {{ pendingCount }}</ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="loading" @click="loadData(pagination.page)">刷新</ElButton>
          <ElButton plain @click="router.push('/tenant/withdraw')">商城提现</ElButton>
        </div>
      </div>

      <ElTable v-loading="loading" :data="orders" row-key="id">
        <ElTableColumn prop="out_trade_no" label="支付订单号" min-width="220" />
        <ElTableColumn label="商品信息" min-width="260">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="line-clamp-1 text-sm font-medium text-g-900">
                {{ row.product_name || `商品#${row.cid}` }}
              </p>
              <p class="mt-1 line-clamp-1 text-xs text-g-500">
                {{ row.course_name || '未记录课程名' }}
              </p>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="account" label="会员账号" min-width="140" />
        <ElTableColumn label="支付方式" width="120" align="center">
          <template #default="{ row }">{{ payTypeLabel(row.pay_type) }}</template>
        </ElTableColumn>
        <ElTableColumn label="金额" width="120" align="right">
          <template #default="{ row }">¥{{ formatMoney(row.money) }}</template>
        </ElTableColumn>
        <ElTableColumn label="支付状态" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="statusMeta(row.status).type" effect="plain">
              {{ statusMeta(row.status).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="关联业务单" width="120" align="center">
          <template #default="{ row }">
            {{ row.order_count || (row.order_id ? 1 : 0) }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="addtime" label="下单时间" width="180" />
        <ElTableColumn label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <ElButton text type="primary" @click="openDetail(row)">详情</ElButton>
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
          @current-change="loadData"
        />
      </div>
    </ElCard>

    <ElDialog v-model="detailVisible" title="商城订单详情" width="980px" destroy-on-close>
      <div v-if="currentOrder" class="grid gap-5 xl:grid-cols-[0.9fr_1.1fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">支付信息</h3>
          </div>

          <div class="mt-5 space-y-4">
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">支付订单号</p>
              <p class="mt-2 break-all text-sm font-semibold text-g-900">{{ currentOrder.out_trade_no }}</p>
            </article>

            <div class="grid gap-4 sm:grid-cols-2">
              <article class="rounded-custom-sm border-full-d p-4">
                <p class="text-xs font-medium text-g-400">支付金额</p>
                <p class="mt-2 text-lg font-semibold text-g-900">¥{{ formatMoney(currentOrder.money) }}</p>
              </article>
              <article class="rounded-custom-sm border-full-d p-4">
                <p class="text-xs font-medium text-g-400">支付状态</p>
                <p class="mt-2">
                  <ElTag :type="statusMeta(currentOrder.status).type" effect="plain">
                    {{ statusMeta(currentOrder.status).text }}
                  </ElTag>
                </p>
              </article>
            </div>

            <article class="rounded-custom-sm border-full-d p-4">
              <p class="text-xs font-medium text-g-400">商品与会员</p>
              <p class="mt-2 text-sm text-g-700">商品：{{ currentOrder.product_name || `商品#${currentOrder.cid}` }}</p>
              <p class="mt-2 text-sm text-g-700">课程：{{ currentOrder.course_name || '-' }}</p>
              <p class="mt-2 text-sm text-g-700">会员：{{ currentOrder.account || '-' }}</p>
              <p class="mt-2 text-sm text-g-700">支付方式：{{ payTypeLabel(currentOrder.pay_type) }}</p>
              <p class="mt-2 text-sm text-g-700">备注：{{ currentOrder.order_remarks || currentOrder.remark || '-' }}</p>
            </article>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="flex items-center justify-between gap-3 border-b-d pb-4">
            <div>
              <h3 class="text-lg font-semibold text-g-900">关联真实订单</h3>
              <p class="mt-1 text-sm text-g-500">商城支付完成后，可能会拆成一条或多条业务订单。</p>
            </div>
            <ElButton plain :loading="detailLoading" @click="currentOrder && openDetail(currentOrder)">
              刷新
            </ElButton>
          </div>

          <div v-loading="detailLoading" class="mt-5">
            <div v-if="linkedOrders.length" class="space-y-4">
              <article
                v-for="item in linkedOrders"
                :key="item.oid"
                class="rounded-custom-sm border-full-d bg-g-100/50 p-4"
              >
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div class="flex flex-wrap items-center gap-2">
                    <ElTag type="primary" effect="plain">订单 #{{ item.oid }}</ElTag>
                    <ElTag :type="linkedStatusTag(item.status)" effect="plain">{{ item.status || '待处理' }}</ElTag>
                  </div>
                  <ElButton size="small" plain @click="loadLogs(item.oid)">查看日志</ElButton>
                </div>

                <div class="mt-4 grid gap-3 sm:grid-cols-2">
                  <div class="text-sm text-g-600">课程：{{ item.kcname || item.ptname || '-' }}</div>
                  <div class="text-sm text-g-600">账号：{{ item.user || '-' }}</div>
                  <div class="text-sm text-g-600">课程 ID：{{ item.kcid || '-' }}</div>
                  <div class="text-sm text-g-600">供货价：¥{{ formatMoney(item.fees) }}</div>
                  <div class="text-sm text-g-600">进度：{{ item.process || '-' }}</div>
                  <div class="text-sm text-g-600">提交时间：{{ item.addtime || '-' }}</div>
                </div>

                <p class="mt-3 text-sm leading-6 text-g-600">备注：{{ item.remarks || '-' }}</p>

                <div v-if="logLoadingMap[item.oid]" class="mt-3 text-sm text-g-500">日志加载中...</div>
                <div v-else-if="orderLogMap[item.oid]?.length" class="mt-3 space-y-3">
                  <article
                    v-for="(log, index) in orderLogMap[item.oid]"
                    :key="`${item.oid}-${index}`"
                    class="rounded-custom-sm border-full-d bg-box p-3"
                  >
                    <div class="flex flex-wrap items-center gap-2">
                      <ElTag effect="plain">{{ log.time || '-' }}</ElTag>
                      <ElTag effect="plain">{{ log.status || '-' }}</ElTag>
                      <span class="text-sm text-g-500">{{ log.process || '-' }}</span>
                    </div>
                    <p class="mt-2 text-sm leading-6 text-g-600">{{ log.remarks || '-' }}</p>
                  </article>
                </div>
              </article>
            </div>
            <ElEmpty v-else description="暂未生成关联业务订单" />
          </div>
        </section>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router'
  import { fetchLegacyOrderLogs, type LegacyOrderLogEntry } from '@/api/legacy/order'
  import {
    fetchTenantMallLinkedOrders,
    fetchTenantMallOrders,
    type LegacyTenantLinkedOrder,
    type LegacyTenantMallPayOrder
  } from '@/api/legacy/tenant'

  defineOptions({ name: 'TenantMallOrdersPage' })

  const router = useRouter()
  const loading = ref(false)
  const detailLoading = ref(false)
  const detailVisible = ref(false)

  const orders = ref<LegacyTenantMallPayOrder[]>([])
  const linkedOrders = ref<LegacyTenantLinkedOrder[]>([])
  const currentOrder = ref<LegacyTenantMallPayOrder | null>(null)
  const orderLogMap = ref<Record<number, LegacyOrderLogEntry[]>>({})
  const logLoadingMap = ref<Record<number, boolean>>({})

  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const pendingCount = computed(
    () => orders.value.filter((item) => Number(item.status) === 0).length
  )
  const linkedCount = computed(
    () => orders.value.filter((item) => Number(item.order_id) > 0 || Number(item.order_count || 0) > 0).length
  )
  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  function payTypeLabel(type: string) {
    if (type === 'alipay') return '支付宝'
    if (type === 'wxpay') return '微信支付'
    if (type === 'qqpay') return 'QQ 支付'
    return type || '-'
  }

  function statusMeta(status: number): { text: string; type: 'danger' | 'info' | 'success' | 'warning' } {
    if (status === 2) return { text: '已下单', type: 'success' }
    if (status === 1) return { text: '已支付', type: 'info' }
    if (status === -1) return { text: '失败', type: 'danger' }
    return { text: '待支付', type: 'warning' }
  }

  function linkedStatusTag(status?: string): 'danger' | 'info' | 'primary' | 'success' | 'warning' {
    if (!status) return 'info'
    if (['已完成', '已上号', '已结课', '已完成待考试'].includes(status)) return 'success'
    if (['进行中', '刷课中', '学习中', '运行中'].includes(status)) return 'primary'
    if (['异常', '出错啦', '失败', '异常待处理', '补刷中'].includes(status)) return 'danger'
    if (['待处理', '等待中'].includes(status)) return 'warning'
    return 'info'
  }

  async function loadData(page = pagination.page) {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchTenantMallOrders({
        page: pagination.page,
        limit: pagination.limit
      })
      orders.value = result.list || []
      pagination.total = Number(result.total || result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  async function openDetail(order: LegacyTenantMallPayOrder) {
    currentOrder.value = order
    detailVisible.value = true
    detailLoading.value = true
    try {
      linkedOrders.value = await fetchTenantMallLinkedOrders(order.id)
    } finally {
      detailLoading.value = false
    }
  }

  async function loadLogs(oid: number) {
    if (logLoadingMap.value[oid]) return
    logLoadingMap.value = { ...logLoadingMap.value, [oid]: true }
    try {
      orderLogMap.value = {
        ...orderLogMap.value,
        [oid]: await fetchLegacyOrderLogs(oid)
      }
    } finally {
      logLoadingMap.value = { ...logLoadingMap.value, [oid]: false }
    }
  }

  onMounted(() => {
    loadData(1)
  })
</script>
