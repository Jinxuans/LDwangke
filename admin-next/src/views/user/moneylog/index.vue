<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <ElCard class="art-table-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d px-5 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">资金明细</h2>
          <ElTag effect="plain">余额流水</ElTag>
          <ElTag effect="plain">共 {{ pagination.total }} 条</ElTag>
          <ElTag type="success" effect="plain">当前页 {{ list.length }} 条</ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElSelect
            v-model="typeFilter"
            clearable
            class="w-[180px]"
            placeholder="按类型筛选"
            @change="handleFilterChange"
          >
            <ElOption label="充值" value="充值" />
            <ElOption label="充值赠送" value="充值赠送" />
            <ElOption label="扣费" value="扣费" />
            <ElOption label="退款" value="退款" />
            <ElOption label="调整" value="调整" />
            <ElOption label="商城代收" value="商城代收" />
            <ElOption label="提现申请" value="提现申请" />
            <ElOption label="提现通过" value="提现通过" />
            <ElOption label="提现驳回" value="提现驳回" />
          </ElSelect>
          <ElButton plain :loading="loading" @click="loadData(pagination.page)">刷新</ElButton>
        </div>
      </div>

      <ElTable :data="list" v-loading="loading" stripe class="w-full">
        <ElTableColumn prop="addtime" label="时间" width="180" />
        <ElTableColumn label="类型" width="140">
          <template #default="{ row }">
            <ElTag :type="getTypeTag(row.type)">{{ row.type }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="金额" width="140" align="right">
          <template #default="{ row }">
            <span
              :class="Number(row.money) >= 0 ? 'text-[var(--el-color-success)]' : 'text-[var(--el-color-danger)]'"
              class="font-semibold"
            >
              {{ Number(row.money) >= 0 ? '+' : '' }}{{ Number(row.money).toFixed(2) }}
            </span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="变动后余额" width="150" align="right">
          <template #default="{ row }">
            <span class="font-medium text-g-800">¥{{ Number(row.balance).toFixed(2) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="remark" label="备注" min-width="320" show-overflow-tooltip />
      </ElTable>

      <div class="flex justify-end border-t-d px-5 py-4">
        <ElPagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.limit"
          :page-sizes="[20, 50, 100]"
          :total="pagination.total"
          @current-change="loadData"
          @size-change="handleSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { fetchLegacyMoneyLogs, type LegacyMoneyLog } from '@/api/legacy/user-center'

  defineOptions({ name: 'UserMoneyLogPage' })

  const loading = ref(false)
  const list = ref<LegacyMoneyLog[]>([])
  const typeFilter = ref('')
  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const getTypeTag = (type: string): 'success' | 'danger' | 'warning' | 'info' | 'primary' => {
    if (['充值', '充值赠送', '商城代收', '提现驳回'].includes(type)) return 'success'
    if (['扣费', '提现申请'].includes(type)) return 'danger'
    if (['退款'].includes(type)) return 'primary'
    if (['提现通过'].includes(type)) return 'warning'
    return 'info'
  }

  const loadData = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyMoneyLogs({
        page: pagination.page,
        limit: pagination.limit,
        type: typeFilter.value || undefined
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const handleFilterChange = async () => {
    await loadData(1)
  }

  const handleSizeChange = async (size: number) => {
    pagination.limit = size
    await loadData(1)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
