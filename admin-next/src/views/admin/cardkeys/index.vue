<template>
  <div class="admin-cardkeys-page art-full-height">
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
          <div class="flex flex-wrap items-center gap-3">
            <ElPopconfirm
              title="确定删除选中的未使用卡密吗？"
              width="260"
              @confirm="handleBatchDelete"
            >
              <template #reference>
                <ElButton plain type="danger" :disabled="selectedIds.length === 0">
                  批量删除
                </ElButton>
              </template>
            </ElPopconfirm>
            <ElButton type="primary" plain @click="openGenerateDialog">生成卡密</ElButton>
            <ElTag effect="plain">已选 {{ selectedIds.length }} 张</ElTag>
            <ElTag type="success" effect="plain">未使用 {{ unusedCount }} 张</ElTag>
            <ElTag type="warning" effect="plain">已使用 {{ usedCount }} 张</ElTag>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog v-model="generateVisible" title="生成卡密" width="520px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">生成参数</h3>
            <p class="mt-1 text-sm text-g-500">面额与数量。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">卡密面额（元）</label>
              <ElInputNumber
                v-model="generateForm.money"
                class="w-full"
                :min="1"
                :max="10000"
                :precision="2"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">生成数量（张）</label>
              <ElInputNumber
                v-model="generateForm.count"
                class="w-full"
                :min="1"
                :max="100"
                :precision="0"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">生成摘要</h3>
            <p class="mt-1 text-sm text-g-500">确认本次批量信息。</p>
          </div>

          <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
            <div class="flex flex-wrap gap-2">
              <ElTag effect="plain">单张 ¥{{ formatMoney(generateForm.money) }}</ElTag>
              <ElTag effect="plain">{{ generateForm.count }} 张</ElTag>
              <ElTag type="primary" effect="plain">
                总额 ¥{{ formatMoney(Number(generateForm.money || 0) * Number(generateForm.count || 0)) }}
              </ElTag>
            </div>
            <p class="mt-3 text-xs leading-5 text-g-500">生成后会弹出结果窗口，可直接复制全部卡密。</p>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="generateVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="generating" @click="handleGenerate">开始生成</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="codesVisible" title="生成结果" width="640px" destroy-on-close>
      <div class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <p class="text-sm text-g-500">共生成 {{ generatedCodes.length }} 张卡密，可直接复制后分发。</p>
          <ElButton plain type="primary" @click="copyCodes">复制全部</ElButton>
        </div>
        <ElInput
          :model-value="generatedCodes.join('\n')"
          type="textarea"
          :rows="12"
          resize="none"
          readonly
        />
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElTag } from 'element-plus'
  import {
    deleteLegacyAdminCardKeys,
    fetchLegacyAdminCardKeys,
    generateLegacyAdminCardKeys,
    type LegacyAdminCardKey
  } from '@/api/legacy/admin-auxiliary'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'AdminCardkeysPage' })

  const loading = ref(false)
  const generating = ref(false)
  const list = ref<LegacyAdminCardKey[]>([])
  const selectedIds = ref<number[]>([])
  const generateVisible = ref(false)
  const codesVisible = ref(false)
  const generatedCodes = ref<string[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    status?: number
  }>({
    status: undefined
  })

  const appliedSearch = reactive({
    status: undefined as number | undefined
  })

  const generateForm = reactive({
    money: 10,
    count: 10
  })

  const unusedCount = computed(() => list.value.filter((item) => Number(item.status) !== 1).length)
  const usedCount = computed(() => list.value.filter((item) => Number(item.status) === 1).length)

  const searchItems = computed(() => [
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '未使用', value: 0 },
          { label: '已使用', value: 1 }
        ]
      }
    }
  ])

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const { columns, columnChecks } = useTableColumns<LegacyAdminCardKey>(() => [
    {
      type: 'selection',
      width: 50,
      selectable: (row: LegacyAdminCardKey) => Number(row.status) !== 1
    },
    {
      prop: 'id',
      label: 'ID',
      width: 70,
      align: 'center'
    },
    {
      prop: 'content',
      label: '卡密内容',
      minWidth: 280,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 break-all' }, row.content || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `面额 ¥${formatMoney(row.money)}`)
        ])
    },
    {
      prop: 'money',
      label: '面额',
      width: 100,
      align: 'right',
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, `¥${formatMoney(row.money)}`)
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: Number(row.status) === 1 ? 'info' : 'success' }, () =>
          Number(row.status) === 1 ? '已使用' : '未使用'
        )
    },
    {
      prop: 'uid',
      label: '使用者',
      width: 110,
      align: 'center',
      formatter: (row) => (row.uid ? `UID ${row.uid}` : '-')
    },
    {
      prop: 'addtime',
      label: '创建时间',
      width: 180
    },
    {
      prop: 'usedtime',
      label: '使用时间',
      width: 180,
      formatter: (row) => row.usedtime || '-'
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminCardKeys({
        page: pagination.current,
        limit: pagination.size,
        status: appliedSearch.status
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
      selectedIds.value = []
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { status?: number }) => {
    appliedSearch.status = typeof params.status === 'number' ? params.status : undefined
    pagination.current = 1
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.status = undefined
    pagination.current = 1
    loadData(1)
  }

  const handleSelectionChange = (rows: LegacyAdminCardKey[]) => {
    selectedIds.value = rows.map((item) => item.id)
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const openGenerateDialog = () => {
    generateForm.money = 10
    generateForm.count = 10
    generateVisible.value = true
  }

  const handleGenerate = async () => {
    if (Number(generateForm.money) < 1) {
      ElMessage.warning('卡密面额至少为 1 元')
      return
    }
    if (Number(generateForm.count) < 1 || Number(generateForm.count) > 100) {
      ElMessage.warning('生成数量需在 1 到 100 之间')
      return
    }

    generating.value = true
    try {
      const result = await generateLegacyAdminCardKeys(generateForm.money, generateForm.count)
      generatedCodes.value = result.codes || []
      generateVisible.value = false
      codesVisible.value = true
      ElMessage.success(`已生成 ${result.count || generatedCodes.value.length} 张卡密`)
      await loadData(1)
    } finally {
      generating.value = false
    }
  }

  const handleBatchDelete = async () => {
    if (!selectedIds.value.length) {
      ElMessage.warning('请先选择要删除的卡密')
      return
    }

    const result = await deleteLegacyAdminCardKeys(selectedIds.value)
    ElMessage.success(`已删除 ${result.deleted || 0} 张卡密`)
    await loadData(pagination.current)
  }

  const copyCodes = async () => {
    if (!generatedCodes.value.length) {
      ElMessage.warning('暂无可复制的卡密')
      return
    }

    try {
      await navigator.clipboard.writeText(generatedCodes.value.join('\n'))
      ElMessage.success('已复制全部卡密')
    } catch {
      ElMessage.warning('复制失败，请手动复制')
    }
  }

  onMounted(() => {
    loadData(1)
  })
</script>
