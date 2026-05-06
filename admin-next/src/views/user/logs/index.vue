<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <ElCard class="art-table-card">
      <div class="border-b-d px-5 py-4">
        <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData(pagination.page)">
          <template #left>
            <ElSpace wrap>
              <ElTag effect="plain">操作日志</ElTag>
              <ElTag type="info" effect="plain">共 {{ pagination.total }} 条</ElTag>
            </ElSpace>
          </template>
        </ArtTableHeader>

        <div class="mt-4 flex flex-wrap gap-3">
          <ElSelect v-model="filterType" clearable class="w-full sm:w-[150px]" placeholder="搜索字段">
            <ElOption label="UID" value="uid" />
            <ElOption label="类型" value="type" />
            <ElOption label="内容" value="text" />
            <ElOption label="金额" value="money" />
            <ElOption label="IP" value="ip" />
          </ElSelect>
          <ElInput
            v-model="keywords"
            clearable
            class="w-full lg:w-[260px]"
            placeholder="输入关键词"
            @keyup.enter="handleSearch"
          />
          <ElButton type="primary" @click="handleSearch">搜索</ElButton>
          <ElButton plain @click="handleReset">重置</ElButton>
        </div>
      </div>

      <ElTable :data="list" v-loading="loading" stripe class="w-full">
        <ElTableColumn prop="id" label="ID" width="90" />
        <ElTableColumn label="类型" width="120">
          <template #default="{ row }">
            <ElTag effect="plain">{{ row.type || '未分类' }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="详情" min-width="340">
          <template #default="{ row }">
            <p class="whitespace-pre-wrap text-sm leading-6 text-g-700">{{ row.text || '-' }}</p>
          </template>
        </ElTableColumn>
        <ElTableColumn label="金额" width="130" align="right">
          <template #default="{ row }">
            <span
              :class="Number(row.money) >= 0 ? 'text-[var(--el-color-success)]' : 'text-[var(--el-color-danger)]'"
              class="font-semibold"
            >
              {{ Number(row.money) >= 0 ? '+' : '' }}{{ Number(row.money).toFixed(2) }}
            </span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="余额" width="130" align="right">
          <template #default="{ row }">
            <span class="font-medium text-g-800">¥{{ Number(row.smoney).toFixed(2) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="ip" label="IP" width="150" />
        <ElTableColumn prop="addtime" label="时间" width="180" />
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
  import { fetchLegacyUserLogs, type LegacyUserLogItem } from '@/api/legacy/user-center'

  defineOptions({ name: 'UserLogsPage' })

  const loading = ref(false)
  const list = ref<LegacyUserLogItem[]>([])
  const filterType = ref('')
  const keywords = ref('')
  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const loadData = async (page = pagination.page) => {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyUserLogs({
        page: pagination.page,
        limit: pagination.limit,
        type: filterType.value || undefined,
        keywords: keywords.value.trim() || undefined
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = async () => {
    await loadData(1)
  }

  const handleReset = async () => {
    filterType.value = ''
    keywords.value = ''
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
