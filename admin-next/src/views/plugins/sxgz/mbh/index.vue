<template>
  <div class="plugin-sxgz-page art-full-height">
    <ArtSearchBar
      v-model="filters"
      class="sxgz-search-bar"
      :items="searchItems"
      :showExpand="false"
      @reset="resetFilters"
      @search="handleSearch"
    />

    <ElCard class="art-table-card" style="margin-top: 0">
      <template #header>
        <div class="flex-cb">
          <h4 class="m-0">实习盖章订单</h4>
          <div class="flex flex-wrap gap-2">
            <ElTag type="success" effect="light">当前页 {{ currentOrderCount }} 条</ElTag>
            <ElTag type="warning" effect="light">待处理 {{ pendingOrderCount }}</ElTag>
            <ElTag type="primary" effect="light">已完成 {{ completedOrderCount }}</ElTag>
            <ElTag type="info" effect="light">退款 {{ refundOrderCount }}</ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="orderLoading"
        layout="refresh,size,fullscreen,columns,settings"
        fullClass="art-table-card"
        @refresh="loadOrders(pagination.page)"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="openOrderDialog">新增订单</ElButton>
            <ElButton plain :loading="announcementLoading" @click="openAnnouncementDialog"
              >公告</ElButton
            >
            <ElButton v-if="isAdmin" plain @click="openConfigDialog">配置</ElButton>
            <ElButton v-if="isAdmin" plain :loading="syncLoading" @click="handleSyncOrders"
              >同步上游</ElButton
            >
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        rowKey="order_id"
        :columns="columns"
        :data="orders"
        empty-height="360px"
        :loading="orderLoading"
        :pagination="tablePagination"
        @pagination:current-change="loadOrders"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog v-model="orderVisible" title="新增订单" width="1180px" destroy-on-close>
      <div class="mx-auto max-w-[1180px]">
        <ElSteps :active="orderStep" align-center finish-status="success">
          <ElStep v-for="item in orderSteps" :key="item.key" :title="item.title" />
        </ElSteps>

        <div class="mt-8">
          <div class="sxgz-step-panel">
            <div class="mb-6">
              <div class="text-lg font-semibold text-g-900">{{ currentStepTitle }}</div>
              <div class="mt-2 text-sm text-g-500">{{ currentStepSubtitle }}</div>
            </div>

            <div
              v-if="currentStepKey === 'service'"
              class="grid gap-5 md:grid-cols-2 xl:grid-cols-4"
            >
              <button
                v-for="item in serviceModeOptions"
                :key="item.value"
                type="button"
                class="sxgz-service-card"
                :class="{ 'is-active': selectedServiceMode === item.value }"
                @click="selectServiceMode(item.value)"
              >
                <span class="sxgz-service-icon" :class="item.iconClass">{{ item.icon }}</span>
                <span class="mt-4 text-sm font-semibold text-g-900">{{ item.label }}</span>
                <span class="mt-2 text-xs text-g-500">{{ item.desc }}</span>
              </button>
            </div>

            <div v-else-if="currentStepKey === 'company'" class="space-y-5">
              <template v-if="selectedServiceMode === 'license_only'">
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">营业执照公司</p>
                  <ElSelect
                    v-model="orderForm.selected_license_companies"
                    class="w-full"
                    filterable
                    multiple
                    placeholder="请选择需要营业执照的公司，可多选"
                    @change="handleLicenseCompanyChange"
                  >
                    <ElOption
                      v-for="item in licenseCompanyOptions"
                      :key="item.cid"
                      :label="licenseCompanyLabel(item)"
                      :value="item.cid"
                    />
                  </ElSelect>
                </div>

                <div v-if="selectedLicenseCompanyRows.length" class="sxgz-selected-list">
                  <div
                    v-for="item in selectedLicenseCompanyRows"
                    :key="item.cid"
                    class="sxgz-selected-row"
                  >
                    <div class="min-w-0">
                      <div class="truncate text-sm font-semibold text-g-900">{{ item.name }}</div>
                      <div v-if="item.content" class="mt-1 truncate text-xs text-g-500">{{
                        item.content
                      }}</div>
                    </div>
                    <div class="shrink-0 text-sm font-semibold text-[var(--el-color-danger)]">
                      ¥{{ Number(item.license_price || item.price || 0).toFixed(2) }}
                    </div>
                  </div>
                </div>
              </template>

              <template v-else>
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">盖章公司</p>
                  <ElSelect
                    v-model="orderForm.company_id"
                    class="w-full"
                    filterable
                    placeholder="请选择公司，可输入关键词搜索"
                    @change="handleCompanyChange"
                  >
                    <ElOption
                      v-for="item in companies"
                      :key="item.cid"
                      :label="companyLabel(item)"
                      :value="item.cid"
                    />
                  </ElSelect>
                  <div v-if="!orderForm.company_id" class="mt-2 text-xs text-g-500">
                    先选盖章公司，再继续填写后续信息
                  </div>
                </div>

                <div v-if="selectedCompany" class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div class="min-w-0">
                      <div class="truncate font-semibold text-g-900">{{
                        selectedCompany.name
                      }}</div>
                      <div class="mt-1 text-xs text-g-500">
                        CID {{ selectedCompany.cid }} / {{ selectedCompany.source || 'cache' }}
                      </div>
                    </div>
                    <ElTag effect="plain" :type="selectedCompany.status ? 'success' : 'info'">
                      {{ selectedCompany.status ? '可用' : '缓存' }}
                    </ElTag>
                  </div>
                  <div v-if="selectedCompany.content" class="mt-3 text-sm leading-6 text-g-600">
                    {{ selectedCompany.content }}
                  </div>
                  <div class="mt-4 grid gap-3 text-sm md:grid-cols-2">
                    <div class="rounded-custom-sm bg-box px-3 py-2">
                      <span class="text-g-500">盖章价格</span>
                      <span class="ml-2 font-semibold text-g-900"
                        >¥{{ Number(selectedCompany.price || 0).toFixed(2) }}</span
                      >
                    </div>
                    <div class="rounded-custom-sm bg-box px-3 py-2">
                      <span class="text-g-500">执照价格</span>
                      <span class="ml-2 font-semibold text-g-900"
                        >¥{{ Number(selectedCompany.license_price || 0).toFixed(2) }}</span
                      >
                    </div>
                  </div>
                </div>

                <div class="rounded-custom-sm border-full-d bg-box p-4">
                  <div class="flex items-center justify-between gap-4">
                    <div>
                      <div class="text-sm font-semibold text-g-900">附加营业执照</div>
                      <div class="mt-1 text-xs text-g-500">默认跟随当前盖章公司</div>
                    </div>
                    <ElSwitch
                      v-model="orderForm.business_license"
                      @change="handleBusinessLicenseChange"
                    />
                  </div>
                  <div
                    v-if="orderForm.business_license"
                    class="mt-3 rounded-custom-sm bg-g-100/80 px-3 py-2 text-sm text-g-600"
                  >
                    将按当前盖章公司同步营业执照费用
                  </div>
                </div>
              </template>
            </div>

            <div v-else-if="currentStepKey === 'material'" class="space-y-5">
              <div class="grid gap-4 md:grid-cols-2">
                <button
                  type="button"
                  class="sxgz-option-card"
                  :class="{ 'is-active': orderForm.material_type === 'upload' }"
                  @click="setMaterialType('upload')"
                >
                  <span class="font-semibold text-g-900">在线上传文件</span>
                  <span class="mt-1 text-xs text-g-500">先创建订单，随后在订单列表上传附件</span>
                </button>
                <button
                  v-if="selectedServiceMode !== 'electronic'"
                  type="button"
                  class="sxgz-option-card"
                  :class="{ 'is-active': orderForm.material_type === 'mail' }"
                  @click="setMaterialType('mail')"
                >
                  <span class="font-semibold text-g-900">邮寄纸质文件</span>
                  <span class="mt-1 text-xs text-g-500">填写寄到工作室的快递信息</span>
                </button>
              </div>

              <div
                v-if="orderForm.material_type === 'mail' && selectedServiceMode !== 'license_only'"
                class="space-y-3"
              >
                <p class="text-sm font-medium text-g-800">盖章方式</p>
                <div class="grid gap-4 md:grid-cols-3">
                  <button
                    v-for="item in stampOptions"
                    :key="item"
                    type="button"
                    class="sxgz-option-card"
                    :class="{ 'is-active': mailStampType === item }"
                    @click="mailStampType = item"
                  >
                    <span class="font-semibold text-g-900">{{ item }}</span>
                    <span class="mt-1 text-xs text-g-500">{{ stampDescription(item) }}</span>
                  </button>
                </div>
              </div>

              <div
                v-if="orderForm.material_type === 'upload'"
                class="rounded-custom-sm border-full-d bg-box p-4"
              >
                <div class="mb-3 flex items-center justify-between gap-3">
                  <div>
                    <div class="text-sm font-semibold text-g-900">上传文件</div>
                    <div class="mt-1 text-xs text-g-500">先选文件，确认下单后会自动挂到新订单</div>
                  </div>
                  <span v-if="pendingMaterialFiles.length" class="text-xs text-g-500"
                    >已选 {{ pendingMaterialFiles.length }} 个</span
                  >
                </div>
                <ElUpload
                  ref="materialUploadRef"
                  :auto-upload="false"
                  :before-upload="beforePendingMaterialFileUpload"
                  :limit="10"
                  :multiple="true"
                  :on-change="handlePendingMaterialFileChange"
                  :on-exceed="handleMaterialFileExceed"
                  :show-file-list="false"
                  accept=".pdf,.doc,.docx,.zip,.rar,.7z,.jpg,.jpeg,.png,.gif"
                  drag
                >
                  <div class="py-10 text-center">
                    <div class="text-sm font-medium text-g-800">拖拽文件到这里或点击选择</div>
                    <div class="mt-1 text-xs text-g-500">支持 PDF、Word、图片和压缩包</div>
                  </div>
                </ElUpload>
                <div v-if="pendingMaterialFiles.length" class="mt-4 space-y-3">
                  <div
                    v-for="(file, index) in pendingMaterialFiles"
                    :key="file.uid"
                    class="rounded-custom-sm border-full-d bg-g-100/60 p-3"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <div class="truncate text-sm font-medium text-g-900">{{ file.name }}</div>
                        <div class="mt-1 text-xs text-g-500">
                          {{ formatFileSize(file.size) }} / 约 {{ filePrintedSheets(file) }} 张
                        </div>
                      </div>
                      <ElButton text type="danger" @click="handlePendingMaterialFileRemove(index)"
                        >删除</ElButton
                      >
                    </div>
                    <div class="mt-3 grid gap-3 md:grid-cols-3 xl:grid-cols-5">
                      <div>
                        <p class="mb-1 text-xs text-g-500">页数/份</p>
                        <ElInputNumber
                          v-model="file.pageCount"
                          class="w-full"
                          :min="1"
                          :max="1000"
                        />
                      </div>
                      <div>
                        <p class="mb-1 text-xs text-g-500">份数</p>
                        <ElInputNumber
                          v-model="file.printOptions.printCount"
                          class="w-full"
                          :min="1"
                          :max="100"
                        />
                      </div>
                      <div>
                        <p class="mb-1 text-xs text-g-500">打印方式</p>
                        <ElSelect v-model="file.printOptions.printMode" class="w-full">
                          <ElOption label="单面打印" value="单面打印" />
                          <ElOption label="双面打印" value="双面打印" />
                        </ElSelect>
                      </div>
                      <div>
                        <p class="mb-1 text-xs text-g-500">颜色</p>
                        <ElSelect v-model="file.printOptions.colorMode" class="w-full">
                          <ElOption label="黑白" value="黑白" />
                          <ElOption label="彩印" value="彩印" />
                        </ElSelect>
                      </div>
                      <div>
                        <p class="mb-1 text-xs text-g-500">纸张</p>
                        <ElSelect v-model="file.printOptions.paperSize" class="w-full">
                          <ElOption label="A4纸" value="A4纸" />
                          <ElOption label="A3纸" value="A3纸" />
                        </ElSelect>
                      </div>
                      <div>
                        <p class="mb-1 text-xs text-g-500">盖章</p>
                        <ElSelect v-model="file.printOptions.stampType" class="w-full">
                          <ElOption label="实体章" value="实体章" />
                          <ElOption label="骑缝章" value="骑缝章" />
                          <ElOption label="实体章+骑缝章" value="实体章+骑缝章" />
                        </ElSelect>
                      </div>
                    </div>
                  </div>
                  <div
                    class="flex flex-wrap items-center justify-between gap-3 rounded-custom-sm bg-g-100/80 px-3 py-2 text-xs text-g-500"
                  >
                    <span
                      >下单后会自动上传这些文件，预计打印 {{ pendingFilePrintSheets }} 张。</span
                    >
                    <ElButton text @click="clearPendingMaterialFiles">清空文件</ElButton>
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="currentStepKey === 'contact'" class="space-y-5">
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">客户姓名</p>
                  <ElInput
                    v-model="orderForm.customer_name"
                    placeholder="请输入客户姓名"
                    @input="scheduleQuote"
                  />
                </div>
                <div v-if="needsEmail">
                  <p class="mb-2 text-sm font-medium text-g-800">接收邮箱</p>
                  <ElInput
                    v-model="orderForm.customer_email"
                    placeholder="请输入接收邮箱"
                    @input="scheduleQuote"
                  />
                </div>
                <div v-if="needsPhone">
                  <p class="mb-2 text-sm font-medium text-g-800">联系电话</p>
                  <ElInput
                    v-model="orderForm.customer_phone"
                    placeholder="请输入联系电话"
                    @input="scheduleQuote"
                  />
                </div>
                <div v-if="needsDelivery">
                  <p class="mb-2 text-sm font-medium text-g-800">回货方式</p>
                  <ElSelect
                    v-model="orderForm.delivery_option"
                    class="w-full"
                    @change="refreshQuote"
                  >
                    <ElOption
                      v-for="item in deliveryOptionKeys"
                      :key="item"
                      :label="item"
                      :value="item"
                    />
                  </ElSelect>
                </div>
              </div>
              <div v-if="needsAddress">
                <p class="mb-2 text-sm font-medium text-g-800">收件地址</p>
                <ElInput
                  v-model="orderForm.customer_address"
                  :rows="3"
                  placeholder="请输入详细收件地址"
                  type="textarea"
                  @input="scheduleQuote"
                />
              </div>
              <div v-if="needsCourier" class="grid gap-4 md:grid-cols-2">
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">寄到工作室快递公司</p>
                  <ElInput
                    v-model="orderForm.courier_company"
                    placeholder="例如顺丰、中通"
                    @input="scheduleQuote"
                  />
                </div>
                <div>
                  <p class="mb-2 text-sm font-medium text-g-800">寄到工作室快递单号</p>
                  <ElInput
                    v-model="orderForm.tracking_number"
                    placeholder="请输入快递单号"
                    @input="scheduleQuote"
                  />
                </div>
              </div>
              <div>
                <p class="mb-2 text-sm font-medium text-g-800">备注说明</p>
                <ElInput
                  v-model="orderForm.special_requirements"
                  :rows="3"
                  placeholder="请输入特殊要求"
                  type="textarea"
                  @input="scheduleQuote"
                />
              </div>
            </div>

            <div v-else>
              <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                <div class="text-sm font-semibold text-g-900">确认费用后提交</div>
                <div class="mt-2 text-sm leading-6 text-g-600">
                  订单将按前面填写的服务、公司、材料和收件信息创建。若本次选择了上传文件，订单创建成功后会自动上传并保存每个文件的打印参数。
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="currentStepKey === 'confirm'"
            class="mt-6 rounded-custom-sm border-full-d bg-g-100/60 p-4"
          >
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm text-g-500">费用预览</span>
              <span class="text-xl font-semibold text-[var(--el-color-danger)]"
                >¥{{ quote.total_price.toFixed(2) }}</span
              >
            </div>
            <div class="mt-4 grid gap-3 text-sm md:grid-cols-2 xl:grid-cols-4">
              <div class="flex items-center justify-between">
                <span class="text-g-500">基础费用</span>
                <span>¥{{ quote.base_price.toFixed(2) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-g-500">打印费用</span>
                <span>¥{{ quote.print_price.toFixed(2) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-g-500">营业执照</span>
                <span>¥{{ quote.license_price.toFixed(2) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-g-500">附加项</span>
                <span>¥{{ quote.extra_options_price.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <ElButton :disabled="orderStep === 0" @click="prevOrderStep">上一步</ElButton>
          <ElButton
            v-if="orderStep < confirmStepIndex"
            type="primary"
            :disabled="!canGoNext"
            @click="nextOrderStep"
          >
            下一步
          </ElButton>
          <ElButton v-else type="primary" :loading="createLoading" @click="handleCreateOrder"
            >确认下单</ElButton
          >
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="configVisible" title="配置" width="1180px" destroy-on-close>
      <div class="space-y-5">
        <div class="flex flex-wrap justify-end gap-3">
          <ElButton plain :loading="configLoading" @click="loadConfig">重新读取</ElButton>
          <ElButton plain :loading="refreshCompanyLoading" @click="handleRefreshCompanies"
            >刷新上游</ElButton
          >
          <ElButton type="primary" :loading="saveConfigLoading" @click="handleSaveConfig"
            >保存配置</ElButton
          >
        </div>

        <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">上游接入</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                维护上游地址、账号凭据、协议类型和文件回传域名。
              </p>
            </div>

            <div class="mt-5 grid gap-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">对接协议</label>
                <ElSelect v-model="configForm.upstream_protocol" class="w-full">
                  <ElOption label="源台（29系统）" value="source29" />
                  <ElOption label="同系统（Go/OpenAPI）" value="same_system" />
                </ElSelect>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">上游地址</label>
                <ElInput
                  v-model="configForm.upstream_url"
                  placeholder="例如 http://127.0.0.1:8080"
                />
              </div>
              <div class="grid gap-4 md:grid-cols-[180px_1fr]">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">上游 UID</label>
                  <ElInputNumber
                    v-model="configForm.upstream_uid"
                    class="w-full"
                    :min="0"
                    :precision="0"
                  />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">上游 Key</label>
                  <ElInput v-model="configForm.upstream_key" show-password />
                </div>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">文件回传域名</label>
                <ElInput v-model="configForm.file_base_url" placeholder="选填" />
              </div>
            </div>
          </section>

          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">价格与同步</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                公司入库时先应用后台倍率，列表和下单时再叠加用户倍率。
              </p>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
                <ElInputNumber
                  v-model="configForm.price_multiplier"
                  class="w-full"
                  :min="0"
                  :step="0.1"
                />
              </div>

              <div class="grid gap-4 md:grid-cols-[1fr_140px]">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">同步间隔(秒)</label>
                  <ElInputNumber
                    v-model="configForm.sync_interval"
                    class="w-full"
                    :min="30"
                    :step="30"
                    :precision="0"
                  />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">自动同步</label>
                  <div class="flex h-8 items-center">
                    <ElSwitch v-model="configForm.auto_sync" />
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">打印计费</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">
              用于上传材料时按文件打印参数计算打印费用。
            </p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-3">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">基础免费份数</label>
              <ElInputNumber
                v-model="configForm.print_pricing.base_free_copies"
                class="w-full"
                :min="0"
                :precision="0"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">每份打印单价</label>
              <ElInputNumber
                v-model="configForm.print_pricing.per_copy_price"
                class="w-full"
                :min="0"
                :step="0.1"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">额外打印单价</label>
              <ElInputNumber
                v-model="configForm.print_pricing.extra_copy_price"
                class="w-full"
                :min="0"
                :step="0.1"
              />
            </div>
          </div>
        </section>

        <div class="grid gap-5 xl:grid-cols-2">
          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">打印选项</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                控制实体章、骑缝章等打印附加项是否参与用户倍率。
              </p>
            </div>

            <div v-if="printOptionCount" class="mt-5 space-y-3">
              <div
                v-for="(rule, key) in configForm.print_options"
                :key="key"
                class="grid gap-3 rounded-custom-sm border-full-d bg-g-100/60 p-3 md:grid-cols-[minmax(0,1fr)_150px_120px] md:items-center"
              >
                <ElInput :model-value="key" disabled />
                <ElInputNumber
                  v-model="rule.price"
                  class="w-full"
                  controls-position="right"
                  :min="0"
                  :step="0.1"
                />
                <div class="flex items-center justify-end gap-2">
                  <span class="w-8 text-xs text-g-500">{{
                    rule.affected_by_user_rate ? '倍率' : '固定'
                  }}</span>
                  <ElSwitch v-model="rule.affected_by_user_rate" />
                </div>
              </div>
            </div>
            <div
              v-else
              class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm text-g-500"
            >
              暂无打印选项配置。
            </div>
          </section>

          <section class="rounded-custom-sm border-full-d bg-box p-5">
            <div class="border-b-d pb-4">
              <h3 class="text-lg font-semibold text-g-900">配送选项</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">
                控制顺丰、普通快递等配送费用是否参与用户倍率。
              </p>
            </div>

            <div v-if="deliveryOptionCount" class="mt-5 space-y-3">
              <div
                v-for="(rule, key) in configForm.delivery_options"
                :key="key"
                class="grid gap-3 rounded-custom-sm border-full-d bg-g-100/60 p-3 md:grid-cols-[minmax(0,1fr)_150px_120px] md:items-center"
              >
                <ElInput :model-value="key" disabled />
                <ElInputNumber
                  v-model="rule.price"
                  class="w-full"
                  controls-position="right"
                  :min="0"
                  :step="0.1"
                />
                <div class="flex items-center justify-end gap-2">
                  <span class="w-8 text-xs text-g-500">{{
                    rule.affected_by_user_rate ? '倍率' : '固定'
                  }}</span>
                  <ElSwitch v-model="rule.affected_by_user_rate" />
                </div>
              </div>
            </div>
            <div
              v-else
              class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm text-g-500"
            >
              暂无配送选项配置。
            </div>
          </section>
        </div>
      </div>
    </ElDialog>

    <ElDialog
      v-model="announcementVisible"
      title="公告"
      width="860px"
      destroy-on-close
      @closed="stopAnnouncementTimer"
    >
      <div class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <ElTag effect="plain">全站公告</ElTag>
          <ElSpace wrap>
            <ElButton plain :loading="announcementLoading" @click="() => loadAnnouncements()"
              >刷新</ElButton
            >
          </ElSpace>
        </div>

        <div class="max-h-[520px] space-y-3 overflow-auto pr-1">
          <div
            v-if="announcementLoading && !announcements.length"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-6 text-center text-sm text-g-500"
          >
            正在加载公告...
          </div>
          <div
            v-else-if="!announcements.length"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-6 text-center text-sm text-g-500"
          >
            暂无公告
          </div>
          <article
            v-for="item in announcements"
            :key="item.AID"
            class="rounded-custom-sm border-full-d bg-box p-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="truncate text-sm font-semibold text-g-900">{{ item.Title || '-' }}</h3>
                  <ElTag v-if="item.Importance === 5" size="small" type="warning" effect="plain"
                    >置顶</ElTag
                  >
                </div>
                <div class="mt-1 text-xs text-g-500">
                  {{ item.PublishDate || '-' }} · AID {{ item.AID || '-' }}
                </div>
              </div>
            </div>
            <div
              class="mt-3 rounded-custom-sm border-full-d bg-g-100/60 p-3 text-sm leading-6 text-g-600"
              v-html="item.Content || '暂无内容'"
            />
          </article>
        </div>

        <div class="flex items-center justify-between gap-3">
          <span class="text-xs text-g-500">共 {{ announcementPagination.total }} 条</span>
          <ElSpace wrap>
            <ElButton
              :disabled="announcementPagination.page <= 1"
              @click="changeAnnouncementPage(-1)"
              >上一页</ElButton
            >
            <span class="text-sm text-g-500">第 {{ announcementPagination.page }} 页</span>
            <ElButton
              :disabled="!announcementPagination.hasMore"
              @click="changeAnnouncementPage(1)"
            >
              下一页
            </ElButton>
          </ElSpace>
        </div>
      </div>
    </ElDialog>

    <ElDialog
      v-model="uploadVisible"
      :title="`上传附件 - ${selectedOrder?.order_no || ''}`"
      width="680px"
    >
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm text-g-600">
          先选择文件并设置每个文件的打印参数，再上传到当前订单，支持多次追加。
        </div>
        <ElUpload
          ref="orderUploadRef"
          :auto-upload="false"
          :before-upload="beforePendingMaterialFileUpload"
          :limit="10"
          :multiple="true"
          :on-change="handleOrderUploadFileChange"
          :on-exceed="handleMaterialFileExceed"
          :show-file-list="false"
          accept=".pdf,.doc,.docx,.zip,.rar,.7z,.jpg,.jpeg,.png,.gif"
          drag
        >
          <div class="py-10 text-center">
            <div class="text-sm font-medium text-g-800">拖拽文件到这里或点击选择</div>
            <div class="mt-1 text-xs text-g-500">支持 PDF、Word、压缩包和图片</div>
          </div>
        </ElUpload>

        <div v-if="orderUploadFiles.length" class="space-y-3">
          <div
            v-for="(file, index) in orderUploadFiles"
            :key="file.uid"
            class="rounded-custom-sm border-full-d bg-g-100/60 p-3"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="truncate text-sm font-medium text-g-900">{{ file.name }}</div>
                <div class="mt-1 text-xs text-g-500">
                  {{ formatFileSize(file.size) }} / 约 {{ filePrintedSheets(file) }} 张
                </div>
              </div>
              <ElButton text type="danger" @click="handleOrderUploadFileRemove(index)"
                >删除</ElButton
              >
            </div>
            <div class="mt-3 grid gap-3 md:grid-cols-3">
              <div>
                <p class="mb-1 text-xs text-g-500">页数/份</p>
                <ElInputNumber v-model="file.pageCount" class="w-full" :min="1" :max="1000" />
              </div>
              <div>
                <p class="mb-1 text-xs text-g-500">份数</p>
                <ElInputNumber
                  v-model="file.printOptions.printCount"
                  class="w-full"
                  :min="1"
                  :max="100"
                />
              </div>
              <div>
                <p class="mb-1 text-xs text-g-500">打印方式</p>
                <ElSelect v-model="file.printOptions.printMode" class="w-full">
                  <ElOption label="单面打印" value="单面打印" />
                  <ElOption label="双面打印" value="双面打印" />
                </ElSelect>
              </div>
              <div>
                <p class="mb-1 text-xs text-g-500">颜色</p>
                <ElSelect v-model="file.printOptions.colorMode" class="w-full">
                  <ElOption label="黑白" value="黑白" />
                  <ElOption label="彩印" value="彩印" />
                </ElSelect>
              </div>
              <div>
                <p class="mb-1 text-xs text-g-500">纸张</p>
                <ElSelect v-model="file.printOptions.paperSize" class="w-full">
                  <ElOption label="A4纸" value="A4纸" />
                  <ElOption label="A3纸" value="A3纸" />
                </ElSelect>
              </div>
              <div>
                <p class="mb-1 text-xs text-g-500">盖章</p>
                <ElSelect v-model="file.printOptions.stampType" class="w-full">
                  <ElOption label="实体章" value="实体章" />
                  <ElOption label="骑缝章" value="骑缝章" />
                  <ElOption label="实体章+骑缝章" value="实体章+骑缝章" />
                </ElSelect>
              </div>
            </div>
          </div>
          <div class="flex justify-end gap-3">
            <ElButton @click="clearOrderUploadFiles">清空</ElButton>
            <ElButton type="primary" :loading="uploadSaving" @click="handleUploadSelectedFiles">
              开始上传
            </ElButton>
          </div>
        </div>

        <div
          v-if="selectedFiles.uploaded.length || selectedFiles.processed.length"
          class="grid gap-4 md:grid-cols-2"
        >
          <div class="rounded-custom-sm border-full-d bg-box p-4">
            <p class="mb-2 text-sm font-medium text-g-800">已上传文件</p>
            <div class="space-y-2 text-sm text-g-600">
              <div
                v-for="item in selectedFiles.uploaded"
                :key="item.url"
                class="flex items-center justify-between gap-3"
              >
                <span class="truncate">{{ item.name }}</span>
                <span>{{ formatFileSize(item.size) }}</span>
              </div>
            </div>
          </div>
          <div class="rounded-custom-sm border-full-d bg-box p-4">
            <p class="mb-2 text-sm font-medium text-g-800">已回传文件</p>
            <div class="space-y-2 text-sm text-g-600">
              <div
                v-for="item in selectedFiles.processed"
                :key="item.url"
                class="flex items-center justify-between gap-3"
              >
                <span class="truncate">{{ item.name }}</span>
                <span>{{ formatFileSize(item.size) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="adminVisible" title="更新订单状态" width="560px">
      <div class="space-y-4">
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">状态</p>
          <ElSelect v-model="adminForm.status" class="w-full">
            <ElOption
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">管理员备注</p>
          <ElInput v-model="adminForm.admin_notes" :rows="3" type="textarea" />
        </div>
        <div>
          <p class="mb-2 text-sm font-medium text-g-800">退款原因</p>
          <ElInput v-model="adminForm.refund_reason" :rows="3" type="textarea" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="adminVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="adminSaving" @click="handleSaveAdminOrder"
            >保存</ElButton
          >
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElMessage, ElMessageBox, ElUpload, ElTag, type UploadFile } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import {
    applyLegacySXGZRefund,
    fetchLegacySXGZAnnouncements,
    createLegacySXGZOrder,
    fetchLegacySXGZAdminOrders,
    fetchLegacySXGZCompanies,
    fetchLegacySXGZConfig,
    fetchLegacySXGZLicenseCompanies,
    fetchLegacySXGZOrders,
    quoteLegacySXGZOrder,
    refreshLegacySXGZCompanies,
    saveLegacySXGZConfig,
    syncLegacySXGZOrders,
    updateLegacySXGZAdminOrder,
    uploadLegacySXGZOrderFile,
    type LegacySXGZAnnouncement,
    type LegacySXGZAnnouncementListResult,
    type LegacySXGZCompany,
    type LegacySXGZConfig,
    type LegacySXGZFileRecord,
    type LegacySXGZFilePrintOption,
    type LegacySXGZOrder,
    type LegacySXGZQuoteRequest,
    type LegacySXGZQuoteResult
  } from '@/api/legacy/plugin-sxgz'

  defineOptions({ name: 'PluginSXGZMBHPage' })

  type ServiceMode = 'both' | 'electronic' | 'license_only' | 'mail'
  type OrderStepKey = 'company' | 'confirm' | 'contact' | 'material' | 'service'
  type FilePrintMode = '单面打印' | '双面打印'
  type FileColorMode = '黑白' | '彩印'
  type FilePaperSize = 'A4纸' | 'A3纸'
  type FileStampType = '实体章' | '骑缝章' | '实体章+骑缝章'

  interface PendingMaterialFilePrintOptions {
    colorMode: FileColorMode
    paperSize: FilePaperSize
    printCount: number
    printMode: FilePrintMode
    stampType: FileStampType
  }

  interface PendingMaterialFile {
    name: string
    pageCount: number
    printOptions: PendingMaterialFilePrintOptions
    raw: File
    size: number
    uid: string
  }

  const userStore = useUserStore()
  const companies = ref<LegacySXGZCompany[]>([])
  const licenseCompanies = ref<LegacySXGZCompany[]>([])
  const orders = ref<LegacySXGZOrder[]>([])
  const orderLoading = ref(false)
  const companyLoading = ref(false)
  const refreshCompanyLoading = ref(false)
  const quoteLoading = ref(false)
  const createLoading = ref(false)
  const syncLoading = ref(false)
  const configLoading = ref(false)
  const saveConfigLoading = ref(false)
  const adminSaving = ref(false)
  const orderVisible = ref(false)
  const configVisible = ref(false)
  const announcementVisible = ref(false)
  const announcementLoading = ref(false)
  const uploadVisible = ref(false)
  const uploadSaving = ref(false)
  const adminVisible = ref(false)
  const selectedOrder = ref<LegacySXGZOrder | null>(null)
  const orderStep = ref(0)
  const selectedServiceMode = ref<ServiceMode>('electronic')
  const mailStampType = ref('实体章')
  const materialUploadRef = ref<InstanceType<typeof ElUpload> | null>(null)
  const orderUploadRef = ref<InstanceType<typeof ElUpload> | null>(null)
  const pendingMaterialFiles = ref<PendingMaterialFile[]>([])
  const orderUploadFiles = ref<PendingMaterialFile[]>([])
  const selectedFiles = reactive({
    processed: [] as LegacySXGZFileRecord[],
    uploaded: [] as LegacySXGZFileRecord[]
  })
  const announcements = ref<LegacySXGZAnnouncement[]>([])
  const announcementTimer = ref<number | null>(null)
  const announcementPagination = reactive({
    page: 1,
    pageSize: 8,
    total: 0,
    hasMore: false,
    type: '全站公告'
  })

  const quote = reactive<LegacySXGZQuoteResult>({
    base_price: 0,
    company_name: '',
    extra_options_price: 0,
    license_price: 0,
    price_multiplier: 1,
    print_price: 0,
    total_price: 0,
    user_rate: 1
  })

  const pagination = reactive({
    page: 1,
    size: 10,
    total: 0
  })

  const filters = reactive({
    search: '',
    status: ''
  })

  const orderForm = reactive({
    business_license: false,
    company_id: null as number | null,
    courier_company: '',
    customer_address: '',
    customer_email: '',
    customer_name: '',
    customer_phone: '',
    delivery_option: '',
    material_type: 'upload',
    only_business_license: false,
    paper_size: 'A4',
    print_copies: 0,
    print_options: [] as string[],
    return_tracking_number: '',
    selected_license_companies: [] as number[],
    service_type: 'electronic',
    special_requirements: '',
    tracking_number: ''
  })

  const configForm = reactive<LegacySXGZConfig>({
    auto_sync: true,
    delivery_options: {},
    file_base_url: '',
    price_multiplier: 1,
    print_options: {},
    print_pricing: {
      base_free_copies: 10,
      extra_copy_price: 0.5,
      per_copy_price: 2
    },
    sync_interval: 300,
    upstream_key: '',
    upstream_protocol: 'source29',
    upstream_uid: 0,
    upstream_url: ''
  })

  const adminForm = reactive({
    admin_notes: '',
    order_id: 0,
    refund_reason: '',
    status: 'pending'
  })

  const tablePagination = computed(() => ({
    current: pagination.page,
    size: pagination.size,
    total: pagination.total
  }))

  const currentOrderCount = computed(() => orders.value.length)
  const pendingOrderCount = computed(
    () => orders.value.filter((item) => item.status === 'pending').length
  )
  const completedOrderCount = computed(
    () =>
      orders.value.filter((item) => item.status === 'completed' || item.status === 'delivered')
        .length
  )
  const refundOrderCount = computed(
    () =>
      orders.value.filter(
        (item) => item.status === 'refund_requested' || item.status === 'refunded'
      ).length
  )

  const isAdmin = computed(() => {
    const roles = userStore.info?.roles || []
    return roles.includes('R_ADMIN') || roles.includes('R_SUPER')
  })

  const deliveryOptionKeys = computed(() => Object.keys(configForm.delivery_options || {}))
  const printOptionCount = computed(() => Object.keys(configForm.print_options || {}).length)
  const deliveryOptionCount = computed(() => Object.keys(configForm.delivery_options || {}).length)
  const licenseCompanyOptions = computed(() =>
    (licenseCompanies.value.length ? licenseCompanies.value : companies.value).filter(
      (item) => Number(item.license_price || item.price || 0) > 0
    )
  )
  const selectedCompany = computed(() =>
    companies.value.find((item) => item.cid === orderForm.company_id)
  )
  const selectedLicenseCompanyRows = computed(
    () =>
      orderForm.selected_license_companies
        .map((id) => licenseCompanyOptions.value.find((item) => item.cid === id))
        .filter(Boolean) as LegacySXGZCompany[]
  )
  const needsEmail = computed(
    () =>
      selectedServiceMode.value === 'electronic' ||
      selectedServiceMode.value === 'both' ||
      selectedServiceMode.value === 'license_only'
  )
  const needsPhone = computed(
    () => selectedServiceMode.value === 'mail' || selectedServiceMode.value === 'both'
  )
  const needsAddress = computed(
    () => selectedServiceMode.value === 'mail' || selectedServiceMode.value === 'both'
  )
  const needsDelivery = computed(
    () =>
      deliveryOptionKeys.value.length > 0 &&
      (selectedServiceMode.value === 'mail' || selectedServiceMode.value === 'both')
  )
  const needsCourier = computed(
    () => selectedServiceMode.value !== 'license_only' && orderForm.material_type === 'mail'
  )
  const needsStampStep = computed(
    () => selectedServiceMode.value !== 'license_only' && orderForm.material_type === 'mail'
  )
  const hasPrintSettings = computed(
    () =>
      selectedServiceMode.value !== 'electronic' &&
      selectedServiceMode.value !== 'license_only' &&
      orderForm.material_type === 'upload'
  )
  const orderSteps = computed<Array<{ key: OrderStepKey; subtitle: string; title: string }>>(() => {
    const labels: Array<{ key: OrderStepKey; subtitle: string; title: string }> = [
      { key: 'service', title: '服务类型', subtitle: '先确定要做电子版、邮寄版还是仅营业执照' },
      {
        key: 'company',
        title: selectedServiceMode.value === 'license_only' ? '选择执照' : '选择公司',
        subtitle:
          selectedServiceMode.value === 'license_only'
            ? '选择需要营业执照的公司，可一次选择多个'
            : '选择盖章公司，需要执照时在本步单独勾选'
      }
    ]
    if (selectedServiceMode.value !== 'license_only') {
      labels.push({
        key: 'material',
        title: '材料处理',
        subtitle: '选择在线上传或邮寄纸质材料，再填写打印要求'
      })
    }
    labels.push({
      key: 'contact',
      title:
        selectedServiceMode.value === 'license_only' || selectedServiceMode.value === 'electronic'
          ? '接收信息'
          : '回货信息',
      subtitle:
        selectedServiceMode.value === 'license_only'
          ? '填写客户姓名和接收营业执照的邮箱'
          : selectedServiceMode.value === 'electronic'
            ? '填写客户姓名和接收电子文件的邮箱'
            : '填写收件人、地址和必要的快递信息'
    })
    labels.push({ key: 'confirm', title: '确认订单', subtitle: '核对服务、联系人和费用后提交' })
    return labels
  })
  const currentStep = computed(() => orderSteps.value[orderStep.value] || orderSteps.value[0])
  const currentStepKey = computed(() => currentStep.value?.key || 'service')
  const confirmStepIndex = computed(() => orderSteps.value.length - 1)
  const currentStepTitle = computed(() => currentStep.value?.title || '')
  const currentStepSubtitle = computed(() => currentStep.value?.subtitle || '')
  const serviceModeOptions = [
    {
      value: 'mail',
      label: '邮寄服务',
      desc: '快递寄送到指定地址',
      icon: '纸',
      iconClass: 'is-mail'
    },
    {
      value: 'electronic',
      label: '电子版',
      desc: '仅提供电子文件',
      icon: '电',
      iconClass: 'is-electronic'
    },
    {
      value: 'both',
      label: '邮寄+电子',
      desc: '电子文件 + 实物邮寄',
      icon: '邮',
      iconClass: 'is-both'
    },
    {
      value: 'license_only',
      label: '仅需营业执照',
      desc: '只需要营业执照',
      icon: '执',
      iconClass: 'is-license'
    }
  ] as const
  const stampOptions = ['实体章', '骑缝章', '实体章+骑缝章']
  const defaultFilePrintOptions = (): PendingMaterialFilePrintOptions => ({
    colorMode: '黑白',
    paperSize: 'A4纸',
    printCount: 1,
    printMode: '单面打印',
    stampType: '实体章'
  })
  const statusOptions = [
    { label: '待处理', value: 'pending' },
    { label: '处理中', value: 'processing' },
    { label: '已完成', value: 'completed' },
    { label: '已送达', value: 'delivered' },
    { label: '已取消', value: 'cancelled' },
    { label: '失败', value: 'failed' },
    { label: '退款申请', value: 'refund_requested' },
    { label: '已退款', value: 'refunded' }
  ]

  const searchItems = computed(() => [
    {
      key: 'search',
      label: '关键词',
      props: {
        clearable: true,
        placeholder: '订单号、客户名或公司名'
      },
      type: 'input'
    },
    {
      key: 'status',
      label: '状态',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: statusOptions
      },
      type: 'select'
    }
  ])

  const serviceTypeLabel = (value: string) => {
    if (value === 'license_only') return '仅需营业执照'
    if (value === 'mail') return '邮寄版'
    if (value === 'both') return '电子版 + 邮寄版'
    return '电子版'
  }

  const companyLabel = (item: LegacySXGZCompany) => {
    return `${item.name} / 普通${Number(item.price || 0).toFixed(2)} / 执照${Number(item.license_price || 0).toFixed(2)}`
  }

  const licenseCompanyLabel = (item: LegacySXGZCompany) => {
    return `${item.name} / 执照${Number(item.license_price || item.price || 0).toFixed(2)}`
  }

  const formatFileSize = (size?: number) => {
    const value = Number(size || 0)
    if (!value) return '0 B'
    if (value < 1024) return `${value} B`
    if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
    return `${(value / 1024 / 1024).toFixed(1)} MB`
  }

  const filePrintedSheets = (file: PendingMaterialFile) => {
    const pages = Math.max(1, Number(file.pageCount || 1))
    const copies = Math.max(1, Number(file.printOptions.printCount || 1))
    const sheetsPerCopy = file.printOptions.printMode === '双面打印' ? Math.ceil(pages / 2) : pages
    return sheetsPerCopy * copies
  }

  const pendingFilePrintSheets = computed(() =>
    pendingMaterialFiles.value.reduce((sum, file) => sum + filePrintedSheets(file), 0)
  )

  const pendingFilePrintOptionNames = computed(() => {
    const options = new Set<string>()
    pendingMaterialFiles.value.forEach((file) => {
      const item = file.printOptions
      ;[item.printMode, item.colorMode, item.paperSize, item.stampType]
        .filter(Boolean)
        .forEach((value) => options.add(value))
    })
    return Array.from(options)
  })

  const buildFilePrintPayload = (file: PendingMaterialFile): LegacySXGZFilePrintOption => ({
    color_mode: file.printOptions.colorMode,
    name: file.name,
    page_count: Math.max(1, Number(file.pageCount || 1)),
    paper_size: file.printOptions.paperSize,
    print_count: Math.max(1, Number(file.printOptions.printCount || 1)),
    print_mode: file.printOptions.printMode,
    size: file.size,
    stamp_type: file.printOptions.stampType
  })

  const filePrintRequirementText = computed(() => {
    if (orderForm.material_type !== 'upload' || !pendingMaterialFiles.value.length) {
      return ''
    }
    const parts = pendingMaterialFiles.value.map((file) => {
      const item = file.printOptions
      return `${file.name}(${item.printCount || 1}份,${item.printMode},${item.colorMode},${item.paperSize},${item.stampType})`
    })
    return `文件打印要求：${parts.join('；')}`
  })

  const cloneConfigValue = <T,>(value: T): T => JSON.parse(JSON.stringify(value || {}))

  const resetOrderForm = () => {
    Object.assign(orderForm, {
      business_license: false,
      company_id: null,
      courier_company: '',
      customer_address: '',
      customer_email: '',
      customer_name: '',
      customer_phone: '',
      delivery_option: '',
      material_type: 'upload',
      only_business_license: false,
      paper_size: 'A4',
      print_copies: 0,
      print_options: [],
      return_tracking_number: '',
      selected_license_companies: [],
      service_type: 'electronic',
      special_requirements: '',
      tracking_number: ''
    })
    orderStep.value = 0
    selectedServiceMode.value = 'electronic'
    mailStampType.value = '实体章'
    materialUploadRef.value?.clearFiles()
    pendingMaterialFiles.value = []
  }

  const selectServiceMode = (mode: ServiceMode) => {
    selectedServiceMode.value = mode
    if (mode === 'license_only') {
      orderForm.service_type = 'electronic'
      orderForm.only_business_license = true
      orderForm.business_license = true
      orderForm.material_type = 'upload'
      orderForm.print_copies = 0
      orderForm.print_options = []
      orderForm.delivery_option = ''
      orderForm.company_id = null
      orderForm.selected_license_companies = []
      materialUploadRef.value?.clearFiles()
      pendingMaterialFiles.value = []
    } else if (mode === 'mail') {
      orderForm.service_type = 'mail'
      orderForm.only_business_license = false
      orderForm.business_license = false
      orderForm.material_type = 'mail'
      orderForm.print_copies = 0
      orderForm.print_options = []
      orderForm.delivery_option = deliveryOptionKeys.value[0] || ''
      orderForm.company_id = null
      orderForm.selected_license_companies = []
      materialUploadRef.value?.clearFiles()
      pendingMaterialFiles.value = []
    } else if (mode === 'both') {
      orderForm.service_type = 'both'
      orderForm.only_business_license = false
      orderForm.business_license = false
      orderForm.material_type = 'upload'
      orderForm.delivery_option = deliveryOptionKeys.value[2] || deliveryOptionKeys.value[0] || ''
      orderForm.print_copies = orderForm.print_copies > 0 ? orderForm.print_copies : 10
      orderForm.company_id = null
      orderForm.selected_license_companies = []
    } else {
      orderForm.service_type = 'electronic'
      orderForm.only_business_license = false
      orderForm.business_license = false
      orderForm.material_type = 'upload'
      orderForm.print_copies = 0
      orderForm.print_options = []
      orderForm.delivery_option = ''
      orderForm.company_id = null
      orderForm.selected_license_companies = []
      materialUploadRef.value?.clearFiles()
    }
    void refreshQuote()
  }

  const setMaterialType = (type: string) => {
    orderForm.material_type = type
    if (type === 'mail') {
      orderForm.print_copies = 0
      orderForm.print_options = []
      materialUploadRef.value?.clearFiles()
      pendingMaterialFiles.value = []
    } else if (selectedServiceMode.value !== 'electronic' && orderForm.print_copies <= 0) {
      orderForm.print_copies = 10
    }
    void refreshQuote()
  }

  const beforePendingMaterialFileUpload = (rawFile: File) => {
    const maxSize = 50 * 1024 * 1024
    const ext = rawFile.name.split('.').pop()?.toLowerCase() || ''
    const allowed = ['pdf', 'doc', 'docx', 'zip', 'rar', '7z', 'jpg', 'jpeg', 'png', 'gif']
    if (!allowed.includes(ext)) {
      ElMessage.warning('支持格式：PDF、Word、图片和压缩包')
      return false
    }
    if (rawFile.size > maxSize) {
      ElMessage.warning('文件大小不能超过 50MB')
      return false
    }
    return true
  }

  const handlePendingMaterialFileChange = (uploadFile: UploadFile) => {
    if (!uploadFile.raw) return
    if (pendingMaterialFiles.value.length >= 10) {
      ElMessage.warning('最多上传 10 个文件')
      materialUploadRef.value?.clearFiles()
      return
    }
    if (!beforePendingMaterialFileUpload(uploadFile.raw)) {
      materialUploadRef.value?.clearFiles()
      return
    }
    const uid = String(uploadFile.uid || `${uploadFile.raw.name}-${uploadFile.raw.size}`)
    const exists = pendingMaterialFiles.value.some((item) => item.uid === uid)
    if (!exists) {
      pendingMaterialFiles.value.push({
        name: uploadFile.raw.name,
        pageCount: 1,
        printOptions: defaultFilePrintOptions(),
        raw: uploadFile.raw,
        size: uploadFile.raw.size,
        uid
      })
    }
    materialUploadRef.value?.clearFiles()
  }

  const handlePendingMaterialFileRemove = (index: number) => {
    pendingMaterialFiles.value.splice(index, 1)
    materialUploadRef.value?.clearFiles()
  }

  const clearPendingMaterialFiles = () => {
    pendingMaterialFiles.value = []
    materialUploadRef.value?.clearFiles()
  }

  const clearOrderUploadFiles = () => {
    orderUploadFiles.value = []
    orderUploadRef.value?.clearFiles()
  }

  const handleOrderUploadFileChange = (uploadFile: UploadFile) => {
    if (!uploadFile.raw) return
    if (orderUploadFiles.value.length >= 10) {
      ElMessage.warning('最多上传 10 个文件')
      orderUploadRef.value?.clearFiles()
      return
    }
    if (!beforePendingMaterialFileUpload(uploadFile.raw)) {
      orderUploadRef.value?.clearFiles()
      return
    }
    const uid = String(uploadFile.uid || `${uploadFile.raw.name}-${uploadFile.raw.size}`)
    const exists = orderUploadFiles.value.some((item) => item.uid === uid)
    if (!exists) {
      orderUploadFiles.value.push({
        name: uploadFile.raw.name,
        pageCount: 1,
        printOptions: defaultFilePrintOptions(),
        raw: uploadFile.raw,
        size: uploadFile.raw.size,
        uid
      })
    }
    orderUploadRef.value?.clearFiles()
  }

  const handleOrderUploadFileRemove = (index: number) => {
    orderUploadFiles.value.splice(index, 1)
    orderUploadRef.value?.clearFiles()
  }

  const handleCompanyChange = () => {
    if (!orderForm.company_id) {
      return
    }
    if (orderForm.business_license) {
      orderForm.selected_license_companies = [orderForm.company_id]
    }
    void refreshQuote()
  }

  const handleLicenseCompanyChange = () => {
    orderForm.company_id = orderForm.selected_license_companies[0] ?? null
    void refreshQuote()
  }

  const handleBusinessLicenseChange = () => {
    if (!orderForm.business_license) {
      orderForm.selected_license_companies = []
    } else if (orderForm.company_id) {
      orderForm.selected_license_companies = [orderForm.company_id]
    }
    void refreshQuote()
  }

  const handleMaterialFileExceed = () => {
    ElMessage.warning('最多上传 10 个文件')
  }

  const prevOrderStep = () => {
    if (orderStep.value > 0) {
      orderStep.value -= 1
    }
  }

  const nextOrderStep = async () => {
    if (currentStepKey.value === 'service') {
      if (!selectedServiceMode.value) {
        ElMessage.warning('请选择服务类型')
        return
      }
    }
    if (currentStepKey.value === 'company') {
      if (selectedServiceMode.value === 'license_only') {
        if (!orderForm.selected_license_companies.length) {
          ElMessage.warning('请选择营业执照公司')
          return
        }
      } else if (!orderForm.company_id) {
        ElMessage.warning('请选择公司')
        return
      }
    }
    if (
      currentStepKey.value === 'material' &&
      selectedServiceMode.value !== 'license_only' &&
      !orderForm.material_type
    ) {
      ElMessage.warning('请选择材料处理方式')
      return
    }
    if (currentStepKey.value === 'contact') {
      if (!orderForm.customer_name) {
        ElMessage.warning('请填写联系人姓名')
        return
      }
      if (needsEmail.value && !orderForm.customer_email) {
        ElMessage.warning('请填写邮箱地址')
        return
      }
      if (needsPhone.value && !orderForm.customer_phone) {
        ElMessage.warning('请填写联系电话')
        return
      }
      if (needsAddress.value && !orderForm.customer_address) {
        ElMessage.warning('邮寄服务请填写收货地址')
        return
      }
      if (needsCourier.value && (!orderForm.courier_company || !orderForm.tracking_number)) {
        ElMessage.warning('邮寄纸质文件请填写寄出快递信息')
        return
      }
    }
    if (orderStep.value < confirmStepIndex.value) {
      orderStep.value += 1
    }
    if (orderStep.value === confirmStepIndex.value) {
      await refreshQuote()
    }
  }

  const canGoNext = computed(() => {
    if (currentStepKey.value === 'service') return true
    if (currentStepKey.value === 'company') {
      return selectedServiceMode.value === 'license_only'
        ? Boolean(orderForm.selected_license_companies.length)
        : Boolean(orderForm.company_id)
    }
    if (currentStepKey.value === 'material') return Boolean(orderForm.material_type)
    if (currentStepKey.value === 'contact') return Boolean(orderForm.customer_name)
    return true
  })

  const stampDescription = (value: string) => {
    if (value === '骑缝章') return '页面连接处盖章'
    if (value === '实体章+骑缝章') return '指定位置和骑缝都盖'
    return '文件指定位置盖章'
  }

  const withStampRequirement = (requirements: string) => {
    const value = requirements.trim()
    if (!needsStampStep.value || !mailStampType.value || value.includes('盖章类型')) {
      return value
    }
    const stampLine = `盖章类型：${mailStampType.value}`
    return value ? `${value}\n${stampLine}` : stampLine
  }

  const finalSpecialRequirements = computed(() => {
    const base = withStampRequirement(orderForm.special_requirements || '')
    if (!filePrintRequirementText.value || base.includes('文件打印要求')) {
      return base
    }
    return base ? `${base}\n${filePrintRequirementText.value}` : filePrintRequirementText.value
  })

  const buildOrderRequest = (): LegacySXGZQuoteRequest => {
    const selectedLicenseCompanies =
      selectedServiceMode.value === 'license_only'
        ? [...orderForm.selected_license_companies]
        : orderForm.business_license && orderForm.company_id
          ? [orderForm.company_id]
          : []
    const companyID =
      selectedServiceMode.value === 'license_only'
        ? selectedLicenseCompanies[0] || 0
        : Number(orderForm.company_id || 0)
    const filePrintOptions = pendingMaterialFiles.value.map((file) => buildFilePrintPayload(file))
    const totalFileSheets =
      selectedServiceMode.value !== 'electronic' && orderForm.material_type === 'upload'
        ? pendingFilePrintSheets.value
        : 0
    const paperSize = filePrintOptions[0]?.paper_size || orderForm.paper_size
    const printOptions =
      selectedServiceMode.value !== 'electronic' && orderForm.material_type === 'upload'
        ? pendingFilePrintOptionNames.value
        : hasPrintSettings.value
          ? [...orderForm.print_options]
          : []

    return {
      business_license: selectedServiceMode.value === 'license_only' || orderForm.business_license,
      company_id: companyID,
      courier_company: needsCourier.value ? orderForm.courier_company : '',
      customer_address: needsAddress.value ? orderForm.customer_address : '',
      customer_email: needsEmail.value ? orderForm.customer_email : '',
      customer_name: orderForm.customer_name,
      customer_phone: needsPhone.value ? orderForm.customer_phone : '',
      delivery_option: needsDelivery.value ? orderForm.delivery_option : '',
      file_print_options: filePrintOptions,
      material_type:
        selectedServiceMode.value === 'license_only' ? 'upload' : orderForm.material_type,
      only_business_license: selectedServiceMode.value === 'license_only',
      paper_size: paperSize,
      print_copies: totalFileSheets || (hasPrintSettings.value ? orderForm.print_copies : 0),
      print_options: printOptions,
      return_tracking_number: '',
      selected_license_companies: selectedLicenseCompanies,
      service_type:
        selectedServiceMode.value === 'license_only' ? 'electronic' : orderForm.service_type,
      special_requirements: finalSpecialRequirements.value,
      tracking_number: needsCourier.value ? orderForm.tracking_number : ''
    }
  }

  const refreshQuote = async () => {
    if (currentStepKey.value !== 'confirm') {
      return
    }
    const payload = buildOrderRequest()
    if (!payload.company_id) {
      quote.base_price = 0
      quote.extra_options_price = 0
      quote.license_price = 0
      quote.print_price = 0
      quote.total_price = 0
      quote.company_name = ''
      return
    }
    quoteLoading.value = true
    try {
      const result = await quoteLegacySXGZOrder(payload)
      Object.assign(quote, result || {})
    } catch {
      quote.base_price = 0
      quote.extra_options_price = 0
      quote.license_price = 0
      quote.print_price = 0
      quote.total_price = 0
      quote.company_name = ''
    } finally {
      quoteLoading.value = false
    }
  }

  const scheduleQuote = () => {}

  const loadCompanies = async (options: { quote?: boolean } = {}) => {
    companyLoading.value = true
    try {
      const [companyList, licenseList] = await Promise.all([
        fetchLegacySXGZCompanies(),
        fetchLegacySXGZLicenseCompanies()
      ])
      companies.value = Array.isArray(companyList) ? companyList : []
      licenseCompanies.value = Array.isArray(licenseList) ? licenseList : companies.value
      if (!orderForm.delivery_option) {
        orderForm.delivery_option = deliveryOptionKeys.value[0] || ''
      }
      if (options.quote !== false) {
        await refreshQuote()
      }
    } finally {
      companyLoading.value = false
    }
  }

  const loadConfig = async () => {
    configLoading.value = true
    try {
      const result = await fetchLegacySXGZConfig()
      const normalized = result || {}
      Object.assign(configForm, {
        ...normalized,
        delivery_options: cloneConfigValue(normalized.delivery_options || {}),
        print_options: cloneConfigValue(normalized.print_options || {}),
        print_pricing: {
          base_free_copies: normalized.print_pricing?.base_free_copies || 10,
          extra_copy_price: normalized.print_pricing?.extra_copy_price || 0.5,
          per_copy_price: normalized.print_pricing?.per_copy_price || 2
        }
      })
      if (!orderForm.delivery_option) {
        orderForm.delivery_option = Object.keys(normalized.delivery_options || {})[0] || ''
      }
    } finally {
      configLoading.value = false
    }
  }

  const loadOrders = async (page = pagination.page) => {
    orderLoading.value = true
    pagination.page = page
    try {
      const api = isAdmin.value ? fetchLegacySXGZAdminOrders : fetchLegacySXGZOrders
      const result = await api({
        page: pagination.page,
        search: filters.search || undefined,
        size: pagination.size,
        status: filters.status || undefined
      })
      orders.value = Array.isArray(result?.list) ? result.list : []
      pagination.total = Number(result?.total || 0)
    } finally {
      orderLoading.value = false
    }
  }

  const resetFilters = () => {
    filters.search = ''
    filters.status = ''
    loadOrders(1)
  }

  const handleSearch = () => loadOrders(1)

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadOrders(1)
  }

  const openOrderDialog = () => {
    resetOrderForm()
    orderVisible.value = true
  }

  const openConfigDialog = () => {
    configVisible.value = true
  }

  const stopAnnouncementTimer = () => {
    if (announcementTimer.value) {
      window.clearInterval(announcementTimer.value)
      announcementTimer.value = null
    }
  }

  const loadAnnouncements = async (page = announcementPagination.page) => {
    announcementLoading.value = true
    announcementPagination.page = page
    try {
      const result = await fetchLegacySXGZAnnouncements({
        page: announcementPagination.page,
        pageSize: announcementPagination.pageSize,
        type: announcementPagination.type
      })
      const normalized = (result || {}) as Partial<LegacySXGZAnnouncementListResult>
      announcements.value = Array.isArray(normalized.data) ? normalized.data : []
      announcementPagination.total = Number(normalized.total || 0)
      announcementPagination.hasMore = Boolean(normalized.hasMore)
      if (!announcementPagination.hasMore && announcementPagination.total > 0) {
        announcementPagination.hasMore =
          announcementPagination.page * announcementPagination.pageSize <
          announcementPagination.total
      }
    } finally {
      announcementLoading.value = false
    }
  }

  const openAnnouncementDialog = async () => {
    announcementVisible.value = true
    announcementPagination.page = 1
    await loadAnnouncements(1)
    stopAnnouncementTimer()
    announcementTimer.value = window.setInterval(() => {
      if (announcementVisible.value) {
        void loadAnnouncements(announcementPagination.page)
      }
    }, 60000)
  }

  const changeAnnouncementPage = async (delta: number) => {
    const nextPage = announcementPagination.page + delta
    if (nextPage < 1) return
    if (delta > 0 && !announcementPagination.hasMore) return
    await loadAnnouncements(nextPage)
  }

  const handleCreateOrder = async () => {
    const payload = buildOrderRequest()
    if (
      selectedServiceMode.value === 'license_only' &&
      !payload.selected_license_companies.length
    ) {
      ElMessage.warning('请先选择营业执照公司')
      return
    }
    if (!payload.company_id) {
      ElMessage.warning('请先选择公司')
      return
    }
    if (!payload.customer_name) {
      ElMessage.warning('请填写联系人姓名')
      return
    }
    createLoading.value = true
    try {
      const result = await createLegacySXGZOrder(payload)
      ElMessage.success(`订单已创建：${result?.order_no || ''}`)
      const pendingFiles = pendingMaterialFiles.value
      if (pendingFiles.length && result?.order_id) {
        const failedFiles: string[] = []
        for (const file of pendingFiles) {
          try {
            await uploadLegacySXGZOrderFile(result.order_id, file.raw, buildFilePrintPayload(file))
          } catch {
            failedFiles.push(file.name)
          }
        }
        if (failedFiles.length) {
          ElMessage.warning(`订单已创建，但有 ${failedFiles.length} 个文件上传失败`)
        } else {
          ElMessage.success(`文件已自动上传 ${pendingFiles.length} 个`)
        }
      }
      clearPendingMaterialFiles()
      orderVisible.value = false
      await loadOrders(1)
    } finally {
      createLoading.value = false
    }
  }

  const openUploadDialog = async (order: LegacySXGZOrder) => {
    selectedOrder.value = order
    const files = order.files || { uploaded: [], processed: [] }
    selectedFiles.uploaded = files.uploaded || []
    selectedFiles.processed = files.processed || []
    clearOrderUploadFiles()
    uploadVisible.value = true
  }

  const refreshOrdersAndSelectedFiles = async () => {
    const refreshed = await (isAdmin.value ? fetchLegacySXGZAdminOrders : fetchLegacySXGZOrders)({
      page: pagination.page,
      search: filters.search || undefined,
      size: pagination.size,
      status: filters.status || undefined
    })
    orders.value = Array.isArray(refreshed?.list) ? refreshed.list : []
    pagination.total = Number(refreshed?.total || 0)
    if (selectedOrder.value) {
      const latest = orders.value.find((item) => item.order_id === selectedOrder.value?.order_id)
      if (latest) {
        selectedOrder.value = latest
        const fileSets = latest.files || { uploaded: [], processed: [] }
        selectedFiles.uploaded = fileSets.uploaded || []
        selectedFiles.processed = fileSets.processed || []
      }
    }
  }

  const handleUploadSelectedFiles = async () => {
    if (!selectedOrder.value) {
      ElMessage.warning('请先选择订单')
      return
    }
    if (!orderUploadFiles.value.length) {
      ElMessage.warning('请先选择文件')
      return
    }
    uploadSaving.value = true
    try {
      const failedFiles: string[] = []
      for (const file of orderUploadFiles.value) {
        try {
          await uploadLegacySXGZOrderFile(
            selectedOrder.value.order_id,
            file.raw,
            buildFilePrintPayload(file)
          )
        } catch {
          failedFiles.push(file.name)
        }
      }
      if (failedFiles.length) {
        ElMessage.warning(`有 ${failedFiles.length} 个文件上传失败`)
      } else {
        ElMessage.success(`文件已上传 ${orderUploadFiles.value.length} 个`)
      }
      clearOrderUploadFiles()
      await refreshOrdersAndSelectedFiles()
    } finally {
      uploadSaving.value = false
    }
  }

  const handleRefund = async (order: LegacySXGZOrder) => {
    if (!canRefundOrder(order)) {
      ElMessage.warning('当前订单状态不允许申请退款')
      return
    }
    const { value } = await ElMessageBox.prompt('请输入退款原因', '申请退款', {
      inputPlaceholder: '填写原因',
      inputType: 'textarea'
    })
    const reason = String(value || '').trim()
    if (!reason) {
      ElMessage.warning('退款原因不能为空')
      return
    }
    await applyLegacySXGZRefund(order.order_id, reason)
    ElMessage.success('退款申请已提交')
    loadOrders(pagination.page)
  }

  const openAdminDialog = (order: LegacySXGZOrder) => {
    adminForm.order_id = order.order_id
    adminForm.status = order.status
    adminForm.admin_notes = order.admin_notes || ''
    adminForm.refund_reason = order.refund_reason || ''
    adminVisible.value = true
  }

  const handleSaveAdminOrder = async () => {
    if (!adminForm.order_id) return
    adminSaving.value = true
    try {
      await updateLegacySXGZAdminOrder(adminForm.order_id, {
        admin_notes: adminForm.admin_notes,
        refund_reason: adminForm.refund_reason,
        status: adminForm.status
      })
      ElMessage.success('订单已更新')
      adminVisible.value = false
      loadOrders(pagination.page)
    } finally {
      adminSaving.value = false
    }
  }

  const handleSyncOrders = async () => {
    syncLoading.value = true
    try {
      const result = await syncLegacySXGZOrders()
      ElMessage.success(`同步完成：${result?.updated || 0} 条`)
      loadOrders(pagination.page)
    } finally {
      syncLoading.value = false
    }
  }

  const handleRefreshCompanies = async () => {
    refreshCompanyLoading.value = true
    try {
      await refreshLegacySXGZCompanies()
      await loadCompanies({ quote: false })
      ElMessage.success('上游公司已刷新')
    } finally {
      refreshCompanyLoading.value = false
    }
  }

  const handleSaveConfig = async () => {
    saveConfigLoading.value = true
    try {
      await saveLegacySXGZConfig({
        ...configForm,
        delivery_options: cloneConfigValue(configForm.delivery_options),
        print_options: cloneConfigValue(configForm.print_options),
        print_pricing: { ...configForm.print_pricing }
      })
      ElMessage.success('配置已保存')
    } finally {
      saveConfigLoading.value = false
    }
  }

  const canRefundOrder = (order: LegacySXGZOrder) =>
    order.status === 'pending' || order.status === 'processing'

  const formatMoney = (value?: number) => `¥${Number(value || 0).toFixed(2)}`

  const materialTypeLabel = (value?: string) => {
    if (value === 'mail') return '邮寄材料'
    return '上传材料'
  }

  const materialTagType = (value?: string) => (value === 'mail' ? 'warning' : 'success')

  const orderFileCount = (order: LegacySXGZOrder) =>
    (order.files?.uploaded?.length || 0) + (order.files?.processed?.length || 0)

  const orderTrackingText = (order: LegacySXGZOrder) => {
    if (order.courier_company || order.tracking_number) {
      return `${order.courier_company || '-'} ${order.tracking_number || ''}`.trim()
    }
    if (order.material_type === 'mail') return '待填写寄件'
    return '-'
  }

  const returnTrackingText = (order: LegacySXGZOrder) => order.return_tracking_number || '待填写'

  const orderRemarkNodes = (order: LegacySXGZOrder) => {
    const nodes = [
      order.special_requirements ? `备注：${order.special_requirements}` : '',
      order.refund_reason ? `退款：${order.refund_reason}` : '',
      order.admin_notes ? `管理：${order.admin_notes}` : ''
    ].filter(Boolean)
    return nodes.length ? nodes : ['-']
  }

  const { columnChecks, columns } = useTableColumns<LegacySXGZOrder>(() => [
    {
      prop: 'order_no',
      label: '订单号',
      minWidth: 170,
      formatter: (row) =>
        h('div', { class: 'sxgz-table-cell' }, [
          h('p', { class: 'sxgz-order-no' }, row.order_no || '-'),
          row.agent_order_id
            ? h('p', { class: 'text-xs text-g-500' }, `上游 ${row.agent_order_id}`)
            : null,
          h('p', { class: 'text-xs text-g-500' }, row.source || '本地')
        ])
    },
    {
      prop: 'customer',
      label: '客户信息',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'sxgz-table-cell' }, [
          h('p', { class: 'sxgz-cell-title' }, row.customer_name || '-'),
          h('p', { class: 'text-xs text-g-500' }, row.customer_phone || row.customer_email || '-'),
          row.customer_address
            ? h('p', { class: 'sxgz-cell-subtle truncate' }, row.customer_address)
            : null
        ])
    },
    {
      prop: 'company_name',
      label: '公司信息',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'sxgz-table-cell' }, [
          h('p', { class: 'sxgz-cell-title' }, row.company_name || '-'),
          h('p', { class: 'text-xs text-g-500' }, `CID ${row.company_id || '-'}`),
          row.only_business_license
            ? h(
                ElTag,
                { class: 'mt-1', effect: 'plain', size: 'small', type: 'warning' },
                () => '仅营业执照'
              )
            : null
        ])
    },
    {
      prop: 'service_type',
      label: '业务类型',
      width: 110,
      align: 'center',
      formatter: (row) =>
        h(ElTag, { effect: 'plain', type: 'primary' }, () => serviceTypeLabel(row.service_type))
    },
    {
      prop: 'service_detail',
      label: '服务详情',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'sxgz-table-cell' }, [
          h(
            ElTag,
            { effect: 'light', size: 'small', type: materialTagType(row.material_type) },
            () => materialTypeLabel(row.material_type)
          ),
          h(
            'p',
            { class: 'mt-1 text-xs text-g-500' },
            row.business_license ? '含营业执照' : '不含营业执照'
          ),
          row.print_copies
            ? h('p', { class: 'text-xs text-g-500' }, `打印 ${row.print_copies} 份`)
            : null
        ])
    },
    {
      prop: 'pricing',
      label: '价格明细',
      minWidth: 170,
      formatter: (row) =>
        h('div', { class: 'sxgz-table-cell' }, [
          h('p', { class: 'sxgz-price-total' }, formatMoney(row.total_price)),
          h(
            'p',
            { class: 'text-xs text-g-500' },
            `基础 ${formatMoney(row.base_price)} / 打印 ${formatMoney(row.print_price)}`
          ),
          row.license_price
            ? h('p', { class: 'text-xs text-g-500' }, `执照 ${formatMoney(row.license_price)}`)
            : null
        ])
    },
    {
      prop: 'files',
      label: '文件/快递',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'sxgz-table-cell' }, [
          h(
            ElTag,
            { effect: 'plain', size: 'small', type: orderFileCount(row) ? 'success' : 'info' },
            () => `${orderFileCount(row)} 个文件`
          ),
          h('p', { class: 'mt-1 truncate text-xs text-g-500' }, orderTrackingText(row))
        ])
    },
    {
      prop: 'return_tracking_number',
      label: '回寄单号',
      minWidth: 130,
      formatter: (row) =>
        h(
          'span',
          { class: row.return_tracking_number ? 'text-g-800' : 'text-g-500' },
          returnTrackingText(row)
        )
    },
    {
      prop: 'status',
      label: '订单状态',
      width: 120,
      align: 'center',
      formatter: (row) =>
        h(ElTag, { effect: 'plain', type: statusType(row.status) }, () => statusLabel(row.status))
    },
    {
      prop: 'created_at',
      label: '创建时间',
      width: 170
    },
    {
      prop: 'remark',
      label: '备注信息',
      minWidth: 220,
      formatter: (row) =>
        h(
          'div',
          { class: 'sxgz-table-cell' },
          orderRemarkNodes(row).map((item) =>
            h('p', { class: 'truncate text-xs text-g-500' }, item)
          )
        )
    },
    {
      prop: 'operation',
      label: '操作',
      width: 178,
      fixed: 'right',
      formatter: (row) =>
        h(
          'div',
          { class: 'flex items-center' },
          [
            h(ArtButtonTable, {
              icon: 'ri:upload-2-line',
              iconClass: 'bg-theme/12 text-theme',
              title: '上传文件',
              onClick: () => openUploadDialog(row)
            }),
            h(ArtButtonTable, {
              type: 'view',
              title: '查看文件',
              onClick: () => openOrderFiles(row)
            }),
            canRefundOrder(row)
              ? h(ArtButtonTable, {
                  icon: 'ri:refund-2-line',
                  iconClass: 'bg-warning/12 text-warning',
                  title: '申请退款',
                  onClick: () => handleRefund(row)
                })
              : null,
            isAdmin.value
              ? h(ArtButtonTable, {
                  icon: 'ri:settings-3-line',
                  iconClass: 'bg-secondary/12 text-secondary',
                  title: '管理订单',
                  onClick: () => openAdminDialog(row)
                })
              : null
          ].filter(Boolean)
        )
    }
  ])

  const openOrderFiles = async (order: LegacySXGZOrder) => {
    selectedOrder.value = order
    selectedFiles.uploaded = order.files?.uploaded || []
    selectedFiles.processed = order.files?.processed || []
    uploadVisible.value = true
  }

  const statusLabel = (status: string) => {
    const item = statusOptions.find((option) => option.value === status)
    return item?.label || status || '未知'
  }

  const statusType = (status: string) => {
    if (status === 'completed' || status === 'delivered') return 'success'
    if (status === 'processing' || status === 'pending') return 'warning'
    if (status === 'refund_requested') return 'info'
    if (status === 'refunded' || status === 'failed' || status === 'cancelled') return 'danger'
    return 'info'
  }

  watch(
    () => orderSteps.value.length,
    () => {
      if (orderStep.value > confirmStepIndex.value) {
        orderStep.value = confirmStepIndex.value
      }
    }
  )

  onMounted(async () => {
    resetOrderForm()
    await Promise.all([loadCompanies(), loadConfig(), loadOrders(1)])
  })

  onBeforeUnmount(() => {
    stopAnnouncementTimer()
  })
