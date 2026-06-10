<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col">
    <section class="art-card-sm p-5" v-loading="loading">
      <div class="mb-4 flex flex-col gap-3 border-b-d pb-4 md:flex-row md:items-center md:justify-between">
        <h1 class="text-lg font-semibold text-g-900">系统设置</h1>
        <div class="flex flex-wrap gap-2 md:justify-end">
          <ElButton plain :loading="loading" @click="loadConfig">刷新</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存当前配置</ElButton>
        </div>
      </div>

      <ElTabs v-model="activeTab" class="settings-tabs">
        <ElTabPane label="网站配置" name="wzpz">
          <div class="grid gap-4 xl:grid-cols-[1.04fr_0.96fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">站点品牌</h2>
                  <p class="mt-1 text-sm text-g-500">站点名称、SEO、Logo 与资源站点标识。</p>
                </div>
                <ElTag effect="plain">基础信息</ElTag>
              </div>

              <div class="mt-5 grid gap-4 md:grid-cols-2">
                <div class="md:col-span-2">
                  <label class="mb-2 block text-sm font-medium text-g-800">站点名称</label>
                  <ElInput v-model="basicForm.sitename" maxlength="60" placeholder="请输入站点名称" />
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">资源版本号</label>
                  <ElInput v-model="basicForm.version" maxlength="40" placeholder="例如：1.0.1" />
                  <p class="mt-2 text-sm text-g-500">用于标识当前站点资源或系统版本，便于后台核对发布批次。</p>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">SEO 关键词</label>
                  <ElInput v-model="basicForm.keywords" maxlength="200" placeholder="例如：网课代刷,在线查课" />
                </div>

                <div class="md:col-span-2">
                  <label class="mb-2 block text-sm font-medium text-g-800">SEO 描述</label>
                  <ElInput v-model="basicForm.description" maxlength="200" placeholder="请输入 SEO 描述" />
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">主 Logo</label>
                  <ElInput v-model="basicForm.logo" placeholder="https://... 或 /uploads/..." />
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">头部 Logo</label>
                  <ElInput v-model="basicForm.hlogo" placeholder="https://... 或 /uploads/..." />
                </div>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">站点开关与公告</h2>
                  <p class="mt-1 text-sm text-g-500">保留旧版网站级开关语义，并集中维护公告文案。</p>
                </div>
                <ElTag effect="plain">网站级</ElTag>
              </div>

              <div class="mt-5 space-y-5">
                <div class="grid gap-3 sm:grid-cols-2">
                  <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">全站水印</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">继续兼容旧版 `sykg`，开启后后台页面显示防截图水印。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.sykg === '1'" @change="(value) => updateSwitch('sykg', value)" />
                    </div>
                  </article>

                  <article
                    class="rounded-custom-sm border-[var(--el-color-danger-light-7)] bg-[var(--el-color-danger-light-9)] px-4 py-3"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">维护模式</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">开启后普通用户前台访问将受限，仅保留管理员处理问题与排查入口。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.bz === '1'" @change="(value) => updateSwitch('bz', value)" />
                    </div>
                  </article>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">登录页公告</label>
                  <ElInput
                    v-model="basicForm.notice"
                    type="textarea"
                    :rows="5"
                    resize="none"
                    placeholder="登录前展示的站点公告"
                  />
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">弹窗公告</label>
                  <ElInput
                    v-model="basicForm.tcgonggao"
                    type="textarea"
                    :rows="4"
                    resize="none"
                    placeholder="全局弹窗公告内容"
                  />
                  <p class="mt-2 text-sm text-g-500">与登录页公告独立，保留旧版 `tcgonggao` 配置语义，便于继续兼容历史弹窗链路。</p>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">商城主域名后缀</label>
                  <ElInput
                    v-model="basicForm.mall_domain_suffix"
                    placeholder="例如：mall.example.com"
                    @blur="basicForm.mall_domain_suffix = normalizeDomain(basicForm.mall_domain_suffix)"
                  />
                  <p class="mt-2 text-sm text-g-500">租户店铺页会基于这个后缀自动切换为子域名前缀模式。</p>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">商城开通价格</label>
                  <ElInput v-model="businessForm.mall_open_price" placeholder="例如：99" />
                  <p class="mt-2 text-sm text-g-500">租户开通商城时展示并使用该价格。</p>
                </div>
              </div>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="代理配置" name="dlpz">
          <div class="grid gap-4 xl:grid-cols-[0.96fr_1.04fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">代理权限与开户</h2>
                  <p class="mt-1 text-sm text-g-500">维护代理迁移、开户与跨户充值主链路设置。</p>
                </div>
                <ElTag effect="plain">主链路</ElTag>
              </div>

              <div class="mt-5 space-y-5">
                <div class="grid gap-3 sm:grid-cols-2">
                  <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">上级迁移</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">控制用户中心的上级迁移入口。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.sjqykg === '1'" @change="(value) => updateSwitch('sjqykg', value)" />
                    </div>
                  </article>

                  <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">允许代理开户</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">关闭后，代理列表中的新增代理链路会被后端直接拦截。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.user_htkh === '1'" @change="(value) => updateSwitch('user_htkh', value)" />
                    </div>
                  </article>

                  <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3 sm:col-span-2">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">邀请注册</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">开启后，注册页会要求并校验邀请码。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.user_yqzc === '1'" @change="(value) => updateSwitch('user_yqzc', value)" />
                    </div>
                  </article>
                </div>

                <div class="grid gap-4 md:grid-cols-2">
                  <div v-for="item in agentNumericItems" :key="item.key">
                    <label class="mb-2 block text-sm font-medium text-g-800">{{ item.label }}</label>
                    <ElInput v-model="businessForm[item.key]" :placeholder="item.placeholder" />
                    <p class="mt-2 text-sm text-g-500">{{ item.desc }}</p>
                  </div>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">跨户充值 UID 白名单</label>
                  <ElInput
                    v-model="basicForm.cross_recharge_uids"
                    type="textarea"
                    :rows="4"
                    resize="none"
                    placeholder="例如：1001,1002,1003"
                  />
                  <p class="mt-2 text-sm text-g-500">保持英文逗号分隔字符串，只有名单内账号可使用跨户充值入口。</p>
                </div>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">平开代理限制</h2>
                  <p class="mt-1 text-sm text-g-500">控制顶级平开与同级平开的限制规则，直接影响新增代理扣费逻辑。</p>
                </div>
                <ElTag effect="plain">{{ agentPolicyItems.length }} 项</ElTag>
              </div>

              <div class="mt-5 grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">平开控制</label>
                  <ElSelect v-model="agentPolicyForm.dl_pkkg" class="w-full">
                    <ElOption
                      v-for="item in agentPolicyOptions"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </ElSelect>
                  <p class="mt-2 text-sm text-g-500">沿用旧后台规则：无限制、禁止顶级平开，或按条件要求双倍余额。</p>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">顶级代理费率定义</label>
                  <ElInput v-model="businessForm.djfl" placeholder="例如：0.50" />
                  <p class="mt-2 text-sm text-g-500">建议填写两位小数；后端会据此判断是否属于顶级平开。</p>
                </div>
              </div>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="支付配置" name="zfpz">
          <div class="grid gap-4 xl:grid-cols-[0.92fr_1.08fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">充值限制</h2>
                  <p class="mt-1 text-sm text-g-500">控制在线充值门槛与代理充值权限。</p>
                </div>
                <ElTag effect="plain">{{ paymentSwitchItems.length + paymentNumericItems.length }} 项</ElTag>
              </div>

              <div class="mt-5 space-y-4">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">最低在线充值金额</label>
                  <ElInput v-model="businessForm.zdpay" placeholder="例如：10" />
                  <p class="mt-2 text-sm text-g-500">用户在线支付时，低于该金额会被后端直接拒绝。</p>
                </div>

                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">非直系在线充值</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">关闭后，非直系代理只能按页面提示联系上级充值。</p>
                    </div>
                    <ElSwitch
                      :model-value="switchForm.non_direct_recharge_enable === '1'"
                      @change="(value) => updateSwitch('non_direct_recharge_enable', value)"
                    />
                  </div>
                </article>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">充值赠送规则</h2>
                  <p class="mt-1 text-sm text-g-500">维护普通充值规则与活动日赠送规则。</p>
                </div>
                <ElTag effect="plain">{{ totalBonusRuleCount }} 条</ElTag>
              </div>

              <div class="mt-5 space-y-5">
                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <div>
                      <p class="text-sm font-medium text-g-900">启用充值赠送</p>
                      <p class="mt-1 text-sm text-g-500">关闭后用户充值页不展示赠送规则。</p>
                    </div>
                    <ElSwitch v-model="bonusForm.enabled" />
                  </div>
                </article>

                <div class="space-y-3">
                  <div class="flex items-center justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">常规规则</p>
                      <p class="mt-1 text-xs text-g-500">按充值金额区间计算赠送比例。</p>
                    </div>
                    <ElButton plain @click="addBonusRule('rules')">新增规则</ElButton>
                  </div>

                  <div class="space-y-3">
                    <article
                      v-for="(rule, index) in bonusForm.rules"
                      :key="`base-rule-${index}`"
                      class="rounded-custom-sm border-full-d bg-g-100/40 p-4"
                    >
                      <div class="grid gap-4 md:grid-cols-[1fr_1fr_1fr_auto] md:items-end">
                        <div>
                          <label class="mb-2 block text-sm font-medium text-g-800">最低金额</label>
                          <ElInputNumber v-model="rule.min" class="w-full" :min="0" :precision="2" />
                        </div>
                        <div>
                          <label class="mb-2 block text-sm font-medium text-g-800">最高金额</label>
                          <ElInputNumber v-model="rule.max" class="w-full" :min="0" :precision="2" />
                        </div>
                        <div>
                          <label class="mb-2 block text-sm font-medium text-g-800">赠送比例(%)</label>
                          <ElInputNumber v-model="rule.bonus_pct" class="w-full" :min="0" :precision="2" />
                        </div>
                        <div class="flex items-end md:h-full">
                          <ElButton text type="danger" @click="removeBonusRule('rules', index)">删除</ElButton>
                        </div>
                      </div>
                    </article>
                  </div>

                  <ElEmpty v-if="!bonusForm.rules.length" description="暂无常规赠送规则" />
                </div>

                <article class="rounded-custom-sm border-full-d p-4">
                  <div class="flex items-center justify-between gap-3">
                    <div>
                      <p class="text-sm font-medium text-g-900">活动日赠送</p>
                      <p class="mt-1 text-sm text-g-500">活动日可覆盖常规规则并展示单独提示文案。</p>
                    </div>
                    <ElSwitch v-model="bonusForm.activity.enabled" />
                  </div>

                  <div class="mt-4 grid gap-4">
                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">活动提示文案</label>
                      <ElInput
                        v-model="bonusForm.activity.hint"
                        maxlength="120"
                        placeholder="例如：今日活动加成已开启，充值更划算。"
                      />
                    </div>

                    <div>
                      <label class="mb-2 block text-sm font-medium text-g-800">生效星期</label>
                      <div class="flex flex-wrap gap-2">
                        <ElButton
                          v-for="day in weekdayOptions"
                          :key="day.value"
                          size="small"
                          :type="bonusForm.activity.weekdays.includes(day.value) ? 'primary' : 'default'"
                          @click="toggleWeekday(day.value)"
                        >
                          {{ day.label }}
                        </ElButton>
                      </div>
                    </div>

                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">活动规则</p>
                        <p class="mt-1 text-xs text-g-500">
                          已选：{{ activeWeekdayLabels || '未选择生效日期' }}
                        </p>
                      </div>
                      <ElButton plain @click="addBonusRule('activity')">新增活动规则</ElButton>
                    </div>

                    <div class="space-y-3">
                      <article
                        v-for="(rule, index) in bonusForm.activity.rules"
                        :key="`activity-rule-${index}`"
                        class="rounded-custom-sm border-full-d bg-g-100/40 p-4"
                      >
                        <div class="grid gap-4 md:grid-cols-[1fr_1fr_1fr_auto] md:items-end">
                          <div>
                            <label class="mb-2 block text-sm font-medium text-g-800">最低金额</label>
                            <ElInputNumber v-model="rule.min" class="w-full" :min="0" :precision="2" />
                          </div>
                          <div>
                            <label class="mb-2 block text-sm font-medium text-g-800">最高金额</label>
                            <ElInputNumber v-model="rule.max" class="w-full" :min="0" :precision="2" />
                          </div>
                          <div>
                            <label class="mb-2 block text-sm font-medium text-g-800">赠送比例(%)</label>
                            <ElInputNumber v-model="rule.bonus_pct" class="w-full" :min="0" :precision="2" />
                          </div>
                          <div class="flex items-end md:h-full">
                            <ElButton text type="danger" @click="removeBonusRule('activity', index)">删除</ElButton>
                          </div>
                        </div>
                      </article>
                    </div>

                    <ElEmpty v-if="!bonusForm.activity.rules.length" description="暂无活动赠送规则" />
                  </div>
                </article>
              </div>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="前台配置" name="qtpz">
          <div class="grid gap-4 xl:grid-cols-[0.98fr_1.02fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">前台入口与公告</h2>
                  <p class="mt-1 text-sm text-g-500">维护前台工具入口、签到区公告和首页展示开关。</p>
                </div>
                <ElTag effect="plain">前台侧</ElTag>
              </div>

              <div class="mt-5 space-y-5">
                <div class="grid gap-3 sm:grid-cols-2">
                  <article
                    v-for="item in businessSwitchItems"
                    :key="item.key"
                    class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">{{ item.label }}</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">{{ item.desc }}</p>
                      </div>
                      <ElSwitch
                        :model-value="switchForm[item.key] === '1'"
                        @change="(value) => updateSwitch(item.key, value)"
                      />
                    </div>
                  </article>

                  <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">扫码下单</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">控制前台是否开启扫码下单模式。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.xdsmopen === '1'" @change="(value) => updateSwitch('xdsmopen', value)" />
                    </div>
                  </article>

                  <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3 sm:col-span-2">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-g-900">签到公告开关</p>
                        <p class="mt-1 text-xs leading-5 text-g-500">控制签到页或前台首页的公告提示区域是否显示。</p>
                      </div>
                      <ElSwitch :model-value="switchForm.qd_notice_open === '1'" @change="(value) => updateSwitch('qd_notice_open', value)" />
                    </div>
                  </article>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">签到页公告</label>
                  <ElInput
                    v-model="basicForm.qd_notice"
                    type="textarea"
                    :rows="4"
                    resize="none"
                    placeholder="签到或前台首页区域展示公告"
                  />
                </div>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">推荐渠道</h2>
                  <p class="mt-1 text-sm text-g-500">前台下单页直接消费这里的渠道配置。</p>
                </div>
                <ElTag effect="plain">{{ recommendChannelCount }} 条</ElTag>
              </div>

              <div class="mt-5 space-y-5">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <p class="text-sm font-semibold text-g-900">推荐渠道列表</p>
                    <p class="mt-1 text-xs text-g-500">保留当前 JSON 存储结构，不改旧版保存契约。</p>
                  </div>
                  <ElButton plain @click="addRecommendChannel">新增渠道</ElButton>
                </div>

                <div class="space-y-3">
                  <article
                    v-for="(item, index) in recommendChannels"
                    :key="`recommend-${index}`"
                    class="rounded-custom-sm border-full-d bg-g-100/40 p-4"
                  >
                    <div class="grid gap-4 md:grid-cols-[1fr_1fr_auto] md:items-start">
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">渠道名称</label>
                        <ElInput v-model="item.name" maxlength="30" placeholder="例如：官方渠道" />
                      </div>
                      <div>
                        <label class="mb-2 block text-sm font-medium text-g-800">渠道说明</label>
                        <ElInput v-model="item.desc" maxlength="80" placeholder="例如：主推支付线路" />
                      </div>
                      <div class="flex items-end md:h-full">
                        <ElButton text type="danger" @click="removeRecommendChannel(index)">删除</ElButton>
                      </div>
                    </div>
                  </article>
                </div>

                <ElEmpty v-if="!recommendChannels.length" description="暂无推荐渠道" />

                <div class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-4">
                  <div class="flex items-center justify-between gap-2">
                    <p class="text-sm font-semibold text-g-900">预览</p>
                    <ElTag effect="plain">{{ recommendChannelCount }} 条</ElTag>
                  </div>
                  <div class="mt-3 max-h-[320px] space-y-2 overflow-auto pr-1">
                    <article
                      v-for="(item, index) in normalizedRecommendChannels"
                      :key="`${item.name}-${index}`"
                      class="rounded-custom-sm border-full-d bg-[var(--el-bg-color)] px-3 py-2"
                    >
                      <p class="text-sm font-semibold text-g-900">{{ item.name }}</p>
                      <p v-if="item.desc" class="mt-1 text-xs leading-5 text-g-500">{{ item.desc }}</p>
                    </article>
                    <ElEmpty v-if="!normalizedRecommendChannels.length" description="暂无可预览渠道" />
                  </div>
                </div>
              </div>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="分类配置" name="flpz">
          <div class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">分类显示控制</h2>
                  <p class="mt-1 text-sm text-g-500">继续沿用旧版 `flkg` 分类开关控制前台分类展示。</p>
                </div>
                <ElTag effect="plain">2 项</ElTag>
              </div>

              <div class="mt-5 space-y-5">
                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">分类开关</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">控制前台课程分类面板是否展示。</p>
                    </div>
                    <ElSwitch :model-value="switchForm.flkg === '1'" @change="(value) => updateSwitch('flkg', value)" />
                  </div>
                </article>

                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">分类类型</label>
                  <ElSelect v-model="basicForm.fllx" class="w-full">
                    <ElOption label="侧边栏分类" value="0" />
                    <ElOption label="下单页面选择框分类" value="1" />
                    <ElOption label="下单页面单选框分类" value="2" />
                  </ElSelect>
                  <p class="mt-2 text-sm text-g-500">订单新增页会按该值切换课程分类选择方式。</p>
                </div>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d bg-g-100/40 p-5">
              <div>
                <h3 class="text-base font-semibold text-g-900">说明</h3>
                <p class="mt-2 text-sm leading-6 text-g-500">
                  这一栏先只保留真正已接通消费链路的分类显示总开关，不把旧版尚未在 `admin-next` 落地的长尾分类配置一起搬进来，避免把设置页重新变成“能存但暂时没人用”的杂项集合。
                </p>
              </div>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="登录设置" name="dlsz">
          <div class="grid gap-4 xl:grid-cols-[1fr_1fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">登录校验</h2>
                  <p class="mt-1 text-sm text-g-500">控制登录链路的校验方式与安全门槛。</p>
                </div>
                <ElTag effect="plain">登录链路</ElTag>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-2">
                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">滑块验证</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">登录时启用滑块校验。</p>
                    </div>
                    <ElSwitch
                      :model-value="switchForm.login_slider_verify === '1'"
                      @change="(value) => updateSwitch('login_slider_verify', value)"
                    />
                  </div>
                </article>

                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">邮箱验证</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">登录或注册链路启用邮箱校验。</p>
                    </div>
                    <ElSwitch
                      :model-value="switchForm.login_email_verify === '1'"
                      @change="(value) => updateSwitch('login_email_verify', value)"
                    />
                  </div>
                </article>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">后台登录附加项</h2>
                  <p class="mt-1 text-sm text-g-500">控制二级密码与找回密码等辅助入口。</p>
                </div>
                <ElTag effect="plain">辅助项</ElTag>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-2">
                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">二级密码</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">控制管理员登录二级密码校验。</p>
                    </div>
                    <ElSwitch :model-value="switchForm.pass2_kg === '1'" @change="(value) => updateSwitch('pass2_kg', value)" />
                  </div>
                </article>

                <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">忘记密码</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">控制登录页是否显示找回密码入口。</p>
                    </div>
                    <ElSwitch
                      :model-value="switchForm.login_forget_pwd === '1'"
                      @change="(value) => updateSwitch('login_forget_pwd', value)"
                    />
                  </div>
                </article>
              </div>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="签到设置" name="qdsz">
          <div class="grid gap-4 xl:grid-cols-[0.92fr_1.08fr]">
            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">签到开关</h2>
                  <p class="mt-1 text-sm text-g-500">控制签到功能启停与参与前置条件。</p>
                </div>
                <ElTag effect="plain">{{ checkinSwitchItems.length }} 项</ElTag>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-2">
                <article
                  v-for="item in checkinSwitchItems"
                  :key="item.key"
                  class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3"
                >
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-g-900">{{ item.label }}</p>
                      <p class="mt-1 text-xs leading-5 text-g-500">{{ item.desc }}</p>
                    </div>
                    <ElSwitch
                      :model-value="switchForm[item.key] === '1'"
                      @change="(value) => updateSwitch(item.key, value)"
                    />
                  </div>
                </article>
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-5">
              <div class="flex items-start justify-between gap-3 border-b-d pb-4">
                <div>
                  <h2 class="text-lg font-semibold text-g-900">签到门槛与奖励</h2>
                  <p class="mt-1 text-sm text-g-500">维护余额门槛、每日人数上限与随机奖励区间。</p>
                </div>
                <ElTag effect="plain">{{ checkinNumericItems.length }} 项</ElTag>
              </div>

              <div class="mt-5 grid gap-4 md:grid-cols-2">
                <div v-for="item in checkinNumericItems" :key="item.key">
                  <label class="mb-2 block text-sm font-medium text-g-800">{{ item.label }}</label>
                  <ElInput v-model="businessForm[item.key]" :placeholder="item.placeholder" />
                  <p class="mt-2 text-sm text-g-500">{{ item.desc }}</p>
                </div>
              </div>
            </article>
          </div>
        </ElTabPane>
      </ElTabs>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { fetchLegacyAdminConfig, saveLegacyAdminConfig } from '@/api/legacy/admin-settings'
  import { useSiteStore } from '@/store/modules/site'
  import type { LegacyAdminConfigMap } from '@/types/legacy-contract'

  defineOptions({ name: 'AdminSettingsPage' })

  interface RecommendChannelItem {
    desc: string
    name: string
  }

  interface BonusRule {
    bonus_pct: number
    max: number
    min: number
  }

  interface BonusActivity {
    enabled: boolean
    hint: string
    rules: BonusRule[]
    weekdays: number[]
  }

  interface BonusConfig {
    activity: BonusActivity
    enabled: boolean
    rules: BonusRule[]
  }

  interface NumericSettingItem {
    key: NumericSettingKey
    label: string
    placeholder: string
    desc: string
  }

  type NumericSettingKey =
    | 'agent_grade_change_fee'
    | 'api_key_free_balance'
    | 'api_key_open_fee'
    | 'checkin_max_reward'
    | 'checkin_max_users'
    | 'checkin_min_balance'
    | 'checkin_min_reward'
    | 'djfl'
    | 'mall_open_price'
    | 'user_ktmoney'
    | 'zdpay'

  type SwitchKey =
    | 'bz'
    | 'checkin_enabled'
    | 'checkin_order_required'
    | 'user_yqzc'
    | 'flkg'
    | 'login_email_verify'
    | 'login_forget_pwd'
    | 'login_slider_verify'
    | 'non_direct_recharge_enable'
    | 'top_consumers_open'
    | 'onlineStore_trdltz'
    | 'pass2_kg'
    | 'qd_notice_open'
    | 'sjqykg'
    | 'sykg'
    | 'user_htkh'
    | 'xdsmopen'

  const siteStore = useSiteStore()

  const loading = ref(false)
  const saving = ref(false)
  const activeTab = ref('wzpz')
  const rawConfig = ref<LegacyAdminConfigMap>({})

  const basicForm = reactive({
    sitename: '',
    version: '',
    keywords: '',
    description: '',
    logo: '',
    hlogo: '',
    notice: '',
    tcgonggao: '',
    qd_notice: '',
    fllx: '1',
    mall_domain_suffix: '',
    cross_recharge_uids: ''
  })

  const switchForm = reactive<Record<SwitchKey, string>>({
    bz: '0',
    checkin_enabled: '0',
    checkin_order_required: '1',
    xdsmopen: '1',
    flkg: '1',
    qd_notice_open: '0',
    top_consumers_open: '0',
    sjqykg: '0',
    pass2_kg: '0',
    onlineStore_trdltz: '0',
    non_direct_recharge_enable: '0',
    user_yqzc: '0',
    user_htkh: '1',
    login_slider_verify: '0',
    login_email_verify: '0',
    login_forget_pwd: '0',
    sykg: '0'
  })

  const businessForm = reactive<Record<NumericSettingKey, string>>({
    mall_open_price: '',
    checkin_min_balance: '',
    checkin_max_users: '',
    checkin_min_reward: '',
    checkin_max_reward: '',
    user_ktmoney: '',
    agent_grade_change_fee: '',
    api_key_open_fee: '',
    api_key_free_balance: '',
    djfl: '',
    zdpay: ''
  })

  const agentPolicyForm = reactive({
    dl_pkkg: '0'
  })

  const recommendChannels = ref<RecommendChannelItem[]>([])

  const createEmptyBonusRule = (): BonusRule => ({
    min: 0,
    max: 0,
    bonus_pct: 0
  })

  const createDefaultBonusConfig = (): BonusConfig => ({
    enabled: false,
    rules: [],
    activity: {
      enabled: false,
      hint: '',
      weekdays: [],
      rules: []
    }
  })

  const bonusForm = reactive<BonusConfig>(createDefaultBonusConfig())

  const switchItems: Array<{ key: SwitchKey; label: string; desc: string }> = [
    { key: 'xdsmopen', label: '扫码下单', desc: '控制前台是否开启扫码下单模式。' },
    { key: 'flkg', label: '分类开关', desc: '控制前台课程分类面板是否展示。' },
    { key: 'qd_notice_open', label: '签到公告', desc: '控制签到公告区域是否启用。' },
    { key: 'sjqykg', label: '上级迁移', desc: '控制用户中心的上级迁移入口。' },
    { key: 'pass2_kg', label: '二级密码', desc: '控制管理员登录二级密码校验。' },
    { key: 'sykg', label: '全站水印', desc: '开启后后台页面显示防截图水印。' },
    { key: 'login_slider_verify', label: '滑块验证', desc: '登录时启用滑块校验。' },
    { key: 'login_email_verify', label: '邮箱验证', desc: '登录或注册链路启用邮箱校验。' },
    { key: 'login_forget_pwd', label: '忘记密码', desc: '控制登录页是否显示找回密码入口。' }
  ]

  const businessSwitchItems: Array<{ key: SwitchKey; label: string; desc: string }> = [
    {
      key: 'onlineStore_trdltz',
      label: '代登入入口',
      desc: '控制用户管理页是否显示管理员代登入按钮。'
    },
    {
      key: 'top_consumers_open',
      label: '消费排行榜',
      desc: '控制控制台右侧是否显示用户消费排行。'
    }
  ]

  const checkinSwitchItems: Array<{ key: 'checkin_enabled' | 'checkin_order_required'; label: string; desc: string }> = [
    {
      key: 'checkin_enabled',
      label: '开启签到',
      desc: '关闭后前台签到功能不可用。'
    },
    {
      key: 'checkin_order_required',
      label: '要求历史订单',
      desc: '开启后仅有历史订单的用户可参与签到。'
    }
  ]

  const paymentSwitchItems: Array<{ key: 'non_direct_recharge_enable'; label: string; desc: string }> = [
    {
      key: 'non_direct_recharge_enable',
      label: '非直系在线充值',
      desc: '关闭后，非直系代理只能按页面提示联系上级充值。'
    }
  ]

  const agentSwitchItems: Array<{ key: 'user_yqzc' | 'user_htkh'; label: string; desc: string; defaultValue: string }> = [
    {
      key: 'user_yqzc',
      label: '邀请注册',
      desc: '控制前台是否要求邀请码注册。',
      defaultValue: '0'
    },
    {
      key: 'user_htkh',
      label: '允许代理开户',
      desc: '关闭后，代理列表中的新增代理链路会被后端直接拦截。',
      defaultValue: '1'
    }
  ]

  const agentPolicyItems = ['dl_pkkg', 'djfl'] as const

  const agentPolicyOptions = [
    { value: '0', label: '无限制（正常开启）' },
    { value: '1', label: '禁止顶级用户平开' },
    { value: '2', label: '顶级平开需双倍余额' },
    { value: '3', label: '同级平开需双倍余额' }
  ]

  const businessNumericItems: NumericSettingItem[] = [
    {
      key: 'mall_open_price',
      label: '商城开通价格',
      placeholder: '例如：99',
      desc: '租户开通商城时展示并使用该价格。'
    }
  ]

  const checkinNumericItems: NumericSettingItem[] = [
    {
      key: 'checkin_min_balance',
      label: '签到最低余额',
      placeholder: '例如：10',
      desc: '用户余额低于该值时无法签到。'
    },
    {
      key: 'checkin_max_users',
      label: '每日签到上限',
      placeholder: '例如：10',
      desc: '限制每天可成功签到的总人数。'
    },
    {
      key: 'checkin_min_reward',
      label: '最小奖励金额',
      placeholder: '例如：0.1',
      desc: '签到随机奖励的最小值。'
    },
    {
      key: 'checkin_max_reward',
      label: '最大奖励金额',
      placeholder: '例如：0.2',
      desc: '签到随机奖励的最大值。'
    }
  ]

  const paymentNumericItems: NumericSettingItem[] = [
    {
      key: 'zdpay',
      label: '最低在线充值金额',
      placeholder: '例如：10',
      desc: '用户在线支付时，低于该金额会被后端直接拒绝。'
    }
  ]

  const agentNumericItems: NumericSettingItem[] = [
    {
      key: 'user_ktmoney',
      label: '代理开户费',
      placeholder: '留空则沿用旧默认值',
      desc: '填写正数后覆盖后端默认开户费；留空时走旧默认值。'
    },
    {
      key: 'agent_grade_change_fee',
      label: '修改等级手续费',
      placeholder: '默认 3',
      desc: '代理修改下级等级或费率时收取，填 0 可免收。'
    },
    {
      key: 'api_key_open_fee',
      label: '接口密钥开通费',
      placeholder: '默认 5',
      desc: '代理给下级开通接口、用户自助开通接口时共用。'
    },
    {
      key: 'api_key_free_balance',
      label: '自助开通免扣余额',
      placeholder: '默认 100',
      desc: '用户自己开通接口时，余额达到该值免扣开通费；填 0 关闭免扣。'
    }
  ]

  const numericSettingItems: NumericSettingItem[] = [
    ...businessNumericItems,
    ...checkinNumericItems,
    ...paymentNumericItems,
    ...agentNumericItems
  ]

  const weekdayOptions = [
    { label: '周日', value: 0 },
    { label: '周一', value: 1 },
    { label: '周二', value: 2 },
    { label: '周三', value: 3 },
    { label: '周四', value: 4 },
    { label: '周五', value: 5 },
    { label: '周六', value: 6 }
  ]

  const safeJsonParse = <T,>(value: string, fallback: T): T => {
    try {
      return JSON.parse(value) as T
    } catch {
      return fallback
    }
  }

  const normalizeDomain = (raw?: string) => {
    return String(raw || '')
      .trim()
      .toLowerCase()
      .replace(/^https?:\/\//, '')
      .replace(/\/.*$/, '')
      .replace(/:\d+$/, '')
  }

  const createEmptyRecommendChannel = (): RecommendChannelItem => ({
    name: '',
    desc: ''
  })

  const normalizeBonusRule = (rule?: Partial<BonusRule>): BonusRule => ({
    min: Number(rule?.min || 0),
    max: Number(rule?.max || 0),
    bonus_pct: Number(rule?.bonus_pct || 0)
  })

  const normalizeRecommendChannels = (items: RecommendChannelItem[]) => {
    return items
      .map((item) => ({
        name: String(item.name || '').trim(),
        desc: String(item.desc || '').trim()
      }))
      .filter((item) => item.name)
  }

  const parseRecommendChannels = (value?: string) => {
    const parsed = safeJsonParse<RecommendChannelItem[]>(value || '[]', [])
    return Array.isArray(parsed)
      ? parsed.map((item) => ({
          name: String(item?.name || ''),
          desc: String(item?.desc || '')
        }))
      : []
  }

  const resetBonusForm = (value?: Partial<BonusConfig>) => {
    const fallback = createDefaultBonusConfig()
    bonusForm.enabled = Boolean(value?.enabled)
    bonusForm.rules = Array.isArray(value?.rules)
      ? value.rules.map((rule) => normalizeBonusRule(rule))
      : fallback.rules
    bonusForm.activity = {
      enabled: Boolean(value?.activity?.enabled),
      hint: String(value?.activity?.hint || ''),
      weekdays: Array.isArray(value?.activity?.weekdays)
        ? value.activity.weekdays
            .map((day) => Number(day))
            .filter((day) => !Number.isNaN(day) && day >= 0 && day <= 6)
            .sort((a, b) => a - b)
        : fallback.activity.weekdays,
      rules: Array.isArray(value?.activity?.rules)
        ? value.activity.rules.map((rule) => normalizeBonusRule(rule))
        : fallback.activity.rules
    }
  }

  const serializeRecommendChannels = () => JSON.stringify(normalizeRecommendChannels(recommendChannels.value))

  const serializeBonusConfig = () => {
    const payload: BonusConfig = {
      enabled: bonusForm.enabled,
      rules: bonusForm.rules.map((rule) => normalizeBonusRule(rule)),
      activity: {
        enabled: bonusForm.activity.enabled,
        hint: String(bonusForm.activity.hint || '').trim(),
        weekdays: [...bonusForm.activity.weekdays].sort((a, b) => a - b),
        rules: bonusForm.activity.rules.map((rule) => normalizeBonusRule(rule))
      }
    }
    return JSON.stringify(payload)
  }

  const normalizedRecommendChannels = computed(() => normalizeRecommendChannels(recommendChannels.value))
  const recommendChannelCount = computed(() => normalizedRecommendChannels.value.length)
  const totalBonusRuleCount = computed(() => bonusForm.rules.length + bonusForm.activity.rules.length)
  const activeWeekdayLabels = computed(() => {
    return weekdayOptions
      .filter((item) => bonusForm.activity.weekdays.includes(item.value))
      .map((item) => item.label)
      .join('、')
  })

  const updateSwitch = (key: SwitchKey, value: string | number | boolean) => {
    switchForm[key] = value ? '1' : '0'
  }

  const addRecommendChannel = () => {
    recommendChannels.value.push(createEmptyRecommendChannel())
  }

  const removeRecommendChannel = (index: number) => {
    recommendChannels.value.splice(index, 1)
  }

  const addBonusRule = (target: 'rules' | 'activity') => {
    if (target === 'rules') {
      bonusForm.rules.push(createEmptyBonusRule())
      return
    }
    bonusForm.activity.rules.push(createEmptyBonusRule())
  }

  const removeBonusRule = (target: 'rules' | 'activity', index: number) => {
    if (target === 'rules') {
      bonusForm.rules.splice(index, 1)
      return
    }
    bonusForm.activity.rules.splice(index, 1)
  }

  const toggleWeekday = (day: number) => {
    if (bonusForm.activity.weekdays.includes(day)) {
      bonusForm.activity.weekdays = bonusForm.activity.weekdays.filter((item) => item !== day)
      return
    }
    bonusForm.activity.weekdays = [...bonusForm.activity.weekdays, day].sort((a, b) => a - b)
  }

  const loadConfig = async () => {
    loading.value = true
    try {
      const config = await fetchLegacyAdminConfig()
      rawConfig.value = { ...config }

      basicForm.sitename = config.sitename || ''
      basicForm.version = config.version || ''
      basicForm.keywords = config.keywords || ''
      basicForm.description = config.description || ''
      basicForm.logo = config.logo || ''
      basicForm.hlogo = config.hlogo || ''
      basicForm.notice = config.notice || ''
      basicForm.tcgonggao = config.tcgonggao || ''
      basicForm.qd_notice = config.qd_notice || ''
      basicForm.fllx = config.fllx || '1'
      basicForm.mall_domain_suffix = normalizeDomain(config.mall_domain_suffix)
      basicForm.cross_recharge_uids = String(config.cross_recharge_uids || '')
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean)
        .join(',')

      switchItems.forEach((item) => {
        switchForm[item.key] = config[item.key] || '0'
      })
      switchForm.bz = config.bz || '0'
      businessSwitchItems.forEach((item) => {
        switchForm[item.key] = config[item.key] || '0'
      })
      switchForm.top_consumers_open = config.top_consumers_open || '0'
      checkinSwitchItems.forEach((item) => {
        switchForm[item.key] = config[item.key] || (item.key === 'checkin_order_required' ? '1' : '0')
      })
      paymentSwitchItems.forEach((item) => {
        switchForm[item.key] = config[item.key] || '0'
      })
      agentSwitchItems.forEach((item) => {
        switchForm[item.key] = config[item.key] || item.defaultValue
      })
      numericSettingItems.forEach((item) => {
        businessForm[item.key] = config[item.key] || ''
      })
      agentPolicyForm.dl_pkkg = config.dl_pkkg || '0'
      if (!config.xdsmopen) {
        switchForm.xdsmopen = '1'
      }
      if (!config.flkg) {
        switchForm.flkg = '1'
      }

      recommendChannels.value = parseRecommendChannels(config.recommend_channels)
      resetBonusForm(safeJsonParse<BonusConfig>(config.recharge_bonus_rules || '{}', createDefaultBonusConfig()))
    } finally {
      loading.value = false
    }
  }

  const handleSave = async () => {
    if (!basicForm.sitename.trim()) {
      ElMessage.warning('请先填写站点名称')
      return
    }

    basicForm.mall_domain_suffix = normalizeDomain(basicForm.mall_domain_suffix)
    basicForm.cross_recharge_uids = basicForm.cross_recharge_uids
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
      .join(',')

    const payload: LegacyAdminConfigMap = {
      ...rawConfig.value,
      sitename: basicForm.sitename.trim(),
      version: basicForm.version.trim(),
      keywords: basicForm.keywords.trim(),
      description: basicForm.description.trim(),
      logo: basicForm.logo.trim(),
      hlogo: basicForm.hlogo.trim(),
      notice: basicForm.notice.trim(),
      tcgonggao: basicForm.tcgonggao.trim(),
      qd_notice: basicForm.qd_notice.trim(),
      fllx: basicForm.fllx,
      mall_domain_suffix: basicForm.mall_domain_suffix,
      cross_recharge_uids: basicForm.cross_recharge_uids,
      recommend_channels: serializeRecommendChannels(),
      recharge_bonus_rules: serializeBonusConfig(),
      dl_pkkg: agentPolicyForm.dl_pkkg,
      ...switchForm,
      ...businessForm
    }

    saving.value = true
    try {
      await saveLegacyAdminConfig(payload)
      ElMessage.success('系统配置已保存')
      await siteStore.initPublicConfig(true)
      await loadConfig()
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadConfig()
  })
</script>
