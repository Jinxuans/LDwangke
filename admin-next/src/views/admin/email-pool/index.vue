<template>
  <div class="admin-email-pool-page art-full-height">
    <ElCard class="art-table-card mt-5">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">可用邮箱 {{ activeCount }} 个</ElTag>
            <ElTag type="warning" effect="plain">异常邮箱 {{ errorCount }} 个</ElTag>
            <ElTag type="primary" effect="plain">今日发送 {{ stats.today_sent || 0 }}</ElTag>
            <ElButton plain @click="handleResetCounters">重置计数</ElButton>
            <ElButton type="primary" plain @click="openCreateDialog">新增邮箱</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="loading" :data="list" :columns="columns" :show-table-header="true" />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑邮箱' : '新增邮箱'"
      width="820px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.1fr_0.9fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">SMTP 配置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">账号、授权码和发件邮箱会直接写入轮询池配置表。</p>
          </div>

          <div class="mt-5 grid gap-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">邮箱名称</label>
              <ElInput v-model="editForm.name" maxlength="40" placeholder="例如 QQ 轮询 01" />
            </div>
            <div class="grid gap-4 md:grid-cols-[1fr_120px]">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">SMTP 地址</label>
                <ElInput v-model="editForm.host" placeholder="smtp.qq.com" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">端口</label>
                <ElInputNumber v-model="editForm.port" class="w-full" :min="1" :max="65535" :precision="0" />
              </div>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">加密方式</label>
              <ElSelect v-model="editForm.encryption" class="w-full">
                <ElOption label="SSL" value="ssl" />
                <ElOption label="STARTTLS" value="starttls" />
                <ElOption label="无加密" value="none" />
              </ElSelect>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">邮箱账号</label>
              <ElInput v-model="editForm.user" placeholder="SMTP 登录邮箱" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">授权码</label>
              <ElInput
                v-model="editForm.password"
                type="password"
                show-password
                :placeholder="isEditing ? '留空表示不修改' : '新建时必填'"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">发件邮箱</label>
              <ElInput v-model="editForm.from_email" placeholder="留空则默认使用邮箱账号" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">轮询策略</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">权重越高越优先，限额达到后系统会自动切换到其他可用邮箱。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">权重</label>
                <ElInputNumber v-model="editForm.weight" class="w-full" :min="1" :max="100" :precision="0" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
                <ElSelect v-model="editForm.status" class="w-full">
                  <ElOption label="启用" :value="1" />
                  <ElOption label="禁用" :value="0" />
                </ElSelect>
              </div>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">日限额</label>
                <ElInputNumber v-model="editForm.day_limit" class="w-full" :min="0" :precision="0" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">小时限额</label>
                <ElInputNumber v-model="editForm.hour_limit" class="w-full" :min="0" :precision="0" />
              </div>
            </div>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">邮箱名称</span>
              <span class="truncate font-medium text-g-900">{{ editForm.name || '未命名邮箱' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">SMTP</span>
              <span class="font-medium text-g-900">
                {{ editForm.host || '未填写 SMTP 地址' }}:{{ editForm.port || 0 }} / {{ editForm.encryption }}
              </span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="editForm.status === 1 ? 'success' : 'info'" effect="plain">
                {{ editForm.status === 1 ? '启用中' : '已禁用' }}
              </ElTag>
              <ElTag type="primary" effect="plain">权重 {{ editForm.weight }}</ElTag>
              <ElTag type="warning" effect="plain">日限 {{ editForm.day_limit || '不限' }}</ElTag>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" @click="handleSave">保存邮箱</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="testVisible" title="测试发送" width="460px" destroy-on-close>
      <ElInput v-model="testTo" placeholder="输入测试收件邮箱地址" />
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="testVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="testLoading" @click="handleTest">发送测试邮件</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteLegacyEmailPoolAccount,
    fetchLegacyEmailPoolAccounts,
    fetchLegacyEmailPoolStats,
    resetLegacyEmailPoolCounters,
    saveLegacyEmailPoolAccount,
    testLegacyEmailPoolAccount,
    toggleLegacyEmailPoolAccount,
    type LegacyEmailPoolAccount,
    type LegacyEmailPoolSaveRequest,
    type LegacyEmailPoolStats
  } from '@/api/legacy/admin-email'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminEmailPoolPage' })

  const loading = ref(false)
  const testLoading = ref(false)
  const list = ref<LegacyEmailPoolAccount[]>([])
  const stats = ref<LegacyEmailPoolStats>({
    total_accounts: 0,
    active_accounts: 0,
    error_accounts: 0,
    today_sent: 0,
    today_fail: 0
  })
  const dialogVisible = ref(false)
  const testVisible = ref(false)
  const testId = ref(0)
  const testTo = ref('')

  const editForm = reactive<LegacyEmailPoolSaveRequest>({
    id: 0,
    name: '',
    host: '',
    port: 465,
    encryption: 'ssl',
    user: '',
    password: '',
    from_email: '',
    weight: 1,
    day_limit: 500,
    hour_limit: 50,
    status: 1
  })

  const isEditing = computed(() => Number(editForm.id || 0) > 0)
  const activeCount = computed(() => list.value.filter((item) => item.status === 1).length)
  const errorCount = computed(() => list.value.filter((item) => item.status === 2).length)

  const { columns, columnChecks } = useTableColumns<LegacyEmailPoolAccount>(() => [
    { type: 'index', label: '序号', width: 70 },
    {
      prop: 'name',
      label: '邮箱信息',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名邮箱'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `${row.user || '-'} / ${row.from_email || row.user || '-'}`),
          h('p', { class: 'text-xs text-g-400 mt-1' }, `${row.host}:${row.port} (${row.encryption})`)
        ])
    },
    {
      prop: 'weight',
      label: '权重',
      width: 80,
      align: 'center'
    },
    {
      prop: 'day_limit',
      label: '限额',
      width: 150,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', `日限 ${row.day_limit || '不限'}`),
          h('p', { class: 'text-xs text-g-500' }, `时限 ${row.hour_limit || '不限'}`)
        ])
    },
    {
      prop: 'today_sent',
      label: '今日 / 累计',
      width: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', `今日 ${row.today_sent} / 本时 ${row.hour_sent}`),
          h('p', { class: 'text-xs text-g-500' }, `累计 ${row.total_sent} 成功 / ${row.total_fail} 失败`)
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h(
            ElTag,
            { type: row.status === 1 ? 'success' : row.status === 2 ? 'danger' : 'info' },
            () => (row.status === 1 ? '启用' : row.status === 2 ? '异常' : '禁用')
          ),
          row.last_error && row.status === 2
            ? h('p', { class: 'text-xs text-[var(--el-color-danger)] line-clamp-2 mt-1' }, row.last_error)
            : null
        ])
    },
    {
      prop: 'last_used',
      label: '最后使用',
      width: 170,
      formatter: (row) => row.last_used || '-'
    },
    {
      prop: 'operation',
      label: '操作',
      minWidth: 260,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
          h(
            'button',
            {
              class: 'rounded-md border border-[var(--el-color-primary-light-6)] bg-[var(--el-color-primary-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-primary)]',
              onClick: () => openTestDialog(row.id)
            },
            '测试'
          ),
          h(
            'button',
            {
              class:
                row.status === 1
                  ? 'rounded-md border border-[var(--el-color-danger-light-6)] bg-[var(--el-color-danger-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-danger)]'
                  : 'rounded-md border border-[var(--el-color-success-light-6)] bg-[var(--el-color-success-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-success)]',
              onClick: () => handleToggle(row.id, row.status === 1 ? 0 : 1)
            },
            row.status === 1 ? '禁用' : '启用'
          ),
          h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
        ])
    }
  ])

  const resetForm = () => {
    editForm.id = 0
    editForm.name = ''
    editForm.host = 'smtp.qq.com'
    editForm.port = 465
    editForm.encryption = 'ssl'
    editForm.user = ''
    editForm.password = ''
    editForm.from_email = ''
    editForm.weight = 1
    editForm.day_limit = 500
    editForm.hour_limit = 50
    editForm.status = 1
  }

  const loadData = async () => {
    loading.value = true
    try {
      const [accounts, nextStats] = await Promise.all([
        fetchLegacyEmailPoolAccounts(),
        fetchLegacyEmailPoolStats()
      ])
      list.value = Array.isArray(accounts) ? accounts : []
      stats.value = nextStats || stats.value
    } finally {
      loading.value = false
    }
  }

  const openCreateDialog = () => {
    resetForm()
    dialogVisible.value = true
  }

  const openEditDialog = (item: LegacyEmailPoolAccount) => {
    editForm.id = item.id
    editForm.name = item.name
    editForm.host = item.host
    editForm.port = item.port
    editForm.encryption = item.encryption
    editForm.user = item.user
    editForm.password = ''
    editForm.from_email = item.from_email
    editForm.weight = item.weight
    editForm.day_limit = item.day_limit
    editForm.hour_limit = item.hour_limit
    editForm.status = item.status
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.host?.trim() || !editForm.user?.trim()) {
      ElMessage.warning('请填写完整的 SMTP 地址和邮箱账号')
      return
    }
    if (!isEditing.value && !editForm.password?.trim()) {
      ElMessage.warning('新增邮箱时必须填写授权码')
      return
    }

    await saveLegacyEmailPoolAccount({ ...editForm })
    ElMessage.success(isEditing.value ? '邮箱已更新' : '邮箱已创建')
    dialogVisible.value = false
    await loadData()
  }

  const handleDelete = async (item: LegacyEmailPoolAccount) => {
    await ElMessageBox.confirm(`确定删除邮箱「${item.name || item.id}」吗？`, '删除邮箱', {
      type: 'warning'
    })
    await deleteLegacyEmailPoolAccount(item.id)
    ElMessage.success('邮箱已删除')
    await loadData()
  }

  const handleToggle = async (id: number, status: number) => {
    await toggleLegacyEmailPoolAccount(id, status)
    ElMessage.success('状态已更新')
    await loadData()
  }

  const openTestDialog = (id: number) => {
    testId.value = id
    testTo.value = ''
    testVisible.value = true
  }

  const handleTest = async () => {
    if (!testTo.value.trim()) {
      ElMessage.warning('请输入测试收件邮箱')
      return
    }
    testLoading.value = true
    try {
      await testLegacyEmailPoolAccount(testId.value, testTo.value.trim())
      ElMessage.success('测试邮件已发送')
      testVisible.value = false
    } finally {
      testLoading.value = false
    }
  }

  const handleResetCounters = async () => {
    await resetLegacyEmailPoolCounters()
    ElMessage.success('计数器已重置')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>
