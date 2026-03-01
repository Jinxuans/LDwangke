<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  DeleteOutlined, HeartFilled, SendOutlined, UploadOutlined,
} from '@ant-design/icons-vue';
import {
  Alert, Button, Card, Col, Input, message, Row, Space,
  Spin, Statistic, Switch, Table, Tag, Tooltip,
} from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { getSiteConfigApi } from '#/api/admin';
import {
  addOrderApi,
  type ClassCategory,
  type ClassItem,
  getClassCategoriesApi,
  getClassListApi,
} from '#/api/class';
import { getFavoritesApi } from '#/api/user-center';
import { aiReviseMultiline } from '#/utils/ai-revise';

// ===== 开关状态（持久化到 localStorage） =====
function loadSwitch(key: string, def: boolean): boolean {
  try { return JSON.parse(localStorage.getItem(key) ?? String(def)); } catch { return def; }
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
const selectedClassId = ref<number | undefined>(undefined);
const submitLoading = ref(false);
const showCategory = ref(true);
const categoryType = ref(0);

// 收藏
const favoriteCourses = ref<string[]>([]);

// 用户输入
const rawText = ref('');

interface ParsedLine {
  key: number;
  school: string;
  user: string;
  pass: string;
  raw: string;
  valid: boolean;
}

const parsedLines = ref<ParsedLine[]>([]);

const filteredClassList = computed(() => {
  let list = classList.value;
  if (activeCateId.value === 'collect') {
    list = list.filter((item) => favoriteCourses.value.includes(String(item.cid)));
  } else if (activeCateId.value) {
    list = list.filter((item) => String(item.fenlei) === activeCateId.value);
  }
  return list;
});

const selectedClass = computed(() => {
  if (!selectedClassId.value) return null;
  return classList.value.find((item) => item.cid === selectedClassId.value);
});

const validCount = computed(() => parsedLines.value.filter((l) => l.valid).length);
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
        },
      }),
    },
  ],
  wrapperClass: 'grid-cols-1',
});

async function loadClassData() {
  classLoading.value = true;
  try {
    try {
      const cfg = await getSiteConfigApi();
      showCategory.value = cfg?.flkg !== '0';
      categoryType.value = Number(cfg?.fllx) || 0;
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
      return { key: idx, school: parts[0]!, user: parts[1]!, pass: parts[2]!, raw: line, valid: true };
    } else if (parts.length === 2) {
      return { key: idx, school: '自动识别', user: parts[0]!, pass: parts[1]!, raw: line, valid: true };
    }
    return { key: idx, school: '', user: '', pass: '', raw: line, valid: false };
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
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-base font-semibold m-0">批量交单</h3>
          <Space>
            <Tooltip title="显示/隐藏分类选择">
              <Switch v-model:checked="showCategoryToggle" checked-children="分类" un-checked-children="隐藏" @change="(v: any) => saveSwitch('batch_show_cate', v)" />
            </Tooltip>
            <Tooltip title="若校正有误，请关闭此功能">
              <Switch v-model:checked="aiFlag" checked-children="AI校正" un-checked-children="关闭" @change="(v: any) => saveSwitch('batch_ai_flag', v)" />
            </Tooltip>
          </Space>
        </div>

        <!-- 分类选择 -->
        <template v-if="showCategory && showCategoryToggle && categoryList.length > 0">
          <div class="mb-4">
            <label class="block text-sm text-gray-500 mb-2">选择分类</label>
            <div class="flex flex-wrap gap-2">
              <Button
                :type="activeCateId === '' ? 'primary' : 'default'"
                size="small"
                @click="activeCateId = ''"
              >
全部课程
</Button>
              <Button
                :style="{ borderColor: '#eb2f96', color: activeCateId === 'collect' ? '' : '#eb2f96' }"
                :type="activeCateId === 'collect' ? 'primary' : 'default'"
                size="small"
                @click="activeCateId = activeCateId === 'collect' ? '' : 'collect'"
              >
                <template #icon><HeartFilled /></template>
                收藏课程
              </Button>
              <Button
                v-for="cat in categoryList"
                :key="cat.id"
                :style="cat.recommend ? { borderColor: '#722ed1', color: activeCateId === String(cat.id) ? '' : '#722ed1', fontWeight: '600' } : {}"
                :type="activeCateId === String(cat.id) ? 'primary' : 'default'"
                size="small"
                @click="activeCateId = String(cat.id)"
              >
{{ cat.name }}
</Button>
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
          <div class="flex items-center justify-between mb-1">
            <label class="text-sm font-medium text-gray-700">批量下单信息</label>
            <div class="flex items-center gap-1.5 cursor-help" title="开启后自动修正输入格式（如符号和多余空格）">
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

        <div class="flex gap-2 items-center">
          <Button class="bg-blue-600 hover:bg-blue-500 flex-1" size="large" type="primary" @click="handleParse">
            <template #icon><UploadOutlined /></template>
            解析数据
          </Button>
          <Button v-if="parsedLines.length > 0" size="large" @click="clearAll">清空</Button>
        </div>
      </Card>

      <!-- 解析结果预览 -->
      <Card v-if="parsedLines.length > 0" class="shadow-sm mb-4" size="small">
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
              <Button danger size="small" type="link" @click="removeLine(record.key)">
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
              <Statistic :precision="2" :value="totalCost" prefix="¥" title="预估费用" />
            </Col>
          </Row>

          <Button
            :disabled="validCount === 0 || !selectedClassId"
            :loading="submitLoading"
            class="bg-green-600 border-green-600 hover:bg-green-500"
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
