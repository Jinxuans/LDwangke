<template>
  <div class="order-list-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      :items="searchItems"
      :showExpand="true"
      :defaultExpanded="false"
      :span="6"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        layout="refresh,size,fullscreen,columns,settings"
        fullClass="art-table-card"
        @refresh="loadOrders(pagination.current)"
      >
        <template #left>
          <ElSpace wrap>
            <ElTag v-if="selectedOrderIds.length" type="primary" effect="plain">
              已选 {{ selectedOrderIds.length }} 条
            </ElTag>
            <ElButton v-if="selectedOrderIds.length" plain size="small" @click="clearSelection">
              清空已选 {{ selectedOrderIds.length }}
            </ElButton>

            <ElDropdown
              v-if="isAdmin && selectedOrderIds.length"
              @command="handleBatchStatusCommand"
            >
              <ElButton plain size="small">
                任务状态
                <i class="i-ep-arrow-down ml-1 text-xs" />
              </ElButton>
              <template #dropdown>
                <ElDropdownMenu>
                  <ElDropdownItem
                    v-for="item in batchStatusOptions"
                    :key="item"
                    :command="item"
                  >
                    {{ item }}
                  </ElDropdownItem>
                </ElDropdownMenu>
              </template>
            </ElDropdown>

            <ElDropdown v-if="isAdmin && selectedOrderIds.length" @command="handleBatchDockCommand">
              <ElButton plain size="small">
                处理状态
                <i class="i-ep-arrow-down ml-1 text-xs" />
              </ElButton>
              <template #dropdown>
                <ElDropdownMenu>
                  <ElDropdownItem
                    v-for="item in dockStatusOptions"
                    :key="item.value"
                    :command="item.value"
                  >
                    {{ item.label }}
                  </ElDropdownItem>
                </ElDropdownMenu>
              </template>
            </ElDropdown>

            <ElButton
              v-if="isAdmin && selectedOrderIds.length"
              plain
              size="small"
              @click="handleManualDock"
            >
              对接上游
            </ElButton>
            <ElButton
              v-if="isAdmin && selectedOrderIds.length"
              plain
              size="small"
              @click="handleSyncProgress"
            >
              同步进度
            </ElButton>
            <ElButton
              v-if="isAdmin && selectedOrderIds.length"
              plain
              size="small"
              @click="handleBatchSync"
            >
              批量同步
            </ElButton>
            <ElButton
              v-if="isAdmin && selectedOrderIds.length"
              plain
              size="small"
              @click="handleBatchResend"
            >
              批量补单
            </ElButton>
            <ElButton
              v-if="isAdmin && selectedOrderIds.length"
              plain
              size="small"
              @click="openRemarksDialog"
            >
              修改备注
            </ElButton>
            <ElButton
              v-if="isAdmin && selectedOrderIds.length"
              type="danger"
              plain
              size="small"
              @click="handleBatchRefund"
            >
              退款
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        rowKey="oid"
        :loading="loading"
        :data="orders"
        :columns="columns"
        :pagination="pagination"
        :pagination-options="{ align: 'center', pageSizes: [20, 50, 100, 200] }"
        :show-table-header="true"
        @selection-change="handleSelectionChange"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      >
        <template #orderInfo="{ row }">
          <div class="leading-6">
            <p class="font-semibold text-g-900">#{{ row.oid }}</p>
            <p class="text-xs text-g-500">{{ row.ptname || '-' }}</p>
          </div>
        </template>

        <template #accountInfo="{ row }">
          <div class="leading-6">
            <p class="truncate text-sm text-g-500">{{ row.school || '-' }}</p>
            <p class="truncate font-medium text-g-900">{{ row.user || '-' }}</p>
            <p class="truncate text-xs text-g-400">密码：{{ row.pass || '-' }}</p>
          </div>
        </template>

        <template #course="{ row }">
          <div class="leading-6">
            <p class="line-clamp-2 text-sm text-g-800">{{ row.kcname || '-' }}</p>
          </div>
        </template>

        <template #status="{ row }">
          <ElTag :type="getStatusTagType(row.status)">{{ row.status || '待处理' }}</ElTag>
        </template>

        <template #process="{ row }">
          <ElProgress :percentage="progressPercent(row.process)" :stroke-width="8" />
        </template>

        <template #remarks="{ row }">
          <p class="line-clamp-2 text-sm leading-6 text-g-600">{{ row.remarks || '-' }}</p>
        </template>

        <template #fees="{ row }">
          <span class="font-semibold text-[var(--el-color-success)]">
            ¥{{ formatAmount(row.fees) }}
          </span>
        </template>

        <template #push="{ row }">
          <div class="flex flex-col items-center gap-1 text-xs leading-5">
            <span v-if="row.pushUid" :class="pushClassName(row.pushStatus)">
              微：{{ row.pushStatus || '已绑' }}
            </span>
            <span v-if="row.pushEmail" :class="pushClassName(row.pushEmailStatus)">
              邮：{{ row.pushEmailStatus || '已绑' }}
            </span>
            <span v-if="row.showdoc_push_url" :class="pushClassName(row.pushShowdocStatus)">
              SD：{{ row.pushShowdocStatus || '已绑' }}
            </span>
            <span v-if="!row.pushUid && !row.pushEmail && !row.showdoc_push_url" class="text-g-400">
              未绑定
            </span>
          </div>
        </template>

        <template #dockstatus="{ row }">
          <ElTag :type="getDockTagType(row.dockstatus)">
            {{ dockStatusLabelMap[row.dockstatus] || '未知' }}
          </ElTag>
        </template>

        <template #actions="{ row }">
          <div class="flex items-center gap-2">
            <ElButton text type="primary" @click="showDetail(row)">详情</ElButton>
            <ElDropdown @command="(command: string) => handleRowAction(command, row)">
              <ElButton text>
                更多
                <i class="i-ep-arrow-down ml-1 text-xs" />
              </ElButton>
              <template #dropdown>
                <ElDropdownMenu>
                  <ElDropdownItem command="chat">聊天</ElDropdownItem>
                  <ElDropdownItem v-if="row.can_pup_login" command="login">Pup 登录</ElDropdownItem>
                </ElDropdownMenu>
              </template>
            </ElDropdown>
          </div>
        </template>
      </ArtTable>
    </ElCard>

    <ElDialog v-model="detailVisible" title="订单详情" width="980px">
      <div v-loading="detailLoading">
        <div v-if="currentOrder" class="grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
          <section class="space-y-5">
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="flex flex-wrap items-center gap-3">
                  <ElTag :type="getStatusTagType(currentOrder.status)">
                    {{ currentOrder.status || '待处理' }}
                  </ElTag>
                  <span class="text-base font-semibold text-g-900">订单 #{{ currentOrder.oid }}</span>
                  <span v-if="isAdmin" class="text-sm text-g-500">UID {{ currentOrder.uid }}</span>
                </div>
                <span class="text-sm text-g-500">{{ currentOrder.addtime }}</span>
              </div>

              <div class="mt-4 grid gap-3 sm:grid-cols-2">
                <div class="text-sm text-g-600">平台：{{ currentOrder.ptname || '-' }}</div>
                <div class="text-sm text-g-600">学校：{{ currentOrder.school || '-' }}</div>
                <div class="text-sm text-g-600">账号：{{ currentOrder.user || '-' }}</div>
                <div class="text-sm text-g-600">密码：{{ currentOrder.pass || '-' }}</div>
                <div class="text-sm text-g-600">课程 ID：{{ currentOrder.kcid || currentOrder.cid || '-' }}</div>
                <div class="text-sm text-g-600">
                  金额：<span class="font-semibold text-[var(--el-color-success)]">¥{{ formatAmount(currentOrder.fees) }}</span>
                </div>
              </div>

              <div class="mt-4 rounded-custom-sm border-full-d bg-box px-4 py-4">
                <p class="text-sm font-medium text-g-900">{{ currentOrder.kcname || '-' }}</p>
                <div class="mt-4 space-y-2">
                  <ElProgress :percentage="progressPercent(currentOrder.process)" :stroke-width="8" />
                  <p class="text-xs text-g-500">{{ currentOrder.process || '0' }}</p>
                </div>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d bg-box p-5">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <h3 class="text-lg font-semibold text-g-900">附加信息</h3>
                <ElTag v-if="isAdmin && currentOrder.yid" effect="plain">
                  上游单号 {{ currentOrder.yid }}
                </ElTag>
              </div>
              <div class="mt-4 grid gap-3 sm:grid-cols-2">
                <div v-if="isAdmin" class="text-sm text-g-600">
                  处理状态：{{ dockStatusLabelMap[currentOrder.dockstatus] || '未知' }}
                </div>
                <div v-if="isAdmin" class="text-sm text-g-600">
                  货源类型：{{ currentOrder.supplier_pt || '-' }}
                </div>
                <div class="text-sm text-g-600">
                  微信推送：{{ currentOrder.pushUid ? currentOrder.pushStatus || '已绑定' : '未绑定' }}
                </div>
                <div class="text-sm text-g-600">
                  邮件推送：{{ currentOrder.pushEmail ? currentOrder.pushEmailStatus || '已绑定' : '未绑定' }}
                </div>
              </div>

              <div
                v-if="currentOrder.remarks"
                class="mt-4 rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4 text-sm leading-7 text-g-700"
              >
                {{ currentOrder.remarks }}
              </div>
            </article>
          </section>

          <section class="space-y-5">
            <article class="rounded-custom-sm border-full-d bg-box p-5">
              <h3 class="text-lg font-semibold text-g-900">快捷操作</h3>
              <div class="mt-4 flex flex-col gap-3">
                <ElButton type="primary" @click="goChat(currentOrder)">去聊天页</ElButton>
                <ElButton
                  v-if="categorySwitches.ticket"
                  plain
                  @click="openTicketDialog(currentOrder.oid)"
                >
                  订单反馈
                </ElButton>
                <ElButton v-if="categorySwitches.log" plain @click="openLogs(currentOrder.oid)">
                  查看日志
                </ElButton>
                <ElButton
                  v-if="categorySwitches.changepass"
                  plain
                  @click="openPasswordDialog(currentOrder.oid)"
                >
                  修改密码
                </ElButton>
                <ElButton
                  plain
                  :loading="resubmitSubmittingOid === currentOrder.oid"
                  @click="handleResubmit(currentOrder.oid)"
                >
                  补单
                </ElButton>
                <ElButton
                  v-if="categorySwitches.allowpause"
                  plain
                  @click="handlePause(currentOrder.oid)"
                >
                  暂停 / 恢复
                </ElButton>
                <ElButton
                  v-if="currentOrder.can_pup_login"
                  plain
                  @click="handlePupLogin(currentOrder.oid)"
                >
                  Pup 登录
                </ElButton>
                <ElButton
                  v-if="currentOrder.can_pup_login"
                  plain
                  @click="openPupResetDialog(currentOrder.oid, 'score')"
                >
                  Pup 重置
                </ElButton>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
              <h3 class="text-lg font-semibold text-g-900">风险操作</h3>
              <div class="mt-4 flex flex-col gap-3">
                <ElButton type="danger" plain @click="handleCancel(currentOrder.oid)">
                  取消订单
                </ElButton>
                <ElButton v-if="isAdmin" type="danger" @click="handleSingleRefund(currentOrder.oid)">
                  退款
                </ElButton>
              </div>
            </article>
          </section>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="logsVisible" title="订单日志" width="760px">
      <div class="mb-3 flex items-center justify-between gap-3">
        <span class="text-sm text-g-500">订单 ID：{{ logOrderId || '-' }}</span>
        <ElButton plain :loading="logsLoading" @click="logOrderId && openLogs(logOrderId)">
          刷新日志
        </ElButton>
      </div>

      <div v-loading="logsLoading">
        <div v-if="logs.length" class="max-h-[520px] space-y-3 overflow-y-auto pr-1">
          <article
            v-for="(item, index) in logs"
            :key="`${item.time}-${index}`"
            class="rounded-custom-sm border-full-d bg-g-100/50 p-4"
          >
            <div class="flex flex-wrap items-center gap-3">
              <ElTag effect="plain">{{ item.time || '未知时间' }}</ElTag>
              <ElTag :type="getStatusTagType(item.status)">{{ item.status || '未知状态' }}</ElTag>
              <span v-if="item.process" class="text-sm text-g-600">{{ item.process }}</span>
            </div>
            <p v-if="item.course" class="mt-3 text-sm text-g-600">课程：{{ item.course }}</p>
            <p v-if="item.remarks" class="mt-2 whitespace-pre-wrap text-sm leading-6 text-g-700">
              {{ item.remarks }}
            </p>
          </article>
        </div>
        <ElEmpty v-else description="暂无日志" />
      </div>
    </ElDialog>

    <ElDialog v-model="ticketVisible" title="订单反馈" width="680px">
      <div class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
        <p class="text-sm text-g-500">订单 ID：{{ ticketForm.oid || '-' }}</p>
        <h3 class="mt-3 text-lg font-semibold text-g-900">提交订单反馈</h3>
        <p class="mt-2 text-sm leading-6 text-g-500">
          提交后会自动创建工单，并把首条反馈消息同步到在线客服会话。
        </p>

        <ElInput
          v-model="ticketForm.content"
          class="mt-5"
          type="textarea"
          :rows="7"
          resize="none"
          placeholder="请尽量写清订单号、异常现象、希望怎么处理。"
        />
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="ticketVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="ticketSubmitting" @click="submitTicket">
            提交并进入聊天
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="passwordVisible" title="修改订单密码" width="420px">
      <div class="space-y-4 py-2">
        <p class="text-sm text-g-500">订单 ID：{{ passwordForm.oid || '-' }}</p>
        <ElInput
          v-model="passwordForm.newPwd"
          show-password
          placeholder="请输入新密码，至少 3 位"
          @keyup.enter="submitPasswordChange"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="passwordVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="passwordSubmitting" @click="submitPasswordChange">
            保存
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="pupResetVisible" title="Pup 重置" width="480px">
      <div class="space-y-5 py-2">
        <p class="text-sm text-g-500">订单 ID：{{ pupResetForm.oid || '-' }}</p>
        <ElSelect v-model="pupResetForm.type" class="w-full">
          <ElOption label="重置分数" value="score" />
          <ElOption label="重置时长" value="duration" />
          <ElOption label="重置周期" value="period" />
        </ElSelect>
        <ElInputNumber
          v-model="pupResetForm.value"
          class="w-full"
          :min="pupResetMin"
          :max="pupResetMax"
        />
        <p class="text-xs text-g-500">
          {{ pupResetHint }}
        </p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="pupResetVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="pupResetSubmitting" @click="submitPupReset">
            提交
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="remarksVisible" title="批量修改备注" width="560px">
      <div class="space-y-4 py-2">
        <p class="text-sm text-g-500">已选订单：{{ selectedOrderIds.length }} 个</p>
        <ElInput
          v-model="remarksValue"
          type="textarea"
          :rows="5"
          resize="none"
          placeholder="输入新的备注内容"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="remarksVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="remarksSubmitting" @click="submitRemarks">
            保存
          </ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { createLegacyChatSession, sendLegacyChatMessage } from '@/api/legacy/chat'
  import {
    fetchLegacyCategorySwitches,
    type LegacyCategorySwitches
  } from '@/api/legacy/class'
  import {
    batchResendLegacyOrders,
    batchSyncLegacyOrderProgress,
    cancelLegacyOrder,
    changeLegacyOrderPassword,
    changeLegacyOrderStatus,
    fetchLegacyOrderDetail,
    fetchLegacyOrderList,
    fetchLegacyOrderLogs,
    fetchLegacyPupLogin,
    manualDockLegacyOrders,
    modifyLegacyOrderRemarks,
    pauseLegacyOrder,
    refundLegacyOrders,
    resetLegacyPupOrder,
    resubmitLegacyOrder,
    syncLegacyOrderProgress,
    type LegacyOrderItem,
    type LegacyOrderListParams,
    type LegacyOrderLogEntry,
    type LegacyOrderStats
  } from '@/api/legacy/order'
  import { createLegacyUserTicket } from '@/api/legacy/ticket'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'OrderListPage' })

  const router = useRouter()
  const userStore = useUserStore()

  const tableRef = ref()
  const loading = ref(false)
  const orders = ref<LegacyOrderItem[]>([])
  const selectedOrders = ref<LegacyOrderItem[]>([])
  const selectedOrderIds = computed(() => selectedOrders.value.map((item) => item.oid))

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const filters = reactive<LegacyOrderListParams>({
    page: 1,
    limit: 20,
    addtime_range: [],
    user: '',
    pass: '',
    school: '',
    oid: '',
    cid: '',
    kcname: '',
    status_text: '',
    dock: '',
    uid: '',
    hid: '',
    search: ''
  })

  const detailVisible = ref(false)
  const detailLoading = ref(false)
  const resubmitSubmittingOid = ref<number>()
  const currentOrder = ref<LegacyOrderItem | null>(null)
  const categorySwitches = ref<LegacyCategorySwitches>({
    allowpause: 0,
    changepass: 1,
    log: 0,
    ticket: 0
  })

  const logsVisible = ref(false)
  const logsLoading = ref(false)
  const logs = ref<LegacyOrderLogEntry[]>([])
  const logOrderId = ref<number>()

  const ticketVisible = ref(false)
  const ticketSubmitting = ref(false)
  const ticketForm = reactive({
    oid: undefined as number | undefined,
    content: ''
  })

  const passwordVisible = ref(false)
  const passwordSubmitting = ref(false)
  const passwordForm = reactive({
    oid: undefined as number | undefined,
    newPwd: ''
  })

  const pupResetVisible = ref(false)
  const pupResetSubmitting = ref(false)
  const pupResetForm = reactive({
    oid: undefined as number | undefined,
    type: 'score' as 'duration' | 'period' | 'score',
    value: 85
  })

  const remarksVisible = ref(false)
  const remarksSubmitting = ref(false)
  const remarksValue = ref('')

  const statusOptions = [
    '待处理',
    '进行中',
    '已完成',
    '异常',
    '已取消',
    '待考试',
    '待时长',
    '待重刷',
    '待上号',
    '已上号',
    '排队中',
    '补刷中',
    '处理中',
    '考试中',
    '队列中',
    '上号中',
    '重刷中',
    '刷课中',
    '时长中',
    '讨论中',
    '暂停中',
    '学习中',
    '运行中',
    '完成次数中',
    '平时分中',
    '平时分',
    '已提取',
    '已提交',
    '已暂停',
    '已结课',
    '已完成待考试',
    '已退款',
    '已退单',
    '异常待处理',
    '异常已处理',
    '等待中',
    '出错啦',
    '问题单',
    '失败',
    '密码错误',
    '登录失败',
    '未找到课程'
  ]

  const batchStatusOptions = ['待处理', '进行中', '已完成', '异常', '已取消']
  const dockStatusOptions = [
    { label: '待处理', value: '0' },
    { label: '处理成功', value: '1' },
    { label: '处理失败', value: '2' },
    { label: '重复下单', value: '3' },
    { label: '已取消', value: '4' },
    { label: '已删除', value: '5' },
    { label: '已退款', value: '6' },
    { label: '自营', value: '99' }
  ]

  const dockStatusLabelMap: Record<string, string> = {
    '0': '待处理',
    '1': '处理成功',
    '2': '处理失败',
    '3': '重复下单',
    '4': '已取消',
    '5': '已删除',
    '6': '已退款',
    '99': '自营'
  }

  const isAdmin = computed(() => {
    const roles = userStore.info?.roles || []
    return roles.includes('R_ADMIN') || roles.includes('R_SUPER')
  })

  const adminColumnProps = ['dockstatus', 'uid']

  const { columns, columnChecks, toggleColumn } = useTableColumns<LegacyOrderItem>(() => [
    {
      type: 'selection',
      width: 50,
      fixed: 'left',
      disabled: true
    },
    {
      prop: 'oid',
      label: '订单 / 平台',
      minWidth: 130,
      useSlot: true,
      slotName: 'orderInfo'
    },
    {
      prop: 'accountInfo',
      label: '账号信息',
      minWidth: 220,
      useSlot: true,
      slotName: 'accountInfo'
    },
    {
      prop: 'course',
      label: '课程',
      minWidth: 220,
      useSlot: true,
      slotName: 'course'
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      useSlot: true
    },
    {
      prop: 'process',
      label: '进度',
      minWidth: 170,
      useSlot: true
    },
    {
      prop: 'remarks',
      label: '日志',
      minWidth: 220,
      useSlot: true
    },
    {
      prop: 'fees',
      label: '金额',
      width: 110,
      align: 'right',
      useSlot: true
    },
    {
      prop: 'push',
      label: '推送',
      width: 120,
      align: 'center',
      useSlot: true
    },
    {
      prop: 'dockstatus',
      label: '处理',
      width: 120,
      align: 'center',
      useSlot: true,
      visible: isAdmin.value,
      disabled: !isAdmin.value
    },
    {
      prop: 'uid',
      label: 'UID',
      width: 90,
      align: 'center',
      visible: isAdmin.value,
      disabled: !isAdmin.value
    },
    {
      prop: 'addtime',
      label: '提交时间',
      width: 180
    },
    {
      prop: 'actions',
      label: '操作',
      width: 150,
      fixed: 'right',
      useSlot: true
    }
  ])

  watch(
    isAdmin,
    (visible) => {
      adminColumnProps.forEach((prop) => {
        toggleColumn(prop, visible)
      })
      columnChecks.value = columnChecks.value.map((item: any) =>
        item.prop && adminColumnProps.includes(item.prop) ? { ...item, disabled: !visible } : item
      )
    },
    { immediate: true }
  )

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'search',
      type: 'input',
      span: 8,
      props: {
        clearable: true,
        placeholder: '搜索账号 / 学校 / 课程 / 订单号'
      }
    },
    {
      label: '提交时间',
      key: 'addtime_range',
      type: 'datetimerange',
      span: 8,
      props: {
        type: 'datetimerange',
        clearable: true,
        valueFormat: 'YYYY-MM-DD HH:mm:ss',
        startPlaceholder: '开始时间',
        endPlaceholder: '结束时间',
        rangeSeparator: '至',
        defaultTime: [
          new Date(2000, 0, 1, 0, 0, 0),
          new Date(2000, 0, 1, 23, 59, 59)
        ]
      }
    },
    {
      label: '任务状态',
      key: 'status_text',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '请选择任务状态',
        options: statusOptions.map((item) => ({ label: item, value: item }))
      }
    },
    {
      label: '处理状态',
      key: 'dock',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '请选择处理状态',
        options: dockStatusOptions
      }
    },
    {
      label: '账号',
      key: 'user',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入账号'
      }
    },
    {
      label: '学校',
      key: 'school',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入学校'
      }
    },
    {
      label: '课程',
      key: 'kcname',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入课程名称'
      }
    },
    {
      label: '订单 ID',
      key: 'oid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入订单 ID'
      }
    },
    {
      label: '课程 ID',
      key: 'cid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入课程 ID'
      }
    },
    {
      label: '货源 HID',
      key: 'hid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入货源 HID'
      }
    },
    {
      label: '密码',
      key: 'pass',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入密码'
      }
    },
    {
      label: '用户 UID',
      key: 'uid',
      type: 'input',
      hidden: !isAdmin.value,
      props: {
        clearable: true,
        placeholder: '输入用户 UID'
      }
    }
  ])

  const pupResetMin = computed(() => {
    if (pupResetForm.type === 'score') return 70
    if (pupResetForm.type === 'duration') return 0
    return 1
  })

  const pupResetMax = computed(() => {
    if (pupResetForm.type === 'score') return 100
    if (pupResetForm.type === 'duration') return 50
    return 20
  })

  const pupResetHint = computed(() => {
    if (pupResetForm.type === 'score') return '分数范围 70-100'
    if (pupResetForm.type === 'duration') return '时长范围 0-50 小时'
    return '周期范围 1-20 天，从下单时间开始计算'
  })

  const loadOrders = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      filters.page = pagination.current
      filters.limit = pagination.size
      const [startTime, endTime] = filters.addtime_range || []
      const params: LegacyOrderListParams = {
        ...filters,
        page: pagination.current,
        limit: pagination.size
      }
      delete params.addtime_range
      if (startTime) params.start_time = startTime
      if (endTime) params.end_time = endTime

      const result = await fetchLegacyOrderList(params)
      orders.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || page)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = async () => {
    await loadOrders(1)
  }

  const handleReset = async () => {
    Object.assign(filters, {
      page: 1,
      limit: 20,
      addtime_range: [],
      user: '',
      pass: '',
      school: '',
      oid: '',
      cid: '',
      kcname: '',
      status_text: '',
      dock: '',
      uid: '',
      hid: '',
      search: ''
    })
    await loadOrders(1)
  }

  const handleCurrentChange = async (page: number) => {
    await loadOrders(page)
  }

  const handleSizeChange = async (size: number) => {
    pagination.size = size
    await loadOrders(1)
  }

  const clearSelection = () => {
    selectedOrders.value = []
    tableRef.value?.elTableRef?.clearSelection?.()
  }

  const handleSelectionChange = (rows: LegacyOrderItem[]) => {
    selectedOrders.value = rows
  }

  const getStatusTagType = (
    status?: string
  ): 'danger' | 'info' | 'primary' | 'success' | 'warning' => {
    if (!status) return 'info'
    if (['已完成', '已上号', '已结课', '已完成待考试'].includes(status)) return 'success'
    if (
      ['进行中', '刷课中', '运行中', '学习中', '时长中', '讨论中', '完成次数中', '平时分中'].includes(
        status
      )
    ) {
      return 'primary'
    }
    if (
      ['异常', '补刷中', '出错啦', '异常待处理', '失败', '密码错误', '登录失败'].includes(status)
    ) {
      return 'danger'
    }
    if (
      ['待处理', '等待中', '待考试', '待时长', '待重刷', '待上号', '排队中', '队列中', '上号中', '处理中', '考试中'].includes(
        status
      )
    ) {
      return 'warning'
    }
    return 'info'
  }

  const getDockTagType = (
    dockStatus?: string
  ): 'danger' | 'info' | 'primary' | 'success' | 'warning' => {
    if (dockStatus === '1' || dockStatus === '99') return 'success'
    if (dockStatus === '2') return 'danger'
    if (dockStatus === '0') return 'primary'
    if (dockStatus === '4' || dockStatus === '6') return 'warning'
    return 'info'
  }

  const progressPercent = (value?: string) => {
    const percent = Number.parseFloat(value || '0')
    if (Number.isNaN(percent)) return 0
    return Math.max(0, Math.min(100, percent))
  }

  const formatAmount = (value?: string) => Number(value || 0).toFixed(2)

  const pushClassName = (status?: string) => {
    if (status === '成功') return 'text-[var(--el-color-success)]'
    if (status === '失败') return 'text-[var(--el-color-danger)]'
    return 'text-[var(--el-color-primary)]'
  }

  const showDetail = async (order: LegacyOrderItem) => {
    detailVisible.value = true
    detailLoading.value = true
    currentOrder.value = { ...order }
    categorySwitches.value = {
      allowpause: 0,
      changepass: 1,
      log: 0,
      ticket: 0
    }

    const [detailResult, switchesResult] = await Promise.allSettled([
      fetchLegacyOrderDetail(order.oid),
      order.cid > 0
        ? fetchLegacyCategorySwitches(order.cid)
        : Promise.resolve({
            allowpause: 0,
            changepass: 1,
            log: 0,
            ticket: 0
          })
    ])

    if (detailResult.status === 'fulfilled') {
      currentOrder.value = detailResult.value
    }

    if (switchesResult.status === 'fulfilled') {
      categorySwitches.value = switchesResult.value
    }

    detailLoading.value = false
  }

  const goChat = async (order?: LegacyOrderItem | null) => {
    const myUid = Number(userStore.info?.userId || 0)
    const targetUid = isAdmin.value && order?.uid && order.uid !== myUid ? order.uid : 1
    const path = targetUid === 1 ? '/chat' : '/admin/chat'
    const session = await createLegacyChatSession(targetUid)
    await router.push({
      path,
      query: session.list_id ? { listId: String(session.list_id) } : undefined
    })
  }

  const handleRowAction = async (command: string, row: LegacyOrderItem) => {
    if (command === 'chat') {
      await goChat(row)
      return
    }
    if (command === 'login') {
      await handlePupLogin(row.oid)
    }
  }

  const ensureSelection = () => {
    if (selectedOrderIds.value.length) return true
    ElMessage.warning('请先选择订单')
    return false
  }

  const confirmBatchAction = async (message: string, action: () => Promise<void>) => {
    if (!ensureSelection()) {
      return
    }
    await ElMessageBox.confirm(message, '批量操作确认', { type: 'warning' })
    await action()
    clearSelection()
  }

  const handleBatchStatusCommand = async (status: string) => {
    await confirmBatchAction(`确定将选中订单的任务状态改为“${status}”吗？`, async () => {
      await changeLegacyOrderStatus({
        status,
        oids: selectedOrderIds.value,
        type: 1
      })
      ElMessage.success('任务状态已更新')
      await loadOrders(pagination.current)
    })
  }

  const handleBatchDockCommand = async (dock: string) => {
    await confirmBatchAction(
      `确定将选中订单的处理状态改为“${dockStatusLabelMap[dock] || dock}”吗？`,
      async () => {
        await changeLegacyOrderStatus({
          status: dock,
          oids: selectedOrderIds.value,
          type: 2
        })
        ElMessage.success('处理状态已更新')
        await loadOrders(pagination.current)
      }
    )
  }

  const handleBatchRefund = async () => {
    await confirmBatchAction('确定退款选中的订单吗？该操作不可逆。', async () => {
      await refundLegacyOrders(selectedOrderIds.value)
      ElMessage.success('退款成功')
      await loadOrders(pagination.current)
    })
  }

  const handleManualDock = async () => {
    await confirmBatchAction('确定将选中的订单提交到上游吗？', async () => {
      const result = await manualDockLegacyOrders(selectedOrderIds.value)
      ElMessage.success(result.msg || `对接完成：成功 ${result.success}，失败 ${result.fail}`)
      await loadOrders(pagination.current)
    })
  }

  const handleSyncProgress = async () => {
    await confirmBatchAction('确定同步选中订单的上游进度吗？', async () => {
      const result = await syncLegacyOrderProgress(selectedOrderIds.value)
      ElMessage.success(result.msg || `同步完成，更新 ${result.updated} 条`)
      await loadOrders(pagination.current)
    })
  }

  const handleBatchSync = async () => {
    await confirmBatchAction('确定执行批量同步吗？', async () => {
      const result = await batchSyncLegacyOrderProgress(selectedOrderIds.value)
      ElMessage.success(result.msg || `批量同步完成，更新 ${result.updated} 条`)
      await loadOrders(pagination.current)
    })
  }

  const handleBatchResend = async () => {
    await confirmBatchAction('确定对选中订单执行批量补单吗？', async () => {
      const result = await batchResendLegacyOrders(selectedOrderIds.value)
      ElMessage.success(result.msg || `补单完成，成功 ${result.success}，失败 ${result.fail}`)
      await loadOrders(pagination.current)
    })
  }

  const openRemarksDialog = () => {
    if (!ensureSelection()) {
      return
    }
    remarksValue.value = ''
    remarksVisible.value = true
  }

  const submitRemarks = async () => {
    if (!selectedOrderIds.value.length) {
      remarksVisible.value = false
      return
    }
    remarksSubmitting.value = true
    try {
      await modifyLegacyOrderRemarks(selectedOrderIds.value, remarksValue.value.trim())
      ElMessage.success('备注已更新')
      remarksVisible.value = false
      clearSelection()
      await loadOrders(pagination.current)
    } finally {
      remarksSubmitting.value = false
    }
  }

  const handleCancel = async (oid: number) => {
    await ElMessageBox.confirm(`确定取消订单 #${oid} 吗？`, '取消订单', { type: 'warning' })
    await cancelLegacyOrder(oid)
    ElMessage.success('订单已取消')
    detailVisible.value = false
    await loadOrders(pagination.current)
  }

  const handleSingleRefund = async (oid: number) => {
    await ElMessageBox.confirm(`确定退款订单 #${oid} 吗？该操作不可逆。`, '订单退款', {
      type: 'warning'
    })
    await refundLegacyOrders([oid])
    ElMessage.success('退款成功')
    detailVisible.value = false
    await loadOrders(pagination.current)
  }

  const handlePause = async (oid: number) => {
    const result = await pauseLegacyOrder(oid)
    ElMessage.success(result.message || result.msg || '操作成功')
    await loadOrders(pagination.current)
    if (currentOrder.value?.oid === oid) {
      await showDetail(currentOrder.value)
    }
  }

  const handleResubmit = async (oid: number) => {
    if (resubmitSubmittingOid.value) {
      return
    }

    resubmitSubmittingOid.value = oid
    try {
      const result = await resubmitLegacyOrder(oid)
      ElMessage.success(result.message || result.msg || '补单成功')
      await loadOrders(pagination.current)
      if (currentOrder.value?.oid === oid) {
        await showDetail(currentOrder.value)
      }
    } finally {
      resubmitSubmittingOid.value = undefined
    }
  }

  const openPasswordDialog = (oid: number) => {
    passwordForm.oid = oid
    passwordForm.newPwd = ''
    passwordVisible.value = true
  }

  const submitPasswordChange = async () => {
    if (!passwordForm.oid) {
      return
    }
    if (!passwordForm.newPwd.trim() || passwordForm.newPwd.trim().length < 3) {
      ElMessage.warning('密码长度至少 3 位')
      return
    }

    passwordSubmitting.value = true
    try {
      await changeLegacyOrderPassword(passwordForm.oid, passwordForm.newPwd.trim())
      ElMessage.success('密码已修改')
      passwordVisible.value = false
      await loadOrders(pagination.current)
      if (currentOrder.value?.oid === passwordForm.oid) {
        await showDetail(currentOrder.value)
      }
    } finally {
      passwordSubmitting.value = false
    }
  }

  const openPupResetDialog = (oid: number, type: 'duration' | 'period' | 'score') => {
    pupResetForm.oid = oid
    pupResetForm.type = type
    pupResetForm.value = type === 'score' ? 85 : type === 'duration' ? 20 : 5
    pupResetVisible.value = true
  }

  const submitPupReset = async () => {
    if (!pupResetForm.oid) {
      return
    }
    pupResetSubmitting.value = true
    try {
      await resetLegacyPupOrder(pupResetForm.oid, pupResetForm.type, Number(pupResetForm.value || 0))
      ElMessage.success('Pup 重置成功')
      pupResetVisible.value = false
      await loadOrders(pagination.current)
    } finally {
      pupResetSubmitting.value = false
    }
  }

  const openLogs = async (oid: number) => {
    logsVisible.value = true
    logsLoading.value = true
    logOrderId.value = oid
    try {
      logs.value = await fetchLegacyOrderLogs(oid)
    } finally {
      logsLoading.value = false
    }
  }

  const openTicketDialog = (oid: number) => {
    ticketForm.oid = oid
    ticketForm.content = ''
    ticketVisible.value = true
  }

  const submitTicket = async () => {
    if (!ticketForm.content.trim()) {
      ElMessage.warning('请先填写反馈内容')
      return
    }

    ticketSubmitting.value = true
    try {
      await createLegacyUserTicket({
        oid: ticketForm.oid,
        type: '订单反馈',
        content: ticketForm.content.trim()
      })

      const session = await createLegacyChatSession(1)
      if (session.list_id) {
        const prefix = ticketForm.oid ? `【订单反馈 #${ticketForm.oid}】` : '【订单反馈】'
        await sendLegacyChatMessage({
          list_id: session.list_id,
          to_uid: 1,
          content: `${prefix}${ticketForm.content.trim()}`
        })
      }

      ticketVisible.value = false
      detailVisible.value = false
      ElMessage.success('反馈已提交，正在进入在线客服')
      await router.push({
        path: '/chat',
        query: session.list_id ? { listId: String(session.list_id) } : undefined
      })
    } finally {
      ticketSubmitting.value = false
    }
  }

  const handlePupLogin = async (oid: number) => {
    const result = await fetchLegacyPupLogin(oid)
    if (!result.url) {
      ElMessage.warning('未获取到登录地址')
      return
    }
    window.open(result.url, '_blank', 'noopener,noreferrer')
  }

  onMounted(() => {
    loadOrders(1)
  })
</script>
