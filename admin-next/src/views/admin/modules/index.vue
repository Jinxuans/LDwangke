<template>
  <div class="admin-modules-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card !pb-0">
      <ElTabs v-model="activeType" class="px-5 pt-3" @tab-change="handleTypeChange">
        <ElTabPane
          v-for="item in typeTabs"
          :key="item.value"
          :label="`${item.label} (${typeCountMap[item.value] || 0})`"
          :name="item.value"
        />
      </ElTabs>

      <div class="px-5 pb-5">
        <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
          <template #left>
            <ElSpace wrap>
              <ElTag effect="plain">模块管理</ElTag>
              <ElTag effect="plain">当前分组 {{ activeTypeLabel }}</ElTag>
              <ElTag type="success" effect="plain">启用 {{ enabledCount }}</ElTag>
              <ElTag type="info" effect="plain">禁用 {{ disabledCount }}</ElTag>
              <ElButton type="primary" plain @click="openCreateDialog">新增模块</ElButton>
            </ElSpace>
          </template>
        </ArtTableHeader>

        <ArtTable :loading="loading" :data="filteredList" :columns="columns" :show-table-header="true" />
      </div>
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑模块' : '新增模块'"
      width="980px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">基础信息</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">
              模块标识、分类、名称和显示价格会直接决定后台菜单与前台大厅展示。
            </p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">模块标识</label>
              <ElInput
                v-model="editForm.app_id"
                :disabled="isEditing"
                maxlength="60"
                placeholder="例如 yyd / appui"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">模块类型</label>
              <ElSelect v-model="editForm.type" class="w-full">
                <ElOption
                  v-for="item in typeTabs.filter((tab) => tab.value !== 'all')"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">模块名称</label>
              <ElInput v-model="editForm.name" maxlength="120" placeholder="请输入模块名称" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">展示价格</label>
              <ElInput v-model="editForm.price" maxlength="40" placeholder="例如 0.5元/次" />
            </div>

            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">模块描述</label>
              <ElInput
                v-model="editForm.description"
                type="textarea"
                :rows="4"
                resize="none"
                maxlength="240"
                placeholder="填写模块简介，用于后台和大厅预览"
              />
            </div>

            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">图标</label>
              <ElInput v-model="editForm.icon" maxlength="80" placeholder="建议填写 Iconify 图标名，例如 ri:apps-2-line" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">入口设置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">维护访问路径、状态和扩展配置。</p>
          </div>

          <div class="mt-5 space-y-4">
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="flex items-start gap-3">
                <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-[var(--el-color-primary-light-9)] text-lg text-[var(--el-color-primary)]">
                  <ArtSvgIcon :icon="previewIcon" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <p class="truncate text-sm font-semibold text-g-900">{{ editForm.name || '未命名模块' }}</p>
                    <ElTag type="primary" effect="plain">{{ typeLabelMap[editForm.type] || '未分类' }}</ElTag>
                    <ElTag :type="Number(editForm.status) === 1 ? 'success' : 'info'" effect="plain">
                      {{ Number(editForm.status) === 1 ? '已启用' : '已禁用' }}
                    </ElTag>
                  </div>
                  <p class="mt-2 text-sm text-g-500">{{ editForm.description || '暂无模块说明' }}</p>
                </div>
              </div>
            </article>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">API 路径</label>
              <ElInput v-model="editForm.api_base" maxlength="255" placeholder="/module/service/api.php" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">前端页面路径</label>
              <ElInput v-model="editForm.view_url" maxlength="255" placeholder="/module/view/index.php" />
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
                <ElInputNumber v-model="editForm.sort" class="w-full" :min="0" :max="999999" />
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
                <ElSegmented
                  v-model="editForm.status"
                  :options="[
                    { label: '启用', value: 1 },
                    { label: '禁用', value: 0 }
                  ]"
                  class="w-full"
                />
              </div>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">扩展配置（JSON）</label>
              <ElInput
                v-model="editForm.config"
                type="textarea"
                :rows="7"
                resize="none"
                placeholder="默认 {}，可填写模块扩展配置"
              />
            </div>

            <div class="space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">模块标识</span>
                <span class="font-medium text-g-900">{{ editForm.app_id || '未填写' }}</span>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">展示价格</span>
                <span class="font-medium text-g-900">{{ editForm.price || '未填写' }}</span>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">排序</span>
                <span class="font-medium text-g-900">{{ editForm.sort ?? '-' }}</span>
              </div>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存模块</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteLegacyDynamicModule,
    fetchLegacyDynamicModules,
    saveLegacyDynamicModule,
    type LegacyDynamicModule
  } from '@/api/legacy/admin-modules'
  import { ElMessage, ElMessageBox, ElSwitch, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminModulesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const dialogVisible = ref(false)

  const list = ref<LegacyDynamicModule[]>([])
  const activeType = ref('all')

  const searchForm = ref({
    keyword: '',
    status: ''
  })

  const appliedSearch = reactive({
    keyword: '',
    status: ''
  })

  const editForm = reactive<LegacyDynamicModule>({
    id: 0,
    app_id: '',
    type: 'sport',
    name: '',
    description: '',
    price: '',
    icon: '',
    api_base: '',
    view_url: '',
    status: 1,
    sort: 10,
    config: '{}'
  })

  const typeTabs = [
    { label: '全部', value: 'all' },
    { label: '运动', value: 'sport' },
    { label: '实习', value: 'intern' },
    { label: '论文', value: 'paper' }
  ]

  const typeLabelMap: Record<string, string> = {
    sport: '运动',
    intern: '实习',
    paper: '论文'
  }

  const isEditing = computed(() => editForm.id > 0)
  const enabledCount = computed(() => list.value.filter((item) => Number(item.status) === 1).length)
  const disabledCount = computed(() => list.value.filter((item) => Number(item.status) !== 1).length)
  const activeTypeLabel = computed(
    () => typeTabs.find((item) => item.value === activeType.value)?.label || '全部'
  )
  const previewIcon = computed(() => normalizeIcon(editForm.icon))
  const typeCountMap = computed(() => {
    const result: Record<string, number> = {
      all: list.value.length,
      sport: 0,
      intern: 0,
      paper: 0
    }

    list.value.forEach((item) => {
      const key = item.type || 'sport'
      result[key] = (result[key] || 0) + 1
    })
    return result
  })

  const filteredList = computed(() =>
    list.value.filter((item) => {
      if (activeType.value !== 'all' && item.type !== activeType.value) {
        return false
      }

      if (appliedSearch.status !== '' && String(item.status) !== appliedSearch.status) {
        return false
      }

      if (!appliedSearch.keyword) {
        return true
      }

      const keyword = appliedSearch.keyword.toLowerCase()
      return [item.name, item.app_id, item.description, item.api_base, item.view_url]
        .join(' ')
        .toLowerCase()
        .includes(keyword)
    })
  )

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索模块名、标识或路径'
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

  const normalizeIcon = (icon?: string) => {
    return icon && icon.includes(':') ? icon : 'ri:apps-2-line'
  }

  const statusTagType = (status: number) => (Number(status) === 1 ? 'success' : 'info')

  const { columns, columnChecks } = useTableColumns<LegacyDynamicModule>(() => [
    {
      prop: 'id',
      label: 'ID',
      width: 80,
      align: 'center'
    },
    {
      prop: 'name',
      label: '模块信息',
      minWidth: 280,
      formatter: (row) =>
        h('div', { class: 'flex items-start gap-3 py-1' }, [
          h(
            'div',
            {
              class:
                'flex h-11 w-11 items-center justify-center rounded-xl bg-[var(--el-color-primary-light-9)] text-lg text-[var(--el-color-primary)]'
            },
            [h(ArtSvgIcon, { icon: normalizeIcon(row.icon) })]
          ),
          h('div', { class: 'min-w-0 flex-1 leading-6' }, [
            h('p', { class: 'truncate font-semibold text-g-900' }, row.name || '未命名模块'),
            h('p', { class: 'mt-1 truncate text-xs text-g-500' }, `app_id: ${row.app_id || '-'}`),
            h('p', { class: 'mt-1 line-clamp-2 text-xs text-g-500' }, row.description || '暂无模块描述')
          ])
        ])
    },
    {
      prop: 'type',
      label: '类型',
      width: 110,
      formatter: (row) =>
        h(ElTag, { type: 'primary', effect: 'plain' }, () => typeLabelMap[row.type] || row.type)
    },
    {
      prop: 'api_base',
      label: '入口配置',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'truncate text-xs text-g-500' }, `API: ${row.api_base || '-'}`),
          h('p', { class: 'mt-1 truncate text-xs text-g-500' }, `页面: ${row.view_url || '-'}`)
        ])
    },
    {
      prop: 'price',
      label: '展示价格',
      width: 120,
      formatter: (row) => row.price || '-'
    },
    {
      prop: 'sort',
      label: '排序',
      width: 90,
      align: 'center'
    },
    {
      prop: 'status',
      label: '状态',
      width: 150,
      formatter: (row) =>
        h('div', { class: 'flex items-center gap-2' }, [
          h(ElSwitch, {
            modelValue: Number(row.status) === 1,
            size: 'small',
            onChange: (value: string | number | boolean) => handleToggleStatus(row, Boolean(value))
          }),
          h(ElTag, { type: statusTagType(row.status), effect: 'plain' }, () =>
            Number(row.status) === 1 ? '启用' : '禁用'
          )
        ])
    },
    {
      prop: 'operation',
      label: '操作',
      width: 140,
      fixed: 'right',
      formatter: (row) =>
        h('div', [
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => openEditDialog(row)
          }),
          h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDelete(row)
          })
        ])
    }
  ])

  const resetEditForm = () => {
    editForm.id = 0
    editForm.app_id = ''
    editForm.type = activeType.value === 'all' ? 'sport' : activeType.value
    editForm.name = ''
    editForm.description = ''
    editForm.price = ''
    editForm.icon = ''
    editForm.api_base = ''
    editForm.view_url = ''
    editForm.status = 1
    editForm.sort = 10
    editForm.config = '{}'
  }

  const loadData = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyDynamicModules()
      list.value = Array.isArray(result) ? result : []
    } finally {
      loading.value = false
    }
  }

  const refreshData = async () => {
    await loadData()
  }

  const handleSearch = (params: { keyword?: string; status?: string }) => {
    appliedSearch.keyword = params.keyword?.trim() || ''
    appliedSearch.status = params.status || ''
  }

  const handleReset = () => {
    appliedSearch.keyword = ''
    appliedSearch.status = ''
  }

  const handleTypeChange = (name: string | number) => {
    activeType.value = String(name)
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyDynamicModule) => {
    editForm.id = record.id
    editForm.app_id = record.app_id || ''
    editForm.type = record.type || 'sport'
    editForm.name = record.name || ''
    editForm.description = record.description || ''
    editForm.price = record.price || ''
    editForm.icon = record.icon || ''
    editForm.api_base = record.api_base || ''
    editForm.view_url = record.view_url || ''
    editForm.status = Number(record.status || 0)
    editForm.sort = Number(record.sort || 0)
    editForm.config = record.config || '{}'
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.app_id.trim()) {
      ElMessage.warning('请先填写模块标识')
      return
    }
    if (!editForm.name.trim()) {
      ElMessage.warning('请先填写模块名称')
      return
    }

    const configText = editForm.config?.trim() || '{}'
    try {
      JSON.parse(configText)
    } catch {
      ElMessage.warning('扩展配置必须是合法 JSON')
      return
    }

    saving.value = true
    try {
      await saveLegacyDynamicModule({
        id: editForm.id || undefined,
        app_id: editForm.app_id.trim(),
        type: editForm.type || 'sport',
        name: editForm.name.trim(),
        description: editForm.description.trim(),
        price: editForm.price.trim(),
        icon: editForm.icon.trim(),
        api_base: editForm.api_base.trim(),
        view_url: editForm.view_url.trim(),
        status: Number(editForm.status || 0),
        sort: Number(editForm.sort || 0),
        config: configText
      })
      ElMessage.success(editForm.id ? '模块已更新' : '模块已创建')
      dialogVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  const handleToggleStatus = async (row: LegacyDynamicModule, enabled: boolean) => {
    await saveLegacyDynamicModule({
      ...row,
      status: enabled ? 1 : 0
    })
    row.status = enabled ? 1 : 0
    ElMessage.success(enabled ? '模块已启用' : '模块已禁用')
  }

  const handleDelete = async (row: LegacyDynamicModule) => {
    await ElMessageBox.confirm(`确定删除模块「${row.name || row.id}」吗？`, '删除模块', {
      type: 'warning'
    })
    await deleteLegacyDynamicModule(row.id)
    ElMessage.success('模块已删除')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>
