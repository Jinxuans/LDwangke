<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Select, SelectOption,
  Table, Checkbox, Tag, Space, Alert, Spin, message,
} from 'ant-design-vue';
import { SearchOutlined, CheckOutlined } from '@ant-design/icons-vue';
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

// 课程数据
const classLoading = ref(false);
const classList = ref<ClassItem[]>([]);
const categoryList = ref<ClassCategory[]>([]);
const activeCateId = ref<string>('');
const selectedClassTips = ref('');
const showCategory = ref(true);
const categoryType = ref(0);
const xdsmopen = ref(false);
const queryLoading = ref(false);
const submitLoading = ref(false);

// 查课结果
const courseResults = ref<CourseQueryResult[]>([]);
const checkedCourses = ref<Array<{ userinfo: string; userName: string; data: CourseItem }>>([]);

// 按分类过滤的课程列表
const filteredClassList = computed(() => {
  if (!activeCateId.value) return classList.value;
  return classList.value.filter((item) => String(item.fenlei) === activeCateId.value);
});

// ===== Vben 表单 =====
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
          const cls = classList.value.find(item => item.cid === val);
          selectedClassTips.value = cls?.content || '';
        },
      }),
    },
    {
      component: 'Textarea',
      fieldName: 'userInfo',
      label: '下单信息',
      componentProps: {
        rows: 5,
        placeholder: '下单格式：\n学校 账号 密码 (空格分开)\n例如：家里蹲大学 13872325008 123456\n多账号换行输入',
      },
      formItemClass: 'items-baseline',
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
    const classes = classesRaw;
    const categories = categoriesRaw;
    classList.value = Array.isArray(classes) ? classes : [];
    categoryList.value = Array.isArray(categories) ? categories : [];
  } catch (e) {
    console.error('加载课程失败:', e);
  } finally {
    classLoading.value = false;
  }
}

// 查课
async function handleQuery() {
  const values = await orderFormApi.getValues();
  if (!values.classId) {
    message.warning('请先选择课程');
    return;
  }
  if (!values.userInfo?.trim()) {
    message.warning('请填写下单信息');
    return;
  }

  const lines = values.userInfo
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((l: string) => l.trim())
    .filter(Boolean);

  courseResults.value = [];
  checkedCourses.value = [];
  queryLoading.value = true;

  try {
    // 并发查课
    const promises = lines.map((line: string) =>
      queryCourseApi(values.classId, line).catch((err: any) => ({
        userinfo: line,
        userName: '',
        msg: err?.message || '查询失败',
        data: [] as CourseItem[],
      })),
    );
    const rawResults = await Promise.all(promises);
    const results = rawResults.map((r: any) => (r?.data && !r.userinfo) ? r.data : r);
    courseResults.value = results.map((r: any) => ({
      ...r,
      data: (r.data || []).map((item: CourseItem, idx: number) => ({
        ...item,
        idx,
        select: false,
      })),
    }));
  } catch (e: any) {
    message.error(e?.message || '查课失败');
  } finally {
    queryLoading.value = false;
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
  // 重建选中列表
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
    orderFormApi.setFieldValue('userInfo', '');
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
        <!-- 分类选择 (受 flkg/fllx 配置控制) -->
        <template v-if="showCategory && categoryList.length > 0">
          <!-- fllx=0 按钮模式 -->
          <div v-if="categoryType === 0" style="display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 16px;">
            <Button
              :type="activeCateId === '' ? 'primary' : 'default'"
              size="small"
              @click="activeCateId = ''"
            >全部课程</Button>
            <Button
              v-for="cat in categoryList"
              :key="cat.id"
              :type="activeCateId === String(cat.id) ? 'primary' : 'default'"
              size="small"
              style="margin: 0;"
              :style="cat.recommend ? { borderColor: '#722ed1', color: activeCateId === String(cat.id) ? '' : '#722ed1', fontWeight: '600' } : {}"
              @click="activeCateId = String(cat.id)"
            >{{ cat.name }}</Button>
          </div>
          <!-- fllx=1 下拉框模式 -->
          <div v-else-if="categoryType === 1" style="margin-bottom: 16px;">
            <Select
              :value="activeCateId || undefined"
              placeholder="选择分类"
              allow-clear
              style="width: 200px"
              @change="(v: any) => activeCateId = v ? String(v) : ''"
            >
              <SelectOption v-for="cat in categoryList" :key="cat.id" :value="String(cat.id)">
                {{ cat.name }}
              </SelectOption>
            </Select>
          </div>
          <!-- fllx=2 单选框模式 -->
          <div v-else style="display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 16px;">
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
        </template>

        <!-- Vben 表单 -->
        <OrderForm />

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
            <span v-if="record.complete" class="text-xs text-green-500 ml-2">[{{ record.complete }}]</span>
          </template>
        </Table.Column>
        <Table.Column title="课程ID" data-index="id" key="id" :width="120" align="center" />
      </Table>

      <div v-else class="text-gray-400 text-center py-4">暂无课程数据</div>
    </Card>
  </Page>
</template>
