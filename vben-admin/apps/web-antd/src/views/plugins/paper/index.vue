<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Tabs, TabPane, Card, Form, FormItem, Input, InputNumber, Select, SelectOption,
  Button, Switch, Table, Tag, Pagination, Upload, Spin, Modal, Radio, RadioGroup,
  message, Textarea, Descriptions, DescriptionsItem, Tooltip, Empty, Space, Alert
} from 'ant-design-vue';
import { InboxOutlined, CopyOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import {
  paperPricesApi,
  paperGenerateTitlesApi,
  paperGenerateOutlineApi,
  paperOutlineStatusApi,
  paperOrderSubmitApi,
  paperOrderListApi,
  paperDownloadApi,
  paperTextRewriteStream,
  paperTextRewriteAigcStream,
  paperParaEditStream,
  paperFileDedupApi,
  paperCountWordsApi,
  paperGetTemplatesApi,
  paperSaveTemplateApi,
  paperGenerateTaskApi,
  paperGenerateProposalApi,
  paperUploadCoverApi,
  type PaperPriceInfo,
} from '#/api/plugins/paper';

// ==================== 共享状态 ====================
const activeTab = ref('order');
const prices = ref<PaperPriceInfo | null>(null);
const priceLoading = ref(false);

async function loadPrices() {
  priceLoading.value = true;
  try {
    prices.value = await paperPricesApi();
  } catch { /* ignore */ }
  priceLoading.value = false;
}

onMounted(() => { loadPrices(); });

// ==================== 论文下单 ====================
const goodsTypeOptions = [
  { value: '6000', label: '论文6000字' },
  { value: '8000', label: '论文8000字' },
  { value: '10000', label: '论文10000字' },
  { value: '12000', label: '论文12000字' },
  { value: '15000', label: '论文15000字' },
];

const orderForm = ref({
  shopcode: '',
  title: '',
  studentName: '',
  major: '',
  requires: '',
  ktbg: 0,
  rws: 0,
  jiangchong: 0,
  tigang: '',
});

const orderPrice = computed(() => {
  if (!prices.value || !orderForm.value.shopcode) return 0;
  const key = `price_${orderForm.value.shopcode}` as keyof PaperPriceInfo;
  let total = Number(prices.value[key]) || 0;
  if (orderForm.value.rws) total += prices.value.price_rws || 0;
  if (orderForm.value.ktbg) total += prices.value.price_ktbg || 0;
  if (orderForm.value.jiangchong) total += prices.value.price_jdaigchj || 0;
  return Math.round(total * 100) / 100;
});

const outlineChapters = ref<any[]>([]);
const isOutlineShow = ref(false);
const outlineLoading = ref(false);
const orderSubmitLoading = ref(false);

// 生成标题对话框
const titleDialogVisible = ref(false);
const titleMajor = ref('');
const titleOptions = ref<string[]>([]);
const selectedTitle = ref('');
const titleLoading = ref(false);

function openTitleDialog() { titleDialogVisible.value = true; titleOptions.value = []; selectedTitle.value = ''; }

async function handleGenerateTitle() {
  if (!titleMajor.value) return;
  titleLoading.value = true;
  try {
    const res = await paperGenerateTitlesApi({ direction: titleMajor.value });
    if (res?.code === 200 && res.data) {
      titleOptions.value = Array.isArray(res.data) ? res.data : [res.data];
    } else {
      message.error(res?.msg || '生成失败');
    }
  } catch (e: any) { message.error(e.message || '生成失败'); }
  titleLoading.value = false;
}

function useSelectedTitle() {
  if (selectedTitle.value) {
    orderForm.value.title = selectedTitle.value;
    titleDialogVisible.value = false;
  }
}

async function handleGenerateOutline() {
  if (!orderForm.value.shopcode) { message.warning('请选择商品类型'); return; }
  if (!orderForm.value.title) { message.warning('请输入论文标题'); return; }
  outlineLoading.value = true;
  try {
    const res = await paperGenerateOutlineApi({
      title: orderForm.value.title,
      wordCount: parseInt(orderForm.value.shopcode),
      major: orderForm.value.major,
      customRequirements: orderForm.value.requires,
    });
    if (res?.code === 200 && res.data) {
      // 大纲可能是异步生成的，需要轮询
      if (res.data.orderId) {
        await pollOutlineStatus(res.data.orderId);
      } else if (Array.isArray(res.data)) {
        outlineChapters.value = res.data;
        isOutlineShow.value = true;
      }
    } else {
      message.error(res?.msg || '生成大纲失败');
    }
  } catch (e: any) { message.error(e.message || '生成大纲失败'); }
  outlineLoading.value = false;
}

async function pollOutlineStatus(orderId: string) {
  for (let i = 0; i < 30; i++) {
    await new Promise(r => setTimeout(r, 2000));
    try {
      const res = await paperOutlineStatusApi(orderId);
      if (res?.code === 200 && res.data) {
        if (res.data.status === 'completed' || res.data.chapters) {
          outlineChapters.value = res.data.chapters || res.data;
          isOutlineShow.value = true;
          return;
        }
      }
    } catch { /* continue polling */ }
  }
  message.error('大纲生成超时，请重试');
}

function addChapter() {
  outlineChapters.value.push({
    chapter_title: '新章节', chapter_desc: '',
    sections: [{ section_title: '新小节', section_desc: '', sub_sections: [] }],
  });
}

function removeChapter(idx: number) {
  outlineChapters.value.splice(idx, 1);
}

function addSection(chapterIdx: number) {
  outlineChapters.value[chapterIdx].sections.push({
    section_title: '新小节', section_desc: '', sub_sections: [],
  });
}

function removeSection(chapterIdx: number, sectionIdx: number) {
  if (outlineChapters.value[chapterIdx].sections.length <= 1) {
    message.warning('每个章节至少需要一个小节');
    return;
  }
  outlineChapters.value[chapterIdx].sections.splice(sectionIdx, 1);
}

async function handleOrderSubmit() {
  if (!orderForm.value.shopcode) { message.warning('请选择商品类型'); return; }
  if (!orderForm.value.title) { message.warning('请输入论文标题'); return; }
  if (!isOutlineShow.value || outlineChapters.value.length === 0) {
    message.warning('请先生成大纲');
    return;
  }

  orderSubmitLoading.value = true;
  try {
    const submitData = {
      ...orderForm.value,
      tigang: JSON.stringify(outlineChapters.value),
    };
    const res = await paperOrderSubmitApi(submitData);
    if (res?.code === 200) {
      message.success(res.msg || '下单成功');
      // 重置表单
      orderForm.value = { shopcode: '', title: '', studentName: '', major: '', requires: '', ktbg: 0, rws: 0, jiangchong: 0, tigang: '' };
      outlineChapters.value = [];
      isOutlineShow.value = false;
    } else {
      message.error(res?.msg || '下单失败');
    }
  } catch (e: any) { message.error(e.message || '下单失败'); }
  orderSubmitLoading.value = false;
}

// ==================== 论文降重/AIGC ====================
const dedupMode = ref<'file' | 'text'>('file');
const dedupTextContent = ref('');
const dedupTextCount = ref(0);
const dedupResultContent = ref('');
const dedupStatusText = ref('');
const dedupResultShow = ref(false);
const dedupTextLoading = ref(false);

// 文件降重
const dedupFile = ref<File | null>(null);
const dedupFileName = ref('');
const dedupFileSize = ref('');
const dedupWordCount = ref(0);
const dedupFileInfoShow = ref(false);
const dedupFileAigc = ref(0);
const dedupFileJiangchong = ref(0);
const dedupFileLoading = ref(false);

function handleDedupTextInput() {
  dedupTextCount.value = dedupTextContent.value.length;
}

async function handleDedupFileChange(info: any) {
  const file = info.file?.originFileObj || info.file;
  if (!file) return;
  dedupFile.value = file;
  dedupFileName.value = file.name;
  dedupFileSize.value = (file.size / 1024).toFixed(2) + 'KB';

  // 统计字数
  const formData = new FormData();
  formData.append('file', file);
  try {
    const res = await paperCountWordsApi(formData);
    if (res?.code === 200) {
      dedupWordCount.value = res.data;
      dedupFileInfoShow.value = true;
    }
  } catch (e: any) { message.error('字数统计失败: ' + e.message); }
}

async function handleFileDedupSubmit() {
  if (!dedupFile.value) { message.warning('请上传文件'); return; }
  if (dedupFileAigc.value === 0 && dedupFileJiangchong.value === 0) {
    message.warning('请至少选择一项降重服务');
    return;
  }
  dedupFileLoading.value = true;
  try {
    const formData = new FormData();
    formData.append('file', dedupFile.value);
    formData.append('wordCount', String(dedupWordCount.value));
    formData.append('aigc', String(dedupFileAigc.value));
    formData.append('jiangchong', String(dedupFileJiangchong.value));
    const res = await paperFileDedupApi(formData);
    if (res?.code === 200) {
      message.success(res.msg || '降重任务已提交');
      dedupFileInfoShow.value = false;
    } else {
      message.error(res?.msg || res?.data || '降重失败');
    }
  } catch (e: any) { message.error(e.message || '降重失败'); }
  dedupFileLoading.value = false;
}

async function handleTextDedup(type: 'rewrite' | 'aigc') {
  if (!dedupTextContent.value) { message.warning('请输入内容'); return; }
  dedupTextLoading.value = true;
  dedupResultContent.value = '';
  dedupResultShow.value = true;
  dedupStatusText.value = '正在处理...';

  try {
    const response = type === 'rewrite'
      ? await paperTextRewriteStream(dedupTextContent.value)
      : await paperTextRewriteAigcStream(dedupTextContent.value);

    if (!response.ok) throw new Error(`HTTP错误: ${response.status}`);

    // 尝试检测JSON错误响应
    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      const json = await response.json();
      if (json.code && json.code !== 200) {
        message.error(json.msg || json.data || '处理失败');
        dedupResultShow.value = false;
        dedupTextLoading.value = false;
        return;
      }
    }

    // 处理SSE流式响应
    const reader = response.body?.getReader();
    if (!reader) throw new Error('无法读取响应流');
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value);
      const events = buffer.split('\n\n');
      buffer = events.pop() || '';

      for (const event of events) {
        const lines = event.split('\n');
        const eventData: Record<string, string> = {};
        lines.forEach(line => {
          const colonIdx = line.indexOf(':');
          if (colonIdx > 0) {
            eventData[line.substring(0, colonIdx).trim()] = line.substring(colonIdx + 1).trim();
          }
        });

        if (eventData.event === 'error') {
          const errorMsg = JSON.parse(eventData.data || '""');
          message.error(typeof errorMsg === 'string' ? errorMsg : '处理失败');
          dedupTextLoading.value = false;
          return;
        }
        if (eventData.event === 'chunk') {
          dedupResultContent.value += JSON.parse(eventData.data || '""');
        } else if (eventData.event === 'status') {
          dedupStatusText.value = JSON.parse(eventData.data || '""');
        }
      }
    }
    dedupStatusText.value = '处理完成';
  } catch (e: any) { message.error(e.message || '处理失败'); }
  dedupTextLoading.value = false;
}

