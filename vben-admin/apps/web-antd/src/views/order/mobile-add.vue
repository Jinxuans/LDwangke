<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';

import {
  HeartFilled,
  HeartOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue';
import { useDebounceFn } from '@vueuse/core';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Empty,
  Input,
  message,
  Select,
  SelectOption,
  Spin,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { getSiteConfigApi } from '#/api/admin';
import {
  addOrderApi,
  type ClassCategory,
  type ClassItem,
  type CourseItem,
  type CourseQueryResult,
  getClassCategoriesApi,
  getClassListPagedApi,
  queryCourseApi,
} from '#/api/class';
import {
  addFavoriteApi,
  getFavoritesApi,
  removeFavoriteApi,
} from '#/api/user-center';

const classLoading = ref(false);
const classList = ref<ClassItem[]>([]);
const categoryList = ref<ClassCategory[]>([]);
const activeCateId = ref<string>('');
const classKeyword = ref('');
const classHasMore = ref(false);
const classPage = ref(1);
const classPageSize = 20;
const classPagingLoading = ref(false);
const classPagingMoreLoading = ref(false);
const selectedClassId = ref<number>();
const selectedClassCache = ref<ClassItem | null>(null);
const selectedClassTips = ref('');
const showCategory = ref(true);
const categoryType = ref(0);
const queryLoading = ref(false);
const submitLoading = ref(false);
const resultsSectionRef = ref<HTMLElement | null>(null);

const school = ref('');
const account = ref('');
const password = ref('');

const favoriteCourses = ref<string[]>([]);
const courseResults = ref<CourseQueryResult[]>([]);
const checkedCourses = ref<
  Array<{ data: CourseItem; userinfo: string; userName: string }>
>([]);

const showResultsSection = computed(
  () => queryLoading.value || courseResults.value.length > 0,
);

function parseCategoryType(raw?: string) {
  const parsed = Number(raw ?? '1');
  return [0, 1, 2].includes(parsed) ? parsed : 1;
}

const selectedClass = computed(() => {
  if (!selectedClassId.value) {
    return null;
  }
  return (
    classList.value.find((item) => item.cid === selectedClassId.value) ??
    selectedClassCache.value
  );
});

const classOptions = computed(() => {
  const list = [...classList.value];
  if (
    selectedClassCache.value &&
    !list.some((item) => item.cid === selectedClassCache.value?.cid)
  ) {
    list.unshift(selectedClassCache.value);
  }
  return list.map((item) => ({
    label: `${item.name}（¥${item.price}）`,
    value: item.cid,
  }));
});

const isFavorite = computed(() =>
  selectedClassId.value
    ? favoriteCourses.value.includes(String(selectedClassId.value))
    : false,
);

async function loadSiteConfig() {
  try {
    const cfg = await getSiteConfigApi();
    showCategory.value = cfg?.flkg !== '0';
    categoryType.value = parseCategoryType(cfg?.fllx);
  } catch {
    /* ignore */
  }
}

function buildClassListParams(page: number) {
  const params: Record<string, number | string> = {
    limit: classPageSize,
    page,
  };
  const keyword = classKeyword.value.trim();
  if (keyword) {
    params.search = keyword;
  }
  if (activeCateId.value === 'collect') {
    params.favorite = 1;
  } else if (activeCateId.value) {
    const fenlei = Number(activeCateId.value);
    if (!Number.isNaN(fenlei) && fenlei > 0) {
      params.fenlei = fenlei;
    }
  }
  return params;
}

function mergeClasses(existing: ClassItem[], incoming: ClassItem[]) {
  const seen = new Set(existing.map((item) => item.cid));
  const merged = [...existing];
  for (const item of incoming) {
    if (!seen.has(item.cid)) {
      merged.push(item);
      seen.add(item.cid);
    }
  }
  return merged;
}

async function fetchClassPage(page: number, append = false) {
  if (append) {
    if (
      classPagingLoading.value ||
      classPagingMoreLoading.value ||
      !classHasMore.value
    ) {
      return;
    }
    classPagingMoreLoading.value = true;
  } else {
    classPagingLoading.value = true;
  }
  try {
    const result = await getClassListPagedApi(buildClassListParams(page));
    const list = Array.isArray(result?.list) ? result.list : [];
    classList.value = append ? mergeClasses(classList.value, list) : list;
    classPage.value = result?.pagination?.page ?? page;
    classHasMore.value = Boolean(result?.pagination?.has_more);
    const currentSelected = classList.value.find(
      (item) => item.cid === selectedClassId.value,
    );
    if (currentSelected) {
      selectedClassCache.value = currentSelected;
      selectedClassTips.value = currentSelected.content || '';
    }
  } catch (error) {
    console.error('加载课程失败:', error);
  } finally {
    classPagingLoading.value = false;
    classPagingMoreLoading.value = false;
  }
}

async function refreshClassList() {
  classPage.value = 1;
  classHasMore.value = false;
  await fetchClassPage(1, false);
}

function resetSelectedClass() {
  selectedClassId.value = undefined;
  selectedClassCache.value = null;
  selectedClassTips.value = '';
}

async function loadMoreClasses() {
  if (!classHasMore.value) {
    return;
  }
  await fetchClassPage(classPage.value + 1, true);
}

const debouncedHandleClassSearch = useDebounceFn((keyword: string) => {
  classKeyword.value = keyword.trim();
  resetSelectedClass();
  void refreshClassList();
}, 300);

function handleClassSearch(keyword: string) {
  debouncedHandleClassSearch(keyword);
}

function handleClassPopupScroll(event: Event) {
  const target = event.target as HTMLElement | null;
  if (!target) {
    return;
  }
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 24) {
    void loadMoreClasses();
  }
}

