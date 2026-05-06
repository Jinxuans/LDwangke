<template>
  <div class="admin-rank-suppliers-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <span class="text-base font-semibold text-g-900">货源排行</span>
            <ElTag effect="plain">活跃货源 {{ activeSupplierCount }} 个</ElTag>
            <ElTag type="success" effect="plain">今日销量 {{ todayTotalCount }} 单</ElTag>
            <ElTag type="warning" effect="plain">榜首 {{ topSupplierName }}</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>
      <ArtTable :loading="loading" :data="list" :columns="columns" :show-table-header="true" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacySupplierRanking,
    type LegacySupplierRankItem
  } from '@/api/legacy/admin-stats'

  defineOptions({ name: 'AdminRankSuppliersPage' })

  const loading = ref(false)
  const list = ref<LegacySupplierRankItem[]>([])

  const activeSupplierCount = computed(() => list.value.filter((item) => Number(item.total_count) > 0).length)
  const todayTotalCount = computed(() => list.value.reduce((sum, item) => sum + Number(item.today_count || 0), 0))
  const topSupplierName = computed(() => list.value[0]?.name || '暂无')

  const { columns, columnChecks } = useTableColumns<LegacySupplierRankItem>(() => [
    {
      type: 'index',
      label: '排名',
      width: 80
    },
    {
      prop: 'name',
      label: '货源名称',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名货源'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `HID ${row.hid}`)
        ])
    },
    {
      prop: 'today_count',
      label: '今日销量',
      width: 120,
      align: 'center'
    },
    {
      prop: 'yesterday_count',
      label: '昨日销量',
      width: 120,
      align: 'center'
    },
    {
      prop: 'total_count',
      label: '累计销量',
      width: 120,
      align: 'center',
      formatter: (row) =>
        h(
          'span',
          { class: row.total_count === list.value[0]?.total_count ? 'font-semibold text-[var(--el-color-warning)]' : '' },
          row.total_count || 0
        )
    }
  ])

  const loadData = async () => {
    loading.value = true
    try {
      const result = await fetchLegacySupplierRanking()
      list.value = Array.isArray(result) ? result : []
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
