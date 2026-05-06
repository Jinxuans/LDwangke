<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm overflow-hidden">
      <ElTabs v-model="tab">
        <ElTabPane label="论文下单" name="order">
          <div class="space-y-5 p-5">
            <section class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
              <div class="grid gap-4 xl:grid-cols-[200px_1fr]">
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">字数档位</p>
                  <ElSelect v-model="form.shopcode" class="w-full" placeholder="选择字数档位">
                    <ElOption v-for="item in goods" :key="item.value" :label="item.label" :value="item.value" />
                  </ElSelect>
                </div>
                <div>
                  <div class="mb-2 flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">论文标题</span>
                    <ElButton link type="primary" @click="titleDialog = true">生成标题</ElButton>
                  </div>
                  <ElInput v-model="form.title" placeholder="请输入论文标题" />
                </div>
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">学生姓名</p>
                  <ElInput v-model="form.studentName" placeholder="请输入姓名" />
                </div>
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">专业</p>
                  <ElInput v-model="form.major" placeholder="请输入专业" />
                </div>
              </div>

              <div class="mt-4">
                <p class="mb-2 text-sm font-medium text-g-800">论文要求</p>
                <ElInput v-model="form.requires" type="textarea" :rows="4" placeholder="请输入论文要求" />
              </div>

              <div class="mt-4 grid gap-4 lg:grid-cols-3">
                <div class="rounded-custom-sm border-full-d bg-box px-4 py-3">
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-g-700">任务书</span>
                    <ElSwitch v-model="form.rws" :active-value="1" :inactive-value="0" />
                  </div>
                </div>
                <div class="rounded-custom-sm border-full-d bg-box px-4 py-3">
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-g-700">开题报告</span>
                    <ElSwitch v-model="form.ktbg" :active-value="1" :inactive-value="0" />
                  </div>
                </div>
                <div class="rounded-custom-sm border-full-d bg-box px-4 py-3">
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-g-700">降低 AIGC 痕迹</span>
                    <ElSwitch v-model="form.jiangchong" :active-value="1" :inactive-value="0" />
                  </div>
                </div>
              </div>

              <div class="mt-4 rounded-custom-sm border-full-d bg-box px-4 py-3">
                <div class="grid gap-3 text-sm md:grid-cols-3">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-g-500">当前档位</span>
                    <span class="font-medium text-g-900">{{ selectedGoodsLabel }}</span>
                  </div>
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-g-500">附加服务</span>
                    <span class="font-medium text-g-900">{{ enabledServices }} 项</span>
                  </div>
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-g-500">预估价格</span>
                    <span class="font-semibold text-[var(--el-color-danger)]">¥{{ orderPrice.toFixed(2) }}</span>
                  </div>
                </div>
              </div>

              <div class="mt-4 flex flex-wrap gap-3">
                <ElButton type="primary" :loading="outlineLoading" @click="makeOutline">生成大纲</ElButton>
                <ElButton type="primary" :loading="submitLoading" :disabled="!outline.length" @click="submitOrder">提交订单</ElButton>
              </div>
            </section>

            <section class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
              <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 class="text-base font-semibold text-g-900">论文大纲</h3>
                  <p class="mt-1 text-sm text-g-500">生成后可以直接调整章节和小节，再提交订单。</p>
                </div>
                <ElButton v-if="outline.length" plain @click="addChapter">新增章节</ElButton>
              </div>

              <div v-if="!outline.length" class="flex min-h-[320px] items-center justify-center rounded-custom-sm border-full-d bg-box">
                <ElEmpty description="暂无论文大纲" />
              </div>
              <div v-else class="space-y-4">
                <article v-for="(chapter,ci) in outline" :key="ci" class="rounded-custom-sm border-full-d bg-box p-4">
                  <div class="flex items-center gap-3">
                    <span class="rounded-custom-sm bg-primary px-3 py-1 text-sm font-semibold text-white">第{{ ci + 1 }}章</span>
                    <ElInput v-model="chapter.chapter_title" placeholder="章节标题" />
                    <ElButton type="danger" plain @click="outline.splice(ci,1)">删除</ElButton>
                  </div>
                  <ElInput v-model="chapter.chapter_desc" class="mt-3" type="textarea" :rows="2" placeholder="章节描述" />
                  <div class="mt-4 space-y-3 border-l-2 border-g-300 pl-4">
                    <div v-for="(section,si) in chapter.sections" :key="si" class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                      <div class="flex items-center gap-3">
                        <span class="text-sm font-medium text-g-700">{{ ci + 1 }}.{{ si + 1 }}</span>
                        <ElInput v-model="section.section_title" placeholder="小节标题" />
                        <ElButton type="danger" plain :disabled="chapter.sections.length <= 1" @click="chapter.sections.splice(si,1)">删除</ElButton>
                      </div>
                      <ElInput v-model="section.section_desc" class="mt-3" type="textarea" :rows="2" placeholder="小节描述" />
                    </div>
                    <ElButton plain @click="chapter.sections.push({ section_title: '新小节', section_desc: '', sub_sections: [] })">新增小节</ElButton>
                  </div>
                </article>
              </div>
            </section>
          </div>
        </ElTabPane>

        <ElTabPane label="文本处理" name="tools">
          <div class="grid gap-5 p-5 xl:grid-cols-2">
            <section class="overflow-hidden rounded-custom-sm border-full-d bg-box">
              <div class="space-y-4 p-5">
                <div class="flex items-center justify-between gap-3">
                  <h3 class="text-base font-semibold text-g-900">文本降重 / AIGC</h3>
                  <ElRadioGroup v-model="rewriteMode"><ElRadioButton label="rewrite">降重</ElRadioButton><ElRadioButton label="aigc">降 AIGC</ElRadioButton></ElRadioGroup>
                </div>

                <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
                  <div class="flex flex-wrap items-center justify-between gap-3 text-sm">
                    <span class="text-g-500">处理模式</span>
                    <span class="font-medium text-g-900">{{ rewriteMode === 'rewrite' ? '文本降重' : '降低 AIGC' }}</span>
                  </div>
                </div>

                <ElInput v-model="rewriteText" type="textarea" :rows="8" placeholder="粘贴需要处理的文本" />
                <div class="flex flex-wrap gap-3">
                  <ElButton type="primary" :loading="rewriteLoading" :disabled="!rewriteText" @click="runRewrite">开始处理</ElButton>
                  <ElButton plain :disabled="!rewriteText" @click="rewriteText = ''">清空</ElButton>
                </div>
                <div v-if="rewriteResultVisible" class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <ElTag :type="rewriteStatus === '处理完成' ? 'success' : 'warning'" effect="plain">{{ rewriteStatus }}</ElTag>
                    <ElButton plain size="small" @click="copy(rewriteResult)">复制</ElButton>
                  </div>
                  <div class="min-h-[180px] whitespace-pre-wrap text-sm leading-7 text-g-700">{{ rewriteResult || '等待结果...' }}</div>
                </div>
              </div>
            </section>

            <section class="overflow-hidden rounded-custom-sm border-full-d bg-box">
              <div class="space-y-4 p-5">
                <h3 class="text-base font-semibold text-g-900">文件降重 / 段落修改</h3>
                <ElUpload drag :auto-upload="false" :show-file-list="false" accept=".doc,.docx" :on-change="pickFile">
                  <div class="py-4 text-center">
                    <p class="text-sm font-medium text-g-800">上传论文文件</p>
                    <p class="mt-1 text-xs text-g-500">支持 doc/docx，统计字数后可直接提交处理</p>
                  </div>
                </ElUpload>
                <div v-if="fileName" class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-7 text-g-700">
                  <div class="grid gap-3 md:grid-cols-3">
                    <div class="flex items-center justify-between gap-3">
                      <span class="text-g-500">文件</span>
                      <span class="truncate font-medium text-g-900">{{ fileName }}</span>
                    </div>
                    <div class="flex items-center justify-between gap-3">
                      <span class="text-g-500">大小</span>
                      <span class="font-medium text-g-900">{{ fileSize }}</span>
                    </div>
                    <div class="flex items-center justify-between gap-3">
                      <span class="text-g-500">字数</span>
                      <span class="font-medium text-g-900">{{ fileWords }}</span>
                    </div>
                  </div>
                  <div class="mt-3 flex flex-wrap gap-4">
                    <label class="flex items-center gap-2"><ElCheckbox v-model="fileDedup" :true-value="1" :false-value="0" /> 降重</label>
                    <label class="flex items-center gap-2"><ElCheckbox v-model="fileAigc" :true-value="1" :false-value="0" /> 降 AIGC</label>
                    <ElButton type="primary" :loading="fileLoading" @click="submitFile">提交文件任务</ElButton>
                  </div>
                </div>
                <ElInput v-model="paraText" type="textarea" :rows="6" placeholder="需要段落修改的文本，至少 100 字" />
                <ElInput v-model="paraAdvice" type="textarea" :rows="3" placeholder="修改意见，选填" />
                <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
                  <div class="flex flex-wrap items-center justify-between gap-3 text-sm">
                    <span class="text-g-500">段落修改门槛</span>
                    <span class="font-medium text-g-900">至少 100 字，当前 {{ paraText.length }} 字</span>
                  </div>
                </div>
                <div class="flex flex-wrap gap-3">
                  <ElButton type="primary" :loading="paraLoading" :disabled="paraText.length < 100" @click="runPara">段落修改</ElButton>
                  <ElButton plain :disabled="!paraText && !paraAdvice" @click="resetPara">清空</ElButton>
                </div>
                <div v-if="paraVisible" class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <ElTag :type="paraStatus === '处理完成' ? 'success' : 'warning'" effect="plain">{{ paraStatus }}</ElTag>
                    <ElButton plain size="small" @click="copy(paraResult)">复制</ElButton>
                  </div>
                  <div class="min-h-[140px] whitespace-pre-wrap text-sm leading-7 text-g-700">{{ paraResult || '等待结果...' }}</div>
                </div>
              </div>
            </section>
          </div>
        </ElTabPane>

        <ElTabPane label="论文管理" name="list">
          <div class="space-y-5 p-5">
            <section class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
              <div class="grid gap-4 xl:grid-cols-[1fr_180px_160px_120px_auto]">
                <ElInput v-model="search.title" clearable placeholder="论文标题" />
                <ElInput v-model="search.shopname" clearable placeholder="商品名称" />
                <ElInput v-model="search.studentName" clearable placeholder="学生姓名" />
                <ElSelect v-model="search.state" clearable placeholder="状态">
                  <ElOption label="待处理" value="0" /><ElOption label="正在处理" value="1" /><ElOption label="处理完成" value="2" /><ElOption label="处理异常" value="3" />
                </ElSelect>
                <div class="flex flex-wrap gap-3">
                  <ElButton type="primary" @click="searchList">查询</ElButton>
                  <ElButton plain @click="resetList">重置</ElButton>
                </div>
              </div>
            </section>

            <section class="overflow-hidden rounded-custom-sm border-full-d bg-box">
              <ArtTableHeader :loading="listLoading" layout="refresh" @refresh="loadList">
                <template #left>
                  <ElSpace wrap>
                    <ElTag effect="plain">当前页 {{ rows.length }} 条</ElTag>
                    <ElTag type="success" effect="plain">已完成 {{ listCompletedCount }}</ElTag>
                    <ElTag type="warning" effect="plain">处理中 {{ listPendingCount }}</ElTag>
                  </ElSpace>
                </template>
              </ArtTableHeader>

              <ElTable v-loading="listLoading" :data="rows" size="large">
                <ElTableColumn prop="shopname" label="商品" min-width="120" />
                <ElTableColumn label="论文名称" min-width="220"><template #default="{ row }"><span class="line-clamp-2 font-medium text-g-800">{{ row.title }}</span></template></ElTableColumn>
                <ElTableColumn prop="studentName" label="姓名" width="100" />
                <ElTableColumn prop="major" label="专业" min-width="120" />
                <ElTableColumn label="服务" min-width="180">
                  <template #default="{ row }">
                    <div class="flex flex-wrap gap-2">
                      <ElTag :type="row.jiangchong === 1 ? 'danger' : 'info'" effect="plain">{{ row.jiangchong === 1 ? '需降重' : '不降重' }}</ElTag>
                      <ElTag :type="row.aigc === 1 ? 'danger' : 'info'" effect="plain">{{ row.aigc === 1 ? '需降AIGC' : '不降AIGC' }}</ElTag>
                    </div>
                  </template>
                </ElTableColumn>
                <ElTableColumn label="价格" width="100" align="right"><template #default="{ row }"><span class="font-semibold text-[var(--el-color-danger)]">¥{{ Number(row.price || 0).toFixed(2) }}</span></template></ElTableColumn>
                <ElTableColumn label="状态" width="120" align="center"><template #default="{ row }"><ElTag :type="stateType(row.state)" effect="plain">{{ stateText(row.state) }}</ElTag></template></ElTableColumn>
                <ElTableColumn prop="createTime" label="下单时间" min-width="160" />
                <ElTableColumn label="操作" width="320" fixed="right">
                  <template #default="{ row }">
                    <div class="flex flex-wrap gap-2">
                      <ElButton size="small" @click="downloadFile(row.url, `论文-${row.title}`)">下载论文</ElButton>
                      <ElButton size="small" @click="row.rws ? downloadFile(row.rws, `任务书-${row.title}`) : makeTask(row.id)"> {{ row.rws ? '下载任务书' : '生成任务书' }} </ElButton>
                      <ElButton size="small" @click="row.ktbg ? downloadFile(row.ktbg, `开题-${row.title}`) : makeProposal(row.id)"> {{ row.ktbg ? '下载开题' : '生成开题' }} </ElButton>
                    </div>
                  </template>
                </ElTableColumn>
              </ElTable>

              <div class="flex justify-end border-t-d px-5 py-4">
                <ElPagination background layout="total, prev, pager, next" :current-page="page" :page-size="pageSize" :total="total" @current-change="changePage" />
              </div>
            </section>
          </div>
        </ElTabPane>
      </ElTabs>
    </section>

    <ElDialog v-model="titleDialog" title="生成标题" width="640px">
      <div class="space-y-4">
        <ElInput v-model="titleMajor" placeholder="输入专业方向，例如：计算机科学、教育学" />
        <label v-for="item in titleOptions" :key="item" class="flex cursor-pointer items-start gap-3 rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <ElRadio v-model="selectedTitle" :value="item" />
          <span class="text-sm leading-6 text-g-700">{{ item }}</span>
        </label>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="titleDialog = false">取消</ElButton>
          <ElButton type="primary" :loading="titleLoading" :disabled="!titleMajor" @click="genTitles">生成标题</ElButton>
          <ElButton type="primary" plain :disabled="!selectedTitle" @click="useTitle">使用选中标题</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    countLegacyPaperWords,
    createLegacyPaperOrder,
    downloadLegacyPaperFile,
    fetchLegacyPaperOrders,
    fetchLegacyPaperOutlineStatus,
    fetchLegacyPaperPrices,
    generateLegacyPaperOutline,
    generateLegacyPaperProposal,
    generateLegacyPaperTask,
    generateLegacyPaperTitles,
    streamLegacyPaperParaEdit,
    streamLegacyPaperTextRewrite,
    streamLegacyPaperTextRewriteAigc,
    submitLegacyPaperFileDedup
  } from '@/api/legacy/plugin-paper'

  defineOptions({ name: 'PluginPaperPage' })

  const goods = [
    { label: '论文 6000 字', value: '6000' },
    { label: '论文 8000 字', value: '8000' },
    { label: '论文 10000 字', value: '10000' },
    { label: '论文 12000 字', value: '12000' },
    { label: '论文 15000 字', value: '15000' }
  ]

  const tab = ref('order')
  const prices = ref<any>(null)
  const titleDialog = ref(false)
  const titleLoading = ref(false)
  const titleMajor = ref('')
  const titleOptions = ref<string[]>([])
  const selectedTitle = ref('')
  const outlineLoading = ref(false)
  const submitLoading = ref(false)
  const form = reactive({ shopcode: '', title: '', studentName: '', major: '', requires: '', ktbg: 0, rws: 0, jiangchong: 0 })
  const outline = ref<any[]>([])

  const rewriteMode = ref<'aigc' | 'rewrite'>('rewrite')
  const rewriteText = ref('')
  const rewriteLoading = ref(false)
  const rewriteResult = ref('')
  const rewriteStatus = ref('')
  const rewriteResultVisible = ref(false)

  const pickedFile = ref<File | null>(null)
  const fileName = ref('')
  const fileSize = ref('')
  const fileWords = ref(0)
  const fileDedup = ref(0)
  const fileAigc = ref(0)
  const fileLoading = ref(false)

  const paraText = ref('')
  const paraAdvice = ref('')
  const paraLoading = ref(false)
  const paraResult = ref('')
  const paraStatus = ref('')
  const paraVisible = ref(false)

  const rows = ref<any[]>([])
  const listLoading = ref(false)
  const page = ref(1)
  const pageSize = ref(20)
  const total = ref(0)
  const search = reactive({ title: '', shopname: '', studentName: '', state: '' })
  const selectedGoodsLabel = computed(() => goods.find((item) => item.value === form.shopcode)?.label || '未选择')
  const enabledServices = computed(() => [form.rws, form.ktbg, form.jiangchong].filter((item) => item === 1).length)
  const listCompletedCount = computed(() => rows.value.filter((item) => Number(item.state) === 2).length)
  const listPendingCount = computed(() => rows.value.filter((item) => [0, 1].includes(Number(item.state))).length)

  const orderPrice = computed(() => {
    if (!prices.value || !form.shopcode) return 0
    let totalPrice = Number(prices.value[`price_${form.shopcode}`] || 0)
    if (form.rws) totalPrice += Number(prices.value.price_rws || 0)
    if (form.ktbg) totalPrice += Number(prices.value.price_ktbg || 0)
    if (form.jiangchong) totalPrice += Number(prices.value.price_jdaigchj || 0)
    return totalPrice
  })

  const loadPrices = async () => { try { prices.value = await fetchLegacyPaperPrices() } catch {} }

  const genTitles = async () => {
    titleLoading.value = true
    try {
      const res = await generateLegacyPaperTitles({ direction: titleMajor.value })
      if (res?.code === 200 && res?.data) titleOptions.value = Array.isArray(res.data) ? res.data : [res.data]
      else ElMessage.error(res?.msg || '生成失败')
    } finally { titleLoading.value = false }
  }
  const useTitle = () => { if (selectedTitle.value) { form.title = selectedTitle.value; titleDialog.value = false } }
  const addChapter = () => outline.value.push({ chapter_title: '新章节', chapter_desc: '', sections: [{ section_title: '新小节', section_desc: '', sub_sections: [] }] })

  const pollOutline = async (orderId: string) => {
    for (let i = 0; i < 30; i += 1) {
      await new Promise((r) => setTimeout(r, 2000))
      const res = await fetchLegacyPaperOutlineStatus(orderId)
      if (res?.code === 200 && res?.data) {
        if (Array.isArray(res.data?.chapters)) { outline.value = res.data.chapters; return true }
        if (Array.isArray(res.data)) { outline.value = res.data; return true }
      }
    }
    return false
  }

  const makeOutline = async () => {
    if (!form.shopcode || !form.title) return ElMessage.warning('请先填写商品类型和论文标题')
    outlineLoading.value = true
    try {
      const res = await generateLegacyPaperOutline({ title: form.title, wordCount: Number(form.shopcode), major: form.major, customRequirements: form.requires })
      if (res?.code === 200 && res?.data) {
        if (res.data.orderId) {
          const ok = await pollOutline(res.data.orderId)
          if (!ok) ElMessage.error('大纲生成超时')
        } else if (Array.isArray(res.data)) outline.value = res.data
        else if (Array.isArray(res.data?.chapters)) outline.value = res.data.chapters
      } else ElMessage.error(res?.msg || '生成失败')
    } finally { outlineLoading.value = false }
  }

  const submitOrder = async () => {
    if (!form.shopcode || !form.title || !outline.value.length) return ElMessage.warning('请先完成标题和大纲')
    submitLoading.value = true
    try {
      const res = await createLegacyPaperOrder({ ...form, tigang: JSON.stringify(outline.value) })
      if (res?.code === 200) {
        ElMessage.success(res?.msg || '下单成功')
        Object.assign(form, { shopcode: '', title: '', studentName: '', major: '', requires: '', ktbg: 0, rws: 0, jiangchong: 0 })
        outline.value = []
      } else ElMessage.error(res?.msg || '下单失败')
    } finally { submitLoading.value = false }
  }

  const consumeStream = async (response: Response, onChunk: (t: string) => void, onStatus: (t: string) => void) => {
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const ct = response.headers.get('content-type') || ''
    if (ct.includes('application/json')) { const j = await response.json(); throw new Error(j?.msg || j?.message || '处理失败') }
    const reader = response.body?.getReader()
    if (!reader) throw new Error('无法读取响应流')
    const decoder = new TextDecoder('utf-8')
    let buf = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buf += decoder.decode(value)
      const events = buf.split('\n\n')
      buf = events.pop() || ''
      for (const event of events) {
        const en = event.match(/event:\s*(.+)/)?.[1]?.trim()
        const data = event.match(/data:\s*(.+)/)?.[1]?.trim()
        if (!en || !data) continue
        const parsed = JSON.parse(data)
        if (en === 'chunk') onChunk(String(parsed || ''))
        else if (en === 'status') onStatus(String(parsed || ''))
        else if (en === 'error') throw new Error(typeof parsed === 'string' ? parsed : '处理失败')
      }
    }
  }

  const runRewrite = async () => {
    rewriteLoading.value = true
    rewriteResultVisible.value = true
    rewriteResult.value = ''
    rewriteStatus.value = '正在处理'
    try {
      const res = rewriteMode.value === 'rewrite' ? await streamLegacyPaperTextRewrite(rewriteText.value) : await streamLegacyPaperTextRewriteAigc(rewriteText.value)
      await consumeStream(res, (t) => (rewriteResult.value += t), (s) => (rewriteStatus.value = s))
      rewriteStatus.value = '处理完成'
    } catch (e: any) {
      rewriteStatus.value = '处理失败'
      ElMessage.error(e?.message || '处理失败')
    } finally { rewriteLoading.value = false }
  }

  const pickFile = async (uploadFile: any) => {
    const file = uploadFile?.raw || uploadFile
    if (!file) return
    pickedFile.value = file
    fileName.value = file.name
    fileSize.value = `${(file.size / 1024).toFixed(2)} KB`
    const fd = new FormData()
    fd.append('file', file)
    const res = await countLegacyPaperWords(fd)
    if (res?.code === 200) fileWords.value = Number(res.data || 0)
  }

  const submitFile = async () => {
    if (!pickedFile.value) return ElMessage.warning('请先上传文件')
    if (!fileDedup.value && !fileAigc.value) return ElMessage.warning('请至少选择一项服务')
    fileLoading.value = true
    try {
      const fd = new FormData()
      fd.append('file', pickedFile.value)
      fd.append('wordCount', String(fileWords.value))
      fd.append('jiangchong', String(fileDedup.value))
      fd.append('aigc', String(fileAigc.value))
      const res = await submitLegacyPaperFileDedup(fd)
      if (res?.code === 200) ElMessage.success(res?.msg || '任务已提交')
      else ElMessage.error(res?.msg || '提交失败')
    } finally { fileLoading.value = false }
  }

  const runPara = async () => {
    paraLoading.value = true
    paraVisible.value = true
    paraResult.value = ''
    paraStatus.value = '正在处理'
    try {
      const res = await streamLegacyPaperParaEdit(paraText.value, paraAdvice.value)
      await consumeStream(res, (t) => (paraResult.value += t), (s) => (paraStatus.value = s))
      paraStatus.value = '处理完成'
    } catch (e: any) {
      paraStatus.value = '处理失败'
      ElMessage.error(e?.message || '处理失败')
    } finally { paraLoading.value = false }
  }

  const resetPara = () => { paraText.value = ''; paraAdvice.value = '' }
  const copy = async (text: string) => { if (text) { await navigator.clipboard.writeText(text); ElMessage.success('复制成功') } }

  const loadList = async () => {
    listLoading.value = true
    try {
      const res = await fetchLegacyPaperOrders({ pageNum: page.value, pageSize: pageSize.value, ...search })
      rows.value = Array.isArray(res?.rows) ? res.rows : []
      total.value = Number(res?.total || 0)
    } finally { listLoading.value = false }
  }
  const searchList = () => { page.value = 1; loadList() }
  const resetList = () => { search.title = ''; search.shopname = ''; search.studentName = ''; search.state = ''; page.value = 1; loadList() }
  const changePage = (p: number) => { page.value = p; loadList() }

  const trigger = (url: string) => {
    const a = document.createElement('a')
    a.href = url.replace('http://', `${window.location.protocol}//`)
    a.style.display = 'none'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }
  const downloadFile = async (orderId: string, fileNameText: string) => {
    if (!orderId) return ElMessage.error('文件不存在')
    const res = await downloadLegacyPaperFile(orderId, fileNameText)
    if (res?.code === 200 && res?.msg) trigger(res.msg)
    else ElMessage.error(res?.msg || '下载失败')
  }
  const makeTask = async (id: string) => {
    await ElMessageBox.confirm('确认生成任务书？', '生成任务书', { type: 'warning' })
    const res = await generateLegacyPaperTask(id)
    if (res?.code === 200) { ElMessage.success(res?.msg || '生成成功'); loadList() } else ElMessage.error(res?.msg || '生成失败')
  }
  const makeProposal = async (id: string) => {
    await ElMessageBox.confirm('确认生成开题报告？', '生成开题报告', { type: 'warning' })
    const res = await generateLegacyPaperProposal(id)
    if (res?.code === 200) { ElMessage.success(res?.msg || '生成成功'); loadList() } else ElMessage.error(res?.msg || '生成失败')
  }
  const stateText = (s: number) => ({ 0: '待处理', 1: '正在处理', 2: '处理完成', 3: '处理异常' }[s] || `状态 ${s}`)
  const stateType = (s: number) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info') as 'danger' | 'info' | 'success' | 'warning'

  watch(tab, (v) => { if (v === 'list') loadList() })
  onMounted(loadPrices)
</script>
