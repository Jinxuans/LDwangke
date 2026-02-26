<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Select, SelectOption, Input, Switch,
  Table, Checkbox, Tag, Space, Alert, Spin, Tooltip, message,
} from 'ant-design-vue';
import { SearchOutlined, CheckOutlined, HeartOutlined, HeartFilled } from '@ant-design/icons-vue';
import { useVbenForm } from '#/adapter/form';
import {
  getClassListApi,
  getClassCategoriesApi,
  queryCourseApi,
  addOrderApi,
  type ClassItem,
  type ClassCategory,
  type CourseQueryResult,
  type CourseItem,
} from '#/api/class';
import { getSiteConfigApi } from '#/api/admin';
import { getFavoritesApi, addFavoriteApi, removeFavoriteApi } from '#/api/user-center';
import { aiReviseMultiline } from '#/utils/ai-revise';

// ===== 开关状态（持久化到 localStorage） =====
function loadSwitch(key: string, def: boolean): boolean {
  try { return JSON.parse(localStorage.getItem(key) ?? String(def)); } catch { return def; }
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
const selectedClassId = ref<number | undefined>(undefined);
const selectedClassTips = ref('');
const showCategory = ref(true);
const categoryType = ref(0);
const xdsmopen = ref(false);
const queryLoading = ref(false);
const submitLoading = ref(false);

// 收藏
const favoriteCourses = ref<string[]>([]);
const isFavorite = computed(() =>
  selectedClassId.value ? favoriteCourses.value.includes(String(selectedClassId.value)) : false,
);

// 查课结果
const courseResults = ref<CourseQueryResult[]>([]);
const checkedCourses = ref<Array<{ userinfo: string; userName: string; data: CourseItem }>>([]);

// 用户输入
const userInfo = ref('');

// 按分类过滤的课程列表
const filteredClassList = computed(() => {
  let list = classList.value;
  if (activeCateId.value === 'collect') {
    list = list.filter((item) => favoriteCourses.value.includes(String(item.cid)));
  } else if (activeCateId.value) {
    list = list.filter((item) => String(item.fenlei) === activeCateId.value);
  }
  return list;
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
        options: filteredClassList.value.map(item => ({
          label: `${item.name}（¥${item.price}）`,
          value: item.cid,
        })),
        showSearch: true,
        allowClear: true,
        filterOption: (input: string, option: any) =>
          option.label?.toLowerCase().includes(input.toLowerCase()),
        placeholder: '请选择课程',
        onChange: (val: number) => {
          selectedClassId.value = val;
          const cls = classList.value.find(item => item.cid === val);
          selectedClassTips.value = cls?.content || '';
        },
      }),
    },
  ],
  wrapperClass: 'grid-cols-1',
});

