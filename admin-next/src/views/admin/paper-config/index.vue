<template>
  <div class="admin-paper-config-page art-full-height">
    <ElCard class="art-table-card mb-4">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">智文论文配置</ElTag>
            <ElTag type="success" effect="plain">价格项 {{ priceFields.length }} 项</ElTag>
            <ElButton type="primary" plain :loading="saving" @click="handleSave">保存配置</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>
    </ElCard>

    <div v-loading="loading" class="grid gap-4 xl:grid-cols-[0.84fr_1.16fr]">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">接口账号</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">维护智文论文平台登录凭证。</p>
        </div>

        <div class="mt-5 grid gap-4">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">登录账号</label>
            <ElInput v-model="form.lunwen_api_username" placeholder="请输入登录账号" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">登录密码</label>
            <ElInput v-model="form.lunwen_api_password" type="password" show-password placeholder="请输入登录密码" />
          </div>
        </div>

        <p class="mt-4 text-sm leading-6 text-g-500">
          保存的是智文论文平台登录凭证，基础价格会作为后续售价和附加服务计算基准。
        </p>
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">价格配置</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">按论文长度和附加服务拆分价格项。</p>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <article
            v-for="field in priceFields"
            :key="field.key"
            class="rounded-custom-sm border-full-d p-4"
          >
            <label class="mb-2 block text-sm font-medium text-g-800">{{ field.label }}</label>
            <ElInput v-model="form[field.key]" :placeholder="field.placeholder" />
            <p class="mt-2 text-xs text-g-400">{{ field.hint }}</p>
          </article>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import {
    fetchPaperConfig,
    savePaperConfig,
    type PaperConfig
  } from '@/api/legacy/admin-upstream'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'AdminPaperConfigPage' })

  const loading = ref(false)
  const saving = ref(false)

  const createDefaultForm = (): PaperConfig => ({
    lunwen_api_username: '',
    lunwen_api_password: '',
    lunwen_api_6000_price: '30',
    lunwen_api_8000_price: '40',
    lunwen_api_10000_price: '50',
    lunwen_api_12000_price: '60',
    lunwen_api_15000_price: '75',
    lunwen_api_rws_price: '10',
    lunwen_api_ktbg_price: '10',
    lunwen_api_jdaigchj_price: '10',
    lunwen_api_xgdl_price: '3',
    lunwen_api_jcl_price: '3',
    lunwen_api_jdaigcl_price: '3'
  })

  const form = reactive<PaperConfig>(createDefaultForm())

  const priceFields: Array<{
    hint: string
    key: keyof PaperConfig
    label: string
    placeholder: string
  }> = [
    { key: 'lunwen_api_6000_price', label: '论文 6000 字', placeholder: '请输入价格', hint: '基础论文价格，单位元。' },
    { key: 'lunwen_api_8000_price', label: '论文 8000 字', placeholder: '请输入价格', hint: '基础论文价格，单位元。' },
    { key: 'lunwen_api_10000_price', label: '论文 10000 字', placeholder: '请输入价格', hint: '基础论文价格，单位元。' },
    { key: 'lunwen_api_12000_price', label: '论文 12000 字', placeholder: '请输入价格', hint: '基础论文价格，单位元。' },
    { key: 'lunwen_api_15000_price', label: '论文 15000 字', placeholder: '请输入价格', hint: '基础论文价格，单位元。' },
    { key: 'lunwen_api_rws_price', label: '任务书', placeholder: '请输入价格', hint: '任务书固定价格，单位元。' },
    { key: 'lunwen_api_ktbg_price', label: '开题报告', placeholder: '请输入价格', hint: '开题报告固定价格，单位元。' },
    { key: 'lunwen_api_jdaigchj_price', label: '降 AIGC + 查重', placeholder: '请输入价格', hint: '附加服务固定价格，单位元。' },
    { key: 'lunwen_api_xgdl_price', label: '修改段落', placeholder: '请输入价格', hint: '一般按千字计价。' },
    { key: 'lunwen_api_jcl_price', label: '文本降重', placeholder: '请输入价格', hint: '一般按千字计价。' },
    { key: 'lunwen_api_jdaigcl_price', label: '降 AIGC 率', placeholder: '请输入价格', hint: '一般按千字计价。' }
  ]

  async function loadData() {
    loading.value = true
    try {
      Object.assign(form, createDefaultForm(), (await fetchPaperConfig()) || {})
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    saving.value = true
    try {
      await savePaperConfig({ ...form })
      ElMessage.success('智文论文配置已保存')
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