function copyDedupResult() {
  navigator.clipboard.writeText(dedupResultContent.value).then(() => message.success('复制成功'));
}

// ==================== 段落修改 ====================
const paraContent = ref('');
const paraYijian = ref('');
const paraContentCount = ref(0);
const paraResultContent = ref('');
const paraStatusText = ref('');
const paraResultShow = ref(false);
const paraLoading = ref(false);

function handleParaInput() { paraContentCount.value = paraContent.value.length; }

async function handleParaSubmit() {
  if (!paraContent.value) { message.warning('请输入内容'); return; }
  if (paraContent.value.length < 100) { message.warning('文本字数不能少于100字'); return; }
  paraLoading.value = true;
  paraResultContent.value = '';
  paraResultShow.value = true;
  paraStatusText.value = '正在处理...';

  try {
    const response = await paperParaEditStream(paraContent.value, paraYijian.value);
    if (!response.ok) throw new Error(`HTTP错误: ${response.status}`);

    const contentType = response.headers.get('content-type') || '';
    if (contentType.includes('application/json')) {
      const json = await response.json();
      if (json.code && json.code !== 200) {
        message.error(json.msg || '处理失败');
        paraResultShow.value = false;
        paraLoading.value = false;
        return;
      }
    }

    const reader = response.body?.getReader();
    if (!reader) throw new Error('无法读取响应流');
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value);
      const events = buffer.split('\n\n');
      buffer = events.pop() || '';

      for (const event of events) {
        const lines = event.split('\n');
        const eventData: Record<string, string> = {};
        lines.forEach(line => {
          const colonIdx = line.indexOf(':');
          if (colonIdx > 0) {
            eventData[line.substring(0, colonIdx).trim()] = line.substring(colonIdx + 1).trim();
          }
        });

        if (eventData.event === 'error') {
          const errorMsg = JSON.parse(eventData.data || '""');
          message.error(typeof errorMsg === 'string' ? errorMsg : '处理失败');
          paraLoading.value = false;
          return;
        }
        if (eventData.event === 'chunk') {
          paraResultContent.value += JSON.parse(eventData.data || '""');
        } else if (eventData.event === 'status') {
          paraStatusText.value = JSON.parse(eventData.data || '""');
        }
      }
    }
    paraStatusText.value = '处理完成';
  } catch (e: any) { message.error(e.message || '处理失败'); }
  paraLoading.value = false;
}

