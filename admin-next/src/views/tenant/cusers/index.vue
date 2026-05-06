<template>
  <div class="tenant-cusers-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">会员管理</ElTag>
            <ElTag effect="plain">会员 {{ pagination.total }} 个</ElTag>
            <ElTag type="success" effect="plain">当前页 {{ members.length }} 条</ElTag>
            <ElButton type="primary" plain @click="openEditDialog()">新增会员</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="members" row-key="id">
        <ElTableColumn prop="account" label="账号" min-width="180" />
        <ElTableColumn prop="nickname" label="昵称" min-width="180" />
        <ElTableColumn prop="addtime" label="注册时间" width="180" />
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <ElButton text type="primary" @click="openEditDialog(row)">编辑</ElButton>
              <ElButton text type="danger" @click="handleDelete(row)">删除</ElButton>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="flex justify-end border-t-d px-5 py-4">
        <ElPagination
          background
          layout="total, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.limit"
          :total="pagination.total"
          @current-change="loadData"
        />
      </div>
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
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    deleteTenantCUser,
    fetchTenantCUsers,
    saveTenantCUser,
    type LegacyTenantCUser
  } from '@/api/legacy/tenant'

  defineOptions({ name: 'TenantCUsersPage' })

  const loading = ref(false)
  const saving = ref(false)
  const editVisible = ref(false)

  const members = ref<LegacyTenantCUser[]>([])

  const pagination = reactive({
    page: 1,
    limit: 20,
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

  async function loadData(page = pagination.page) {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchTenantCUsers({
        page: pagination.page,
        limit: pagination.limit
      })
      members.value = result.list || []
      pagination.total = Number(result.total || result.pagination?.total || 0)
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
      await loadData(editForm.id ? pagination.page : 1)
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
    await loadData(pagination.page)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
