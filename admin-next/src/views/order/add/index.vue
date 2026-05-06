<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-4">
    <section class="art-card-sm p-5">
      <div class="space-y-5">
        <div class="space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <p class="text-base font-semibold text-g-900">选择课程</p>
            <div v-if="showCategoryPanel" class="flex flex-wrap items-center gap-2">
              <ElTag v-if="shouldShowCategoryPanel && activeCategory === 'collect'" type="warning" effect="plain" size="small">收藏中</ElTag>
              <ElButton text size="small" @click="showCategoryToggle = !showCategoryToggle">
                {{ shouldShowCategoryPanel ? '收起分类' : '展开分类' }}
              </ElButton>
            </div>
          </div>

          <div v-if="shouldShowCategoryPanel" class="space-y-3">
            <div v-if="categoryType === 1" class="max-w-[340px]">
              <ElSelect
                :model-value="activeCategory || undefined"
                clearable
                placeholder="全部课程"
                class="w-full"
                @change="handleCategorySelect"
              >
                <ElOption label="收藏课程" value="collect" />
                <ElOption
                  v-for="item in categories"
                  :key="item.id"
                  :label="item.name"
                  :value="String(item.id)"
                />
              </ElSelect>
            </div>
            <div v-else class="flex flex-wrap gap-1">
              <ElButton
                size="small"
                class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                :class="
                  activeCategory === ''
                    ? '!border-[var(--el-color-primary-light-5)] !bg-[var(--el-color-primary-light-9)] !text-[var(--el-color-primary)]'
                    : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                "
                @click="changeCategory('')"
              >
                全部课程
              </ElButton>
              <ElButton
                size="small"
                class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                :class="
                  activeCategory === 'collect'
                    ? '!border-[#ffc9de] !bg-[#fff4f8] !text-[#eb2f96]'
                    : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                "
                @click="changeCategory(activeCategory === 'collect' ? '' : 'collect')"
              >
                收藏课程
              </ElButton>
              <ElButton
                v-for="item in categories"
                :key="item.id"
                size="small"
                class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                :class="
                  activeCategory === String(item.id)
                    ? '!border-[var(--el-color-primary-light-5)] !bg-[var(--el-color-primary-light-9)] !text-[var(--el-color-primary)]'
                    : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                "
                @click="changeCategory(String(item.id))"
              >
                <span>{{ item.name }}</span>
                <span
                  v-if="item.recommend && activeCategory !== String(item.id)"
                  class="ml-1 rounded-full bg-[#f3e8ff] px-1.5 py-0.5 text-[10px] leading-none text-[#722ed1]"
                >
                  荐
                </span>
              </ElButton>
            </div>
          </div>

          <ElSelect
            :model-value="selectedClassId"
            class="w-full"
            clearable
            filterable
            remote
            reserve-keyword
            placeholder="搜索并选择课程"
            no-data-text="当前筛选条件下没有课程"
            popper-class="order-add-course-select-dropdown"
            :remote-method="handleCourseKeywordSearch"
            @change="handleCourseSelect"
            @clear="handleCourseQueryClear"
            @visible-change="handleCourseDropdownVisibleChange"
          >
            <ElOption
              v-for="course in courseSelectOptions"
              :key="course.cid"
              :label="course.name"
              :value="course.cid"
            >
              <span class="text-sm text-g-900">{{ course.name }}</span>
              <span class="ml-1 text-g-500">¥{{ formatPrice(course.price) }}</span>
            </ElOption>
          </ElSelect>

          <div
            v-if="selectedClass?.content"
            class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3"
          >
            <div class="flex flex-wrap items-center justify-between gap-2 border-b-d pb-2">
              <p class="text-sm font-medium text-g-800">课程说明</p>
              <ElButton
                v-if="selectedClass"
                text
                size="small"
                class="!text-g-500"
                @click="toggleFavorite(selectedClass.cid)"
              >
                {{ favoriteIds.includes(selectedClass.cid) ? '取消收藏' : '添加收藏' }}
              </ElButton>
            </div>
            <div class="mt-3 text-sm leading-6 text-g-600" v-html="selectedClass.content"></div>
          </div>
        </div>

        <div class="border-t-d pt-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <p class="text-base font-semibold text-g-900">下单信息</p>
            <div class="flex flex-wrap items-center gap-2 text-xs">
              <ElButton text size="small" @click="showRemarks = !showRemarks">
                {{ showRemarks ? '收起备注' : '添加备注' }}
              </ElButton>
              <ElButton text size="small" @click="isMultiline = !isMultiline">
                {{ isMultiline ? '多行输入' : '单行输入' }}
              </ElButton>
            </div>
          </div>

          <div class="mt-3 space-y-3">
            <ElInput
              v-model="userInfo"
              :type="isMultiline ? 'textarea' : 'text'"
              :rows="isMultiline ? 8 : undefined"
              resize="none"
              :placeholder="
                isMultiline
                  ? '请输入下单信息\n学校 账号 密码（空格分开），例如：家里蹲大学 13872325008 123456\n手机号 密码（空格分开），例如：13872325008 123456\n多账号请换行输入，查课后可继续勾选提交'
                  : '请输入学校 账号 密码，或手机号 密码'
              "
              @keyup.enter="!isMultiline && handleQuery()"
            />

            <ElInput
              v-if="showRemarks"
              v-model="remarks"
              type="textarea"
              :rows="2"
              resize="none"
              placeholder="选填备注：给这批订单追加说明"
            />

            <div class="mt-3 flex flex-wrap justify-end gap-3">
              <ElButton type="primary" :disabled="!selectedClass" :loading="queryLoading" @click="handleQuery">
                查询课程
              </ElButton>
              <ElButton plain @click="clearForm">清空输入</ElButton>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section v-if="queryLoading || queryResults.length > 0" ref="resultsSectionRef" class="art-card-sm p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <p class="text-base font-semibold text-g-900">查课结果</p>
        <ElButton
          type="primary"
          :loading="submitLoading"
          :disabled="selectedCourseCount === 0"
          @click="submitOrder"
        >
          提交订单（{{ selectedCourseCount }}）
        </ElButton>
      </div>

      <div v-if="queryLoading" class="py-16 text-center">
        <ElIcon class="is-loading text-2xl text-[var(--el-color-primary)]"><Loading /></ElIcon>
        <p class="mt-4 text-sm text-g-500">正在逐个账号查课，请稍候...</p>
      </div>

      <div v-else class="mt-5 space-y-4">
        <article
          v-for="(result, index) in queryResults"
          :key="`${result.userinfo}-${index}`"
          class="overflow-hidden rounded-custom-sm border-full-d bg-box"
        >
          <div class="flex flex-wrap items-center justify-between gap-3 border-b-d bg-g-100/50 px-5 py-4">
            <div>
              <p class="text-base font-semibold text-g-900">{{ result.userName || '未识别账号' }}</p>
              <p class="mt-1 text-sm text-g-500">{{ result.userinfo }}</p>
            </div>
            <ElTag :type="isQuerySuccess(result.msg) ? 'success' : 'danger'">{{ result.msg || '查询结果' }}</ElTag>
          </div>

          <div class="p-5">
            <div v-if="result.data.length" class="space-y-3">
              <div class="flex justify-end">
                <ElButton text type="primary" @click="toggleAllCourses(result)">
                  {{ areAllCoursesSelected(result) ? '取消全选' : '全选本账号课程' }}
                </ElButton>
              </div>

              <div class="grid gap-3">
                <button
                  v-for="course in result.data"
                  :key="`${result.userinfo}-${course.idx}`"
                  type="button"
                  class="flex items-center justify-between gap-4 rounded-custom-sm border px-4 py-4 text-left transition"
                  :class="
                    course.select
                      ? 'border-[var(--el-color-primary)] bg-[var(--el-color-primary-light-9)]'
                      : 'border-full-d bg-box hover:border-[var(--el-color-primary-light-5)]'
                  "
                  @click="toggleCourse(result, course)"
                >
                  <div class="min-w-0 flex-1">
                    <p class="line-clamp-2 text-sm font-medium text-g-900">{{ course.name }}</p>
                    <p class="mt-1 text-xs text-g-400">课程 ID：{{ course.id || '-' }}</p>
                  </div>
                  <ElCheckbox :model-value="Boolean(course.select)" />
                </button>
              </div>
            </div>

            <div
              v-else-if="isQuerySuccess(result.msg)"
              class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-5 text-sm text-[var(--el-color-success)]"
            >
              该账号无需选课，已自动加入待下单列表。
            </div>

            <div v-else class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-5 text-sm text-[var(--el-color-danger)]">
              {{ result.msg || '查询失败' }}
            </div>
          </div>
        </article>
      </div>
    </section>

  </div>
