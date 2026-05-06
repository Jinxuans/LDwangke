<template>
  <div class="admin-moneylog-page art-full-height">
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
            <ElTag effect="plain">全站流水</ElTag>
            <ElTag effect="plain">UID {{ appliedSearch.uid || '全部' }}</ElTag>
            <ElTag type="primary" effect="plain">类型 {{ appliedSearch.type || '全部' }}</ElTag>
            <ElTag type="success" effect="plain">涉及用户 {{ currentUserCount }} 个</ElTag>
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
  import { ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyAdminMoneyLogs,
    type LegacyAdminMoneyLogItem
  } from '@/api/legacy/admin-stats'

  defineOptions({ name: 'AdminMoneylogPage' })

  const logTypeOptions = ['扣费', '充值', '退款', '调整', '商城代收', '提现申请', '提现通过', '提现驳回']

  const loading = ref(false)
  const list = ref<LegacyAdminMoneyLogItem[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    uid: '',
    type: ''
  })

  const appliedSearch = reactive({
    uid: '',
    type: ''
  })

  const currentUserCount = computed(() => new Set(list.value.map((item) => item.uid)).size)

  const searchItems = computed(() => [
    {
      label: '用户 UID',
      key: 'uid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入用户 UID'
      }
    },
    {
      label: '流水类型',
      key: 'type',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部类型',
        options: logTypeOptions.map((item) => ({ label: item, value: item }))
      }
    }
  ])

  const getTypeTag = (type: string) => {
    if (['充值', '商城代收', '提现驳回'].includes(type)) {
      return 'success'
    }
    if (['扣费', '提现申请'].includes(type)) {
      return 'danger'
    }
    if (['调整', '提现通过'].includes(type)) {
      return 'warning'
    }
    if (type === '退款') {
      return 'primary'
    }
    return 'info'
  }

  const { columns, columnChecks } = useTableColumns<LegacyAdminMoneyLogItem>(() => [
    {
      prop: 'id',
      label: 'ID',
      width: 80,
      align: 'center'
    },
    {
      prop: 'uid',
      label: '用户',
      width: 150,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.username || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `UID ${row.uid}`)
        ])
    },
    {
      prop: 'type',
      label: '类型',
      width: 120,
      formatter: (row) => h(ElTag, { type: getTypeTag(row.type), effect: 'plain' }, () => row.type || '-')
    },
    {
      prop: 'money',
      label: '金额',
      width: 120,
      align: 'right',
      formatter: (row) =>
        h(
          'span',
          { class: Number(row.money) >= 0 ? 'font-semibold text-[var(--el-color-success)]' : 'font-semibold text-[var(--el-color-danger)]' },
          `${Number(row.money) >= 0 ? '+' : ''}${Number(row.money || 0).toFixed(2)}`
        )
    },
    {
      prop: 'balance',
      label: '变动后余额',
      width: 130,
      align: 'right',
      formatter: (row) => `¥${Number(row.balance || 0).toFixed(2)}`
    },
    {
      prop: 'remark',
      label: '备注',
      minWidth: 260,
      formatter: (row) => row.remark || '-'
    },
    {
      prop: 'addtime',
      label: '时间',
      width: 180,
      formatter: (row) => row.addtime || '-'
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminMoneyLogs({
        page: pagination.current,
        limit: pagination.size,
        uid: appliedSearch.uid || undefined,
        type: appliedSearch.type || undefined
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { type?: string; uid?: string }) => {
    appliedSearch.uid = params.uid?.trim() || ''
    appliedSearch.type = params.type || ''
    pagination.current = 1
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.uid = ''
    appliedSearch.type = ''
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