</script>

<style scoped>
  .sxgz-step-panel {
    border: 1px solid var(--art-border-color, #e5e7eb);
    border-radius: 6px;
    background: var(--art-main-bg-color, #fff);
  }

  .sxgz-step-panel {
    padding: 22px;
  }

  .sxgz-service-card {
    display: flex;
    min-height: 150px;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--art-border-color, #e5e7eb);
    border-radius: 6px;
    background: var(--art-main-bg-color, #fff);
    padding: 24px 16px;
    transition:
      border-color 0.18s ease,
      box-shadow 0.18s ease,
      transform 0.18s ease;
  }

  .sxgz-service-card:hover,
  .sxgz-service-card.is-active {
    border-color: var(--el-color-primary);
    box-shadow: 0 8px 20px rgb(0 0 0 / 6%);
    transform: translateY(-1px);
  }

  .sxgz-service-icon {
    display: inline-flex;
    width: 46px;
    height: 46px;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    font-size: 20px;
    font-weight: 700;
    line-height: 1;
  }

  .sxgz-service-icon.is-mail {
    color: #2563eb;
    background: rgb(37 99 235 / 10%);
  }

  .sxgz-service-icon.is-electronic {
    color: #16a34a;
    background: rgb(22 163 74 / 10%);
  }

  .sxgz-service-icon.is-both {
    color: #d97706;
    background: rgb(217 119 6 / 10%);
  }

  .sxgz-service-icon.is-license {
    color: #e11d48;
    background: rgb(225 29 72 / 10%);
  }

  .sxgz-option-card {
    display: flex;
    min-height: 96px;
    width: 100%;
    flex-direction: column;
    justify-content: center;
    border: 1px solid var(--art-border-color, #e5e7eb);
    border-radius: 6px;
    background: var(--art-main-bg-color, #fff);
    padding: 18px;
    text-align: left;
    transition:
      border-color 0.18s ease,
      box-shadow 0.18s ease;
  }

  .sxgz-option-card:hover,
  .sxgz-option-card.is-active {
    border-color: var(--el-color-primary);
    box-shadow: 0 6px 16px rgb(0 0 0 / 5%);
  }

  .sxgz-selected-list {
    overflow: hidden;
    border: 1px solid var(--art-border-color, #e5e7eb);
    border-radius: 6px;
    background: var(--art-main-bg-color, #fff);
  }

  .sxgz-selected-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
  }

  .sxgz-selected-row + .sxgz-selected-row {
    border-top: 1px solid var(--art-border-color, #e5e7eb);
  }

  .sxgz-search-bar {
    margin-bottom: 12px;
  }

  .sxgz-table-cell {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 2px;
    line-height: 1.45;
  }

  .sxgz-order-no {
    font-size: 13px;
    font-weight: 600;
    color: var(--g-text-title, #111827);
  }

  .sxgz-cell-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--g-text-title, #111827);
  }

  .sxgz-cell-subtle {
    color: var(--g-text-secondary, #6b7280);
  }

  .sxgz-price-total {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-color-danger);
  }
</style>
