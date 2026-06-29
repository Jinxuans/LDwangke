<template>
  <div class="admin-sync-monitor-page art-full-height">
    <section v-loading="loading" class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-4 border-b-d pb-4">
        <div class="flex flex-wrap gap-2">
          <ElTag :type="autoStatus.enabled ? 'success' : 'info'" effect="plain">
            自动同步 {{ autoStatus.enabled ? '已开启' : '未开启' }}
          </ElTag>
          <ElTag :type="autoStatus.running ? 'warning' : 'primary'" effect="plain">
            {{ autoStatus.running ? '执行中' : '空闲' }}
          </ElTag>
          <ElTag effect="plain">上次 {{ autoStatus.last_run_time || '暂无' }}</ElTag>
          <ElTag effect="plain">结果 {{ autoStatus.last_result || '暂无' }}</ElTag>
        </div>
        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="loading" @click="refreshAll">刷新数据</ElButton>
          <ElButton plain :loading="saving" @click="saveConfig">保存同步设置</ElButton>
        </div>
      </div>

      <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">监听货源</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ monitoredSuppliers.length }}</p>
          <p class="mt-2 text-sm text-g-500">已选中同步供应商</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">本地商品</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ totalLocalCount }}</p>
          <p class="mt-2 text-sm text-g-500">当前监听范围内</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">在架商品</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ totalActiveCount }}</p>
          <p class="mt-2 text-sm text-g-500">处于可售状态</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">同步日志</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ logPagination.total }}</p>
          <p class="mt-2 text-sm text-g-500">累计变更记录</p>
        </article>
      </div>
    </section>

    <ElTabs v-model="activeTab" class="mt-4">
      <ElTabPane label="同步概况" name="overview">
        <div class="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-4 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">监听货源</h3>
                <p class="mt-1.5 text-sm leading-6 text-g-500">直接预览差异并执行同步。</p>
              </div>
              <ElTag effect="plain">{{ autoStatus.last_result || '暂无结果' }}</ElTag>
            </div>

            <div v-if="monitoredSuppliers.length" class="mt-5 grid gap-4 md:grid-cols-2">
              <article
                v-for="item in monitoredSuppliers"
                :key="item.hid"
                class="rounded-custom-sm border-full-d p-4"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate text-base font-semibold text-g-900">{{ item.name }}</p>
                    <p class="mt-1 text-sm text-g-500">{{ item.pt_name || item.pt || '-' }}</p>
                  </div>
                  <ElTag :type="item.status === '1' ? 'success' : 'info'" effect="plain">
                    {{ item.status === '1' ? '启用' : '停用' }}
                  </ElTag>
                </div>

                <div class="mt-4 grid grid-cols-3 gap-3">
                  <div class="rounded-lg bg-g-100/60 px-3 py-3 text-center">
                    <p class="text-xs text-g-400">本地</p>
                    <p class="mt-2 text-base font-semibold text-g-900">{{ item.local_count }}</p>
                  </div>
                  <div class="rounded-lg bg-g-100/60 px-3 py-3 text-center">
                    <p class="text-xs text-g-400">在架</p>
                    <p class="mt-2 text-base font-semibold text-g-900">{{ item.active_count }}</p>
                  </div>
                  <div class="rounded-lg bg-g-100/60 px-3 py-3 text-center">
                    <p class="text-xs text-g-400">余额</p>
                    <p class="mt-2 text-base font-semibold text-g-900">{{ item.money || '0' }}</p>
                  </div>
                </div>

                <div class="mt-4">
                  <ElButton size="small" type="primary" plain @click="openPreview(item)">预览差异</ElButton>
                </div>
              </article>
            </div>

            <ElEmpty v-else description="还没有配置监听货源" />
          </section>

          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">自动同步状态</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">按分钟间隔直接执行商品同步。</p>
            </div>

            <div class="mt-5 grid gap-4 sm:grid-cols-2">
              <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                <p class="text-xs font-medium text-g-400">状态</p>
                <p class="mt-2 text-base font-semibold text-g-900">
                  {{ autoStatus.enabled ? (autoStatus.running ? '执行中' : '已开启') : '未开启' }}
                </p>
                <p class="mt-2 text-sm text-g-500">间隔 {{ autoStatus.interval || 0 }} 分钟</p>
              </article>
              <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                <p class="text-xs font-medium text-g-400">累计执行</p>
                <p class="mt-2 text-base font-semibold text-g-900">{{ autoStatus.total_runs || 0 }}</p>
                <p class="mt-2 text-sm text-g-500">下次 {{ autoStatus.next_run_time || '暂无' }}</p>
              </article>
            </div>
          </section>
        </div>
      </ElTabPane>

      <ElTabPane label="同步设置" name="settings">
        <div class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">基础配置</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">维护监听货源、倍率和自动同步开关。</p>
            </div>

            <div class="mt-5 space-y-5">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">监听货源</label>
                <ElSelect
                  v-model="selectedSupplierIds"
                  class="w-full"
                  multiple
                  filterable
                  clearable
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="选择需要监听的货源"
                  @change="handleSupplierChange"
                >
                  <ElOption
                    v-for="item in supplierOptions"
                    :key="item.hid"
                    :label="`${item.name} (HID:${item.hid})`"
                    :value="item.hid"
                  />
                </ElSelect>
              </div>

              <div class="grid gap-4 md:grid-cols-2">
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">同步价格</span>
                    <ElSwitch v-model="syncConfig.sync_price" />
                  </div>
                </article>
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">同步上下架</span>
                    <ElSwitch v-model="syncConfig.sync_status" />
                  </div>
                </article>
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">同步说明</span>
                    <ElSwitch v-model="syncConfig.sync_content" />
                  </div>
                </article>
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">同步名称</span>
                    <ElSwitch v-model="syncConfig.sync_name" />
                  </div>
                </article>
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">允许克隆</span>
                    <ElSwitch v-model="syncConfig.clone_enabled" />
                  </div>
                </article>
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-800">自动同步</span>
                    <ElSwitch v-model="syncConfig.auto_sync_enabled" />
                  </div>
                </article>
              </div>

              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">同步保密价倍率</label>
                  <ElInputNumber v-model="syncConfig.secret_price_rate" class="w-full" :min="0" :step="0.1" :precision="2" />
                  <p class="mt-1.5 text-xs text-g-500">写入商品保密价，不影响用户密价规则</p>
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">自动同步间隔（分钟）</label>
                  <ElInputNumber v-model="syncConfig.auto_sync_interval" class="w-full" :min="5" :max="1440" />
                </div>
              </div>

              <div>
                <h4 class="mb-3 text-sm font-semibold text-g-900">货源倍率</h4>
                <div v-if="selectedSupplierIds.length" class="space-y-3">
                  <div
                    v-for="hid in selectedSupplierIds"
                    :key="hid"
                    class="flex items-center gap-3 rounded-custom-sm border-full-d p-4"
                  >
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm font-medium text-g-900">{{ supplierName(hid) }}</p>
                      <p class="mt-1 text-xs text-g-500">本地售价 = 上游价格 × 倍率</p>
                    </div>
                    <ElInputNumber
                      :model-value="syncConfig.price_rates[String(hid)] || 1"
                      :min="0.01"
                      :max="100"
                      :step="0.1"
                      :precision="2"
                      @update:model-value="updateSupplierRate(hid, Number($event || 1))"
                    />
                  </div>
                </div>
                <ElEmpty v-else description="先选择监听货源" />
              </div>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">扩展规则</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">保留最常用的跳过分类和名称替换。</p>
            </div>

            <div class="mt-5 space-y-5">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">跳过本地分类</label>
                <ElSelect
                  v-model="syncConfig.skip_categories"
                  class="w-full"
                  multiple
                  filterable
                  clearable
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="这些分类不会参与同步"
                >
                  <ElOption
                    v-for="item in categoryOptions"
                    :key="item.id"
                    :label="item.name"
                    :value="String(item.id)"
                  />
                </ElSelect>
              </div>

              <div>
                <div class="mb-3 flex items-center justify-between">
                  <label class="text-sm font-medium text-g-800">名称替换</label>
                  <ElButton text type="primary" @click="addNameReplace">添加规则</ElButton>
                </div>

                <div class="grid gap-3 sm:grid-cols-[1fr_auto_1fr]">
                  <ElInput v-model="replaceDraft.oldValue" placeholder="原文字" />
                  <div class="flex items-center justify-center text-g-400">→</div>
                  <ElInput v-model="replaceDraft.newValue" placeholder="替换为（留空代表删除）" />
                </div>

                <div v-if="replaceEntries.length" class="mt-4 space-y-3">
                  <div
                    v-for="item in replaceEntries"
                    :key="item.oldValue"
                    class="flex items-center gap-3 rounded-custom-sm border-full-d p-4"
                  >
                    <ElTag type="danger" effect="plain">{{ item.oldValue }}</ElTag>
                    <span class="text-g-400">→</span>
                    <ElTag type="success" effect="plain">{{ item.newValue || '(删除)' }}</ElTag>
                    <ElButton text type="danger" @click="removeNameReplace(item.oldValue)">删除</ElButton>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </ElTabPane>

      <ElTabPane label="变更日志" name="logs">
        <ElCard class="art-table-card">
          <ArtTableHeader v-model:columns="logColumnChecks" :loading="logLoading" @refresh="loadLogs" />
          <ArtTable
            :loading="logLoading"
            :data="logs"
            :columns="logColumns"
            :pagination="logPagination"
            @pagination:current-change="handleLogPageChange"
            @pagination:size-change="handleLogSizeChange"
          />
        </ElCard>
      </ElTabPane>

      <ElTabPane label="龙龙工具" name="longlong">
        <div class="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">基础配置</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">集中维护常用参数和运行操作。</p>
            </div>

            <div class="mt-5 grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">源台地址</label>
                <ElInput v-model="longlongConfig.long_host" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Access Key</label>
                <ElInput v-model="longlongConfig.access_key" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">绑定货源 HID</label>
                <ElInput v-model="longlongConfig.docking" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
                <ElInput v-model="longlongConfig.rate" />
              </div>
            </div>

            <div class="mt-5 flex flex-wrap gap-3">
              <ElButton plain :loading="longlongSaving" @click="saveLonglongConfig">保存配置</ElButton>
              <ElButton plain :loading="longlongInstalling" @click="installCli">安装 CLI</ElButton>
              <ElButton type="primary" :loading="longlongSyncing" @click="runLonglongSync">立即同步产品</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">运行状态</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">从后端运行态读取当前同步和监听情况。</p>
            </div>

            <div class="mt-5 space-y-4">
              <article class="rounded-custom-sm border-full-d p-4">
                <div class="flex items-center justify-between gap-3">
                  <span class="text-sm font-medium text-g-800">CLI 状态</span>
                  <ElTag :type="cliStatus.installed ? 'success' : 'warning'" effect="plain">
                    {{ cliStatus.installed ? '已安装' : '未安装' }}
                  </ElTag>
                </div>
                <p class="mt-2 text-sm leading-6 text-g-500">{{ cliStatus.message || '暂无信息' }}</p>
              </article>
              <article class="rounded-custom-sm border-full-d p-4">
                <p class="text-xs font-medium text-g-400">产品同步</p>
                <p class="mt-2 text-base font-semibold text-g-900">
                  {{ longlongStatus.sync_running ? '运行中' : '空闲' }}
                </p>
                <p class="mt-2 text-sm text-g-500">{{ longlongStatus.last_sync_msg || '暂无结果' }}</p>
              </article>
              <article class="rounded-custom-sm border-full-d p-4">
                <p class="text-xs font-medium text-g-400">订单监听</p>
                <p class="mt-2 text-base font-semibold text-g-900">
                  {{ longlongStatus.listen_running ? '运行中' : '空闲' }}
                </p>
                <p class="mt-2 text-sm text-g-500">{{ longlongStatus.last_listen_msg || '暂无结果' }}</p>
              </article>
            </div>
          </section>
        </div>
      </ElTabPane>
    </ElTabs>

    <ElDialog v-model="previewVisible" title="同步差异预览" width="1080px" destroy-on-close>
      <div class="grid gap-4 md:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">上游商品</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ previewResult?.upstream_count || 0 }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">本地商品</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ previewResult?.local_count || 0 }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">差异总数</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ diffCount }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">货源</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ previewResult?.supplier_name || '-' }}</p>
        </article>
      </div>

      <div class="mt-4">
        <ArtTable :data="previewResult?.diffs || []" :columns="previewColumns" :show-table-header="true" />
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="previewVisible = false">关闭</ElButton>
          <ElButton type="primary" :loading="executeLoading" :disabled="!diffCount" @click="executePreview">
            执行同步
          </ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { fetchLegacyAdminCategoryOptions, type LegacyAdminCategory } from '@/api/legacy/admin-categories'
  import { fetchLegacyAdminSuppliers, type LegacyAdminSupplier } from '@/api/legacy/admin-suppliers'
  import {
    executeLegacySync,
    fetchLegacyAutoSyncStatus,
    fetchLegacyLonglongCliStatus,
    fetchLegacyLonglongToolConfig,
    fetchLegacyLonglongToolStatus,
    fetchLegacyMonitoredSuppliers,
    fetchLegacySyncConfig,
    fetchLegacySyncLogs,
    fetchLegacySyncPreview,
    installLegacyLonglongCli,
    runLegacyLonglongToolSync,
    saveLegacyLonglongToolConfig,
    saveLegacySyncConfig,
    type LegacyAutoSyncStatus,
    type LegacyLonglongCliStatus,
    type LegacyLonglongToolConfig,
    type LegacyLonglongToolStatus,
    type LegacyMonitoredSupplier,
    type LegacySyncConfig,
    type LegacySyncDiffItem,
    type LegacySyncLogItem,
    type LegacySyncPreviewResult
  } from '@/api/legacy/admin-sync'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminSyncMonitorPage' })

  const loading = ref(false)
  const saving = ref(false)
  const logLoading = ref(false)
  const executeLoading = ref(false)
  const longlongSaving = ref(false)
  const longlongSyncing = ref(false)
  const longlongInstalling = ref(false)

  const activeTab = ref('overview')
  const monitoredSuppliers = ref<LegacyMonitoredSupplier[]>([])
  const supplierOptions = ref<LegacyAdminSupplier[]>([])
  const categoryOptions = ref<LegacyAdminCategory[]>([])
  const logs = ref<LegacySyncLogItem[]>([])
  const selectedSupplierIds = ref<number[]>([])

  const syncConfig = reactive<LegacySyncConfig>({
    supplier_ids: '',
    price_rates: {},
    category_rates: {},
    sync_price: true,
    sync_status: true,
    sync_content: true,
    sync_name: false,
    clone_enabled: false,
    force_price_up: false,
    clone_category: false,
    skip_categories: [],
    name_replace: {},
    secret_price_rate: 0,
    auto_sync_enabled: false,
    auto_sync_interval: 30
  })

  const autoStatus = ref<LegacyAutoSyncStatus>({
    enabled: false,
    interval: 30,
    running: false,
    last_run_time: '',
    last_result: '',
    total_runs: 0,
    next_run_time: ''
  })

  const longlongConfig = reactive<LegacyLonglongToolConfig>({
    long_host: '',
    access_key: '',
    mysql_host: '',
    mysql_port: '3306',
    mysql_user: '',
    mysql_password: '',
    mysql_database: '',
    class_table: '',
    order_table: '',
    docking: '',
    rate: '1.5',
    name_prefix: '',
    category: '',
    cover_price: true,
    cover_desc: true,
    cover_name: false,
    sort: '0',
    cron_value: '30',
    cron_unit: 'minute'
  })

  const longlongStatus = ref<LegacyLonglongToolStatus>({
    sync_running: false,
    listen_running: false,
    last_sync_time: '',
    last_sync_msg: '',
    last_listen_at: '',
    last_listen_msg: '',
    sync_count: 0,
    listen_count: 0
  })

  const cliStatus = ref<LegacyLonglongCliStatus>({
    installed: false,
    path: '',
    os: '',
    message: ''
  })

  const replaceDraft = reactive({
    oldValue: '',
    newValue: ''
  })

  const logPagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const previewVisible = ref(false)
  const previewResult = ref<LegacySyncPreviewResult | null>(null)

  const totalLocalCount = computed(() =>
    monitoredSuppliers.value.reduce((sum, item) => sum + Number(item.local_count || 0), 0)
  )
  const totalActiveCount = computed(() =>
    monitoredSuppliers.value.reduce((sum, item) => sum + Number(item.active_count || 0), 0)
  )
  const replaceEntries = computed(() =>
    Object.entries(syncConfig.name_replace || {}).map(([oldValue, newValue]) => ({ oldValue, newValue }))
  )
  const diffCount = computed(() => previewResult.value?.diffs?.length || 0)

  const supplierName = (hid: number) =>
    supplierOptions.value.find((item) => item.hid === hid)?.name || `HID ${hid}`

  const diffTagType = (action: string): 'success' | 'warning' | 'danger' | 'primary' | 'info' => {
    if (['更新价格', '更新说明', '更新名称'].includes(action)) return 'primary'
    if (['上架', '克隆上架'].includes(action)) return 'success'
    if (action === '下架') return 'warning'
    if (action === '新增分类') return 'info'
    return 'danger'
  }

  const handleSupplierChange = (value: number[]) => {
    selectedSupplierIds.value = value
    syncConfig.supplier_ids = value.join(',')
    value.forEach((hid) => {
      if (!syncConfig.price_rates[String(hid)]) syncConfig.price_rates[String(hid)] = 1
    })
  }

  const updateSupplierRate = (hid: number, value: number) => {
    syncConfig.price_rates[String(hid)] = value
  }

  const addNameReplace = () => {
    if (!replaceDraft.oldValue.trim()) {
      ElMessage.warning('请先填写原文字')
      return
    }
    syncConfig.name_replace[replaceDraft.oldValue.trim()] = replaceDraft.newValue.trim()
    replaceDraft.oldValue = ''
    replaceDraft.newValue = ''
  }

  const removeNameReplace = (key: string) => {
    delete syncConfig.name_replace[key]
    syncConfig.name_replace = { ...syncConfig.name_replace }
  }

  const loadConfig = async () => {
    Object.assign(syncConfig, await fetchLegacySyncConfig())
    selectedSupplierIds.value = String(syncConfig.supplier_ids || '')
      .split(',')
      .map((item) => Number(item))
      .filter((item) => item > 0)
  }

  const loadOverview = async () => {
    const [suppliers, auto] = await Promise.all([
      fetchLegacyMonitoredSuppliers().catch(() => []),
      fetchLegacyAutoSyncStatus().catch(() => autoStatus.value)
    ])
    monitoredSuppliers.value = suppliers || []
    autoStatus.value = auto
  }

  const loadBaseOptions = async () => {
    const [suppliers, categories] = await Promise.all([
      fetchLegacyAdminSuppliers().catch(() => []),
      fetchLegacyAdminCategoryOptions().catch(() => [])
    ])
    supplierOptions.value = suppliers || []
    categoryOptions.value = categories || []
  }

  const loadLogs = async () => {
    logLoading.value = true
    try {
      const result = await fetchLegacySyncLogs({
        page: logPagination.current,
        page_size: logPagination.size
      })
      logs.value = result.list || []
      logPagination.total = Number(result.total || 0)
    } finally {
      logLoading.value = false
    }
  }

  const loadLonglong = async () => {
    const [config, status, cli] = await Promise.all([
      fetchLegacyLonglongToolConfig().catch(() => longlongConfig),
      fetchLegacyLonglongToolStatus().catch(() => longlongStatus.value),
      fetchLegacyLonglongCliStatus().catch(() => cliStatus.value)
    ])
    Object.assign(longlongConfig, config)
    longlongStatus.value = status
    cliStatus.value = cli
  }

  const refreshAll = async () => {
    loading.value = true
    try {
      await Promise.all([loadConfig(), loadOverview(), loadBaseOptions(), loadLogs(), loadLonglong()])
    } finally {
      loading.value = false
    }
  }

  const saveConfig = async () => {
    saving.value = true
    try {
      await saveLegacySyncConfig({
        ...syncConfig,
        supplier_ids: selectedSupplierIds.value.join(',')
      })
      ElMessage.success('同步配置已保存')
      await refreshAll()
    } finally {
      saving.value = false
    }
  }

  const openPreview = async (supplier: LegacyMonitoredSupplier) => {
    previewVisible.value = true
    previewResult.value = await fetchLegacySyncPreview(supplier.hid)
  }

  const executePreview = async () => {
    if (!previewResult.value) return
    await ElMessageBox.confirm('确认执行当前差异同步吗？', '执行同步', { type: 'warning' })
    executeLoading.value = true
    try {
      const result = await executeLegacySync(previewResult.value.supplier_id)
      ElMessage.success(`同步完成，应用 ${result.applied} 项，失败 ${result.failed} 项`)
      previewVisible.value = false
      await Promise.all([loadOverview(), loadLogs()])
    } finally {
      executeLoading.value = false
    }
  }

  const saveLonglongConfig = async () => {
    longlongSaving.value = true
    try {
      await saveLegacyLonglongToolConfig({ ...longlongConfig })
      ElMessage.success('龙龙配置已保存')
      await loadLonglong()
    } finally {
      longlongSaving.value = false
    }
  }

  const runLonglongSync = async () => {
    longlongSyncing.value = true
    try {
      const result = await runLegacyLonglongToolSync()
      ElMessage.success(result.msg || '同步已触发')
      await loadLonglong()
    } finally {
      longlongSyncing.value = false
    }
  }

  const installCli = async () => {
    longlongInstalling.value = true
    try {
      const result = await installLegacyLonglongCli()
      ElMessage.success(result.msg || 'CLI 安装命令已执行')
      await loadLonglong()
    } finally {
      longlongInstalling.value = false
    }
  }

  const handleLogPageChange = (page: number) => {
    logPagination.current = page
    loadLogs()
  }

  const handleLogSizeChange = (size: number) => {
    logPagination.size = size
    logPagination.current = 1
    loadLogs()
  }

  const { columns: previewColumns } = useTableColumns<LegacySyncDiffItem>(() => [
    {
      prop: 'action',
      label: '操作',
      width: 120,
      formatter: (row) => h(ElTag, { type: diffTagType(row.action), effect: 'plain' }, () => row.action || '-')
    },
    { prop: 'cid', label: 'CID', width: 90, formatter: (row) => row.cid || '-' },
    { prop: 'name', label: '商品名称', minWidth: 220 },
    { prop: 'category', label: '分类', width: 140, formatter: (row) => row.category || '-' },
    { prop: 'old_value', label: '变更前', minWidth: 180, formatter: (row) => row.old_value || '-' },
    { prop: 'new_value', label: '变更后', minWidth: 180, formatter: (row) => row.new_value || '-' }
  ])

  const { columns: logColumns, columnChecks: logColumnChecks } = useTableColumns<LegacySyncLogItem>(() => [
    { prop: 'sync_time', label: '时间', width: 180 },
    { prop: 'supplier_name', label: '货源', width: 160 },
    {
      prop: 'action',
      label: '操作',
      width: 120,
      formatter: (row) => h(ElTag, { type: diffTagType(row.action), effect: 'plain' }, () => row.action || '-')
    },
    { prop: 'product_name', label: '商品', minWidth: 220, formatter: (row) => row.product_name || '-' },
    { prop: 'category_name', label: '分类', width: 140, formatter: (row) => row.category_name || '-' },
    { prop: 'data_before', label: '变更前', minWidth: 180, formatter: (row) => row.data_before || '-' },
    { prop: 'data_after', label: '变更后', minWidth: 180, formatter: (row) => row.data_after || '-' }
  ])

  onMounted(() => {
    refreshAll()
  })
</script>
