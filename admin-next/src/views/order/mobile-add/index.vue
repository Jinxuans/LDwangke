<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <div class="space-y-4">
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
            <div v-if="categoryType === 1" class="max-w-full sm:max-w-[320px]">
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
            popper-class="order-mobile-course-select-dropdown"
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
              <span class="ml-1 text-g-500">¥{{ formatLegacyPrice(course.price) }}</span>
            </ElOption>
          </ElSelect>

          <div v-if="selectedClass?.content" class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
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

        <div class="rounded-custom-sm border-full-d bg-box p-4">
          <p class="text-base font-semibold text-g-900">下单信息</p>

          <div class="mt-3 grid gap-3">
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">学校</p>
              <ElInput v-model="school" placeholder="选填，不填则按账号密码查询" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">账号</p>
              <ElInput v-model="account" placeholder="请输入账号" @keyup.enter="handleQuery" />
            </div>
            <div>
              <p class="mb-2 text-sm font-medium text-g-800">密码</p>
              <ElInput v-model="password" placeholder="请输入密码" show-password @keyup.enter="handleQuery" />
            </div>
          </div>

          <div class="mt-5 flex flex-wrap gap-3">
            <ElButton type="primary" :disabled="!selectedClass" :loading="queryLoading" @click="handleQuery">
              一键查课
            </ElButton>
            <ElButton plain @click="clearAccountForm">清空</ElButton>
          </div>
        </div>
      </div>
    </section>

    <section v-if="queryLoading || queryResults.length > 0" ref="resultsSectionRef" class="art-card-sm p-4">
      <div class="flex items-center justify-between gap-3">
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
        <p class="mt-4 text-sm text-g-500">正在查询课程，请稍候...</p>
      </div>

      <div v-else class="mt-5 space-y-4">
        <article
          v-for="(result, index) in queryResults"
          :key="`${result.userinfo}-${index}`"
          class="overflow-hidden rounded-custom-sm border-full-d bg-box"
        >
          <div class="flex flex-wrap items-center justify-between gap-3 border-b-d bg-g-100/50 px-5 py-4">
            <div>
              <p class="text-base font-semibold text-g-900">{{ result.userName || account || '未识别账号' }}</p>
              <p class="mt-1 text-sm text-g-500">{{ result.userinfo }}</p>
            </div>
            <ElTag :type="isLegacyQuerySuccess(result.msg) ? 'success' : 'danger'">{{ result.msg || '查询结果' }}</ElTag>
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
              v-else-if="isLegacyQuerySuccess(result.msg)"
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
  import {
    createLegacyOrder,
    queryLegacyCourses,
    type LegacyCourseItem,
    type LegacyCourseQueryResult,
    type LegacyOrderAddItem
  } from '@/api/legacy/class'
  import {
    formatLegacyPrice,
    isLegacyQuerySuccess,
    useLegacyCourseCatalog
  } from '../shared/useLegacyCourseCatalog'

  defineOptions({ name: 'OrderMobileAddPage' })

  const CATEGORY_TOGGLE_KEY = 'order_show_cate'

  const {
    categories,
    activeCategory,
    keyword,
    courseLoading,
    courses,
    favoriteIds,
    selectedClassId,
    selectedClass,
    hasMoreCourses,
    showCategoryPanel,
    categoryType,
    loadBaseData,
    loadCourses,
    loadMoreCourses,
    changeCategory,
    selectCourse,
    toggleFavorite
  } = useLegacyCourseCatalog()

  const showCategoryToggle = ref(loadStoredBoolean(CATEGORY_TOGGLE_KEY, true))
  const shouldShowCategoryPanel = computed(() => showCategoryPanel.value && showCategoryToggle.value)

  const school = ref('')
  const account = ref('')
  const password = ref('')
  const queryLoading = ref(false)
  const submitLoading = ref(false)
  const queryResults = ref<LegacyCourseQueryResult[]>([])
  const checkedCourses = ref<LegacyOrderAddItem[]>([])
  const resultsSectionRef = ref<HTMLElement | null>(null)
  const courseDropdownScrollWrap = ref<HTMLElement | null>(null)

  const courseSelectOptions = computed(() => courses.value)

  const selectedCourseCount = computed(() => checkedCourses.value.length)

  function loadStoredBoolean(key: string, defaultValue: boolean) {
    try {
      return JSON.parse(localStorage.getItem(key) || JSON.stringify(defaultValue))
    } catch {
      return defaultValue
    }
  }

  watch(showCategoryToggle, (value) => {
    localStorage.setItem(CATEGORY_TOGGLE_KEY, JSON.stringify(value))
  })

  const buildUserInfo = () =>
    [school.value.trim(), account.value.trim(), password.value.trim()].filter(Boolean).join(' ')

  const clearAccountForm = () => {
    school.value = ''
    account.value = ''
    password.value = ''
  }

  const handleCategorySelect = async (value?: string) => {
    await changeCategory(value ? String(value) : '')
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

  const handleCourseDropdownVisibleChange = async (visible: boolean) => {
    if (visible) {
      await nextTick()
      bindCourseDropdownScroll('.order-mobile-course-select-dropdown .el-select-dropdown__wrap')
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

  const scrollToResults = async () => {
    await nextTick()
    resultsSectionRef.value?.scrollIntoView({
      behavior: 'smooth',
      block: 'start'
    })
  }

  const normalizeQueryResult = (
    payload: LegacyCourseQueryResult,
    fallbackUserInfo: string
  ): LegacyCourseQueryResult => ({
    ...payload,
    userinfo: payload.userinfo || fallbackUserInfo,
    userName: payload.userName || account.value.trim(),
    data: Array.isArray(payload.data)
      ? payload.data.map((item, index) => ({
          ...item,
          idx: index,
          select: false
        }))
      : []
  })

  const buildCheckedKey = (userinfo: string, course: LegacyCourseItem) =>
    `${userinfo}_${course.idx ?? course.id ?? course.name}`

  const buildDirectOrder = (result: LegacyCourseQueryResult): LegacyOrderAddItem => ({
    userinfo: result.userinfo,
    userName: result.userName || account.value.trim(),
    data: {
      id: '',
      kcjs: '',
      name: selectedClass.value?.name || '无需查课直接下单',
      idx: -1,
      select: true
    }
  })

  const handleQuery = async () => {
    if (!selectedClassId.value) {
      ElMessage.warning('请先选择课程')
      return
    }
    if (!account.value.trim()) {
      ElMessage.warning('请填写账号')
      return
    }
    if (!password.value.trim()) {
      ElMessage.warning('请填写密码')
      return
    }

    queryLoading.value = true
    queryResults.value = []
    checkedCourses.value = []

    const userinfo = buildUserInfo()
    await scrollToResults()

    try {
      const result = normalizeQueryResult(await queryLegacyCourses(selectedClassId.value, userinfo), userinfo)
      queryResults.value = [result]

      if (isLegacyQuerySuccess(result.msg) && result.data.length === 0) {
        checkedCourses.value = [buildDirectOrder(result)]
      }
    } catch (error: any) {
      queryResults.value = [
        {
          userinfo,
          userName: account.value.trim(),
          msg: error?.message || '查询失败',
          data: []
        }
      ]
    } finally {
      queryLoading.value = false
      await scrollToResults()
    }
  }

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
        data: checkedCourses.value
      })
      ElMessage.success(result.msg || '下单成功')
      queryResults.value = []
      checkedCourses.value = []
      clearAccountForm()
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
