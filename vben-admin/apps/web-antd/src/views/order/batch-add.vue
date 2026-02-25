<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Table, Tag, Space, message, Spin, Alert, Statistic, Row, Col,
} from 'ant-design-vue';
import {
  UploadOutlined, DeleteOutlined, SendOutlined,
} from '@ant-design/icons-vue';
import { useVbenForm } from '#/adapter/form';
import {
  getClassListApi,
  getClassCategoriesApi,
  addOrderApi,
  type ClassItem,
  type ClassCategory,
} from '#/api/class';

const classLoading = ref(false);
const classList = ref<ClassItem[]>([]);
const categoryList = ref<ClassCategory[]>([]);
const activeCateId = ref<string>('');
const selectedClassId = ref<number | undefined>(undefined);
const submitLoading = ref(false);

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
  if (!activeCateId.value) return classList.value;
  return classList.value.filter((item) => String(item.fenlei) === activeCateId.value);
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

// ===== Vben 表单 =====
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
    {
      component: 'Textarea',
      fieldName: 'rawText',
      label: '批量下单信息',
      componentProps: {
        rows: 8,
        placeholder: '每行一个账号，格式：\n学校 账号 密码（空格分隔）\n例如：\n家里蹲大学 13800138000 123456\n清华大学 13900139000 654321\n\n也支持无学校格式：\n13800138000 123456',
      },
      formItemClass: 'items-baseline',
    },
  ],
  wrapperClass: 'grid-cols-1',
});

async function loadClassData() {
  classLoading.value = true;
  try {
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

async function handleParse() {
  const values = await batchFormApi.getValues();
  if (!values.rawText?.trim()) {
    message.warning('请粘贴下单信息');
    return;
  }

  const lines = values.rawText
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
  batchFormApi.setFieldValue('rawText', '');
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
    // 批量提交：每个账号作为一条，无课程选择（kcid/kcname留空，由后台处理）
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
        <!-- 分类选择 -->
        <div class="flex flex-wrap gap-2 mb-4">
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
            @click="activeCateId = String(cat.id)"
          >{{ cat.name }}</Button>
        </div>

        <!-- Vben 表单 -->
        <BatchForm />

        <Alert
          v-if="selectedClass"
          type="info"
          show-icon
          class="my-3"
          :message="`单价: ¥${selectedClass.price}（实际价格以服务器计算为准）`"
        />

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
