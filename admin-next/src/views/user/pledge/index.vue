<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">我的质押</h2>
            <p class="mt-1 text-sm leading-6 text-g-500">查看生效中的折扣方案、到期时间和剩余天数。</p>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <ElTag effect="plain">{{ activePledges.length ? '已生效' : '暂无记录' }}</ElTag>
            <ElTag type="warning" effect="plain">即将到期 {{ expiringCount }} 项</ElTag>
            <ElButton plain :loading="loading" @click="loadData">刷新</ElButton>
          </div>
        </div>

        <div v-if="loading" class="mt-5">
          <ElSkeleton :rows="5" animated />
        </div>

        <div v-else-if="activePledges.length" class="mt-5 grid gap-4 xl:grid-cols-2">
          <article
            v-for="item in activePledges"
            :key="item.id"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-4"
            :class="remainDays(item.addtime, Number(item.days || 0)) <= 3 ? 'pledge-card--expiring' : ''"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <h3 class="text-base font-semibold text-g-900">{{ item.category_name || '未命名分类' }}</h3>
                <p class="mt-1 text-sm text-g-500">方案 ID {{ item.config_id }} / 记录 ID {{ item.id }}</p>
              </div>
              <ElTag :type="getPledgeTagType(item)" effect="plain">
                {{ getPledgeStatusText(item) }}
              </ElTag>
            </div>

            <ElProgress
              class="mt-4"
              :percentage="progressPercent(item.addtime, Number(item.days || 0))"
              :status="getProgressStatus(item)"
              :show-text="false"
              :stroke-width="8"
            />

            <div class="mt-4 grid gap-3 sm:grid-cols-2">
              <div class="rounded-custom-sm border-full-d bg-box p-3">
                <p class="text-xs font-medium text-g-400">质押金额</p>
                <p class="mt-2 text-lg font-semibold text-g-900">
                  ¥{{ formatMoney(item.amount) }}
                </p>
              </div>
              <div class="rounded-custom-sm border-full-d bg-box p-3">
                <p class="text-xs font-medium text-g-400">折扣率</p>
                <p class="mt-2 text-lg font-semibold text-[var(--el-color-primary)]">
                  {{ formatDiscount(item.discount_rate) }}
                </p>
              </div>
              <div class="rounded-custom-sm border-full-d bg-box p-3">
                <p class="text-xs font-medium text-g-400">质押开始</p>
                <p class="mt-2 text-sm font-medium text-g-800">{{ item.addtime || '-' }}</p>
              </div>
              <div class="rounded-custom-sm border-full-d bg-box p-3">
                <p class="text-xs font-medium text-g-400">剩余天数</p>
                <p
                  class="mt-2 text-lg font-semibold"
                  :class="
                    remainDays(item.addtime, Number(item.days || 0)) <= 3
                      ? 'text-[var(--el-color-warning)]'
                      : 'text-[var(--el-color-success)]'
                  "
                >
                  {{ remainDays(item.addtime, Number(item.days || 0)) }} 天
                </p>
              </div>
            </div>

            <div class="mt-4 flex items-center justify-between gap-3">
              <p class="text-sm text-g-500">
                到期时间 {{ formatDate(calcExpiry(item.addtime, Number(item.days || 0))) }}
              </p>
              <ElPopconfirm
                title="提前取消会扣除部分质押金，确认继续吗？"
                width="260"
                @confirm="handleCancel(item.id)"
              >
                <template #reference>
                  <ElButton plain type="danger">取消质押</ElButton>
                </template>
              </ElPopconfirm>
            </div>
          </article>
        </div>

        <div v-else class="mt-8">
          <ElEmpty description="暂无生效中的质押记录" />
        </div>
      </article>

      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">可用方案</h2>
            <p class="mt-1 text-sm leading-6 text-g-500">选择适合的分类方案，确认后会直接从余额扣除质押金额。</p>
          </div>
          <ElTag effect="plain">{{ configs.length }} 个方案</ElTag>
        </div>

        <div v-if="loading" class="mt-5">
          <ElSkeleton :rows="5" animated />
        </div>

        <div v-else-if="configs.length" class="mt-5 space-y-4">
          <article
            v-for="config in configs"
            :key="config.id"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <h3 class="text-base font-semibold text-g-900">
                  {{ config.category_name || '未命名分类' }}
                </h3>
                <p class="mt-1 text-sm text-g-500">质押 {{ config.days }} 天后自动返还本金</p>
              </div>
              <ElTag type="primary" effect="plain">{{ formatDiscount(config.discount_rate) }}</ElTag>
            </div>

            <div class="mt-4 grid gap-3 sm:grid-cols-2">
              <div class="rounded-custom-sm border-full-d bg-box p-3">
                <p class="text-xs font-medium text-g-400">质押金额</p>
                <p class="mt-2 text-lg font-semibold text-g-900">
                  ¥{{ formatMoney(config.amount) }}
                </p>
              </div>
              <div class="rounded-custom-sm border-full-d bg-box p-3">
                <p class="text-xs font-medium text-g-400">取消手续费</p>
                <p class="mt-2 text-lg font-semibold text-[var(--el-color-warning)]">
                  {{ formatPercent(config.cancel_fee) }}
                </p>
              </div>
            </div>

            <div class="mt-4 flex items-center justify-between gap-3">
              <p class="text-sm text-g-500">生效后对应分类将按该折扣率下单。</p>
              <ElPopconfirm
                :title="`确认质押 ¥${formatMoney(config.amount)} 并启用当前方案吗？`"
                width="280"
                @confirm="handleCreate(config.id)"
              >
                <template #reference>
                  <ElButton type="primary">立即质押</ElButton>
                </template>
              </ElPopconfirm>
            </div>
          </article>

          <ElAlert
            type="info"
            show-icon
            :closable="false"
            title="质押后对应分类课程会按折扣价结算，到期自动退还。提前取消将按方案收取手续费。"
          />
        </div>

        <div v-else class="mt-8">
          <ElEmpty description="当前暂无可用质押方案" />
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
  import type { ProgressProps, TagProps } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import {
    cancelLegacyPledge,
    createLegacyPledge,
    fetchLegacyMyPledges,
    fetchLegacyPledgeConfigs,
    type LegacyPledgeConfig,
    type LegacyPledgeRecord
  } from '@/api/legacy/auxiliary'

  defineOptions({ name: 'UserPledgePage' })

  const loading = ref(false)
  const configs = ref<LegacyPledgeConfig[]>([])
  const pledges = ref<LegacyPledgeRecord[]>([])

  const activePledges = computed(() => pledges.value.filter((item) => Number(item.status || 1) === 1))
  const expiringCount = computed(
    () => activePledges.value.filter((item) => remainDays(item.addtime, Number(item.days || 0)) <= 3).length
  )
  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)
  const formatPercent = (value?: number | string) => `${(Number(value || 0) * 100).toFixed(0)}%`
  const formatDiscount = (value?: number | string) => `${(Number(value || 0) * 100).toFixed(0)}% 折`

  const calcExpiry = (addtime: string, days: number) => {
    const start = new Date(addtime)
    return new Date(start.getTime() + days * 24 * 60 * 60 * 1000)
  }

  const formatDate = (date: Date) => {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hour = String(date.getHours()).padStart(2, '0')
    const minute = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day} ${hour}:${minute}`
  }

  const remainDays = (addtime: string, days: number) => {
    if (!addtime || days <= 0) {
      return 0
    }
    const diff = calcExpiry(addtime, days).getTime() - Date.now()
    return Math.max(0, Math.ceil(diff / (24 * 60 * 60 * 1000)))
  }

  const progressPercent = (addtime: string, days: number) => {
    if (!addtime || days <= 0) {
      return 0
    }
    const start = new Date(addtime).getTime()
    const total = days * 24 * 60 * 60 * 1000
    const elapsed = Date.now() - start
    return Math.min(100, Math.max(0, Math.round((elapsed / total) * 100)))
  }

  const getProgressStatus = (item: LegacyPledgeRecord): ProgressProps['status'] => {
    const remain = remainDays(item.addtime, Number(item.days || 0))
    if (remain <= 0) {
      return 'success'
    }
    if (remain <= 3) {
      return 'warning'
    }
    return ''
  }

  const getPledgeTagType = (item: LegacyPledgeRecord): TagProps['type'] => {
    const remain = remainDays(item.addtime, Number(item.days || 0))
    if (remain <= 0) {
      return 'info'
    }
    if (remain <= 3) {
      return 'warning'
    }
    return 'success'
  }

  const getPledgeStatusText = (item: LegacyPledgeRecord) => {
    const remain = remainDays(item.addtime, Number(item.days || 0))
    if (remain <= 0) {
      return '已到期'
    }
    if (remain <= 3) {
      return '即将到期'
    }
    return '生效中'
  }

  const loadData = async () => {
    loading.value = true
    try {
      const [configResult, pledgeResult] = await Promise.all([
        fetchLegacyPledgeConfigs(),
        fetchLegacyMyPledges()
      ])
      configs.value = Array.isArray(configResult) ? configResult : []
      pledges.value = Array.isArray(pledgeResult) ? pledgeResult : []
    } finally {
      loading.value = false
    }
  }

  const handleCreate = async (configId: number) => {
    await createLegacyPledge(configId)
    ElMessage.success('质押成功')
    await loadData()
  }

  const handleCancel = async (id: number) => {
    await cancelLegacyPledge(id)
    ElMessage.success('质押已取消，余额将按规则返还')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>

<style scoped lang="scss">
  .pledge-card--expiring {
    border-color: var(--el-color-warning-light-5);
    box-shadow: 0 0 0 1px var(--el-color-warning-light-8) inset;
  }
</style>
