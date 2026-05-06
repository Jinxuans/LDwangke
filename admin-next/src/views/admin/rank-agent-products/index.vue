<template>
  <div class="admin-rank-agent-products-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <span class="text-base font-semibold text-g-900">代理商品排行</span>
            <ElSegmented
              v-model="timeType"
              :options="timeOptions"
              :disabled="!hasValidUid"
              @change="handleTimeChange"
            />
            <ElTag effect="plain">代理 UID {{ appliedUid || '-' }}</ElTag>
            <ElTag type="primary" effect="plain">统计范围 {{ timeLabelMap[timeType] }}</ElTag>
            <ElTag type="success" effect="plain">商品数 {{ list.length }}</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        v-if="searched"
        :loading="loading"
        :data="list"
        :columns="columns"
        :show-table-header="true"
      />
      <div v-else class="py-16 text-center text-sm text-g-500">请输入代理 UID 后查询商品排行</div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElSegmented, ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyAgentProductRanking,
    type LegacyAgentProductRankItem
  } from '@/api/legacy/admin-stats'

  defineOptions({ name: 'AdminRankAgentProductsPage' })

  const loading = ref(false)
  const searched = ref(false)
  const list = ref<LegacyAgentProductRankItem[]>([])

  const searchForm = ref({
    uid: ''
  })

  const appliedUid = ref<number>()
  const timeType = ref<'today' | 'yesterday' | 'week' | 'month'>('today')

  const timeLabelMap = {
    today: '今日',
    yesterday: '昨日',
    week: '本周',
    month: '本月'
  }

  const timeOptions = [
    { label: '今日', value: 'today' },
    { label: '昨日', value: 'yesterday' },
    { label: '本周', value: 'week' },
    { label: '本月', value: 'month' }
  ]

  const hasValidUid = computed(() => Number(appliedUid.value || 0) > 0)

  const searchItems = computed(() => [
    {
      label: '代理 UID',
      key: 'uid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入代理 UID'
      }
    }
  ])

  const { columns, columnChecks } = useTableColumns<LegacyAgentProductRankItem>(() => [
    {
      prop: 'rank',
      label: '排名',
      width: 90,
      align: 'center',
      formatter: (row) =>
        h(
          ElTag,
          { type: row.rank <= 3 ? 'warning' : 'info', effect: row.rank <= 3 ? 'dark' : 'plain' },
          () => `#${row.rank}`
        )
    },
    {
      prop: 'ptname',
      label: '商品名称',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.ptname || '未命名商品'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `统计范围 ${timeLabelMap[timeType.value]}`)
        ])
    },
    {
      prop: 'count',
      label: '订单量',
      width: 120,
      align: 'center',
      formatter: (row) => h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, `${row.count} 单`)
    },
    {
      prop: 'latest',
      label: '最后下单时间',
      width: 180,
      formatter: (row) => row.latest || '-'
    }
  ])

  const loadData = async () => {
    if (!hasValidUid.value) {
      return
    }
    loading.value = true
    try {
      const result = await fetchLegacyAgentProductRanking(appliedUid.value!, timeType.value)
      list.value = Array.isArray(result) ? result : []
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { uid?: string }) => {
    const nextUid = Number(params.uid || 0)
    if (!nextUid || nextUid <= 0) {
      ElMessage.warning('请输入有效的代理 UID')
      return
    }
    appliedUid.value = nextUid
    searched.value = true
    loadData()
  }

  const handleReset = () => {
    searchForm.value.uid = ''
    appliedUid.value = undefined
    searched.value = false
    list.value = []
    timeType.value = 'today'
  }

  const handleTimeChange = () => {
    if (hasValidUid.value && searched.value) {
      loadData()
    }
  }
</script>
