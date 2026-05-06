<template>
  <div class="admin-mail-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader :loading="loading || previewing" layout="refresh" @refresh="refreshAll">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">邮件群发</ElTag>
            <ElTag effect="plain">当前范围 {{ targetPreviewLabel }}</ElTag>
            <ElTag type="primary" effect="plain">预估收件 {{ previewCount }} 个</ElTag>
            <ElTag type="success" effect="plain">启用模板 {{ enabledTemplates.length }} 个</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>
    </ElCard>

    <section class="mt-5 grid gap-5 xl:grid-cols-[1.2fr_0.8fr]">
      <ElCard class="art-table-card">
        <template #header>
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="text-base font-semibold text-g-900">发送邮件</p>
              <p class="mt-1 text-sm text-g-500">选择收件范围、模板和正文内容，确认后会生成异步发送任务。</p>
            </div>
            <ElButton plain :loading="previewing" @click="loadPreview">预览收件人</ElButton>
          </div>
        </template>

        <div class="grid gap-4">
          <div class="grid gap-4 md:grid-cols-[180px_1fr]">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">收件范围</label>
              <ElSelect v-model="targetType" class="w-full">
                <ElOption label="全部用户" value="all" />
                <ElOption label="直属用户" value="direct" />
                <ElOption label="非直属用户" value="indirect" />
                <ElOption label="按等级" value="grade" />
                <ElOption label="指定 UID" value="uids" />
              </ElSelect>
            </div>

            <div v-if="targetType === 'grade'">
              <label class="mb-2 block text-sm font-medium text-g-800">等级方案</label>
              <ElSelect v-model="targetGrade" class="w-full" placeholder="请选择等级">
                <ElOption
                  v-for="item in gradeOptions"
                  :key="item.id"
                  :label="`${item.name}（费率 ${item.rate}）`"
                  :value="String(item.id)"
                />
              </ElSelect>
            </div>

            <div v-else-if="targetType === 'uids'">
              <label class="mb-2 block text-sm font-medium text-g-800">目标 UID</label>
              <ElInput
                v-model="targetUids"
                placeholder="多个 UID 用英文逗号分隔，例如 1,2,3"
                @blur="loadPreview"
              />
            </div>

            <div v-else class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-600">
              当前范围将按旧系统规则自动筛选有效邮箱，发送前可点击右上角预览收件人数。
            </div>
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">邮件模板</label>
            <ElSelect
              v-model="selectedTemplateId"
              class="w-full"
              clearable
              filterable
              placeholder="选择模板后自动填充标题和内容"
              @change="handleTemplateChange"
            >
              <ElOption
                v-for="item in enabledTemplates"
                :key="item.id"
                :label="`${item.name}（${item.code}）`"
                :value="item.id"
              />
            </ElSelect>
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">邮件标题</label>
            <ElInput v-model="subject" maxlength="200" placeholder="请输入邮件标题" />
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">邮件正文</label>
            <ElInput
              v-model="content"
              type="textarea"
              :rows="10"
              resize="none"
              placeholder="支持 HTML 内容，发送时会直接进入异步任务队列"
            />
          </div>

          <div class="flex flex-wrap items-center gap-3">
            <ElButton type="primary" :loading="sending" @click="handleSend">发送邮件</ElButton>
            <ElButton plain @click="clearComposer">清空内容</ElButton>
            <span class="text-sm text-g-500">
              当前预计向 {{ previewCount }} 个有效邮箱发信，任务创建后可在下方查看状态。
            </span>
          </div>
        </div>
      </ElCard>

      <div class="grid gap-5">
        <ElCard class="art-table-card">
          <template #header>
            <div>
              <p class="text-base font-semibold text-g-900">发送摘要</p>
              <p class="mt-1 text-sm text-g-500">基于当前填写内容和最近发送记录汇总。</p>
            </div>
          </template>

          <div class="space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">目标范围</span>
              <span class="font-medium text-g-900">{{ targetPreviewLabel }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">邮件标题</span>
              <span class="line-clamp-1 text-right font-medium text-g-900">
                {{ subject || '未填写标题' }}
              </span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">最近记录</span>
              <span class="line-clamp-1 text-right font-medium text-g-900">
                {{ logs[0]?.subject || '暂无发送记录' }}
              </span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag type="primary" effect="plain">{{ previewCount }} 个收件人</ElTag>
              <ElTag v-if="logs[0]" type="info" effect="plain">
                {{ formatTarget(logs[0].target) }}
              </ElTag>
            </div>
          </div>
        </ElCard>

        <ElCard class="art-table-card">
          <template #header>
            <div>
              <p class="text-base font-semibold text-g-900">模板提示</p>
              <p class="mt-1 text-sm text-g-500">优先选择启用模板，避免手填内容与旧系统通知模板不一致。</p>
            </div>
          </template>

          <div class="flex flex-wrap gap-2">
            <ElTag
              v-for="item in enabledTemplates"
              :key="item.id"
              type="info"
              effect="plain"
            >
              {{ item.name }}
            </ElTag>
            <span v-if="!enabledTemplates.length" class="text-sm text-g-500">暂无启用模板</span>
          </div>
        </ElCard>
      </div>
    </section>

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadLogs(pagination.current)">
        <template #left>
          <ElTag effect="plain">发送记录 {{ logs.length }} 条</ElTag>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="logs"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { fetchLegacyUserGrades, type LegacyGradeOption } from '@/api/legacy/user-center'
  import {
    fetchLegacyEmailTemplates,
    fetchLegacyMassEmailLogs,
    previewLegacyEmailRecipients,
    sendLegacyMassEmail,
    type LegacyEmailTemplate,
    type LegacyMassEmailLog
  } from '@/api/legacy/admin-email'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminMailPage' })

  const loading = ref(false)
  const sending = ref(false)
  const previewing = ref(false)

  const logs = ref<LegacyMassEmailLog[]>([])
  const templates = ref<LegacyEmailTemplate[]>([])
  const gradeOptions = ref<LegacyGradeOption[]>([])

  const targetType = ref<'all' | 'direct' | 'indirect' | 'grade' | 'uids'>('all')
  const targetGrade = ref('')
  const targetUids = ref('')
  const selectedTemplateId = ref<number | undefined>()
  const subject = ref('')
  const content = ref('')
  const previewCount = ref(0)

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const enabledTemplates = computed(() => templates.value.filter((item) => Number(item.status) === 1))
  const targetValue = computed(() => {
    if (targetType.value === 'all') {
      return 'all'
    }
    if (targetType.value === 'direct') {
      return 'direct'
    }
    if (targetType.value === 'indirect') {
      return 'indirect'
    }
    if (targetType.value === 'grade') {
      return targetGrade.value ? `grade:${targetGrade.value}` : ''
    }
    return targetUids.value.trim() ? `uids:${targetUids.value.trim()}` : ''
  })
  const targetPreviewLabel = computed(() => formatTarget(targetValue.value))

  const statusMap: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'danger' }> = {
    sending: { label: '发送中', type: 'warning' },
    done: { label: '已完成', type: 'success' },
    partial: { label: '部分失败', type: 'warning' },
    failed: { label: '发送失败', type: 'danger' }
  }

  const formatTarget = (value: string) => {
    if (!value) {
      return '未选择范围'
    }
    if (value === 'all') {
      return '全部用户'
    }
    if (value === 'direct') {
      return '直属用户'
    }
    if (value === 'indirect') {
      return '非直属用户'
    }
    if (value.startsWith('grade:')) {
      const gradeId = value.slice(6)
      const grade = gradeOptions.value.find((item) => String(item.id) === gradeId)
      return grade ? `等级 ${grade.name}` : `等级 ${gradeId}`
    }
    if (value.startsWith('uids:')) {
      return `指定 UID (${value.slice(5)})`
    }
    return value
  }

  const getStatusMeta = (status: string) =>
    statusMap[status] || {
      label: status || '未知状态',
      type: 'info' as const
    }

  const { columns, columnChecks } = useTableColumns<LegacyMassEmailLog>(() => [
    {
      type: 'index',
      label: '序号',
      width: 70
    },
    {
      prop: 'status',
      label: '状态',
      width: 110,
      formatter: (row) =>
        h(ElTag, { type: getStatusMeta(row.status).type }, () => getStatusMeta(row.status).label)
    },
    {
      prop: 'subject',
      label: '邮件标题',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.subject || '无标题'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, formatTarget(row.target))
        ])
    },
    {
      prop: 'total',
      label: '发送结果',
      width: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, `总计 ${row.total || 0}`),
          h(
            'p',
            { class: 'text-xs text-g-500 mt-1' },
            `成功 ${row.success_count || 0} / 失败 ${row.fail_count || 0}`
          )
        ])
    },
    {
      prop: 'addtime',
      label: '创建时间',
      width: 180,
      formatter: (row) => row.addtime || '-'
    }
  ])

  const clearComposer = () => {
    selectedTemplateId.value = undefined
    subject.value = ''
    content.value = ''
  }

  const loadTemplates = async () => {
    const result = await fetchLegacyEmailTemplates()
    templates.value = Array.isArray(result) ? result : []
  }

  const loadGradeOptions = async () => {
    const result = await fetchLegacyUserGrades()
    gradeOptions.value = Array.isArray(result) ? result : []
    if (!targetGrade.value && gradeOptions.value.length) {
      targetGrade.value = String(gradeOptions.value[0].id)
    }
  }

  const loadLogs = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyMassEmailLogs({
        page: pagination.current,
        limit: pagination.size
      })
      logs.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const loadPreview = async () => {
    if (!targetValue.value) {
      previewCount.value = 0
      return
    }
    previewing.value = true
    try {
      const result = await previewLegacyEmailRecipients(targetValue.value)
      previewCount.value = Number(result.count || 0)
    } catch {
      previewCount.value = 0
    } finally {
      previewing.value = false
    }
  }

  const refreshAll = async () => {
    await Promise.all([loadLogs(pagination.current), loadTemplates(), loadPreview()])
  }

  const handleTemplateChange = (templateId?: number) => {
    const matched = enabledTemplates.value.find((item) => item.id === Number(templateId))
    if (!matched) {
      return
    }
    subject.value = matched.subject || ''
    content.value = matched.content || ''
  }

  const handleSend = async () => {
    if (!targetValue.value) {
      ElMessage.warning('请先选择有效的收件范围')
      return
    }
    if (!subject.value.trim() || !content.value.trim()) {
      ElMessage.warning('请填写邮件标题和正文')
      return
    }
    if (previewCount.value <= 0) {
      ElMessage.warning('当前范围没有可发送的有效邮箱')
      return
    }

    await ElMessageBox.confirm(
      `将向 ${previewCount.value} 个邮箱创建发送任务，确认继续吗？`,
      '确认发送邮件',
      { type: 'warning' }
    )

    sending.value = true
    try {
      const result = await sendLegacyMassEmail({
        target: targetValue.value,
        subject: subject.value.trim(),
        content: content.value
      })
      ElMessage.success(result.message || '发送任务已创建')
      clearComposer()
      await Promise.all([loadLogs(1), loadPreview()])
    } finally {
      sending.value = false
    }
  }

  const handleCurrentChange = (page: number) => {
    loadLogs(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadLogs(1)
  }

  watch(
    [targetType, targetGrade],
    () => {
      if (targetType.value !== 'uids') {
        loadPreview()
      }
    },
    { immediate: false }
  )

  onMounted(async () => {
    await Promise.all([loadGradeOptions(), loadTemplates()])
    await Promise.all([loadLogs(1), loadPreview()])
  })
</script>
