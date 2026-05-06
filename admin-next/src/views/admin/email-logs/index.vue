<template>
  <div class="admin-email-logs-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">日志 {{ pagination.total }} 条</ElTag>
            <ElTag effect="plain">收件关键词 {{ appliedSearch.to_email || '全部' }}</ElTag>
            <ElTag type="primary" effect="plain">邮件类型 {{ currentTypeLabel }}</ElTag>
            <ElTag :type="appliedSearch.status === 0 ? 'danger' : 'success'" effect="plain">
              状态 {{ currentStatusLabel }}
            </ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyEmailSendLogs,
    type LegacyEmailSendLog
  } from '@/api/legacy/admin-email'
  import { ElTag } from 'element-plus'

  defineOptions({ name: 'AdminEmailLogsPage' })

  const mailTypeOptions = [
    { label: '注册验证码', value: 'register' },
    { label: '重置密码', value: 'reset' },
    { label: '系统通知', value: 'notify' },
    { label: '群发邮件', value: 'mass' },
    { label: '登录提醒', value: 'login_alert' },
    { label: '邮箱变更', value: 'change_email' },
    { label: '邮箱池测试', value: 'pool_test' }
  ]

  const loading = ref(false)
  const list = ref<LegacyEmailSendLog[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    mail_type?: string
    status?: number
    to_email?: string
  }>({
    mail_type: undefined,
    status: undefined,
    to_email: undefined
  })

  const appliedSearch = reactive<{
    mail_type?: string
    status?: number
    to_email?: string
  }>({
    mail_type: undefined,
    status: undefined,
    to_email: undefined
  })

  const currentTypeLabel = computed(() => getMailTypeLabel(appliedSearch.mail_type))
  const currentStatusLabel = computed(() => {
    if (appliedSearch.status === 1) {
      return '成功'
    }
    if (appliedSearch.status === 0) {
      return '失败'
    }
    return '全部'
  })

  const searchItems = computed(() => [
    {
      label: '邮件类型',
      key: 'mail_type',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部类型',
        options: mailTypeOptions
      }
    },
    {
      label: '发送状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '成功', value: 1 },
          { label: '失败', value: 0 }
        ]
      }
    },
    {
      label: '收件邮箱',
      key: 'to_email',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入收件邮箱关键词'
      }
    }
  ])

  const getMailTypeLabel = (value?: string) => {
    return mailTypeOptions.find((item) => item.value === value)?.label || '全部'
  }

  const getStatusTagType = (status: number) => (status === 1 ? 'success' : 'danger')

  const { columns, columnChecks } = useTableColumns<LegacyEmailSendLog>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'mail_type',
      label: '类型',
      width: 120,
      formatter: (row) =>
        h(ElTag, { type: 'primary', effect: 'plain' }, () => getMailTypeLabel(row.mail_type))
    },
    {
      prop: 'from_email',
      label: '发件 / 池编号',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 break-all' }, row.from_email || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `邮箱池 ID ${row.pool_id || 0}`)
        ])
    },
    {
      prop: 'to_email',
      label: '收件邮箱',
      minWidth: 220,
      formatter: (row) => h('span', { class: 'break-all text-g-800' }, row.to_email || '-')
    },
    {
      prop: 'subject',
      label: '邮件主题',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'line-clamp-1 font-semibold text-g-900' }, row.subject || '无主题'),
          row.error
            ? h('p', { class: 'line-clamp-2 text-xs text-[var(--el-color-danger)] mt-1' }, row.error)
            : h('p', { class: 'text-xs text-g-500 mt-1' }, '发送正常，无错误信息')
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: getStatusTagType(Number(row.status)) as any }, () =>
          Number(row.status) === 1 ? '成功' : '失败'
        )
    },
    {
      prop: 'addtime',
      label: '发送时间',
      width: 180,
      formatter: (row) => row.addtime || '-'
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyEmailSendLogs({
        page: pagination.current,
        limit: pagination.size,
        mail_type: appliedSearch.mail_type,
        status: appliedSearch.status,
        to_email: appliedSearch.to_email
      })
      list.value = result.list || []
      pagination.total = Number(result.total || 0)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { mail_type?: string; status?: number; to_email?: string }) => {
    appliedSearch.mail_type = params.mail_type || undefined
    appliedSearch.status = typeof params.status === 'number' ? params.status : undefined
    appliedSearch.to_email = params.to_email?.trim() || undefined
    pagination.current = 1
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.mail_type = undefined
    appliedSearch.status = undefined
    appliedSearch.to_email = undefined
    pagination.current = 1
    loadData(1)
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