function handleCategoryChange(value: string | undefined) {
  activeCateId.value = value ? String(value) : '';
}

function handleCourseChange(value: number | undefined) {
  selectedClassId.value = value;
  selectedClassCache.value =
    classList.value.find((item) => item.cid === value) || null;
}

function handleCourseDropdownVisibleChange(open: boolean) {
  if (open && classList.value.length === 0 && !classPagingLoading.value) {
    void refreshClassList();
  }
}

function scrollToResults() {
  resultsSectionRef.value?.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  });
}

async function loadPageData() {
  classLoading.value = true;
  try {
    await loadSiteConfig();
    const categoriesRaw = await getClassCategoriesApi();
    categoryList.value = Array.isArray(categoriesRaw) ? categoriesRaw : [];
    try {
      const favs = await getFavoritesApi();
      favoriteCourses.value = (Array.isArray(favs) ? favs : []).map(String);
    } catch {
      /* ignore */
    }
    await refreshClassList();
  } catch (error) {
    console.error('加载页面失败:', error);
  } finally {
    classLoading.value = false;
  }
}

async function toggleFavorite() {
  if (!selectedClassId.value) {
    message.warning('请先选择课程');
    return;
  }
  try {
    if (isFavorite.value) {
      await removeFavoriteApi(selectedClassId.value);
      favoriteCourses.value = favoriteCourses.value.filter(
        (id) => id !== String(selectedClassId.value),
      );
      message.success('已取消收藏');
    } else {
      await addFavoriteApi(selectedClassId.value);
      favoriteCourses.value.push(String(selectedClassId.value));
      message.success('已添加收藏');
    }
    if (activeCateId.value === 'collect') {
      await refreshClassList();
    }
  } catch (error: any) {
    message.error(error?.message || '操作失败');
  }
}

function buildUserInfo() {
  const parts = [
    school.value.trim(),
    account.value.trim(),
    password.value.trim(),
  ].filter(Boolean);
  return parts.join(' ');
}

