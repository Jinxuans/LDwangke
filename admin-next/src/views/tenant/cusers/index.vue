<template>
  <div class="tenant-cusers-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">会员管理</ElTag>
            <ElTag effect="plain">会员 {{ pagination.total }} 个</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ members.length }} 条</ElTag>
            <ElButton type="primary" plain @click="openEditDialog()">新增会员</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="members"
        :columns="columns"
        :pagination="pagination"
        row-key="id"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog v-model="editVisible" :title="editForm.id ? '编辑会员' : '新增会员'" width="620px" destroy-on-close>
      <div class="space-y-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">账号</label>
          <ElInput v-model="editForm.account" maxlength="40" placeholder="请输入登录账号" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">
            {{ editForm.id ? '新密码（留空不修改）' : '登录密码' }}
          </label>
          <ElInput v-model="editForm.password" type="password" show-password placeholder="请输入密码" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">昵称</label>
          <ElInput v-model="editForm.nickname" maxlength="40" placeholder="可选，默认同账号" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElButton, ElMessage, ElMessageBox } from 'element-plus'
  import {
    deleteTenantCUser,
    fetchTenantCUsers,
    saveTenantCUser,
    type LegacyTenantCUser
  } from '@/api/legacy/tenant'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'TenantCUsersPage' })

  const loading = ref(false)
  const saving = ref(false)
  const editVisible = ref(false)

  const members = ref<LegacyTenantCUser[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const editForm = reactive({
    id: 0,
    account: '',
    password: '',
    nickname: ''
  })

  function resetEditForm() {
    editForm.id = 0
    editForm.account = ''
    editForm.password = ''
    editForm.nickname = ''
  }

  const { columns, columnChecks } = useTableColumns<LegacyTenantCUser>(() => [
    { prop: 'account', label: '账号', minWidth: 180 },
    {
      prop: 'nickname',
      label: '昵称',
      minWidth: 180,
      formatter: (row) => row.nickname || '-'
    },
    { prop: 'addtime', label: '注册时间', width: 180 },
    {
      prop: 'operation',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex items-center gap-2' }, [
          h(ElButton, { text: true, type: 'primary', onClick: () => openEditDialog(row) }, () => '编辑'),
          h(ElButton, { text: true, type: 'danger', onClick: () => handleDelete(row) }, () => '删除')
        ])
    }
  ])

  async function loadData(page = pagination.current) {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchTenantCUsers({
        page: pagination.current,
        limit: pagination.size
      })
      members.value = result.list || []
      pagination.total = Number(result.total || result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  function openEditDialog(member?: LegacyTenantCUser) {
    resetEditForm()
    if (member) {
      editForm.id = member.id
      editForm.account = member.account
      editForm.nickname = member.nickname || ''
    }
    editVisible.value = true
  }

  async function handleSave() {
    if (!editForm.account.trim()) {
      ElMessage.warning('请先填写账号')
      return
    }
    if (!editForm.id && !editForm.password.trim()) {
      ElMessage.warning('请先填写密码')
      return
    }

    saving.value = true
    try {
      await saveTenantCUser({
        id: editForm.id || undefined,
        account: editForm.account.trim(),
        password: editForm.password.trim() || undefined,
        nickname: editForm.nickname.trim() || undefined
      })
      ElMessage.success('会员已保存')
      editVisible.value = false
      await loadData(editForm.id ? pagination.current : 1)
    } finally {
      saving.value = false
    }
  }

  async function handleDelete(member: LegacyTenantCUser) {
    try {
      await ElMessageBox.confirm(`确定删除会员「${member.account}」吗？`, '删除会员', {
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteTenantCUser(member.id)
    ElMessage.success('会员已删除')
    await loadData(pagination.current)
  }

  function handleCurrentChange(page: number) {
    loadData(page)
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    loadData(1)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
