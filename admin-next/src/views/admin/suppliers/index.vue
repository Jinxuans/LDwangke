<template>
  <div class="admin-suppliers-page art-full-height">
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
          <div class="flex flex-wrap items-center gap-2">
            <ElTag effect="plain">货源管理</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增货源</ElButton>
            <ElTag effect="plain">共 {{ filteredList.length }} 条</ElTag>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="loading" :data="filteredList" :columns="columns" :showTableHeader="true" />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑货源' : '新增货源'"
      width="900px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">货源基础信息</h3>
            <p class="mt-1 text-sm text-g-500">平台、地址与鉴权信息。</p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">平台类型</label>
              <ElSelect
                v-model="editForm.pt"
                class="w-full"
                clearable
                filterable
                placeholder="请选择平台类型"
              >
                <ElOption
                  v-for="(label, key) in platformNames"
                  :key="key"
                  :label="`${label} (${key})`"
                  :value="key"
                />
              </ElSelect>
            </div>

            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">货源名称</label>
              <ElInput v-model="editForm.name" maxlength="60" placeholder="请输入货源名称" />
            </div>

            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
              <ElInput v-model="editForm.url" placeholder="http://example.com/api" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">账号</label>
              <ElInput v-model="editForm.user" placeholder="上游账号" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">密钥</label>
              <ElInput v-model="editForm.pass" placeholder="上游密钥 / pass" show-password />
            </div>

            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
              <ElInput v-model="editForm.token" placeholder="可选，部分平台需要" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <p class="text-lg font-semibold text-g-900">状态</p>
            <p class="mt-1 text-sm text-g-500">启停与当前摘要。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">货源状态</label>
              <ElSegmented
                v-model="editForm.status"
                :options="[
                  { label: '启用', value: '1' },
                  { label: '禁用', value: '0' }
                ]"
                class="w-full"
              />
            </div>
          </div>

          <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
            <div class="flex flex-wrap gap-2">
              <ElTag :type="editForm.status === '1' ? 'success' : 'info'" effect="plain">
                {{ editForm.status === '1' ? '启用中' : '已禁用' }}
              </ElTag>
              <ElTag type="primary" effect="plain">{{ getPlatformLabel(editForm.pt) }}</ElTag>
              <ElTag :type="editForm.token ? 'warning' : 'info'" effect="plain">
                {{ editForm.token ? '含 Token' : '无 Token' }}
              </ElTag>
            </div>
            <p class="mt-3 text-sm font-medium text-g-900">{{ editForm.name || '未命名货源' }}</p>
            <p class="mt-1 text-xs leading-5 text-g-500">{{ editForm.url || '未填写接口地址' }}</p>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存货源</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="importVisible" title="一键对接" width="720px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">对接参数</h3>
            <p class="mt-1 text-sm text-g-500">倍率、分类与导入名称。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">货源名称</label>
              <ElInput :model-value="importForm.name" disabled />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
              <ElInputNumber
                v-model="importForm.pricee"
                class="w-full"
                :min="0.01"
                :max="100"
                :step="0.1"
                :precision="2"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">分类筛选</label>
              <ElInput v-model="importForm.category" placeholder="999999 表示全部分类" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">新建分类名称</label>
              <ElInput v-model="importForm.importName" placeholder="留空则默认使用货源名称" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <p class="text-lg font-semibold text-g-900">导入模式</p>
            <p class="mt-1 text-sm text-g-500">选择同步方式。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">同步模式</label>
              <ElSegmented
                v-model="importForm.fd"
                :options="[
                  { label: '全量', value: 0 },
                  { label: '仅更新已有', value: 1 }
                ]"
                class="w-full"
              />
            </div>
          </div>

          <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
            <div class="flex flex-wrap gap-2">
              <ElTag effect="plain">{{ importForm.name || '未选择货源' }}</ElTag>
              <ElTag type="primary" effect="plain">{{ importForm.pricee.toFixed(2) }} 倍</ElTag>
              <ElTag :type="importForm.fd === 0 ? 'success' : 'info'" effect="plain">
                {{ importForm.fd === 0 ? '全量导入' : '仅更新已有' }}
              </ElTag>
            </div>
            <p class="mt-3 text-xs leading-5 text-g-500">
              分类 {{ importForm.category || '999999' }}，新分类名 {{ importForm.importName || importForm.name || '-' }}。
            </p>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="importVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="importing" @click="handleImport">开始对接</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteLegacyAdminSupplier,
    fetchLegacyAdminSuppliers,
    fetchLegacyPlatformNames,
    importLegacySupplier,
    queryLegacySupplierBalance,
    saveLegacyAdminSupplier,
    syncLegacySupplierStatus,
    type LegacyAdminSupplier
  } from '@/api/legacy/admin-suppliers'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminSuppliersPage' })

  const loading = ref(false)
  const saving = ref(false)
  const importing = ref(false)
  const dialogVisible = ref(false)
  const importVisible = ref(false)
  const list = ref<LegacyAdminSupplier[]>([])
  const platformNames = ref<Record<string, string>>({})
  const balanceLoadingMap = reactive<Record<number, boolean>>({})
  const syncLoadingMap = reactive<Record<number, boolean>>({})

  const searchForm = ref<{
    keyword?: string
    pt?: string
    status?: string
  }>({
    keyword: undefined,
    pt: undefined,
    status: undefined
  })

  const appliedSearch = reactive({
    keyword: undefined as string | undefined,
    pt: undefined as string | undefined,
    status: undefined as string | undefined
  })

  const editForm = reactive({
    hid: 0,
    pt: '',
    name: '',
    url: '',
    user: '',
    pass: '',
    token: '',
    status: '1'
  })

  const importForm = reactive({
    hid: 0,
    name: '',
    importName: '',
    pricee: 1,
    category: '999999',
    fd: 0
  })

  const isEditing = computed(() => editForm.hid > 0)

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索货源名称 / 地址 / 账号'
      }
    },
    {
      label: '平台',
      key: 'pt',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '全部平台',
        options: Object.entries(platformNames.value).map(([value, label]) => ({
          label: `${label} (${value})`,
          value
        }))
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '启用', value: '1' },
          { label: '禁用', value: '0' }
        ]
      }
    }
  ])

  const filteredList = computed(() => {
    const keyword = appliedSearch.keyword?.toLowerCase() || ''
    return [...list.value].filter((item) => {
      const matchesKeyword =
        !keyword ||
        item.name?.toLowerCase().includes(keyword) ||
        item.url?.toLowerCase().includes(keyword) ||
        item.user?.toLowerCase().includes(keyword)

      const matchesPlatform = !appliedSearch.pt || item.pt === appliedSearch.pt
      const matchesStatus = !appliedSearch.status || item.status === appliedSearch.status

      return matchesKeyword && matchesPlatform && matchesStatus
    })
  })

  const getPlatformLabel = (pt?: string) => platformNames.value[pt || ''] || pt || '未配置平台'
  const formatMoney = (value?: string) => {
    const amount = Number(value || 0)
    return Number.isFinite(amount) ? `¥${amount.toFixed(2)}` : value || '-'
  }

  const { columns, columnChecks } = useTableColumns<LegacyAdminSupplier>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'name',
      label: '货源信息',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名货源'),
          h('p', { class: 'mt-1 text-xs text-g-500 line-clamp-1' }, row.url || '未填写接口地址'),
          h('p', { class: 'mt-1 text-xs text-g-400' }, `HID ${row.hid} / 账号 ${row.user || '-'}`)
        ])
    },
    {
      prop: 'pt',
      label: '平台',
      width: 130,
      formatter: (row) => h(ElTag, { type: 'primary' }, () => getPlatformLabel(row.pt))
    },
    {
      prop: 'money',
      label: '余额',
      width: 140,
      formatter: (row) =>
        h('div', { class: 'flex items-center gap-2' }, [
          h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, formatMoney(row.money)),
          h(
            'button',
            {
              class:
                'rounded-md border border-[var(--el-color-primary-light-6)] bg-[var(--el-color-primary-light-9)] px-2 py-1 text-xs text-[var(--el-color-primary)] transition hover:bg-[var(--el-color-primary-light-8)]',
              onClick: () => handleQueryBalance(row)
            },
            balanceLoadingMap[row.hid] ? '查询中' : '刷新'
          )
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: row.status === '1' ? 'success' : 'info' }, () =>
          row.status === '1' ? '启用' : '禁用'
        )
    },
    {
      prop: 'addtime',
      label: '添加时间',
      width: 180
    },
    {
      prop: 'operation',
      label: '操作',
      minWidth: 260,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap items-center gap-2' }, [
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => openEditDialog(row)
          }),
          h(
            'button',
            {
              class:
                'rounded-md border border-[var(--el-color-primary-light-6)] bg-[var(--el-color-primary-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-primary)] transition hover:bg-[var(--el-color-primary-light-8)]',
              onClick: () => openImportDialog(row)
            },
            '对接'
          ),
          h(
            'button',
            {
              class:
                'rounded-md border border-[var(--art-card-border)] bg-[var(--el-fill-color-light)] px-3 py-1.5 text-xs text-g-700 transition hover:bg-g-100/70',
              onClick: () => handleSyncStatus(row)
            },
            syncLoadingMap[row.hid] ? '同步中' : '同步状态'
          ),
          h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDelete(row)
          })
        ])
    }
  ])

  const resetEditForm = () => {
    editForm.hid = 0
    editForm.pt = ''
    editForm.name = ''
    editForm.url = ''
    editForm.user = ''
    editForm.pass = ''
    editForm.token = ''
    editForm.status = '1'
  }

  const loadData = async () => {
    loading.value = true
    try {
      const [suppliers, platforms] = await Promise.all([
        fetchLegacyAdminSuppliers(),
        fetchLegacyPlatformNames().catch(() => ({}))
      ])
      list.value = Array.isArray(suppliers) ? suppliers : []
      platformNames.value = platforms || {}
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { keyword?: string; pt?: string; status?: string }) => {
    appliedSearch.keyword = params.keyword?.trim() || undefined
    appliedSearch.pt = params.pt || undefined
    appliedSearch.status = params.status || undefined
  }

  const handleReset = () => {
    appliedSearch.keyword = undefined
    appliedSearch.pt = undefined
    appliedSearch.status = undefined
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyAdminSupplier) => {
    editForm.hid = record.hid
    editForm.pt = record.pt || ''
    editForm.name = record.name || ''
    editForm.url = record.url || ''
    editForm.user = record.user || ''
    editForm.pass = record.pass || ''
    editForm.token = record.token || ''
    editForm.status = record.status || '1'
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.name.trim()) {
      ElMessage.warning('请先填写货源名称')
      return
    }

    if (!editForm.url.trim()) {
      ElMessage.warning('请先填写 API 地址')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminSupplier({
        hid: editForm.hid || undefined,
        pt: editForm.pt,
        name: editForm.name.trim(),
        url: editForm.url.trim(),
        user: editForm.user.trim(),
        pass: editForm.pass.trim(),
        token: editForm.token.trim(),
        status: editForm.status
      })
      ElMessage.success(editForm.hid ? '货源已更新' : '货源已创建')
      dialogVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (record: LegacyAdminSupplier) => {
    await ElMessageBox.confirm(`确定删除货源「${record.name || record.hid}」吗？`, '删除货源', {
      type: 'warning'
    })
    await deleteLegacyAdminSupplier(record.hid)
    ElMessage.success('货源已删除')
    await loadData()
  }

  const handleQueryBalance = async (record: LegacyAdminSupplier) => {
    balanceLoadingMap[record.hid] = true
    try {
      const result = await queryLegacySupplierBalance(record.hid)
      const target = list.value.find((item) => item.hid === record.hid)
      if (target) {
        target.money = result.money || target.money
      }
      ElMessage.success(`${record.name}: 当前余额 ${result.money || '-'}`)
    } finally {
      balanceLoadingMap[record.hid] = false
    }
  }

  const handleSyncStatus = async (record: LegacyAdminSupplier) => {
    syncLoadingMap[record.hid] = true
    try {
      const result = await syncLegacySupplierStatus(record.hid)
      ElMessage.success(result.msg || '同步完成')
    } finally {
      syncLoadingMap[record.hid] = false
    }
  }

  const openImportDialog = (record: LegacyAdminSupplier) => {
    importForm.hid = record.hid
    importForm.name = record.name || ''
    importForm.importName = record.name || ''
    importForm.pricee = 1
    importForm.category = '999999'
    importForm.fd = 0
    importVisible.value = true
  }

  const handleImport = async () => {
    if (!importForm.hid) {
      ElMessage.warning('请先选择货源')
      return
    }

    importing.value = true
    try {
      const result = await importLegacySupplier({
        hid: importForm.hid,
        pricee: importForm.pricee,
        category: importForm.category.trim() || '999999',
        name: importForm.importName.trim() || importForm.name,
        fd: importForm.fd
      })
      ElMessage.success(result.msg || '对接完成')
      importVisible.value = false
    } finally {
      importing.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