async function handleQuery() {
  if (!selectedClassId.value) {
    message.warning('请先选择课程');
    return;
  }
  if (!account.value.trim()) {
    message.warning('请填写账号');
    return;
  }
  if (!password.value.trim()) {
    message.warning('请填写密码');
    return;
  }

  queryLoading.value = true;
  courseResults.value = [];
  checkedCourses.value = [];

  try {
    const raw = await queryCourseApi(selectedClassId.value, buildUserInfo());
    const result = raw?.data && !raw.userinfo ? raw.data : raw;
    const items = (result.data || []).map((item: CourseItem, idx: number) => ({
      ...item,
      idx,
      select: false,
    }));
    const resultItem = {
      ...result,
      data: items,
    };
    courseResults.value = [resultItem];

    if (items.length === 0) {
      checkedCourses.value.push({
        data: { id: '', kcjs: '', name: selectedClass.value?.name || '' },
        userName: account.value.trim(),
        userinfo: buildUserInfo(),
      });
    }
  } catch (error: any) {
    courseResults.value = [
      {
        data: [],
        msg: error?.message || '查询失败',
        userName: account.value.trim(),
        userinfo: buildUserInfo(),
      },
    ];
    checkedCourses.value = [];
  } finally {
    queryLoading.value = false;
  }

  await nextTick();
  scrollToResults();
}

function toggleCourse(result: CourseQueryResult, course: CourseItem) {
  course.select = !course.select;
  const idx = checkedCourses.value.findIndex(
    (item) => item.userinfo === result.userinfo && item.data.idx === course.idx,
  );
  if (course.select && idx === -1) {
    checkedCourses.value.push({
      data: course,
      userName: result.userName,
      userinfo: result.userinfo,
    });
  } else if (!course.select && idx !== -1) {
    checkedCourses.value.splice(idx, 1);
  }
}

function toggleAll(result: CourseQueryResult) {
  const allSelected = result.data.every((item) => item.select);
  result.data.forEach((item) => {
    item.select = !allSelected;
  });
  checkedCourses.value = checkedCourses.value.filter(
    (item) => item.userinfo !== result.userinfo,
  );
  if (!allSelected) {
    result.data.forEach((item) => {
      checkedCourses.value.push({
        data: item,
        userName: result.userName,
        userinfo: result.userinfo,
      });
    });
  }
}

async function handleSubmit() {
  if (!selectedClassId.value) {
    message.warning('请先选择课程');
    return;
  }
  if (checkedCourses.value.length === 0) {
    message.warning('请选择要下单的课程');
    return;
  }

  submitLoading.value = true;
  try {
    await addOrderApi({
      cid: selectedClassId.value,
      data: checkedCourses.value,
    });
    message.success('下单成功');
    courseResults.value = [];
    checkedCourses.value = [];
    school.value = '';
    account.value = '';
    password.value = '';
  } catch (error: any) {
    message.error(error?.message || '下单失败');
  } finally {
    submitLoading.value = false;
  }
}

onMounted(loadPageData);
onActivated(loadSiteConfig);

watch(selectedClass, (cls) => {
  selectedClassTips.value = cls?.content || '';
});

watch(activeCateId, async () => {
  classKeyword.value = '';
  resetSelectedClass();
  await refreshClassList();
});
</script>

