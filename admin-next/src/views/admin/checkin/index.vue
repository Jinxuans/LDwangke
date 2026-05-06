<template>
  <div class="admin-checkin-page art-full-height">
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
            <ElTag effect="plain">统计日期 {{ appliedSearch.date }}</ElTag>
            <ElTag type="success" effect="plain">签到人数 {{ totalUsers }}</ElTag>
            <ElTag type="warning" effect="plain">奖励发放 ¥{{ totalReward }}</ElTag>
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
  import {
    fetchLegacyAdminCheckinStats,
    type LegacyAdminCheckinRecord
  } from '@/api/legacy/admin-auxiliary'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'AdminCheckinPage' })

  const getCurrentDate = () => {
    const date = new Date()
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }

  const loading = ref(false)
  const list = ref<LegacyAdminCheckinRecord[]>([])
  const totalUsers = ref(0)
  const totalReward = ref('0.00')

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    date: getCurrentDate()
  })

  const appliedSearch = reactive({
    date: getCurrentDate()
  })

  const searchItems = computed(() => [
    {
      label: '统计日期',
      key: 'date',
      type: 'date',
      props: {
        clearable: false,
        valueFormat: 'YYYY-MM-DD',
        placeholder: '选择统计日期'
      }
    }
  ])

  const { columns, columnChecks } = useTableColumns<LegacyAdminCheckinRecord>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'uid',
      label: 'UID',
      width: 90,
      align: 'center'
    },
    {
      prop: 'username',
      label: '用户',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.username || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `UID ${row.uid}`)
        ])
    },
    {
      prop: 'reward_money',
      label: '奖励金额',
      width: 120,
      align: 'right',
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-warning)]' }, `¥${Number(row.reward_money || 0).toFixed(2)}`)
    },
    {
      prop: 'addtime',
      label: '签到时间',
      width: 180
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminCheckinStats({
        date: appliedSearch.date,
        page: pagination.current,
        limit: pagination.size
      })
      list.value = result.list || []
      totalUsers.value = Number(result.total_users || 0)
      totalReward.value = Number(result.total_reward || 0).toFixed(2)
      pagination.total = Number(result.total || 0)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { date?: string }) => {
    appliedSearch.date = params.date || getCurrentDate()
    pagination.current = 1
    loadData(1)
  }

  const handleReset = () => {
    const currentDate = getCurrentDate()
    searchForm.value.date = currentDate
    appliedSearch.date = currentDate
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
