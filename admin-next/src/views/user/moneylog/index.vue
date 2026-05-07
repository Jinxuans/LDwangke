<template>
  <div class="user-moneylog-page art-full-height">
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
            <ElTag effect="plain">余额流水</ElTag>
            <ElTag effect="plain">共 {{ pagination.total }} 条</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ list.length }} 条</ElTag>
            <ElTag v-if="appliedSearch.type" type="primary" effect="plain">类型 {{ appliedSearch.type }}</ElTag>
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
  import { h } from 'vue'
  import { ElTag } from 'element-plus'
  import { fetchLegacyMoneyLogs, type LegacyMoneyLog } from '@/api/legacy/user-center'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'UserMoneyLogPage' })

  const logTypeOptions = ['充值', '充值赠送', '扣费', '退款', '调整', '商城代收', '提现申请', '提现通过', '提现驳回']

  const loading = ref(false)
  const list = ref<LegacyMoneyLog[]>([])
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    type: ''
  })

  const appliedSearch = reactive({
    type: ''
  })

  const searchItems = computed(() => [
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

  const getTypeTag = (type: string): 'success' | 'danger' | 'warning' | 'info' | 'primary' => {
    if (['充值', '充值赠送', '商城代收', '提现驳回'].includes(type)) return 'success'
    if (['扣费', '提现申请'].includes(type)) return 'danger'
    if (['退款'].includes(type)) return 'primary'
    if (['提现通过'].includes(type)) return 'warning'
    return 'info'
  }

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const { columns, columnChecks } = useTableColumns<LegacyMoneyLog>(() => [
    {
      prop: 'addtime',
      label: '时间',
      width: 180
    },
    {
      prop: 'type',
      label: '类型',
      width: 140,
      formatter: (row) => h(ElTag, { type: getTypeTag(row.type), effect: 'plain' }, () => row.type || '-')
    },
    {
      prop: 'money',
      label: '金额',
      width: 140,
      align: 'right',
      formatter: (row) =>
        h(
          'span',
          { class: Number(row.money) >= 0 ? 'font-semibold text-[var(--el-color-success)]' : 'font-semibold text-[var(--el-color-danger)]' },
          `${Number(row.money) >= 0 ? '+' : ''}${formatMoney(row.money)}`
        )
    },
    {
      prop: 'balance',
      label: '变动后余额',
      width: 150,
      align: 'right',
      formatter: (row) => h('span', { class: 'font-medium text-g-800' }, `¥${formatMoney(row.balance)}`)
    },
    {
      prop: 'remark',
      label: '备注',
      minWidth: 320,
      formatter: (row) => row.remark || '-'
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyMoneyLogs({
        page: pagination.current,
        limit: pagination.size,
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

  const handleSearch = (params: { type?: string }) => {
    appliedSearch.type = params.type || ''
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.type = ''
    loadData(1)
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = async (size: number) => {
    pagination.size = size
    await loadData(1)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
