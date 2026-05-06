<template>
  <div v-if="isAdminUser" v-loading="loading">
    <ElRow :gutter="20" class="flex">
      <ElCol v-for="item in dashboardMetricCards" :key="item.title" :sm="12" :md="6" :lg="6">
        <div class="art-card relative mb-5 flex h-35 flex-col justify-center px-5 max-sm:mb-4">
          <span class="pr-16 text-sm text-g-700">{{ item.title }}</span>
          <ArtCountTo
            class="mt-2 pr-16 text-[26px] font-medium"
            :target="item.value"
            :duration="1400"
            :decimals="item.decimals"
            :prefix="item.prefix"
            :suffix="item.suffix"
            separator=","
          />
          <div class="flex-c mt-1">
            <span class="text-xs text-g-600">{{ item.noteLabel }}</span>
            <span class="ml-1 text-xs font-semibold" :class="item.noteClass">{{ item.noteValue }}</span>
          </div>
          <div
            class="absolute top-0 right-5 bottom-0 m-auto flex size-12.5 items-center justify-center rounded-xl"
            :class="item.badgeClass"
          >
            <ArtSvgIcon :icon="item.icon" class="text-xl" :class="item.iconClass" />
          </div>
        </div>
      </ElCol>
    </ElRow>

    <ElRow :gutter="20">
      <ElCol :sm="24" :md="12" :lg="10">
        <section class="art-card console-card-section mb-5 h-105 overflow-hidden p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>待处理事项</h4>
              <p>待关注<span class="text-warning">{{ attentionCount }}</span></p>
            </div>
            <ElTag size="small" effect="plain">最后刷新 {{ lastRefreshLabel }}</ElTag>
          </div>

          <div v-if="dashboardPendingItems.length" class="console-scroll-list mt-4 pr-1">
            <article
              v-for="item in dashboardPendingItems"
              :key="item.title"
              class="border-b border-[var(--default-border)] py-3 last:border-b-0"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <p class="line-clamp-1 text-sm font-medium text-g-900">{{ item.title }}</p>
                    <ElTag size="small" :type="item.tagType" effect="plain">{{ item.count }}</ElTag>
                  </div>
                  <p class="mt-1 text-xs leading-5 text-g-500">{{ item.description }}</p>
                  <div class="mt-2 flex items-center justify-between gap-3 text-xs text-g-500">
                    <span class="line-clamp-1">{{ item.footer }}</span>
                    <ElButton text type="primary" @click="goTo(item.path)">{{ item.actionText }}</ElButton>
                  </div>
                </div>
              </div>
            </article>
          </div>

          <ElEmpty v-else description="当前没有待处理事项" />
        </section>
      </ElCol>

      <ElCol :sm="24" :md="12" :lg="14">
        <section class="art-card console-card-section console-trend-card mb-5 p-4 max-sm:mb-4">
          <template v-if="trendRows.length">
            <ArtBarChart
              class="box-border px-2 pt-2"
              height="13.7rem"
              barWidth="50%"
              :showAxisLine="false"
              :showSplitLine="false"
              :data="trendChartOrders"
              :xAxisData="trendChartLabels"
            />

            <div class="mt-4 px-1">
              <h3 class="text-lg font-medium text-g-900">近 7 天经营趋势</h3>
            </div>

            <div class="console-trend-summary mt-3 grid gap-3 px-1 sm:grid-cols-2 xl:grid-cols-4">
              <article
                v-for="item in trendSummaryItems"
                :key="item.label"
                class="relative overflow-hidden rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5"
              >
                <p class="pr-14 text-[13px] text-g-500">{{ item.label }}</p>
                <p class="mt-2 pr-14 text-[22px] leading-none font-medium text-g-900">{{ item.value }}</p>
                <p class="mt-2 text-xs" :class="item.noteClass">{{ item.note }}</p>
                <div
                  class="absolute top-3.5 right-4 flex size-10 shrink-0 items-center justify-center rounded-xl"
                  :class="item.iconBoxClass"
                >
                  <ArtSvgIcon :icon="item.icon" class="text-[18px]" :class="item.iconClass" />
                </div>
              </article>
            </div>
          </template>

          <ElEmpty v-else description="暂无趋势数据" />
        </section>
      </ElCol>
    </ElRow>

    <ElRow :gutter="20">
      <ElCol :sm="24" :md="24" :lg="12">
        <section class="art-card console-card-section mb-5 h-128 overflow-hidden p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>最新订单</h4>
              <p>最近进入系统<span class="text-[var(--theme-color)]">{{ recentOrders.length }}</span></p>
            </div>
            <ElButton text type="primary" @click="goTo('/order/list')">查看全部</ElButton>
          </div>

          <div class="console-table-wrap mt-4">
            <ArtTable :data="recentOrders" :columns="recentOrderColumns" :show-table-header="true" />
          </div>
        </section>
      </ElCol>

      <ElCol :sm="24" :md="12" :lg="6">
        <section class="art-card console-card-section mb-5 h-128 overflow-hidden p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>订单状态</h4>
              <p>完成率<span class="text-success">{{ completionRate }}%</span></p>
            </div>
            <ElTag size="small" effect="plain">共 {{ totalStatusCount }} 单</ElTag>
          </div>

          <div v-if="statusRows.length" class="console-scroll-list mt-4 pr-1">
            <article
              v-for="item in statusRows"
              :key="item.status"
              class="border-b border-[var(--default-border)] py-3 last:border-b-0"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <ElTag size="small" :type="orderStatusTagType(item.status)" effect="plain">{{ item.status }}</ElTag>
                    <span class="text-sm font-semibold text-g-900">{{ item.count }}</span>
                  </div>
                  <div class="mt-2 flex items-center justify-between gap-3 text-xs text-g-500">
                    <span>占比 {{ statusPercent(item.count) }}%</span>
                    <span>{{ totalStatusCount ? `${item.count}/${totalStatusCount}` : '0/0' }}</span>
                  </div>
                  <ElProgress
                    class="mt-2"
                    :percentage="statusPercent(item.count)"
                    :stroke-width="8"
                    :show-text="false"
                  />
                </div>
              </div>
            </article>
          </div>

          <ElEmpty v-else description="暂无状态统计" />
        </section>
      </ElCol>

      <ElCol :sm="24" :md="12" :lg="6">
        <section class="art-card console-card-section mb-5 h-128 overflow-hidden p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>高消费用户</h4>
              <p>近 7 天 Top<span class="text-warning">{{ topUserRows.length }}</span></p>
            </div>
            <ElTag size="small" effect="plain">最近 7 天</ElTag>
          </div>

          <div v-if="topUserRows.length" class="console-scroll-list mt-4 pr-1">
            <article
              v-for="item in topUserRows"
              :key="item.uid"
              class="border-b border-[var(--default-border)] py-3 last:border-b-0"
            >
              <div class="flex items-start gap-3">
                <ElAvatar class="console-top-user-card__avatar" :size="40" :src="item.avatar">
                  {{ item.avatarInitial }}
                </ElAvatar>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-3">
                    <p class="line-clamp-1 text-sm font-medium text-g-900">{{ item.maskedName }}</p>
                    <ElTag :type="item.rank <= 3 ? 'warning' : 'info'" size="small" :effect="item.rank <= 3 ? 'dark' : 'plain'">
                      TOP {{ item.rank }}
                    </ElTag>
                  </div>
                  <p class="mt-1 text-xs text-g-500">近 7 天订单 {{ item.orders }}</p>
                  <div class="mt-2 flex items-center justify-between gap-3 text-xs text-g-500">
                    <span>消费金额</span>
                    <span class="font-semibold text-[var(--el-color-primary)]">{{ moneyLabel(item.total) }}</span>
                  </div>
                </div>
              </div>
            </article>
          </div>

          <ElEmpty v-else description="暂无高消费用户" />
        </section>
      </ElCol>
    </ElRow>

    <section class="art-card mb-5 overflow-hidden p-5 max-sm:mb-4">
      <div class="console-overview">
        <div>
          <div class="art-card-header console-overview__header">
            <div class="title">
              <h4>运行概览</h4>
              <p>保留真实业务入口与系统状态</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <ElButton size="small" plain :loading="loading" @click="refreshDashboard">刷新看板</ElButton>
              <ElButton size="small" plain :type="autoRefresh ? 'primary' : 'default'" @click="toggleAutoRefresh">
                {{ autoRefresh ? '自动刷新中' : '开启自动刷新' }}
              </ElButton>
              <ElButton size="small" type="primary" plain @click="goTo('/admin/tickets')">工单中心</ElButton>
              <ElButton size="small" type="primary" plain @click="goTo('/admin/chat')">客服会话</ElButton>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <ElTag size="small" :type="autoRefresh ? 'success' : 'info'" effect="plain">
              {{ autoRefresh ? '自动刷新开启' : '自动刷新关闭' }}
            </ElTag>
            <ElTag size="small" :type="scheduler?.running ? 'warning' : 'primary'" effect="plain">
              调度 {{ scheduler?.running ? '运行中' : '空闲' }}
            </ElTag>
            <ElTag size="small" :type="syncStatusType" effect="plain">进度同步 {{ syncStatusLabel }}</ElTag>
            <ElTag v-if="failedSections.length" size="small" type="warning" effect="plain">
              部分接口延迟 {{ failedSections.length }} 项
            </ElTag>
          </div>

          <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <article class="relative overflow-hidden rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5">
              <p class="pr-14 text-[13px] text-g-500">订单总量</p>
              <p class="mt-2 pr-14 text-[22px] leading-none font-medium text-g-900">{{ stats?.total_orders || 0 }}</p>
              <p class="mt-2 text-xs text-g-500">已完成 {{ stats?.completed_orders || 0 }}</p>
              <div class="absolute top-3.5 right-4 flex size-10 items-center justify-center rounded-xl bg-[var(--el-color-primary-light-9)]">
                <ArtSvgIcon icon="ri:file-list-3-line" class="text-[18px] text-[var(--el-color-primary)]" />
              </div>
            </article>
            <article class="relative overflow-hidden rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5">
              <p class="pr-14 text-[13px] text-g-500">进行中订单</p>
              <p class="mt-2 pr-14 text-[22px] leading-none font-medium text-g-900">{{ stats?.processing_orders || 0 }}</p>
              <p class="mt-2 text-xs text-g-500">等待回执或继续同步</p>
              <div class="absolute top-3.5 right-4 flex size-10 items-center justify-center rounded-xl bg-[var(--el-color-warning-light-9)]">
                <ArtSvgIcon icon="ri:loader-4-line" class="text-[18px] text-[var(--el-color-warning)]" />
              </div>
            </article>
            <article class="relative overflow-hidden rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5">
              <p class="pr-14 text-[13px] text-g-500">待审核提现</p>
              <p class="mt-2 pr-14 text-[22px] leading-none font-medium text-g-900">{{ pendingReviewCount }}</p>
              <p class="mt-2 text-xs text-g-500">普通 {{ pendingWithdrawCount }} / 商城 {{ pendingMallWithdrawCount }}</p>
              <div class="absolute top-3.5 right-4 flex size-10 items-center justify-center rounded-xl bg-[var(--el-color-danger-light-9)]">
                <ArtSvgIcon icon="ri:wallet-3-line" class="text-[18px] text-[var(--el-color-danger)]" />
              </div>
            </article>
            <article class="relative overflow-hidden rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5">
              <p class="pr-14 text-[13px] text-g-500">用户总余额</p>
              <p class="mt-2 pr-14 text-[22px] leading-none font-medium text-g-900">{{ moneyLabel(stats?.total_balance) }}</p>
              <p class="mt-2 text-xs text-g-500">结合提现审核关注</p>
              <div class="absolute top-3.5 right-4 flex size-10 items-center justify-center rounded-xl bg-[var(--el-color-success-light-9)]">
                <ArtSvgIcon icon="ri:money-cny-circle-line" class="text-[18px] text-[var(--el-color-success)]" />
              </div>
            </article>
          </div>

          <div class="mt-4 grid gap-3 xl:grid-cols-2">
            <article class="rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">对接调度</p>
                  <p class="mt-1 text-xs text-g-500">触发来源 {{ scheduler?.last_trigger || '暂无' }}</p>
                </div>
                <ElTag size="small" :type="scheduler?.running ? 'warning' : 'success'" effect="plain">
                  {{ scheduler?.running ? '运行中' : '空闲' }}
                </ElTag>
              </div>
              <div class="mt-3 flex items-center justify-between gap-3 text-xs text-g-500">
                <span>待对接 {{ scheduler?.pending || 0 }} / 成功 {{ scheduler?.last_success || 0 }}</span>
                <ElButton text type="primary" @click="goTo('/admin/queue')">查看调度</ElButton>
              </div>
            </article>
            <article class="rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3.5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">订单进度同步</p>
                  <p class="mt-1 text-xs text-g-500">批量模式 {{ progressSync?.batch_enabled ? '开启' : '关闭' }}</p>
                </div>
                <ElTag size="small" :type="syncStatusType" effect="plain">{{ syncStatusLabel }}</ElTag>
              </div>
              <div class="mt-3 flex items-center justify-between gap-3 text-xs text-g-500">
                <span>更新 {{ progressUpdatedCount }} / 失败 {{ progressFailedCount }}</span>
                <ElButton text type="primary" @click="goTo('/admin/order-progress-sync')">查看同步</ElButton>
              </div>
            </article>
          </div>
        </div>

        <section class="console-announcement-panel">
          <div class="art-card-header items-start">
            <div class="title">
              <h4>公告通知</h4>
              <p>同步后台公告与置顶状态</p>
            </div>
            <ElButton text type="primary" @click="goTo('/admin/announcements')">查看全部</ElButton>
          </div>

          <div v-if="announcements.length" class="console-scroll-list mt-4 pr-1">
            <article
              v-for="item in announcements"
              :key="item.id"
              class="border-b border-[var(--default-border)] py-3 last:border-b-0"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0 flex-1">
                  <p class="line-clamp-1 text-sm font-medium text-g-900">
                    {{ item.title || `公告 #${item.id}` }}
                  </p>
                  <p class="mt-1 text-xs text-g-500">{{ item.author || '系统' }} · {{ item.time || '-' }}</p>
                  <p class="mt-2 line-clamp-2 text-xs leading-5 text-g-500">
                    {{ item.content || '暂无公告摘要' }}
                  </p>
                </div>
                <ElTag size="small" :type="item.zhiding === '1' ? 'warning' : 'info'" effect="plain">
                  {{ item.zhiding === '1' ? '置顶' : '普通' }}
                </ElTag>
              </div>
            </article>
          </div>

          <ElEmpty v-else description="暂无公告" />
        </section>
      </div>
    </section>
  </div>

  <div v-else v-loading="userLoading" class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card p-5">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-g-900">{{ userDisplayName }}</h2>
          <p class="mt-1 text-sm text-g-500">UID {{ userProfile?.uid || userStore.info.userId || '-' }} / 账号 {{ userProfile?.user || userStore.info.username || '-' }}</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <ElTag v-if="userFailures.length" type="warning" effect="plain">部分数据延迟 {{ userFailures.length }} 项</ElTag>
        </div>
      </div>

      <div class="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <article
          v-for="item in userMetricCards"
          :key="item.title"
          class="relative overflow-hidden rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-4"
        >
          <p class="pr-14 text-[13px] text-g-500">{{ item.title }}</p>
          <p class="mt-2 pr-14 text-[24px] leading-none font-medium text-g-900">{{ item.value }}</p>
          <p class="mt-2 text-xs text-g-500">{{ item.note }}</p>
          <div
            class="absolute top-4 right-4 flex size-10 items-center justify-center rounded-xl"
            :class="item.iconBoxClass"
          >
            <ArtSvgIcon :icon="item.icon" class="text-[18px]" :class="item.iconClass" />
          </div>
        </article>
      </div>

      <div class="mt-5 flex flex-wrap gap-3">
        <ElButton
          v-for="item in userQuickActions"
          :key="item.path"
          :type="item.primary ? 'primary' : 'default'"
          plain
          @click="goTo(item.path)"
        >
          {{ item.label }}
        </ElButton>
      </div>
    </section>

    <ElRow :gutter="20">
      <ElCol :sm="24" :lg="14">
        <section class="art-card console-card-section mb-5 overflow-hidden p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>我的最近订单</h4>
              <p>最近提交<span class="text-[var(--theme-color)]">{{ userOrders.length }}</span></p>
            </div>
            <ElButton text type="primary" @click="goTo('/order/list')">查看全部</ElButton>
          </div>

          <div class="mt-4">
            <ArtTable :data="userOrders" :columns="userOrderColumns" :show-table-header="true" />
          </div>
        </section>
      </ElCol>

      <ElCol :sm="24" :lg="10">
        <section class="art-card console-card-section mb-5 p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>资金与工单</h4>
              <p>余额变动和反馈状态</p>
            </div>
            <ElButton text type="primary" @click="goTo('/user/recharge')">充值</ElButton>
          </div>

          <div class="mt-4 grid gap-4 xl:grid-cols-2">
            <div>
              <div class="mb-3 flex items-center justify-between gap-3">
                <p class="text-sm font-semibold text-g-900">最近资金变动</p>
                <ElButton text type="primary" @click="goTo('/user/moneylog')">明细</ElButton>
              </div>
              <div v-if="userMoneyLogs.length" class="space-y-3">
                <article
                  v-for="item in userMoneyLogs"
                  :key="item.id"
                  class="rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3"
                >
                  <div class="flex items-center justify-between gap-3">
                    <div class="min-w-0">
                      <p class="line-clamp-1 text-sm font-medium text-g-900">{{ item.type || '余额变动' }}</p>
                      <p class="mt-1 text-xs text-g-500">{{ item.addtime || '-' }}</p>
                    </div>
                    <span
                      class="shrink-0 text-sm font-semibold"
                      :class="Number(item.money || 0) >= 0 ? 'text-[var(--el-color-success)]' : 'text-[var(--el-color-danger)]'"
                    >
                      {{ Number(item.money || 0) >= 0 ? '+' : '' }}{{ Number(item.money || 0).toFixed(2) }}
                    </span>
                  </div>
                </article>
              </div>
              <ElEmpty v-else description="暂无资金变动" />
            </div>

            <div>
              <div class="mb-3 flex items-center justify-between gap-3">
                <p class="text-sm font-semibold text-g-900">我的工单</p>
                <ElButton text type="primary" @click="goTo('/user/ticket')">工单中心</ElButton>
              </div>
              <div v-if="userTickets.length" class="space-y-3">
                <article
                  v-for="item in userTickets"
                  :key="item.id"
                  class="rounded-custom-sm border-full-d bg-[var(--default-box-color)] px-4 py-3"
                >
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <p class="line-clamp-1 text-sm font-medium text-g-900">{{ item.type || `工单 #${item.id}` }}</p>
                      <p class="mt-1 line-clamp-1 text-xs text-g-500">{{ item.content || '-' }}</p>
                    </div>
                    <ElTag size="small" :type="item.status === 1 ? 'success' : item.status === 2 ? 'info' : 'warning'" effect="plain">
                      {{ item.status === 1 ? '已回复' : item.status === 2 ? '已关闭' : '处理中' }}
                    </ElTag>
                  </div>
                </article>
              </div>
              <ElEmpty v-else description="暂无工单" />
            </div>
          </div>
        </section>
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, onUnmounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElAvatar, ElTag } from 'element-plus'
  import defaultAvatar from '@/assets/images/user/avatar.webp'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useUserStore } from '@/store/modules/user'
  import {
    fetchLegacyAdminDashboardStats,
    type LegacyDashboardRecentOrder,
    type LegacyDashboardStats
  } from '@/api/legacy/admin-dashboard'
  import { fetchLegacyOrderList, type LegacyOrderItem } from '@/api/legacy/order'
  import {
    fetchLegacyMoneyLogs,
    fetchLegacyUserProfile,
    type LegacyMoneyLog,
    type LegacyUserProfile
  } from '@/api/legacy/user-center'
  import {
    fetchLegacyAdminTicketStats,
    fetchLegacyUserTickets,
    type LegacyTicket,
    type LegacyTicketStats
  } from '@/api/legacy/ticket'
  import {
    fetchLegacyAdminChatSessions,
    fetchLegacyAdminChatStats,
    type LegacyAdminChatSession,
    type LegacyAdminChatStats
  } from '@/api/legacy/admin-chat'
  import {
    fetchLegacyDockSchedulerStats,
    fetchLegacyOrderProgressSyncStatus,
    type LegacyDockSchedulerStats,
    type LegacyOrderProgressSyncStatus
  } from '@/api/legacy/admin-sync'
  import {
    fetchLegacyAdminWithdrawRequests,
    fetchLegacyAdminMallCUserWithdrawRequests
  } from '@/api/legacy/admin-stats'
  import {
    fetchLegacyAdminAnnouncements,
    type LegacyAdminAnnouncement
  } from '@/api/legacy/admin-content'

  defineOptions({ name: 'Console' })

  const router = useRouter()
  const userStore = useUserStore()

  const loading = ref(false)
  const userLoading = ref(false)
  const autoRefresh = ref(false)
  const intervalId = ref<number | null>(null)
  const lastRefresh = ref('')
  const failedSections = ref<string[]>([])
  const userFailures = ref<string[]>([])

  const stats = ref<LegacyDashboardStats | null>(null)
  const ticketStats = ref<LegacyTicketStats | null>(null)
  const chatStats = ref<LegacyAdminChatStats | null>(null)
  const chatSessions = ref<LegacyAdminChatSession[]>([])
  const scheduler = ref<LegacyDockSchedulerStats | null>(null)
  const progressSync = ref<LegacyOrderProgressSyncStatus | null>(null)
  const announcements = ref<LegacyAdminAnnouncement[]>([])
  const pendingWithdrawCount = ref(0)
  const pendingMallWithdrawCount = ref(0)
  const userProfile = ref<LegacyUserProfile | null>(null)
  const userOrders = ref<LegacyOrderItem[]>([])
  const userMoneyLogs = ref<LegacyMoneyLog[]>([])
  const userTickets = ref<LegacyTicket[]>([])

  const isAdminUser = computed(() => {
    const roles = userStore.info?.roles || []
    return roles.includes('R_SUPER') || roles.includes('R_ADMIN')
  })
  const userDisplayName = computed(
    () => userProfile.value?.name || userProfile.value?.user || userStore.info.realName || userStore.info.username || '我的首页'
  )
  const trendRows = computed(() => stats.value?.trend || [])
  const statusRows = computed(() => stats.value?.status_distribution || [])
  const recentOrders = computed(() => stats.value?.recent_orders || [])
  const topUserRows = computed(() =>
    (stats.value?.top_users || []).map((item, index) => ({
      ...item,
      rank: index + 1,
      total: Number(item.total || 0),
      orders: Number(item.orders || 0),
      avatar: getAvatarUrl(item.username),
      avatarInitial: getAvatarInitial(item.username),
      maskedName: maskTopUserName(item.username, index + 1)
    }))
  )

  const recentProcessingOrderCount = computed(
    () => recentOrders.value.filter((item) => ['进行中', '处理中'].includes(String(item.status || ''))).length
  )
  const recentFailedOrderCount = computed(
    () => recentOrders.value.filter((item) => ['异常', '失败'].includes(String(item.status || ''))).length
  )
  const maxTrendOrders = computed(() => {
    const values = trendRows.value.map((item) => Number(item.orders || 0))
    return values.length ? Math.max(...values, 1) : 1
  })
  const maxTrendIncome = computed(() => {
    const values = trendRows.value.map((item) => Number(item.income || 0))
    return values.length ? Math.max(...values, 0) : 0
  })
  const weekOrders = computed(() =>
    trendRows.value.reduce((sum, item) => sum + Number(item.orders || 0), 0)
  )
  const weekIncome = computed(() =>
    trendRows.value.reduce((sum, item) => sum + Number(item.income || 0), 0)
  )
  const activeTrendDays = computed(() =>
    trendRows.value.filter((item) => Number(item.orders || 0) > 0).length
  )
  const lastRefreshLabel = computed(() => lastRefresh.value || '未刷新')
  const completionRate = computed(() => {
    const total = Number(stats.value?.total_orders || 0)
    const completed = Number(stats.value?.completed_orders || 0)
    if (!total) return 0
    return Math.round((completed / total) * 100)
  })
  const ordersDiff = computed(() => {
    const today = Number(stats.value?.today_orders || 0)
    const yesterday = Number(stats.value?.yesterday_orders || 0)
    if (!yesterday) return today > 0 ? 100 : 0
    return Math.round(((today - yesterday) / yesterday) * 100)
  })
  const incomeDiff = computed(() => {
    const today = Number(stats.value?.today_income || 0)
    const yesterday = Number(stats.value?.yesterday_income || 0)
    if (!yesterday) return today > 0 ? 100 : 0
    return Math.round(((today - yesterday) / yesterday) * 100)
  })
  const totalStatusCount = computed(() =>
    statusRows.value.reduce((sum, item) => sum + Number(item.count || 0), 0)
  )
  const unreadSessionCount = computed(() =>
    chatSessions.value.filter((item) => Number(item.unread_count || 0) > 0).length
  )
  const pendingReviewCount = computed(
    () => pendingWithdrawCount.value + pendingMallWithdrawCount.value
  )
  const progressFailedCount = computed(
    () => Number(progressSync.value?.last_failed || 0) + Number(progressSync.value?.batch_last_failed || 0)
  )
  const progressUpdatedCount = computed(
    () => Number(progressSync.value?.last_updated || 0) + Number(progressSync.value?.batch_last_updated || 0)
  )
  const attentionCount = computed(
    () =>
      Number(ticketStats.value?.pending || 0) +
      unreadSessionCount.value +
      pendingReviewCount.value +
      progressFailedCount.value
  )
  const syncStatusLabel = computed(() => {
    if (!progressSync.value) return '未知'
    if (progressSync.value.running || progressSync.value.batch_running) return '运行中'
    if (progressSync.value.enabled || progressSync.value.batch_enabled) return '已启用'
    return '未启用'
  })
  const dashboardPendingItems = computed(() => {
    const items = [] as Array<{
      title: string
      description: string
      count: number
      footer: string
      path: string
      actionText: string
      tagType: 'warning' | 'danger'
    }>

    if (Number(ticketStats.value?.pending || 0) > 0) {
      items.push({
        title: '工单待处理',
        description: `上游待同步 ${ticketStats.value?.upstream_pending || 0}，需优先跟进异常与回复。`,
        count: Number(ticketStats.value?.pending || 0),
        footer: '待回复工单',
        path: '/admin/tickets',
        actionText: '立即处理',
        tagType: 'warning'
      })
    }

    if (unreadSessionCount.value > 0) {
      items.push({
        title: '客服未读会话',
        description: `总会话 ${chatStats.value?.session_count || 0}，未读消息需尽快回捞。`,
        count: unreadSessionCount.value,
        footer: `消息 ${chatStats.value?.msg_count || 0}`,
        path: '/admin/chat',
        actionText: '查看会话',
        tagType: 'danger'
      })
    }

    if (pendingWithdrawCount.value > 0) {
      items.push({
        title: '普通提现审核',
        description: '普通用户提现请求待审核，避免资金链路滞留。',
        count: pendingWithdrawCount.value,
        footer: '待审核提现',
        path: '/admin/withdraw',
        actionText: '前往审核',
        tagType: 'warning'
      })
    }

    if (pendingMallWithdrawCount.value > 0) {
      items.push({
        title: '商城提现审核',
        description: '租户侧商城用户提现单独处理，避免漏单。',
        count: pendingMallWithdrawCount.value,
        footer: '租户提现链路',
        path: '/admin/mall-cuser-withdraw',
        actionText: '前往审核',
        tagType: 'warning'
      })
    }

    if (Number(scheduler.value?.pending || 0) > 0) {
      items.push({
        title: '待对接调度',
        description: `上次执行 ${scheduler.value?.last_run_time || '暂无'}，持续关注积压情况。`,
        count: Number(scheduler.value?.pending || 0),
        footer: `成功 ${scheduler.value?.last_success || 0} / 批量 ${scheduler.value?.batch_limit || 0}`,
        path: '/admin/queue',
        actionText: '查看调度',
        tagType: 'warning'
      })
    }

    if (progressFailedCount.value > 0) {
      items.push({
        title: '订单进度同步',
        description: `下次运行 ${progressSync.value?.next_run_time || '未计划'}，当前存在失败记录需继续处理。`,
        count: progressFailedCount.value,
        footer: `更新 ${progressUpdatedCount.value} / 失败 ${progressFailedCount.value}`,
        path: '/admin/order-progress-sync',
        actionText: '查看同步',
        tagType: 'danger'
      })
    }

    return items
  })

  const syncStatusType = computed<'success' | 'warning' | 'danger' | 'info'>(() => {
    if (!progressSync.value) return 'info'
    if (progressFailedCount.value > 0) return 'danger'
    if (progressSync.value.running || progressSync.value.batch_running) return 'warning'
    if (progressSync.value.enabled || progressSync.value.batch_enabled) return 'success'
    return 'info'
  })
  const trendChartLabels = computed(() => trendRows.value.map((item) => formatTrendDateLabel(item.date)))
  const trendChartOrders = computed(() => trendRows.value.map((item) => Number(item.orders || 0)))
  const trendPeakRow = computed(() => {
    if (!trendRows.value.length) return null
    return trendRows.value.reduce((max, item) =>
      Number(item.orders || 0) > Number(max.orders || 0) ? item : max
    )
  })
  const trendPeakDayLabel = computed(() => formatTrendDateLabel(trendPeakRow.value?.date))
  const userOpenTicketCount = computed(() =>
    userTickets.value.filter((item) => ![1, 2].includes(Number(item.status))).length
  )

  const userMetricCards = computed(() => [
    {
      title: '账户余额',
      value: moneyLabel(userProfile.value?.money),
      note: `储值 ${moneyLabel(userProfile.value?.cdmoney)}`,
      icon: 'ri:wallet-3-line',
      iconBoxClass: 'bg-[var(--el-color-primary-light-9)]',
      iconClass: 'text-[var(--el-color-primary)]'
    },
    {
      title: '今日订单',
      value: `${Number(userProfile.value?.today_orders || 0)}`,
      note: `累计 ${Number(userProfile.value?.order_total || 0)} 单`,
      icon: 'ri:file-list-3-line',
      iconBoxClass: 'bg-[var(--el-color-success-light-9)]',
      iconClass: 'text-[var(--el-color-success)]'
    },
    {
      title: '今日消费',
      value: moneyLabel(userProfile.value?.today_spend),
      note: `总充值 ${moneyLabel(userProfile.value?.zcz)}`,
      icon: 'ri:coins-line',
      iconBoxClass: 'bg-[var(--el-color-warning-light-9)]',
      iconClass: 'text-[var(--el-color-warning)]'
    },
    {
      title: '待跟进工单',
      value: `${userOpenTicketCount.value}`,
      note: `最近工单 ${userTickets.value.length} 条`,
      icon: 'ri:customer-service-2-line',
      iconBoxClass: 'bg-[var(--el-color-info-light-9)]',
      iconClass: 'text-[var(--el-color-info)]'
    }
  ])

  const userQuickActions = [
    { label: '查课交单', path: '/order/add', primary: true },
    { label: '批量交单', path: '/order/batch-add', primary: false },
    { label: '订单列表', path: '/order/list', primary: false },
    { label: '余额充值', path: '/user/recharge', primary: false },
    { label: '提交工单', path: '/user/ticket', primary: false },
    { label: '我的资料', path: '/user/profile', primary: false }
  ]

  const getAvatarUrl = (account?: string) => {
    const value = String(account || '').trim()
    return value ? `//q2.qlogo.cn/headimg_dl?dst_uin=${value}&spec=640` : defaultAvatar
  }

  const getAvatarInitial = (text?: string) => {
    const value = String(text || '').trim()
    return (value || 'U').slice(0, 1).toUpperCase()
  }

  const maskTopUserName = (username?: string, rank?: number) => {
    const label = String(username || '').trim()

    if (!label) {
      return `高消费用户 ${rank || '-'}`
    }

    if (label.length === 1) {
      return `${label}**`
    }

    if (label.length === 2) {
      return `${label.slice(0, 1)}*`
    }

    return `${label.slice(0, 1)}${'*'.repeat(Math.min(3, label.length - 2))}${label.slice(-1)}`
  }

  const moneyLabel = (value?: number) => `¥${Number(value || 0).toFixed(2)}`

  const formatTrendDateLabel = (value?: string) => {
    const normalized = String(value || '').trim().split('T')[0].replace(/\//g, '-')

    if (/^\d{4}-\d{2}-\d{2}$/.test(normalized)) {
      return normalized.slice(5)
    }

    return normalized || '-'
  }

  const formatDelta = (value: number) => `${value >= 0 ? '+' : ''}${value}%`

  const metricDeltaClass = (value: number) => {
    if (value > 0) return 'text-success'
    if (value < 0) return 'text-danger'
    return 'text-g-500'
  }

  const dashboardMetricCards = computed(() => [
    {
      title: '今日订单',
      value: Number(stats.value?.today_orders || 0),
      decimals: 0,
      prefix: '',
      suffix: '',
      noteLabel: '较昨日',
      noteValue: formatDelta(ordersDiff.value),
      noteClass: metricDeltaClass(ordersDiff.value),
      icon: 'ri:clipboard-line',
      badgeClass: 'console-metric-card__badge--primary',
      iconClass: 'text-[var(--el-color-primary)]'
    },
    {
      title: '今日收入',
      value: Number(stats.value?.today_income || 0),
      decimals: 2,
      prefix: '¥',
      suffix: '',
      noteLabel: '较昨日',
      noteValue: formatDelta(incomeDiff.value),
      noteClass: metricDeltaClass(incomeDiff.value),
      icon: 'ri:coins-line',
      badgeClass: 'console-metric-card__badge--success',
      iconClass: 'text-[var(--el-color-success)]'
    },
    {
      title: '用户总数',
      value: Number(stats.value?.user_count || 0),
      decimals: 0,
      prefix: '',
      suffix: '',
      noteLabel: '近 7 天活跃',
      noteValue: `${activeTrendDays.value} 天`,
      noteClass: 'text-[var(--el-color-primary)]',
      icon: 'ri:team-line',
      badgeClass: 'console-metric-card__badge--warning',
      iconClass: 'text-[var(--el-color-warning)]'
    },
    {
      title: '今日新增',
      value: Number(stats.value?.today_new_users || 0),
      decimals: 0,
      prefix: '',
      suffix: '',
      noteLabel: '全站用户',
      noteValue: `${Number(stats.value?.user_count || 0)}`,
      noteClass: 'text-[var(--el-color-warning)]',
      icon: 'ri:user-add-line',
      badgeClass: 'console-metric-card__badge--info',
      iconClass: 'text-[var(--el-color-warning-dark-2)]'
    }
  ])

  const trendSummaryItems = computed(() => [
    {
      label: '近 7 天订单',
      value: `${weekOrders.value}`,
      note: `订单峰值 ${maxTrendOrders.value} 单`,
      noteClass: 'text-xs text-g-500',
      icon: 'ri:shopping-cart-2-line',
      iconBoxClass: 'bg-[var(--el-color-primary-light-9)]',
      iconClass: 'text-[var(--el-color-primary)]'
    },
    {
      label: '近 7 天收入',
      value: moneyLabel(weekIncome.value),
      note: `最高日收入 ${moneyLabel(maxTrendIncome.value)}`,
      noteClass: 'text-xs text-[var(--el-color-success)]',
      icon: 'ri:money-cny-circle-line',
      iconBoxClass: 'bg-[var(--el-color-success-light-9)]',
      iconClass: 'text-[var(--el-color-success)]'
    },
    {
      label: '活跃天数',
      value: `${activeTrendDays.value} 天`,
      note: '按有订单统计',
      noteClass: 'text-xs text-g-500',
      icon: 'ri:calendar-check-line',
      iconBoxClass: 'bg-[var(--el-color-warning-light-9)]',
      iconClass: 'text-[var(--el-color-warning)]'
    },
    {
      label: '峰值日期',
      value: trendPeakDayLabel.value,
      note: '最近 7 天订单最高日',
      noteClass: 'text-xs text-[var(--el-color-info)]',
      icon: 'ri:medal-line',
      iconBoxClass: 'bg-[var(--el-color-info-light-9)]',
      iconClass: 'text-[var(--el-color-info)]'
    }
  ])

  const orderStatusTagType = (status: string): 'success' | 'warning' | 'danger' | 'info' => {
    if (['已完成'].includes(status)) return 'success'
    if (['进行中', '处理中'].includes(status)) return 'warning'
    if (['异常', '失败'].includes(status)) return 'danger'
    return 'info'
  }

  const statusPercent = (count: number) => {
    if (!totalStatusCount.value) return 0
    return Math.round((Number(count || 0) / totalStatusCount.value) * 100)
  }

  const goTo = (path: string) => {
    router.push(path)
  }

  const loadDashboard = async () => {
    loading.value = true
    try {
      const [
        dashboardResult,
        ticketResult,
        chatStatsResult,
        chatSessionsResult,
        withdrawResult,
        mallWithdrawResult,
        schedulerResult,
        progressResult,
        announcementResult
      ] = await Promise.allSettled([
        fetchLegacyAdminDashboardStats(),
        fetchLegacyAdminTicketStats(),
        fetchLegacyAdminChatStats(),
        fetchLegacyAdminChatSessions(),
        fetchLegacyAdminWithdrawRequests({ page: 1, limit: 1, status: '0' }),
        fetchLegacyAdminMallCUserWithdrawRequests({ page: 1, limit: 1, status: '0' }),
        fetchLegacyDockSchedulerStats(),
        fetchLegacyOrderProgressSyncStatus(),
        fetchLegacyAdminAnnouncements({ page: 1, limit: 5 })
      ])

      const failures: string[] = []

      if (dashboardResult.status === 'fulfilled') {
        stats.value = dashboardResult.value
      } else {
        stats.value = null
        failures.push('经营数据')
      }

      if (ticketResult.status === 'fulfilled') {
        ticketStats.value = ticketResult.value
      } else {
        ticketStats.value = null
        failures.push('工单统计')
      }

      if (chatStatsResult.status === 'fulfilled') {
        chatStats.value = chatStatsResult.value
      } else {
        chatStats.value = null
        failures.push('聊天统计')
      }

      if (chatSessionsResult.status === 'fulfilled') {
        chatSessions.value = chatSessionsResult.value || []
      } else {
        chatSessions.value = []
        failures.push('聊天会话')
      }

      if (withdrawResult.status === 'fulfilled') {
        pendingWithdrawCount.value = Number(withdrawResult.value?.pagination?.total || 0)
      } else {
        pendingWithdrawCount.value = 0
        failures.push('普通提现')
      }

      if (mallWithdrawResult.status === 'fulfilled') {
        pendingMallWithdrawCount.value = Number(mallWithdrawResult.value?.pagination?.total || 0)
      } else {
        pendingMallWithdrawCount.value = 0
        failures.push('商城提现')
      }

      if (schedulerResult.status === 'fulfilled') {
        scheduler.value = schedulerResult.value
      } else {
        scheduler.value = null
        failures.push('对接调度')
      }

      if (progressResult.status === 'fulfilled') {
        progressSync.value = progressResult.value
      } else {
        progressSync.value = null
        failures.push('进度同步')
      }

      if (announcementResult.status === 'fulfilled') {
        announcements.value = announcementResult.value?.list || []
      } else {
        announcements.value = []
        failures.push('公告')
      }

      failedSections.value = failures
      lastRefresh.value = new Date().toLocaleString('zh-CN', { hour12: false })
    } finally {
      loading.value = false
    }
  }

  const refreshDashboard = async () => {
    await loadDashboard()
  }

  const loadUserDashboard = async () => {
    userLoading.value = true
    try {
      const [profileResult, ordersResult, moneyResult, ticketsResult] = await Promise.allSettled([
        fetchLegacyUserProfile(),
        fetchLegacyOrderList({ page: 1, limit: 6 }),
        fetchLegacyMoneyLogs({ page: 1, limit: 5 }),
        fetchLegacyUserTickets(1, 5)
      ])

      const failures: string[] = []

      if (profileResult.status === 'fulfilled') {
        userProfile.value = profileResult.value
      } else {
        userProfile.value = null
        failures.push('账户资料')
      }

      if (ordersResult.status === 'fulfilled') {
        userOrders.value = ordersResult.value?.list || []
      } else {
        userOrders.value = []
        failures.push('最近订单')
      }

      if (moneyResult.status === 'fulfilled') {
        userMoneyLogs.value = moneyResult.value?.list || []
      } else {
        userMoneyLogs.value = []
        failures.push('资金明细')
      }

      if (ticketsResult.status === 'fulfilled') {
        userTickets.value = ticketsResult.value?.list || []
      } else {
        userTickets.value = []
        failures.push('工单')
      }

      userFailures.value = failures
    } finally {
      userLoading.value = false
    }
  }

  const stopAutoRefresh = () => {
    if (intervalId.value !== null) {
      window.clearInterval(intervalId.value)
      intervalId.value = null
    }
    autoRefresh.value = false
  }

  const toggleAutoRefresh = () => {
    if (autoRefresh.value) {
      stopAutoRefresh()
      return
    }
    autoRefresh.value = true
    intervalId.value = window.setInterval(() => {
      loadDashboard()
    }, 60000)
  }

  const { columns: recentOrderColumns } = useTableColumns<LegacyDashboardRecentOrder>(() => [
    { prop: 'oid', label: '订单号', width: 90 },
    {
      prop: 'ptname',
      label: '订单信息',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.ptname || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500 line-clamp-1' }, row.kcname || '未记录课程')
        ])
    },
    { prop: 'user', label: '账号', width: 150 },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row) => h(ElTag, { type: orderStatusTagType(row.status), effect: 'plain' }, () => row.status || '-')
    },
    {
      prop: 'fees',
      label: '费用',
      width: 120,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, moneyLabel(row.fees))
    },
    { prop: 'addtime', label: '时间', width: 180 }
  ])

  const { columns: userOrderColumns } = useTableColumns<LegacyOrderItem>(() => [
    { prop: 'oid', label: '订单号', width: 90 },
    {
      prop: 'ptname',
      label: '订单信息',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.ptname || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500 line-clamp-1' }, row.kcname || '未记录课程')
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row) => h(ElTag, { type: orderStatusTagType(row.status), effect: 'plain' }, () => row.status || '-')
    },
    {
      prop: 'fees',
      label: '费用',
      width: 120,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, moneyLabel(Number(row.fees || 0)))
    },
    { prop: 'addtime', label: '时间', width: 180 }
  ])

  onMounted(() => {
    if (isAdminUser.value) {
      loadDashboard()
    } else {
      loadUserDashboard()
    }
  })

  onUnmounted(() => {
    stopAutoRefresh()
  })
