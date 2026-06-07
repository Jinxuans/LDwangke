<template>
  <div class="baitan-page art-full-height">
    <ArtSearchBar
      :model-value="query"
      :items="searchItems"
      :showExpand="false"
      @update:model-value="assignQuery"
      @search="loadOrders(1)"
      @reset="resetQuery"
    />

    <section class="rounded-custom-sm border-full-d bg-box p-5">
      <ArtTableHeader
        :loading="loading || syncLoading || orderInfoLoading"
        layout="refresh,size,fullscreen,settings"
        fullClass="baitan-page"
        @refresh="loadOrders(pagination.page)"
      >
        <template #left>
          <div class="baitan-toolbar">
            <h4 class="baitan-title">摆摊实习打卡</h4>
            <ElButton type="primary" :icon="Plus" @click="openOrderDrawer()">提交订单</ElButton>
            <ElButton plain :icon="Link" @click="openUserPage">{{ uiSettings.user_page_text }}</ElButton>
            <ElButton plain :icon="Bell" :loading="noticeLoading" @click="openNotice(true)">{{ uiSettings.notice_refresh_text }}</ElButton>
            <ElButton plain :icon="Refresh" :loading="loading" @click="loadOrders(pagination.page)">刷新订单</ElButton>
            <ElButton v-if="isAdmin" plain :icon="Setting" @click="openConfig">配置接入</ElButton>
            <ElButton v-if="isAdmin" plain :icon="Refresh" :loading="syncLoading" @click="syncAll">同步订单</ElButton>
          </div>
        </template>
      </ArtTableHeader>

      <div v-if="noticeHtml" class="bt-notice" v-html="noticeHtml"></div>

      <ElTable :data="orders" border v-loading="loading" class="mt-4" row-key="id">
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn label="平台/账号" min-width="190">
          <template #default="{ row }">
            <div class="bt-cell">
              <strong>{{ row.platform_label || platformLabel(row.type) }}</strong>
              <span>{{ row.userName }}</span>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="姓名/学校" min-width="180">
          <template #default="{ row }">
            <div class="bt-cell">
              <strong>{{ row.nikeName || '-' }}</strong>
              <span>{{ row.sid || '-' }}</span>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="周期/报告" min-width="170">
          <template #default="{ row }">
            <div class="bt-cell">
              <span>{{ formatWeeks(row.weeks || row.week) }}</span>
              <span>{{ formatReports(row.reports || row.report) }}</span>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="endDate" label="到期时间" width="120" />
        <ElTableColumn label="金额" width="130" align="right">
          <template #default="{ row }">
            <div class="bt-money">
              <strong>{{ money(row.pre_deduct) }}</strong>
              <span v-if="row.difference !== null" :class="diffClass(row.difference)">{{ signedMoney(row.difference) }}</span>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="statusType(row.status)" effect="plain">{{ statusLabel(row.status) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn v-if="isAdmin" label="用户" width="130">
          <template #default="{ row }">UID {{ row.uid }}<br /><span class="text-xs text-g-500">{{ row.username || '-' }}</span></template>
        </ElTableColumn>
        <ElTableColumn prop="createTime" label="创建时间" width="170" />
        <ElTableColumn label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <span v-if="isRefunded(row)" class="bt-muted-action">已退款</span>
            <template v-else>
              <ElButton link type="primary" @click="openOrderDrawer(row)">编辑</ElButton>
              <ElButton link type="primary" @click="syncOne(row)">同步</ElButton>
              <ElDropdown trigger="click" @command="(cmd: string) => handleCommand(row, cmd)">
                <ElButton link type="primary">更多</ElButton>
                <template #dropdown>
                  <ElDropdownMenu>
                    <ElDropdownItem command="logs">日志</ElDropdownItem>
                    <ElDropdownItem command="addDays">加天数</ElDropdownItem>
                    <ElDropdownItem command="buka">补签</ElDropdownItem>
                    <ElDropdownItem command="source">源台信息</ElDropdownItem>
                    <ElDropdownItem command="delete" divided>退款</ElDropdownItem>
                  </ElDropdownMenu>
                </template>
              </ElDropdown>
            </template>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="mt-4 flex justify-end">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :total="pagination.total"
          :page-sizes="[10, 20, 30, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadOrders"
          @size-change="handleSizeChange"
        />
      </div>
    </section>

    <ElDrawer v-model="orderDrawer" :title="orderForm.id ? '编辑订单' : '提交订单'" size="min(900px, 100vw)" class="baitan-drawer">
      <div class="drawer-body">
        <ElAlert
          v-if="orderInfoLoading"
          title="正在同步源台信息"
          type="info"
          show-icon
          :closable="false"
          class="mb-4"
        />
        <section class="rounded-custom-sm border-full-d bg-box p-4">
          <h3 class="bt-section-title">账号与期限</h3>
          <div class="mt-4 grid gap-4 md:grid-cols-3">
            <div>
              <label>平台</label>
              <ElSelect v-model="orderForm.type" class="w-full" filterable :disabled="Boolean(orderForm.id)" @change="onPlatformChange">
                <ElOption v-for="item in platforms" :key="item.value" :label="`${item.label} · ${money(item.price)}/天`" :value="item.value" />
              </ElSelect>
            </div>
            <div>
              <label>账号</label>
              <ElInput v-model="orderForm.userName" :disabled="Boolean(orderForm.id)" />
            </div>
            <div>
              <label>密码</label>
              <ElInput v-model="orderForm.passWord" show-password />
            </div>
            <div v-if="!orderForm.id">
              <label>到期日期</label>
              <ElDatePicker
                v-model="orderForm.endDate"
                class="w-full"
                value-format="YYYY-MM-DD"
                type="date"
                :disabled-date="disablePastDate"
                @change="updateOrderDays"
              />
            </div>
            <div v-if="!orderForm.id" class="bt-action-field">
              <label>查询</label>
              <ElButton class="w-full" type="primary" :loading="phoneInfoLoading" @click="loadPhoneInfo">获取打卡信息</ElButton>
            </div>
          </div>
        </section>

        <section class="mt-4 rounded-custom-sm border-full-d bg-box p-4">
          <h3 class="bt-section-title">打卡信息</h3>
          <div class="mt-4 grid gap-4 md:grid-cols-3">
            <div>
              <label>姓名</label>
              <ElInput v-model="orderForm.nikeName" />
            </div>
            <div>
              <label>学校</label>
              <ElSelect v-if="currentPlatform?.dict_key" v-model="orderForm.sid" class="w-full" filterable :loading="schoolLoading" @visible-change="(v: boolean) => v && loadSchools()">
                <ElOption v-for="item in schools" :key="schoolValue(item)" :label="schoolLabel(item)" :value="schoolValue(item)" />
              </ElSelect>
              <ElInput v-else v-model="orderForm.sid" />
            </div>
            <div v-if="orderForm.type === '10'">
              <label>邀请码</label>
              <ElInput v-model="orderForm.version" />
            </div>
            <div>
              <label>岗位名称</label>
              <ElInput v-model="orderForm.prof" />
            </div>
            <div>
              <label>公司名称</label>
              <ElInput v-model="orderForm.name" />
            </div>
            <div>
              <label>节假日</label>
              <ElRadioGroup v-model="orderForm.km">
                <ElRadioButton label="0">不打卡</ElRadioButton>
                <ElRadioButton label="1">打卡</ElRadioButton>
              </ElRadioGroup>
            </div>
          </div>
        </section>

        <section class="mt-4 rounded-custom-sm border-full-d bg-box p-4">
          <h3 class="bt-section-title">任务设置</h3>
          <div class="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <label>打卡周期</label>
              <div class="bt-week-grid">
                <button
                  v-for="item in weekOptions"
                  :key="item.value"
                  type="button"
                  class="bt-choice bt-week-choice"
                  :class="{ 'is-active': orderForm.weeks.includes(item.value) }"
                  :aria-pressed="orderForm.weeks.includes(item.value)"
                  @click="toggleWeek(item.value)"
                >
                  <span>周{{ item.label }}</span>
                </button>
              </div>
            </div>
            <div>
              <label>报告类型</label>
              <div class="bt-report-grid">
                <button
                  v-for="item in reportOptions"
                  :key="item.value"
                  type="button"
                  class="bt-choice bt-report-choice"
                  :class="{ 'is-active': orderForm.report.includes(item.value) }"
                  :aria-pressed="orderForm.report.includes(item.value)"
                  @click="toggleReport(item.value)"
                >
                  <span>{{ item.label }}</span>
                </button>
              </div>
            </div>
          </div>
          <div class="mt-4 grid gap-4 md:grid-cols-2" v-if="orderForm.report.includes('4') || orderForm.report.includes('5')">
            <div v-if="orderForm.report.includes('4')">
              <label>周报提交时间</label>
              <ElInputNumber v-model="orderForm.weekNum" class="w-full" :min="1" :max="7" />
            </div>
            <div v-if="orderForm.report.includes('5')">
              <label>月报提交日</label>
              <ElInputNumber v-model="orderForm.monthNum" class="w-full" :min="1" :max="28" />
            </div>
          </div>
        </section>

        <section class="mt-4 rounded-custom-sm border-full-d bg-box p-4">
          <h3 class="bt-section-title">定位与时间</h3>
          <div class="mt-4 grid gap-4 md:grid-cols-[1fr_160px_160px]">
            <div>
              <label>打卡地址</label>
              <ElInput v-model="orderForm.address" />
            </div>
            <div>
              <label>经度</label>
              <ElInput v-model="orderForm.lon" placeholder="108.01111" />
            </div>
            <div>
              <label>纬度</label>
              <ElInput v-model="orderForm.lat" placeholder="30.01111" />
            </div>
          </div>
          <div class="mt-4 grid gap-4 md:grid-cols-3">
            <div>
              <label>省</label>
              <ElInput v-model="orderForm.province" />
            </div>
            <div>
              <label>市</label>
              <ElInput v-model="orderForm.market" />
            </div>
            <div>
              <label>区/县</label>
              <ElInput v-model="orderForm.zone" />
            </div>
          </div>
          <div class="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <label>上班时间</label>
              <ElTimePicker v-model="orderForm.startTime" class="w-full" value-format="HH:mm" format="HH:mm" />
            </div>
            <div>
              <label>下班时间</label>
              <ElTimePicker v-model="orderForm.endTime" class="w-full" value-format="HH:mm" format="HH:mm" />
            </div>
          </div>
        </section>
      </div>
      <template #footer>
        <div class="bt-drawer-footer">
          <span class="bt-footer-estimate">
            预计扣费
            <strong>{{ money(estimatedAmount) }}</strong>
            <em>{{ estimateHint }}</em>
          </span>
          <div class="bt-footer-actions">
            <ElButton @click="orderDrawer = false">取消</ElButton>
            <ElButton type="primary" :loading="saving" @click="saveOrder">{{ orderForm.id ? '修改' : '添加' }}</ElButton>
          </div>
        </div>
      </template>
    </ElDrawer>

    <ElDialog v-model="configVisible" title="摆摊打卡配置" width="980px">
      <ElTabs class="bt-config-tabs">
        <ElTabPane label="上游接入">
          <section class="bt-config-panel">
            <div>
              <label>对接类型</label>
              <ElSegmented v-model="configForm.upstream_protocol" :options="protocolOptions" />
            </div>
            <div>
              <label>{{ upstreamUrlLabel }}</label>
              <ElInput v-model="configForm.upstream_url" :placeholder="upstreamUrlPlaceholder" />
            </div>
            <div v-if="configForm.upstream_protocol === 'source'">
              <label>源台 Token</label>
              <ElInput v-model="configForm.token" show-password placeholder="源台 api.php 中配置的 token" />
            </div>
            <div v-else class="grid gap-4 md:grid-cols-2">
              <div>
                <label>同系统 UID</label>
                <ElInputNumber v-model="configForm.upstream_uid" class="w-full" :min="0" placeholder="对方系统开放 UID" />
              </div>
              <div>
                <label>同系统 Key</label>
                <ElInput v-model="configForm.upstream_key" show-password placeholder="对方系统开放 Key" />
              </div>
            </div>
            <div class="bt-config-grid">
              <div>
                <label>补签单价</label>
                <div class="bt-number-field">
                  <ElInputNumber
                    v-model="configForm.buka_unit_price"
                    class="bt-number-input"
                    controls-position="right"
                    :min="0.01"
                    :precision="2"
                    :step="0.01"
                  />
                  <span>元/次</span>
                </div>
              </div>
              <div>
                <label>同步间隔</label>
                <div class="bt-number-field">
                  <ElInputNumber
                    v-model="configForm.sync_interval"
                    class="bt-number-input"
                    controls-position="right"
                    :min="60"
                    :step="30"
                  />
                  <span>秒</span>
                </div>
              </div>
              <div class="bt-switch-field">
                <label>自动同步</label>
                <ElSwitch v-model="configForm.auto_sync" active-text="开启" inactive-text="关闭" inline-prompt />
              </div>
            </div>
          </section>
        </ElTabPane>
        <ElTabPane label="界面展示">
          <section class="bt-config-panel">
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label>用户页面按钮</label>
                <ElInput v-model="configForm.user_page_text" placeholder="用户页面" />
              </div>
              <div>
                <label>公告刷新按钮</label>
                <ElInput v-model="configForm.notice_refresh_text" placeholder="刷新公告" />
              </div>
            </div>
            <div>
              <label>用户页面地址</label>
              <ElInput v-model="configForm.user_page_url" placeholder="https://phone.mmalbasa.net.eu.org/" />
            </div>
            <div>
              <label>用户页面说明</label>
              <ElInput
                v-model="configForm.user_page_intro"
                type="textarea"
                :rows="3"
                maxlength="200"
                show-word-limit
                placeholder="可发给用户进行图片上传、立即执行、开启暂停。"
              />
            </div>
            <div class="bt-config-notice">
              <div class="bt-config-notice-head">
                <label>本地公告</label>
                <ElSwitch v-model="configForm.notice_enabled" active-text="显示" inactive-text="隐藏" inline-prompt />
              </div>
              <ElInput
                v-model="configForm.notice_content"
                type="textarea"
                :rows="4"
                maxlength="1000"
                show-word-limit
                placeholder="填写后展示在订单列表上方，支持简单 HTML"
              />
            </div>
          </section>
        </ElTabPane>
        <ElTabPane label="平台价格">
          <section class="bt-config-panel">
          <ElTable :data="platforms" height="460" border>
            <ElTableColumn prop="label" label="平台" min-width="150" />
            <ElTableColumn label="基础单价" width="180">
              <template #default="{ row }">
                <ElInputNumber v-model="configForm.platform_prices[row.value]" class="w-full" :min="0.01" :precision="2" :step="0.1" />
              </template>
            </ElTableColumn>
          </ElTable>
          </section>
        </ElTabPane>
      </ElTabs>
      <template #footer>
        <ElButton @click="configVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="configSaving" @click="saveConfig">保存配置</ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="bukaVisible" title="补签设置" width="620px">
      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label>补签类型</label>
          <ElSelect v-model="bukaForm.type" class="w-full">
            <ElOption label="签到" value="1" />
            <ElOption label="日报" value="3" />
            <ElOption label="周报" value="4" />
            <ElOption label="月报" value="5" />
          </ElSelect>
        </div>
        <div>
          <label>账号</label>
          <ElInput v-model="bukaForm.userName" disabled />
        </div>
        <div>
          <label>开始日期</label>
          <ElDatePicker v-model="bukaForm.startDate" class="w-full" type="date" value-format="YYYY-MM-DD" :disabled-date="disableFutureDate" @change="refreshBukaEstimate" />
        </div>
        <div>
          <label>结束日期</label>
          <ElDatePicker v-model="bukaForm.endDate" class="w-full" type="date" value-format="YYYY-MM-DD" :disabled-date="disableInvalidBukaEndDate" @change="refreshBukaEstimate" />
        </div>
      </div>
      <div v-if="bukaEstimate" class="bt-summary mt-4">
        <span>{{ bukaEstimate.units }} {{ bukaEstimate.unitLabel }}</span>
        <strong>{{ money(bukaEstimate.money) }}</strong>
      </div>
      <div v-else class="bt-summary bt-summary-muted mt-4">
        <span>选择日期后显示预计扣费</span>
        <strong>{{ money(0) }}</strong>
      </div>
      <template #footer>
        <ElButton @click="bukaVisible = false">取消</ElButton>
        <ElButton :loading="bukaEstimating" @click="() => loadBukaEstimate()">估算</ElButton>
        <ElButton type="primary" :loading="bukaSaving" @click="submitBuka">提交补签</ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="detailDialog.visible" :title="detailDialog.title" width="760px">
      <div class="bt-detail-grid">
        <div v-for="item in detailDialog.items" :key="item.label" class="bt-detail-item">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </div>
      <ElCollapse v-if="detailDialog.raw" class="mt-4">
        <ElCollapseItem title="原始数据" name="raw">
          <pre class="bt-json">{{ detailDialog.raw }}</pre>
        </ElCollapseItem>
      </ElCollapse>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { Bell, Link, Plus, Refresh, Setting } from '@element-plus/icons-vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import {
    addBaitanDays,
    createBaitanOrder,
    editBaitanOrder,
    estimateBaitanBuka,
    fetchBaitanConfig,
    fetchBaitanLogs,
    fetchBaitanNotice,
    fetchBaitanOrders,
    fetchBaitanPhoneInfo,
    fetchBaitanPlatforms,
    fetchBaitanSchools,
    fetchBaitanUISettings,
    queryBaitanSourceOrder,
    refundBaitanOrder,
    saveBaitanConfig,
    submitBaitanBuka,
    syncBaitanOrder,
    syncBaitanOrders,
    type BaitanBukaEstimate,
    type BaitanConfig,
    type BaitanNotice,
    type BaitanOrder,
    type BaitanOrderForm,
    type BaitanPlatform,
    type BaitanUISettings
  } from '@/api/legacy/plugin-baitan'

  defineOptions({ name: 'BaitanIndexPage' })

  const userStore = useUserStore()
  const isAdmin = computed(() => {
    const info = userStore.info as any
    const uid = Number(info?.uid || info?.userId || 0)
    const role = String(info?.role || info?.roleCode || '').toLowerCase()
    const roles = Array.isArray(info?.roles) ? info.roles.map((item: string) => item.toLowerCase()) : []
    return uid === 1 || role === 'admin' || role === 'super' || roles.includes('r_admin') || roles.includes('r_super')
  })

  const platforms = ref<BaitanPlatform[]>([])
  const orders = ref<BaitanOrder[]>([])
  const schools = ref<any[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const syncLoading = ref(false)
  const schoolLoading = ref(false)
  const noticeLoading = ref(false)
  const noticeHtml = ref('')
  const phoneInfoLoading = ref(false)
  const orderInfoLoading = ref(false)
  const configSaving = ref(false)
  const bukaEstimating = ref(false)
  const bukaSaving = ref(false)
  const orderDrawer = ref(false)
  const configVisible = ref(false)
  const bukaVisible = ref(false)
  const bukaEstimate = ref<BaitanBukaEstimate | null>(null)
  const detailDialog = reactive({ items: [] as Array<{ label: string; value: string }>, raw: '', title: '', visible: false })
  const pagination = reactive({ page: 1, limit: 20, total: 0 })
  const query = reactive({ keyword: '', search: '', status: '' })

  const weekOptions = [
    { label: '一', value: '1' }, { label: '二', value: '2' }, { label: '三', value: '3' },
    { label: '四', value: '4' }, { label: '五', value: '5' }, { label: '六', value: '6' }, { label: '日', value: '7' }
  ]
  const reportOptions = [
    { label: '日报', value: '3' }, { label: '周报', value: '4' }, { label: '月报', value: '5' }
  ]
  const protocolOptions = [
    { label: '源台对接', value: 'source' },
    { label: '同系统对接', value: 'same_system' }
  ]
  const searchItems = computed(() => [
    { label: '搜索字段', key: 'search', type: 'select', props: { clearable: true, options: [{ label: '综合搜索', value: '' }, { label: '订单ID', value: 'id' }, { label: '姓名', value: 'nikeName' }, { label: '平台', value: 'type' }, { label: '用户UID', value: 'uid' }] } },
    { label: '关键词', key: 'keyword', type: 'input', props: { clearable: true, placeholder: '账号 / 姓名 / 学校' } },
    { label: '状态', key: 'status', type: 'select', props: { clearable: true, options: [{ label: '启用', value: 'active' }, { label: '异常', value: 'error' }, { label: '已结算', value: 'settled' }] } }
  ])

  const defaultOrder = (): BaitanOrderForm => ({
    type: platforms.value[0]?.value || '1',
    userName: '',
    passWord: '',
    nikeName: '',
    sid: '',
    endDate: '',
    days: 1,
    weeks: ['1', '2', '3', '4', '5'],
    report: ['3'],
    address: '',
    lon: '',
    lat: '',
    version: '',
    weekNum: 6,
    monthNum: 25,
    prof: '',
    province: '',
    market: '',
    zone: '',
    startTime: '',
    endTime: '',
    name: '',
    post: '',
    phone: '',
    holidays: '',
    planName: '',
    planId: '',
    planStartDate: '',
    planEndDate: '',
    moduleId: '',
    projectId: '',
    traineeId: '',
    adCode: '',
    other: '',
    sname: '',
    km: '1'
  })
  const defaultConfig = (): BaitanConfig => ({
    upstream_protocol: 'source',
    upstream_url: '',
    upstream_uid: 0,
    upstream_key: '',
    token: '',
    platform_prices: {},
    buka_unit_price: 0.1,
    auto_sync: true,
    sync_interval: 300,
    timeout: 30,
    user_page_url: 'https://phone.mmalbasa.net.eu.org/',
    user_page_text: '用户页面',
    user_page_intro: '可发给用户进行图片上传、立即执行、开启暂停。',
    notice_refresh_text: '刷新公告',
    notice_enabled: false,
    notice_content: ''
  })
  const orderForm = reactive<BaitanOrderForm>(defaultOrder())
  const configForm = reactive<BaitanConfig>(defaultConfig())
  const uiSettings = reactive<BaitanUISettings>({
    user_page_url: 'https://phone.mmalbasa.net.eu.org/',
    user_page_text: '用户页面',
    user_page_intro: '可发给用户进行图片上传、立即执行、开启暂停。',
    notice_refresh_text: '刷新公告'
  })
  const bukaForm = reactive({ userName: '', platformType: '', type: '4', startDate: '', endDate: '' })
  const currentPlatform = computed(() => platforms.value.find((item) => item.value === orderForm.type))
  const upstreamUrlLabel = computed(() => configForm.upstream_protocol === 'same_system' ? '同系统站点地址' : '源台 API 地址')
  const upstreamUrlPlaceholder = computed(() => {
    if (configForm.upstream_protocol === 'same_system') return 'https://example.com'
    return 'https://example.com/prod-api/api/v2'
  })
  const estimatedAmount = computed(() => {
    if (orderForm.id || !orderForm.endDate) return 0
    return Number(((currentPlatform.value?.price || 0) * Number(orderForm.days || 0)).toFixed(2))
  })
  const estimateHint = computed(() => {
    if (orderForm.id) return '编辑订单不重新预估'
    if (!orderForm.endDate) return '选择到期日期后计算'
    return `按剩余天数计算：${Number(orderForm.days || 0)} 天 x ${money(currentPlatform.value?.price || 0)}/天`
  })

  const loadPlatforms = async () => { platforms.value = (await fetchBaitanPlatforms()) || [] }
  const loadUISettings = async () => {
    try {
      Object.assign(uiSettings, await fetchBaitanUISettings())
    } catch {
      // 使用默认界面文案
    }
  }
  const assignQuery = (value: Record<string, string>) => { Object.assign(query, value || {}) }
  const loadOrders = async (page = pagination.page) => {
    loading.value = true; pagination.page = page
    try { const res = await fetchBaitanOrders({ ...query, page, limit: pagination.limit }); orders.value = res?.list || []; pagination.total = Number(res?.total || 0) } finally { loading.value = false }
  }
  const loadConfig = async () => { Object.assign(configForm, defaultConfig(), await fetchBaitanConfig()); for (const p of platforms.value) if (!configForm.platform_prices[p.value]) configForm.platform_prices[p.value] = 1 }
  const openConfig = async () => { await loadConfig(); configVisible.value = true }
  const saveConfig = async () => {
    const message = validateConfigForm()
    if (message) {
      ElMessage.warning(message)
      return
    }
    configSaving.value = true
    try {
      await saveBaitanConfig({ ...configForm })
      ElMessage.success('配置已保存')
      configVisible.value = false
      await loadUISettings()
      await loadPlatforms()
    } finally {
      configSaving.value = false
    }
  }
  const resetQuery = () => { query.keyword = ''; query.search = ''; query.status = ''; loadOrders(1) }
  const handleSizeChange = (size: number) => { pagination.limit = size; loadOrders(1) }
  const toggleListValue = (list: string[], value: string, order: string[]) => {
    const index = list.indexOf(value)
    if (index >= 0) list.splice(index, 1)
    else list.push(value)
    list.sort((a, b) => order.indexOf(a) - order.indexOf(b))
  }
  const toggleWeek = (value: string) => { toggleListValue(orderForm.weeks, value, weekOptions.map((item) => item.value)) }
  const toggleReport = (value: string) => { toggleListValue(orderForm.report, value, reportOptions.map((item) => item.value)) }
  const onPlatformChange = () => { schools.value = [] }
  const updateOrderDays = () => { orderForm.days = countOrderDays(orderForm.endDate) }
  const openOrderDrawer = async (row?: BaitanOrder) => {
    if (row && isRefunded(row)) {
      ElMessage.warning('订单已退款，不能继续操作')
      return
    }
    const base = row ? orderFromRow(row) : {}
    Object.assign(orderForm, defaultOrder(), base)
    orderDrawer.value = true
    if (row) {
      orderInfoLoading.value = true
      try {
        mergeOrderForm(unwrapSourceRow(await queryBaitanSourceOrder(row.id)))
      } catch {
        ElMessage.warning('源台信息获取失败，已使用本地订单信息')
      } finally {
        orderInfoLoading.value = false
      }
    }
    if (currentPlatform.value?.dict_key) await loadSchools()
  }
  const saveOrder = async () => {
    const message = validateOrderForm()
    if (message) {
      ElMessage.warning(message)
      return
    }
    saving.value = true
    try {
      if (orderForm.id) await editBaitanOrder(orderForm.id, { ...orderForm })
      else await createBaitanOrder({ ...orderForm })
      ElMessage.success('保存成功')
      orderDrawer.value = false
      await loadOrders(1)
    } finally {
      saving.value = false
    }
  }
  const loadSchools = async () => { if (!currentPlatform.value?.dict_key) return; schoolLoading.value = true; try { const res = await fetchBaitanSchools({ platform: orderForm.type, dictKey: currentPlatform.value.dict_key }); schools.value = normalizeList(res) } finally { schoolLoading.value = false } }
  const loadPhoneInfo = async () => {
    if (!orderForm.userName || !orderForm.passWord) {
      ElMessage.warning('请先填写账号和密码')
      return
    }
    if (!orderForm.endDate) {
      ElMessage.warning('请先配置到期时间')
      return
    }
    updateOrderDays()
    if (Number(orderForm.days || 0) <= 0) {
      ElMessage.warning('到期时间必须晚于当前日期')
      return
    }
    phoneInfoLoading.value = true
    try {
      mergeOrderForm(unwrapSourceRow(await fetchBaitanPhoneInfo({ ...orderForm })))
      ElMessage.success('打卡信息已获取')
    } finally {
      phoneInfoLoading.value = false
    }
  }
  const openNotice = async (notify = false) => {
    noticeLoading.value = true
    try {
      const res = await fetchBaitanNotice()
      const notice = normalizeNotice(res)
      noticeHtml.value = notice.has_notice ? notice.content : ''
      if (notify) ElMessage.success(noticeHtml.value ? '公告已更新' : '暂无公告')
    } catch (error) {
      if (notify) ElMessage.error('公告获取失败')
    } finally {
      noticeLoading.value = false
    }
  }
  const openUserPage = async () => {
    const url = String(uiSettings.user_page_url || '').trim()
    if (!isValidHttpUrl(url)) {
      ElMessage.warning('用户页面地址未配置或格式不正确')
      return
    }
    try {
      await ElMessageBox.confirm(
        `${uiSettings.user_page_intro}\n\n${url}`,
        uiSettings.user_page_text,
        {
          confirmButtonText: '打开页面',
          cancelButtonText: '取消',
          type: 'info'
        }
      )
      window.open(url, '_blank', 'noopener,noreferrer')
    } catch {
      // 用户取消
    }
  }
  const syncOne = async (row: BaitanOrder) => {
    if (isRefunded(row)) {
      ElMessage.warning('订单已退款，不能继续操作')
      return
    }
    await syncBaitanOrder(row.id); ElMessage.success('同步完成'); await loadOrders(pagination.page)
  }
  const syncAll = async () => { syncLoading.value = true; try { const res = await syncBaitanOrders(100); ElMessage.success(`同步完成，更新 ${res.updated || 0} 条`); await loadOrders(pagination.page) } finally { syncLoading.value = false } }
  const handleCommand = async (row: BaitanOrder, cmd: string) => {
    if (isRefunded(row)) {
      ElMessage.warning('订单已退款，不能继续操作')
      return
    }
    if (cmd === 'logs') return showDetail('订单日志', await fetchBaitanLogs(row.id))
    if (cmd === 'source') return showDetail('源台信息', await queryBaitanSourceOrder(row.id))
    if (cmd === 'addDays') { const { value } = await ElMessageBox.prompt('请输入增加天数', '增加天数', { inputValue: '1' }); await addBaitanDays(row.id, Number(value || 0)); ElMessage.success('增加成功'); return loadOrders(pagination.page) }
    if (cmd === 'delete') { await ElMessageBox.confirm(`确定为账号 ${row.userName} 按剩余天数退款吗？订单记录会保留。`, '退款确认', { type: 'warning' }); await refundBaitanOrder(row.id); ElMessage.success('退款成功'); return loadOrders(pagination.page) }
    if (cmd === 'buka') { bukaForm.userName = row.userName; bukaForm.platformType = row.type; bukaForm.type = '4'; bukaForm.startDate = ''; bukaForm.endDate = ''; bukaEstimate.value = null; bukaVisible.value = true }
  }
  const refreshBukaEstimate = async () => {
    bukaEstimate.value = null
    if (bukaForm.startDate && bukaForm.endDate) await loadBukaEstimate(false)
  }
  const loadBukaEstimate = async (notify = true) => {
    const message = validateBukaForm()
    if (message) {
      bukaEstimate.value = null
      if (notify) ElMessage.warning(message)
      return
    }
    bukaEstimating.value = true
    try { bukaEstimate.value = await estimateBaitanBuka({ ...bukaForm }) } finally { bukaEstimating.value = false }
  }
  const submitBuka = async () => {
    const message = validateBukaForm()
    if (message) {
      ElMessage.warning(message)
      return
    }
    if (!bukaEstimate.value) await loadBukaEstimate()
    if (!bukaEstimate.value || Number(bukaEstimate.value.units || 0) <= 0) {
      ElMessage.warning('补签日期范围无效')
      return
    }
    await ElMessageBox.confirm(
      `确认提交补签并扣除 ${money(bukaEstimate.value.money)}？`,
      '补签确认',
      { type: 'warning', confirmButtonText: '提交补签', cancelButtonText: '取消' }
    )
    bukaSaving.value = true
    try { await submitBaitanBuka({ ...bukaForm }); ElMessage.success('补签已提交'); bukaVisible.value = false } finally { bukaSaving.value = false }
  }

  const normalizeList = (res: any) => Array.isArray(res?.data) ? res.data : Array.isArray(res?.data?.list) ? res.data.list : Array.isArray(res?.list) ? res.list : Array.isArray(res) ? res : []
  const normalizeNotice = (res: BaitanNotice | any): BaitanNotice => {
    const data = res?.data ?? res
    const content = sanitizeNoticeHtml(String(data?.content ?? data?.notice ?? data?.data ?? '').trim())
    return { has_notice: Boolean(data?.has_notice ?? content), content, raw: data?.raw ?? data }
  }
  const sanitizeNoticeHtml = (raw: string) => raw
    .replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
    .replace(/<style[\s\S]*?>[\s\S]*?<\/style>/gi, '')
    .replace(/\son\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '')
    .replace(/href\s*=\s*("|')\s*javascript:[\s\S]*?\1/gi, 'href="#"')
  const validateOrderForm = () => {
    const required: Array<[keyof BaitanOrderForm, string]> = [
      ['type', '请选择打卡平台'],
      ['userName', '请输入账号'],
      ['passWord', '请输入密码'],
      ['prof', '请输入岗位名称'],
      ['address', '请输入打卡地址'],
      ['lon', '请输入经度'],
      ['lat', '请输入纬度']
    ]
    if (!orderForm.id) required.splice(3, 0, ['endDate', '请选择到期日期'])
    const missing = required.find(([key]) => !String(orderForm[key] || '').trim())
    if (missing) return missing[1]
    if (!orderForm.id) {
      updateOrderDays()
      if (Number(orderForm.days || 0) <= 0) return '下单天数必须大于0'
    }
    return ''
  }
  const validateConfigForm = () => {
    if (!String(configForm.upstream_url || '').trim()) return configForm.upstream_protocol === 'same_system' ? '请填写同系统站点地址' : '请填写源台 API 地址'
    if (configForm.upstream_protocol === 'source' && !String(configForm.token || '').trim()) return '请填写源台 Token'
    if (configForm.upstream_protocol === 'same_system') {
      if (!Number(configForm.upstream_uid || 0)) return '请填写同系统 UID'
      if (!String(configForm.upstream_key || '').trim()) return '请填写同系统 Key'
    }
    return ''
  }
  const validateBukaForm = () => {
    if (!bukaForm.userName) return '补签账号不能为空'
    if (!bukaForm.platformType) return '补签平台不能为空'
    if (!bukaForm.type) return '请选择补签类型'
    if (!bukaForm.startDate) return '请选择开始日期'
    if (!bukaForm.endDate) return '请选择结束日期'
    if (new Date(bukaForm.endDate).getTime() < new Date(bukaForm.startDate).getTime()) return '结束日期不能早于开始日期'
    return ''
  }
  const orderFromRow = (row: BaitanOrder): Partial<BaitanOrderForm> => ({ id: row.id, type: row.type, userName: row.userName, passWord: row.passWord, nikeName: row.nikeName, sid: row.sid, endDate: row.endDate, days: 1, weeks: normalizeStringList(row.weeks || row.week), report: normalizeStringList(row.reports || row.report), address: row.address, lon: row.lon, lat: row.lat, version: row.version, weekNum: row.weekNum || 6, monthNum: row.monthNum || 25 })
  const unwrapSourceRow = (payload: any): Record<string, any> => {
    const data = payload?.data ?? payload
    if (Array.isArray(data)) return data[0] || {}
    if (Array.isArray(data?.list)) return data.list[0] || {}
    if (Array.isArray(data?.rows)) return data.rows[0] || {}
    return data && typeof data === 'object' ? data : {}
  }
  const normalizeStringList = (value: any): string[] => {
    const list = Array.isArray(value) ? value : String(value || '').split(',')
    return list.map((item: any) => String(item).trim()).filter(Boolean)
  }
  const mergeOrderForm = (source: Record<string, any>) => {
    if (!source || !Object.keys(source).length) return
    const keep = { id: orderForm.id, endDate: orderForm.endDate, days: orderForm.days }
    Object.assign(orderForm, source)
    if (!orderForm.type && source.platformType) orderForm.type = String(source.platformType)
    if (!orderForm.sid && source.schoolId) orderForm.sid = String(source.schoolId)
    orderForm.weeks = normalizeStringList(source.weeks ?? source.week ?? orderForm.weeks)
    orderForm.report = normalizeStringList(source.report ?? source.reports ?? orderForm.report)
    orderForm.weekNum = Number(source.weekNum || orderForm.weekNum || 6)
    orderForm.monthNum = Number(source.monthNum || orderForm.monthNum || 25)
    orderForm.id = keep.id
    orderForm.endDate = keep.endDate
    orderForm.days = keep.days
  }
  const schoolLabel = (item: any) => {
    const id = String(item?.dictValue || item?.value || item?.id || '')
    const name = String(item?.dictLabel || item?.label || item?.name || item?.schoolName || item?.value || item)
    if (id && name && item?.dictLabel) return orderForm.type === '10' ? `${id}-${name}` : `${id}--${name}`
    return name
  }
  const schoolValue = (item: any) => String(item?.dictValue || item?.value || item?.id || schoolLabel(item))
  const countOrderDays = (endDate: string) => {
    if (!endDate) return 0
    const end = new Date(endDate)
    if (Number.isNaN(end.getTime())) return 0
    return Math.ceil((end.getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  }
  const disablePastDate = (time: Date) => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    return time.getTime() < today.getTime()
  }
  const disableFutureDate = (time: Date) => time.getTime() > Date.now()
  const disableInvalidBukaEndDate = (time: Date) => {
    if (disableFutureDate(time)) return true
    if (!bukaForm.startDate) return false
    const start = new Date(bukaForm.startDate)
    start.setHours(0, 0, 0, 0)
    return time.getTime() < start.getTime()
  }
  const money = (v: unknown) => `¥${Number(v || 0).toFixed(2)}`
  const signedMoney = (v: unknown) => { const n = Number(v || 0); if (n > 0) return `+${money(n)}`; if (n < 0) return `-${money(Math.abs(n))}`; return money(0) }
  const diffClass = (v: unknown) => Number(v || 0) > 0 ? 'text-danger' : Number(v || 0) < 0 ? 'text-success' : 'text-g-500'
  const statusLabel = (v: string) => ({ active: '正常', pending: '待处理', settled: '已结算', error: '异常', refunded: '已退款' })[String(v || '')] || v || '-'
  const statusType = (v: string) => ['active', 'settled'].includes(v) ? 'success' : ['error', 'refunded'].includes(v) ? 'danger' : 'info'
  const isRefunded = (row: BaitanOrder) => row.status === 'refunded' || row.payment_status === 'refunded'
  const platformLabel = (v: string) => platforms.value.find((item) => item.value === v)?.label || v
  const formatWeeks = (v: any) => (Array.isArray(v) ? v : String(v || '').split(',')).filter(Boolean).map((x: string) => `周${weekOptions.find((i) => i.value === x)?.label || x}`).join('、') || '-'
  const formatReports = (v: any) => (Array.isArray(v) ? v : String(v || '').split(',')).filter(Boolean).map((x: string) => reportOptions.find((i) => i.value === x)?.label || x).join('、') || '-'
  const isValidHttpUrl = (raw: string) => {
    try {
      const url = new URL(raw)
      return url.protocol === 'http:' || url.protocol === 'https:'
    } catch {
      return false
    }
  }
  const showDetail = (title: string, data: any) => {
    const row = unwrapSourceRow(data)
    detailDialog.title = title
    detailDialog.items = buildDetailItems(row)
    detailDialog.raw = typeof data === 'string' ? data : JSON.stringify(data, null, 2)
    detailDialog.visible = true
  }
  const buildDetailItems = (row: Record<string, any>) => {
    const fields = [
      ['状态', row.status || row.order_status || row.code],
      ['消息', row.msg || row.message || row.error],
      ['订单ID', row.id || row.order_id || row.sxdkId],
      ['账号', row.userName || row.username || row.phone],
      ['姓名', row.nikeName || row.name],
      ['学校', row.sid || row.schoolName || row.school],
      ['到期时间', row.endDate || row.end_time],
      ['更新时间', row.updateTime || row.updated_at || row.time]
    ]
    const items = fields
      .map(([label, value]) => ({ label: String(label), value: formatDetailValue(value) }))
      .filter((item) => item.value && item.value !== '-')
    return items.length ? items : [{ label: '结果', value: '已返回数据，详情见原始数据' }]
  }
  const formatDetailValue = (value: any) => {
    if (value === undefined || value === null || value === '') return '-'
    if (typeof value === 'object') return JSON.stringify(value)
    return String(value)
  }

  onMounted(async () => { await Promise.all([loadPlatforms(), loadUISettings()]); await Promise.all([loadOrders(1), openNotice(false)]) })
</script>

<style scoped>
  .baitan-page { display: flex; flex-direction: column; gap: 16px; }
  .baitan-toolbar { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; min-height: 34px; }
  .baitan-title { margin: 0 8px 0 0; color: var(--el-text-color-primary); font-size: 16px; font-weight: 600; line-height: 32px; white-space: nowrap; }
  .bt-cell { display: grid; gap: 2px; line-height: 1.45; }
  .bt-cell strong { color: var(--el-text-color-primary); font-weight: 600; }
  .bt-cell span, .bt-money span { color: var(--el-text-color-secondary); font-size: 12px; }
  .bt-money { display: grid; gap: 2px; }
  .bt-money strong { color: var(--el-color-danger); }
  .bt-muted-action { color: var(--el-text-color-secondary); font-size: 13px; }
  .bt-notice { margin-top: 12px; border: 1px solid var(--el-color-warning-light-7); border-radius: 6px; background: var(--el-color-warning-light-9); padding: 10px 12px; color: var(--el-text-color-primary); font-size: 13px; line-height: 1.7; }
  .bt-notice :deep(p) { margin: 0 0 6px; }
  .bt-notice :deep(p:last-child) { margin-bottom: 0; }
  .bt-notice :deep(a) { color: var(--el-color-primary); }
  .drawer-body label, .baitan-page label { display: block; margin-bottom: 8px; color: var(--el-text-color-primary); font-size: 14px; font-weight: 500; }
  .bt-week-grid { display: grid; grid-template-columns: repeat(7, minmax(42px, 1fr)); gap: 8px; }
  .bt-report-grid { display: grid; grid-template-columns: repeat(3, minmax(78px, 1fr)); gap: 8px; }
  .bt-choice {
    appearance: none;
    display: inline-flex;
    min-height: 34px;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
    background: var(--el-fill-color-blank);
    color: var(--el-text-color-regular);
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.2;
    transition: border-color .16s ease, background-color .16s ease, color .16s ease, box-shadow .16s ease;
  }
  .bt-choice:hover { border-color: var(--el-color-primary-light-5); color: var(--el-color-primary); }
  .bt-choice.is-active { border-color: var(--el-color-primary); background: var(--el-color-primary-light-9); color: var(--el-color-primary); box-shadow: inset 0 0 0 1px var(--el-color-primary-light-7); }
  .bt-choice:focus-visible { outline: 2px solid var(--el-color-primary-light-5); outline-offset: 2px; }
  .bt-report-choice { min-height: 38px; }
  .bt-section-title { margin: 0; color: var(--el-text-color-primary); font-size: 15px; font-weight: 700; }
  .bt-config-grid { display: grid; grid-template-columns: minmax(150px, 1fr) minmax(150px, 1fr) minmax(120px, auto); gap: 16px; align-items: end; }
  .bt-config-panel { display: grid; gap: 16px; padding: 4px 0 0; }
  .bt-number-field { display: flex; align-items: center; gap: 8px; }
  .bt-number-field span { flex: none; color: var(--el-text-color-secondary); font-size: 13px; }
  .bt-number-input { width: 148px; }
  .bt-switch-field { min-width: 120px; }
  .bt-switch-field label { margin-bottom: 10px; }
  .bt-config-notice { margin-top: 16px; }
  .bt-config-notice-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
  .bt-config-notice-head label { margin-bottom: 0; }
  .bt-summary { display: flex; align-items: center; justify-content: space-between; border: 1px solid var(--art-border-color); border-radius: 6px; padding: 12px 14px; }
  .bt-summary strong { color: var(--el-color-danger); }
  .bt-summary-muted { color: var(--el-text-color-secondary); }
  .bt-summary-muted strong { color: var(--el-text-color-secondary); }
  .bt-detail-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .bt-detail-item { display: grid; gap: 4px; border: 1px solid var(--el-border-color-lighter); border-radius: 6px; padding: 10px 12px; }
  .bt-detail-item span { color: var(--el-text-color-secondary); font-size: 12px; }
  .bt-detail-item strong { min-width: 0; overflow-wrap: anywhere; color: var(--el-text-color-primary); font-size: 13px; font-weight: 600; }
  .bt-drawer-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; text-align: left; }
  .bt-footer-estimate { display: flex; min-width: 0; align-items: baseline; gap: 8px; color: var(--el-text-color-secondary); font-size: 13px; line-height: 1.5; }
  .bt-footer-estimate strong { flex: none; color: var(--el-color-danger); font-size: 15px; }
  .bt-footer-estimate em { min-width: 0; overflow: hidden; font-style: normal; text-overflow: ellipsis; white-space: nowrap; }
  .bt-footer-actions { display: flex; flex: none; gap: 8px; }
  .bt-json { max-height: 60vh; overflow: auto; margin: 0; border: 1px solid var(--art-border-color); border-radius: 6px; background: var(--art-main-bg-color); padding: 14px; color: var(--el-text-color-regular); font-size: 12px; line-height: 1.7; white-space: pre-wrap; }
  @media (max-width: 640px) {
    .bt-config-grid { grid-template-columns: 1fr; align-items: stretch; }
    .bt-detail-grid { grid-template-columns: 1fr; }
    .bt-number-input { width: 100%; }
    .bt-week-grid { grid-template-columns: repeat(4, minmax(54px, 1fr)); }
    .bt-report-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .bt-drawer-footer { align-items: stretch; flex-direction: column; }
    .bt-footer-estimate { display: grid; justify-items: start; gap: 2px; }
    .bt-footer-estimate em { white-space: normal; }
    .bt-footer-actions { justify-content: flex-end; }
  }
  :global(.baitan-drawer .el-drawer__header) { margin-bottom: 0; padding: 18px 20px; border-bottom: 1px solid var(--el-border-color-lighter); }
  :global(.baitan-drawer .el-drawer__body) { padding: 16px; background: var(--art-main-bg-color); }
  :global(.baitan-drawer .el-drawer__footer) { padding: 12px 16px; border-top: 1px solid var(--el-border-color-lighter); }
</style>
