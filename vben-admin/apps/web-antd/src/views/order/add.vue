<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core';
import { ref, computed, onActivated, onMounted, watch } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card,
  Button,
  Select,
  SelectOption,
  Input,
  Switch,
  Table,
  Checkbox,
  Tag,
  Space,
  Alert,
  Spin,
  Tooltip,
  message,
  Modal,
  Empty,
} from 'ant-design-vue';
import {
  SearchOutlined,
  CheckOutlined,
  HeartOutlined,
  HeartFilled,
  EditOutlined,
  PlusOutlined,
  DeleteOutlined,
  SaveOutlined,
  CloseOutlined,
} from '@ant-design/icons-vue';
import { useVbenForm } from '#/adapter/form';
import {
  getClassListPagedApi,
  getClassCategoriesApi,
  queryCourseApi,
  addOrderApi,
  type ClassItem,
  type ClassCategory,
  type CourseQueryResult,
  type CourseItem,
} from '#/api/class';
import { getSiteConfigApi, saveConfigApi } from '#/api/admin';
import {
  getFavoritesApi,
  addFavoriteApi,
  removeFavoriteApi,
} from '#/api/user-center';
import { aiReviseMultiline } from '#/utils/ai-revise';
import { useAccessStore } from '@vben/stores';

// ===== 开关状态（持久化到 localStorage） =====
function loadSwitch(key: string, def: boolean): boolean {
  try {
    return JSON.parse(localStorage.getItem(key) ?? String(def));
  } catch {
    return def;
  }
}
function saveSwitch(key: string, val: boolean) {
  localStorage.setItem(key, JSON.stringify(val));
}

const isMultiline = ref(loadSwitch('order_multiline', true));
const aiFlag = ref(loadSwitch('order_ai_flag', true));
const showCategoryToggle = ref(loadSwitch('order_show_cate', true));

// 课程数据
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
const selectedClassId = ref<number | undefined>(undefined);
const selectedClassCache = ref<ClassItem | null>(null);
const selectedClassTips = ref('');
const showCategory = ref(true);
const categoryType = ref(0);
const xdsmopen = ref(false);
const queryLoading = ref(false);
const submitLoading = ref(false);
const courseModalVisible = ref(false);

// 收藏
const favoriteCourses = ref<string[]>([]);
const isFavorite = computed(() =>
  selectedClassId.value
    ? favoriteCourses.value.includes(String(selectedClassId.value))
    : false,
);

// 查课结果
const courseResults = ref<CourseQueryResult[]>([]);
const checkedCourses = ref<
  Array<{ userinfo: string; userName: string; data: CourseItem }>
>([]);

// 用户输入
const userInfo = ref('');

// ===== 推荐渠道 =====
const accessStore = useAccessStore();
const hasAdminRole = computed(() => {
  const codes = accessStore.accessCodes;
  return Array.isArray(codes) && codes.includes('admin');
});

interface RecommendChannel {
  name: string;
  desc: string;
}

const recommendChannels = ref<RecommendChannel[]>([]);
const editingChannels = ref(false);
const editChannelList = ref<RecommendChannel[]>([]);
const savingChannels = ref(false);

function parseCategoryType(raw?: string) {
  const parsed = Number(raw ?? '1');
  return [0, 1, 2].includes(parsed) ? parsed : 1;
}

function startEditChannels() {
  editChannelList.value = JSON.parse(JSON.stringify(recommendChannels.value));
  editingChannels.value = true;
}

function cancelEditChannels() {
  editingChannels.value = false;
}

function addChannel() {
  editChannelList.value.push({ name: '', desc: '' });
}

function removeChannel(idx: number) {
  editChannelList.value.splice(idx, 1);
}