</template>

<script setup lang="ts">
  import { nextTick } from 'vue'
  import { Loading } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import {
    addLegacyFavoriteCourse,
    createLegacyOrder,
    fetchLegacyClassCategories,
    fetchLegacyClassListPaged,
    fetchLegacyFavoriteCourseIds,
    queryLegacyCourses,
    removeLegacyFavoriteCourse,
    type LegacyClassCategory,
    type LegacyClassItem,
    type LegacyCourseItem,
    type LegacyCourseQueryResult,
    type LegacyOrderAddItem
  } from '@/api/legacy/class'
  import { storeToRefs } from 'pinia'
  import { useSiteStore } from '@/store/modules/site'
  import { parseLegacyCategoryType } from '../shared/useLegacyCourseCatalog'

  defineOptions({ name: 'OrderAddPage' })

  const STORAGE_KEY = 'admin-next-order-add-multiline'
  const CATEGORY_TOGGLE_KEY = 'order_show_cate'

  const siteStore = useSiteStore()
  const { config } = storeToRefs(siteStore)

  const categories = ref<LegacyClassCategory[]>([])
  const activeCategory = ref('')
  const keyword = ref('')
  const courseLoading = ref(false)
  const courses = ref<LegacyClassItem[]>([])
  const favoriteIds = ref<number[]>([])
  const selectedClassId = ref<number>()
  const selectedClassCache = ref<LegacyClassItem | null>(null)
  const hasMoreCourses = ref(false)

  const coursePagination = reactive({
    page: 1,
    limit: 12,
    total: 0
  })

  const isMultiline = ref(loadStoredBoolean(STORAGE_KEY, true))
  const showCategoryToggle = ref(loadStoredBoolean(CATEGORY_TOGGLE_KEY, true))
  const showRemarks = ref(false)
  const userInfo = ref('')
  const remarks = ref('')
  const queryLoading = ref(false)
  const submitLoading = ref(false)
  const queryResults = ref<LegacyCourseQueryResult[]>([])
  const checkedCourses = ref<LegacyOrderAddItem[]>([])
  const resultsSectionRef = ref<HTMLElement | null>(null)
  const courseDropdownScrollWrap = ref<HTMLElement | null>(null)

  function loadStoredBoolean(key: string, defaultValue: boolean) {
    try {
      return JSON.parse(localStorage.getItem(key) || JSON.stringify(defaultValue))
    } catch {
      return defaultValue
    }
  }

  watch(isMultiline, (value) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
  })

  watch(showCategoryToggle, (value) => {
    localStorage.setItem(CATEGORY_TOGGLE_KEY, JSON.stringify(value))
  })

  const selectedClass = computed(() => {
    if (!selectedClassId.value) return null
    return courses.value.find((item) => item.cid === selectedClassId.value) || selectedClassCache.value
  })

  const selectedCourseCount = computed(() => checkedCourses.value.length)

  const showCategoryPanel = computed(
    () => String(config.value.flkg ?? '1') !== '0' && categories.value.length > 0
  )

  const shouldShowCategoryPanel = computed(() => showCategoryPanel.value && showCategoryToggle.value)

  const categoryType = computed(() => parseLegacyCategoryType(config.value.fllx))

  const courseSelectOptions = computed(() => courses.value)

  const isQuerySuccess = (msg?: string) =>
    msg === '查询成功' || msg === '此课程无需查课，直接下单即可'

  const formatPrice = (value?: number | string) => Number(value || 0).toFixed(2)

  const scrollToResults = async () => {
    await nextTick()
    resultsSectionRef.value?.scrollIntoView({
      behavior: 'smooth',
      block: 'start'
    })
  }

  const buildCourseParams = (page = coursePagination.page) => {
    const params: Record<string, number | string> = {
      page,
      limit: coursePagination.limit
    }

    if (keyword.value.trim()) {
      params.search = keyword.value.trim()
    }

    if (activeCategory.value === 'collect') {
      params.favorite = 1
    } else if (activeCategory.value) {
      const fenlei = Number(activeCategory.value)
      if (!Number.isNaN(fenlei) && fenlei > 0) {
        params.fenlei = fenlei
      }
    }

    return params
  }

  const loadBaseData = async () => {
    await Promise.all([
      siteStore.initPublicConfig(true),
      fetchLegacyClassCategories().then((result) => {
        categories.value = result || []
      }),
      fetchLegacyFavoriteCourseIds().then((result) => {
        favoriteIds.value = Array.isArray(result) ? result : []
      })
    ])
  }

  const loadCourses = async (page = coursePagination.page, append = false) => {
    if (courseLoading.value) {
      return
    }

    courseLoading.value = true
    coursePagination.page = page
    try {
      const result = await fetchLegacyClassListPaged(buildCourseParams(page))
      const nextList = Array.isArray(result?.list) ? result.list : []
      courses.value = append ? [...courses.value, ...nextList] : nextList
      coursePagination.page = Number(result?.pagination?.page || page)
      coursePagination.total = Number(result?.pagination?.total || 0)
      hasMoreCourses.value = Boolean(result?.pagination?.has_more)
      const current = courses.value.find((item) => item.cid === selectedClassId.value)
      if (current) {
        selectedClassCache.value = current
      }
    } finally {
      courseLoading.value = false
    }
  }

  const loadMoreCourses = async () => {
    if (!hasMoreCourses.value || courseLoading.value) {
      return
    }

    void loadCourses(coursePagination.page + 1, true)
  }

  const changeCategory = async (value: string) => {
    activeCategory.value = value
    await loadCourses(1)
  }

  const handleCourseKeywordSearch = async (value: string) => {
    keyword.value = value.trim()
    await loadCourses(1)
  }

  const handleCourseQueryClear = async () => {
    selectedClassId.value = undefined
    keyword.value = ''
    await loadCourses(1)
  }

  const handleCategorySelect = async (value?: string) => {
    await changeCategory(value ? String(value) : '')
  }

  const selectCourse = (course: LegacyClassItem) => {
    selectedClassId.value = course.cid
    selectedClassCache.value = course
  }

  const handleCourseDropdownVisibleChange = async (visible: boolean) => {
    if (visible) {
      await nextTick()
      bindCourseDropdownScroll('.order-add-course-select-dropdown .el-select-dropdown__wrap')
      return
    }

    unbindCourseDropdownScroll()
  }

  const handleCourseDropdownScroll = () => {
    const wrap = courseDropdownScrollWrap.value
    if (!wrap || courseLoading.value || !hasMoreCourses.value) {
      return
    }

    const remaining = wrap.scrollHeight - wrap.scrollTop - wrap.clientHeight
    const preloadDistance = Math.max(160, wrap.clientHeight * 0.6)
    if (remaining <= preloadDistance) {
      void loadMoreCourses()
    }
  }

  const bindCourseDropdownScroll = (selector: string) => {
    unbindCourseDropdownScroll()
    const wrap = document.querySelector(selector) as HTMLElement | null
    if (!wrap) {
      return
    }

    courseDropdownScrollWrap.value = wrap
    wrap.addEventListener('scroll', handleCourseDropdownScroll, { passive: true })
  }

  const unbindCourseDropdownScroll = () => {
    if (!courseDropdownScrollWrap.value) {
      return
    }

    courseDropdownScrollWrap.value.removeEventListener('scroll', handleCourseDropdownScroll)
    courseDropdownScrollWrap.value = null
  }

  const handleCourseSelect = (cid?: number) => {
    if (!cid) {
      selectedClassId.value = undefined
      return
    }

    const course = courses.value.find((item) => item.cid === cid)
    if (course) {
      selectCourse(course)
    }
  }

  const toggleFavorite = async (cid: number) => {
    if (favoriteIds.value.includes(cid)) {
      await removeLegacyFavoriteCourse(cid)
      favoriteIds.value = favoriteIds.value.filter((item) => item !== cid)
      ElMessage.success('已取消收藏')
      if (activeCategory.value === 'collect') {
        await loadCourses(1)
      }
      return
    }

    await addLegacyFavoriteCourse(cid)
    favoriteIds.value = [...favoriteIds.value, cid]
    ElMessage.success('已加入收藏')
  }

  const clearForm = () => {
    userInfo.value = ''
    remarks.value = ''
  }

  const normalizeQueryResult = (
    payload: LegacyCourseQueryResult,
    fallbackUserInfo: string
  ): LegacyCourseQueryResult => ({
    ...payload,
    userinfo: payload.userinfo || fallbackUserInfo,
    userName: payload.userName || '',
    data: Array.isArray(payload.data)
      ? payload.data.map((item, index) => ({
          ...item,
          idx: index,
          select: false
        }))
      : []
  })

  const buildDirectOrder = (result: LegacyCourseQueryResult): LegacyOrderAddItem => ({
    userinfo: result.userinfo,
    userName: result.userName,
    data: {
      id: '',
      name: '无需查课直接下单',
      idx: -1,
      select: true
    }
  })

  const handleQuery = async () => {
    if (!selectedClassId.value) {
      ElMessage.warning('请先选择课程')
      return
    }

    const lines = userInfo.value
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean)

    if (!lines.length) {
      ElMessage.warning('请先填写下单信息')
      return
    }

    queryLoading.value = true
    queryResults.value = []
    checkedCourses.value = []
    await scrollToResults()

    const tasks = lines.map(async (line) => {
      try {
        const result = normalizeQueryResult(await queryLegacyCourses(selectedClassId.value!, line), line)
        queryResults.value.push(result)
        if (isQuerySuccess(result.msg) && result.data.length === 0) {
          checkedCourses.value.push(buildDirectOrder(result))
        }
      } catch (error: any) {
        queryResults.value.push({
          userinfo: line,
          userName: '',
          msg: error?.message || '查询失败',
          data: []
        })
      }
    })

    await Promise.allSettled(tasks)
    queryLoading.value = false
    await scrollToResults()

    if (queryResults.value.some((item) => isQuerySuccess(item.msg))) {
      ElMessage.success('查课完成')
    }
  }

  const buildCheckedKey = (userinfo: string, course: LegacyCourseItem) =>
    `${userinfo}_${course.idx ?? course.id ?? course.name}`

  const toggleCourse = (result: LegacyCourseQueryResult, course: LegacyCourseItem) => {
    course.select = !course.select
    const key = buildCheckedKey(result.userinfo, course)
    const index = checkedCourses.value.findIndex(
      (item) => buildCheckedKey(item.userinfo, item.data) === key
    )

    if (course.select && index === -1) {
      checkedCourses.value.push({
        userinfo: result.userinfo,
        userName: result.userName,
        data: course
      })
    }

    if (!course.select && index >= 0) {
      checkedCourses.value.splice(index, 1)
    }
  }

  const areAllCoursesSelected = (result: LegacyCourseQueryResult) =>
    result.data.length > 0 && result.data.every((item) => item.select)

  const toggleAllCourses = (result: LegacyCourseQueryResult) => {
    const nextState = !areAllCoursesSelected(result)

    checkedCourses.value = checkedCourses.value.filter((item) => item.userinfo !== result.userinfo)
    result.data.forEach((course) => {
      course.select = nextState
      if (nextState) {
        checkedCourses.value.push({
          userinfo: result.userinfo,
          userName: result.userName,
          data: course
        })
      }
    })
  }

  const submitOrder = async () => {
    if (!selectedClassId.value) {
      ElMessage.warning('请先选择课程')
      return
    }
    if (!checkedCourses.value.length) {
      ElMessage.warning('请先勾选待下单课程')
      return
    }

    submitLoading.value = true
    try {
      const result = await createLegacyOrder({
        cid: selectedClassId.value,
        data: checkedCourses.value,
        remarks: remarks.value.trim() || undefined
      })
      ElMessage.success(result.msg || '下单成功')
      queryResults.value = []
      checkedCourses.value = []
      clearForm()
    } finally {
      submitLoading.value = false
    }
  }

  onMounted(async () => {
    await loadBaseData()
    await loadCourses(1)
  })

  onUnmounted(() => {
    unbindCourseDropdownScroll()
  })
</script>

<style>
</style>