function copyParaResult() {
  navigator.clipboard.writeText(paraResultContent.value).then(() => message.success('复制成功'));
}

// ==================== 论文管理 ====================
const listData = ref<any[]>([]);
const listPage = ref(1);
const listPageSize = ref(20);
const listTotal = ref(0);
const listLoading = ref(false);
const listSearch = ref({ title: '', shopname: '', studentName: '', state: '' });

const listColumns = [
  { title: '商品名称', dataIndex: 'shopname', width: 100 },
  { title: '论文名称', dataIndex: 'title', ellipsis: true },
  { title: '姓名', dataIndex: 'studentName', width: 80 },
  { title: '专业', dataIndex: 'major', width: 80 },
  { title: '降重', dataIndex: 'jiangchong', width: 80 },
  { title: 'AIGC', dataIndex: 'aigc', width: 80 },
  { title: '价格', dataIndex: 'price', width: 80 },
  { title: '状态', dataIndex: 'state', width: 90 },
  { title: '下单时间', dataIndex: 'createTime', width: 170 },
  { title: '操作', key: 'action', width: 200 },
];

async function loadList() {
  listLoading.value = true;
  try {
    const params: Record<string, any> = {
      pageNum: listPage.value,
      pageSize: listPageSize.value,
    };
    Object.entries(listSearch.value).forEach(([k, v]) => { if (v) params[k] = v; });
    const res = await paperOrderListApi(params);
    listData.value = res?.rows ?? [];
    listTotal.value = res?.total ?? 0;
  } catch (e: any) { message.error(e.message || '获取列表失败'); }
  listLoading.value = false;
}