</script>

<style scoped>
  .console-card-section {
    box-sizing: border-box;
  }

  .console-scroll-list {
    max-height: calc(100% - 56px);
    overflow: auto;
  }

  .console-scroll-list :deep(.el-tag) {
    flex-shrink: 0;
  }

  .console-table-wrap {
    height: calc(100% - 56px);
    overflow: hidden;
  }

  .console-table-wrap :deep(.art-table) {
    border-radius: 16px;
    overflow: hidden;
  }

  .console-trend-card {
    min-height: 420px;
  }

  .console-trend-summary {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .console-overview {
    display: grid;
    gap: 20px;
    align-items: start;
    grid-template-columns: minmax(0, 1fr) 360px;
  }

  .console-overview__header {
    align-items: flex-start;
  }

  .console-announcement-panel {
    border-left: 1px solid var(--default-border);
    padding-left: 20px;
  }

  .console-top-user-card__avatar {
    flex-shrink: 0;
    border: 1px solid var(--art-card-border);
    background: var(--default-box-color);
  }

  @media (max-width: 1400px) {
    .console-overview {
      grid-template-columns: 1fr;
    }

    .console-announcement-panel {
      border-left: 0;
      border-top: 1px solid var(--default-border);
      padding-top: 20px;
      padding-left: 0;
    }
  }

  @media (max-width: 1280px) {
    .console-trend-summary {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 640px) {
    .console-trend-card {
      min-height: 0;
    }

    .console-trend-summary {
      grid-template-columns: 1fr;
    }
  }
</style>
