<template>
  <div class="admin-upstream-config-page art-full-height">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap gap-3">
        <ElTag effect="plain">对接中心</ElTag>
        <ElTag effect="plain">模块 {{ configuredModulesCount }}/11</ElTag>
        <ElTag :type="configuredModulesCount === 11 ? 'success' : 'warning'" effect="plain">
          待配置 {{ 11 - configuredModulesCount }}
        </ElTag>
        <ElTag type="info" effect="plain">
          YF {{ yfdkProjects.length }} / XM {{ xmProviders.length }} / 项目 {{ xmProjects.length }}
        </ElTag>
      </div>
      <ElButton plain :loading="loading" @click="refreshAll">刷新全部</ElButton>
    </div>

    <ElTabs v-model="activeTab">
      <ElTabPane label="基础对接" name="basic">
        <div v-loading="loading" class="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">图图强国</h3>
                <p class="mt-1.5 text-sm text-g-500">基础地址、密钥和加价倍率。</p>
              </div>
              <ElTag :type="isConfigured(tutuqgConfig.base_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(tutuqgConfig.base_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
                <ElInput v-model="tutuqgConfig.base_url" placeholder="https://api.example.com" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">密钥</label>
                <ElInput v-model="tutuqgConfig.key" type="password" show-password placeholder="请输入密钥" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">加价倍率</label>
                <ElInputNumber v-model="tutuqgConfig.price_increment" class="w-full" :min="0" :step="0.1" :precision="2" />
              </div>
              <ElButton plain :loading="savingKey === 'tutuqg'" @click="saveTutuQG">保存图图强国</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">YF 打卡</h3>
                <p class="mt-1.5 text-sm text-g-500">用于 YF 订单创建与项目同步。</p>
              </div>
              <ElTag :type="isConfigured(yfdkConfig.base_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(yfdkConfig.base_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
                <ElInput v-model="yfdkConfig.base_url" placeholder="https://dk.blwl.fun/api/" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
                <ElInput v-model="yfdkConfig.token" type="password" show-password placeholder="请输入 Token" />
              </div>
              <ElButton plain :loading="savingKey === 'yfdk'" @click="saveYFDK">保存 YF 打卡</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">泰山打卡</h3>
                <p class="mt-1.5 text-sm text-g-500">对接地址、令牌和管理账号。</p>
              </div>
              <ElTag :type="isConfigured(sxdkConfig.base_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(sxdkConfig.base_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
                <ElInput v-model="sxdkConfig.base_url" placeholder="http://..." />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
                <ElInput v-model="sxdkConfig.token" type="password" show-password placeholder="请输入 Token" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">管理账号</label>
                <ElInput v-model="sxdkConfig.admin" placeholder="请输入管理账号" />
              </div>
              <ElButton plain :loading="savingKey === 'sxdk'" @click="saveSXDK">保存泰山打卡</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">HZW 实时进度</h3>
                <p class="mt-1 text-sm text-g-500">Socket 地址与推送启停。</p>
              </div>
              <ElTag :type="isConfigured(hzwSocketConfig.socket_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(hzwSocketConfig.socket_url) ? '已启用' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Socket 地址</label>
                <ElInput v-model="hzwSocketConfig.socket_url" placeholder="http://socket.biedawo.org" />
              </div>
              <p class="text-xs leading-5 text-g-500">保存后会自动启用实时推送。</p>
              <ElButton plain :loading="savingKey === 'hzw'" @click="saveHZW">保存 HZW Socket</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">Appui 打卡</h3>
                <p class="mt-1.5 text-sm text-g-500">保留最常用的连接参数和加价配置。</p>
              </div>
              <ElTag :type="isConfigured(appuiConfig.base_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(appuiConfig.base_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
                <ElInput v-model="appuiConfig.base_url" placeholder="请输入 Appui 地址" />
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">UID</label>
                  <ElInput v-model="appuiConfig.uid" placeholder="上游 UID" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">Key</label>
                  <ElInput v-model="appuiConfig.key" type="password" show-password placeholder="上游 Key" />
                </div>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">加价</label>
                <ElInputNumber v-model="appuiConfig.price_increment" class="w-full" :min="0" :step="0.1" :precision="2" />
              </div>
              <ElButton plain :loading="savingKey === 'appui'" @click="saveAppui">保存 Appui</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">闪电运动</h3>
                <p class="mt-1.5 text-sm text-g-500">基础地址、超时与默认售价。</p>
              </div>
              <ElTag :type="isConfigured(sdxyConfig.base_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(sdxyConfig.base_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">基础地址</label>
                <ElInput v-model="sdxyConfig.base_url" placeholder="请输入闪电运动地址" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Endpoint</label>
                <ElInput v-model="sdxyConfig.endpoint" placeholder="/flash/api.php" />
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">UID</label>
                  <ElInput v-model="sdxyConfig.uid" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">Key</label>
                  <ElInput v-model="sdxyConfig.key" type="password" show-password />
                </div>
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">超时</label>
                  <ElInputNumber v-model="sdxyConfig.timeout" class="w-full" :min="1" :step="1" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">默认售价</label>
                  <ElInputNumber v-model="sdxyConfig.price" class="w-full" :min="0" :step="0.1" :precision="2" />
                </div>
              </div>
              <ElButton plain :loading="savingKey === 'sdxy'" @click="saveSDXY">保存闪电运动</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">运动世界</h3>
                <p class="mt-1.5 text-sm text-g-500">晨跑、锻炼和真实成本倍率。</p>
              </div>
              <ElTag :type="isConfigured(ydsjConfig.base_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(ydsjConfig.base_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">基础地址</label>
                <ElInput v-model="ydsjConfig.base_url" placeholder="请输入运动世界地址" />
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
                  <ElInput v-model="ydsjConfig.token" type="password" show-password />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">UID</label>
                  <ElInput v-model="ydsjConfig.uid" />
                </div>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Key</label>
                <ElInput v-model="ydsjConfig.key" type="password" show-password />
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
                  <ElInputNumber v-model="ydsjConfig.price_multiple" class="w-full" :min="0" :step="0.1" :precision="2" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">真实成本倍率</label>
                  <ElInputNumber v-model="ydsjConfig.real_cost_multiple" class="w-full" :min="0" :step="0.1" :precision="2" />
                </div>
              </div>
              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">晨跑售价</label>
                  <ElInputNumber v-model="ydsjConfig.xbd_morning_price" class="w-full" :min="0" :step="0.1" :precision="2" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">锻炼售价</label>
                  <ElInputNumber v-model="ydsjConfig.xbd_exercise_price" class="w-full" :min="0" :step="0.1" :precision="2" />
                </div>
              </div>
              <ElButton plain :loading="savingKey === 'ydsj'" @click="saveYDSJ">保存运动世界</ElButton>
            </div>
          </section>

          <section class="art-card-sm p-5 md:col-span-2 2xl:col-span-3">
            <div class="flex items-start justify-between gap-3 border-b-d pb-4">
              <div>
                <h3 class="text-lg font-semibold text-g-900">永夜运动</h3>
                <p class="mt-1.5 text-sm text-g-500">维护 API、定价和公告配置。</p>
              </div>
              <ElTag :type="isConfigured(yongyeConfig.api_url) ? 'success' : 'info'" effect="plain">
                {{ isConfigured(yongyeConfig.api_url) ? '已配置' : '未配置' }}
              </ElTag>
            </div>

            <div class="mt-5 grid gap-4 lg:grid-cols-2 xl:grid-cols-4">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
                <ElInput v-model="yongyeConfig.api_url" placeholder="请输入永夜 API" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
                <ElInput v-model="yongyeConfig.token" type="password" show-password />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">单价</label>
                <ElInputNumber v-model="yongyeConfig.dj" class="w-full" :min="0" :step="0.1" :precision="2" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">折算</label>
                <ElInputNumber v-model="yongyeConfig.zs" class="w-full" :min="0" :step="0.01" :precision="2" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">倍率</label>
                <ElInputNumber v-model="yongyeConfig.beis" class="w-full" :min="0" :step="0.01" :precision="2" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">续增单价</label>
                <ElInputNumber v-model="yongyeConfig.xzdj" class="w-full" :min="0" :step="0.1" :precision="2" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">续增默认值</label>
                <ElInputNumber v-model="yongyeConfig.xzmo" class="w-full" :min="0" :step="1" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">退款比例</label>
                <ElInputNumber v-model="yongyeConfig.tk" class="w-full" :min="0" :step="0.01" :precision="2" />
              </div>
            </div>

            <div class="mt-4 grid gap-4 lg:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">说明内容</label>
                <ElInput v-model="yongyeConfig.content" type="textarea" :rows="3" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">弹窗公告</label>
                <ElInput v-model="yongyeConfig.tcgg" type="textarea" :rows="3" />
              </div>
            </div>

            <div class="mt-4">
              <ElButton plain :loading="savingKey === 'yongye'" @click="saveYongye">保存永夜运动</ElButton>
            </div>
          </section>
        </div>
      </ElTabPane>
      <ElTabPane label="论文与打卡商品" name="content">
        <div v-loading="loading" class="grid gap-4 xl:grid-cols-[0.85fr_1.15fr]">
          <div class="space-y-4">
            <section class="art-card-sm p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h3 class="text-lg font-semibold text-g-900">土拨鼠论文</h3>
                  <p class="mt-1 text-sm text-g-500">倍率与页面入口控制。</p>
                </div>
                <ElTag :type="tuboshuConfig.price_ratio > 0 ? 'success' : 'info'" effect="plain">
                  {{ tuboshuConfig.price_ratio > 0 ? '已配置' : '未配置' }}
                </ElTag>
              </div>

              <div class="mt-5 space-y-4">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
                  <ElInputNumber v-model="tuboshuConfig.price_ratio" class="w-full" :min="0.1" :step="0.5" :precision="1" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">页面显示控制</label>
                  <ElCheckboxGroup v-model="selectedTuboshuPages" class="grid gap-3 md:grid-cols-2">
                    <ElCheckbox v-for="item in tuboshuPageOptions" :key="item.key" :value="item.key">
                      {{ item.label }}
                    </ElCheckbox>
                  </ElCheckboxGroup>
                </div>
                <p class="text-xs leading-5 text-g-500">详细价格仍在论文页面维护，这里只保留全局入口配置。</p>
                <ElButton plain :loading="savingKey === 'tuboshu'" @click="saveTuboshu">保存土拨鼠</ElButton>
              </div>
            </section>

            <section class="art-card-sm p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h3 class="text-lg font-semibold text-g-900">凸知打卡</h3>
                  <p class="mt-1.5 text-sm text-g-500">维护账号配置和商品覆盖价。</p>
                </div>
                <ElTag :type="isConfigured(tuzhiConfig.daka_api_username) ? 'success' : 'info'" effect="plain">
                  {{ isConfigured(tuzhiConfig.daka_api_username) ? '已配置' : '未配置' }}
                </ElTag>
              </div>

              <div class="mt-5 space-y-4">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">上游账号</label>
                  <ElInput v-model="tuzhiConfig.daka_api_username" placeholder="请输入上游账号" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">上游密码</label>
                  <ElInput v-model="tuzhiConfig.daka_api_password" type="password" show-password placeholder="请输入上游密码" />
                </div>
                <div class="flex flex-wrap gap-3">
                  <ElButton plain :loading="savingKey === 'tuzhi'" @click="saveTuzhi">保存凸知配置</ElButton>
                  <ElButton plain :loading="tuzhiGoodsLoading" @click="loadTuzhiGoods">刷新商品</ElButton>
                  <ElButton plain :loading="tuzhiOverridesSaving" @click="saveTuzhiOverrides">保存商品覆盖</ElButton>
                </div>
              </div>
            </section>
          </div>

          <div class="space-y-4">
            <section class="art-card-sm p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h3 class="text-lg font-semibold text-g-900">智文论文</h3>
                  <p class="mt-1 text-sm text-g-500">论文相关价格集中维护。</p>
                </div>
                <ElTag :type="isConfigured(paperConfig.lunwen_api_username) ? 'success' : 'info'" effect="plain">
                  {{ isConfigured(paperConfig.lunwen_api_username) ? '已配置' : '未配置' }}
                </ElTag>
              </div>

              <div class="mt-5 grid gap-4 xl:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">API 账号</label>
                  <ElInput v-model="paperConfig.lunwen_api_username" placeholder="请输入登录账号" />
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">API 密码</label>
                  <ElInput v-model="paperConfig.lunwen_api_password" type="password" show-password placeholder="请输入登录密码" />
                </div>
                <div v-for="field in paperFields" :key="field.key">
                  <label class="mb-2 block text-sm font-medium text-g-800">{{ field.label }}</label>
                  <ElInput v-model="paperConfig[field.key]" />
                </div>
              </div>

              <div class="mt-4">
                <ElButton plain :loading="savingKey === 'paper'" @click="savePaper">保存智文论文</ElButton>
              </div>
            </section>

            <ElCard class="art-table-card">
              <ArtTableHeader v-model:columns="tuzhiColumnChecks" :loading="tuzhiGoodsLoading" @refresh="loadTuzhiGoods">
                <template #left>
                  <div class="flex flex-wrap items-center gap-2">
                    <ElTag effect="plain">凸知商品 {{ tuzhiGoods.length }} 个</ElTag>
                  </div>
                </template>
              </ArtTableHeader>

              <ArtTable :data="tuzhiGoods" :columns="tuzhiColumns" :show-table-header="true">
                <template #upstream_price="{ row }">
                  <span class="font-medium text-g-900">{{ currency(row.price) }}</span>
                </template>
                <template #billing="{ row }">
                  <ElTag :type="row.billing_method === 2 ? 'warning' : 'primary'" effect="plain">
                    {{ row.billing_method === 2 ? '按月' : '按日' }}
                  </ElTag>
                </template>
                <template #override_price="{ row }">
                  <ElInputNumber
                    :model-value="getTuzhiOverride(row.id).price"
                    :min="0"
                    :step="0.01"
                    :precision="2"
                    class="w-full"
                    @update:model-value="setTuzhiOverrideField(row.id, 'price', Number($event || 0))"
                  />
                </template>
                <template #enabled="{ row }">
                  <ElSelect
                    :model-value="getTuzhiOverride(row.id).enabled"
                    class="w-full"
                    @update:model-value="setTuzhiOverrideField(row.id, 'enabled', Number($event))"
                  >
                    <ElOption :value="1" label="上架" />
                    <ElOption :value="0" label="下架" />
                  </ElSelect>
                </template>
              </ArtTable>
            </ElCard>
          </div>
        </div>
      </ElTabPane>
      <ElTabPane label="项目管理" name="projects">
        <div class="grid gap-4 xl:grid-cols-[0.82fr_1.18fr]">
          <ElCard class="art-table-card">
            <ArtTableHeader v-model:columns="yfdkColumnChecks" :loading="yfdkLoading" @refresh="loadYFDKProjects">
              <template #left>
                <div class="flex flex-wrap items-center gap-2">
                  <ElButton plain size="small" :loading="yfdkSyncing" @click="syncYFDKProjectList">从上游同步</ElButton>
                  <ElTag effect="plain">YF 项目 {{ yfdkProjects.length }} 个</ElTag>
                </div>
              </template>
            </ArtTableHeader>

            <ArtTable :data="yfdkProjects" :columns="yfdkColumns" :show-table-header="true">
              <template #action="{ row }">
                <div class="flex flex-wrap gap-2">
                  <ElButton link type="primary" @click="openYFDKEdit(row)">编辑</ElButton>
                  <ElButton link type="danger" @click="removeYFDKProject(row.id)">删除</ElButton>
                </div>
              </template>
            </ArtTable>
          </ElCard>

          <ElCard class="art-table-card">
            <ArtTableHeader v-model:columns="xmProviderColumnChecks" :loading="xmProvidersLoading" @refresh="loadXMProviders">
              <template #left>
                <div class="flex flex-wrap items-center gap-2">
                  <ElButton plain size="small" @click="openXMProviderAdd">新增连接</ElButton>
                  <ElTag effect="plain">XM 连接 {{ xmProviders.length }} 条</ElTag>
                </div>
              </template>
            </ArtTableHeader>

            <ArtTable :data="xmProviders" :columns="xmProviderColumns" :show-table-header="true">
              <template #actions="{ row }">
                <div class="flex flex-wrap gap-2">
                  <ElButton link type="primary" @click="openXMProviderEdit(row)">编辑</ElButton>
                  <ElButton link @click="handleXMProviderTest(row)">测试</ElButton>
                  <ElButton link @click="openXMImport(row)">导入项目</ElButton>
                  <ElButton link @click="openXMSync(row)">同步项目</ElButton>
                  <ElButton link type="danger" @click="removeXMProvider(row.id)">删除</ElButton>
                </div>
              </template>
            </ArtTable>
          </ElCard>
        </div>

        <ElCard class="art-table-card mt-4">
          <ArtTableHeader v-model:columns="xmProjectColumnChecks" :loading="xmProjectsLoading" @refresh="loadXMProjects">
            <template #left>
              <div class="flex flex-wrap items-center gap-2">
                <ElTag effect="plain">XM 项目 {{ xmProjects.length }} 个</ElTag>
              </div>
            </template>
          </ArtTableHeader>

          <ArtTable :data="xmProjects" :columns="xmProjectColumns" :show-table-header="true">
            <template #actions="{ row }">
              <div class="flex flex-wrap gap-2">
                <ElButton link type="primary" @click="openXMProjectEdit(row)">编辑</ElButton>
                <ElButton link type="danger" @click="removeXMProject(row.id)">删除</ElButton>
              </div>
            </template>
          </ArtTable>
        </ElCard>
      </ElTabPane>
    </ElTabs>

    <ElDialog v-model="yfdkEditVisible" title="编辑 YF 项目" width="520px">
      <div class="grid gap-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">说明</label>
          <ElInput v-model="yfdkEditForm.content" type="textarea" :rows="3" placeholder="项目描述" />
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">售价</label>
            <ElInputNumber v-model="yfdkEditForm.sell_price" class="w-full" :min="0" :step="0.01" :precision="2" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
            <ElInputNumber v-model="yfdkEditForm.sort" class="w-full" :min="0" :step="1" />
          </div>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
          <ElSelect v-model="yfdkEditForm.enabled" class="w-full">
            <ElOption :value="1" label="启用" />
            <ElOption :value="0" label="禁用" />
          </ElSelect>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="yfdkEditVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingKey === 'yfdk-project'" @click="submitYFDKEdit">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="xmProviderVisible" :title="xmProviderForm.id ? '编辑连接' : '新增连接'" width="720px">
      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">连接名称</label>
          <ElInput v-model="xmProviderForm.name" placeholder="例如：Spiderman 主号" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">认证方式</label>
          <ElSelect v-model="xmProviderForm.auth_type" class="w-full">
            <ElOption :value="0" label="UID + Key" />
            <ElOption :value="1" label="Token" />
          </ElSelect>
        </div>
      </div>

      <div class="mt-4">
        <label class="mb-2 block text-sm font-medium text-g-800">API 地址</label>
        <ElInput v-model="xmProviderForm.base_url" placeholder="例如：https://spiderman.sbs/api/xm_apis.php" />
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-3">
        <div v-if="Number(xmProviderForm.auth_type || 0) === 0">
          <label class="mb-2 block text-sm font-medium text-g-800">UID</label>
          <ElInput v-model="xmProviderForm.uid" placeholder="上游 UID" />
        </div>
        <div v-if="Number(xmProviderForm.auth_type || 0) === 0">
          <label class="mb-2 block text-sm font-medium text-g-800">Key</label>
          <ElInput v-model="xmProviderForm.key" type="password" show-password placeholder="上游 Key" />
        </div>
        <div v-if="Number(xmProviderForm.auth_type || 0) === 1">
          <label class="mb-2 block text-sm font-medium text-g-800">Token</label>
          <ElInput v-model="xmProviderForm.token" type="password" show-password placeholder="上游 Token" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
          <ElSelect v-model="xmProviderForm.status" class="w-full">
            <ElOption :value="0" label="正常" />
            <ElOption :value="1" label="停用" />
          </ElSelect>
        </div>
      </div>

      <div class="mt-4">
        <label class="mb-2 block text-sm font-medium text-g-800">备注</label>
        <ElInput v-model="xmProviderForm.remark" type="textarea" :rows="3" placeholder="记录来源、限制或同步说明" />
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="xmProviderVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingKey === 'xm-provider'" @click="submitXMProvider">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="xmProjectVisible" title="编辑 XM 项目" width="720px">
      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">项目名称</label>
          <ElInput v-model="xmProjectForm.name" placeholder="本地展示名称" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">来源连接</label>
          <ElInput :model-value="xmProjectForm.provider_name" disabled />
        </div>
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">上游项目 ID</label>
          <ElInput :model-value="xmProjectForm.p_id" disabled />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
          <ElInputNumber v-model="xmProjectForm.sort_order" class="w-full" :min="0" :step="1" />
        </div>
      </div>

      <div class="mt-4">
        <label class="mb-2 block text-sm font-medium text-g-800">项目说明</label>
        <ElInput v-model="xmProjectForm.description" type="textarea" :rows="3" placeholder="本地展示说明" />
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">本地售价</label>
          <ElInputNumber v-model="xmProjectForm.price" class="w-full" :min="0" :step="0.1" :precision="4" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">上游价格</label>
          <ElInputNumber v-model="xmProjectForm.upstream_price" class="w-full" :min="0" :step="0.1" :precision="4" disabled />
        </div>
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-3">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
          <ElSelect v-model="xmProjectForm.status" class="w-full">
            <ElOption :value="0" label="正常" />
            <ElOption :value="1" label="下架" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">支持查询</label>
          <ElSelect v-model="xmProjectForm.query" class="w-full">
            <ElOption :value="1" label="是" />
            <ElOption :value="0" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">需要密码</label>
          <ElSelect v-model="xmProjectForm.password" class="w-full">
            <ElOption :value="1" label="是" />
            <ElOption :value="0" label="否" />
          </ElSelect>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="xmProjectVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingKey === 'xm-project'" @click="submitXMProject">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="xmImportVisible" title="导入上游项目" width="980px">
      <ElAlert
        class="mb-4"
        type="info"
        show-icon
        :closable="false"
        :title="xmImportProvider ? `当前连接：${xmImportProvider.name}` : '请选择连接'"
      />

      <div class="grid gap-4 md:grid-cols-3">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">本地价格倍率</label>
          <ElInputNumber v-model="xmImportForm.price_multiplier" class="w-full" :min="0" :step="0.1" :precision="2" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">附加价</label>
          <ElInputNumber v-model="xmImportForm.price_addition" class="w-full" :step="0.1" :precision="2" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">覆盖本地价格</label>
          <ElSelect v-model="xmImportForm.overwrite_local_price" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
      </div>

      <ElTable
        class="mt-4"
        :data="xmImportProjects"
        border
        height="420"
        row-key="id"
        v-loading="xmImportLoading"
        @selection-change="handleXMImportSelectionChange"
      >
        <ElTableColumn type="selection" width="48" />
        <ElTableColumn prop="id" label="上游 ID" width="100" />
        <ElTableColumn prop="name" label="项目名称" min-width="220" />
        <ElTableColumn prop="price" label="上游价格" width="120">
          <template #default="{ row }">{{ currency(row.price) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="query" label="支持查询" width="110">
          <template #default="{ row }">
            <ElTag :type="row.query === 1 ? 'success' : 'info'" effect="plain">
              {{ row.query === 1 ? '支持' : '不支持' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="password" label="需要密码" width="110">
          <template #default="{ row }">
            <ElTag :type="row.password === 1 ? 'warning' : 'info'" effect="plain">
              {{ row.password === 1 ? '需要' : '不需要' }}
            </ElTag>
          </template>
        </ElTableColumn>
      </ElTable>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="xmImportVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingKey === 'xm-import'" @click="submitXMImport">确认导入</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="xmSyncVisible" title="同步上游项目" width="760px">
      <ElAlert
        class="mb-4"
        type="info"
        show-icon
        :closable="false"
        :title="xmSyncProvider ? `将从连接 ${xmSyncProvider.name} 拉取并更新本地项目。` : '请选择连接'"
      />

      <div class="grid gap-4 md:grid-cols-3">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">同步名称</label>
          <ElSelect v-model="xmSyncForm.sync_name" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">同步说明</label>
          <ElSelect v-model="xmSyncForm.sync_description" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">同步上游价格</label>
          <ElSelect v-model="xmSyncForm.sync_upstream_price" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">同步查询能力</label>
          <ElSelect v-model="xmSyncForm.sync_query" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">同步密码要求</label>
          <ElSelect v-model="xmSyncForm.sync_password" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">重算本地价格</label>
          <ElSelect v-model="xmSyncForm.overwrite_local_price" class="w-full">
            <ElOption :value="true" label="是" />
            <ElOption :value="false" label="否" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">价格倍率</label>
          <ElInputNumber v-model="xmSyncForm.price_multiplier" class="w-full" :min="0" :step="0.1" :precision="2" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">附加价</label>
          <ElInputNumber v-model="xmSyncForm.price_addition" class="w-full" :step="0.1" :precision="2" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="xmSyncVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="savingKey === 'xm-sync'" @click="submitXMSync">开始同步</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    deleteXMProject,
    deleteXMProvider,
    deleteYFDKProject,
    fetchAppuiConfig,
    fetchHZWSocketConfig,
    fetchPaperConfig,
    fetchSDXYConfig,
    fetchSXDKConfig,
    fetchTutuQGConfig,
    fetchTuboshuConfig,
    fetchTuZhiConfig,
    fetchTuZhiGoods,
    fetchTuZhiGoodsOverrides,
    fetchXMProjects,
    fetchXMProviderProjects,
    fetchXMProviders,
    fetchYDSJConfig,
    fetchYFDKConfig,
    fetchYFDKProjects,
    fetchYongyeConfig,
    importXMProviderProjects,
    saveAppuiConfig,
    saveHZWSocketConfig,
    savePaperConfig,
    saveSDXYConfig,
    saveSXDKConfig,
    saveTutuQGConfig,
    saveTuboshuConfig,
    saveTuZhiConfig,
    saveTuZhiGoodsOverrides,
    saveXMProject,
    saveXMProvider,
    saveYDSJConfig,
    saveYFDKConfig,
    saveYongyeConfig,
    syncXMProviderProjects,
    syncYFDKProjects,
    testXMProvider,
    updateYFDKProject,
    type AppuiConfig,
    type HZWSocketConfig,
    type PaperConfig,
    type SDXYConfig,
    type SXDKConfig,
    type TutuQGUpstreamConfig,
    type TuboshuUpstreamConfig,
    type TuZhiAdminGoods,
    type TuZhiConfig,
    type TuZhiGoodsOverride,
    type XMProjectItem,
    type XMProviderItem,
    type XMUpstreamProjectItem,
    type YDSJConfig,
    type YFDKAdminProject,
    type YFDKConfig,
    type YongyeConfig
  } from '@/api/legacy/admin-upstream'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminUpstreamConfigPage' })

  const activeTab = ref('basic')
  const loading = ref(false)
  const savingKey = ref('')

  const tutuqgConfig = reactive<TutuQGUpstreamConfig>({ base_url: '', key: '', price_increment: 0 })
  const yfdkConfig = reactive<YFDKConfig>({ base_url: '', token: '' })
  const sxdkConfig = reactive<SXDKConfig>({ base_url: '', token: '', admin: '' })
  const hzwSocketConfig = reactive<HZWSocketConfig>({ socket_url: '' })
  const tuboshuConfig = reactive<TuboshuUpstreamConfig>({ price_ratio: 5, price_config: {}, page_visibility: {} })
  const appuiConfig = reactive<AppuiConfig>({ base_url: '', uid: '', key: '', price_increment: 0, courses: [] })
  const sdxyConfig = reactive<SDXYConfig>({
    base_url: '',
    endpoint: '/flash/api.php',
    uid: '',
    key: '',
    timeout: 30,
    price: 10
  })
  const ydsjConfig = reactive<YDSJConfig>({
    base_url: '',
    token: '',
    uid: '',
    key: '',
    price_multiple: 5,
    xbd_morning_price: 6,
    xbd_exercise_price: 6.5,
    real_cost_multiple: 1
  })
  const yongyeConfig = reactive<YongyeConfig>({
    api_url: '',
    token: '',
    dj: 0,
    zs: 1.25,
    beis: 1.3,
    xzdj: 0,
    xzmo: 100,
    tk: 0.01,
    content: '',
    tcgg: ''
  })
  const paperConfig = reactive<PaperConfig>({
    lunwen_api_username: '',
    lunwen_api_password: '',
    lunwen_api_6000_price: '30',
    lunwen_api_8000_price: '40',
    lunwen_api_10000_price: '50',
    lunwen_api_12000_price: '60',
    lunwen_api_15000_price: '75',
    lunwen_api_rws_price: '10',
    lunwen_api_ktbg_price: '10',
    lunwen_api_jdaigchj_price: '10',
    lunwen_api_xgdl_price: '3',
    lunwen_api_jcl_price: '3',
    lunwen_api_jdaigcl_price: '3'
  })
  const tuzhiConfig = reactive<TuZhiConfig>({ daka_api_username: '', daka_api_password: '' })

  const tuboshuPageOptions = [
    { key: 'ComponentStagePage', label: '分步对话' },
    { key: 'ChatPage', label: 'AI 对话' },
    { key: 'ChartPage', label: '图表生成' },
    { key: 'TemplatePage', label: '模板中心' },
    { key: 'ReductionPage', label: '论文降重' },
    { key: 'AccountTable', label: '账户管理' },
    { key: 'TicketPage', label: '工单系统' }
  ]

  const paperFields: Array<{ key: keyof PaperConfig; label: string }> = [
    { key: 'lunwen_api_6000_price', label: '6000 字价格' },
    { key: 'lunwen_api_8000_price', label: '8000 字价格' },
    { key: 'lunwen_api_10000_price', label: '10000 字价格' },
    { key: 'lunwen_api_12000_price', label: '12000 字价格' },
    { key: 'lunwen_api_15000_price', label: '15000 字价格' },
    { key: 'lunwen_api_rws_price', label: '任务书价格' },
    { key: 'lunwen_api_ktbg_price', label: '开题报告价格' },
    { key: 'lunwen_api_jdaigchj_price', label: '降 AIGC + 查重' },
    { key: 'lunwen_api_xgdl_price', label: '段落修改' },
    { key: 'lunwen_api_jcl_price', label: '文本降重' },
    { key: 'lunwen_api_jdaigcl_price', label: '降 AIGC 率' }
  ]

  const selectedTuboshuPages = computed({
    get: () => tuboshuPageOptions.filter((item) => tuboshuConfig.page_visibility?.[item.key] !== false).map((item) => item.key),
    set: (values: string[]) => {
      const next: Record<string, boolean> = {}
      for (const item of tuboshuPageOptions) next[item.key] = values.includes(item.key)
      tuboshuConfig.page_visibility = next
    }
  })

  const yfdkProjects = ref<YFDKAdminProject[]>([])
  const yfdkLoading = ref(false)
  const yfdkSyncing = ref(false)
  const yfdkEditVisible = ref(false)
  const yfdkEditForm = reactive({ id: 0, sell_price: 0, enabled: 1, sort: 0, content: '' })

  const tuzhiGoods = ref<TuZhiAdminGoods[]>([])
  const tuzhiOverrides = ref<TuZhiGoodsOverride[]>([])
  const tuzhiGoodsLoading = ref(false)
  const tuzhiOverridesSaving = ref(false)

  const xmProviders = ref<XMProviderItem[]>([])
  const xmProvidersLoading = ref(false)
  const xmProviderVisible = ref(false)
  const xmProviderForm = reactive<Partial<XMProviderItem>>({
    id: 0,
    name: '',
    base_url: '',
    auth_type: 0,
    uid: '',
    key: '',
    token: '',
    status: 0,
    remark: ''
  })

  const xmProjects = ref<XMProjectItem[]>([])
  const xmProjectsLoading = ref(false)
  const xmProjectVisible = ref(false)
  const xmProjectForm = reactive<Partial<XMProjectItem>>({
    id: 0,
    provider_id: 0,
    provider_name: '',
    name: '',
    description: '',
    price: 0,
    upstream_price: 0,
    query: 1,
    password: 1,
    p_id: '',
    status: 0,
    sort_order: 0,
    sync_mode: 1
  })

  const xmImportVisible = ref(false)
  const xmImportLoading = ref(false)
  const xmImportProvider = ref<XMProviderItem | null>(null)
  const xmImportProjects = ref<XMUpstreamProjectItem[]>([])
  const xmSelectedImportProjectIds = ref<string[]>([])
  const xmImportForm = reactive({ price_multiplier: 1, price_addition: 0, overwrite_local_price: true })

  const xmSyncVisible = ref(false)
  const xmSyncProvider = ref<XMProviderItem | null>(null)
  const xmSyncForm = reactive({
    provider_id: 0,
    sync_name: true,
    sync_description: true,
    sync_upstream_price: true,
    sync_query: true,
    sync_password: true,
    overwrite_local_price: false,
    price_multiplier: 1,
    price_addition: 0
  })

  const moduleStatusList = computed(() => [
    { key: 'tutuqg', label: '图图强国', ready: isConfigured(tutuqgConfig.base_url), detail: tutuqgConfig.base_url || '等待配置地址' },
    { key: 'yfdk', label: 'YF 打卡', ready: isConfigured(yfdkConfig.base_url), detail: yfdkConfig.base_url || '等待配置地址' },
    { key: 'sxdk', label: '泰山打卡', ready: isConfigured(sxdkConfig.base_url), detail: sxdkConfig.base_url || '等待配置地址' },
    { key: 'hzw', label: 'HZW 实时进度', ready: isConfigured(hzwSocketConfig.socket_url), detail: hzwSocketConfig.socket_url || '等待配置地址' },
    { key: 'tuboshu', label: '土拨鼠论文', ready: tuboshuConfig.price_ratio > 0, detail: `倍率 ${tuboshuConfig.price_ratio || 0}` },
    { key: 'appui', label: 'Appui 打卡', ready: isConfigured(appuiConfig.base_url), detail: appuiConfig.base_url || '等待配置地址' },
    { key: 'sdxy', label: '闪电运动', ready: isConfigured(sdxyConfig.base_url), detail: sdxyConfig.base_url || '等待配置地址' },
    { key: 'ydsj', label: '运动世界', ready: isConfigured(ydsjConfig.base_url), detail: ydsjConfig.base_url || '等待配置地址' },
    { key: 'yongye', label: '永夜运动', ready: isConfigured(yongyeConfig.api_url), detail: yongyeConfig.api_url || '等待配置地址' },
    { key: 'paper', label: '智文论文', ready: isConfigured(paperConfig.lunwen_api_username), detail: paperConfig.lunwen_api_username || '等待配置账号' },
    { key: 'tuzhi', label: '凸知打卡', ready: isConfigured(tuzhiConfig.daka_api_username), detail: tuzhiConfig.daka_api_username || '等待配置账号' }
  ])

  const configuredModulesCount = computed(() => moduleStatusList.value.filter((item) => item.ready).length)

  const currency = (value: number | string | undefined) => `¥${Number(value || 0).toFixed(2)}`

  function isConfigured(value: string | number | undefined | null) {
    if (typeof value === 'number') return value > 0
    return String(value || '').trim().length > 0
  }

  function applyReactive<T extends object>(target: T, value?: Partial<T> | null) {
    if (value) Object.assign(target, value)
  }

  function notifyPartialFailures(results: PromiseSettledResult<unknown>[], labels: string[]) {
    const failed = results
      .map((result, index) => ({ result, label: labels[index] }))
      .filter((item) => item.result.status === 'rejected')
      .map((item) => item.label)
    if (failed.length) ElMessage.warning(`部分配置加载失败：${failed.join('、')}`)
  }

  async function loadConfigs() {
    const results = await Promise.allSettled([
      fetchTutuQGConfig(),
      fetchYFDKConfig(),
      fetchSXDKConfig(),
      fetchHZWSocketConfig(),
      fetchTuboshuConfig(),
      fetchAppuiConfig(),
      fetchSDXYConfig(),
      fetchYDSJConfig(),
      fetchYongyeConfig(),
      fetchPaperConfig(),
      fetchTuZhiConfig()
    ])

    notifyPartialFailures(results, ['图图强国', 'YF 打卡', '泰山打卡', 'HZW Socket', '土拨鼠', 'Appui', '闪电运动', '运动世界', '永夜运动', '智文论文', '凸知打卡'])

    if (results[0].status === 'fulfilled') applyReactive(tutuqgConfig, results[0].value)
    if (results[1].status === 'fulfilled') applyReactive(yfdkConfig, results[1].value)
    if (results[2].status === 'fulfilled') applyReactive(sxdkConfig, results[2].value)
    if (results[3].status === 'fulfilled') applyReactive(hzwSocketConfig, results[3].value)
    if (results[4].status === 'fulfilled') applyReactive(tuboshuConfig, results[4].value)
    if (results[5].status === 'fulfilled') applyReactive(appuiConfig, results[5].value)
    if (results[6].status === 'fulfilled') applyReactive(sdxyConfig, results[6].value)
    if (results[7].status === 'fulfilled') applyReactive(ydsjConfig, results[7].value)
    if (results[8].status === 'fulfilled') applyReactive(yongyeConfig, results[8].value)
    if (results[9].status === 'fulfilled') applyReactive(paperConfig, results[9].value)
    if (results[10].status === 'fulfilled') applyReactive(tuzhiConfig, results[10].value)
  }

  async function loadYFDKProjects() {
    yfdkLoading.value = true
    try {
      yfdkProjects.value = (await fetchYFDKProjects()) || []
    } finally {
      yfdkLoading.value = false
    }
  }

  async function loadTuzhiGoods() {
    tuzhiGoodsLoading.value = true
    try {
      const [goods, overrides] = await Promise.all([fetchTuZhiGoods(), fetchTuZhiGoodsOverrides()])
      tuzhiGoods.value = goods || []
      tuzhiOverrides.value = overrides || []
    } finally {
      tuzhiGoodsLoading.value = false
    }
  }

  async function loadXMProviders() {
    xmProvidersLoading.value = true
    try {
      xmProviders.value = (await fetchXMProviders()) || []
    } finally {
      xmProvidersLoading.value = false
    }
  }

  async function loadXMProjects() {
    xmProjectsLoading.value = true
    try {
      xmProjects.value = (await fetchXMProjects()) || []
    } finally {
      xmProjectsLoading.value = false
    }
  }

  async function refreshAll() {
    loading.value = true
    try {
      await Promise.all([loadConfigs(), loadYFDKProjects(), loadTuzhiGoods(), loadXMProviders(), loadXMProjects()])
    } finally {
      loading.value = false
    }
  }

  async function withSaving(key: string, action: () => Promise<void>, successMessage: string) {
    savingKey.value = key
    try {
      await action()
      ElMessage.success(successMessage)
    } finally {
      savingKey.value = ''
    }
  }

  const saveTutuQG = () => withSaving('tutuqg', async () => saveTutuQGConfig({ ...tutuqgConfig }), '图图强国配置已保存')
  const saveYFDK = () => withSaving('yfdk', async () => saveYFDKConfig({ ...yfdkConfig }), 'YF 打卡配置已保存')
  const saveSXDK = () => withSaving('sxdk', async () => saveSXDKConfig({ ...sxdkConfig }), '泰山打卡配置已保存')
  const saveHZW = () => withSaving('hzw', async () => saveHZWSocketConfig({ ...hzwSocketConfig }), 'HZW Socket 配置已保存')
  const saveTuboshu = () => withSaving('tuboshu', async () => saveTuboshuConfig({ ...tuboshuConfig }), '土拨鼠配置已保存')
  const saveAppui = () => withSaving('appui', async () => saveAppuiConfig({ ...appuiConfig }), 'Appui 配置已保存')
  const saveSDXY = () => withSaving('sdxy', async () => saveSDXYConfig({ ...sdxyConfig }), '闪电运动配置已保存')
  const saveYDSJ = () => withSaving('ydsj', async () => saveYDSJConfig({ ...ydsjConfig }), '运动世界配置已保存')
  const saveYongye = () => withSaving('yongye', async () => saveYongyeConfig({ ...yongyeConfig }), '永夜运动配置已保存')
  const savePaper = () => withSaving('paper', async () => savePaperConfig({ ...paperConfig }), '智文论文配置已保存')
  const saveTuzhi = () => withSaving('tuzhi', async () => saveTuZhiConfig({ ...tuzhiConfig }), '凸知打卡配置已保存')

  function getTuzhiOverride(goodsId: number) {
    return tuzhiOverrides.value.find((item) => item.goods_id === goodsId) || { goods_id: goodsId, price: 0, enabled: 1 }
  }

  function setTuzhiOverrideField(goodsId: number, field: 'price' | 'enabled', value: number) {
    const found = tuzhiOverrides.value.find((item) => item.goods_id === goodsId)
    if (found) {
      if (field === 'price') found.price = value
      else found.enabled = value
      return
    }
    const next: TuZhiGoodsOverride = { goods_id: goodsId, price: 0, enabled: 1 }
    if (field === 'price') next.price = value
    else next.enabled = value
    tuzhiOverrides.value.push(next)
  }

  async function saveTuzhiOverrides() {
    tuzhiOverridesSaving.value = true
    try {
      await saveTuZhiGoodsOverrides(tuzhiOverrides.value)
      ElMessage.success('凸知商品覆盖已保存')
    } finally {
      tuzhiOverridesSaving.value = false
    }
  }

  async function syncYFDKProjectList() {
    yfdkSyncing.value = true
    try {
      const result = await syncYFDKProjects()
      await loadYFDKProjects()
      ElMessage.success(result?.msg || `YF 项目同步完成，共 ${result?.count || 0} 条`)
    } finally {
      yfdkSyncing.value = false
    }
  }

  function openYFDKEdit(record: YFDKAdminProject) {
    Object.assign(yfdkEditForm, {
      id: record.id,
      sell_price: Number(record.sell_price || 0),
      enabled: Number(record.enabled || 0),
      sort: Number(record.sort || 0),
      content: record.content || ''
    })
    yfdkEditVisible.value = true
  }

  async function submitYFDKEdit() {
    await withSaving('yfdk-project', async () => {
      await updateYFDKProject({ ...yfdkEditForm })
      yfdkEditVisible.value = false
      await loadYFDKProjects()
    }, 'YF 项目已保存')
  }

  async function removeYFDKProject(id: number) {
    await ElMessageBox.confirm('确定删除这个 YF 项目吗？', '删除确认', { type: 'warning' })
    await deleteYFDKProject(id)
    ElMessage.success('YF 项目已删除')
    await loadYFDKProjects()
  }

  function openXMProviderAdd() {
    Object.assign(xmProviderForm, {
      id: 0,
      name: '',
      base_url: '',
      auth_type: 0,
      uid: '',
      key: '',
      token: '',
      status: 0,
      remark: ''
    })
    xmProviderVisible.value = true
  }

  function openXMProviderEdit(record: XMProviderItem) {
    Object.assign(xmProviderForm, { ...record })
    xmProviderVisible.value = true
  }

  async function submitXMProvider() {
    if (!String(xmProviderForm.name || '').trim()) return ElMessage.warning('连接名称不能为空')
    if (!String(xmProviderForm.base_url || '').trim()) return ElMessage.warning('API 地址不能为空')
    await withSaving('xm-provider', async () => {
      await saveXMProvider({ ...xmProviderForm })
      xmProviderVisible.value = false
      await Promise.all([loadXMProviders(), loadXMProjects()])
    }, xmProviderForm.id ? '连接已更新' : '连接已新增')
  }

  async function handleXMProviderTest(record: XMProviderItem) {
    savingKey.value = `xm-provider-test-${record.id}`
    try {
      const result = await testXMProvider(record.id)
      ElMessage.success(`${result?.message || '连接成功'}，拉到 ${result?.project_count || 0} 个项目`)
    } finally {
      savingKey.value = ''
    }
  }

  async function removeXMProvider(id: number) {
    await ElMessageBox.confirm('确定删除这个上游连接吗？', '删除确认', { type: 'warning' })
    await deleteXMProvider(id)
    ElMessage.success('上游连接已删除')
    await Promise.all([loadXMProviders(), loadXMProjects()])
  }

  async function openXMImport(record: XMProviderItem) {
    xmImportLoading.value = true
    xmImportProvider.value = record
    try {
      const projects = await fetchXMProviderProjects(record.id)
      xmImportProjects.value = projects || []
      xmSelectedImportProjectIds.value = xmImportProjects.value.map((item) => String(item.id))
      Object.assign(xmImportForm, { price_multiplier: 1, price_addition: 0, overwrite_local_price: true })
      xmImportVisible.value = true
    } finally {
      xmImportLoading.value = false
    }
  }

  function handleXMImportSelectionChange(rows: XMUpstreamProjectItem[]) {
    xmSelectedImportProjectIds.value = rows.map((item) => String(item.id))
  }

  async function submitXMImport() {
    if (!xmImportProvider.value) return
    if (!xmSelectedImportProjectIds.value.length) return ElMessage.warning('请至少选择一个项目')
    await withSaving('xm-import', async () => {
      const result = await importXMProviderProjects({
        provider_id: xmImportProvider.value!.id,
        project_ids: xmSelectedImportProjectIds.value,
        price_multiplier: xmImportForm.price_multiplier,
        price_addition: xmImportForm.price_addition,
        overwrite_local_price: xmImportForm.overwrite_local_price
      })
      xmImportVisible.value = false
      await Promise.all([loadXMProviders(), loadXMProjects()])
      ElMessage.success(`导入完成：新增 ${result?.summary?.created || 0}，更新 ${result?.summary?.updated || 0}`)
    }, 'XM 项目导入完成')
  }

  function openXMSync(record: XMProviderItem) {
    xmSyncProvider.value = record
    Object.assign(xmSyncForm, {
      provider_id: record.id,
      sync_name: true,
      sync_description: true,
      sync_upstream_price: true,
      sync_query: true,
      sync_password: true,
      overwrite_local_price: false,
      price_multiplier: 1,
      price_addition: 0
    })
    xmSyncVisible.value = true
  }

  async function submitXMSync() {
    await withSaving('xm-sync', async () => {
      const result = await syncXMProviderProjects({ ...xmSyncForm })
      xmSyncVisible.value = false
      await Promise.all([loadXMProviders(), loadXMProjects()])
      ElMessage.success(`同步完成：更新 ${result?.summary?.updated || 0}，跳过 ${result?.summary?.skipped || 0}`)
    }, 'XM 项目同步完成')
  }

  function openXMProjectEdit(record: XMProjectItem) {
    Object.assign(xmProjectForm, { ...record })
    xmProjectVisible.value = true
  }

  async function submitXMProject() {
    await withSaving('xm-project', async () => {
      await saveXMProject({ ...xmProjectForm })
      xmProjectVisible.value = false
      await loadXMProjects()
    }, 'XM 项目已保存')
  }

  async function removeXMProject(id: number) {
    await ElMessageBox.confirm('确定删除这个 XM 项目吗？', '删除确认', { type: 'warning' })
    await deleteXMProject(id)
    ElMessage.success('XM 项目已删除')
    await loadXMProjects()
  }

  const { columns: tuzhiColumns, columnChecks: tuzhiColumnChecks } = useTableColumns<TuZhiAdminGoods>(() => [
    { prop: 'id', label: 'ID', width: 80 },
    { prop: 'name', label: '商品名称', minWidth: 220 },
    { prop: 'upstream_price', label: '上游价格', width: 120, useSlot: true },
    { prop: 'billing', label: '计费方式', width: 110, useSlot: true },
    { prop: 'override_price', label: '覆盖售价', width: 160, useSlot: true },
    { prop: 'enabled', label: '上架', width: 120, useSlot: true }
  ])

  const { columns: yfdkColumns, columnChecks: yfdkColumnChecks } = useTableColumns<YFDKAdminProject>(() => [
    { prop: 'cid', label: '上游 ID', width: 110 },
    { prop: 'name', label: '项目名称', minWidth: 180 },
    { prop: 'cost_price', label: '成本价', width: 100, formatter: (row) => currency(row.cost_price) },
    { prop: 'sell_price', label: '售价', width: 100, formatter: (row) => h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, currency(row.sell_price)) },
    {
      prop: 'enabled',
      label: '状态',
      width: 100,
      formatter: (row) => h(ElTag, { type: row.enabled === 1 ? 'success' : 'danger', effect: 'plain' }, () => (row.enabled === 1 ? '启用' : '禁用'))
    },
    { prop: 'sort', label: '排序', width: 80, align: 'center' },
    { prop: 'action', label: '操作', width: 140, fixed: 'right', useSlot: true }
  ])

  const { columns: xmProviderColumns, columnChecks: xmProviderColumnChecks } = useTableColumns<XMProviderItem>(() => [
    { prop: 'id', label: 'ID', width: 70 },
    { prop: 'name', label: '连接名称', minWidth: 160 },
    { prop: 'auth_type', label: '认证', width: 100, formatter: (row) => (Number(row.auth_type) === 1 ? 'Token' : 'UID + Key') },
    { prop: 'base_url', label: 'API 地址', minWidth: 240 },
    { prop: 'project_count', label: '项目数', width: 90, align: 'center' },
    { prop: 'last_sync_at', label: '最近同步', width: 180 },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) => h(ElTag, { type: Number(row.status) === 0 ? 'success' : 'info', effect: 'plain' }, () => (Number(row.status) === 0 ? '正常' : '停用'))
    },
    { prop: 'actions', label: '操作', minWidth: 280, fixed: 'right', useSlot: true }
  ])

  const { columns: xmProjectColumns, columnChecks: xmProjectColumnChecks } = useTableColumns<XMProjectItem>(() => [
    { prop: 'id', label: 'ID', width: 70 },
    { prop: 'name', label: '项目名称', minWidth: 180 },
    { prop: 'provider_name', label: '来源连接', width: 140 },
    { prop: 'price', label: '本地售价', width: 110, formatter: (row) => currency(row.price) },
    { prop: 'upstream_price', label: '上游价格', width: 110, formatter: (row) => currency(row.upstream_price) },
    { prop: 'p_id', label: '上游项目 ID', width: 120 },
    {
      prop: 'status',
      label: '状态',
      width: 100,
      formatter: (row) => h(ElTag, { type: Number(row.status) === 0 ? 'success' : 'info', effect: 'plain' }, () => (Number(row.status) === 0 ? '正常' : '下架'))
    },
    { prop: 'actions', label: '操作', width: 140, fixed: 'right', useSlot: true }
  ])

  onMounted(() => {
    refreshAll()
  })
</script>

<style scoped>
  .admin-upstream-config-page :deep(.el-checkbox) {
    margin-right: 0;
  }
</style>
