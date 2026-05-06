<template>
  <div class="admin-tenants-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">当前页 {{ list.length }} 个店铺</ElTag>
            <ElTag type="success" effect="plain">启用 {{ enabledCount }}</ElTag>
            <ElTag type="warning" effect="plain">禁用 {{ disabledCount }}</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新建店铺</ElButton>
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

    <ElDialog v-model="dialogVisible" title="新建店铺" width="920px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">基础信息</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">先选主账号，再填写店铺名称。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">绑定用户</label>
              <ElSelect
                v-model="createForm.uid"
                class="w-full"
                clearable
                filterable
                remote
                reserve-keyword
                :loading="userLoading"
                placeholder="输入用户名、账号或 UID 搜索"
                :remote-method="searchUsers"
                @change="handleUserChange"
              >
                <ElOption
                  v-for="item in userOptions"
                  :key="item.uid"
                  :label="formatUserLabel(item)"
                  :value="item.uid"
                />
              </ElSelect>
              <p class="mt-2 text-xs text-g-500">建议输入账号或用户名关键词，接口会返回前 10 个匹配用户。</p>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">店铺名称</label>
              <ElInput
                v-model="createForm.shop_name"
                maxlength="60"
                placeholder="请输入店铺名称"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">创建设置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">确认绑定账号和店铺摘要。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div class="space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">店铺名称</span>
                <span class="font-medium text-g-900">{{ createForm.shop_name.trim() || '未填写店铺名称' }}</span>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">绑定账号</span>
                <span class="font-medium text-g-900">
                  {{ selectedUser?.name || selectedUser?.user || '未选择用户' }}
                </span>
              </div>
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">账户余额</span>
                <span class="font-medium text-g-900">
                  {{ selectedUser ? moneyLabel(selectedUser.balance) : '请先选择账号' }}
                </span>
              </div>
              <div class="flex flex-wrap gap-2 pt-1">
                <ElTag type="primary" effect="plain">
                  {{ selectedUser ? `UID ${selectedUser.uid}` : '未选择用户' }}
                </ElTag>
                <ElTag :type="selectedUser?.status === 1 ? 'success' : 'warning'" effect="plain">
                  {{ selectedUser?.status === 1 ? '用户正常' : selectedUser ? '用户状态异常' : '待选择用户' }}
                </ElTag>
              </div>
            </div>

            <p class="text-sm leading-7 text-g-500">
              创建成功后会生成对应租户记录，域名、Logo 和商城配置继续在租户侧页面维护。
            </p>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleCreate">创建店铺</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    createLegacyAdminTenant,
    fetchLegacyAdminTenants,
    fetchLegacyAdminUsers,
    setLegacyAdminTenantStatus,
    type LegacyAdminTenantItem,
    type LegacyAdminUserItem
  } from '@/api/legacy/admin-tenants'
  import { ElButton, ElMessage, ElSwitch, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminTenantsPage' })

  const loading = ref(false)
  const saving = ref(false)
  const dialogVisible = ref(false)
  const userLoading = ref(false)

  const list = ref<LegacyAdminTenantItem[]>([])
  const userOptions = ref<LegacyAdminUserItem[]>([])
  const selectedUser = ref<LegacyAdminUserItem | null>(null)
  const searchTimer = ref<number | null>(null)

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const createForm = reactive({
    uid: undefined as number | undefined,
    shop_name: ''
  })

  const enabledCount = computed(() => list.value.filter((item) => Number(item.status) === 1).length)
  const disabledCount = computed(() => list.value.filter((item) => Number(item.status) !== 1).length)

  const moneyLabel = (value?: number) => `¥${Number(value || 0).toFixed(2)}`

  const formatUserLabel = (item: LegacyAdminUserItem) =>
    `${item.name || item.user || '未命名用户'}（${item.user || '-'}）UID:${item.uid}`

  const statusTagType = (status: number) => (Number(status) === 1 ? 'success' : 'warning')

  const { columns, columnChecks } = useTableColumns<LegacyAdminTenantItem>(() => [
    {
      prop: 'tid',
      label: 'TID',
      width: 90,
      align: 'center'
    },
    {
      prop: 'uid',
      label: '绑定用户',
      width: 120,
      align: 'center',
      formatter: (row) => `UID ${row.uid}`
    },
    {
      prop: 'shop_name',
      label: '店铺信息',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.shop_name || '未命名店铺'),
          h('p', { class: 'mt-1 text-xs text-g-500 line-clamp-1' }, row.shop_desc || '暂无店铺说明')
        ])
    },
    {
      prop: 'domain',
      label: '域名',
      minWidth: 220,
      formatter: (row) => row.domain || '-'
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
            Number(row.status) === 1 ? '正常' : '禁用'
          )
        ])
    },
    {
      prop: 'addtime',
      label: '创建时间',
      width: 180
    },
    {
      prop: 'operation',
      label: '操作',
      width: 120,
      fixed: 'right',
      formatter: (row) =>
        h(
          ElButton,
          {
            text: true,
            type: Number(row.status) === 1 ? 'danger' : 'primary',
            onClick: () => handleToggleStatus(row, Number(row.status) !== 1)
          },
          () => (Number(row.status) === 1 ? '禁用' : '启用')
        )
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminTenants({
        page: pagination.current,
        limit: pagination.size
      })
      list.value = result.list || []
      pagination.total = Number(result.total || 0)
    } finally {
      loading.value = false
    }
  }

  const searchUsers = (keyword: string) => {
    const value = keyword.trim()
    if (searchTimer.value !== null) {
      window.clearTimeout(searchTimer.value)
      searchTimer.value = null
    }

    if (!value) {
      userOptions.value = []
      return
    }

    searchTimer.value = window.setTimeout(async () => {
      userLoading.value = true
      try {
        const result = await fetchLegacyAdminUsers({
          keywords: value,
          limit: 10
        })
        userOptions.value = result.list || []
      } finally {
        userLoading.value = false
      }
    }, 250)
  }

  const handleUserChange = (uid?: number) => {
    selectedUser.value = userOptions.value.find((item) => item.uid === uid) || null
  }

  const openCreateDialog = () => {
    createForm.uid = undefined
    createForm.shop_name = ''
    userOptions.value = []
    selectedUser.value = null
    dialogVisible.value = true
  }

  const handleCreate = async () => {
    if (!createForm.uid) {
      ElMessage.warning('请先选择绑定用户')
      return
    }
    if (!createForm.shop_name.trim()) {
      ElMessage.warning('请先填写店铺名称')
      return
    }

    saving.value = true
    try {
      await createLegacyAdminTenant({
        uid: createForm.uid,
        shop_name: createForm.shop_name.trim()
      })
      ElMessage.success('店铺已创建')
      dialogVisible.value = false
      await loadData(1)
    } finally {
      saving.value = false
    }
  }

  const handleToggleStatus = async (row: LegacyAdminTenantItem, enabled: boolean) => {
    await setLegacyAdminTenantStatus(row.tid, enabled ? 1 : 0)
    row.status = enabled ? 1 : 0
    ElMessage.success(enabled ? '店铺已启用' : '店铺已禁用')
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

  onUnmounted(() => {
    if (searchTimer.value !== null) {
      window.clearTimeout(searchTimer.value)
    }
  })
</script>