function handleListSearch() { listPage.value = 1; loadList(); }
function handleListReset() {
  listSearch.value = { title: '', shopname: '', studentName: '', state: '' };
  listPage.value = 1;
  loadList();
}

async function handlePaperDownload(row: any) {
  if (!row.url) { message.error('下载地址不存在'); return; }
  try {
    const res = await paperDownloadApi(row.url, '-' + row.title);
    if (res?.code === 200 && res.msg) {
      const link = document.createElement('a');
      link.href = res.msg.replace('http://', window.location.protocol + '//');
      link.style.display = 'none';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } else {
      message.error(res?.msg || '下载失败');
    }
  } catch (e: any) { message.error(e.message || '下载失败'); }
}

async function handleGenerateTask(row: any) {
  Modal.confirm({
    title: '确认', content: '确定要生成任务书吗？',
    onOk: async () => {
      try {
        const res = await paperGenerateTaskApi(row.id);
        if (res?.code === 200) { message.success('生成任务书成功'); loadList(); }
        else { message.error(res?.msg || '生成失败'); }
      } catch (e: any) { message.error(e.message || '生成失败'); }
    },
  });
}

async function handleDownloadTask(row: any) {
  try {
    const res = await paperDownloadApi(row.rws, '任务书-' + row.title);
    if (res?.code === 200 && res.msg) {
      const link = document.createElement('a');
      link.href = res.msg.replace('http://', window.location.protocol + '//');
      link.style.display = 'none';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } else { message.error(res?.msg || '下载失败'); }
  } catch (e: any) { message.error(e.message || '下载失败'); }
}

async function handleGenerateProposal(row: any) {
  Modal.confirm({
    title: '确认', content: '确定要生成开题报告吗？',
    onOk: async () => {
      try {
        const res = await paperGenerateProposalApi(row.id);
        if (res?.code === 200) { message.success('生成开题报告成功'); loadList(); }
        else { message.error(res?.msg || '生成失败'); }
      } catch (e: any) { message.error(e.message || '生成失败'); }
    },
  });
}

async function handleDownloadProposal(row: any) {
  try {
    const res = await paperDownloadApi(row.ktbg, '开题报告-' + row.title);
    if (res?.code === 200 && res.msg) {
      const link = document.createElement('a');
      link.href = res.msg.replace('http://', window.location.protocol + '//');
      link.style.display = 'none';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } else { message.error(res?.msg || '下载失败'); }
  } catch (e: any) { message.error(e.message || '下载失败'); }
}

function getStateTag(state: number) {
  switch (state) {
    case 1: return { text: '正在处理', color: 'orange' };
    case 2: return { text: '处理完成', color: 'green' };
    case 3: return { text: '处理异常', color: 'red' };
    default: return { text: '待处理', color: 'default' };
  }
}

watch(activeTab, (val) => {
  if (val === 'list') loadList();
});
</script>

