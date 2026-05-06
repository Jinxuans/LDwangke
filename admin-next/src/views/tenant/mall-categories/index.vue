<template>
  <div class="tenant-mall-categories-page art-full-height">
    <ElCard class="art-table-card">
      <div class="grid gap-4 lg:grid-cols-[1fr_auto]">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">关键词</label>
          <ElInput v-model="keyword" clearable placeholder="搜索分类名称" />
        </div>
        <div class="flex items-end gap-3">
          <ElButton @click="keyword = ''">重置</ElButton>
          <ElButton type="primary" @click="openCreateDialog">新增</ElButton>
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card mt-4">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">商城分类</ElTag>
            <ElTag effect="plain">分类 {{ filteredList.length }} 个</ElTag>
            <ElTag v-if="sortChangedCount" type="warning" effect="plain">待保存排序 {{ sortChangedCount }} 项</ElTag>
            <ElButton plain :disabled="!sortChangedCount" :loading="sortSaving" @click="handleSaveSort">
              保存排序
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="filteredList" row-key="id">
        <ElTableColumn prop="id" label="ID" width="90" align="center" />
        <ElTableColumn prop="name" label="分类名称" min-width="220" />
        <ElTableColumn label="排序" width="160" align="center">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.sort"
              class="w-full"
              :min="0"
              :step="1"
              @change="markSortChanged(row)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="Number(row.status) === 1 ? 'success' : 'info'" effect="plain">
              {{ Number(row.status) === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="addtime" label="添加时间" width="180" />
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <ElButton text type="primary" @click="openEditDialog(row)">编辑</ElButton>
              <ElButton text type="danger" @click="handleDelete(row)">删除</ElButton>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="form.id ? '编辑分类' : '新增分类'" width="520px" destroy-on-close>
      <div class="space-y-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">分类名称</label>
          <ElInput v-model="form.name" maxlength="60" placeholder="请输入分类名称" />
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
          <ElInputNumber v-model="form.sort" class="w-full" :min="0" :step="1" />
        </div>

        <div class="rounded-custom-sm border-full-d p-4">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-medium text-g-900">分类状态</p>
              <p class="mt-1 text-sm text-g-500">禁用后前台不再展示该分类，但商品数据仍保留。</p>
            </div>
            <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    deleteTenantMallCategory,
    fetchTenantMallCategories,
    saveTenantMallCategory,
    updateTenantMallCategorySort,
    type LegacyTenantMallCategory
  } from '@/api/legacy/tenant'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'TenantMallCategoriesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const sortSaving = ref(false)
  const dialogVisible = ref(false)
  const keyword = ref('')

  const list = ref<LegacyTenantMallCategory[]>([])
  const originalSortMap = ref<Record<number, number>>({})

  const form = reactive<Partial<LegacyTenantMallCategory>>({
    id: 0,
    name: '',
    sort: 10,
    status: 1
  })

  const filteredList = computed(() => {
    const value = keyword.value.trim().toLowerCase()
    return [...list.value]
      .filter((item) => !value || item.name?.toLowerCase().includes(value))
      .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0) || Number(a.id || 0) - Number(b.id || 0))
  })

  const sortChangedCount = computed(
    () =>
      list.value.filter((item) => Number(originalSortMap.value[item.id]) !== Number(item.sort || 0)).length
  )

  function resetForm() {
    Object.assign(form, {
      id: 0,
      name: '',
      sort: 10,
      status: 1
    })
  }

  function syncOriginalSortMap(items: LegacyTenantMallCategory[]) {
    originalSortMap.value = items.reduce<Record<number, number>>((result, item) => {
      result[item.id] = Number(item.sort || 0)
      return result
    }, {})
  }

  async function loadData() {
    loading.value = true
    try {
      const result = (await fetchTenantMallCategories()) || []
      list.value = Array.isArray(result) ? result : []
      syncOriginalSortMap(list.value)
    } finally {
      loading.value = false
    }
  }

  function openCreateDialog() {
    resetForm()
    dialogVisible.value = true
  }

  function openEditDialog(row: LegacyTenantMallCategory) {
    Object.assign(form, {
      id: row.id,
      name: row.name,
      sort: row.sort,
      status: row.status
    })
    dialogVisible.value = true
  }

  function markSortChanged(row: LegacyTenantMallCategory) {
    row.sort = Number(row.sort || 0)
  }

  async function handleSave() {
    if (!String(form.name || '').trim()) {
      ElMessage.warning('请先填写分类名称')
      return
    }

    saving.value = true
    try {
      await saveTenantMallCategory({
        ...form,
        name: String(form.name || '').trim(),
        sort: Number(form.sort || 0),
        status: Number(form.status || 0)
      })
      ElMessage.success('分类已保存')
      dialogVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  async function handleSaveSort() {
    if (!sortChangedCount.value) return
    sortSaving.value = true
    try {
      await updateTenantMallCategorySort(
        list.value.map((item) => ({
          id: item.id,
          sort: Number(item.sort || 0)
        }))
      )
      ElMessage.success('排序已保存')
      syncOriginalSortMap(list.value)
    } finally {
      sortSaving.value = false
    }
  }

  async function handleDelete(row: LegacyTenantMallCategory) {
    try {
      await ElMessageBox.confirm(`确定删除分类「${row.name}」吗？`, '删除分类', {
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteTenantMallCategory(row.id)
    ElMessage.success('分类已删除')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>
