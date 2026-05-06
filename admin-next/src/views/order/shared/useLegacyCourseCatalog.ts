import { computed, reactive, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useSiteStore } from '@/store/modules/site'
import {
  addLegacyFavoriteCourse,
  fetchLegacyClassCategories,
  fetchLegacyClassListPaged,
  fetchLegacyFavoriteCourseIds,
  removeLegacyFavoriteCourse,
  type LegacyClassCategory,
  type LegacyClassItem
} from '@/api/legacy/class'

export function isLegacyQuerySuccess(msg?: string) {
  return msg === '查询成功' || msg === '此课程无需查课，直接下单即可'
}

export function formatLegacyPrice(value?: number | string) {
  return Number(value || 0).toFixed(2)
}

export function parseLegacyCategoryType(raw?: string) {
  const parsed = Number(raw ?? '1')
  return [0, 1, 2].includes(parsed) ? parsed : 1
}

export function useLegacyCourseCatalog(pageSize = 12) {
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
    limit: pageSize,
    total: 0
  })

  const selectedClass = computed(() => {
    if (!selectedClassId.value) {
      return null
    }
    return courses.value.find((item) => item.cid === selectedClassId.value) || selectedClassCache.value
  })

  const showCategoryPanel = computed(
    () => String(config.value.flkg ?? '1') !== '0' && categories.value.length > 0
  )

  const categoryType = computed(() => parseLegacyCategoryType(config.value.fllx))

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
        categories.value = Array.isArray(result) ? result : []
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

  const refreshCourses = async () => {
    await loadCourses(1)
  }

  const handleCourseSearch = async () => {
    await loadCourses(1)
  }

  const resetCourseFilters = async () => {
    keyword.value = ''
    activeCategory.value = ''
    await loadCourses(1)
  }

  const changeCategory = async (value: string) => {
    activeCategory.value = value
    await loadCourses(1)
  }

  const selectCourse = (course: LegacyClassItem) => {
    selectedClassId.value = course.cid
    selectedClassCache.value = course
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

  return {
    categories,
    activeCategory,
    keyword,
    courseLoading,
    courses,
    favoriteIds,
    selectedClassId,
    selectedClass,
    coursePagination,
    hasMoreCourses,
    showCategoryPanel,
    categoryType,
    loadBaseData,
    loadCourses,
    loadMoreCourses,
    refreshCourses,
    handleCourseSearch,
    resetCourseFilters,
    changeCategory,
    selectCourse,
    toggleFavorite
  }
}
