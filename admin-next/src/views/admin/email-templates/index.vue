<template>
  <div class="admin-email-templates-page flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm overflow-hidden">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">模板总数 {{ templates.length }}</ElTag>
            <ElTag type="success" effect="plain">启用模板 {{ enabledCount }}</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>
    </section>

    <section v-if="templates.length" class="grid gap-5 xl:grid-cols-2">
      <article
        v-for="item in templates"
        :key="item.id"
        class="art-card-sm p-5"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="flex flex-wrap items-center gap-2">
              <ElTag :type="codeMeta(item.code).type" effect="plain">{{ codeMeta(item.code).label }}</ElTag>
              <ElTag :type="item.status === 1 ? 'success' : 'info'">{{ item.status === 1 ? '启用' : '禁用' }}</ElTag>
            </div>
            <h2 class="mt-3 text-lg font-semibold text-g-900">{{ item.name }}</h2>
            <p class="mt-1 text-sm text-g-500">{{ codeMeta(item.code).desc }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <ElButton text type="primary" @click="handlePreview(item.code)">预览</ElButton>
            <ElButton text type="primary" @click="openTestDialog(item.code)">测试</ElButton>
            <ElButton type="primary" plain @click="openEditDialog(item)">编辑</ElButton>
          </div>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">邮件标题</p>
          <p class="mt-2 text-sm font-semibold text-g-900">{{ item.subject || '未填写标题' }}</p>
        </div>

        <div class="mt-4 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">可用变量</p>
          <div class="mt-3 flex flex-wrap gap-2">
            <ElTag
              v-for="variable in formatVariables(item.variables)"
              :key="`${item.id}-${variable}`"
              type="info"
              effect="plain"
            >
              {{ '{' + variable + '}' }}
            </ElTag>
            <span v-if="!formatVariables(item.variables).length" class="text-sm text-g-500">无变量</span>
          </div>
        </div>

        <p class="mt-4 text-xs text-g-500">最后更新：{{ item.updated_at || item.created_at || '-' }}</p>
      </article>
    </section>

    <section v-else class="art-card-sm px-6 py-16">
      <ElEmpty description="暂无邮件模板数据" />
    </section>

    <ElDialog v-model="editVisible" title="编辑邮件模板" width="860px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.08fr_0.92fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">模板内容</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">标题和正文都支持模板变量，内容支持 HTML。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">邮件标题</label>
              <ElInput v-model="editForm.subject" maxlength="200" placeholder="{site_name} - 注册验证码" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">邮件正文</label>
              <ElInput
                v-model="editForm.content"
                type="textarea"
                :rows="14"
                resize="none"
                placeholder="支持 HTML 内容和模板变量"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">模板设置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">确认模板类型、变量和启用状态。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div class="space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="text-g-500">模板名称</span>
                <span class="font-medium text-g-900">{{ editMeta.label }}</span>
              </div>
              <p class="text-sm text-g-500">{{ editMeta.desc }}</p>
            </div>
            <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">可用变量</p>
              <div class="mt-3 flex flex-wrap gap-2">
                <ElTag
                  v-for="variable in formatVariables(editCurrentVariables)"
                  :key="`edit-${variable}`"
                  type="info"
                  effect="plain"
                >
                  {{ '{' + variable + '}' }}
                </ElTag>
                <span v-if="!formatVariables(editCurrentVariables).length" class="text-sm text-g-500">
                  无变量
                </span>
              </div>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">模板状态</label>
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
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存模板</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="previewVisible" title="模板预览" width="760px" destroy-on-close>
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">邮件标题</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ previewSubject || '无标题' }}</p>
        </div>
        <div class="rounded-custom-sm border-full-d bg-box p-4">
          <div v-html="previewHtml"></div>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="testVisible" title="测试发送" width="460px" destroy-on-close>
      <ElInput v-model="testTo" placeholder="输入测试收件邮箱地址" />
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="testVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="testLoading" @click="handleTest">发送测试邮件</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import {
    fetchLegacyEmailTemplates,
    previewLegacyEmailTemplate,
    saveLegacyEmailTemplate,
    testLegacyEmailTemplate,
    type LegacyEmailTemplate
  } from '@/api/legacy/admin-email'

  defineOptions({ name: 'AdminEmailTemplatesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const testLoading = ref(false)
  const templates = ref<LegacyEmailTemplate[]>([])
  const editVisible = ref(false)
  const previewVisible = ref(false)
  const testVisible = ref(false)
  const previewHtml = ref('')
  const previewSubject = ref('')
  const testCode = ref('')
  const testTo = ref('')
  const editCurrentCode = ref('')
  const editCurrentVariables = ref('')

  const editForm = reactive({
    id: 0,
    subject: '',
    content: '',
    status: 1
  })

  const templateCodeMap: Record<string, { desc: string; label: string; type: 'info' | 'primary' | 'success' | 'warning' }> = {
    register: { label: '注册验证码', type: 'primary', desc: '用户注册时发送验证码邮件。' },
    reset_password: { label: '重置密码', type: 'warning', desc: '用户找回密码时发送验证码邮件。' },
    system_notify: { label: '系统通知', type: 'success', desc: '后台群发或系统通知邮件。' }
  }

  const enabledCount = computed(() => templates.value.filter((item) => item.status === 1).length)
  const editMeta = computed(() => codeMeta(editCurrentCode.value))

  const codeMeta = (code: string) => {
    return templateCodeMap[code] || {
      label: code || '未知模板',
      type: 'info' as const,
      desc: '自定义模板'
    }
  }

  const formatVariables = (variables?: string) =>
    String(variables || '')
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)

  const loadData = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyEmailTemplates()
      templates.value = Array.isArray(result) ? result : []
    } finally {
      loading.value = false
    }
  }

  const openEditDialog = (item: LegacyEmailTemplate) => {
    editForm.id = item.id
    editForm.subject = item.subject || ''
    editForm.content = item.content || ''
    editForm.status = item.status
    editCurrentCode.value = item.code
    editCurrentVariables.value = item.variables || ''
    editVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.subject.trim() || !editForm.content.trim()) {
      ElMessage.warning('邮件标题和正文不能为空')
      return
    }
    saving.value = true
    try {
      await saveLegacyEmailTemplate({
        id: editForm.id,
        subject: editForm.subject.trim(),
        content: editForm.content,
        status: editForm.status
      })
      ElMessage.success('模板已保存')
      editVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  const handlePreview = async (code: string) => {
    const result = await previewLegacyEmailTemplate(code)
    previewSubject.value = result.subject || ''
    previewHtml.value = result.html || ''
    previewVisible.value = true
  }

  const openTestDialog = (code: string) => {
    testCode.value = code
    testTo.value = ''
    testVisible.value = true
  }

  const handleTest = async () => {
    if (!testTo.value.trim()) {
      ElMessage.warning('请输入测试收件邮箱')
      return
    }
    testLoading.value = true
    try {
      await testLegacyEmailTemplate(testCode.value, testTo.value.trim())
      ElMessage.success('测试邮件已发送')
      testVisible.value = false
    } finally {
      testLoading.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
