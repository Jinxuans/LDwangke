<template>
  <div class="agent-list-page art-full-height">
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
          <ElSpace wrap>
            <ElTag effect="plain">代理列表</ElTag>
            <ElTag effect="plain">当前角色：{{ isAdmin ? '管理员' : '代理' }}</ElTag>
            <ElTag :type="crossRechargeAllowed ? 'success' : 'info'" effect="plain">
              {{ crossRechargeAllowed ? '已开通跨户充值' : '跨户充值关闭' }}
            </ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增代理</ElButton>
            <ElButton v-if="isAdmin || crossRechargeAllowed" plain @click="openCrossRechargeDialog">
              跨户充值
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="tableData"
        :columns="columns"
        :pagination="pagination"
        :pagination-options="{ align: 'right' }"
        :show-table-header="true"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog v-model="createVisible" title="新增代理" width="720px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">账号信息</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">按原后台接口逻辑，先做费用预校验，再确认创建。</p>
          </div>

          <div class="mt-5 grid gap-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">用户昵称</label>
              <ElInput v-model="createForm.nickname" maxlength="32" placeholder="请输入用户昵称" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">用户账号</label>
              <ElInput v-model="createForm.user" maxlength="32" placeholder="请输入 QQ 号或账号" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">登录密码</label>
              <ElInput
                v-model="createForm.pass"
                type="password"
                show-password
                placeholder="请输入登录密码"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">等级方案</label>
              <ElSelect
                v-model="createForm.gradeId"
                class="w-full"
                clearable
                filterable
                placeholder="请选择等级"
              >
                <ElOption
                  v-for="item in gradeOptions"
                  :key="item.id"
                  :label="`${item.name}（费率 ${item.rate} / 开通 ${item.money} 元）`"
                  :value="item.id"
                />
              </ElSelect>
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">创建摘要</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">提交前会再次校验费用和等级信息。</p>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">昵称</span>
              <span class="truncate font-medium text-g-900">{{ createForm.nickname || '未填写昵称' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">账号</span>
              <span class="truncate font-medium text-g-900">{{ createForm.user || '未填写账号' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">等级</span>
              <span class="truncate text-right font-medium text-g-900">{{ selectedCreateGradeText }}</span>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="createVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="createSubmitting" @click="handleCreateAgent">
            确认创建
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="gradeVisible" title="修改等级" width="520px" destroy-on-close>
      <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
        为当前代理切换新的价格等级，接口会先返回费用确认信息。
      </div>

      <ElSelect
        v-model="gradeForm.gradeId"
        class="mt-4 w-full"
        clearable
        filterable
        placeholder="请选择等级"
      >
        <ElOption
          v-for="item in gradeOptions"
          :key="item.id"
          :label="`${item.name}（费率 ${item.rate} / 开通 ${item.money} 元）`"
          :value="item.id"
        />
      </ElSelect>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="gradeVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="gradeSubmitting" @click="handleChangeGrade">
            保存等级
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="superiorVisible" title="调整上级" width="560px" destroy-on-close>
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
          该操作会直接修改代理的上级归属，请确认新的上级 UID 正确无误。
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">目标代理</label>
          <ElInput :model-value="superiorForm.targetName" disabled />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">当前上级 UID</label>
          <ElInputNumber :model-value="superiorForm.currentSuperiorUid" class="w-full" disabled />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">新上级 UID</label>
          <ElInputNumber
            v-model="superiorForm.superiorUid"
            class="w-full"
            :min="1"
            :precision="0"
            placeholder="请输入新的上级 UID"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="superiorVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="superiorSubmitting" @click="handleChangeSuperior">
            确认调整
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="crossVisible" title="跨户充值" width="520px" destroy-on-close>
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
          充值金额将直接进入目标账户，实际扣费按当前账户费率和目标费率换算。
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">目标 UID</label>
          <ElInputNumber
            v-model="crossForm.uid"
            class="w-full"
            :min="1"
            :precision="0"
            placeholder="请输入目标 UID"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">充值金额</label>
          <ElInputNumber
            v-model="crossForm.money"
            class="w-full"
            :min="0.01"
            :precision="2"
            placeholder="请输入充值金额"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="crossVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="crossSubmitting" @click="handleCrossRecharge">
            确认充值
          </ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonMore, {
    type ButtonMoreItem
  } from '@/components/core/forms/art-button-more/index.vue'
  import { ElAvatar, ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { fetchGetUserInfo } from '@/api/auth'
  import {
    adminChangeLegacyAgentSuperior,
    changeLegacyAgentGrade,
    changeLegacyAgentStatus,
    checkLegacyCrossRechargePermission,
    createLegacyAgent,
    deductLegacyAgent,
    fetchLegacyAgentList,
    legacyAdminImpersonate,
    openLegacyAgentKey,
    rechargeLegacyAgent,
    resetLegacyAgentPassword,
    setLegacyAgentInviteCode,
    submitLegacyCrossRecharge,
    type LegacyAgentListItem
  } from '@/api/legacy/agent'
  import { fetchLegacyUserGrades, type LegacyGradeOption } from '@/api/legacy/user-center'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'AgentListPage' })

  const userStore = useUserStore()

  const loading = ref(false)
  const createSubmitting = ref(false)
  const gradeSubmitting = ref(false)
  const superiorSubmitting = ref(false)
  const crossSubmitting = ref(false)

  const tableData = ref<LegacyAgentListItem[]>([])
  const gradeOptions = ref<LegacyGradeOption[]>([])
  const crossRechargeAllowed = ref(false)

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    keywords?: string
    type?: string
  }>({
    type: '2',
    keywords: undefined
  })

  const appliedSearch = reactive({
    keywords: undefined as string | undefined,
    type: '2'
  })

  const createVisible = ref(false)
  const createForm = reactive({
    nickname: '',
    user: '',
    pass: '',
    gradeId: undefined as number | undefined
  })

  const gradeVisible = ref(false)
  const gradeForm = reactive({
    uid: 0,
    gradeId: undefined as number | undefined
  })

  const superiorVisible = ref(false)
  const superiorForm = reactive({
    uid: 0,
    targetName: '',
    currentSuperiorUid: 0,
    superiorUid: undefined as number | undefined
  })

  const crossVisible = ref(false)
  const crossForm = reactive({
    uid: undefined as number | undefined,
    money: undefined as number | undefined
  })

  const roleSet = computed(() => new Set(userStore.getUserInfo.roles || []))
  const isAdmin = computed(
    () => roleSet.value.has('R_ADMIN') || roleSet.value.has('R_SUPER')
  )
  const selectedCreateGradeText = computed(() => {
    const matched = gradeOptions.value.find((item) => item.id === createForm.gradeId)
    if (!matched) {
      return '未选择等级'
    }
    return `${matched.name} / 费率 ${matched.rate} / 开通 ${matched.money} 元`
  })

  const searchItems = computed(() => [
    {
      label: '搜索类型',
      key: 'type',
      type: 'select',
      props: {
        placeholder: '请选择搜索类型',
        options: [
          { label: 'UID', value: '1' },
          { label: '用户名', value: '2' },
          { label: '邀请码', value: '3' },
          { label: '昵称', value: '4' },
          { label: '等级', value: '5' },
          { label: '余额', value: '6' },
          { label: '在线时间', value: '7' }
        ]
      }
    },
    {
      label: '关键词',
      key: 'keywords',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入搜索内容'
      }
    }
  ])

  const getAvatarUrl = (account?: string) =>
    account ? `//q2.qlogo.cn/headimg_dl?dst_uin=${account}&spec=640` : ''
  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)
  const statusTagType = (value: number) => (value === 1 ? 'success' : 'danger')

  const { columns, columnChecks } = useTableColumns<LegacyAgentListItem>(() => {
    const baseColumns: any[] = []

    if (isAdmin.value) {
      baseColumns.push({
        prop: 'uuid',
        label: '上级 UID',
        width: 100,
        align: 'center'
      })
    }

    baseColumns.push(
      {
        prop: 'uid',
        label: '代理信息',
        minWidth: 230,
        formatter: (row: LegacyAgentListItem) =>
          h('div', { class: 'flex items-center gap-3' }, [
            h(ElAvatar, { size: 40, src: getAvatarUrl(row.user) }, () =>
              h('span', row.user?.slice(0, 1) || 'A')
            ),
            h('div', { class: 'leading-6 min-w-0' }, [
              h('p', { class: 'truncate font-semibold text-g-900' }, row.name || row.user || '未命名代理'),
              h('p', { class: 'truncate text-xs text-g-500' }, `UID ${row.uid} / 账号 ${row.user || '-'}`)
            ])
          ])
      },
      {
        prop: 'money',
        label: '余额 / 总充值',
        width: 160,
        formatter: (row: LegacyAgentListItem) =>
          h('div', { class: 'leading-6' }, [
            h('p', { class: 'font-semibold text-[var(--el-color-success)]' }, `¥${formatMoney(row.money)}`),
            h('p', { class: 'text-xs text-g-500' }, `总充值 ¥${formatMoney(row.zcz)}`)
          ])
      },
      {
        prop: 'addprice',
        label: '费率',
        width: 100,
        formatter: (row: LegacyAgentListItem) =>
          h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, Number(row.addprice || 0).toFixed(2))
      },
      {
        prop: 'dd',
        label: '订单量',
        width: 90,
        align: 'center'
      }
    )

    if (isAdmin.value) {
      baseColumns.push({
        prop: 'active',
        label: '状态',
        width: 100,
        formatter: (row: LegacyAgentListItem) =>
          h(
            ElTag,
            { type: statusTagType(Number(row.active || 0)) as any, effect: 'plain' },
            () => (Number(row.active || 0) === 1 ? '正常' : '封禁')
          )
      })
    }

    baseColumns.push(
      {
        prop: 'key',
        label: '密钥',
        width: 110,
        formatter: (row: LegacyAgentListItem) =>
          h(
            ElTag,
            { type: Number(row.key) === 1 ? 'success' : 'warning', effect: 'plain' },
            () => (Number(row.key) === 1 ? '已开通' : '未开通')
          )
      },
      {
        prop: 'yqm',
        label: '邀请码',
        width: 120,
        formatter: (row: LegacyAgentListItem) =>
          h(
            ElTag,
            { type: row.yqm ? 'success' : 'info', effect: 'plain' },
            () => row.yqm || '未设置'
          )
      },
      {
        prop: 'endtime',
        label: '最后在线',
        width: 170
      },
      {
        prop: 'addtime',
        label: '创建时间',
        width: 170
      },
      {
        prop: 'operation',
        label: '操作',
        width: isAdmin.value ? 190 : 140,
        fixed: 'right',
        formatter: (row: LegacyAgentListItem) =>
          h('div', { class: 'flex items-center gap-1.5' }, [
            h(
              ElButton,
              {
                type: 'primary',
                plain: true,
                size: 'small',
                onClick: () => promptRecharge(row.uid)
              },
              () => '充值'
            ),
            ...(isAdmin.value
              ? [
                  h(
                    ElButton,
                    {
                      type: 'warning',
                      plain: true,
                      size: 'small',
                      onClick: () => promptDeduct(row.uid)
                    },
                    () => '扣款'
                  )
                ]
              : []),
            h(ArtButtonMore, {
              list: getAgentMoreActions(row),
              onClick: (item: ButtonMoreItem) => handleAgentMoreAction(row, item)
            })
          ])
      }
    )

    return baseColumns
  })

  const resetCreateForm = () => {
    createForm.nickname = ''
    createForm.user = ''
    createForm.pass = ''
    createForm.gradeId = undefined
  }

  const loadGrades = async () => {
    const result = await fetchLegacyUserGrades()
    gradeOptions.value = Array.isArray(result) ? result : []
  }

  const loadCrossRechargePermission = async () => {
    try {
      const result = await checkLegacyCrossRechargePermission()
      crossRechargeAllowed.value = Boolean(result.allowed)
    } catch {
      crossRechargeAllowed.value = false
    }
  }

  const loadData = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyAgentList({
        page: pagination.current,
        limit: pagination.size,
        type: appliedSearch.type,
        keywords: appliedSearch.keywords
      })
      tableData.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { keywords?: string; type?: string }) => {
    appliedSearch.keywords = params.keywords?.trim() || undefined
    appliedSearch.type = params.type || '2'
    pagination.current = 1
    loadData()
  }

  const handleReset = () => {
    appliedSearch.keywords = undefined
    appliedSearch.type = '2'
    pagination.current = 1
    loadData()
  }

  const handleCurrentChange = (page: number) => {
    pagination.current = page
    loadData()
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    pagination.current = 1
    loadData()
  }

  const getAgentMoreActions = (row: LegacyAgentListItem): ButtonMoreItem[] => {
    const statusActive = Number(row.active || 0) === 1
    const keyOpened = Number(row.key) === 1

    return [
      {
        key: 'grade',
        label: '修改等级',
        icon: 'ri:price-tag-3-line'
      },
      {
        key: 'invite',
        label: row.yqm ? '修改邀请码' : '设置邀请码',
        icon: 'ri:coupon-3-line'
      },
      {
        key: 'key',
        label: keyOpened ? '密钥已开通' : '开通密钥',
        icon: 'ri:key-2-line',
        disabled: keyOpened
      },
      ...(isAdmin.value
        ? [
            {
              key: 'impersonate',
              label: '进入账号',
              icon: 'ri:login-circle-line',
              color: 'var(--el-color-success)'
            },
            {
              key: 'superior',
              label: '调整上级',
              icon: 'ri:node-tree',
              color: 'var(--el-color-warning)'
            },
            {
              key: 'reset',
              label: '重置密码',
              icon: 'ri:lock-password-line'
            }
          ]
        : []),
      {
        key: 'status',
        label: statusActive ? '封禁账号' : '解封账号',
        icon: statusActive ? 'ri:forbid-2-line' : 'ri:checkbox-circle-line',
        color: statusActive ? 'var(--el-color-danger)' : 'var(--el-color-success)'
      }
    ]
  }

  const handleAgentMoreAction = async (row: LegacyAgentListItem, item: ButtonMoreItem) => {
    switch (item.key) {
      case 'grade':
        await openGradeDialog(row.uid)
        break
      case 'invite':
        await promptInviteCode(row.uid)
        break
      case 'key':
        if (Number(row.key) !== 1) {
          await handleOpenKey(row.uid)
        }
        break
      case 'impersonate':
        await handleImpersonate(row.uid)
        break
      case 'superior':
        openSuperiorDialog(row)
        break
      case 'reset':
        await handleResetPassword(row.uid)
        break
      case 'status':
        await handleChangeStatus(row.uid, Number(row.active || 0))
        break
    }
  }

  const openCreateDialog = async () => {
    if (!gradeOptions.value.length) {
      await loadGrades()
    }
    resetCreateForm()
    createVisible.value = true
  }

  const handleCreateAgent = async () => {
    if (!createForm.nickname.trim() || !createForm.user.trim() || !createForm.pass.trim() || !createForm.gradeId) {
      ElMessage.warning('请填写完整的代理信息')
      return
    }

    createSubmitting.value = true
    try {
      const preview = await createLegacyAgent({
        nickname: createForm.nickname.trim(),
        user: createForm.user.trim(),
        pass: createForm.pass.trim(),
        gradeId: createForm.gradeId,
        type: 0
      })

      await ElMessageBox.confirm(
        String(preview.message || preview.msg || '确认创建当前代理账号吗？'),
        '确认新增代理',
        { type: 'warning' }
      )

      const result = await createLegacyAgent({
        nickname: createForm.nickname.trim(),
        user: createForm.user.trim(),
        pass: createForm.pass.trim(),
        gradeId: createForm.gradeId,
        type: 1
      })
      ElMessage.success(String(result.message || result.msg || '代理创建成功'))
      createVisible.value = false
      resetCreateForm()
      pagination.current = 1
      await loadData()
    } finally {
      createSubmitting.value = false
    }
  }

  const openGradeDialog = async (uid: number) => {
    if (!gradeOptions.value.length) {
      await loadGrades()
    }
    gradeForm.uid = uid
    gradeForm.gradeId = undefined
    gradeVisible.value = true
  }

  const handleChangeGrade = async () => {
    if (!gradeForm.gradeId) {
      ElMessage.warning('请选择新的等级')
      return
    }

    gradeSubmitting.value = true
    try {
      const preview = await changeLegacyAgentGrade(gradeForm.uid, gradeForm.gradeId, 0)
      await ElMessageBox.confirm(
        String(preview.message || preview.msg || '确认修改当前代理等级吗？'),
        '确认修改等级',
        { type: 'warning' }
      )
      const result = await changeLegacyAgentGrade(gradeForm.uid, gradeForm.gradeId, 1)
      ElMessage.success(String(result.message || result.msg || '等级修改成功'))
      gradeVisible.value = false
      await loadData()
    } finally {
      gradeSubmitting.value = false
    }
  }

  const promptRecharge = async (uid: number) => {
    const { value } = await ElMessageBox.prompt('请输入充值金额', '账户充值', {
      inputPlaceholder: '例如 100',
      inputValidator: (val) => Number(val) > 0 || '请输入大于 0 的金额'
    })
    await rechargeLegacyAgent(uid, Number(value))
    ElMessage.success('充值成功')
    await loadData()
  }

  const promptDeduct = async (uid: number) => {
    const { value } = await ElMessageBox.prompt('请输入扣款金额', '账户扣款', {
      inputPlaceholder: '例如 100',
      inputValidator: (val) => Number(val) > 0 || '请输入大于 0 的金额'
    })
    await deductLegacyAgent(uid, Number(value))
    ElMessage.success('扣款成功')
    await loadData()
  }

  const handleChangeStatus = async (uid: number, currentStatus: number) => {
    const nextStatus = currentStatus === 1 ? 0 : 1
    await ElMessageBox.confirm(
      `确认${nextStatus === 1 ? '解封' : '封禁'}该代理账号吗？`,
      '状态变更',
      { type: 'warning' }
    )
    await changeLegacyAgentStatus(uid, nextStatus)
    ElMessage.success('状态已更新')
    await loadData()
  }

  const handleResetPassword = async (uid: number) => {
    await ElMessageBox.confirm('确认重置该代理的登录密码吗？', '重置密码', {
      type: 'warning'
    })
    const result = await resetLegacyAgentPassword(uid)
    ElMessage.success(String(result.message || result.msg || '密码已重置'))
    await loadData()
  }

  const handleOpenKey = async (uid: number) => {
    await ElMessageBox.confirm('确认开通该代理的接口密钥吗？', '开通密钥', {
      type: 'warning'
    })
    await openLegacyAgentKey(uid)
    ElMessage.success('接口密钥已开通')
    await loadData()
  }

  const promptInviteCode = async (uid: number) => {
    const { value } = await ElMessageBox.prompt('请输入新的邀请码', '设置邀请码', {
      inputPlaceholder: '最少 4 位',
      inputValidator: (val) => (val?.trim().length || 0) >= 4 || '邀请码最少 4 位'
    })
    await setLegacyAgentInviteCode(uid, value.trim())
    ElMessage.success('邀请码已更新')
    await loadData()
  }

  const openSuperiorDialog = (record: LegacyAgentListItem) => {
    superiorForm.uid = record.uid
    superiorForm.targetName = `${record.name || record.user || '-'}（UID ${record.uid}）`
    superiorForm.currentSuperiorUid = Number(record.uuid || 0)
    superiorForm.superiorUid = Number(record.uuid || 0) || undefined
    superiorVisible.value = true
  }

  const handleChangeSuperior = async () => {
    if (!superiorForm.superiorUid || superiorForm.superiorUid <= 0) {
      ElMessage.warning('请输入有效的上级 UID')
      return
    }

    superiorSubmitting.value = true
    try {
      await adminChangeLegacyAgentSuperior(superiorForm.uid, superiorForm.superiorUid)
      ElMessage.success('上级归属已调整')
      superiorVisible.value = false
      await loadData()
    } finally {
      superiorSubmitting.value = false
    }
  }

  const openCrossRechargeDialog = () => {
    crossForm.uid = undefined
    crossForm.money = undefined
    crossVisible.value = true
  }

  const handleCrossRecharge = async () => {
    if (!crossForm.uid || crossForm.uid <= 0 || !crossForm.money || crossForm.money <= 0) {
      ElMessage.warning('请填写正确的目标 UID 和充值金额')
      return
    }

    crossSubmitting.value = true
    try {
      await submitLegacyCrossRecharge(crossForm.uid, crossForm.money)
      ElMessage.success('跨户充值成功')
      crossVisible.value = false
      await loadData()
    } finally {
      crossSubmitting.value = false
    }
  }

  const handleImpersonate = async (uid: number) => {
    await ElMessageBox.confirm(`确认以 UID ${uid} 的身份进入系统吗？`, '管理员进入代理', {
      type: 'warning'
    })

    const currentToken = userStore.accessToken
    if (currentToken) {
      localStorage.setItem('legacy-admin-backup-token', currentToken)
    }

    const result = await legacyAdminImpersonate(uid)
    if (!result?.accessToken) {
      ElMessage.warning('切换身份失败，未获取到新的访问令牌')
      return
    }

    userStore.setToken(result.accessToken, result.refreshToken)
    userStore.setLoginStatus(true)
    userStore.setUserInfo(await fetchGetUserInfo())
    window.location.href = '/'
  }

  onMounted(async () => {
    await Promise.all([loadGrades(), loadCrossRechargePermission()])
    await loadData()
  })
</script>
