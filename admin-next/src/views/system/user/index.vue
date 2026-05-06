<template>
  <div class="system-user-page art-full-height">
    <ElCard class="art-card-sm mb-4">
      <div class="grid gap-4 lg:grid-cols-[1fr_auto]">
        <div class="grid gap-4 md:grid-cols-[1fr_auto_auto]">
          <ElInput
            v-model="keywords"
            clearable
            placeholder="搜索用户名 / UID"
            @keyup.enter="loadUsers(1)"
          />
          <ElButton type="primary" @click="loadUsers(1)">搜索</ElButton>
          <ElButton @click="resetFilters">重置</ElButton>
        </div>
        <div class="flex items-center justify-end text-sm text-g-500">支持按账号、用户名和 UID 快速筛选。</div>
      </div>
    </ElCard>

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadUsers(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">用户管理</ElTag>
            <ElTag effect="plain">共 {{ pagination.total }} 个用户</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ userList.length }} 条</ElTag>
            <ElTag type="info" effect="plain">等级 {{ gradeOptions.length }} 个</ElTag>
            <ElTag effect="plain">分页 {{ pagination.current }}/{{ totalPage }}</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :data="userList"
        :columns="columns"
        :loading="loading"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      >
        <template #actions="{ row }">
          <div class="flex flex-wrap gap-2">
            <ElButton link type="primary" @click="openEdit(row)">编辑</ElButton>
            <ElButton link type="danger" @click="handleResetPass(row.uid)">重置密码</ElButton>
            <ElButton v-if="loginAsEnabled" link @click="handleLoginAs(row.uid)">登入</ElButton>
          </div>
        </template>
      </ArtTable>
    </ElCard>

    <ElDialog v-model="editVisible" title="编辑用户" width="520px">
      <div v-if="editUser" class="grid gap-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">UID</label>
          <ElInput :model-value="String(editUser.uid)" disabled />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">账号</label>
          <ElInput :model-value="editUser.user" disabled />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">余额</label>
          <ElInputNumber v-model="editBalance" class="w-full" :min="0" :step="10" :precision="2" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">等级</label>
          <ElSelect v-model="editGradeId" class="w-full" filterable clearable placeholder="请选择等级">
            <ElOption
              v-for="item in gradeOptions"
              :key="item.id"
              :label="`${item.name}（费率 ${item.rate}）`"
              :value="item.id"
            />
          </ElSelect>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSaveEdit">保存</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useUserStore } from '@/store/modules/user'
  import {
    fetchLegacyAdminUsers,
    impersonateLegacyAdminUser,
    resetLegacyAdminUserPassword,
    updateLegacyAdminUserBalance,
    updateLegacyAdminUserGrade,
    type LegacyAdminUserItem
  } from '@/api/legacy/admin-user-tools'
  import { fetchLegacyAdminGrades, type LegacyAdminGrade } from '@/api/legacy/admin-grades'
  import { fetchLegacySiteConfig } from '@/api/legacy/site'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'User' })

  const userStore = useUserStore()

  const loading = ref(false)
  const saving = ref(false)
  const loginAsEnabled = ref(false)
  const userList = ref<LegacyAdminUserItem[]>([])
  const keywords = ref('')
  const gradeOptions = ref<LegacyAdminGrade[]>([])
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const editVisible = ref(false)
  const editUser = ref<LegacyAdminUserItem | null>(null)
  const editBalance = ref(0)
  const editGradeId = ref<number | undefined>()

  const totalPage = computed(() => {
    if (!pagination.total) return 1
    return Math.max(1, Math.ceil(pagination.total / pagination.size))
  })

  async function loadGrades() {
    gradeOptions.value = (await fetchLegacyAdminGrades()) || []
  }

  async function loadSiteConfig() {
    const config = await fetchLegacySiteConfig()
    loginAsEnabled.value = config?.onlineStore_trdltz === '1'
  }

  async function loadUsers(page = 1) {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminUsers({
        page: pagination.current,
        limit: pagination.size,
        keywords: keywords.value.trim()
      })
      userList.value = result?.list || []
      pagination.total = Number(result?.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  function resetFilters() {
    keywords.value = ''
    loadUsers(1)
  }

  function openEdit(user: LegacyAdminUserItem) {
    editUser.value = { ...user }
    editBalance.value = Number(user.balance || 0)
    editGradeId.value = user.grade_id || undefined
    editVisible.value = true
  }

  async function handleResetPass(uid: number) {
    const result = await ElMessageBox.prompt('请输入新密码', '重置密码', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      inputPlaceholder: '密码不能少于 6 位',
      inputValidator: (value) => (value && value.length >= 6 ? true : '密码不能少于 6 位')
    })
    await resetLegacyAdminUserPassword(uid, result.value)
    ElMessage.success('密码已重置')
  }

  async function handleSaveEdit() {
    if (!editUser.value) return
    if (!editGradeId.value) {
      ElMessage.warning('请选择等级')
      return
    }
    saving.value = true
    try {
      await Promise.all([
        updateLegacyAdminUserBalance(editUser.value.uid, Number(editBalance.value || 0)),
        updateLegacyAdminUserGrade(editUser.value.uid, Number(editGradeId.value))
      ])
      editVisible.value = false
      ElMessage.success('保存成功')
      await loadUsers(pagination.current)
    } finally {
      saving.value = false
    }
  }

  async function handleLoginAs(uid: number) {
    const result = await impersonateLegacyAdminUser(uid)
    if (!result?.accessToken) {
      ElMessage.warning('代登入失败')
      return
    }
    userStore.setToken(result.accessToken, result.refreshToken)
    userStore.setLoginStatus(true)
    window.location.href = '/'
  }

  function handleCurrentChange(page: number) {
    loadUsers(page)
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    loadUsers(1)
  }

  const { columns, columnChecks } = useTableColumns<LegacyAdminUserItem>(() => [
    { prop: 'uid', label: 'UID', width: 80 },
    { prop: 'user', label: '账号', minWidth: 150 },
    { prop: 'name', label: '昵称', minWidth: 140 },
    {
      prop: 'grade_name',
      label: '等级',
      width: 140,
      formatter: (row) => h(ElTag, { type: 'primary', effect: 'plain' }, () => row.grade_name || '未知')
    },
    { prop: 'addprice', label: '费率', width: 100 },
    {
      prop: 'balance',
      label: '余额',
      width: 120,
      formatter: (row) => h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, `¥${Number(row.balance || 0).toFixed(2)}`)
    },
    { prop: 'phone', label: '手机号', width: 150 },
    { prop: 'addtime', label: '注册时间', width: 180 },
    { prop: 'actions', label: '操作', minWidth: 190, fixed: 'right', useSlot: true }
  ])

  onMounted(async () => {
    await Promise.all([loadGrades(), loadSiteConfig()])
    await loadUsers(1)
  })
</script>
