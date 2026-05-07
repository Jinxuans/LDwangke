<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">当前充值方案</h2>
            <p class="mt-1 text-sm leading-6 text-g-500">
              在线支付和卡密兑换都支持，活动赠送规则自动取后端配置。
            </p>
          </div>
          <ElButton
            v-if="hasOnlineRechargePermission"
            plain
            :loading="channelLoading"
            @click="loadChannels"
          >
            刷新渠道
          </ElButton>
        </div>

        <div class="mt-4 flex flex-wrap gap-3">
          <ElTag effect="plain">账户充值</ElTag>
          <ElTag :type="hasOnlineRechargePermission ? 'success' : 'warning'" effect="plain">
            {{ hasOnlineRechargePermission ? '在线充值已开启' : '仅卡密充值' }}
          </ElTag>
          <ElTag v-if="isActivityDay" type="warning" effect="plain">
            活动赠送进行中
          </ElTag>
        </div>

        <div
          v-if="isActivityDay"
          class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4"
        >
          <p class="text-sm font-medium text-g-700">
            {{ bonusConfig?.activity?.hint || '今日活动加成已开启，充值更划算。' }}
          </p>
        </div>

        <div
          v-if="bonusPreview"
          class="mt-4 rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4"
        >
          <p class="text-sm text-g-700">
            充值 <span class="font-semibold">¥{{ formatMoney(amount) }}</span>，赠送
            <span class="font-semibold text-[var(--el-color-warning)]">
              ¥{{ bonusPreview.bonus.toFixed(2) }}
            </span>
            ，实际到账
            <span class="font-semibold text-[var(--el-color-success)]">
              ¥{{ bonusPreview.total.toFixed(2) }}
            </span>
          </p>
        </div>

        <div class="mt-5 flex flex-wrap gap-3">
          <ElButton
            v-for="item in quickAmounts"
            :key="item"
            :type="amount === item ? 'primary' : 'default'"
            round
            @click="amount = item"
          >
            ¥{{ item }}
          </ElButton>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <ElInputNumber
            v-model="amount"
            class="w-full"
            :min="1"
            :max="100000"
            :precision="2"
            :step="10"
            placeholder="充值金额"
          />
          <ElButton
            type="primary"
            :disabled="!hasOnlineRechargePermission || !selectedChannel || amount < 1"
            :loading="creating"
            @click="handlePay"
          >
            立即创建支付单
          </ElButton>
        </div>
      </article>

      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">卡密充值</h2>
            <p class="mt-1 text-sm text-g-500">没有在线支付权限时，也可以直接使用卡密完成充值。</p>
          </div>
          <ElTag effect="plain">即时到账</ElTag>
        </div>

        <div class="mt-5 space-y-4">
          <ElInput
            v-model="cardKeyCode"
            clearable
            placeholder="输入卡密内容"
            @keyup.enter="handleCardKey"
          />
          <ElButton type="primary" class="w-full" :loading="cardKeyLoading" @click="handleCardKey">
            兑换卡密
          </ElButton>
        </div>

        <div v-if="displayRules.length" class="mt-5 grid gap-4 md:grid-cols-2">
          <article
            v-for="(rule, index) in displayRules"
            :key="`${rule.min}-${rule.max}-${index}`"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-4"
          >
            <p class="text-xs font-medium text-g-400">赠送规则</p>
            <p class="mt-3 text-base font-semibold text-g-900">充值 ¥{{ rule.min }} ~ ¥{{ rule.max }}</p>
            <p class="mt-2 text-sm text-g-500">赠送 {{ rule.bonus_pct }}%</p>
          </article>
        </div>
      </article>
    </section>

    <section v-if="hasOnlineRechargePermission" class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-lg font-semibold text-g-900">在线支付方式</h2>
          <p class="mt-1 text-sm text-g-500">选择支付渠道后即可创建充值订单，支付回跳后也会自动检测到账状态。</p>
        </div>
        <ElButton plain :loading="channelLoading" @click="loadChannels">刷新渠道</ElButton>
      </div>

      <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <button
          v-for="channel in channels"
          :key="channel.key"
          type="button"
          class="rounded-custom-sm border px-5 py-5 text-left transition-all"
          :class="
            selectedChannel === channel.key
              ? 'border-[var(--el-color-primary)] bg-[var(--el-color-primary-light-9)]'
              : 'border-full-d bg-box hover:border-[var(--el-color-primary-light-5)]'
          "
          @click="selectedChannel = channel.key"
        >
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-xs font-medium text-g-400">支付渠道</p>
                <p class="mt-2 text-lg font-semibold text-g-900">{{ channel.label }}</p>
              </div>
            <span
              class="rounded-full px-3 py-1 text-xs"
              :class="
                selectedChannel === channel.key
                  ? 'bg-[var(--el-color-primary-light-8)] text-[var(--el-color-primary)]'
                  : 'bg-g-100 text-g-500'
              "
            >
              {{ selectedChannel === channel.key ? '已选中' : channel.key }}
            </span>
          </div>
        </button>
      </div>
    </section>

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadOrders(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">充值记录</ElTag>
            <ElTag type="info" effect="plain">共 {{ pagination.total }} 条</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ orders.length }} 条</ElTag>
            <ElTag v-if="selectedChannel" type="primary" effect="plain">渠道 {{ selectedChannel }}</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="orders"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { h, nextTick } from 'vue'
  import { storeToRefs } from 'pinia'
  import { ElButton, ElMessage, ElTag } from 'element-plus'
  import { useRouter } from 'vue-router'
  import {
    checkLegacyPayStatus,
    createLegacyPayOrder,
    fetchLegacyPayChannels,
    fetchLegacyPayOrders,
    useLegacyCardKey,
    type LegacyPayChannel,
    type LegacyPayOrder
  } from '@/api/legacy/user-center'
  import { useSiteStore } from '@/store/modules/site'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'UserRechargePage' })

  interface BonusRule {
    min: number
    max: number
    bonus_pct: number
  }

  interface BonusActivity {
    enabled: boolean
    hint?: string
    rules: BonusRule[]
    weekdays: number[]
  }

  interface BonusConfig {
    activity: BonusActivity
    enabled: boolean
    rules: BonusRule[]
  }

  const quickAmounts = [50, 100, 200, 500, 1000]

  const router = useRouter()
  const siteStore = useSiteStore()
  const { config } = storeToRefs(siteStore)

  const channelLoading = ref(false)
  const channels = ref<LegacyPayChannel[]>([])
  const selectedChannel = ref('')
  const amount = ref(50)
  const creating = ref(false)

  const cardKeyCode = ref('')
  const cardKeyLoading = ref(false)

  const loading = ref(false)
  const orders = ref<LegacyPayOrder[]>([])
  const pagination = reactive({
    current: 1,
    size: 10,
    total: 0
  })

  const bonusConfig = computed<BonusConfig | null>(() => {
    const raw = config.value.recharge_bonus_rules
    if (!raw) {
      return null
    }
    try {
      return JSON.parse(raw)
    } catch {
      return null
    }
  })

  const isActivityDay = computed(() => {
    if (!bonusConfig.value?.activity?.enabled) {
      return false
    }
    const weekday = new Date().getDay()
    return (bonusConfig.value.activity.weekdays || []).includes(weekday)
  })

  const displayRules = computed(() => {
    if (!bonusConfig.value?.enabled) {
      return []
    }
    if (isActivityDay.value && bonusConfig.value.activity?.rules?.length) {
      return bonusConfig.value.activity.rules
    }
    return bonusConfig.value.rules || []
  })

  const bonusPreview = computed(() => {
    if (!displayRules.value.length || !amount.value) {
      return null
    }

    const current = Number(amount.value || 0)
    const matched = displayRules.value.find((item) => current >= item.min && current < item.max)
    if (!matched || matched.bonus_pct <= 0) {
      return null
    }

    const bonus = (current * matched.bonus_pct) / 100
    return {
      pct: matched.bonus_pct,
      bonus,
      total: current + bonus
    }
  })

  const hasOnlineRechargePermission = computed(() => channels.value.length > 0)

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const getStatusText = (status: number) => {
    if (status === 2) return '已到账'
    if (status === 1) return '已支付'
    return '待支付'
  }

  const getStatusType = (status: number): 'warning' | 'success' | 'primary' => {
    if (status >= 1) {
      return 'success'
    }
    if (status === 0) {
      return 'warning'
    }
    return 'primary'
  }

  const { columns, columnChecks } = useTableColumns<LegacyPayOrder>(() => [
    { prop: 'out_trade_no', label: '订单号', minWidth: 220 },
    {
      prop: 'money',
      label: '金额',
      width: 120,
      align: 'right',
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, `¥${formatMoney(row.money)}`)
    },
    {
      prop: 'status',
      label: '状态',
      width: 110,
      formatter: (row) => h(ElTag, { type: getStatusType(row.status), effect: 'plain' }, () => getStatusText(row.status))
    },
    { prop: 'addtime', label: '创建时间', width: 180 },
    {
      prop: 'operation',
      label: '操作',
      width: 120,
      fixed: 'right',
      formatter: (row) =>
        row.status === 0
          ? h(ElButton, { text: true, type: 'primary', onClick: () => handleCheckPay(row.out_trade_no) }, () => '检测到账')
          : h('span', { class: 'text-sm text-g-400' }, '已完成')
    }
  ])

  const loadChannels = async () => {
    channelLoading.value = true
    try {
      const result = await fetchLegacyPayChannels()
      channels.value = Array.isArray(result) ? result : []
      if (!selectedChannel.value && channels.value.length) {
        selectedChannel.value = channels.value[0]?.key || ''
      }
    } finally {
      channelLoading.value = false
    }
  }

  const loadOrders = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyPayOrders(pagination.current, pagination.size)
      orders.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handlePay = async () => {
    const amountValue = Number(amount.value || 0)
    if (amountValue < 1) {
      ElMessage.warning('请输入有效充值金额')
      return
    }
    if (!selectedChannel.value) {
      ElMessage.warning('请选择支付方式')
      return
    }

    creating.value = true
    try {
      const result = await createLegacyPayOrder(amountValue, selectedChannel.value)
      ElMessage.success(result.pay_url ? '订单已创建，正在跳转支付' : '充值订单已创建')
      if (result.pay_url) {
        window.open(result.pay_url, '_blank', 'noopener,noreferrer')
      }
      await loadOrders(1)
    } finally {
      creating.value = false
    }
  }

  const handleCardKey = async () => {
    if (!cardKeyCode.value.trim()) {
      ElMessage.warning('请输入卡密')
      return
    }

    cardKeyLoading.value = true
    try {
      const result = await useLegacyCardKey(cardKeyCode.value.trim())
      ElMessage.success(result.msg || `充值成功，到账 ¥${formatMoney(result.money)}`)
      cardKeyCode.value = ''
      await loadOrders(1)
    } finally {
      cardKeyLoading.value = false
    }
  }

  const handleCheckPay = async (outTradeNo: string) => {
    const result = await checkLegacyPayStatus(outTradeNo)
    if (result.status === 1) {
      ElMessage.success(result.msg || '订单已到账')
      await loadOrders(pagination.current)
      return
    }
    ElMessage.info(result.msg || '订单暂未到账')
  }

  const handleCurrentChange = (page: number) => {
    loadOrders(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadOrders(1)
  }

  const getPayReturnParams = () => {
    const fromSearch = new URLSearchParams(window.location.search)
    const hash = window.location.hash || ''
    const queryIndex = hash.indexOf('?')
    const fromHash = queryIndex >= 0 ? new URLSearchParams(hash.slice(queryIndex + 1)) : new URLSearchParams()
    const read = (key: string) => fromHash.get(key) || fromSearch.get(key) || ''
    return {
      outTradeNo: read('out_trade_no')
    }
  }

  const clearPayReturnParams = async () => {
    await router.replace('/user/recharge')
    await nextTick()
    window.history.replaceState(
      {},
      '',
      `${window.location.origin}${window.location.pathname}${window.location.hash}`
    )
  }

  const handlePayReturn = async () => {
    const params = getPayReturnParams()
    if (!params.outTradeNo) {
      return
    }

    try {
      const result = await checkLegacyPayStatus(params.outTradeNo)
      if (result.status === 1) {
        ElMessage.success(result.msg || '支付成功')
      } else {
        ElMessage.info(result.msg || '订单尚未支付成功')
      }
      await loadOrders(1)
    } finally {
      await clearPayReturnParams()
    }
  }

  onMounted(async () => {
    await siteStore.initPublicConfig()
    await Promise.all([loadChannels(), loadOrders(1)])
    await handlePayReturn()
  })
</script>