// 初始化
async function loadClassData() {
  classLoading.value = true;
  try {
    // 加载分类配置
    try {
      const cfg = await getSiteConfigApi();
      showCategory.value = cfg?.flkg !== '0';
      categoryType.value = Number(cfg?.fllx) || 0;
      xdsmopen.value = cfg?.xdsmopen === '1';
    } catch { /* ignore */ }

    const [classesRaw, categoriesRaw] = await Promise.all([
      getClassListApi(),
      getClassCategoriesApi(),
    ]);
    classList.value = Array.isArray(classesRaw) ? classesRaw : [];
    categoryList.value = Array.isArray(categoriesRaw) ? categoriesRaw : [];

    // 加载收藏
    try {
      const favs = await getFavoritesApi();
      favoriteCourses.value = (Array.isArray(favs) ? favs : []).map(String);
    } catch { /* ignore */ }
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
      favoriteCourses.value = favoriteCourses.value.filter(id => id !== String(selectedClassId.value));
      message.success('已取消收藏');
    } else {
      await addFavoriteApi(selectedClassId.value);
      favoriteCourses.value.push(String(selectedClassId.value));
      message.success('已添加收藏');
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

  // 渐进式加载：每个请求独立处理
  const requests = lines.map((line: string) =>
    queryCourseApi(values.classId, line)
      .then((res: any) => {
        const r = res?.data && !res.userinfo ? res.data : res;
        courseResults.value.push({
          ...r,
          data: (r.data || []).map((item: CourseItem, idx: number) => ({
            ...item,
            idx,
            select: false,
          })),
        });
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

  // 第一个完成后取消加载动画
  Promise.race(requests).finally(() => {
    queryLoading.value = false;
  });

  // 等待全部完成
  await Promise.allSettled(requests);
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
</script>

<template>
  <Page title="查课交单" content-class="p-4">
    <Spin :spinning="classLoading">
      <Card class="mb-4">
        <!-- 顶部开关栏 -->
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-base font-semibold m-0">查课交单</h3>
          <Space>
            <Tooltip title="显示/隐藏分类选择">
              <Switch v-model:checked="showCategoryToggle" checked-children="分类" un-checked-children="隐藏" @change="(v: any) => saveSwitch('order_show_cate', v)" />
            </Tooltip>
            <Tooltip title="切换输入模式">
              <Switch v-model:checked="isMultiline" checked-children="多行" un-checked-children="单行" @change="(v: any) => saveSwitch('order_multiline', v)" />
            </Tooltip>
            <Tooltip title="若校正有误，请关闭此功能">
              <Switch v-model:checked="aiFlag" checked-children="AI校正" un-checked-children="关闭" @change="(v: any) => saveSwitch('order_ai_flag', v)" />
            </Tooltip>
          </Space>
        </div>

        <!-- 分类选择 -->
        <template v-if="showCategory && showCategoryToggle && categoryList.length > 0">
          <div class="mb-4">
            <label class="block text-sm text-gray-500 mb-2">选择分类</label>
            <!-- fllx=0 按钮模式 -->
            <div v-if="categoryType === 0" class="flex flex-wrap gap-2">
              <Button
                :type="activeCateId === '' ? 'primary' : 'default'"
                size="small"
                @click="activeCateId = ''"
              >全部课程</Button>
              <Button
                :type="activeCateId === 'collect' ? 'primary' : 'default'"
                size="small"
                :style="{ borderColor: '#eb2f96', color: activeCateId === 'collect' ? '' : '#eb2f96' }"
                @click="activeCateId = activeCateId === 'collect' ? '' : 'collect'"
              >
                <template #icon><HeartFilled /></template>
                收藏课程
              </Button>
              <Button
                v-for="cat in categoryList"
                :key="cat.id"
                :type="activeCateId === String(cat.id) ? 'primary' : 'default'"
                size="small"
                :style="cat.recommend ? { borderColor: '#722ed1', color: activeCateId === String(cat.id) ? '' : '#722ed1', fontWeight: '600' } : {}"
                @click="activeCateId = String(cat.id)"
              >{{ cat.name }}</Button>
            </div>
            <!-- fllx=1 下拉框模式 -->
            <div v-else-if="categoryType === 1">
              <Select
                :value="activeCateId || undefined"
                placeholder="选择分类"
                allow-clear
                style="width: 200px"
                @change="(v: any) => activeCateId = v ? String(v) : ''"
              >
                <SelectOption value="collect">收藏课程</SelectOption>
                <SelectOption v-for="cat in categoryList" :key="cat.id" :value="String(cat.id)">
                  {{ cat.name }}
                </SelectOption>
              </Select>
            </div>
            <!-- fllx=2 单选框模式 -->
            <div v-else class="flex flex-wrap gap-3">
              <label class="cursor-pointer" @click="activeCateId = activeCateId === 'collect' ? '' : 'collect'">
                <input type="radio" :checked="activeCateId === 'collect'" class="mr-1" />
                收藏课程
              </label>
              <label
                v-for="cat in categoryList"
                :key="cat.id"
                class="cursor-pointer"
                @click="activeCateId = activeCateId === String(cat.id) ? '' : String(cat.id)"
              >
                <input type="radio" :checked="activeCateId === String(cat.id)" class="mr-1" />
                {{ cat.name }}
              </label>
            </div>
          </div>
        </template>

        <!-- 课程选择 + 收藏按钮 -->
        <div class="flex items-end gap-2 mb-3">
          <div class="flex-1">
            <OrderForm />
          </div>
          <Tooltip :title="isFavorite ? '取消收藏' : '添加收藏'" v-if="selectedClassId">
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
          <label class="block text-sm text-gray-500 mb-1">下单信息</label>
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
        <Space>
          <Button type="primary" :loading="queryLoading" @click="handleQuery">
            <template #icon><SearchOutlined /></template>
            立即查询
          </Button>
          <Button
            type="primary"
            :loading="submitLoading"
            :disabled="checkedCourses.length === 0"
            @click="handleSubmit"
            class="bg-green-600 border-green-600 hover:bg-green-500"
          >
            <template #icon><CheckOutlined /></template>
            确认下单（{{ checkedCourses.length }}）
          </Button>
          <Tag v-if="xdsmopen" color="blue">扫码下单已开启</Tag>
        </Space>
      </Card>
    </Spin>

    <!-- 查课结果 -->
    <Card
      v-for="(result, index) in courseResults"
      :key="index"
      class="mb-4"
      size="small"
    >
      <template #title>
        <Space>
          <span class="font-bold">{{ result.userName || '用户' }}</span>
          <span class="text-gray-500 text-sm">{{ result.userinfo }}</span>
          <Tag v-if="result.msg === '查询成功'" color="green">{{ result.msg }}</Tag>
          <Tag v-else color="red">{{ result.msg }}</Tag>
        </Space>
      </template>

      <Table
        v-if="result.data && result.data.length > 0"
        :data-source="result.data"
        :pagination="false"
        row-key="idx"
        size="small"
        bordered
        :row-class-name="(record: CourseItem) => record.select ? 'bg-blue-50' : ''"
        :custom-row="(record: CourseItem) => ({
          onClick: () => toggleCourse(result, record),
          style: { cursor: 'pointer' },
        })"
      >
        <Table.Column title="" :width="50" align="center">
          <template #default="{ record }">
            <Checkbox :checked="record.select" />
          </template>
          <template #title>
            <Checkbox
              :checked="result.data.every((c: CourseItem) => c.select)"
              :indeterminate="result.data.some((c: CourseItem) => c.select) && !result.data.every((c: CourseItem) => c.select)"
              @change="toggleAll(result)"
            />
          </template>
        </Table.Column>
        <Table.Column title="课程名" data-index="name" key="name">
          <template #default="{ record }">
            <span>{{ record.name }}</span>
            <span v-if="record.studyStartTime" class="text-xs text-gray-400 ml-2">[开始：{{ record.studyStartTime }}]</span>
            <span v-if="record.studyEndTime" class="text-xs text-gray-400 ml-2">[结束：{{ record.studyEndTime }}]</span>
            <span v-if="record.examStartTime" class="text-xs text-orange-400 ml-2">[考试开始：{{ record.examStartTime }}]</span>
            <span v-if="record.examEndTime" class="text-xs text-orange-400 ml-2">[考试结束：{{ record.examEndTime }}]</span>
            <span v-if="record.learnStatusName" class="text-xs text-blue-400 ml-2">[学习状态：{{ record.learnStatusName }}]</span>
            <span v-if="record.complete" class="text-xs text-green-500 ml-2">[{{ record.complete }}]</span>
          </template>
        </Table.Column>
        <Table.Column title="课程ID" data-index="id" key="id" :width="120" align="center" />
      </Table>

      <div v-else class="text-gray-400 text-center py-4">暂无课程数据</div>
    </Card>
  </Page>
</template>
