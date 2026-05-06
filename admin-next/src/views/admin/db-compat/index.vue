<template>
  <div class="admin-db-compat-page art-full-height">
    <ElCard class="art-table-card mb-4">
      <ArtTableHeader :loading="checking" layout="refresh" @refresh="loadCompatCheck">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">数据库工具</ElTag>
            <ElTag effect="plain">检测时间 {{ compatCheck?.check_time || '暂无' }}</ElTag>
            <ElTag :type="compatHealthy ? 'success' : 'warning'" effect="plain">
              {{ compatHealthy ? '结构兼容' : '需要修复' }}
            </ElTag>
            <ElTag :type="syncReady ? 'success' : 'info'" effect="plain">
              {{ syncReady ? '可执行导入' : '等待预检查' }}
            </ElTag>
            <ElButton plain :loading="testing" @click="runTest">预检查</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>
    </ElCard>

    <ElTabs v-model="activeTab">
      <ElTabPane label="同步数据" name="sync">
        <div class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">源数据库连接</h3>
              <p class="mt-1 text-sm text-g-500">填写旧库连接并执行预检查。</p>
            </div>

            <div class="mt-5 grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">数据库地址</label>
                <ElInput v-model="syncForm.host" placeholder="localhost" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">数据库端口</label>
                <ElInputNumber v-model="syncForm.port" class="w-full" :min="1" :max="65535" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">数据库名</label>
                <ElInput v-model="syncForm.db_name" placeholder="请输入数据库名" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">数据库用户名</label>
                <ElInput v-model="syncForm.user" placeholder="root" />
              </div>
            </div>

            <div class="mt-4">
              <label class="mb-2 block text-sm font-medium text-g-800">数据库密码</label>
              <ElInput v-model="syncForm.password" type="password" show-password placeholder="请输入数据库密码" />
            </div>

            <div class="mt-4 rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">更新已存在数据</p>
                  <p class="mt-1 text-xs text-g-500">开启后会覆盖同主键记录</p>
                </div>
                <ElSwitch v-model="syncForm.update_existing" />
              </div>
            </div>

            <p class="mt-4 text-sm text-[var(--el-color-warning)]">
              导入前请先备份当前数据库。确认令牌仅在预检查通过后短时间内有效。
            </p>

            <div class="mt-4 flex flex-wrap gap-3">
              <ElButton plain :loading="testing" :disabled="!canSync" @click="runTest">预检查</ElButton>
              <ElButton type="primary" :loading="syncing" :disabled="!syncReady" @click="runSync">开始导入</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">预检查结果</h3>
              <p class="mt-1 text-sm text-g-500">确认连接状态、结构兼容与导入范围。</p>
            </div>

            <div class="mt-5 space-y-4">
              <div
                v-if="testing"
                class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3 text-sm text-g-500"
              >
                正在进行预检查，请稍候。
              </div>

              <template v-else-if="syncTest">
                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">
                        {{ syncTest.connected ? (syncTest.ready ? '预检查通过' : '预检查未通过') : '连接失败' }}
                      </p>
                      <p class="mt-2 text-sm leading-6 text-g-500">{{ syncTest.summary }}</p>
                    </div>
                    <ElTag :type="syncTest.connected ? (syncTest.ready ? 'success' : 'warning') : 'danger'" effect="plain">
                      {{ syncTest.connected ? (syncTest.ready ? '可导入' : '需处理') : '失败' }}
                    </ElTag>
                  </div>
                </article>

                <div v-if="syncTest.warnings?.length" class="flex flex-wrap gap-2">
                  <ElTag
                    v-for="(warning, index) in syncTest.warnings"
                    :key="index"
                    type="warning"
                    effect="plain"
                  >
                    {{ warning }}
                  </ElTag>
                </div>

                <div v-if="syncTest.tables" class="grid gap-3 sm:grid-cols-2">
                  <article
                    v-for="(count, tableName) in syncTest.tables"
                    :key="tableName"
                    class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3"
                  >
                    <p class="text-xs text-g-400">{{ tableLabel(tableName) }}</p>
                    <p class="mt-2 text-base font-semibold text-g-900">
                      {{ Number(count) >= 0 ? `${count} 条` : '表不存在' }}
                    </p>
                  </article>
                </div>
              </template>

              <ElEmpty v-else description="先执行一次预检查" />
            </div>
          </section>
        </div>

        <ElCard v-if="syncTest" class="art-table-card mt-4">
          <ArtTableHeader v-model:columns="syncCheckColumnChecks" :loading="testing" @refresh="runTest" />
          <ArtTable :data="syncTest.table_checks" :columns="syncCheckColumns" :show-table-header="true" />
        </ElCard>

        <section v-if="syncResult" class="art-card-sm mt-4 p-5">
          <div class="flex items-start justify-between gap-3 border-b-d pb-4">
            <div>
              <h3 class="text-lg font-semibold text-g-900">导入结果</h3>
              <p class="mt-1 text-sm text-g-500">{{ syncResult.summary }}</p>
            </div>
            <ElTag :type="syncResult.success ? 'success' : 'warning'" effect="plain">
              {{ syncResult.success ? '已完成' : '完成但有警告' }}
            </ElTag>
          </div>

          <ElCard class="art-table-card mt-5 !shadow-none">
            <ArtTableHeader v-model:columns="syncResultColumnChecks" :loading="syncing" />
            <ArtTable :data="syncResult.details" :columns="syncResultColumns" :show-table-header="true" />
          </ElCard>

          <div v-if="syncResult.errors.length" class="mt-4 space-y-3">
            <ElAlert
              v-for="(error, index) in syncResult.errors"
              :key="index"
              type="error"
              show-icon
              :closable="false"
              :title="error"
            />
          </div>
        </section>
      </ElTabPane>

      <ElTabPane label="结构检测" name="compat">
        <div class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">结构概览</h3>
              <p class="mt-1 text-sm text-g-500">检测核心表与字段差异。</p>
            </div>

            <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <ElTag effect="plain">{{ compatCheck?.check_time || '暂无检测时间' }}</ElTag>
                <ElTag :type="compatHealthy ? 'success' : 'warning'" effect="plain">
                  {{ compatHealthy ? '结构兼容' : '存在差异' }}
                </ElTag>
              </div>
              <p class="mt-3 text-sm text-g-500">{{ compatCheck?.summary || '尚未检测' }}</p>
            </div>

            <div class="mt-5 flex flex-wrap gap-3">
              <ElButton plain :loading="checking" @click="loadCompatCheck">重新检测</ElButton>
              <ElButton type="primary" plain :loading="fixing" :disabled="checking" @click="runCompatFix">一键修复</ElButton>
            </div>

            <p class="mt-4 text-sm text-g-500">修复会自动创建缺失表和字段，建议先备份数据库。</p>
          </section>

          <section class="art-card-sm p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">差异明细</h3>
              <p class="mt-1 text-sm text-g-500">按缺失表、缺失列与现有核心表展示。</p>
            </div>

            <div v-if="compatCheck" class="mt-5 space-y-5">
              <div>
                <div class="mb-3 flex items-center justify-between">
                  <h4 class="text-sm font-semibold text-g-900">缺失表</h4>
                  <ElTag :type="compatCheck.missing_tables.length ? 'warning' : 'success'" effect="plain">
                    {{ compatCheck.missing_tables.length }} 个
                  </ElTag>
                </div>
                <div v-if="compatCheck.missing_tables.length" class="flex flex-wrap gap-2">
                  <ElTag v-for="item in compatCheck.missing_tables" :key="item" type="warning" effect="plain">{{ item }}</ElTag>
                </div>
                <ElEmpty v-else description="没有缺失表" />
              </div>

              <div>
                <div class="mb-3 flex items-center justify-between">
                  <h4 class="text-sm font-semibold text-g-900">已存在核心表</h4>
                  <ElTag effect="plain">{{ compatCheck.existing_tables.length }} 个</ElTag>
                </div>
                <div class="flex flex-wrap gap-2">
                  <ElTag v-for="item in compatCheck.existing_tables" :key="item" type="success" effect="plain">{{ item }}</ElTag>
                </div>
              </div>

              <div v-if="compatCheck.extra_tables.length">
                <div class="mb-3 flex items-center justify-between">
                  <h4 class="text-sm font-semibold text-g-900">数据库其他表</h4>
                  <ElTag effect="plain">{{ compatCheck.extra_tables.length }} 个</ElTag>
                </div>
                <div class="flex flex-wrap gap-2">
                  <ElTag v-for="item in compatCheck.extra_tables" :key="item" effect="plain">{{ item }}</ElTag>
                </div>
              </div>
            </div>

            <ElEmpty v-else description="正在等待结构检测" />
          </section>
        </div>

        <ElCard v-if="compatCheck?.missing_columns.length" class="art-table-card mt-4">
          <ArtTableHeader v-model:columns="missingColumnChecks" :loading="checking" @refresh="loadCompatCheck" />
          <ArtTable :data="compatCheck.missing_columns" :columns="missingColumns" :show-table-header="true" />
        </ElCard>

        <section v-if="fixResult" class="art-card-sm mt-4 p-5">
          <div class="flex items-start justify-between gap-3 border-b-d pb-4">
            <div>
              <h3 class="text-lg font-semibold text-g-900">修复结果</h3>
              <p class="mt-1 text-sm text-g-500">{{ fixResult.summary }}</p>
            </div>
            <ElTag type="success" effect="plain">{{ fixResult.fix_time }}</ElTag>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <p class="text-sm font-semibold text-g-900">创建的表</p>
              <div class="mt-3 flex flex-wrap gap-2">
                <ElTag v-for="item in fixResult.tables_created" :key="item" type="success" effect="plain">{{ item }}</ElTag>
              </div>
            </article>
            <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <p class="text-sm font-semibold text-g-900">添加的列</p>
              <div class="mt-3 flex flex-wrap gap-2">
                <ElTag v-for="item in fixResult.columns_added" :key="item" type="primary" effect="plain">{{ item }}</ElTag>
              </div>
            </article>
          </div>

          <p v-if="fixResult.admin_created" class="mt-4 text-sm text-[var(--el-color-warning)]">
            已自动创建管理员账号：admin / admin123，请尽快修改密码。
          </p>

          <div v-if="fixResult.errors.length" class="mt-4 space-y-3">
            <ElAlert
              v-for="(error, index) in fixResult.errors"
              :key="index"
              type="error"
              show-icon
              :closable="false"
              :title="error"
            />
          </div>
        </section>
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyDBCompatCheck,
    runLegacyDBCompatFix,
    runLegacyDBSyncExecute,
    runLegacyDBSyncTest,
    type LegacyDBCompatCheckResult,
    type LegacyDBCompatFixResult,
    type LegacyDBSyncResult,
    type LegacyDBSyncTableCheck,
    type LegacyDBSyncTableInfo,
    type LegacyDBSyncTestResult,
    type LegacyMissingColumnInfo
  } from '@/api/legacy/admin-db-tools'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminDbCompatPage' })

  const activeTab = ref('sync')
  const checking = ref(false)
  const fixing = ref(false)
  const testing = ref(false)
  const syncing = ref(false)

  const compatCheck = ref<LegacyDBCompatCheckResult | null>(null)
  const fixResult = ref<LegacyDBCompatFixResult | null>(null)
  const syncTest = ref<LegacyDBSyncTestResult | null>(null)
  const syncResult = ref<LegacyDBSyncResult | null>(null)
  const confirmationToken = ref('')

  const syncForm = reactive({
    host: 'localhost',
    port: 3306,
    db_name: '',
    user: 'root',
    password: '',
    update_existing: false
  })

  const canSync = computed(() => {
    return Boolean(syncForm.host && syncForm.db_name && syncForm.user)
  })

  const syncReady = computed(() => {
    return Boolean(syncTest.value?.connected && syncTest.value?.ready && confirmationToken.value)
  })

  const compatHealthy = computed(() => {
    return Boolean(
      compatCheck.value &&
        compatCheck.value.missing_tables.length === 0 &&
        compatCheck.value.missing_columns.length === 0
    )
  })

  const tableLabelMap: Record<string, string> = {
    qingka_wangke_dengji: '等级',
    qingka_wangke_huoyuan: '货源',
    qingka_wangke_user: '用户',
    qingka_wangke_fenlei: '分类',
    qingka_wangke_class: '商品',
    qingka_wangke_config: '配置',
    qingka_wangke_gonggao: '公告',
    qingka_wangke_mijia: '密价',
    qingka_wangke_km: '卡密',
    qingka_wangke_order: '订单',
    qingka_wangke_pay: '支付'
  }

  function tableLabel(tableName: string) {
    return tableLabelMap[tableName] || tableName
  }

  async function loadCompatCheck() {
    checking.value = true
    try {
      compatCheck.value = await fetchLegacyDBCompatCheck()
    } finally {
      checking.value = false
    }
  }

  async function runCompatFix() {
    await ElMessageBox.confirm('将自动创建缺失的表和列，此操作不可逆。建议先备份数据库。', '确认修复', {
      type: 'warning'
    })
    fixing.value = true
    try {
      fixResult.value = await runLegacyDBCompatFix()
      ElMessage.success('修复完成')
      await loadCompatCheck()
    } finally {
      fixing.value = false
    }
  }

  async function runTest() {
    if (!canSync.value) {
      ElMessage.warning('请填写完整的数据库连接信息')
      return
    }
    testing.value = true
    syncResult.value = null
    confirmationToken.value = ''
    try {
      syncTest.value = await runLegacyDBSyncTest({ ...syncForm })
      confirmationToken.value = syncTest.value.confirmation_token || ''
      if (!syncTest.value.connected) {
        ElMessage.error(syncTest.value.error || '连接失败')
      } else if (syncTest.value.ready) {
        ElMessage.success('预检查通过')
      } else {
        ElMessage.warning('预检查未通过，请先处理结构差异')
      }
    } finally {
      testing.value = false
    }
  }

  async function runSync() {
    if (!syncReady.value || !syncTest.value) {
      ElMessage.warning('请先完成预检查并确保通过')
      return
    }
    const warningText = syncTest.value.warnings?.length ? `\n\n风险提示：${syncTest.value.warnings.join('；')}` : ''
    await ElMessageBox.confirm(
      `请确认已完成当前数据库备份。导入将写入当前系统核心数据表。${warningText}`,
      '确认开始导入',
      { type: 'warning' }
    )
    syncing.value = true
    try {
      syncResult.value = await runLegacyDBSyncExecute({
        ...syncForm,
        confirmation_token: confirmationToken.value
      })
      confirmationToken.value = ''
      ElMessage.success(syncResult.value.success ? '导入完成' : '导入完成，但有部分警告')
    } finally {
      syncing.value = false
    }
  }

  const { columns: syncCheckColumns, columnChecks: syncCheckColumnChecks } =
    useTableColumns<LegacyDBSyncTableCheck>(() => [
      { prop: 'label', label: '数据类型', width: 120 },
      { prop: 'source_table', label: '命中源表', width: 180 },
      { prop: 'source_count', label: '源库条数', width: 100, align: 'center' },
      { prop: 'local_count', label: '本地条数', width: 100, align: 'center' },
      {
        prop: 'ready',
        label: '状态',
        width: 120,
        formatter: (row) =>
          h(
            ElTag,
            {
              type: row.skip ? 'warning' : row.ready ? 'success' : 'danger',
              effect: 'plain'
            },
            () => (row.skip ? '将跳过' : row.ready ? '可导入' : '需处理')
          )
      },
      {
        prop: 'missing_local_columns',
        label: '缺失字段',
        minWidth: 180,
        formatter: (row) => row.missing_local_columns?.length ? row.missing_local_columns.join(', ') : '无'
      },
      { prop: 'message', label: '说明', minWidth: 220 }
    ])

  const { columns: syncResultColumns, columnChecks: syncResultColumnChecks } =
    useTableColumns<LegacyDBSyncTableInfo>(() => [
      { prop: 'label', label: '数据类型', width: 120 },
      { prop: 'source_table', label: '源表', width: 180 },
      { prop: 'total', label: '总数', width: 90, align: 'center' },
      { prop: 'inserted', label: '新增', width: 90, align: 'center' },
      { prop: 'updated', label: '更新', width: 90, align: 'center' },
      { prop: 'skipped', label: '跳过', width: 90, align: 'center' },
      { prop: 'failed', label: '失败', width: 90, align: 'center' },
      { prop: 'message', label: '说明', minWidth: 220 }
    ])

  const { columns: missingColumns, columnChecks: missingColumnChecks } =
    useTableColumns<LegacyMissingColumnInfo>(() => [
      { prop: 'table', label: '表名', minWidth: 220 },
      { prop: 'column', label: '列名', minWidth: 180 },
      { prop: 'type', label: '类型', minWidth: 200 }
    ])

  onMounted(async () => {
    await loadCompatCheck()
  })
</script>
