<template>
  <div class="tenant-pay-config-page art-full-height">
    <section v-if="!available" class="art-card-sm p-5">
      <ElAlert
        title="当前租户尚未开通商城，暂时无法维护支付参数。"
        type="warning"
        :closable="false"
        show-icon
      />
      <div class="mt-4">
        <RouterLink class="text-sm text-[var(--el-color-primary)]" to="/tenant/shop">
          前往店铺设置页开通商城
        </RouterLink>
      </div>
    </section>

    <div v-else class="grid gap-4 xl:grid-cols-[0.96fr_1.04fr]">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 class="text-lg font-semibold text-g-900">收款模式</h3>
              <p class="mt-1.5 text-sm leading-6 text-g-500">先确定资金链路，再决定是否需要自配易支付接口。</p>
            </div>
            <div class="flex flex-wrap items-center gap-3">
              <ElTag effect="plain">支付配置</ElTag>
              <ElTag type="success" effect="plain">商城已开通</ElTag>
              <ElTag effect="plain">
                当前模式 {{ isSiteBalanceMode ? '站长代收' : '商家直连' }}
              </ElTag>
              <ElButton plain :loading="loading" @click="loadData">刷新</ElButton>
              <ElButton type="primary" plain :loading="saving" @click="handleSave">保存配置</ElButton>
            </div>
          </div>
        </div>

        <div class="mt-5 space-y-4">
          <ElRadioGroup v-model="form.pay_mode" class="w-full">
            <div class="grid gap-4">
              <label class="rounded-custom-sm border-full-d p-4">
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <p class="text-sm font-medium text-g-900">站长代收</p>
                    <p class="mt-2 text-sm leading-6 text-g-500">
                      C 端支付成功后，资金先进入主站余额，再自动扣供货价下单，利润保留在余额中。
                    </p>
                  </div>
                  <ElRadioButton value="site_balance">站长代收</ElRadioButton>
                </div>
              </label>

              <label class="rounded-custom-sm border-full-d p-4">
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <p class="text-sm font-medium text-g-900">商家直连</p>
                    <p class="mt-2 text-sm leading-6 text-g-500">
                      用户购买商品时直接跳转到你配置的易支付接口收款，资金不经过主站余额。
                    </p>
                  </div>
                  <ElRadioButton value="merchant_epay">商家直连</ElRadioButton>
                </div>
              </label>
            </div>
          </ElRadioGroup>

          <ElAlert
            :title="modeDescription"
            :type="isSiteBalanceMode ? 'success' : 'info'"
            :closable="false"
            show-icon
          />
        </div>
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">支付参数</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">
            商家直连模式下填写易支付参数；站长代收模式下这里只保留说明，不再暴露无效字段。
          </p>
        </div>

        <div class="mt-5">
          <div v-if="isSiteBalanceMode" class="space-y-4">
            <ElAlert
              title="当前为站长代收模式"
              type="success"
              :closable="false"
              show-icon
            />

            <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-sm font-semibold text-g-900">说明</p>
              <ul class="mt-3 space-y-2 text-sm leading-6 text-g-500">
                <li>商城展示的支付方式会直接读取平台管理员在系统设置中的支付开关。</li>
                <li>商家不需要单独维护易支付参数，也不需要担心回调地址和签名问题。</li>
                <li>后续如切换为商家直连，再回到此页补齐参数即可。</li>
              </ul>
            </div>
          </div>

          <div v-else class="space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">易支付 API 地址</label>
              <ElInput v-model="form.epay_api" placeholder="https://your-epay.com/" />
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">商户 PID</label>
                <ElInput v-model="form.epay_pid" placeholder="请输入商户 PID" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">商户密钥</label>
                <ElInput v-model="form.epay_key" type="password" show-password placeholder="请输入商户密钥" />
              </div>
            </div>

            <div class="rounded-custom-sm border-full-d p-4">
              <p class="text-sm font-semibold text-g-900">对外展示支付方式</p>
              <p class="mt-2 text-sm leading-6 text-g-500">至少开启一种支付渠道，否则前台无法完成下单支付。</p>

              <div class="mt-4 grid gap-4 md:grid-cols-3">
                <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-900">支付宝</span>
                    <ElSwitch
                      :model-value="form.is_alipay === '1'"
                      @change="(value) => (form.is_alipay = value ? '1' : '0')"
                    />
                  </div>
                </div>
                <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-900">微信支付</span>
                    <ElSwitch
                      :model-value="form.is_wxpay === '1'"
                      @change="(value) => (form.is_wxpay = value ? '1' : '0')"
                    />
                  </div>
                </div>
                <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium text-g-900">QQ 支付</span>
                    <ElSwitch
                      :model-value="form.is_qqpay === '1'"
                      @change="(value) => (form.is_qqpay = value ? '1' : '0')"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { fetchTenantShop, saveTenantPayConfig } from '@/api/legacy/tenant'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'TenantPayConfigPage' })

  type PayMode = 'merchant_epay' | 'site_balance'

  interface PayConfig {
    epay_api: string
    epay_key: string
    epay_pid: string
    is_alipay: string
    is_qqpay: string
    is_wxpay: string
    pay_mode: PayMode
  }

  const loading = ref(false)
  const saving = ref(false)
  const available = ref(false)

  const form = reactive<PayConfig>({
    epay_api: '',
    epay_key: '',
    epay_pid: '',
    is_alipay: '1',
    is_qqpay: '0',
    is_wxpay: '0',
    pay_mode: 'site_balance'
  })

  const isSiteBalanceMode = computed(() => form.pay_mode === 'site_balance')

  const modeDescription = computed(() =>
    isSiteBalanceMode.value
      ? '默认使用站长代收，商城支付金额会先进入主站余额，再自动扣供货价。'
      : '当前为商家直连模式，C 端购买商品时会跳转到你的易支付接口完成付款。'
  )

  function createDefaultForm(): PayConfig {
    return {
      epay_api: '',
      epay_key: '',
      epay_pid: '',
      is_alipay: '1',
      is_qqpay: '0',
      is_wxpay: '0',
      pay_mode: 'site_balance'
    }
  }

  function resolvePayMode(config: Record<string, any>): PayMode {
    const mode = String(config.pay_mode || config.mode || '').trim()
    if (mode === 'merchant_epay' || mode === 'site_balance') {
      return mode
    }
    if (config.epay_api || config.epay_pid || config.epay_key) {
      return 'merchant_epay'
    }
    return 'site_balance'
  }

  async function loadData() {
    loading.value = true
    try {
      const result = await fetchTenantShop()
      available.value = true
      Object.assign(form, createDefaultForm())
      if (result?.pay_config) {
        try {
          const parsed = JSON.parse(result.pay_config)
          Object.assign(form, {
            epay_api: String(parsed?.epay_api || ''),
            epay_key: String(parsed?.epay_key || ''),
            epay_pid: String(parsed?.epay_pid || ''),
            is_alipay: parsed?.is_alipay === '1' || parsed?.is_alipay === true ? '1' : '0',
            is_qqpay: parsed?.is_qqpay === '1' || parsed?.is_qqpay === true ? '1' : '0',
            is_wxpay: parsed?.is_wxpay === '1' || parsed?.is_wxpay === true ? '1' : '0',
            pay_mode: resolvePayMode(parsed)
          })
        } catch {
          Object.assign(form, createDefaultForm())
        }
      }
    } catch {
      available.value = false
      Object.assign(form, createDefaultForm())
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    if (!available.value) {
      ElMessage.warning('请先开通商城')
      return
    }

    if (!isSiteBalanceMode.value) {
      if (!form.epay_api.trim()) {
        ElMessage.warning('请先填写易支付 API 地址')
        return
      }
      if (!form.epay_pid.trim()) {
        ElMessage.warning('请先填写商户 PID')
        return
      }
      if (!form.epay_key.trim()) {
        ElMessage.warning('请先填写商户密钥')
        return
      }
      if (form.is_alipay !== '1' && form.is_wxpay !== '1' && form.is_qqpay !== '1') {
        ElMessage.warning('请至少开启一种支付方式')
        return
      }
    }

    saving.value = true
    try {
      const payload = isSiteBalanceMode.value
        ? { pay_mode: 'site_balance' }
        : { ...form }
      await saveTenantPayConfig(JSON.stringify(payload))
      ElMessage.success('支付配置已保存')
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
