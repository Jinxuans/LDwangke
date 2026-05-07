<template>
  <div class="user-logs-page art-full-height">
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
            <ElTag effect="plain">操作日志</ElTag>
            <ElTag type="info" effect="plain">共 {{ pagination.total }} 条</ElTag>
            <ElTag v-if="appliedSearch.type" type="primary" effect="plain">字段 {{ appliedSearch.type }}</ElTag>
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
  import { fetchLegacyUserLogs, type LegacyUserLogItem } from '@/api/legacy/user-center'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'UserLogsPage' })

  const loading = ref(false)
  const list = ref<LegacyUserLogItem[]>([])
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    type: '',
    keywords: ''
  })

  const appliedSearch = reactive({
    type: '',
    keywords: ''
  })

  const searchItems = computed(() => [
    {
      label: '搜索字段',
      key: 'type',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部字段',
        options: [
          { label: 'UID', value: 'uid' },
          { label: '类型', value: 'type' },
          { label: '内容', value: 'text' },
          { label: '金额', value: 'money' },
          { label: 'IP', value: 'ip' }
        ]
      }
    },
    {
      label: '关键词',
      key: 'keywords',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入关键词'
      }
    }
  ])

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const { columns, columnChecks } = useTableColumns<LegacyUserLogItem>(() => [
    { prop: 'id', label: 'ID', width: 90 },
    {
      prop: 'type',
      label: '类型',
      width: 120,
      formatter: (row) => h(ElTag, { effect: 'plain' }, () => row.type || '未分类')
    },
    {
      prop: 'text',
      label: '详情',
      minWidth: 340,
      formatter: (row) => h('p', { class: 'whitespace-pre-wrap text-sm leading-6 text-g-700' }, row.text || '-')
    },
    {
      prop: 'money',
      label: '金额',
      width: 130,
      align: 'right',
      formatter: (row) =>
        h(
          'span',
          { class: Number(row.money) >= 0 ? 'font-semibold text-[var(--el-color-success)]' : 'font-semibold text-[var(--el-color-danger)]' },
          `${Number(row.money) >= 0 ? '+' : ''}${formatMoney(row.money)}`
        )
    },
    {
      prop: 'smoney',
      label: '余额',
      width: 130,
      align: 'right',
      formatter: (row) => h('span', { class: 'font-medium text-g-800' }, `¥${formatMoney(row.smoney)}`)
    },
    { prop: 'ip', label: 'IP', width: 150 },
    { prop: 'addtime', label: '时间', width: 180 }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyUserLogs({
        page: pagination.current,
        limit: pagination.size,
        type: appliedSearch.type || undefined,
        keywords: appliedSearch.keywords || undefined
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { type?: string; keywords?: string }) => {
    appliedSearch.type = params.type || ''
    appliedSearch.keywords = params.keywords?.trim() || ''
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.type = ''
    appliedSearch.keywords = ''
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