async function saveChannels() {
  const valid = editChannelList.value.filter((c) => c.name.trim());
  savingChannels.value = true;
  try {
    await saveConfigApi({ recommend_channels: JSON.stringify(valid) });
    recommendChannels.value = valid;
    editingChannels.value = false;
    message.success('推荐渠道已保存');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    savingChannels.value = false;
  }
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

// ===== Vben 表单（仅课程选择） =====
const [OrderForm, orderFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: 'w-full' },
  },
  showDefaultActions: false,
  layout: 'vertical',
  schema: [
    {
      component: 'Select',
      fieldName: 'classId',
      label: '选择课程',
      componentProps: () => ({
        options: classOptions.value,
        showSearch: true,
        allowClear: true,
        filterOption: false,
        loading: classPagingLoading.value || classPagingMoreLoading.value,
        notFoundContent: classPagingLoading.value
          ? '课程加载中...'
          : '暂无课程',
        onDropdownVisibleChange: (open: boolean) => {
          if (
            open &&
            classList.value.length === 0 &&
            !classPagingLoading.value
          ) {
            void refreshClassList();
          }
        },
        placeholder: '请选择课程',
        onPopupScroll: handleClassPopupScroll,
        onSearch: handleClassSearch,
        onChange: (val: number) => {
          selectedClassId.value = val;
          const cls = classList.value.find((item) => item.cid === val);
          selectedClassCache.value = cls || null;
          selectedClassTips.value = cls?.content || '';
        },
      }),
    },
  ],
  wrapperClass: 'grid-cols-1',
});

