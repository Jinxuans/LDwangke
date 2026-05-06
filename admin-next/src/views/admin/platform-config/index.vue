<template>
  <div class="admin-platform-config-page art-full-height">
    <section class="art-card-sm p-5">
      <div class="grid gap-4 lg:grid-cols-[1.2fr_0.8fr_0.8fr_auto]">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">关键词</label>
          <ElInput v-model="searchForm.keyword" clearable placeholder="搜索平台标识、名称、路径" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">认证方式</label>
          <ElSelect v-model="searchForm.auth_type" class="w-full" clearable placeholder="全部认证">
            <ElOption
              v-for="item in authTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">能力筛选</label>
          <ElSelect v-model="searchForm.capability" class="w-full" clearable placeholder="全部能力">
            <ElOption
              v-for="item in capabilityOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </div>
        <div class="flex items-end gap-3">
          <ElButton type="primary" @click="loadData">刷新列表</ElButton>
          <ElButton @click="resetSearch">重置</ElButton>
        </div>
      </div>
    </section>

    <section class="art-card-sm mt-4 overflow-hidden">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">平台 {{ filteredList.length }} 条</ElTag>
            <ElTag v-if="searchActiveCount" type="primary" effect="plain">筛选 {{ searchActiveCount }} 项</ElTag>
            <ElTag type="success" effect="plain">余额 {{ balanceCapabilityCount }}</ElTag>
            <ElTag type="warning" effect="plain">自定义查课 {{ customQueryCount }}</ElTag>
            <ElButton plain @click="openPHPImport">PHP 导入</ElButton>
            <ElButton plain @click="openDetect">自动检测</ElButton>
            <ElButton type="primary" plain @click="openAdd">新增平台</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="loading" :data="filteredList" :columns="columns" :show-table-header="true" />
    </section>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? `编辑平台 ${editForm.pt || ''}` : '新增平台配置'"
      width="1220px"
      destroy-on-close
    >
      <div class="grid gap-5 xl:grid-cols-[0.82fr_1.18fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">基础资料</h3>
            <p class="mt-1 text-sm text-g-500">标识、认证与基础开关。</p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">平台标识</label>
              <ElInput v-model="editForm.pt" maxlength="40" placeholder="如 longlong" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">平台名称</label>
              <ElInput v-model="editForm.name" maxlength="60" placeholder="请输入平台名称" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">认证方式</label>
              <ElSelect v-model="editForm.auth_type" class="w-full">
                <ElOption v-for="item in authTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </ElSelect>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">成功码</label>
              <ElInput v-model="editForm.success_codes" placeholder="默认 0，可填多个逗号分隔" />
            </div>
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">查课驱动</label>
              <ElSelect v-model="editForm.query_act" class="w-full" clearable placeholder="默认按路径请求；只有特殊平台才选驱动">
                <ElOption v-for="item in queryDriverOptions" :key="item" :label="item" :value="item" />
              </ElSelect>
            </div>
          </div>

          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">JSON 请求</p>
                  <p class="mt-1 text-xs text-g-500">请求体切换</p>
                </div>
                <ElSwitch v-model="editForm.use_json" />
              </div>
            </div>
            <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">需要代理</p>
                  <p class="mt-1 text-xs text-g-500">代理转发</p>
                </div>
                <ElSwitch v-model="editForm.need_proxy" />
              </div>
            </div>
            <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">返回 YID</p>
                  <p class="mt-1 text-xs text-g-500">下单结果包含上游单号</p>
                </div>
                <ElSwitch v-model="editForm.returns_yid" />
              </div>
            </div>
            <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">额外参数</p>
                  <p class="mt-1 text-xs text-g-500">扩展映射</p>
                </div>
                <ElSwitch v-model="editForm.extra_params" />
              </div>
            </div>
          </div>

          <div class="mt-5 flex flex-wrap gap-2">
            <ElTag type="primary" effect="plain">{{ editForm.pt || 'new-platform' }}</ElTag>
            <ElTag effect="plain">{{ editForm.name || '未命名平台' }}</ElTag>
            <ElTag effect="plain">{{ editForm.auth_type || '未选择认证' }}</ElTag>
            <ElTag :type="editForm.use_json ? 'success' : 'info'" effect="plain">
              {{ editForm.use_json ? 'JSON' : '表单/自动' }}
            </ElTag>
            <ElTag :type="editForm.returns_yid ? 'warning' : 'info'" effect="plain">
              {{ editForm.returns_yid ? '返回 YID' : '不返回 YID' }}
            </ElTag>
          </div>

          <div class="mt-5">
            <label class="mb-2 block text-sm font-medium text-g-800">来源代码 / 备注</label>
            <ElInput
              v-model="editForm.source_code"
              type="textarea"
              :rows="6"
              placeholder="可粘贴解析来源的 PHP 片段，便于后续核对。"
            />
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">接口细节</h3>
            <p class="mt-1 text-sm text-g-500">按能力填写对应路径与参数映射。</p>
          </div>

          <ElTabs v-model="activeEditTab" class="mt-5">
            <ElTabPane label="查课 / 下单" name="query-order">
              <div class="grid gap-4 xl:grid-cols-2">
                <section class="rounded-custom-sm border-full-d p-4">
                  <p class="text-base font-semibold text-g-900">查课接口</p>
                  <div class="mt-4 grid gap-4">
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">查课路径</label>
                      <ElInput v-model="editForm.query_path" placeholder="/api.php?act=get" />
                    </div>
                    <div class="grid gap-4 md:grid-cols-3">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">请求方式</label>
                        <ElSelect v-model="editForm.query_method" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">Body 类型</label>
                        <ElSelect v-model="editForm.query_body_type" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">参数风格</label>
                        <ElInput v-model="editForm.query_param_style" placeholder="可留空" />
                      </div>
                    </div>
                    <div class="grid gap-4 md:grid-cols-3">
                      <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                        <div class="flex items-center justify-between gap-3">
                          <div>
                            <p class="text-sm font-medium text-g-900">轮询查课</p>
                            <p class="mt-1 text-xs text-g-500">用于异步返回</p>
                          </div>
                          <ElSwitch v-model="editForm.query_polling" />
                        </div>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">最大尝试次数</label>
                        <ElInputNumber v-model="editForm.query_max_attempts" class="w-full" :min="0" :step="1" />
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">轮询间隔(ms)</label>
                        <ElInputNumber v-model="editForm.query_interval" class="w-full" :min="0" :step="100" />
                      </div>
                    </div>
                    <p class="text-xs leading-5 text-g-500">{{ queryParamHelp }}</p>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">参数映射</label>
                      <ElInput v-model="editForm.query_param_map" type="textarea" :rows="4" :placeholder="queryParamMapExample" />
                    </div>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">响应映射</label>
                      <ElInput v-model="editForm.query_response_map" type="textarea" :rows="4" placeholder='如 {"data.list":"classes"}' />
                    </div>
                  </div>
                </section>

                <section class="rounded-custom-sm border-full-d p-4">
                  <p class="text-base font-semibold text-g-900">下单接口</p>
                  <div class="mt-4 grid gap-4">
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">下单路径</label>
                      <ElInput v-model="editForm.order_path" placeholder="/api.php?act=add" />
                    </div>
                    <div class="grid gap-4 md:grid-cols-3">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">请求方式</label>
                        <ElSelect v-model="editForm.order_method" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">Body 类型</label>
                        <ElSelect v-model="editForm.order_body_type" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                      <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                        <div class="flex items-center justify-between gap-3">
                          <div>
                            <p class="text-sm font-medium text-g-900">YID 在 data 数组</p>
                            <p class="mt-1 text-xs text-g-500">兼容旧返回结构</p>
                          </div>
                          <ElSwitch v-model="editForm.yid_in_data_array" />
                        </div>
                      </div>
                    </div>
                    <p class="text-xs leading-5 text-g-500">{{ orderParamHelp }}</p>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">参数映射</label>
                      <ElInput v-model="editForm.order_param_map" type="textarea" :rows="6" :placeholder="orderParamMapExample" />
                    </div>
                  </div>
                </section>
              </div>
            </ElTabPane>

            <ElTabPane label="进度 / 目录" name="progress-catalog">
              <div class="space-y-4">
                <section
                  v-for="section in progressSections"
                  :key="section.key"
                  class="rounded-custom-sm border-full-d p-4"
                >
                  <p class="text-base font-semibold text-g-900">{{ section.title }}</p>
                  <p class="mt-1 text-xs leading-5 text-g-500">{{ section.help }}</p>
                  <div class="mt-4 grid gap-4">
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">{{ section.pathLabel }}</label>
                      <ElInput v-model="editForm[`${section.key}_path`]" :placeholder="section.pathPlaceholder" />
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">请求方式</label>
                        <ElSelect v-model="editForm[`${section.key}_method`]" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">Body 类型</label>
                        <ElSelect v-model="editForm[`${section.key}_body_type`]" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">参数映射</label>
                      <ElInput v-model="editForm[`${section.key}_param_map`]" type="textarea" :rows="4" :placeholder="section.example" />
                    </div>
                  </div>
                </section>

                <div class="grid gap-4 xl:grid-cols-2">
                  <section
                    v-for="section in catalogSections"
                    :key="section.key"
                    class="rounded-custom-sm border-full-d p-4"
                  >
                    <p class="text-base font-semibold text-g-900">{{ section.title }}</p>
                    <div class="mt-4 grid gap-4">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">接口路径</label>
                        <ElInput v-model="editForm[`${section.key}_path`]" :placeholder="section.pathPlaceholder" />
                      </div>
                      <div class="grid gap-4 md:grid-cols-2">
                        <div>
                          <label class="mb-2 block text-sm font-medium text-g-800">请求方式</label>
                          <ElSelect v-model="editForm[`${section.key}_method`]" class="w-full" clearable placeholder="自动">
                            <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                          </ElSelect>
                        </div>
                        <div>
                          <label class="mb-2 block text-sm font-medium text-g-800">Body 类型</label>
                          <ElSelect v-model="editForm[`${section.key}_body_type`]" class="w-full" clearable placeholder="自动">
                            <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                          </ElSelect>
                        </div>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">参数映射</label>
                        <ElInput v-model="editForm[`${section.key}_param_map`]" type="textarea" :rows="4" :placeholder="section.example" />
                      </div>
                    </div>
                  </section>
                </div>
              </div>
            </ElTabPane>

            <ElTabPane label="动作接口" name="actions">
              <div class="space-y-4">
                <section
                  v-for="section in actionSections"
                  :key="section.key"
                  class="rounded-custom-sm border-full-d p-4"
                >
                  <p class="text-base font-semibold text-g-900">{{ section.title }}</p>
                  <p v-if="section.help" class="mt-1 text-xs leading-5 text-g-500">{{ section.help }}</p>
                  <div class="mt-4 grid gap-4">
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">接口路径</label>
                        <ElInput v-model="editForm[`${section.key}_path`]" :placeholder="section.pathPlaceholder" />
                      </div>
                      <div v-if="section.extraFieldKey">
                        <label class="mb-2 block text-sm font-medium text-g-800">{{ section.extraFieldLabel }}</label>
                        <ElInput v-model="editForm[section.extraFieldKey]" :placeholder="section.extraFieldPlaceholder" />
                      </div>
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">请求方式</label>
                        <ElSelect v-model="editForm[`${section.key}_method`]" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">Body 类型</label>
                        <ElSelect v-model="editForm[`${section.key}_body_type`]" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <div v-if="section.idFieldKey" class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">{{ section.idFieldLabel }}</label>
                        <ElInput v-model="editForm[section.idFieldKey]" :placeholder="section.idFieldPlaceholder" />
                      </div>
                    </div>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">参数映射</label>
                      <ElInput v-model="editForm[`${section.key}_param_map`]" type="textarea" :rows="4" :placeholder="section.example" />
                    </div>
                  </div>
                </section>
              </div>
            </ElTabPane>

            <ElTabPane label="余额 / 工单" name="balance-report">
              <div class="grid gap-4 xl:grid-cols-2">
                <section class="rounded-custom-sm border-full-d p-4">
                  <p class="text-base font-semibold text-g-900">余额接口</p>
                  <div class="mt-4 grid gap-4">
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">余额路径</label>
                      <ElInput v-model="editForm.balance_path" placeholder="/api.php?act=getmoney" />
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">请求方式</label>
                        <ElSelect v-model="editForm.balance_method" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">Body 类型</label>
                        <ElSelect v-model="editForm.balance_body_type" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">金额字段</label>
                        <ElSelect v-model="editForm.balance_money_field" class="w-full" clearable placeholder="请选择字段路径">
                          <ElOption v-for="item in balanceFieldOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">认证覆盖</label>
                        <ElSelect v-model="editForm.balance_auth_type" class="w-full" clearable placeholder="跟随全局">
                          <ElOption v-for="item in balanceAuthTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">参数映射</label>
                      <ElInput v-model="editForm.balance_param_map" type="textarea" :rows="4" :placeholder="balanceParamMapExample" />
                    </div>
                  </div>
                </section>

                <section class="rounded-custom-sm border-full-d p-4">
                  <p class="text-base font-semibold text-g-900">工单 / 刷新</p>
                  <div class="mt-4 grid gap-4">
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">提交工单路径</label>
                        <ElInput v-model="editForm.report_path" placeholder="如 /api/report" />
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">查询工单路径</label>
                        <ElInput v-model="editForm.get_report_path" placeholder="如 /api/report/status" />
                      </div>
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">提交方式</label>
                        <ElSelect v-model="editForm.report_method" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">提交 Body</label>
                        <ElSelect v-model="editForm.report_body_type" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">参数风格</label>
                        <ElSelect v-model="editForm.report_param_style" class="w-full" clearable placeholder="standard">
                          <ElOption v-for="item in reportParamStyleOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">认证覆盖</label>
                        <ElSelect v-model="editForm.report_auth_type" class="w-full" clearable placeholder="跟随全局">
                          <ElOption v-for="item in reportAuthTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <p class="text-xs leading-5 text-g-500">{{ reportParamHelp }}</p>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">提交参数映射</label>
                      <ElInput v-model="editForm.report_param_map" type="textarea" :rows="4" :placeholder="reportParamMapExample" />
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">查询方式</label>
                        <ElSelect v-model="editForm.get_report_method" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in methodOptions" :key="item" :label="item" :value="item" />
                        </ElSelect>
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">查询 Body</label>
                        <ElSelect v-model="editForm.get_report_body_type" class="w-full" clearable placeholder="自动">
                          <ElOption v-for="item in bodyTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                      </div>
                    </div>
                    <p class="text-xs leading-5 text-g-500">{{ getReportParamHelp }}</p>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">查询参数映射</label>
                      <ElInput v-model="editForm.get_report_param_map" type="textarea" :rows="4" :placeholder="getReportParamMapExample" />
                    </div>
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">刷新路径</label>
                      <ElInput v-model="editForm.refresh_path" placeholder="如 /api/refresh" />
                    </div>
                  </div>
                </section>
              </div>
            </ElTabPane>
          </ElTabs>
        </section>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存平台配置</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="phpVisible" title="从 PHP 代码导入" width="780px" destroy-on-close>
      <p class="mb-4 text-sm text-g-500">粘贴旧 PHP 平台分支，系统会先解析字段，再由你在编辑器中确认保存。</p>
      <ElInput
        v-model="phpCode"
        type="textarea"
        :rows="14"
        placeholder='if ($type == "newplat") { ... }'
      />

      <div class="mt-4 flex justify-between gap-3">
        <div class="text-sm text-g-500">解析后不会直接落库，仍需进入编辑弹窗核对再保存。</div>
        <ElButton type="primary" :loading="phpLoading" @click="parsePHP">解析代码</ElButton>
      </div>

      <div v-if="phpResult" class="mt-5 rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
        <div class="flex flex-wrap gap-2">
          <ElTag type="primary" effect="plain">{{ phpResult.pt || '未识别平台' }}</ElTag>
          <ElTag effect="plain">{{ phpResult.auth_type || '未识别认证' }}</ElTag>
          <ElTag type="success" effect="plain">置信度 {{ phpResult.confidence }}%</ElTag>
        </div>
        <p class="mt-3 text-sm text-g-500">
          {{ phpResult.name || '未识别名称' }}，查课 {{ formatEndpoint(phpResult.query_path, phpResult.query_act) }}。
        </p>
        <div class="mt-4 flex flex-wrap gap-2">
          <ElTag type="primary" effect="plain">查课 {{ formatEndpoint(phpResult.query_path, phpResult.query_act) }}</ElTag>
          <ElTag type="success" effect="plain">下单 {{ formatEndpoint(phpResult.order_path) }}</ElTag>
          <ElTag type="warning" effect="plain">进度 {{ formatEndpoint(phpResult.progress_path) }}</ElTag>
        </div>
        <div v-if="phpResult.warnings?.length" class="mt-4 flex flex-wrap gap-2">
          <ElTag
            v-for="item in phpResult.warnings"
            :key="item"
            type="warning"
            effect="plain"
          >
            {{ item }}
          </ElTag>
        </div>
        <div class="mt-4 flex justify-end">
          <ElButton type="primary" @click="applyPHPResult">应用到编辑器</ElButton>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="detectVisible" title="自动检测平台" width="880px" destroy-on-close>
      <div class="grid gap-5 xl:grid-cols-[0.9fr_1.1fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">检测参数</h3>
            <p class="mt-1 text-sm text-g-500">输入地址与凭证后探测接口。</p>
          </div>
          <div class="mt-5 grid gap-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">平台 URL</label>
              <ElInput v-model="detectForm.url" placeholder="http://xxx.com" />
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">UID</label>
                <ElInput v-model="detectForm.uid" placeholder="上游账号" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Key</label>
                <ElInput v-model="detectForm.key" placeholder="上游密钥" />
              </div>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
                <ElInput v-model="detectForm.token" placeholder="可选" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Cookie</label>
                <ElInput v-model="detectForm.cookie" placeholder="可选" />
              </div>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">平台标识</label>
                <ElInput v-model="detectForm.pt" placeholder="一键保存时必填" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">平台名称</label>
                <ElInput v-model="detectForm.name" placeholder="留空则自动建议" />
              </div>
            </div>
            <div class="flex flex-wrap gap-3">
              <ElButton type="primary" :loading="detectLoading" @click="runDetect">仅检测</ElButton>
              <ElButton plain :loading="detectSaveLoading" @click="autoDetectSave">检测并保存</ElButton>
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">检测结果</h3>
            <p class="mt-1 text-sm text-g-500">确认结果后可直接填入编辑器。</p>
          </div>

          <div v-if="detectResult" class="mt-5 space-y-4">
            <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <ElTag :type="detectResult.success ? 'success' : 'warning'" effect="plain">
                  {{ detectResult.success ? '检测成功' : '未识别可用接口' }}
                </ElTag>
                <ElTag effect="plain">{{ detectResult.suggested_name || '未生成建议名称' }}</ElTag>
                <ElTag effect="plain">{{ detectResult.auth_type || '未识别认证' }}</ElTag>
                <ElTag effect="plain">成功码 {{ detectResult.success_code || '-' }}</ElTag>
              </div>
              <p class="mt-3 text-sm text-g-500">
                平台 URL {{ detectForm.url || '-' }}，可根据下方探测项决定是否应用到编辑器。
              </p>
            </div>

            <div class="flex flex-wrap gap-2">
              <ElTag :type="detectResult.balance_ok ? 'success' : 'info'" effect="plain">
                余额 {{ detectResult.balance_ok ? detectResult.balance_money || '可用' : '不可用' }}
              </ElTag>
              <ElTag :type="detectResult.query_ok ? 'success' : 'info'" effect="plain">
                查课 {{ detectResult.query_ok ? '可用' : '不可用' }}
              </ElTag>
              <ElTag :type="detectResult.class_list_ok ? 'success' : 'info'" effect="plain">
                课程列表 {{ detectResult.class_list_ok ? '可用' : '不可用' }}
              </ElTag>
              <ElTag :type="detectResult.category_ok ? 'success' : 'info'" effect="plain">
                分类 {{ detectResult.category_ok ? '可用' : '不可用' }}
              </ElTag>
            </div>

            <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <p class="text-sm font-semibold text-g-900">探测详情</p>
              <div class="mt-3 space-y-2">
                <article
                  v-for="(probe, index) in detectResult.probes"
                  :key="index"
                  class="flex flex-wrap items-center gap-2 rounded-custom-sm border-full-d bg-[var(--el-bg-color)] px-3 py-2 text-sm"
                >
                  <ElTag :type="probeTagType(probe.status)" effect="plain">{{ probe.status }}</ElTag>
                  <ElTag effect="plain">{{ probe.method }}</ElTag>
                  <span class="font-medium text-g-900">{{ probe.endpoint }}</span>
                  <span class="text-g-500">{{ probe.msg }}</span>
                </article>
              </div>
            </div>

            <div class="flex justify-end gap-3">
              <ElButton :disabled="!detectResult.success" @click="applyDetectResult">应用到编辑器</ElButton>
            </div>
          </div>

            <ElEmpty v-else description="先执行一次检测" />
        </section>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    autoDetectSaveLegacyPlatform,
    deleteLegacyPlatformConfig,
    detectLegacyPlatform,
    fetchLegacyPlatformConfigs,
    parseLegacyPlatformPHPCode,
    saveLegacyPlatformConfig,
    type LegacyParsedPHPResult,
    type LegacyPlatformConfig,
    type LegacyPlatformDetectResult
  } from '@/api/legacy/admin-platform-config'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminPlatformConfigPage' })

  const methodOptions = ['POST', 'GET', 'PUT']
  const queryDriverOptions = ['local_time', 'local_script', 'xxt_query', 'KUN_custom', 'simple_custom', 'yyy_custom', 'tuboshu_custom', 'nx_custom', 'lgwk_custom']
  const authTypeOptions = [
    { label: 'UID + Key', value: 'uid_key' },
    { label: 'Token', value: 'token_only' },
    { label: 'Bearer Token', value: 'bearer_token' },
    { label: 'Cookie', value: 'cookie' }
  ]
  const capabilityOptions = [
    { label: '支持余额', value: 'balance' },
    { label: '支持批量进度', value: 'batch_progress' },
    { label: '支持工单', value: 'report' },
    { label: '支持暂停', value: 'pause' },
    { label: '支持恢复', value: 'resume' },
    { label: '自定义查课驱动', value: 'custom_query' }
  ]
  const bodyTypeOptions = [
    { label: '自动', value: '' },
    { label: 'form', value: 'form' },
    { label: 'json', value: 'json' },
    { label: 'query', value: 'query' }
  ]
  const balanceFieldOptions = [
    { label: 'money（根级）', value: 'money' },
    { label: 'data.money', value: 'data.money' },
    { label: 'data', value: 'data' },
    { label: 'data.remainscore', value: 'data.remainscore' }
  ]
  const balanceAuthTypeOptions = [
    { label: '跟随全局', value: '' },
    { label: 'UID + Key', value: 'uid_key' },
    { label: 'Token', value: 'token_only' },
    { label: 'Bearer Token', value: 'bearer_token' }
  ]
  const reportParamStyleOptions = [
    { label: 'standard', value: '' },
    { label: 'token', value: 'token' }
  ]
  const reportAuthTypeOptions = [
    { label: '跟随全局', value: '' },
    { label: 'token_only', value: 'token_only' }
  ]

  const loading = ref(false)
  const saving = ref(false)
  const list = ref<LegacyPlatformConfig[]>([])

  const searchForm = reactive({
    keyword: '',
    auth_type: '',
    capability: ''
  })

  const dialogVisible = ref(false)
  const phpVisible = ref(false)
  const detectVisible = ref(false)
  const isEditing = ref(false)
  const activeEditTab = ref('query-order')

  const phpLoading = ref(false)
  const phpCode = ref('')
  const phpResult = ref<LegacyParsedPHPResult | null>(null)

  const detectLoading = ref(false)
  const detectSaveLoading = ref(false)
  const detectResult = ref<LegacyPlatformDetectResult | null>(null)

  const detectForm = reactive({
    url: '',
    uid: '',
    key: '',
    token: '',
    cookie: '',
    pt: '',
    name: ''
  })

  const editForm = reactive(createDefaultEditForm() as ReturnType<typeof createDefaultEditForm> & Record<string, any>)

  const actionParamHelp =
    '接口配置现在只认显式命名空间模板变量。通用变量：{{supplier.uid}} {{supplier.key}} {{supplier.token}} {{order.yid}}。'
  const queryParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","school":"{{action.school}}","user":"{{action.user}}","pass":"{{action.password}}","platform":"{{action.platform}}"}')
  const queryParamHelp = `查课动作变量：{{action.school}} {{action.user}} {{action.password}} {{action.platform}}。默认模板：${queryParamMapExample}`
  const orderParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","platform":"{{order.noun}}","school":"{{order.school}}","user":"{{order.user}}","pass":"{{order.pass}}","kcid":"{{order.kcid}}","kcname":"{{order.kcname}}"}')
  const orderParamHelp = `下单动作默认模板：${orderParamMapExample}`
  const progressParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}')
  const batchProgressParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}')
  const categoryParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}')
  const classListParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}')
  const pauseParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}')
  const resumeParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}')
  const changePassParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","newPwd":"{{action.new_password}}"}')
  const changePassParamHelp = `可用变量：{{supplier.uid}} {{supplier.key}} {{order.yid}} {{action.new_password}}。默认模板：${changePassParamMapExample}`
  const resubmitParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}')
  const logParamMapExample = formatJsonTemplate('{"dtoken":"{{supplier.token}}","account":"{{order.user}}","password":"{{order.pass}}","course":"{{order.kcname}}","courseId":"{{order.kcid}}"}')
  const balanceParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}')
  const reportParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","question":"{{action.content}}"}')
  const reportParamHelp = `提工单动作变量：{{action.content}} {{action.ticket_type}}。默认模板：${reportParamMapExample}`
  const getReportParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","reportId":"{{action.report_id}}"}')
  const getReportParamHelp = `查询工单变量：{{action.report_id}}。默认模板：${getReportParamMapExample}`

  const progressSections = [
    {
      key: 'progress',
      title: '单条进度',
      pathLabel: '进度路径',
      pathPlaceholder: '/api.php?act=chadan2',
      example: progressParamMapExample,
      help: '单条进度一般会携带订单账号、课程和上游订单号。'
    },
    {
      key: 'batch_progress',
      title: '批量进度',
      pathLabel: '批量进度路径',
      pathPlaceholder: '如 /api/batch-progress',
      example: batchProgressParamMapExample,
      help: '批量进度通常只带供应商级凭证，也可附带批次上下文。'
    }
  ]

  const catalogSections = [
    {
      key: 'category',
      title: '分类拉取',
      pathPlaceholder: '/api.php?act=getcate',
      example: categoryParamMapExample
    },
    {
      key: 'class_list',
      title: '课程列表',
      pathPlaceholder: '/api.php?act=getclass',
      example: classListParamMapExample
    }
  ]

  const actionSections = [
    {
      key: 'pause',
      title: '暂停订单',
      pathPlaceholder: '/api.php?act=zt',
      example: pauseParamMapExample,
      idFieldKey: 'pause_id_param',
      idFieldLabel: '订单 ID 参数',
      idFieldPlaceholder: 'id'
    },
    {
      key: 'resume',
      title: '恢复订单',
      pathPlaceholder: '如 /api/resume',
      example: resumeParamMapExample
    },
    {
      key: 'change_pass',
      title: '修改密码',
      pathPlaceholder: '/api.php?act=gaimi',
      example: changePassParamMapExample,
      help: changePassParamHelp,
      extraFieldKey: 'change_pass_param',
      extraFieldLabel: '密码参数名',
      extraFieldPlaceholder: 'newPwd',
      idFieldKey: 'change_pass_id_param',
      idFieldLabel: '订单 ID 参数',
      idFieldPlaceholder: 'id'
    },
    {
      key: 'resubmit',
      title: '补单',
      pathPlaceholder: '/api.php?act=budan',
      example: resubmitParamMapExample,
      idFieldKey: 'resubmit_id_param',
      idFieldLabel: '订单 ID 参数',
      idFieldPlaceholder: 'id'
    },
    {
      key: 'log',
      title: '日志查询',
      pathPlaceholder: '/api.php?act=xq',
      example: logParamMapExample,
      idFieldKey: 'log_id_param',
      idFieldLabel: '日志 ID 参数',
      idFieldPlaceholder: 'id',
      help: actionParamHelp
    }
  ]

  const filteredList = computed(() => {
    const keyword = searchForm.keyword.trim().toLowerCase()
    return list.value.filter((item) => {
      const matchesKeyword =
        !keyword ||
        item.pt?.toLowerCase().includes(keyword) ||
        item.name?.toLowerCase().includes(keyword) ||
        item.query_path?.toLowerCase().includes(keyword) ||
        item.order_path?.toLowerCase().includes(keyword)
      const matchesAuth = !searchForm.auth_type || item.auth_type === searchForm.auth_type
      const matchesCapability = !searchForm.capability || hasCapability(item, searchForm.capability)
      return matchesKeyword && matchesAuth && matchesCapability
    })
  })

  const searchActiveCount = computed(
    () => [searchForm.keyword, searchForm.auth_type, searchForm.capability].filter(Boolean).length
  )

  const balanceCapabilityCount = computed(() => list.value.filter((item) => Boolean(item.balance_path)).length)

  const customQueryCount = computed(() => list.value.filter((item) => Boolean(item.query_act)).length)

  function createDefaultEditForm() {
    return {
      pt: '',
      name: '',
      auth_type: 'uid_key',
      success_codes: '0',
      use_json: false,
      need_proxy: false,
      returns_yid: false,
      extra_params: false,
      query_act: '',
      query_path: '',
      query_method: '',
      query_body_type: '',
      query_param_style: '',
      query_param_map: '',
      query_polling: false,
      query_max_attempts: 0,
      query_interval: 0,
      query_response_map: '',
      order_path: '',
      order_method: '',
      order_body_type: '',
      order_param_map: '',
      yid_in_data_array: false,
      progress_path: '',
      progress_method: '',
      progress_body_type: '',
      progress_param_map: '',
      batch_progress_path: '',
      batch_progress_method: '',
      batch_progress_body_type: '',
      batch_progress_param_map: '',
      category_path: '',
      category_method: '',
      category_body_type: '',
      category_param_map: '',
      class_list_path: '',
      class_list_method: '',
      class_list_body_type: '',
      class_list_param_map: '',
      pause_path: '',
      pause_method: '',
      pause_body_type: '',
      pause_param_map: '',
      pause_id_param: '',
      resume_path: '',
      resume_method: '',
      resume_body_type: '',
      resume_param_map: '',
      change_pass_path: '',
      change_pass_method: '',
      change_pass_body_type: '',
      change_pass_param_map: '',
      change_pass_param: '',
      change_pass_id_param: '',
      resubmit_path: '',
      resubmit_method: '',
      resubmit_body_type: '',
      resubmit_param_map: '',
      resubmit_id_param: '',
      log_path: '',
      log_method: '',
      log_body_type: '',
      log_param_map: '',
      log_id_param: '',
      balance_path: '',
      balance_money_field: '',
      balance_method: '',
      balance_body_type: '',
      balance_param_map: '',
      balance_auth_type: '',
      report_param_style: '',
      report_auth_type: '',
      report_path: '',
      report_method: '',
      report_body_type: '',
      report_param_map: '',
      get_report_path: '',
      get_report_method: '',
      get_report_body_type: '',
      get_report_param_map: '',
      refresh_path: '',
      source_code: ''
    }
  }

  function fillEditForm(value?: Partial<LegacyPlatformConfig>) {
    Object.assign(editForm, createDefaultEditForm(), value || {})
  }

  function formatJsonTemplate(raw: string) {
    try {
      return JSON.stringify(JSON.parse(raw), null, 2)
    } catch {
      return raw
    }
  }

  function firstPath(path?: string, fallbackPath = '') {
    const normalized = path?.trim()
    if (normalized) return normalized
    return fallbackPath.trim()
  }

  function formatEndpoint(path?: string, queryDriver = '') {
    if (path?.trim()) return path.trim()
    if (queryDriver?.trim()) return `专用驱动: ${queryDriver.trim()}`
    return '-'
  }

  function hasCapability(item: LegacyPlatformConfig, capability: string) {
    if (capability === 'balance') return Boolean(item.balance_path)
    if (capability === 'batch_progress') return Boolean(item.batch_progress_path)
    if (capability === 'report') return Boolean(item.report_path || item.get_report_path)
    if (capability === 'pause') return Boolean(item.pause_path)
    if (capability === 'resume') return Boolean(item.resume_path)
    if (capability === 'custom_query') return Boolean(item.query_act)
    return true
  }

  function featureTagNodes(row: LegacyPlatformConfig) {
    const features: Array<{ label: string; type: 'info' | 'success' | 'warning' | 'primary' }> = []
    if (row.use_json) features.push({ label: 'JSON', type: 'info' })
    if (row.returns_yid) features.push({ label: '返回YID', type: 'warning' })
    if (row.balance_path) features.push({ label: '余额', type: 'success' })
    if (row.batch_progress_path) features.push({ label: '批量进度', type: 'primary' })
    if (row.report_path || row.get_report_path) features.push({ label: '工单', type: 'warning' })
    if (row.pause_path) features.push({ label: '暂停', type: 'info' })
    if (row.resume_path) features.push({ label: '恢复', type: 'info' })
    if (row.query_act) features.push({ label: row.query_act, type: 'primary' })
    if (!features.length) features.push({ label: '基础配置', type: 'info' })

    return h(
      'div',
      { class: 'flex flex-wrap gap-2' },
      features.map((item) => h(ElTag, { type: item.type, effect: 'plain' }, () => item.label))
    )
  }

  const { columns, columnChecks } = useTableColumns<LegacyPlatformConfig>(() => [
    {
      prop: 'pt',
      label: '平台',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || row.pt || '未命名平台'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, row.pt || '-')
        ])
    },
    {
      prop: 'auth_type',
      label: '认证',
      width: 120,
      formatter: (row) =>
        h(ElTag, { type: row.auth_type === 'uid_key' ? 'success' : 'warning', effect: 'plain' }, () => row.auth_type || '-')
    },
    {
      prop: 'query_path',
      label: '查课',
      minWidth: 180,
      formatter: (row) => formatEndpoint(row.query_path, row.query_act)
    },
    {
      prop: 'order_path',
      label: '下单',
      minWidth: 180,
      formatter: (row) => formatEndpoint(row.order_path)
    },
    {
      prop: 'progress_path',
      label: '进度',
      minWidth: 180,
      formatter: (row) => formatEndpoint(row.progress_path)
    },
    {
      prop: 'features',
      label: '能力',
      minWidth: 260,
      formatter: (row) => featureTagNodes(row)
    },
    {
      prop: 'updated_at',
      label: '更新时间',
      width: 180
    },
    {
      prop: 'operation',
      label: '操作',
      width: 150,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex items-center gap-2' }, [
          h(ArtButtonTable, { type: 'edit', onClick: () => openEdit(row) }),
          h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
        ])
    }
  ])

  function resetSearch() {
    searchForm.keyword = ''
    searchForm.auth_type = ''
    searchForm.capability = ''
  }

  async function loadData() {
    loading.value = true
    try {
      list.value = (await fetchLegacyPlatformConfigs()) || []
    } finally {
      loading.value = false
    }
  }

  function openAdd() {
    isEditing.value = false
    activeEditTab.value = 'query-order'
    fillEditForm()
    dialogVisible.value = true
  }

  function openEdit(record: LegacyPlatformConfig) {
    isEditing.value = true
    activeEditTab.value = 'query-order'
    fillEditForm(record)
    dialogVisible.value = true
  }

  async function handleSave() {
    if (!String(editForm.pt || '').trim()) return ElMessage.warning('请先填写平台标识')
    if (!String(editForm.name || '').trim()) return ElMessage.warning('请先填写平台名称')
    saving.value = true
    try {
      await saveLegacyPlatformConfig({ ...editForm })
      dialogVisible.value = false
      ElMessage.success('平台配置已保存')
      await loadData()
    } finally {
      saving.value = false
    }
  }

  async function handleDelete(record: LegacyPlatformConfig) {
    await ElMessageBox.confirm(`确定删除平台「${record.name || record.pt}」吗？`, '删除平台', {
      type: 'warning'
    })
    await deleteLegacyPlatformConfig(record.pt)
    ElMessage.success('平台已删除')
    await loadData()
  }

  function openPHPImport() {
    phpCode.value = ''
    phpResult.value = null
    phpVisible.value = true
  }

  async function parsePHP() {
    if (!phpCode.value.trim()) return ElMessage.warning('请先粘贴 PHP 代码')
    phpLoading.value = true
    try {
      phpResult.value = await parseLegacyPlatformPHPCode(phpCode.value)
      ElMessage.success(`解析完成，置信度 ${phpResult.value?.confidence || 0}%`)
    } finally {
      phpLoading.value = false
    }
  }

  function applyPHPResult() {
    if (!phpResult.value) return
    const result = phpResult.value
    fillEditForm({
      pt: result.pt,
      name: result.name,
      auth_type: result.auth_type || 'uid_key',
      success_codes: result.success_codes || '0',
      use_json: result.use_json,
      query_act: result.query_act || '',
      query_path: firstPath(result.query_path, '/api.php?act=get'),
      query_method: 'POST',
      order_path: firstPath(result.order_path, '/api.php?act=add'),
      order_method: 'POST',
      progress_path: firstPath(result.progress_path, '/api.php?act=chadan2'),
      progress_method: result.progress_method || 'POST',
      category_path: '/api.php?act=getcate',
      category_method: 'POST',
      class_list_path: '/api.php?act=getclass',
      class_list_method: 'POST',
      pause_path: firstPath(result.pause_path, '/api.php?act=zt'),
      pause_method: 'POST',
      pause_id_param: result.pause_id_param || 'id',
      change_pass_path: firstPath(result.change_pass_path, '/api.php?act=gaimi'),
      change_pass_method: 'POST',
      change_pass_param: result.change_pass_param || 'newPwd',
      change_pass_id_param: result.change_pass_id_param || 'id',
      resubmit_path: '/api.php?act=budan',
      resubmit_method: 'POST',
      resubmit_id_param: 'id',
      log_path: firstPath(result.log_path, '/api.php?act=xq'),
      log_method: 'POST',
      log_id_param: result.log_id_param || 'id',
      returns_yid: result.returns_yid,
      balance_path: firstPath(result.balance_path, '/api.php?act=getmoney'),
      balance_money_field: result.balance_money_field || 'money',
      balance_method: 'POST',
      source_code: phpCode.value
    })
    phpVisible.value = false
    dialogVisible.value = true
    ElMessage.success('已将解析结果填入编辑器')
  }

  function openDetect() {
    detectForm.url = ''
    detectForm.uid = ''
    detectForm.key = ''
    detectForm.token = ''
    detectForm.cookie = ''
    detectForm.pt = ''
    detectForm.name = ''
    detectResult.value = null
    detectVisible.value = true
  }

  async function runDetect() {
    if (!detectForm.url.trim()) return ElMessage.warning('请先填写平台 URL')
    detectLoading.value = true
    try {
      detectResult.value = await detectLegacyPlatform({ ...detectForm })
      ElMessage.success(detectResult.value?.success ? '检测完成' : '检测完成，但未识别到稳定接口')
    } finally {
      detectLoading.value = false
    }
  }

  async function autoDetectSave() {
    if (!detectForm.url.trim()) return ElMessage.warning('请先填写平台 URL')
    if (!detectForm.pt.trim()) return ElMessage.warning('一键保存前请填写平台标识')
    detectSaveLoading.value = true
    try {
      const result = await autoDetectSaveLegacyPlatform({ ...detectForm })
      detectResult.value = result.detect
      ElMessage.success(result.msg || '检测已完成')
      if (result.success) {
        await loadData()
      }
    } finally {
      detectSaveLoading.value = false
    }
  }

  function applyDetectResult() {
    if (!detectResult.value) return
    const config = detectResult.value.config || {}
    fillEditForm({
      pt: detectForm.pt || '',
      name: detectForm.name || detectResult.value.suggested_name || '',
      auth_type: config.auth_type || 'uid_key',
      success_codes: config.success_codes || '0',
      use_json: config.use_json === 'true',
      query_act: config.query_act || '',
      query_path: firstPath(config.query_path, '/api.php?act=get'),
      query_method: config.query_method || 'POST',
      query_body_type: config.query_body_type || '',
      query_param_style: config.query_param_style || '',
      query_param_map: config.query_param_map || '',
      progress_path: firstPath(config.progress_path, '/api.php?act=chadan2'),
      progress_method: config.progress_method || 'POST',
      progress_body_type: config.progress_body_type || '',
      progress_param_map: config.progress_param_map || '',
      batch_progress_path: firstPath(config.batch_progress_path),
      batch_progress_method: config.batch_progress_method || '',
      batch_progress_body_type: config.batch_progress_body_type || '',
      batch_progress_param_map: config.batch_progress_param_map || '',
      order_path: firstPath(config.order_path, '/api.php?act=add'),
      order_method: config.order_method || 'POST',
      order_body_type: config.order_body_type || '',
      order_param_map: config.order_param_map || '',
      category_path: firstPath(config.category_path, '/api.php?act=getcate'),
      category_method: config.category_method || 'POST',
      category_body_type: config.category_body_type || '',
      category_param_map: config.category_param_map || '',
      class_list_path: firstPath(config.class_list_path, '/api.php?act=getclass'),
      class_list_method: config.class_list_method || 'POST',
      class_list_body_type: config.class_list_body_type || '',
      class_list_param_map: config.class_list_param_map || '',
      pause_path: firstPath(config.pause_path, '/api.php?act=zt'),
      pause_method: config.pause_method || 'POST',
      pause_body_type: config.pause_body_type || '',
      pause_param_map: config.pause_param_map || '',
      pause_id_param: config.pause_id_param || 'id',
      resume_path: firstPath(config.resume_path),
      resume_method: config.resume_method || 'POST',
      resume_body_type: config.resume_body_type || '',
      resume_param_map: config.resume_param_map || '',
      change_pass_path: firstPath(config.change_pass_path, '/api.php?act=gaimi'),
      change_pass_method: config.change_pass_method || 'POST',
      change_pass_body_type: config.change_pass_body_type || '',
      change_pass_param_map: config.change_pass_param_map || '',
      change_pass_param: config.change_pass_param || 'newPwd',
      change_pass_id_param: config.change_pass_id_param || 'id',
      resubmit_path: firstPath(config.resubmit_path, '/api.php?act=budan'),
      resubmit_method: config.resubmit_method || 'POST',
      resubmit_body_type: config.resubmit_body_type || '',
      resubmit_param_map: config.resubmit_param_map || '',
      resubmit_id_param: config.resubmit_id_param || 'id',
      log_path: firstPath(config.log_path, '/api.php?act=xq'),
      log_method: config.log_method || 'POST',
      log_body_type: config.log_body_type || '',
      log_param_map: config.log_param_map || '',
      log_id_param: config.log_id_param || 'id',
      returns_yid: config.returns_yid === 'true',
      balance_path: firstPath(config.balance_path, '/api.php?act=getmoney'),
      balance_money_field: config.balance_money_field || 'money',
      balance_method: config.balance_method || 'POST',
      balance_body_type: config.balance_body_type || '',
      balance_param_map: config.balance_param_map || '',
      balance_auth_type: config.balance_auth_type || '',
      report_path: config.report_path || '',
      report_method: config.report_method || 'POST',
      report_body_type: config.report_body_type || '',
      report_param_map: config.report_param_map || '',
      report_param_style: config.report_param_style || '',
      report_auth_type: config.report_auth_type || '',
      get_report_path: config.get_report_path || '',
      get_report_method: config.get_report_method || 'POST',
      get_report_body_type: config.get_report_body_type || '',
      get_report_param_map: config.get_report_param_map || '',
      refresh_path: config.refresh_path || ''
    })
    detectVisible.value = false
    dialogVisible.value = true
    ElMessage.success('已将检测结果填入编辑器')
  }

  function probeTagType(status: string): 'danger' | 'info' | 'primary' | 'success' | 'warning' {
    if (status === 'ok') return 'success'
    if (status === 'fail') return 'danger'
    if (status === 'timeout' || status === 'error') return 'warning'
    return 'info'
  }

  onMounted(() => {
    loadData()
  })
</script>

<style scoped>
  .platform-edit-tab-placeholder {
    padding: 16px;
    border: 1px dashed var(--art-card-border);
    border-radius: var(--custom-radius-sm);
    color: var(--art-gray-500);
  }
</style>
