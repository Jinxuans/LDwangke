<template>
  <div class="admin-docking-page art-full-height">
    <section class="rounded-custom-sm border-full-d bg-box p-5">
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">当前货源</label>
          <ElSelect
            v-model="selectedSupplierId"
            class="w-full"
            clearable
            filterable
            placeholder="请选择要拉取商品的货源"
            @change="handleSupplierChange"
          >
            <ElOption
              v-for="item in supplierOptions"
              :key="item.hid"
              :label="`${item.name} (HID:${item.hid})`"
              :value="item.hid"
            />
          </ElSelect>
          <p class="mt-2 text-sm text-g-500">
            选择货源后将自动拉取商品，并按本地课程库标记是否已入库。
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2 lg:justify-end">
          <ElTag v-if="currentSupplier" type="primary" effect="plain">
            {{ currentSupplier.name }}
          </ElTag>
          <ElTag v-if="currentSupplier" effect="plain">
            {{ currentSupplier?.pt || '未标记平台' }}
          </ElTag>
          <ElTag v-if="currentSupplier" type="info" effect="plain">HID {{ currentSupplier.hid }}</ElTag>
          <ElTag v-else type="info" effect="plain">未选择货源</ElTag>
        </div>
      </div>
    </section>

    <ArtSearchBar
      class="mt-5"
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        @refresh="handleRefreshProducts"
      >
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">对接插件</ElTag>
            <ElTag effect="plain">上游商品 {{ rawProducts.length }}</ElTag>
            <ElTag effect="plain">筛选结果 {{ filteredProducts.length }}</ElTag>
            <ElTag type="success" effect="plain">已入库 {{ dockedCount }}</ElTag>
            <ElTag type="warning" effect="plain">当前已选 {{ selectedProductIds.length }}</ElTag>
            <ElButton plain :disabled="!filteredProducts.length" @click="selectAllFiltered">
              全选筛选结果
            </ElButton>
            <ElButton plain :disabled="!selectedProductIds.length" @click="handleClearSelection">
              清空选择
            </ElButton>
            <ElButton plain :disabled="!selectedSupplierId" @click="openSyncPreview">同步预览</ElButton>
            <ElButton plain :disabled="!selectedSupplierId" @click="handleSyncStatus">检查失效商品</ElButton>
            <ElDropdown
              trigger="click"
              :disabled="!selectedSupplierId || statusBatching"
              @command="handleSupplierStatusCommand"
            >
              <ElButton plain :loading="statusBatching" :disabled="!selectedSupplierId">
                批量状态
                <ElIcon class="el-icon--right"><ArrowDown /></ElIcon>
              </ElButton>
              <template #dropdown>
                <ElDropdownMenu>
                  <ElDropdownItem command="up">本货源全部上架</ElDropdownItem>
                  <ElDropdownItem command="down" divided>本货源全部下架</ElDropdownItem>
                </ElDropdownMenu>
              </template>
            </ElDropdown>
            <ElButton plain :disabled="!selectedSupplierId" @click="openImportDialog">一键对接</ElButton>
            <ElButton
              type="primary"
              plain
              :disabled="!selectedProductIds.length"
              @click="openBatchDialog"
            >
              批量上架
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        row-key="cid"
        :loading="loading"
        :data="pagedProducts"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <div class="mt-5 grid gap-5 xl:grid-cols-2">
      <ElCard class="art-table-card">
        <template #header>
          <div>
            <h3 class="text-lg font-semibold text-g-900">批量替换关键词</h3>
            <p class="mt-1 text-sm text-g-500">对本地课程名做统一替换，适合清理上游品牌或敏感词。</p>
          </div>
        </template>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">原关键词</label>
            <ElInput v-model="replaceForm.search" placeholder="请输入要替换的关键词" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">替换为</label>
            <ElInput v-model="replaceForm.replace" placeholder="留空则删除关键词" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">范围</label>
            <ElSelect v-model="replaceForm.scope" class="w-full" @change="replaceForm.scopeId = ''">
              <ElOption label="全部课程" value="all" />
              <ElOption label="按分类" value="cate" />
              <ElOption label="按对接平台 ID" value="docking" />
            </ElSelect>
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">
              {{ replaceForm.scope === 'cate' ? '目标分类' : '范围值' }}
            </label>
            <ElSelect
              v-if="replaceForm.scope === 'cate'"
              v-model="replaceForm.scopeId"
              class="w-full"
              clearable
              filterable
              placeholder="请选择分类"
            >
              <ElOption
                v-for="item in categoryOptions"
                :key="item.id"
                :label="item.name"
                :value="String(item.id)"
              />
            </ElSelect>
            <ElInput
              v-else
              v-model="replaceForm.scopeId"
              :placeholder="replaceForm.scope === 'docking' ? '请输入对接平台 ID' : '全部范围无需填写'"
              :disabled="replaceForm.scope === 'all'"
            />
          </div>
        </div>

        <div class="mt-4 flex justify-end">
          <ElButton type="primary" :loading="replaceLoading" @click="handleReplaceKeyword">
            执行替换
          </ElButton>
        </div>
      </ElCard>

      <ElCard class="art-table-card">
        <template #header>
          <div>
            <h3 class="text-lg font-semibold text-g-900">批量添加前缀</h3>
            <p class="mt-1 text-sm text-g-500">对本地课程名统一加前缀，便于区分活动期或渠道期。</p>
          </div>
        </template>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">前缀内容</label>
            <ElInput v-model="prefixForm.prefix" placeholder="例如：[活动]" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">范围</label>
            <ElSelect v-model="prefixForm.scope" class="w-full" @change="prefixForm.scopeId = ''">
              <ElOption label="全部课程" value="all" />
              <ElOption label="按分类" value="cate" />
              <ElOption label="按对接平台 ID" value="docking" />
            </ElSelect>
          </div>
          <div class="md:col-span-2">
            <label class="mb-2 block text-sm font-medium text-g-800">
              {{ prefixForm.scope === 'cate' ? '目标分类' : '范围值' }}
            </label>
            <ElSelect
              v-if="prefixForm.scope === 'cate'"
              v-model="prefixForm.scopeId"
              class="w-full"
              clearable
              filterable
              placeholder="请选择分类"
            >
              <ElOption
                v-for="item in categoryOptions"
                :key="item.id"
                :label="item.name"
                :value="String(item.id)"
              />
            </ElSelect>
            <ElInput
              v-else
              v-model="prefixForm.scopeId"
              :placeholder="prefixForm.scope === 'docking' ? '请输入对接平台 ID' : '全部范围无需填写'"
              :disabled="prefixForm.scope === 'all'"
            />
          </div>
        </div>

        <div class="mt-4 flex justify-end">
          <ElButton type="primary" :loading="prefixLoading" @click="handleAddPrefix">
            添加前缀
          </ElButton>
        </div>
      </ElCard>
    </div>

    <ElDialog v-model="addDialogVisible" title="单个上架商品" width="920px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">商品对接信息</h3>
          </div>
          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">课程名称</label>
              <ElInput v-model="addForm.name" maxlength="120" placeholder="请输入课程名称" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">售价</label>
              <ElInput v-model="addForm.price" placeholder="例如 9.90" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">倍率参考</label>
              <div class="flex gap-2">
                <ElInputNumber
                  v-model="addRate"
                  class="flex-1"
                  :min="0.01"
                  :max="100"
                  :step="0.1"
                  :precision="2"
                />
                <ElButton plain @click="applySuggestedPrice">回填售价</ElButton>
              </div>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
              <ElInputNumber v-model="addSortNumber" class="w-full" :min="0" :max="999999" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">分类</label>
              <ElSelect
                v-model="addForm.fenlei"
                class="w-full"
                clearable
                filterable
                placeholder="可不选择"
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
              <label class="mb-2 block text-sm font-medium text-g-800">查询参数</label>
              <ElInput v-model="addForm.getnoun" placeholder="通常使用上游 CID" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">对接参数</label>
              <ElInput v-model="addForm.noun" placeholder="通常使用上游 CID" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">加价方式</label>
              <ElSelect v-model="addForm.yunsuan" class="w-full">
                <ElOption label="乘法 (*)" value="*" />
                <ElOption label="加法 (+)" value="+" />
              </ElSelect>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">上架状态</label>
              <ElSegmented
                v-model="addForm.status"
                :options="[
                  { label: '上架', value: '1' },
                  { label: '下架', value: '0' }
                ]"
                class="w-full"
              />
            </div>
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">说明</label>
              <ElInput
                v-model="addForm.content"
                type="textarea"
                :rows="4"
                resize="none"
                placeholder="可补充对接说明"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">当前商品</span>
              <span class="truncate font-medium text-g-900">{{ addForm.name || '未命名商品' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">价格参考</span>
              <span class="font-medium text-g-900">
                原价 ¥{{ currentAddProduct ? Number(currentAddProduct.price || 0).toFixed(2) : '0.00' }}
                / 上架 ¥{{ addForm.price || '0.00' }}
              </span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="addForm.status === '1' ? 'success' : 'info'" effect="plain">
                {{ addForm.status === '1' ? '上架中' : '已下架' }}
              </ElTag>
              <ElTag type="primary" effect="plain">{{ currentSupplier?.name || '未选货源' }}</ElTag>
              <ElTag :type="addForm.fenlei ? 'warning' : 'info'" effect="plain">
                {{ getCategoryLabel(addForm.fenlei) }}
              </ElTag>
            </div>
          </div>

          <div class="mt-5">
            <label class="mb-2 block text-sm font-medium text-g-800">快速新建分类</label>
            <div class="flex gap-2">
              <ElInput v-model="newCategoryName" placeholder="输入新分类名称" />
              <ElButton plain :loading="categorySaving" @click="createCategoryFor('single')">
                新建
              </ElButton>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="addDialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="addSaving" @click="submitAddProduct">确认上架</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="batchDialogVisible" title="批量上架" width="760px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1fr_0.92fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">批量上架参数</h3>
          </div>
          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
              <ElInputNumber
                v-model="batchForm.rate"
                class="w-full"
                :min="0.01"
                :max="100"
                :step="0.1"
                :precision="2"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">默认排序</label>
              <ElInputNumber v-model="batchForm.sort" class="w-full" :min="0" :max="999999" />
            </div>
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">目标分类</label>
              <ElSelect
                v-model="batchForm.fenlei"
                class="w-full"
                clearable
                filterable
                placeholder="可不选择"
              >
                <ElOption
                  v-for="item in categoryOptions"
                  :key="item.id"
                  :label="item.name"
                  :value="String(item.id)"
                />
              </ElSelect>
            </div>
            <div class="md:col-span-2 rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-g-900">跳过已入库商品</p>
                  <p class="mt-1 text-sm text-g-500">开启后仅对未入库商品执行批量上架。</p>
                </div>
                <ElSwitch v-model="batchForm.skipExists" />
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">已选择商品</span>
              <span class="font-medium text-g-900">{{ selectedProductIds.length }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">本次待执行</span>
              <span class="font-medium text-g-900">{{ batchCandidates.length }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">上架模式</span>
              <span class="font-medium text-g-900">{{ batchForm.skipExists ? '跳过已入库' : '全部执行' }}</span>
            </div>
          </div>

          <div class="mt-4">
            <label class="mb-2 block text-sm font-medium text-g-800">快速新建分类</label>
            <div class="flex gap-2">
              <ElInput v-model="newCategoryName" placeholder="输入新分类名称" />
              <ElButton plain :loading="categorySaving" @click="createCategoryFor('batch')">
                新建
              </ElButton>
            </div>
          </div>

          <ElProgress v-if="batchRunning" class="mt-4" :percentage="batchProgress" :stroke-width="14" />
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton :disabled="batchRunning" @click="batchDialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="batchRunning" @click="submitBatchAdd">开始上架</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="importDialogVisible" title="一键对接" width="760px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">导入参数</h3>
          </div>
          <div class="mt-5 grid gap-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">货源名称</label>
              <ElInput :model-value="currentSupplier?.name || ''" disabled />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">上游分类</label>
              <ElSelect v-model="importForm.category" class="w-full" filterable>
                <ElOption label="全部分类" value="999999" />
                <ElOption
                  v-for="item in upstreamCategoryOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">本地分类名</label>
              <ElSelect
                v-model="importForm.name"
                class="w-full"
                clearable
                filterable
                placeholder="默认使用货源名"
              >
                <ElOption
                  v-for="item in categoryOptions"
                  :key="item.id"
                  :label="item.name"
                  :value="item.name"
                />
              </ElSelect>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">快速新建分类</label>
              <div class="flex gap-2">
                <ElInput v-model="newCategoryName" placeholder="输入新分类名称" />
                <ElButton plain :loading="categorySaving" @click="createCategoryFor('import')">
                  新建
                </ElButton>
              </div>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
              <ElInputNumber
                v-model="importForm.pricee"
                class="w-full"
                :min="0.01"
                :max="100"
                :step="0.1"
                :precision="2"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <label class="mb-2 block text-sm font-medium text-g-800">导入模式</label>
          <ElSegmented
            v-model="importForm.fd"
            :options="[
              { label: '全量导入', value: 0 },
              { label: '仅更新已有', value: 1 }
            ]"
            class="w-full"
          />

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">当前货源</span>
              <span class="font-medium text-g-900">{{ currentSupplier?.name || '未选择货源' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">价格倍率</span>
              <span class="font-medium text-g-900">{{ importForm.pricee.toFixed(2) }} 倍</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">执行方式</span>
              <span class="font-medium text-g-900">
                {{ importForm.fd === 0 ? '新增并更新商品' : '仅更新已有商品' }}
              </span>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="importDialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="importing" @click="submitImport">开始对接</ElButton>
        </div>
      </template>
    </ElDialog>
    <ElDialog v-model="previewVisible" title="同步差异预览" width="1080px" destroy-on-close>
      <div v-loading="previewLoading" class="grid gap-4 md:grid-cols-4">
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
          <p class="mt-2 text-base font-semibold text-g-900">{{ previewResult?.supplier_name || currentSupplier?.name || '-' }}</p>
        </article>
      </div>

      <div class="mt-4">
        <ArtTable :data="previewResult?.diffs || []" :columns="previewColumns" :show-table-header="true" />
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="previewVisible = false">关闭</ElButton>
          <ElButton type="primary" :loading="executeLoading" :disabled="!diffCount" @click="executeSyncPreview">
            执行同步
          </ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, nextTick, onMounted, reactive, ref, watch } from 'vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    addLegacyDockingClass,
    batchAddLegacyClassPrefix,
    batchReplaceLegacyClassKeyword,
    fetchLegacySupplierProducts,
    type LegacyDockingProduct
  } from '@/api/legacy/admin-docking'
  import {
    fetchLegacyAdminCategoryOptions,
    saveLegacyAdminCategory,
    type LegacyAdminCategory
  } from '@/api/legacy/admin-categories'
  import {
    batchUpdateLegacySupplierClassStatus,
    fetchLegacyAdminSuppliers,
    importLegacySupplier,
    syncLegacySupplierStatus,
    type LegacyAdminSupplier
  } from '@/api/legacy/admin-suppliers'
  import {
    executeLegacySync,
    fetchLegacySyncPreview,
    type LegacySyncDiffItem,
    type LegacySyncPreviewResult
  } from '@/api/legacy/admin-sync'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { ArrowDown } from '@element-plus/icons-vue'

  defineOptions({ name: 'AdminDockingPage' })

  type BatchScope = 'all' | 'cate' | 'docking'
  type CategoryAssignTarget = 'single' | 'batch' | 'import'
  type SupplierStatusCommand = 'up' | 'down'

  const tableRef = ref()
  const loading = ref(false)
  const addSaving = ref(false)
  const batchRunning = ref(false)
  const importing = ref(false)
  const replaceLoading = ref(false)
  const prefixLoading = ref(false)
  const categorySaving = ref(false)
  const statusBatching = ref(false)
  const batchProgress = ref(0)

  const previewLoading = ref(false)
  const executeLoading = ref(false)

  const addDialogVisible = ref(false)
  const batchDialogVisible = ref(false)
  const importDialogVisible = ref(false)
  const previewVisible = ref(false)

  const supplierOptions = ref<LegacyAdminSupplier[]>([])
  const categoryOptions = ref<LegacyAdminCategory[]>([])
  const rawProducts = ref<LegacyDockingProduct[]>([])
  const previewResult = ref<LegacySyncPreviewResult | null>(null)
  const selectedSupplierId = ref<number>()
  const selectedProductIds = ref<string[]>([])
  const currentAddProduct = ref<LegacyDockingProduct>()
  const newCategoryName = ref('')

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    keyword?: string
    upstreamCategory?: string
    state?: string
  }>({
    keyword: undefined,
    upstreamCategory: undefined,
    state: undefined
  })

  const appliedSearch = reactive({
    keyword: undefined as string | undefined,
    upstreamCategory: undefined as string | undefined,
    state: undefined as string | undefined
  })

  const addForm = reactive({
    sort: '10',
    name: '',
    price: '',
    getnoun: '',
    noun: '',
    content: '',
    queryplat: '0',
    docking: '0',
    yunsuan: '*',
    status: '1',
    fenlei: ''
  })

  const addRate = ref(1.1)
  const addSortNumber = ref(10)

  const batchForm = reactive({
    rate: 1.1,
    sort: 10,
    fenlei: '',
    skipExists: true
  })

  const importForm = reactive({
    category: '999999',
    name: '',
    pricee: 1.1,
    fd: 0
  })

  const replaceForm = reactive({
    search: '',
    replace: '',
    scope: 'all' as BatchScope,
    scopeId: ''
  })

  const prefixForm = reactive({
    prefix: '',
    scope: 'all' as BatchScope,
    scopeId: ''
  })

  const currentSupplier = computed(() =>
    supplierOptions.value.find((item) => item.hid === selectedSupplierId.value)
  )

  const dockedCount = computed(() => rawProducts.value.filter((item) => item.states === 1).length)

  const diffCount = computed(() => previewResult.value?.diffs?.length || 0)

  const diffTagType = (action: string): 'success' | 'warning' | 'danger' | 'primary' | 'info' => {
    if (['更新价格', '更新说明', '更新名称'].includes(action)) return 'primary'
    if (['上架', '克隆上架'].includes(action)) return 'success'
    if (action === '下架') return 'warning'
    if (action === '新增分类') return 'info'
    return 'danger'
  }

  const upstreamCategoryOptions = computed(() => {
    const seen = new Map<string, string>()
    for (const item of rawProducts.value) {
      const key = String(item.fenlei || '').trim()
      if (!key) continue
      const label = `${item.category_name || item.fenlei} (${item.fenlei})`
      seen.set(key, label)
    }
    return Array.from(seen.entries()).map(([value, label]) => ({ value, label }))
  })

  const filteredProducts = computed(() => {
    const keyword = appliedSearch.keyword?.toLowerCase() || ''
    const upstreamCategory = appliedSearch.upstreamCategory || ''
    const state = appliedSearch.state || ''

    return rawProducts.value.filter((item) => {
      const hitKeyword = keyword
        ? [item.name, item.content, item.cid].some((field) =>
            String(field || '')
              .toLowerCase()
              .includes(keyword)
          )
        : true
      const hitCategory = upstreamCategory ? String(item.fenlei || '') === upstreamCategory : true
      const hitState =
        state === 'docked' ? item.states === 1 : state === 'undocked' ? item.states !== 1 : true
      return hitKeyword && hitCategory && hitState
    })
  })

  const pagedProducts = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return filteredProducts.value.slice(start, start + pagination.size)
  })

  const selectedProducts = computed(() => {
    const idSet = new Set(selectedProductIds.value.map((item) => String(item)))
    return rawProducts.value.filter((item) => idSet.has(String(item.cid)))
  })

  const batchCandidates = computed(() => {
    return batchForm.skipExists
      ? selectedProducts.value.filter((item) => item.states !== 1)
      : selectedProducts.value
  })

  const normalizeNumber = (value: unknown) => {
    const numberValue = Number(value)
    return Number.isFinite(numberValue) ? numberValue : null
  }

  const normalizePositiveNumber = (value: unknown) => {
    const numberValue = normalizeNumber(value)
    return numberValue !== null && numberValue > 0 ? numberValue : null
  }

  const normalizeNonNegativeNumber = (value: unknown) => {
    const numberValue = normalizeNumber(value)
    return numberValue !== null && numberValue >= 0 ? numberValue : null
  }

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索商品名 / 说明 / CID'
      }
    },
    {
      label: '上游分类',
      key: 'upstreamCategory',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '全部分类',
        options: upstreamCategoryOptions.value
      }
    },
    {
      label: '入库状态',
      key: 'state',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '已入库', value: 'docked' },
          { label: '未入库', value: 'undocked' }
        ]
      }
    }
  ])

  const getCategoryLabel = (value?: string) =>
    categoryOptions.value.find((item) => String(item.id) === String(value || ''))?.name || '未分类'

  const updatePaginationTotal = () => {
    pagination.total = filteredProducts.value.length
    const maxPage = Math.max(1, Math.ceil(Math.max(pagination.total, 1) / pagination.size))
    if (pagination.current > maxPage) {
      pagination.current = maxPage
    }
  }

  const loadBaseOptions = async () => {
    const [suppliers, categories] = await Promise.all([
      fetchLegacyAdminSuppliers(),
      fetchLegacyAdminCategoryOptions()
    ])
    supplierOptions.value = Array.isArray(suppliers) ? suppliers : []
    categoryOptions.value = Array.isArray(categories) ? categories : []
  }

  const clearSelection = () => {
    selectedProductIds.value = []
    tableRef.value?.elTableRef?.clearSelection?.()
  }

  const handleClearSelection = () => {
    clearSelection()
    ElMessage.info('已清空选择')
  }

  const selectAllFiltered = async () => {
    selectedProductIds.value = Array.from(
      new Set(filteredProducts.value.map((item) => String(item.cid)))
    )
    await syncPageSelection()
    ElMessage.success(`已选中全部 ${selectedProductIds.value.length} 个筛选结果`)
  }

  const syncPageSelection = async () => {
    await nextTick()
    const selectedSet = new Set(selectedProductIds.value.map((item) => String(item)))
    tableRef.value?.elTableRef?.clearSelection?.()
    for (const row of pagedProducts.value) {
      if (selectedSet.has(String(row.cid))) {
        tableRef.value?.elTableRef?.toggleRowSelection?.(row, true)
      }
    }
  }

  const reloadProducts = async () => {
    if (!selectedSupplierId.value) {
      rawProducts.value = []
      updatePaginationTotal()
      clearSelection()
      return
    }

    loading.value = true
    try {
      const data = await fetchLegacySupplierProducts(selectedSupplierId.value)
      rawProducts.value = Array.isArray(data) ? data : []
      pagination.current = 1
      updatePaginationTotal()
      clearSelection()
      ElMessage.success(`已拉取 ${rawProducts.value.length} 个上游商品`)
    } finally {
      loading.value = false
    }
  }

  const resetAddForm = () => {
    addForm.sort = '10'
    addForm.name = ''
    addForm.price = ''
    addForm.getnoun = ''
    addForm.noun = ''
    addForm.content = ''
    addForm.queryplat = String(selectedSupplierId.value || 0)
    addForm.docking = String(selectedSupplierId.value || 0)
    addForm.yunsuan = '*'
    addForm.status = '1'
    addForm.fenlei = ''
    addRate.value = 1.1
    addSortNumber.value = 10
  }

  const handleSupplierChange = async () => {
    await reloadProducts()
  }

  const handleRefreshProducts = async () => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }
    await reloadProducts()
  }

  const openSyncPreview = async () => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }

    previewVisible.value = true
    previewLoading.value = true
    previewResult.value = null
    try {
      previewResult.value = await fetchLegacySyncPreview(selectedSupplierId.value)
    } finally {
      previewLoading.value = false
    }
  }

  const executeSyncPreview = async () => {
    if (!previewResult.value) return

    await ElMessageBox.confirm('确认执行当前差异同步吗？', '执行同步', { type: 'warning' })
    executeLoading.value = true
    try {
      const result = await executeLegacySync(previewResult.value.supplier_id)
      ElMessage.success(`同步完成，应用 ${result.applied} 项，失败 ${result.failed} 项`)
      previewVisible.value = false
      await reloadProducts()
    } finally {
      executeLoading.value = false
    }
  }

  const supplierStatusActions: Record<
    SupplierStatusCommand,
    { status: number; label: string; title: string }
  > = {
    up: { status: 1, label: '上架', title: '本货源全部上架' },
    down: { status: 0, label: '下架', title: '本货源全部下架' }
  }

  const handleSupplierStatusCommand = async (command: string | number | object) => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }
    const action = supplierStatusActions[String(command) as SupplierStatusCommand]
    if (!action) return

    statusBatching.value = true
    try {
      const preview = await batchUpdateLegacySupplierClassStatus(
        selectedSupplierId.value,
        action.status,
        true
      )
      const total = Number(preview.total || 0)
      const changed = Number(preview.changed || 0)
      if (total <= 0) {
        ElMessage.warning('当前货源没有本地商品')
        return
      }
      if (changed <= 0) {
        ElMessage.info(`当前货源下 ${total} 个商品已全部${action.label}`)
        return
      }

      const supplierName = currentSupplier.value?.name || `HID ${selectedSupplierId.value}`
      await ElMessageBox.confirm(
        `将把「${supplierName}」下 ${total} 个本地商品设为${action.label}，其中 ${changed} 个需要变更。是否继续？`,
        action.title,
        {
          type: action.status === 1 ? 'warning' : 'error',
          confirmButtonText: action.title,
          cancelButtonText: '取消'
        }
      )

      const result = await batchUpdateLegacySupplierClassStatus(
        selectedSupplierId.value,
        action.status
      )
      ElMessage.success(result.msg || `${action.title}完成`)
    } finally {
      statusBatching.value = false
    }
  }

  const handleSyncStatus = async () => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }

    await ElMessageBox.confirm(
      '将从上游重新核对商品状态，并自动处理本地已不存在的对接课程。是否继续？',
      '检查失效商品',
      { type: 'warning' }
    )
    const result = await syncLegacySupplierStatus(selectedSupplierId.value)
    ElMessage.success(result.msg || '检查完成')
    await reloadProducts()
  }

  const handleSearch = (params: {
    keyword?: string
    upstreamCategory?: string
    state?: string
  }) => {
    appliedSearch.keyword = params.keyword?.trim() || undefined
    appliedSearch.upstreamCategory = params.upstreamCategory || undefined
    appliedSearch.state = params.state || undefined
    pagination.current = 1
    updatePaginationTotal()
  }

  const handleReset = () => {
    appliedSearch.keyword = undefined
    appliedSearch.upstreamCategory = undefined
    appliedSearch.state = undefined
    pagination.current = 1
    updatePaginationTotal()
  }

  const handleSelectionChange = (rows: LegacyDockingProduct[]) => {
    const currentIds = new Set(pagedProducts.value.map((item) => String(item.cid)))
    const preserved = selectedProductIds.value.filter((id) => !currentIds.has(id))
    selectedProductIds.value = Array.from(
      new Set([...preserved, ...rows.map((item) => String(item.cid)).filter((id) => id.trim())])
    )
  }

  const handleCurrentChange = (current: number) => {
    pagination.current = current
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    pagination.current = 1
    updatePaginationTotal()
  }

  const openAddDialog = (record: LegacyDockingProduct) => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }

    currentAddProduct.value = record
    resetAddForm()
    addForm.name = record.name || ''
    addForm.getnoun = record.cid || ''
    addForm.noun = record.cid || ''
    addForm.content = record.content || ''
    addSortNumber.value = Number(record.sort || 10)
    applySuggestedPrice()
    addDialogVisible.value = true
  }

  const applySuggestedPrice = () => {
    const basePrice = normalizeNonNegativeNumber(currentAddProduct.value?.price)
    const rate = normalizePositiveNumber(addRate.value)
    const sort = normalizeNonNegativeNumber(addSortNumber.value)
    addForm.sort = String(sort ?? 10)
    if (basePrice === null || rate === null) {
      addForm.price = ''
      return
    }
    const price = basePrice * rate
    addForm.price = Number.isFinite(price) ? price.toFixed(2) : ''
  }

  const submitAddProduct = async () => {
    if (!addForm.name.trim()) {
      ElMessage.warning('请先填写课程名称')
      return
    }
    const price = normalizeNonNegativeNumber(addForm.price)
    if (price === null) {
      ElMessage.warning('请先填写售价')
      return
    }

    const sort = normalizeNonNegativeNumber(addSortNumber.value)
    addForm.sort = String(sort ?? 10)
    addSaving.value = true
    try {
      await addLegacyDockingClass({
        ...addForm,
        name: addForm.name.trim(),
        price: price.toFixed(2),
        getnoun: addForm.getnoun.trim(),
        noun: addForm.noun.trim(),
        content: addForm.content.trim()
      })
      if (currentAddProduct.value) {
        currentAddProduct.value.states = 1
      }
      ElMessage.success('商品已上架')
      addDialogVisible.value = false
    } finally {
      addSaving.value = false
    }
  }

  const openBatchDialog = () => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }
    if (!selectedProductIds.value.length) {
      ElMessage.warning('请先勾选要上架的商品')
      return
    }
    batchForm.rate = 1.1
    batchForm.sort = 10
    batchForm.fenlei = ''
    batchForm.skipExists = true
    batchDialogVisible.value = true
  }

  const submitBatchAdd = async () => {
    if (!batchCandidates.value.length) {
      ElMessage.warning('当前没有可上架的商品')
      return
    }
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }
    const rate = normalizePositiveNumber(batchForm.rate)
    if (rate === null) {
      ElMessage.warning('请先填写正确的价格倍率')
      return
    }
    const defaultSort = normalizeNonNegativeNumber(batchForm.sort) ?? 10

    batchRunning.value = true
    batchProgress.value = 0
    let successCount = 0
    let failCount = 0

    try {
      for (const [index, item] of batchCandidates.value.entries()) {
        const basePrice = normalizeNonNegativeNumber(item.price)
        if (basePrice === null) {
          failCount += 1
          batchProgress.value = Math.round(((index + 1) / batchCandidates.value.length) * 100)
          continue
        }
        const price = basePrice * rate
        const sort = normalizeNonNegativeNumber(item.sort) ?? defaultSort
        try {
          await addLegacyDockingClass({
            sort: String(sort),
            name: item.name,
            price: price.toFixed(2),
            getnoun: item.cid,
            noun: item.cid,
            content: item.content || '',
            queryplat: String(selectedSupplierId.value),
            docking: String(selectedSupplierId.value),
            yunsuan: '*',
            status: '1',
            fenlei: batchForm.fenlei
          })
          item.states = 1
          successCount += 1
        } catch {
          failCount += 1
        }
        batchProgress.value = Math.round(((index + 1) / batchCandidates.value.length) * 100)
      }

      ElMessage.success(`批量上架完成：成功 ${successCount} 个，失败 ${failCount} 个`)
      clearSelection()
      if (failCount === 0) {
        batchDialogVisible.value = false
      }
    } finally {
      batchRunning.value = false
    }
  }

  const openImportDialog = () => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }
    importForm.category = '999999'
    importForm.name = currentSupplier.value?.name || ''
    importForm.pricee = 1.1
    importForm.fd = 0
    importDialogVisible.value = true
  }

  const submitImport = async () => {
    if (!selectedSupplierId.value) {
      ElMessage.warning('请先选择货源')
      return
    }
    const pricee = normalizePositiveNumber(importForm.pricee)
    if (pricee === null) {
      ElMessage.warning('请先填写正确的价格倍率')
      return
    }

    importing.value = true
    try {
      const result = await importLegacySupplier({
        hid: selectedSupplierId.value,
        pricee,
        category: importForm.category,
        name: importForm.name.trim() || currentSupplier.value?.name || '',
        fd: importForm.fd
      })
      ElMessage.success(result.msg || '对接完成')
      importDialogVisible.value = false
      await reloadProducts()
    } finally {
      importing.value = false
    }
  }

  const createCategoryFor = async (target: CategoryAssignTarget) => {
    const name = newCategoryName.value.trim()
    if (!name) {
      ElMessage.warning('请先填写分类名称')
      return
    }

    categorySaving.value = true
    try {
      await saveLegacyAdminCategory({
        name,
        sort: 10,
        status: '1',
        recommend: 0,
        log: 0,
        ticket: 0,
        changepass: 1,
        allowpause: 0,
        supplier_report: 0,
        supplier_report_hid: 0
      })
      await loadBaseOptions()
      const created = categoryOptions.value.find((item) => item.name === name)
      if (created) {
        if (target === 'single') {
          addForm.fenlei = String(created.id)
        } else if (target === 'batch') {
          batchForm.fenlei = String(created.id)
        } else {
          importForm.name = created.name
        }
      }
      newCategoryName.value = ''
      ElMessage.success('分类已创建')
    } finally {
      categorySaving.value = false
    }
  }

  const normalizeScopePayload = (scope: BatchScope, scopeId: string) => ({
    scope,
    scope_id: scope === 'all' ? '' : scopeId.trim()
  })

  const handleReplaceKeyword = async () => {
    if (!replaceForm.search.trim()) {
      ElMessage.warning('请先填写原关键词')
      return
    }
    if (replaceForm.scope !== 'all' && !replaceForm.scopeId.trim()) {
      ElMessage.warning('请先填写范围值')
      return
    }

    replaceLoading.value = true
    try {
      const result = await batchReplaceLegacyClassKeyword({
        search: replaceForm.search.trim(),
        replace: replaceForm.replace.trim(),
        ...normalizeScopePayload(replaceForm.scope, replaceForm.scopeId)
      })
      ElMessage.success(result.msg || '批量替换完成')
    } finally {
      replaceLoading.value = false
    }
  }

  const handleAddPrefix = async () => {
    if (!prefixForm.prefix.trim()) {
      ElMessage.warning('请先填写前缀')
      return
    }
    if (prefixForm.scope !== 'all' && !prefixForm.scopeId.trim()) {
      ElMessage.warning('请先填写范围值')
      return
    }

    prefixLoading.value = true
    try {
      const result = await batchAddLegacyClassPrefix({
        prefix: prefixForm.prefix.trim(),
        ...normalizeScopePayload(prefixForm.scope, prefixForm.scopeId)
      })
      ElMessage.success(result.msg || '批量添加前缀完成')
    } finally {
      prefixLoading.value = false
    }
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

  const { columns, columnChecks } = useTableColumns<LegacyDockingProduct>(() => [

    {
      type: 'selection',
      width: 50,
      fixed: 'left',
      reserveSelection: true
    },
    {
      type: 'globalIndex',
      label: '序号',
      width: 70
    },
    {
      prop: 'name',
      label: '上游商品',
      minWidth: 280,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名商品'),
          h(
            'p',
            { class: 'mt-1 text-xs text-g-500 line-clamp-1' },
            row.content || row.category_name || '暂无商品说明'
          ),
          h('p', { class: 'mt-1 text-xs text-g-400' }, `CID ${row.cid} / 分类 ${row.fenlei || '-'}`)
        ])
    },
    {
      prop: 'price',
      label: '上游价格',
      width: 120,
      formatter: (row) =>
        h(
          'span',
          { class: 'font-semibold text-[var(--el-color-success)]' },
          `¥${Number(row.price || 0).toFixed(2)}`
        )
    },
    {
      prop: 'category_name',
      label: '上游分类',
      minWidth: 180,
      formatter: (row) => row.category_name || row.fenlei || '-'
    },
    {
      prop: 'states',
      label: '入库状态',
      width: 110,
      formatter: (row) =>
        h(ElTag, { type: row.states === 1 ? 'success' : 'info' }, () =>
          row.states === 1 ? '已入库' : '未入库'
        )
    },
    {
      prop: 'operation',
      label: '操作',
      minWidth: 180,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap items-center gap-2' }, [
          h(
            'button',
            {
              class:
                'rounded-md border border-[var(--el-color-primary-light-6)] bg-[var(--el-color-primary-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-primary)] transition hover:bg-[var(--el-color-primary-light-8)]',
              onClick: () => openAddDialog(row)
            },
            row.states === 1 ? '再次上架' : '单个上架'
          )
        ])
    }
  ])

  watch(
    () => filteredProducts.value.length,
    () => {
      updatePaginationTotal()
    }
  )

  watch(
    () => [pagination.current, pagination.size, pagedProducts.value.map((item) => item.cid).join(',')],
    async () => {
      await syncPageSelection()
    }
  )

  onMounted(async () => {
    await loadBaseOptions()
  })
</script>