<template>
  <Page title="智文论文" description="AI论文写作、降重、段落修改">
    <Tabs v-model:activeKey="activeTab">
      <!-- ==================== 论文下单 ==================== -->
      <TabPane key="order" tab="论文下单">
        <div class="flex gap-4 flex-col lg:flex-row">
          <!-- 左侧表单 -->
          <Card title="论文信息" class="lg:w-[400px] flex-shrink-0">
            <Form :labelCol="{ xs: { span: 24 }, sm: { span: 6 }, lg: { span: 6 } }" :wrapperCol="{ xs: { span: 24 }, sm: { span: 18 }, lg: { span: 18 } }">
              <FormItem label="商品类型" required>
                <Select v-model:value="orderForm.shopcode" placeholder="请选择">
                  <SelectOption v-for="o in goodsTypeOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="商品价格">
                <span class="text-red-500 text-xl font-bold">¥{{ orderPrice }}</span>
              </FormItem>
              <FormItem label="论文标题" required>
                <div class="flex gap-2">
                  <Input v-model:value="orderForm.title" placeholder="请输入论文标题" class="flex-1" />
                  <Button type="link" @click="openTitleDialog">生成标题</Button>
                </div>
              </FormItem>
              <FormItem label="姓名">
                <Input v-model:value="orderForm.studentName" placeholder="选填" />
              </FormItem>
              <FormItem label="专业">
                <Input v-model:value="orderForm.major" placeholder="选填" />
              </FormItem>
              <FormItem label="论文要求">
                <Textarea v-model:value="orderForm.requires" :rows="4" placeholder="请输入论文具体要求（选填）" />
              </FormItem>
              <FormItem label="附加服务">
                <div class="space-y-2">
                  <div class="flex items-center justify-between border rounded p-2">
                    <span>任务书</span>
                    <span class="text-orange-500">¥{{ prices?.price_rws ?? '-' }}</span>
                    <Switch v-model:checked="orderForm.rws" :checkedValue="1" :unCheckedValue="0" />
                  </div>
                  <div class="flex items-center justify-between border rounded p-2">
                    <span>开题报告</span>
                    <span class="text-orange-500">¥{{ prices?.price_ktbg ?? '-' }}</span>
                    <Switch v-model:checked="orderForm.ktbg" :checkedValue="1" :unCheckedValue="0" />
                  </div>
                  <div class="flex items-center justify-between border rounded p-2">
                    <span>降低AIGC痕迹</span>
                    <span class="text-orange-500">¥{{ prices?.price_jdaigchj ?? '-' }}</span>
                    <Switch v-model:checked="orderForm.jiangchong" :checkedValue="1" :unCheckedValue="0" />
                  </div>
                </div>
              </FormItem>
              <FormItem :wrapperCol="{ offset: 5 }">
                <Button type="primary" @click="handleGenerateOutline" :loading="outlineLoading" class="mr-2">生成大纲</Button>
                <Button type="primary" @click="handleOrderSubmit" :loading="orderSubmitLoading" :disabled="!isOutlineShow">提交订单</Button>
              </FormItem>
            </Form>
          </Card>

          <!-- 右侧大纲 -->
          <Card title="论文大纲" class="flex-1">
            <template v-if="!isOutlineShow">
              <Empty description="暂无大纲数据，请先填写左侧信息并点击生成大纲" class="mt-10" />
            </template>
            <Spin :spinning="outlineLoading" v-else>
              <div class="mb-4">
                <Alert message="可自由增加、删除章节和小节。调整满意后即可提交订单。" type="info" show-icon />
              </div>
              <div class="text-right mb-3">
                <Button type="primary" size="small" @click="addChapter">
                  <PlusOutlined /> 添加章节
                </Button>
              </div>
              <div class="space-y-4">
                <div v-for="(chapter, ci) in outlineChapters" :key="ci" class="border rounded-md p-4 bg-gray-50/50 dark:bg-[#1f1f1f] dark:border-[#303030]">
                  <div class="flex items-center gap-3 mb-3">
                    <span class="text-lg font-bold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 px-2 py-1 rounded">第{{ ci + 1 }}章</span>
                    <Input v-model:value="chapter.chapter_title" placeholder="章节标题" class="flex-1" />
                    <Button type="text" danger @click="removeChapter(ci)" title="删除章节"><DeleteOutlined /></Button>
                  </div>
                  <div class="mb-4">
                    <Textarea v-model:value="chapter.chapter_desc" placeholder="章节描述或写作要点" :auto-size="{ minRows: 2, maxRows: 4 }" />
                  </div>
                  
                  <div class="pl-6 border-l-2 border-gray-200 dark:border-[#303030] space-y-3">
                    <div v-for="(section, si) in chapter.sections" :key="si" class="bg-white dark:bg-[#141414] p-3 rounded shadow-sm border border-gray-100 dark:border-[#303030]">
                      <div class="flex items-center gap-2 mb-2">
                        <Tag color="blue">{{ ci + 1 }}.{{ si + 1 }}</Tag>
                        <Input v-model:value="section.section_title" placeholder="小节标题" size="small" class="flex-1" />
                        <Button type="text" danger size="small" @click="removeSection(ci, si)" :disabled="chapter.sections.length <= 1" title="删除小节"><DeleteOutlined /></Button>
                      </div>
                      <Textarea v-model:value="section.section_desc" placeholder="小节具体写作要点" :auto-size="{ minRows: 2, maxRows: 4 }" />
                    </div>
                    <div>
                      <Button type="dashed" size="small" @click="addSection(ci)" block>
                        <PlusOutlined /> 添加小节
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </Spin>
          </Card>
        </div>

        <!-- 生成标题对话框 -->
        <Modal v-model:open="titleDialogVisible" title="生成论文标题" :width="600">
          <Input v-model:value="titleMajor" placeholder="请输入专业方向，例如：计算机科学、经济学、教育学等" class="mb-3" />
          <RadioGroup v-if="titleOptions.length" v-model:value="selectedTitle" class="w-full">
            <div v-for="t in titleOptions" :key="t" class="mb-2">
              <Radio :value="t">{{ t }}</Radio>
            </div>
          </RadioGroup>
          <template #footer>
            <Button @click="titleDialogVisible = false">取消</Button>
            <Button type="primary" @click="handleGenerateTitle" :loading="titleLoading" :disabled="!titleMajor">
              {{ titleLoading ? '生成中...' : '生成标题' }}
            </Button>
            <Button type="primary" @click="useSelectedTitle" :disabled="!selectedTitle" class="bg-green-600">使用选中标题</Button>
          </template>
        </Modal>
      </TabPane>

      <!-- ==================== 论文降重/AIGC ==================== -->
      <TabPane key="dedup" tab="论文降重/AIGC">
        <Card>
          <template #title>
            <div class="flex justify-between items-center">
              <span>论文降重服务</span>
              <RadioGroup v-model:value="dedupMode" buttonStyle="solid" size="small">
                <Radio.Button value="file">文件降重</Radio.Button>
                <Radio.Button value="text">文本降重</Radio.Button>
              </RadioGroup>
              <div class="space-x-2">
                <Tag>降重单价：{{ prices?.price_jcl ?? '-' }}元/千字</Tag>
                <Tag>AI降重单价：{{ prices?.price_jdaigcl ?? '-' }}元/千字</Tag>
              </div>
            </div>
          </template>

          <!-- 文件降重 -->
          <div v-if="dedupMode === 'file'">
            <Upload.Dragger
              accept=".docx"
              :maxCount="1"
              :customRequest="() => {}"
              @change="handleDedupFileChange"
              :showUploadList="false"
            >
              <p class="text-4xl text-gray-400"><InboxOutlined /></p>
              <p>将文件拖到此处，或<span class="text-blue-500">点击上传</span></p>
              <p class="text-gray-400 text-xs mt-2">只能上传docx格式文件，不超过20MB</p>
            </Upload.Dragger>

            <div v-if="dedupFileInfoShow" class="mt-4">
              <Descriptions title="文件信息" :column="1" bordered size="small">
                <DescriptionsItem label="文件名称">{{ dedupFileName }}</DescriptionsItem>
                <DescriptionsItem label="文件大小">{{ dedupFileSize }}</DescriptionsItem>
                <DescriptionsItem label="字数统计">{{ dedupWordCount }}</DescriptionsItem>
              </Descriptions>
              <div class="mt-3 space-y-2">
                <div class="flex items-center justify-between border rounded p-2">
                  <span>论文降重</span>
                  <Tag>{{ prices?.price_jcl ?? '-' }}元/千字</Tag>
                  <Switch v-model:checked="dedupFileJiangchong" :checkedValue="1" :unCheckedValue="0" />
                </div>
                <div class="flex items-center justify-between border rounded p-2">
                  <span>降低AIGC痕迹</span>
                  <Tag>{{ prices?.price_jdaigcl ?? '-' }}元/千字</Tag>
                  <Switch v-model:checked="dedupFileAigc" :checkedValue="1" :unCheckedValue="0" />
                </div>
              </div>
              <div class="text-center mt-4">
                <Button type="primary" @click="handleFileDedupSubmit" :loading="dedupFileLoading"
                  :disabled="dedupFileAigc === 0 && dedupFileJiangchong === 0">提交降重</Button>
              </div>
            </div>
          </div>

          <!-- 文本降重 -->
          <div v-if="dedupMode === 'text'">
            <Form :labelCol="{ xs: { span: 24 }, sm: { span: 3 }, lg: { span: 3 } }" :wrapperCol="{ xs: { span: 24 }, sm: { span: 21 }, lg: { span: 21 } }">
              <FormItem label="原文内容">
                <Textarea v-model:value="dedupTextContent" :auto-size="{ minRows: 6, maxRows: 12 }" @input="handleDedupTextInput"
                  placeholder="粘贴论文内容，建议250字以上" />
              </FormItem>
              <FormItem :wrapperCol="{ xs: { span: 24 }, sm: { offset: 3, span: 21 } }">
                <Tag>字数统计：{{ dedupTextCount }}字</Tag>
              </FormItem>
              <FormItem :wrapperCol="{ xs: { span: 24 }, sm: { offset: 3, span: 21 } }">
                <Space wrap>
                  <Button type="primary" @click="handleTextDedup('rewrite')" :loading="dedupTextLoading" :disabled="!dedupTextContent">降低重复率</Button>
                  <Button style="background-color: #10b981; color: white; border: none;" @click="handleTextDedup('aigc')" :loading="dedupTextLoading" :disabled="!dedupTextContent">降低AIGC率</Button>
                  <Button @click="dedupTextContent = ''; dedupTextCount = 0" :disabled="!dedupTextContent">清空内容</Button>
                </Space>
              </FormItem>
            </Form>
          </div>

          <!-- 结果展示 -->
          <div v-if="dedupResultShow" class="mt-6 border-t dark:border-[#303030] pt-4">
            <div class="flex justify-between items-center mb-3">
              <div class="flex items-center gap-2">
                <span class="font-bold text-lg">处理结果</span>
                <Tag :color="dedupStatusText === '处理完成' ? 'green' : 'orange'">{{ dedupStatusText }}</Tag>
              </div>
              <Button type="primary" ghost size="small" @click="copyDedupResult"><CopyOutlined /> 复制结果</Button>
            </div>
            <div class="bg-gray-50 dark:bg-[#141414] rounded-md p-4 min-h-[150px] border dark:border-[#303030] relative">
              <Spin :spinning="dedupStatusText === '正在处理...' && !dedupResultContent">
                <div class="whitespace-pre-wrap leading-relaxed">{{ dedupResultContent || '等待处理结果...' }}</div>
              </Spin>
            </div>
          </div>
        </Card>
      </TabPane>

      <!-- ==================== 段落修改 ==================== -->
      <TabPane key="para" tab="段落修改">
        <Card>
          <template #title>
            <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2">
              <span>文本段落修改</span>
              <Tag color="blue" class="w-fit">修改单价：{{ prices?.price_xgdl ?? '-' }}元/千字</Tag>
            </div>
          </template>
          <Form :labelCol="{ xs: { span: 24 }, sm: { span: 3 }, lg: { span: 3 } }" :wrapperCol="{ xs: { span: 24 }, sm: { span: 21 }, lg: { span: 21 } }">
            <FormItem label="原文内容">
              <Textarea v-model:value="paraContent" :auto-size="{ minRows: 5, maxRows: 10 }" @input="handleParaInput" placeholder="请输入需要修改的原文内容" />
            </FormItem>
            <FormItem label="修改意见">
              <Textarea v-model:value="paraYijian" :auto-size="{ minRows: 3, maxRows: 6 }" placeholder="请输入具体的修改建议（选填）" />
            </FormItem>
            <FormItem :wrapperCol="{ xs: { span: 24 }, sm: { offset: 3, span: 21 } }">
              <Tag>字数统计：<span :class="paraContentCount < 100 ? 'text-red-500' : 'text-green-500'">{{ paraContentCount }}</span>字</Tag>
              <Tag color="error" v-if="paraContentCount > 0 && paraContentCount < 100">文本字数不能少于100字</Tag>
            </FormItem>
            <FormItem :wrapperCol="{ xs: { span: 24 }, sm: { offset: 3, span: 21 } }">
              <Space wrap>
                <Button type="primary" @click="handleParaSubmit" :loading="paraLoading" :disabled="!paraContent || paraContentCount < 100">提交修改</Button>
                <Button @click="paraContent = ''; paraYijian = ''; paraContentCount = 0" :disabled="!paraContent && !paraYijian">清空</Button>
              </Space>
            </FormItem>
          </Form>

          <div v-if="paraResultShow" class="mt-6 border-t dark:border-[#303030] pt-4">
            <div class="flex justify-between items-center mb-3">
              <div class="flex items-center gap-2">
                <span class="font-bold text-lg">修改结果</span>
                <Tag :color="paraStatusText === '处理完成' ? 'green' : 'orange'">{{ paraStatusText }}</Tag>
              </div>
              <Button type="primary" ghost size="small" @click="copyParaResult"><CopyOutlined /> 复制结果</Button>
            </div>
            <div class="bg-gray-50 dark:bg-[#141414] rounded-md p-4 min-h-[150px] border dark:border-[#303030] relative">
              <Spin :spinning="paraStatusText === '正在处理...' && !paraResultContent">
                <div class="whitespace-pre-wrap leading-relaxed">{{ paraResultContent || '等待处理结果...' }}</div>
              </Spin>
            </div>
          </div>
        </Card>
      </TabPane>

      <!-- ==================== 论文管理 ==================== -->
      <TabPane key="list" tab="论文管理">
        <Card>
          <div class="flex gap-4 mb-5 flex-wrap items-end bg-gray-50 dark:bg-[#1f1f1f] p-4 rounded-md border border-gray-100 dark:border-dark-400">
            <div class="w-full sm:w-[160px]">
              <div class="text-xs text-gray-500 mb-1">论文名称</div>
              <Input v-model:value="listSearch.title" placeholder="模糊搜索" size="middle" class="w-full" allowClear />
            </div>
            <div class="w-full sm:w-[140px]">
              <div class="text-xs text-gray-500 mb-1">商品名称</div>
              <Input v-model:value="listSearch.shopname" placeholder="商品名称" size="middle" class="w-full" allowClear />
            </div>
            <div class="w-full sm:w-[140px]">
              <div class="text-xs text-gray-500 mb-1">学生姓名</div>
              <Input v-model:value="listSearch.studentName" placeholder="学生姓名" size="middle" class="w-full" allowClear />
            </div>
            <div class="w-full sm:w-[140px]">
              <div class="text-xs text-gray-500 mb-1">状态</div>
              <Select v-model:value="listSearch.state" placeholder="全部状态" size="middle" class="w-full" allowClear>
                <SelectOption value="">全部状态</SelectOption>
                <SelectOption value="0">待处理</SelectOption>
                <SelectOption value="1">正在处理</SelectOption>
                <SelectOption value="2">处理完成</SelectOption>
                <SelectOption value="3">处理异常</SelectOption>
              </Select>
            </div>
            <div class="w-full sm:w-auto flex gap-2">
              <Button type="primary" @click="handleListSearch" class="flex-1 sm:flex-none">搜索</Button>
              <Button @click="handleListReset" class="flex-1 sm:flex-none">重置</Button>
            </div>
          </div>

          <Table :dataSource="listData" :columns="listColumns" :loading="listLoading"
            :pagination="false" rowKey="id" size="middle" :scroll="{ x: 1200 }"
            class="border rounded-md">
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'title'">
                <Tooltip :title="record.title">
                  <div class="truncate max-w-[200px] font-medium text-blue-600">{{ record.title }}</div>
                </Tooltip>
              </template>
              <template v-if="column.dataIndex === 'jiangchong'">
                <Tag :color="record.jiangchong === 1 ? 'red' : 'default'">{{ record.jiangchong === 1 ? '需降重' : '无需降重' }}</Tag>
              </template>
              <template v-if="column.dataIndex === 'aigc'">
                <Tag :color="record.aigc === 1 ? 'red' : 'default'">{{ record.aigc === 1 ? '需降AIGC' : '无需降AIGC' }}</Tag>
              </template>
              <template v-if="column.dataIndex === 'price'">
                <span class="text-orange-500 font-bold">¥{{ record.price }}</span>
              </template>
              <template v-if="column.dataIndex === 'state'">
                <Tag :color="getStateTag(record.state).color">{{ getStateTag(record.state).text }}</Tag>
              </template>
              <template v-if="column.key === 'action'">
                <div class="space-x-2">
                  <Button type="primary" ghost size="small" @click="handlePaperDownload(record)">下载</Button>
                  <Button type="default" size="small" v-if="!isNaN(record.shopcode)"
                    @click="record.rws ? handleDownloadTask(record) : handleGenerateTask(record)">
                    {{ record.rws ? '下载任务书' : '生成任务书' }}
                  </Button>
                  <Button type="default" size="small" v-if="!isNaN(record.shopcode)"
                    @click="record.ktbg ? handleDownloadProposal(record) : handleGenerateProposal(record)">
                    {{ record.ktbg ? '下载开题报告' : '生成开题报告' }}
                  </Button>
                </div>
              </template>
            </template>
          </Table>

          <div class="mt-4 flex justify-end">
            <Pagination v-model:current="listPage" :total="listTotal" :pageSize="listPageSize"
              @change="(p: number) => { listPage = p; loadList(); }" :showSizeChanger="false" :showTotal="(total: number) => `共 ${total} 条`" />
          </div>
        </Card>
      </TabPane>
    </Tabs>
  </Page>
</template>

<style scoped>
.truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
