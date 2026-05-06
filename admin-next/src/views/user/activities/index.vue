<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section v-if="loading" class="art-card-sm px-6 py-16">
      <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap gap-3">
          <ElTag effect="plain">活动中心</ElTag>
          <ElTag effect="plain">进行中 {{ activeActivities.length }} 项</ElTag>
        </div>
        <ElButton plain :loading="loading" @click="loadActivities">刷新活动</ElButton>
      </div>

      <div class="flex justify-center">
        <ElSkeleton :rows="5" animated class="max-w-[720px] w-full" />
      </div>
    </section>

    <section v-else-if="activeActivities.length" class="grid gap-5 lg:grid-cols-2 2xl:grid-cols-3">
      <div class="lg:col-span-2 2xl:col-span-3 flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap gap-3">
          <ElTag effect="plain">活动中心</ElTag>
          <ElTag effect="plain">进行中 {{ activeActivities.length }} 项</ElTag>
          <ElTag type="primary" effect="plain">邀新活动 {{ inviteActivities.length }} 项</ElTag>
          <ElTag type="success" effect="plain">订单活动 {{ orderActivities.length }} 项</ElTag>
        </div>
        <ElButton plain :loading="loading" @click="loadActivities">刷新活动</ElButton>
      </div>

      <article
        v-for="activity in activeActivities"
        :key="activity.hid"
        class="art-card-sm overflow-hidden p-0"
      >
        <div class="border-b-d px-5 py-5">
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-start gap-3">
              <div
                class="flex h-12 w-12 shrink-0 items-center justify-center rounded-custom-sm"
                :class="
                  activity.type === '1'
                    ? 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]'
                    : 'bg-[var(--el-color-success-light-9)] text-[var(--el-color-success)]'
                "
              >
                <ElIcon :size="22">
                  <UserFilled v-if="activity.type === '1'" />
                  <Tickets v-else />
                </ElIcon>
              </div>

              <div class="min-w-0">
                <h2 class="truncate text-lg font-semibold text-g-900">{{ activity.name }}</h2>
                <p class="mt-1 text-sm text-g-500">
                  {{ activity.type === '1' ? '邀新达标后发放奖励' : '订单完成后发放奖励' }}
                </p>
              </div>
            </div>

            <div class="flex flex-col items-end gap-2">
              <ElTag :type="activity.type === '1' ? 'primary' : 'success'" effect="plain">
                {{ activity.type === '1' ? '邀新活动' : '订单活动' }}
              </ElTag>
              <ElTag :type="activity.status_ok === '1' ? 'success' : 'warning'" effect="plain">
                {{ activity.status_ok === '1' ? '已达成' : '进行中' }}
              </ElTag>
            </div>
          </div>
        </div>

        <div class="space-y-5 px-5 py-5">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">活动要求</p>
            <p class="mt-2 text-sm leading-7 text-g-700">{{ activity.yaoqiu || '以页面展示规则为准' }}</p>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <article class="rounded-custom-sm border-full-d bg-box p-4">
              <p class="text-xs font-medium text-g-400">要求数量</p>
              <p class="mt-2 text-xl font-semibold text-g-900">{{ activity.num || '-' }}</p>
            </article>
            <article class="rounded-custom-sm border-full-d bg-box p-4">
              <p class="text-xs font-medium text-g-400">奖励金额</p>
              <p class="mt-2 text-xl font-semibold text-[var(--el-color-danger)]">
                ¥{{ formatMoney(activity.money) }}
              </p>
            </article>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <article class="rounded-custom-sm border-full-d bg-box p-4">
              <p class="text-xs font-medium text-g-400">开始时间</p>
              <p class="mt-2 text-sm font-medium text-g-800">{{ activity.addtime || '-' }}</p>
            </article>
            <article class="rounded-custom-sm border-full-d bg-box p-4">
              <p class="text-xs font-medium text-g-400">结束时间</p>
              <p class="mt-2 text-sm font-medium text-g-800">{{ activity.endtime || '-' }}</p>
            </article>
          </div>
        </div>
      </article>
    </section>

    <section v-else class="art-card-sm px-6 py-16">
      <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap gap-3">
          <ElTag effect="plain">活动中心</ElTag>
          <ElTag effect="plain">进行中 0 项</ElTag>
        </div>
        <ElButton plain :loading="loading" @click="loadActivities">刷新活动</ElButton>
      </div>

      <ElEmpty description="当前暂无进行中的活动" />
    </section>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { Tickets, UserFilled } from '@element-plus/icons-vue'
  import {
    fetchLegacyPublicActivities,
    type LegacyPublicActivity
  } from '@/api/legacy/auxiliary'

  defineOptions({ name: 'UserActivitiesPage' })

  const loading = ref(false)
  const activities = ref<LegacyPublicActivity[]>([])

  const activeActivities = computed(() =>
    activities.value.filter((item) => String(item.status || '1') !== '0')
  )
  const inviteActivities = computed(() =>
    activeActivities.value.filter((item) => String(item.type || '') === '1')
  )
  const orderActivities = computed(() =>
    activeActivities.value.filter((item) => String(item.type || '') !== '1')
  )
  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const loadActivities = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyPublicActivities()
      activities.value = Array.isArray(result) ? result : []
    } catch (error) {
      console.error('[UserActivities] 加载活动失败', error)
      ElMessage.error('活动加载失败，请稍后重试')
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadActivities()
  })
</script>
