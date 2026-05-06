<template>
  <div class="min-h-screen bg-[var(--el-bg-color-page)] px-4 py-6 md:px-6 md:py-8">
    <div class="mx-auto max-w-6xl">
      <section class="art-card-sm p-5 md:p-6">
        <div class="flex flex-wrap gap-2">
          <ElTag effect="plain">公开查单</ElTag>
          <ElTag effect="plain">结果 {{ list.length }}</ElTag>
          <ElTag type="success" effect="plain">微信 {{ wxBoundCount }}</ElTag>
          <ElTag type="primary" effect="plain">邮箱 {{ emailBoundCount }}</ElTag>
        </div>

        <div class="mt-6 grid gap-4 md:grid-cols-[1fr_1fr_auto]">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">账号</label>
            <ElInput
              v-model="searchUser"
              clearable
              size="large"
              placeholder="输入下单账号"
              @keyup.enter="handleSearch"
            />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">订单号</label>
            <ElInput
              v-model="searchOid"
              clearable
              size="large"
              placeholder="或输入订单号"
              @keyup.enter="handleSearch"
            />
          </div>
          <div class="flex items-end gap-3">
            <ElButton type="primary" size="large" :loading="loading" @click="handleSearch">查询</ElButton>
            <ElButton size="large" @click="handleReset">清空</ElButton>
          </div>
        </div>

        <div class="mt-5 rounded-2xl border border-dashed border-[var(--art-card-border)] bg-[var(--el-fill-color-lighter)] px-4 py-3 text-sm text-g-500">
          账号和订单号至少填写一项。若同一账号下有多笔订单，会返回最近的最多 50 条记录。
        </div>
      </section>

      <section class="art-card-sm mt-6 p-5 md:p-6">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="text-sm text-g-500">查询结果</div>
          <div class="flex flex-wrap items-center gap-3">
            <ElTag effect="plain">查询关键词 {{ searchSummary }}</ElTag>
            <ElTag type="success" effect="plain">完成 {{ completedCount }}</ElTag>
            <ElTag type="warning" effect="plain">进行中 {{ processingCount }}</ElTag>
          </div>
        </div>

        <ElSkeleton v-if="loading && !searched" class="mt-6" animated :rows="6" />

        <div v-else class="mt-6">
          <ElEmpty v-if="searched && !list.length && !loading" description="未找到相关订单" />

          <div v-else>
            <div class="space-y-4 md:hidden">
              <article
                v-for="item in list"
                :key="item.oid"
                class="rounded-2xl border border-[var(--art-card-border)] bg-[var(--el-fill-color-lighter)] p-4"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <p class="text-base font-semibold text-g-900">{{ item.kcname || '未命名课程' }}</p>
                    <p class="mt-1 text-xs text-g-400">订单号 {{ item.oid }} / {{ item.ptname || '未知平台' }}</p>
                  </div>
                  <ElTag :type="statusMeta(item.status).type">{{ statusMeta(item.status).label }}</ElTag>
                </div>

                <dl class="mt-4 grid gap-3 text-sm">
                  <div class="flex justify-between gap-4">
                    <dt class="text-g-400">账号</dt>
                    <dd class="truncate text-right text-g-700">{{ item.account || '-' }}</dd>
                  </div>
                  <div class="flex justify-between gap-4">
                    <dt class="text-g-400">进度</dt>
                    <dd class="text-right text-g-700">{{ item.process || '-' }}</dd>
                  </div>
                  <div class="flex justify-between gap-4">
                    <dt class="text-g-400">备注</dt>
                    <dd class="text-right text-g-700">{{ item.remarks || '-' }}</dd>
                  </div>
                  <div class="flex justify-between gap-4">
                    <dt class="text-g-400">时间</dt>
                    <dd class="text-right text-g-700">{{ item.addtime || '-' }}</dd>
                  </div>
                </dl>

                <div class="mt-4 flex flex-wrap gap-2">
                  <ElTag v-if="item.pushUid" type="success" effect="plain">微信已绑定</ElTag>
                  <ElTag v-if="item.pushEmail" type="primary" effect="plain">邮箱已绑定</ElTag>
                </div>

                <div class="mt-4 flex flex-wrap gap-2">
                  <ElButton size="small" @click="item.pushUid ? handleUnbindWx(item.account || '') : openWxBind(item.account || '')">
                    {{ item.pushUid ? '解绑微信' : '绑定微信' }}
                  </ElButton>
                  <ElButton size="small" @click="item.pushEmail ? handleUnbindEmail(item.account || '') : openEmailBind(item.account || '')">
                    {{ item.pushEmail ? '解绑邮箱' : '绑定邮箱' }}
                  </ElButton>
                  <ElButton
                    v-if="item.status !== '等待中'"
                    size="small"
                    type="primary"
                    plain
                    @click="handlePupLogin(item.oid)"
                  >
                    一键登录
                  </ElButton>
                </div>
              </article>
            </div>

            <div class="hidden md:block">
              <ElTable
                :data="list"
                border
                stripe
                empty-text="暂无数据"
                style="width: 100%"
              >
                <ElTableColumn prop="oid" label="订单号" width="96" align="center" />
                <ElTableColumn prop="ptname" label="平台" min-width="140" />
                <ElTableColumn prop="account" label="账号信息" min-width="220" show-overflow-tooltip />
                <ElTableColumn prop="kcname" label="课程名称" min-width="220" show-overflow-tooltip />
                <ElTableColumn label="状态" width="110" align="center">
                  <template #default="{ row }">
                    <ElTag :type="statusMeta(row.status).type">{{ statusMeta(row.status).label }}</ElTag>
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="process" label="进度" width="130" align="center" show-overflow-tooltip />
                <ElTableColumn prop="remarks" label="备注" min-width="180" show-overflow-tooltip />
                <ElTableColumn prop="addtime" label="时间" width="170" />
                <ElTableColumn label="操作" min-width="240" fixed="right">
                  <template #default="{ row }">
                    <div class="flex flex-wrap gap-2">
                      <ElButton
                        size="small"
                        @click="row.pushUid ? handleUnbindWx(row.account || '') : openWxBind(row.account || '')"
                      >
                        {{ row.pushUid ? '解绑微信' : '绑定微信' }}
                      </ElButton>
                      <ElButton
                        size="small"
                        @click="row.pushEmail ? handleUnbindEmail(row.account || '') : openEmailBind(row.account || '')"
                      >
                        {{ row.pushEmail ? '解绑邮箱' : '绑定邮箱' }}
                      </ElButton>
                      <ElButton
                        v-if="row.status !== '等待中'"
                        size="small"
                        type="primary"
                        plain
                        @click="handlePupLogin(row.oid)"
                      >
                        一键登录
                      </ElButton>
                    </div>
                  </template>
                </ElTableColumn>
              </ElTable>
            </div>
          </div>
        </div>
      </section>
    </div>

    <ElDialog v-model="wxDialogVisible" title="绑定微信推送" width="420px" @closed="stopWxPolling">
      <div class="py-2 text-center">
        <p class="text-sm text-g-500">请使用微信扫码，扫码成功后会自动完成绑定。</p>
        <div class="mt-5 flex justify-center">
          <div class="rounded-2xl border border-[var(--art-card-border)] bg-white p-4">
            <QrcodeVue v-if="wxQrUrl" :value="wxQrUrl" :size="200" level="M" />
            <div v-else class="flex h-[200px] w-[200px] items-center justify-center text-sm text-g-400">
              二维码加载中
            </div>
          </div>
        </div>
        <p class="mt-4 text-xs text-g-400">当前账号：{{ currentAccount || '-' }}</p>
      </div>
    </ElDialog>

    <ElDialog v-model="emailDialogVisible" title="绑定邮箱推送" width="460px">
      <ElForm label-position="top">
        <ElFormItem label="接收邮箱">
          <ElInput v-model="emailForm.email" placeholder="请输入接收通知的邮箱地址" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="emailDialogVisible = false">取消</ElButton>
          <ElButton type="primary" @click="submitEmailBind">确认绑定</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onUnmounted, reactive, ref } from 'vue'
  import QrcodeVue from 'qrcode.vue'
  import {
    bindLegacyEmailPush,
    bindLegacyWxPush,
    checkLegacyOrder,
    fetchLegacyPublicPupLogin,
    fetchLegacyWxQrcode,
    fetchLegacyWxScanUid,
    type LegacyCheckOrderResult,
    unbindLegacyEmailPush,
    unbindLegacyWxPush
  } from '@/api/legacy/auxiliary'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'PublicQueryPage' })

  const searchUser = ref('')
  const searchOid = ref('')
  const loading = ref(false)
  const searched = ref(false)
  const list = ref<LegacyCheckOrderResult[]>([])

  const wxDialogVisible = ref(false)
  const wxQrUrl = ref('')
  const wxQrCode = ref('')
  const currentAccount = ref('')
  const wxPollingTimer = ref<number>()

  const emailDialogVisible = ref(false)
  const emailForm = reactive({
    account: '',
    email: ''
  })

  const wxBoundCount = computed(() => list.value.filter((item) => Boolean(item.pushUid)).length)
  const emailBoundCount = computed(() => list.value.filter((item) => Boolean(item.pushEmail)).length)
  const completedCount = computed(() => list.value.filter((item) => item.status === '已完成').length)
  const processingCount = computed(() => list.value.filter((item) => item.status === '进行中').length)
  const searchSummary = computed(() => {
    const values = [searchUser.value.trim(), searchOid.value.trim()].filter(Boolean)
    return values.length ? values.join(' / ') : '未查询'
  })

  const statusMeta = (status?: string) => {
    switch (status) {
      case '已完成':
        return { label: '已完成', type: 'success' as const }
      case '已退款':
        return { label: '已退款', type: 'warning' as const }
      case '异常':
        return { label: '异常', type: 'danger' as const }
      case '进行中':
        return { label: '进行中', type: 'primary' as const }
      default:
        return { label: status || '等待中', type: 'info' as const }
    }
  }

  const stopWxPolling = () => {
    if (wxPollingTimer.value) {
      window.clearInterval(wxPollingTimer.value)
      wxPollingTimer.value = undefined
    }
  }

  const refreshResults = async () => {
    if (!searched.value) {
      return
    }
    await handleSearch()
  }

  const handleSearch = async () => {
    const user = searchUser.value.trim()
    const oid = searchOid.value.trim()

    if (!user && !oid) {
      ElMessage.warning('请输入账号或订单号')
      return
    }

    loading.value = true
    searched.value = true
    try {
      const result = await checkLegacyOrder({
        user: user || undefined,
        oid: oid || undefined
      })
      list.value = result.list || []
    } finally {
      loading.value = false
    }
  }

  const handleReset = () => {
    searchUser.value = ''
    searchOid.value = ''
    searched.value = false
    list.value = []
  }

  const startWxPolling = () => {
    stopWxPolling()
    wxPollingTimer.value = window.setInterval(async () => {
      if (!wxDialogVisible.value || !wxQrCode.value || !currentAccount.value) {
        return
      }

      try {
        const result = await fetchLegacyWxScanUid({ code: wxQrCode.value })
        if (result.uid) {
          const oids = list.value
            .filter((item) => item.account === currentAccount.value)
            .map((item) => item.oid)
            .join(',')
          await bindLegacyWxPush({
            account: currentAccount.value,
            pushUid: result.uid,
            oids
          })
          ElMessage.success('微信推送绑定成功')
          wxDialogVisible.value = false
          stopWxPolling()
          await refreshResults()
        }
      } catch {
        // 轮询期间忽略未扫码状态
      }
    }, 3000)
  }

  const openWxBind = async (account: string) => {
    if (!account) {
      ElMessage.warning('当前订单缺少账号信息')
      return
    }

    currentAccount.value = account
    const result = await fetchLegacyWxQrcode({ account })
    wxQrUrl.value = result.url
    wxQrCode.value = result.code
    wxDialogVisible.value = true
    startWxPolling()
  }

  const handleUnbindWx = async (account: string) => {
    if (!account) {
      ElMessage.warning('当前订单缺少账号信息')
      return
    }
    await unbindLegacyWxPush({ account })
    ElMessage.success('微信推送已解绑')
    await refreshResults()
  }

  const openEmailBind = (account: string) => {
    if (!account) {
      ElMessage.warning('当前订单缺少账号信息')
      return
    }
    emailForm.account = account
    emailForm.email = ''
    emailDialogVisible.value = true
  }

  const submitEmailBind = async () => {
    if (!emailForm.email.trim()) {
      ElMessage.warning('请输入邮箱地址')
      return
    }

    await bindLegacyEmailPush({
      account: emailForm.account,
      pushEmail: emailForm.email.trim()
    })
    ElMessage.success('邮箱推送绑定成功')
    emailDialogVisible.value = false
    await refreshResults()
  }

  const handleUnbindEmail = async (account: string) => {
    if (!account) {
      ElMessage.warning('当前订单缺少账号信息')
      return
    }
    await unbindLegacyEmailPush({ account })
    ElMessage.success('邮箱推送已解绑')
    await refreshResults()
  }

  const handlePupLogin = async (oid: number) => {
    const result = await fetchLegacyPublicPupLogin(oid)
    if (!result.url) {
      ElMessage.warning('未获取到登录地址')
      return
    }
    window.open(result.url, '_blank', 'noopener,noreferrer')
  }

  onUnmounted(() => {
    stopWxPolling()
  })
</script>
