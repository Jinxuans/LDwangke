<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-g-900">批量下单</h2>
          <p class="mt-1 text-sm text-g-500">批量粘贴账号信息，解析后统一提交，适合桌面端集中处理。</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <ElTag effect="plain">{{ autoNormalize ? '自动整理' : '保留原样' }}</ElTag>
          <ElTag :type="selectedClass ? 'success' : 'info'" effect="plain">
            {{ selectedClass ? '已选课程' : '未选课程' }}
          </ElTag>
          <ElSwitch v-model="autoNormalize" inline-prompt active-text="整理" inactive-text="原样" />
        </div>
      </div>

      <div class="mt-5 grid gap-5 xl:grid-cols-[minmax(0,1.15fr)_320px]">
        <div>
          <ElInput
            v-model="rawText"
            type="textarea"
            :rows="10"
            resize="none"
            placeholder="例如：&#10;家里蹲大学 13800138000 123456&#10;清华大学 13900139000 654321&#10;&#10;或：&#10;13800138000 123456"
            @blur="handleNormalizeOnBlur"
          />

          <div class="mt-4 rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-600">
            解析前会按空格拆分。开启“整理”后，会自动替换中英文逗号、制表符和多余空格。
          </div>

          <div class="mt-5 flex flex-wrap gap-3">
            <ElButton type="primary" :disabled="!selectedClass" @click="handleParse">解析数据</ElButton>
            <ElButton plain @click="clearAll">清空</ElButton>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm font-medium text-g-800">当前课程</p>
              <ElButton text type="primary" @click="refreshCourses">刷新课程</ElButton>
            </div>
            <p class="mt-3 text-base font-semibold text-g-900">{{ selectedClass?.name || '请选择一门课程' }}</p>
            <p class="mt-2 text-sm text-g-500">
              {{ selectedClass ? `价格 ¥${formatLegacyPrice(selectedClass.price)}${selectedClass.noun ? ` · ${selectedClass.noun}` : ''}` : '课程选中后会在这里显示价格和预计费用。' }}
            </p>
          </div>

          <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-sm font-medium text-g-800">解析结果</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <ElTag effect="plain">有效 {{ validCount }}</ElTag>
              <ElTag effect="plain" type="danger">无效 {{ invalidCount }}</ElTag>
              <ElTag v-if="selectedClass && validCount" type="success" effect="plain">预估 ¥{{ totalCost }}</ElTag>
            </div>
          </div>

          <div
            v-if="selectedClass?.content"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-6 text-g-700"
          >
            <p class="mb-2 text-sm font-semibold text-g-900">课程说明</p>
            <div v-html="selectedClass.content"></div>
          </div>
        </div>
      </div>
    </section>

    <section class="art-card-sm overflow-hidden">
      <div class="border-b-d px-5 py-4">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-g-900">课程面板</h2>
            <p class="mt-1 text-sm text-g-500">先确定课程，再处理批量数据，避免提交到错误课程。</p>
          </div>
          <div class="flex flex-wrap gap-3">
            <ElInput
              v-model="keyword"
              class="w-[260px]"
              clearable
              placeholder="搜索课程名称"
              @keyup.enter="handleCourseSearch"
            />
            <ElButton type="primary" @click="handleCourseSearch">搜索</ElButton>
            <ElButton plain @click="resetCourseFilters">重置</ElButton>
          </div>
        </div>

        <div v-if="showCategoryPanel" class="mt-4 flex flex-wrap gap-2">
          <ElButton :type="activeCategory === '' ? 'primary' : 'default'" round @click="changeCategory('')">
            全部课程
          </ElButton>
          <ElButton :type="activeCategory === 'collect' ? 'danger' : 'default'" round @click="changeCategory('collect')">
            收藏课程
          </ElButton>
          <ElButton
            v-for="item in categories"
            :key="item.id"
            :type="activeCategory === String(item.id) ? 'primary' : 'default'"
            round
            @click="changeCategory(String(item.id))"
          >
            {{ item.name }}
          </ElButton>
        </div>
      </div>

      <ArtTableHeader :loading="courseLoading" layout="refresh" @refresh="refreshCourses">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">共 {{ coursePagination.total }} 门</ElTag>
            <ElTag v-if="favoriteIds.length" type="warning" effect="plain">收藏 {{ favoriteIds.length }}</ElTag>
            <ElTag v-if="selectedClass" type="success" effect="plain">
              已选 {{ selectedClass.name }}
            </ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <div v-loading="courseLoading" class="p-5">
        <div v-if="courses.length" class="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
          <article
            v-for="course in courses"
            :key="course.cid"
            class="group rounded-custom-sm border p-5 transition-all"
            :class="
              selectedClassId === course.cid
                ? 'border-[var(--el-color-primary)] bg-[var(--el-color-primary-light-9)]'
                : 'border-full-d bg-box hover:border-[var(--el-color-primary-light-5)]'
            "
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0 flex-1">
                <p class="line-clamp-2 text-base font-semibold text-g-900">{{ course.name }}</p>
                <p class="mt-2 text-xs font-medium text-g-400">课程 ID {{ course.cid }}</p>
              </div>
              <button
                type="button"
                class="rounded-full border px-3 py-1 text-xs transition"
                :class="
                  favoriteIds.includes(course.cid)
                    ? 'border-[var(--el-color-warning-light-7)] bg-[var(--el-color-warning-light-9)] text-[var(--el-color-warning)]'
                    : 'border-full-d bg-box text-g-500'
                "
                @click.stop="toggleFavorite(course.cid)"
              >
                {{ favoriteIds.includes(course.cid) ? '已收藏' : '收藏' }}
              </button>
            </div>

            <div class="mt-4 flex items-center justify-between gap-3">
              <div>
                <p class="text-sm text-g-500">价格</p>
                <p class="mt-1 text-xl font-semibold text-[var(--el-color-success)]">¥{{ formatLegacyPrice(course.price) }}</p>
              </div>
              <ElButton :type="selectedClassId === course.cid ? 'primary' : 'default'" round @click="selectCourse(course)">
                {{ selectedClassId === course.cid ? '已选中' : '选择课程' }}
              </ElButton>
            </div>
          </article>
        </div>

        <ElEmpty v-else description="当前筛选条件下没有课程" />
      </div>

      <div class="flex justify-end border-t-d px-5 py-4">
        <ElPagination
          background
          layout="total, prev, pager, next"
          :current-page="coursePagination.page"
          :page-size="coursePagination.limit"
          :total="coursePagination.total"
          @current-change="loadCourses"
        />
      </div>
    </section>

    <section v-if="parsedLines.length > 0" class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-g-900">数据预览</h2>
          <p class="mt-1 text-sm text-g-500">可删除无效行，确认数量后再提交。</p>
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <ElTag effect="plain" type="success">有效 {{ validCount }}</ElTag>
          <ElTag effect="plain" type="danger">无效 {{ invalidCount }}</ElTag>
          <ElTag v-if="selectedClass" effect="plain">预估 ¥{{ totalCost }}</ElTag>
          <ElButton
            type="primary"
            :loading="submitLoading"
            :disabled="validCount === 0 || !selectedClassId"
            @click="handleSubmit"
          >
            确认批量提交（{{ validCount }}）
          </ElButton>
        </div>
      </div>

      <ArtTableHeader class="mt-5" :loading="submitLoading">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">预览 {{ parsedLines.length }} 条</ElTag>
            <ElTag type="success" effect="plain">有效 {{ validCount }}</ElTag>
            <ElTag type="danger" effect="plain">无效 {{ invalidCount }}</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable :data="parsedLines" :columns="columns" :show-table-header="true" row-key="key" />
    </section>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElButton, ElMessage, ElTag } from 'element-plus'
  import {
    createLegacyOrder,
    type LegacyOrderAddItem
  } from '@/api/legacy/class'
  import {
    formatLegacyPrice,
    useLegacyCourseCatalog
  } from '../shared/useLegacyCourseCatalog'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'OrderBatchAddPage' })

  interface ParsedLine {
    key: number
    pass: string
    raw: string
    school: string
    user: string
    valid: boolean
  }

  const STORAGE_KEY = 'admin-next-order-batch-normalize'

  const {
    categories,
    activeCategory,
    keyword,
    courseLoading,
    courses,
    favoriteIds,
    selectedClassId,
    selectedClass,
    coursePagination,
    showCategoryPanel,
    loadBaseData,
    loadCourses,
    refreshCourses,
    handleCourseSearch,
    resetCourseFilters,
    changeCategory,
    selectCourse,
    toggleFavorite
  } = useLegacyCourseCatalog()

  const loadNormalizeSwitch = () => {
    try {
      return JSON.parse(localStorage.getItem(STORAGE_KEY) || 'true')
    } catch {
      return true
    }
  }

  const autoNormalize = ref(loadNormalizeSwitch())
  const rawText = ref('')
  const parsedLines = ref<ParsedLine[]>([])
  const submitLoading = ref(false)

  watch(autoNormalize, (value) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
  })

  const validCount = computed(() => parsedLines.value.filter((item) => item.valid).length)
  const invalidCount = computed(() => parsedLines.value.length - validCount.value)
  const totalCost = computed(() => {
    if (!selectedClass.value) {
      return '0.00'
    }
    return (validCount.value * Number(selectedClass.value.price || 0)).toFixed(2)
  })

  const { columns } = useTableColumns<ParsedLine>(() => [
    { prop: 'school', label: '学校', minWidth: 180, formatter: (row) => row.school || '自动识别' },
    { prop: 'user', label: '账号', minWidth: 180, formatter: (row) => row.user || '-' },
    { prop: 'pass', label: '密码', minWidth: 160, formatter: (row) => row.pass || '-' },
    {
      prop: 'valid',
      label: '状态',
      width: 100,
      align: 'center',
      formatter: (row) => h(ElTag, { type: row.valid ? 'success' : 'danger' }, () => (row.valid ? '有效' : '无效'))
    },
    { prop: 'raw', label: '原始内容', minWidth: 220, formatter: (row) => h('span', { class: 'text-g-500' }, row.raw || '-') },
    {
      prop: 'operation',
      label: '操作',
      width: 90,
      align: 'center',
      formatter: (row) => h(ElButton, { text: true, type: 'danger', onClick: () => removeLine(row.key) }, () => '删除')
    }
  ])

  const normalizeInputText = (value: string) =>
    value
      .replaceAll('\r\n', '\n')
      .replaceAll('，', ' ')
      .replaceAll('、', ' ')
      .replaceAll('\t', ' ')
      .replaceAll('\u3000', ' ')
      .split('\n')
      .map((line) => line.trim().replace(/\s+/g, ' '))
      .join('\n')

  const handleNormalizeOnBlur = () => {
    if (!autoNormalize.value) {
      return
    }
    rawText.value = normalizeInputText(rawText.value)
  }

  const handleParse = () => {
    if (!selectedClassId.value) {
      ElMessage.warning('请先选择课程')
      return
    }
    if (!rawText.value.trim()) {
      ElMessage.warning('请先粘贴批量下单信息')
      return
    }

    if (autoNormalize.value) {
      rawText.value = normalizeInputText(rawText.value)
    }

    const lines = rawText.value
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)

    parsedLines.value = lines.map((line, index) => {
      const parts = line.split(/\s+/).filter(Boolean)

      if (parts.length >= 3) {
        const [school, user, ...rest] = parts
        return {
          key: index,
          school,
          user,
          pass: rest.join(' '),
          raw: line,
          valid: true
        }
      }

      if (parts.length === 2) {
        const [user, pass] = parts
        return {
          key: index,
          school: '',
          user,
          pass,
          raw: line,
          valid: true
        }
      }

      return {
        key: index,
        school: '',
        user: '',
        pass: '',
        raw: line,
        valid: false
      }
    })
  }

  const removeLine = (key: number) => {
    parsedLines.value = parsedLines.value.filter((item) => item.key !== key)
  }

  const clearAll = () => {
    rawText.value = ''
    parsedLines.value = []
  }

  const buildUserInfo = (line: ParsedLine) =>
    [line.school.trim(), line.user.trim(), line.pass.trim()].filter(Boolean).join(' ')

  const handleSubmit = async () => {
    if (!selectedClassId.value) {
      ElMessage.warning('请先选择课程')
      return
    }

    const validLines = parsedLines.value.filter((item) => item.valid)
    if (!validLines.length) {
      ElMessage.warning('没有可提交的有效数据')
      return
    }

    submitLoading.value = true
    try {
      const payload: LegacyOrderAddItem[] = validLines.map((line) => ({
        userinfo: buildUserInfo(line),
        userName: line.user,
        data: {
          id: '',
          kcjs: '',
          name: selectedClass.value?.name || '',
          idx: -1,
          select: true
        }
      }))

      const result = await createLegacyOrder({
        cid: selectedClassId.value,
        data: payload
      })
      ElMessage.success(result.msg || `批量提交成功，共 ${validLines.length} 条`)
      clearAll()
    } finally {
      submitLoading.value = false
    }
  }

  onMounted(async () => {
    await loadBaseData()
    await loadCourses(1)
  })
</script>