<template>
  <Page content-class="p-4" title="手机下单">
    <div class="mx-auto w-full max-w-[720px]">
      <Spin :spinning="classLoading">
        <Card class="overflow-hidden rounded-2xl shadow-sm">
          <div class="mb-5">
            <h3 class="m-0 text-lg font-semibold">手机下单</h3>
            <div class="mt-1 text-sm text-gray-500">
              适合手机用户，直接填写学校、账号、密码后查课提交
            </div>
          </div>

          <template v-if="showCategory && categoryList.length > 0">
            <div class="mb-4">
              <label class="mb-2 block text-sm text-gray-500">选择分类</label>
              <div v-if="categoryType === 1">
                <Select
                  :value="activeCateId || undefined"
                  allow-clear
                  placeholder="选择分类"
                  style="width: 100%"
                  @change="handleCategoryChange"
                >
                  <SelectOption value="collect">收藏课程</SelectOption>
                  <SelectOption
                    v-for="cat in categoryList"
                    :key="cat.id"
                    :value="String(cat.id)"
                  >
                    {{ cat.name }}
                  </SelectOption>
                </Select>
              </div>
              <div v-else class="flex flex-wrap gap-2">
                <Button
                  :type="activeCateId === '' ? 'primary' : 'default'"
                  size="small"
                  @click="activeCateId = ''"
                >
                  全部课程
                </Button>
                <Button
                  :style="{
                    borderColor: '#eb2f96',
                    color: activeCateId === 'collect' ? '' : '#eb2f96',
                  }"
                  :type="activeCateId === 'collect' ? 'primary' : 'default'"
                  size="small"
                  @click="
                    activeCateId = activeCateId === 'collect' ? '' : 'collect'
                  "
                >
                  <template #icon><HeartFilled /></template>
                  收藏课程
                </Button>
                <Button
                  v-for="cat in categoryList"
                  :key="cat.id"
                  :style="
                    cat.recommend
                      ? {
                          borderColor: '#722ed1',
                          color:
                            activeCateId === String(cat.id) ? '' : '#722ed1',
                          fontWeight: '600',
                        }
                      : {}
                  "
                  :type="
                    activeCateId === String(cat.id) ? 'primary' : 'default'
                  "
                  size="small"
                  @click="activeCateId = String(cat.id)"
                >
                  {{ cat.name }}
                </Button>
              </div>
            </div>
          </template>

          <div class="mb-4">
            <label class="mb-2 block text-sm text-gray-500">选择课程</label>
            <div class="flex items-end gap-2">
              <Select
                :filter-option="false"
                :loading="classPagingLoading || classPagingMoreLoading"
                :not-found-content="
                  classPagingLoading ? '课程加载中...' : '暂无课程'
                "
                :value="selectedClassId"
                allow-clear
                class="flex-1"
                placeholder="搜索课程名称"
                show-search
                @change="handleCourseChange"
                @dropdown-visible-change="handleCourseDropdownVisibleChange"
                @popup-scroll="handleClassPopupScroll"
                @search="handleClassSearch"
              >
                <SelectOption
                  v-for="option in classOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </SelectOption>
              </Select>
              <Tooltip
                v-if="selectedClassId"
                :title="isFavorite ? '取消收藏' : '添加收藏'"
              >
                <Button
                  :type="isFavorite ? 'primary' : 'default'"
                  danger
                  size="large"
                  @click="toggleFavorite"
                >
                  <template #icon>
                    <HeartFilled v-if="isFavorite" />
                    <HeartOutlined v-else />
                  </template>
                </Button>
              </Tooltip>
            </div>
          </div>

          <Alert v-if="selectedClassTips" class="mb-4" show-icon type="warning">
            <template #message>
              <div class="text-sm" v-html="selectedClassTips"></div>
            </template>
          </Alert>

          <div class="grid grid-cols-1 gap-4">
            <div>
              <label class="mb-2 block text-sm text-gray-500">学校</label>
              <Input
                v-model:value="school"
                placeholder="选填，不填则按账号密码查询"
                size="large"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm text-gray-500">账号</label>
              <Input
                v-model:value="account"
                placeholder="请输入账号"
                size="large"
                @press-enter="handleQuery"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm text-gray-500">密码</label>
              <Input
                v-model:value="password"
                placeholder="请输入密码"
                size="large"
                @press-enter="handleQuery"
              />
            </div>
          </div>

          <Alert
            class="mt-4"
            message="手机端建议逐个账号提交，查课后再勾选课程下单。"
            show-icon
            type="info"
          />

          <div class="mt-5">
            <Button
              :loading="queryLoading"
              block
              class="bg-blue-600 hover:bg-blue-500"
              size="large"
              type="primary"
              @click="handleQuery"
            >
              <template #icon><SearchOutlined /></template>
              一键查课
            </Button>
          </div>
        </Card>
      </Spin>
    </div>

    <div
      v-if="showResultsSection"
      ref="resultsSectionRef"
      class="mx-auto mt-4 w-full max-w-[720px]"
    >
      <Card class="overflow-hidden rounded-2xl shadow-sm">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="m-0 text-lg font-semibold">查课结果</h3>
            <div class="mt-1 text-sm text-gray-500">
              勾选课程后可直接在页面底部提交
            </div>
          </div>
          <Button
            :disabled="checkedCourses.length === 0"
            :loading="submitLoading"
            class="bg-green-600 hover:bg-green-500"
            type="primary"
            @click="handleSubmit"
          >
            提交 {{ checkedCourses.length }}
          </Button>
        </div>

        <div
          v-if="queryLoading"
          class="flex flex-col items-center justify-center rounded-lg bg-white py-16"
        >
          <Spin size="large" />
          <div class="mt-4 text-base text-gray-500">正在查询中，请稍候...</div>
        </div>

        <div
          v-else-if="courseResults.length === 0"
          class="rounded-lg bg-white py-12"
        >
          <Empty description="暂无查课结果" />
        </div>

        <template v-else>
          <Alert
            :message="`已选择 ${checkedCourses.length} 个课程准备下单`"
            class="mb-4"
            show-icon
            type="info"
          />

          <div class="space-y-4">
            <Card
              v-for="(result, index) in courseResults"
              :key="index"
              :body-style="{ padding: 0 }"
              bordered
            >
              <div
                class="flex items-center justify-between border-b border-gray-100 bg-gray-50 px-4 py-3"
              >
                <div class="flex flex-col leading-tight">
                  <span class="text-sm font-bold text-gray-800">
                    {{ result.userName || '未知账号' }}
                  </span>
                  <span class="mt-0.5 text-xs text-gray-400">
                    {{ result.userinfo }}
                  </span>
                </div>
                <Tag
                  :color="
                    result.msg === '查询成功' ||
                    result.msg === '此课程无需查课，直接下单即可'
                      ? 'success'
                      : 'error'
                  "
                  class="m-0"
                >
                  {{ result.msg }}
                </Tag>
              </div>

              <div class="p-4">
                <Table
                  v-if="result.data && result.data.length > 0"
                  :custom-row="
                    (record: CourseItem) => ({
                      onClick: () => toggleCourse(result, record),
                      style: { cursor: 'pointer' },
                    })
                  "
                  :data-source="result.data"
                  :pagination="false"
                  :row-class-name="
                    (record: CourseItem) => (record.select ? 'bg-blue-50' : '')
                  "
                  :scroll="{ x: 720 }"
                  row-key="idx"
                  size="middle"
                >
                  <Table.Column :width="60" align="center" title="">
                    <template #default="{ record }">
                      <Checkbox :checked="record.select" />
                    </template>
                    <template #title>
                      <Checkbox
                        :checked="
                          result.data.length > 0 &&
                          result.data.every((item: CourseItem) => item.select)
                        "
                        :indeterminate="
                          result.data.some((item: CourseItem) => item.select) &&
                          !result.data.every((item: CourseItem) => item.select)
                        "
                        @change="toggleAll(result)"
                      />
                    </template>
                  </Table.Column>
                  <Table.Column key="name" data-index="name" title="课程信息">
                    <template #default="{ record }">
                      <div
                        class="max-w-[420px] overflow-hidden text-ellipsis whitespace-nowrap text-sm"
                      >
                        {{ record.name }}
                      </div>
                    </template>
                  </Table.Column>
                  <Table.Column
                    key="id"
                    :width="120"
                    align="center"
                    data-index="id"
                    title="课程ID"
                  >
                    <template #default="{ record }">
                      <span
                        class="whitespace-nowrap font-mono text-xs text-gray-500"
                      >
                        {{ record.id }}
                      </span>
                    </template>
                  </Table.Column>
                </Table>

                <div
                  v-else-if="
                    result.msg === '查询成功' ||
                    result.msg === '此课程无需查课，直接下单即可'
                  "
                  class="flex flex-col items-center justify-center rounded-md border border-green-100 bg-green-50 py-6 text-green-600"
                >
                  <span class="mb-1 text-lg">无需选课</span>
                  <span class="text-sm">此账号可直接下单</span>
                </div>

                <div
                  v-else
                  class="flex flex-col items-center justify-center rounded-md border border-red-100 bg-red-50 py-6 text-red-500"
                >
                  <span class="mb-1 text-lg">查询失败</span>
                  <span class="text-sm">{{ result.msg }}</span>
                </div>
              </div>
            </Card>
          </div>

          <div class="mt-4">
            <Button
              :disabled="checkedCourses.length === 0"
              :loading="submitLoading"
              block
              class="bg-green-600 hover:bg-green-500"
              size="large"
              type="primary"
              @click="handleSubmit"
            >
              提交已选课程（{{ checkedCourses.length }}）
            </Button>
          </div>
        </template>
      </Card>
    </div>
  </Page>
</template>
