<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Input, Switch, Table, Tag, Space, message,
  Spin, Alert, Statistic, Row, Col, Tooltip,
} from 'ant-design-vue';
import {
  UploadOutlined, DeleteOutlined, SendOutlined, HeartFilled,
} from '@ant-design/icons-vue';
import { useVbenForm } from '#/adapter/form';
import {
  getClassListApi,
  getClassCategoriesApi,
  addOrderApi,
  type ClassItem,
  type ClassCategory,
} from '#/api/class';
import { getSiteConfigApi } from '#/api/admin';
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
  } catch (e) {
    console.error('加载课程失败:', e);
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
    .replace(/\r\n/g, '\n')
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
  } catch (e: any) {
    message.error(e?.message || '批量提交失败');
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
  <Page title="批量交单" content-class="p-4">
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
          </div>
        </template>

        <!-- 课程选择 -->
        <BatchForm />

        <Alert
          v-if="selectedClass"
          type="info"
          show-icon
          class="my-3"
          :message="`单价: ¥${selectedClass.price}（实际价格以服务器计算为准）`"
        />

        <!-- 批量下单信息 -->
        <div class="mb-4">
          <label class="block text-sm text-gray-500 mb-1">批量下单信息</label>
          <Input.TextArea
            v-model:value="rawText"
            :rows="8"
            placeholder="每行一个账号，格式：&#10;学校 账号 密码（空格分隔）&#10;例如：&#10;家里蹲大学 13800138000 123456&#10;清华大学 13900139000 654321&#10;&#10;也支持无学校格式：&#10;13800138000 123456"
            @blur="handleBlurRevise"
          />
        </div>

        <Space>
          <Button type="primary" @click="handleParse">
            <template #icon><UploadOutlined /></template>
            解析数据
          </Button>
          <Button @click="clearAll" v-if="parsedLines.length > 0">清空</Button>
        </Space>
      </Card>

      <!-- 解析结果预览 -->
      <Card v-if="parsedLines.length > 0" class="mb-4">
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
          :data-source="parsedLines"
          :columns="tableColumns"
          :pagination="false"
          row-key="key"
          size="small"
          bordered
          :scroll="{ y: 400 }"
        >
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'index'">{{ index + 1 }}</template>
            <template v-if="column.key === 'status'">
              <Tag :color="record.valid ? 'green' : 'red'">
                {{ record.valid ? '有效' : '无效' }}
              </Tag>
            </template>
            <template v-if="column.key === 'action'">
              <Button type="link" danger size="small" @click="removeLine(record.key)">
                <template #icon><DeleteOutlined /></template>
              </Button>
            </template>
          </template>
        </Table>

        <div class="mt-4 flex items-center justify-between">
          <Row :gutter="24">
            <Col>
              <Statistic title="有效条数" :value="validCount" class="mr-8" />
            </Col>
            <Col v-if="selectedClass">
              <Statistic title="预估费用" :value="totalCost" :precision="2" prefix="¥" />
            </Col>
          </Row>

          <Button
            type="primary"
            size="large"
            :loading="submitLoading"
            :disabled="validCount === 0 || !selectedClassId"  
            @click="handleSubmit"
            class="bg-green-600 border-green-600 hover:bg-green-500"
          >
            <template #icon><SendOutlined /></template>
            确认批量提交（{{ validCount }} 条）
          </Button>
        </div>
      </Card>
    </Spin>
  </Page>
</template>
