<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="grid gap-5 xl:grid-cols-[1.1fr_0.9fr]">
      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3 border-b-d pb-4">
          <div>
            <h2 class="text-lg font-semibold text-g-900">签到状态</h2>
            <p class="mt-1 text-sm leading-6 text-g-500">签到成功后，奖励会直接计入余额。</p>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <ElTag :type="checkedIn ? 'success' : 'warning'" effect="plain">
              {{ checkedIn ? '已完成' : '待签到' }}
            </ElTag>
            <ElButton plain :loading="loading" @click="loadStatus">刷新</ElButton>
          </div>
        </div>

        <div class="mt-5 grid gap-4 sm:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">今日奖励</p>
            <p class="mt-2 text-2xl font-semibold text-[var(--el-color-primary)]">
              ¥{{ formatMoney(reward) }}
            </p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">签到状态</p>
            <p class="mt-2 text-2xl font-semibold text-g-900">{{ checkedIn ? '已完成' : '待签到' }}</p>
          </article>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <div class="flex items-center gap-3">
            <ElIcon
              :size="20"
              :class="
                checkedIn
                  ? 'text-[var(--el-color-success)]'
                  : 'text-[var(--el-color-warning)]'
              "
            >
              <CircleCheckFilled v-if="checkedIn" />
              <Present v-else />
            </ElIcon>
            <p class="text-sm font-medium text-g-900">
              {{
                checkedIn
                  ? `本次签到奖励 ¥${formatMoney(reward)} 已到账，明天可继续签到。`
                  : '点击按钮即可领取今日随机奖励。'
              }}
            </p>
          </div>

          <div class="mt-4 flex flex-wrap gap-3">
            <ElButton
              v-if="!checkedIn"
              type="primary"
              :loading="submitting"
              @click="handleCheckin"
            >
              立即签到
            </ElButton>
            <ElButton v-else plain @click="loadStatus">刷新状态</ElButton>
          </div>
        </div>
      </article>

      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">签到说明</h2>
            <p class="mt-1 text-sm leading-6 text-g-500">签到规则和到账方式都在这里。</p>
          </div>
          <ElTag effect="plain">规则</ElTag>
        </div>

        <div class="mt-5 space-y-4">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center gap-2">
              <ElIcon class="text-[var(--el-color-primary)]"><Calendar /></ElIcon>
              <p class="text-sm font-semibold text-g-900">每天限领一次</p>
            </div>
            <p class="mt-2 text-sm leading-6 text-g-500">当天签到成功后，重复点击不会再次发奖。</p>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center gap-2">
              <ElIcon class="text-[var(--el-color-success)]"><Wallet /></ElIcon>
              <p class="text-sm font-semibold text-g-900">奖励直接到账</p>
            </div>
            <p class="mt-2 text-sm leading-6 text-g-500">奖励会直接累加到余额，无需手动领取或审核。</p>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center gap-2">
              <ElIcon class="text-[var(--el-color-warning)]"><Opportunity /></ElIcon>
              <p class="text-sm font-semibold text-g-900">建议固定访问</p>
            </div>
            <p class="mt-2 text-sm leading-6 text-g-500">把本页加入常用菜单，避免遗漏每日奖励。</p>
          </article>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import {
    Calendar,
    CircleCheckFilled,
    Opportunity,
    Present,
    Wallet
  } from '@element-plus/icons-vue'
  import {
    fetchLegacyUserCheckinStatus,
    submitLegacyUserCheckin
  } from '@/api/legacy/auxiliary'

  defineOptions({ name: 'UserCheckinPage' })

  const loading = ref(false)
  const submitting = ref(false)
  const checkedIn = ref(false)
  const reward = ref(0)

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const loadStatus = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyUserCheckinStatus()
      checkedIn.value = Boolean(result.checked_in)
      reward.value = Number(result.reward_money || 0)
    } finally {
      loading.value = false
    }
  }

  const handleCheckin = async () => {
    submitting.value = true
    try {
      const result = await submitLegacyUserCheckin()
      reward.value = Number(result.reward_money || 0)
      checkedIn.value = true
      ElMessage.success(`签到成功，奖励 ¥${formatMoney(result.reward_money)}`)
    } finally {
      submitting.value = false
    }
  }

  onMounted(() => {
    loadStatus()
  })
</script>
