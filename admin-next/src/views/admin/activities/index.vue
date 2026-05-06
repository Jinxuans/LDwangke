<template>
  <div class="admin-activities-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">活动总数 {{ pagination.total }}</ElTag>
            <ElTag type="primary" effect="plain">邀新活动 {{ inviteCount }}</ElTag>
            <ElTag type="success" effect="plain">订单活动 {{ orderCount }}</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增活动</ElButton>
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

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑活动' : '新增活动'"
      width="780px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.08fr_0.92fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">活动内容</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">名称、要求和时间区间会直接在前台活动中心展示。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">活动名称</label>
              <ElInput v-model="editForm.name" maxlength="80" placeholder="请输入活动名称" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">活动要求</label>
              <ElInput
                v-model="editForm.yaoqiu"
                type="textarea"
                :rows="4"
                resize="none"
                placeholder="例如：邀请 10 个用户注册并完成首单"
              />
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">要求数量</label>
                <ElInput v-model="editForm.num" placeholder="例如 10" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">奖励金额（元）</label>
                <ElInput v-model="editForm.money" placeholder="例如 50" />
              </div>
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">开始时间</label>
                <ElInput v-model="editForm.addtime" placeholder="2026-01-01 00:00:00" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">结束时间</label>
                <ElInput v-model="editForm.endtime" placeholder="2026-12-31 23:59:59" />
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">活动设置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">统一控制类型、状态和奖励摘要。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">活动类型</label>
              <ElSegmented
                v-model="editForm.type"
                :options="[
                  { label: '邀新活动', value: '1' },
                  { label: '订单活动', value: '2' }
                ]"
                class="w-full"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">活动状态</label>
              <ElSegmented
                v-model="editForm.status_ok"
                :options="[
                  { label: '进行中', value: '1' },
                  { label: '已结束', value: '2' }
                ]"
                class="w-full"
              />
            </div>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">活动名称</span>
              <span class="truncate font-medium text-g-900">{{ editForm.name || '未命名活动' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">奖励金额</span>
              <span class="font-medium text-[var(--el-color-warning)]">¥{{ formatMoney(editForm.money) }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">要求数量</span>
              <span class="font-medium text-g-900">{{ editForm.num || '未填写' }}</span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="editForm.type === '1' ? 'primary' : 'success'" effect="plain">
                {{ editForm.type === '1' ? '邀新活动' : '订单活动' }}
              </ElTag>
              <ElTag :type="editForm.status_ok === '1' ? 'success' : 'info'" effect="plain">
                {{ editForm.status_ok === '1' ? '进行中' : '已结束' }}
              </ElTag>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存活动</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteLegacyAdminActivity,
    fetchLegacyAdminActivities,
    saveLegacyAdminActivity,
    type LegacyAdminActivity
  } from '@/api/legacy/admin-auxiliary'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminActivitiesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const dialogVisible = ref(false)
  const list = ref<LegacyAdminActivity[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const editForm = reactive({
    hid: 0,
    name: '',
    yaoqiu: '',
    type: '1',
    num: '',
    money: '',
    addtime: '',
    endtime: '',
    status_ok: '1'
  })

  const isEditing = computed(() => editForm.hid > 0)
  const inviteCount = computed(() => list.value.filter((item) => item.type === '1').length)
  const orderCount = computed(() => list.value.filter((item) => item.type !== '1').length)

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const { columns, columnChecks } = useTableColumns<LegacyAdminActivity>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'name',
      label: '活动名称',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.name || '未命名活动'),
          h('p', { class: 'text-xs text-g-500 line-clamp-2 mt-1' }, row.yaoqiu || '暂无要求描述')
        ])
    },
    {
      prop: 'type',
      label: '类型',
      width: 110,
      formatter: (row) =>
        h(ElTag, { type: row.type === '1' ? 'primary' : 'success' }, () =>
          row.type === '1' ? '邀新活动' : '订单活动'
        )
    },
    {
      prop: 'num',
      label: '要求数量',
      width: 100,
      align: 'center'
    },
    {
      prop: 'money',
      label: '奖励金额',
      width: 110,
      align: 'right',
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-warning)]' }, `¥${formatMoney(row.money)}`)
    },
    {
      prop: 'addtime',
      label: '开始时间',
      width: 170
    },
    {
      prop: 'endtime',
      label: '结束时间',
      width: 170
    },
    {
      prop: 'status_ok',
      label: '状态',
      width: 100,
      formatter: (row) =>
        h(ElTag, { type: row.status_ok === '1' ? 'success' : 'info' }, () =>
          row.status_ok === '1' ? '进行中' : '已结束'
        )
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
    editForm.hid = 0
    editForm.name = ''
    editForm.yaoqiu = ''
    editForm.type = '1'
    editForm.num = ''
    editForm.money = ''
    editForm.addtime = ''
    editForm.endtime = ''
    editForm.status_ok = '1'
  }

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminActivities({
        page: pagination.current,
        limit: pagination.size
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyAdminActivity) => {
    editForm.hid = record.hid
    editForm.name = record.name || ''
    editForm.yaoqiu = record.yaoqiu || ''
    editForm.type = record.type || '1'
    editForm.num = record.num || ''
    editForm.money = record.money || ''
    editForm.addtime = record.addtime || ''
    editForm.endtime = record.endtime || ''
    editForm.status_ok = record.status_ok || '1'
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.name.trim()) {
      ElMessage.warning('请先填写活动名称')
      return
    }
    if (!editForm.num.trim() || !editForm.money.trim()) {
      ElMessage.warning('请先填写要求数量和奖励金额')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminActivity({ ...editForm })
      ElMessage.success(editForm.hid ? '活动已更新' : '活动已创建')
      dialogVisible.value = false
      await loadData(pagination.current)
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (record: LegacyAdminActivity) => {
    await ElMessageBox.confirm(`确定删除活动「${record.name || record.hid}」吗？`, '删除活动', {
      type: 'warning'
    })
    await deleteLegacyAdminActivity(record.hid)
    ElMessage.success('活动已删除')
    await loadData(pagination.current)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