async function loadSiteConfig() {
  try {
    const cfg = await getSiteConfigApi();
    // 分类开关默认开启：只有显式配置为 0 时才关闭。
    showCategory.value = cfg?.flkg !== '0';
    // 分类类型允许 0/1/2，缺省或异常时才回退到 1。
    categoryType.value = parseCategoryType(cfg?.fllx);
    // 下单说明默认开启：只有显式配置为 0 时才关闭。
    xdsmopen.value = cfg?.xdsmopen !== '0';
    // 加载推荐渠道
    if (cfg?.recommend_channels) {
      try {
        recommendChannels.value = JSON.parse(cfg.recommend_channels);
      } catch {
        /* ignore */
      }
    } else {
      recommendChannels.value = [];
    }
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
  void orderFormApi.setFieldValue('classId', undefined);
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

// 初始化
async function loadClassData() {
  classLoading.value = true;
  try {
    await loadSiteConfig();

    const categoriesRaw = await getClassCategoriesApi();
    categoryList.value = Array.isArray(categoriesRaw) ? categoriesRaw : [];

    // 加载收藏
    try {
      const favs = await getFavoritesApi();
      favoriteCourses.value = (Array.isArray(favs) ? favs : []).map(String);
    } catch {
      /* ignore */
    }

    await refreshClassList();
  } catch (e) {
    console.error('加载课程失败:', e);
  } finally {
    classLoading.value = false;
  }
}

// 收藏切换
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
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// AI 校正（输入框失焦时）
function handleBlurRevise() {
  if (!aiFlag.value) return;
  userInfo.value = aiReviseMultiline(userInfo.value);
}

// 查课（渐进式加载：每完成一个请求立即显示结果）
async function handleQuery() {
  const values = await orderFormApi.getValues();
  if (!values.classId) {
    message.warning('请先选择课程');
    return;
  }
  if (!userInfo.value?.trim()) {
    message.warning('请填写下单信息');
    return;
  }

  const lines = userInfo.value
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((l: string) => l.trim())
    .filter(Boolean);

  courseResults.value = [];
  checkedCourses.value = [];
  queryLoading.value = true;
  let hasSuccess = false;

  // 渐进式加载：每个请求独立处理
  const requests = lines.map((line: string) =>
    queryCourseApi(values.classId, line)
      .then((res: any) => {
        hasSuccess = true;
        const r = res?.data && !res.userinfo ? res.data : res;
        const resultItem = {
          ...r,
          data: (r.data || []).map((item: CourseItem, idx: number) => ({
            ...item,
            idx,
            select: false,
          })),
        };
        courseResults.value.push(resultItem);
        // 无需选课时（data 为空且查询成功），自动加入待下单列表
        if (
          resultItem.data.length === 0 &&
          (r.msg === '查询成功' || r.msg === '此课程无需查课，直接下单即可')
        ) {
          checkedCourses.value.push({
            userinfo: r.userinfo || line,
            userName: r.userName || '',
            data: { id: '', name: '', idx: 0, select: true } as any,
          });
        }
      })
      .catch((err: any) => {
        courseResults.value.push({
          userinfo: line,
          userName: '',
          msg: err?.message || '查询失败',
          data: [],
        });
      }),
  );

  courseModalVisible.value = true;

  // 等待全部完成
  await Promise.allSettled(requests);
  queryLoading.value = false;

  if (hasSuccess) {
    message.success('查课成功');
  }
}

// 选中/取消选中单个课程
function toggleCourse(result: CourseQueryResult, course: CourseItem) {
  course.select = !course.select;
  const key = `${result.userinfo}_${course.idx}`;
  const idx = checkedCourses.value.findIndex(
    (c) => `${c.userinfo}_${c.data.idx}` === key,
  );
  if (course.select && idx === -1) {
    checkedCourses.value.push({
      userinfo: result.userinfo,
      userName: result.userName,
      data: course,
    });
  } else if (!course.select && idx >= 0) {
    checkedCourses.value.splice(idx, 1);
  }
}

// 全选/全不选
function toggleAll(result: CourseQueryResult) {
  const allSelected = result.data.every((c) => c.select);
  result.data.forEach((c) => {
    c.select = !allSelected;
  });
  checkedCourses.value = checkedCourses.value.filter(
    (c) => c.userinfo !== result.userinfo,
  );
  if (!allSelected) {
    result.data.forEach((c) => {
      checkedCourses.value.push({
        userinfo: result.userinfo,
        userName: result.userName,
        data: c,
      });
    });
  }
}

// 下单
async function handleSubmit() {
  const values = await orderFormApi.getValues();
  if (!values.classId) {
    message.warning('请先查课');
    return;
  }
  if (checkedCourses.value.length === 0) {
    message.warning('请选择要下单的课程');
    return;
  }

  submitLoading.value = true;
  try {
    await addOrderApi({
      cid: values.classId,
      data: checkedCourses.value,
    });
    message.success('下单成功');
    courseModalVisible.value = false;
    courseResults.value = [];
    checkedCourses.value = [];
    userInfo.value = '';
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    submitLoading.value = false;
  }
}

onMounted(loadClassData);
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
  <Page title="查课交单" content-class="p-4">
    <div
      class="flex flex-col gap-4 lg:flex-row"
      style="align-items: flex-start"
    >
      <!-- 左侧：查课交单主区域 -->
      <div class="w-full lg:min-w-0 lg:flex-1">
        <Spin :spinning="classLoading">
          <Card class="mb-4">
            <!-- 顶部开关栏 -->
            <div class="mb-4 flex items-center justify-between">
              <h3 class="m-0 text-base font-semibold">查课交单</h3>
              <Space>
                <Tooltip title="显示/隐藏分类选择">
                  <Switch
                    v-model:checked="showCategoryToggle"
                    checked-children="分类"
                    un-checked-children="隐藏"
                    @change="(v: any) => saveSwitch('order_show_cate', v)"
                  />
                </Tooltip>
                <Tooltip title="切换输入模式">
                  <Switch
                    v-model:checked="isMultiline"
                    checked-children="多行"
                    un-checked-children="单行"
                    @change="(v: any) => saveSwitch('order_multiline', v)"
                  />
                </Tooltip>
                <Tooltip title="若校正有误，请关闭此功能">
                  <Switch
                    v-model:checked="aiFlag"
                    checked-children="AI校正"
                    un-checked-children="关闭"
                    @change="(v: any) => saveSwitch('order_ai_flag', v)"
                  />
                </Tooltip>
              </Space>
            </div>

            <!-- 分类选择 -->
            <template
              v-if="
                showCategory && showCategoryToggle && categoryList.length > 0
              "
            >
              <div class="mb-4">
                <label class="mb-2 block text-sm text-gray-500">选择分类</label>
                <!-- fllx=1 下拉框模式 -->
                <div v-if="categoryType === 1">
                  <Select
                    :value="activeCateId || undefined"
                    placeholder="选择分类"
                    allow-clear
                    style="width: 200px"
                    @change="(v: any) => (activeCateId = v ? String(v) : '')"
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
                <!-- fllx=0/2 按钮模式 -->
                <div v-else class="flex flex-wrap gap-2">
                  <Button
                    :type="activeCateId === '' ? 'primary' : 'default'"
                    size="small"
                    @click="activeCateId = ''"
                    >全部课程</Button
                  >
                  <Button
                    :type="activeCateId === 'collect' ? 'primary' : 'default'"
                    size="small"
                    :style="{
                      borderColor: '#eb2f96',
                      color: activeCateId === 'collect' ? '' : '#eb2f96',
                    }"
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
                    :type="
                      activeCateId === String(cat.id) ? 'primary' : 'default'
                    "
                    size="small"
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
                    @click="activeCateId = String(cat.id)"
                    >{{ cat.name }}</Button
                  >
                </div>
              </div>
            </template>

            <!-- 课程选择 + 收藏按钮 -->
            <div class="mb-3 flex items-end gap-2">
              <div class="flex-1">
                <OrderForm />
              </div>
              <Tooltip
                :title="isFavorite ? '取消收藏' : '添加收藏'"
                v-if="selectedClassId"
              >
                <Button
                  :type="isFavorite ? 'primary' : 'default'"
                  danger
                  @click="toggleFavorite"
                  class="mb-[24px]"
                >
                  <template #icon>
                    <HeartFilled v-if="isFavorite" />
                    <HeartOutlined v-else />
                  </template>
                  {{ isFavorite ? '已收藏' : '收藏' }}
                </Button>
              </Tooltip>
            </div>

            <!-- 课程说明 -->
            <Alert
              v-if="selectedClassTips"
              type="warning"
              show-icon
              class="my-3"
            >
              <template #message>
                <div class="text-sm" v-html="selectedClassTips"></div>
              </template>
            </Alert>

            <!-- 下单信息（单行 / 多行） -->
            <div class="mb-4">
              <div class="mb-1 flex items-center justify-between">
                <label class="text-sm font-medium text-gray-700"
                  >下单信息</label
                >
              </div>
              <Input.TextArea
                v-if="isMultiline"
                v-model:value="userInfo"
                :rows="5"
                placeholder="下单格式：&#10;学校 账号 密码 (空格分开)&#10;例如：家里蹲大学 13872325008 123456&#10;多账号换行输入"
                @blur="handleBlurRevise"
              />
              <Input
                v-else
                v-model:value="userInfo"
                placeholder="单行模式：家里蹲大学 13872325008 123456"
                @blur="handleBlurRevise"
                @press-enter="handleQuery"
              />
            </div>

            <!-- 操作按钮 -->
            <div class="flex items-center gap-2">
              <Button
                type="primary"
                class="flex-1 bg-blue-600 hover:bg-blue-500"
                size="large"
                :loading="queryLoading"
                @click="handleQuery"
              >
                <template #icon><SearchOutlined /></template>
                一键查课
              </Button>
              <Tag v-if="xdsmopen" color="blue">扫码下单已开启</Tag>
            </div>
          </Card>
        </Spin>
      </div>

      <!-- 右侧：推荐渠道 -->
      <div class="w-full lg:w-[300px]" style="flex-shrink: 0">
        <Card size="small" style="position: sticky; top: 16px">
          <template #title>
            <div class="flex items-center gap-2">
              <span class="text-base font-semibold">{{
                hasAdminRole ? '推荐渠道编辑' : '推荐渠道'
              }}</span>
            </div>
          </template>
          <template #extra v-if="hasAdminRole">
            <Button
              v-if="!editingChannels"
              type="link"
              size="small"
              @click="startEditChannels"
            >
              <EditOutlined /> 编辑
            </Button>
            <Space v-else>
              <Button
                type="link"
                size="small"
                :loading="savingChannels"
                @click="saveChannels"
              >
                <SaveOutlined /> 保存
              </Button>
              <Button
                type="link"
                size="small"
                danger
                @click="cancelEditChannels"
              >
                <CloseOutlined /> 取消
              </Button>
            </Space>
          </template>

          <!-- 编辑模式 (管理员) -->
          <template v-if="editingChannels && hasAdminRole">
            <div class="space-y-3">
              <div
                v-for="(ch, idx) in editChannelList"
                :key="idx"
                class="rounded-lg border border-gray-200 p-3 dark:border-gray-700"
              >
                <div class="mb-2 flex items-center gap-2">
                  <Input
                    v-model:value="ch.name"
                    placeholder="渠道名称"
                    size="small"
                    class="flex-1"
                  />
                  <Button
                    type="text"
                    danger
                    size="small"
                    @click="removeChannel(idx)"
                  >
                    <DeleteOutlined />
                  </Button>
                </div>
                <Input.TextArea
                  v-model:value="ch.desc"
                  placeholder="渠道说明（选填）"
                  :rows="2"
                  size="small"
                />
              </div>
              <Button type="dashed" block @click="addChannel">
                <PlusOutlined /> 添加渠道
              </Button>
            </div>
          </template>

          <!-- 展示模式 -->
          <template v-else>
            <div v-if="recommendChannels.length > 0" class="space-y-3">
              <div
                v-for="(ch, idx) in recommendChannels"
                :key="idx"
                class="rounded-lg border border-blue-100 bg-blue-50/50 p-3 dark:border-blue-900 dark:bg-blue-950/30"
              >
                <div class="mb-1 flex items-center gap-2">
                  <Tag color="blue" class="m-0">{{ idx + 1 }}</Tag>
                  <span
                    class="text-sm font-semibold text-gray-800 dark:text-gray-200"
                    >{{ ch.name }}</span
                  >
                </div>
                <div
                  v-if="ch.desc"
                  class="mt-1 whitespace-pre-wrap text-xs leading-relaxed text-gray-500 dark:text-gray-400"
                >
                  {{ ch.desc }}
                </div>
              </div>
            </div>
            <div
              v-else
              class="flex flex-col items-center py-8 text-gray-400 dark:text-gray-500"
            >
              <span class="text-sm">暂无推荐渠道</span>
              <span v-if="hasAdminRole" class="mt-1 text-xs"
                >点击右上角编辑按钮添加</span
              >
            </div>
          </template>
        </Card>
      </div>
    </div>

    <!-- 选课弹窗 -->
    <Modal
      v-model:open="courseModalVisible"
      title="📦 选择代刷课程"
      width="900px"
      @ok="handleSubmit"
      :confirmLoading="submitLoading"
      okText="立即提交订单"
      cancelText="取消"
      :okButtonProps="{
        disabled: checkedCourses.length === 0,
        size: 'large',
        style: { backgroundColor: '#16a34a', borderColor: '#16a34a' },
      }"
      :cancelButtonProps="{ size: 'large' }"
      :bodyStyle="{
        maxHeight: '65vh',
        overflowY: 'auto',
        backgroundColor: '#f3f4f6',
        padding: '20px',
      }"
    >
      <div
        v-if="queryLoading"
        class="flex flex-col items-center justify-center rounded-lg bg-white py-16 shadow-sm"
      >
        <Spin size="large" />
        <div class="mt-4 text-base text-gray-500">
          正在疯狂查询中，请稍候...
        </div>
      </div>

      <div
        v-else-if="courseResults.length === 0"
        class="rounded-lg bg-white py-12 shadow-sm"
      >
        <Empty description="暂无查课结果" />
      </div>

      <template v-else>
        <!-- 悬浮提示条 -->
        <div class="sticky top-0 z-10 bg-[#f3f4f6] pb-4">
          <Alert
            :message="`已选择 ${checkedCourses.length} 个课程准备下单`"
            type="info"
            show-icon
            class="border-blue-200 bg-blue-50"
          >
            <template #action v-if="checkedCourses.length > 0">
              <span class="text-sm font-medium text-blue-600"
                >点击底部按钮确认提交</span
              >
            </template>
          </Alert>
        </div>

        <div class="space-y-4">
          <Card
            v-for="(result, index) in courseResults"
            :key="index"
            :bordered="false"
            class="overflow-hidden rounded-lg shadow-sm"
            :bodyStyle="{ padding: 0 }"
          >
            <!-- 账号头部 -->
            <div
              class="flex items-center justify-between border-b border-gray-100 bg-gray-50 px-4 py-3"
            >
              <div class="flex items-center gap-3">
                <div
                  class="flex h-9 w-9 items-center justify-center rounded-full bg-blue-100 text-lg font-bold text-blue-600 shadow-inner"
                >
                  {{
                    result.userName
                      ? result.userName.charAt(0).toUpperCase()
                      : 'U'
                  }}
                </div>
                <div class="flex flex-col leading-tight">
                  <span class="text-sm font-bold text-gray-800">{{
                    result.userName || '未知账号'
                  }}</span>
                  <span class="mt-0.5 text-xs text-gray-400">{{
                    result.userinfo
                  }}</span>
                </div>
              </div>
              <Tag
                :color="
                  result.msg === '查询成功' ||
                  result.msg === '此课程无需查课，直接下单即可'
                    ? 'success'
                    : 'error'
                "
                class="m-0 rounded-md border-0 px-2 py-0.5 shadow-sm"
              >
                {{ result.msg }}
              </Tag>
            </div>

            <!-- 课程列表内容 -->
            <div class="p-4">
              <Table
                v-if="result.data && result.data.length > 0"
                :data-source="result.data"
                :pagination="false"
                :scroll="{ x: 780 }"
                row-key="idx"
                size="middle"
                :row-class-name="
                  (record: CourseItem) => (record.select ? 'bg-blue-50' : '')
                "
                :custom-row="
                  (record: CourseItem) => ({
                    onClick: () => toggleCourse(result, record),
                    style: { cursor: 'pointer', transition: 'all 0.3s ease' },
                  })
                "
                class="overflow-hidden rounded-md border border-gray-100"
              >
                <Table.Column title="" :width="60" align="center">
                  <template #default="{ record }">
                    <Checkbox :checked="record.select" />
                  </template>
                  <template #title>
                    <Checkbox
                      :checked="
                        result.data.length > 0 &&
                        result.data.every((c: CourseItem) => c.select)
                      "
                      :indeterminate="
                        result.data.some((c: CourseItem) => c.select) &&
                        !result.data.every((c: CourseItem) => c.select)
                      "
                      @change="toggleAll(result)"
                    />
                  </template>
                </Table.Column>

                <Table.Column title="课程信息" data-index="name" key="name">
                  <template #default="{ record }">
                    <div
                      class="max-w-[520px] overflow-hidden text-ellipsis whitespace-nowrap py-1 text-sm font-medium text-gray-800"
                    >
                      {{ record.name }}
                    </div>
                  </template>
                </Table.Column>

                <Table.Column
                  title="课程ID"
                  data-index="id"
                  key="id"
                  :width="120"
                  align="center"
                >
                  <template #default="{ record }">
                    <span class="font-mono text-xs text-gray-400">{{
                      record.id
                    }}</span>
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
                <span class="mb-1 text-lg">✅</span>
                <span class="font-medium"
                  >无需选课，此账号已自动添加并可直接下单</span
                >
              </div>
              <div v-else class="py-6">
                <Empty description="暂无课程数据" />
              </div>
            </div>
          </Card>
        </div>
      </template>
    </Modal>
  </Page>
</template>
