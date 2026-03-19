<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core';
import { computed, onActivated, onMounted, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';

import {
  DeleteOutlined,
  HeartFilled,
  SendOutlined,
  UploadOutlined,
} from '@ant-design/icons-vue';
import {
  Alert,
  Button,
  Card,
  Col,
  Input,
  message,
  Row,
  Space,
  Select,
  SelectOption,
  Spin,
  Statistic,
  Switch,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { getSiteConfigApi } from '#/api/admin';
import {
  addOrderApi,
  type ClassCategory,
  type ClassItem,
  getClassCategoriesApi,
  getClassListPagedApi,
} from '#/api/class';
import { getFavoritesApi } from '#/api/user-center';
import { aiReviseMultiline } from '#/utils/ai-revise';

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

const aiFlag = ref(loadSwitch('batch_ai_flag', true));
const showCategoryToggle = ref(loadSwitch('batch_show_cate', true));

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
const submitLoading = ref(false);
const showCategory = ref(true);
const categoryType = ref(0);

// 收藏
const favoriteCourses = ref<string[]>([]);

// 用户输入
const rawText = ref('');

function parseCategoryType(raw?: string) {
  const parsed = Number(raw ?? '1');
  return [0, 1, 2].includes(parsed) ? parsed : 1;
}

interface ParsedLine {
  key: number;
  school: string;
  user: string;
  pass: string;
  raw: string;
  valid: boolean;
}

const parsedLines = ref<ParsedLine[]>([]);

const selectedClass = computed(() => {
  if (!selectedClassId.value) return null;
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

const validCount = computed(
  () => parsedLines.value.filter((l) => l.valid).length,
);
const totalCost = computed(() => {
  if (!selectedClass.value) return 0;
  return validCount.value * Number(selectedClass.value.price);
});

// ===== Vben 表单（仅课程选择） =====
const [BatchForm, batchFormApi] = useVbenForm({
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
        onPopupScroll: handleClassPopupScroll,
        onSearch: handleClassSearch,
        placeholder: '请选择课程',
        onChange: (val: number) => {
          selectedClassId.value = val;
          selectedClassCache.value =
            classList.value.find((item) => item.cid === val) || null;
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
  void batchFormApi.setFieldValue('classId', undefined);
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
  } catch (error) {
    console.error('加载课程失败:', error);
  } finally {
    classLoading.value = false;
  }
}

// AI 校正（输入框失焦时）
function handleBlurRevise() {
  if (!aiFlag.value) return;
  rawText.value = aiReviseMultiline(rawText.value);
}

function handleParse() {
  if (!rawText.value?.trim()) {
    message.warning('请粘贴下单信息');
    return;
  }

  // 解析前先执行 AI 校正
  if (aiFlag.value) {
    rawText.value = aiReviseMultiline(rawText.value);
  }

  const lines = rawText.value
    .replaceAll('\r\n', '\n')
    .split('\n')
    .map((l: string) => l.trim())
    .filter(Boolean);

  parsedLines.value = lines.map((line, idx) => {
    const parts = line.split(/\s+/);
    if (parts.length >= 3) {
      return {
        key: idx,
        school: parts[0]!,
        user: parts[1]!,
        pass: parts[2]!,
        raw: line,
        valid: true,
      };
    } else if (parts.length === 2) {
      return {
        key: idx,
        school: '自动识别',
        user: parts[0]!,
        pass: parts[1]!,
        raw: line,
        valid: true,
      };
    }
    return {
      key: idx,
      school: '',
      user: '',
      pass: '',
      raw: line,
      valid: false,
    };
  });
}

function removeLine(key: number) {
  parsedLines.value = parsedLines.value.filter((l) => l.key !== key);
}

function clearAll() {
  parsedLines.value = [];
  rawText.value = '';
}

async function handleSubmit() {
  const values = await batchFormApi.getValues();
  if (!values.classId) {
    message.warning('请先选择课程');
    return;
  }
  const validLines = parsedLines.value.filter((l) => l.valid);
  if (validLines.length === 0) {
    message.warning('没有有效的下单数据');
    return;
  }

  submitLoading.value = true;
  try {
    await addOrderApi({
      cid: values.classId,
      data: validLines.map((l) => ({
        userinfo: `${l.school} ${l.user} ${l.pass}`.trim(),
        userName: l.user,
        data: { id: '', name: selectedClass.value?.name || '', kcjs: '' },
      })),
    });
    message.success(`批量提交成功，共 ${validLines.length} 条`);
    clearAll();
  } catch (error: any) {
    message.error(error?.message || '批量提交失败');
  } finally {
    submitLoading.value = false;
  }
}

onMounted(loadClassData);
onActivated(loadSiteConfig);

watch(activeCateId, async () => {
  classKeyword.value = '';
  resetSelectedClass();
  await refreshClassList();
});

const tableColumns = [
  { title: '#', key: 'index', width: 50, align: 'center' as const },
  { title: '学校', dataIndex: 'school', key: 'school', width: 150 },
  { title: '账号', dataIndex: 'user', key: 'user', width: 200 },
  { title: '密码', dataIndex: 'pass', key: 'pass', width: 150 },
  { title: '状态', key: 'status', width: 80, align: 'center' as const },
  { title: '操作', key: 'action', width: 80, align: 'center' as const },
];
</script>

<template>
  <Page content-class="p-4" title="批量交单">
    <Spin :spinning="classLoading">
      <Card class="mb-4">
        <!-- 顶部开关栏 -->
        <div class="mb-4 flex items-center justify-between">
          <h3 class="m-0 text-base font-semibold">批量交单</h3>
          <Space>
            <Tooltip title="显示/隐藏分类选择">
              <Switch
                v-model:checked="showCategoryToggle"
                checked-children="分类"
                un-checked-children="隐藏"
                @change="(v: any) => saveSwitch('batch_show_cate', v)"
              />
            </Tooltip>
            <Tooltip title="若校正有误，请关闭此功能">
              <Switch
                v-model:checked="aiFlag"
                checked-children="AI校正"
                un-checked-children="关闭"
                @change="(v: any) => saveSwitch('batch_ai_flag', v)"
              />
            </Tooltip>
          </Space>
        </div>

        <!-- 分类选择 -->
        <template
          v-if="showCategory && showCategoryToggle && categoryList.length > 0"
        >
          <div class="mb-4">
            <label class="mb-2 block text-sm text-gray-500">选择分类</label>
            <div v-if="categoryType === 1">
              <Select
                :value="activeCateId || undefined"
                allow-clear
                placeholder="选择分类"
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
            <div v-else class="flex flex-wrap gap-2">
              <Button
                :type="activeCateId === '' ? 'primary' : 'default'"
                size="small"
                @click="activeCateId = ''"
                >全部课程</Button
              >
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
                        color: activeCateId === String(cat.id) ? '' : '#722ed1',
                        fontWeight: '600',
                      }
                    : {}
                "
                :type="activeCateId === String(cat.id) ? 'primary' : 'default'"
                size="small"
                @click="activeCateId = String(cat.id)"
                >{{ cat.name }}</Button
              >
            </div>
          </div>
        </template>

        <!-- 课程选择 -->
        <BatchForm />

        <Alert
          v-if="selectedClass"
          :message="`单价: ¥${selectedClass.price}（实际价格以服务器计算为准）`"
          class="my-3"
          show-icon
          type="info"
        />

        <!-- 批量下单信息 -->
        <div class="mb-4">
          <div class="mb-1 flex items-center justify-between">
            <label class="text-sm font-medium text-gray-700"
              >批量下单信息</label
            >
            <div
              class="flex cursor-help items-center gap-1.5"
              title="开启后自动修正输入格式（如符号和多余空格）"
            >
              <Switch v-model:checked="aiFlag" size="small" />
              <span class="text-xs text-gray-500">AI纠错</span>
            </div>
          </div>
          <Input.TextArea
            v-model:value="rawText"
            :rows="8"
            placeholder="每行一个账号，格式：&#10;学校 账号 密码（空格分隔）&#10;例如：&#10;家里蹲大学 13800138000 123456&#10;清华大学 13900139000 654321&#10;&#10;也支持无学校格式：&#10;13800138000 123456"
            @blur="handleBlurRevise"
          />
        </div>

        <div class="flex items-center gap-2">
          <Button
            class="flex-1 bg-blue-600 hover:bg-blue-500"
            size="large"
            type="primary"
            @click="handleParse"
          >
            <template #icon><UploadOutlined /></template>
            解析数据
          </Button>
          <Button v-if="parsedLines.length > 0" size="large" @click="clearAll"
            >清空</Button
          >
        </div>
      </Card>

      <!-- 解析结果预览 -->
      <Card v-if="parsedLines.length > 0" class="mb-4 shadow-sm" size="small">
        <template #title>
          <Space>
            <span>数据预览</span>
            <Tag color="green">有效: {{ validCount }}</Tag>
            <Tag v-if="parsedLines.length - validCount > 0" color="red">
              无效: {{ parsedLines.length - validCount }}
            </Tag>
          </Space>
        </template>

        <Table
          :columns="tableColumns"
          :data-source="parsedLines"
          :pagination="false"
          :scroll="{ y: 400 }"
          bordered
          row-key="key"
          size="small"
        >
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'index'">{{ index + 1 }}</template>
            <template v-if="column.key === 'status'">
              <Tag :color="record.valid ? 'green' : 'red'">
                {{ record.valid ? '有效' : '无效' }}
              </Tag>
            </template>
            <template v-if="column.key === 'action'">
              <Button
                danger
                size="small"
                type="link"
                @click="removeLine(record.key)"
              >
                <template #icon><DeleteOutlined /></template>
              </Button>
            </template>
          </template>
        </Table>

        <div class="mt-4 flex items-center justify-between">
          <Row :gutter="24">
            <Col>
              <Statistic :value="validCount" class="mr-8" title="有效条数" />
            </Col>
            <Col v-if="selectedClass">
              <Statistic
                :precision="2"
                :value="totalCost"
                prefix="¥"
                title="预估费用"
              />
            </Col>
          </Row>

          <Button
            :disabled="validCount === 0 || !selectedClassId"
            :loading="submitLoading"
            class="border-green-600 bg-green-600 hover:bg-green-500"
            size="large"
            type="primary"
            @click="handleSubmit"
          >
            <template #icon><SendOutlined /></template>
            确认批量提交（{{ validCount }} 条）
          </Button>
        </div>
      </Card>
    </Spin>
  </Page>
</template>
