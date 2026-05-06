<template>
  <div class="tenant-shop-page art-full-height">
    <div v-loading="loading">
      <section v-if="!opened" class="art-card-sm p-6">
        <article class="mx-auto w-full max-w-[760px] rounded-custom-sm border-full-d bg-g-100/60 p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-sm font-medium text-g-900">开通商城</p>
              <p class="mt-1 text-sm leading-6 text-g-500">
                开通后即可继续维护店铺名称、公告、会员注册、返利和客服入口。
              </p>
            </div>
            <div class="flex flex-wrap items-center gap-3">
              <ElTag effect="plain">店铺设置</ElTag>
              <ElTag type="warning" effect="plain">商城未开通</ElTag>
              <ElTag type="warning" effect="plain">开通价格 {{ moneyLabel(openPrice) }}</ElTag>
              <ElButton plain :loading="loading" @click="loadData">刷新</ElButton>
            </div>
          </div>

          <div class="mt-5 flex flex-col gap-3 sm:flex-row">
            <ElInput
              v-model="openShopName"
              class="flex-1"
              maxlength="60"
              placeholder="请输入准备开通的店铺名称"
            />
            <ElButton type="primary" :loading="opening" @click="handleOpen">立即开通商城</ElButton>
          </div>

          <ElAlert
            class="mt-4"
            title="如平台已配置商城主域名，开通后可直接设置店铺子域名前缀。"
            type="info"
            :closable="false"
            show-icon
          />
        </article>
      </section>

      <div v-else class="grid gap-4 xl:grid-cols-[1.02fr_0.98fr]">
        <section class="art-card-sm p-5">
          <div class="border-b-d pb-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 class="text-lg font-semibold text-g-900">基础资料</h3>
                <p class="mt-1.5 text-sm leading-6 text-g-500">维护店铺基础信息、访问地址和开关状态。</p>
              </div>
              <div class="flex flex-wrap items-center gap-3">
                <ElTag effect="plain">店铺设置</ElTag>
                <ElTag type="success" effect="plain">商城已开通</ElTag>
                <ElTag effect="plain">店铺 ID {{ form.tid || 0 }}</ElTag>
                <ElTag v-if="mallDomainEnabled" type="info" effect="plain">已启用主域名</ElTag>
                <ElButton plain :loading="loading" @click="loadData">刷新</ElButton>
                <ElButton type="primary" plain :loading="saving" @click="handleSave">保存设置</ElButton>
              </div>
            </div>
          </div>

          <div class="mt-5 space-y-4">
            <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">商城访问地址</p>
              <div class="mt-3 flex flex-wrap items-center gap-3">
                <ElLink v-if="mallUrl" :href="mallUrl" type="primary" target="_blank">{{ mallUrl }}</ElLink>
                <span v-else class="text-sm text-g-500">尚未生成访问地址</span>
                <ElButton v-if="mallUrl" size="small" plain @click="copyMallUrl">复制链接</ElButton>
                <ElButton v-if="mallUrl" size="small" type="primary" plain @click="openMallUrl">打开商城</ElButton>
              </div>
              <p class="mt-2 text-sm leading-6 text-g-500">
                {{ form.domain ? '当前已启用域名直达，用户无需再带店铺 ID 访问。' : '当前仍使用默认商城路径访问。' }}
              </p>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">店铺名称</label>
              <ElInput v-model="form.shop_name" maxlength="60" placeholder="请输入店铺名称" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">Logo 地址</label>
              <ElInput v-model="form.shop_logo" placeholder="https://example.com/logo.png" />
              <div v-if="form.shop_logo" class="mt-3">
                <img :src="form.shop_logo" alt="logo" class="h-20 w-20 rounded-custom-sm border-full-d object-cover" />
              </div>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">跑马灯公告</label>
              <ElInput
                v-model="form.shop_desc"
                type="textarea"
                :rows="4"
                placeholder="展示在商城首页滚动公告区域的内容"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">商城域名</label>
              <template v-if="mallDomainEnabled">
                <ElInput
                  v-model="subdomainPrefix"
                  placeholder="请输入店铺子域名前缀"
                  @blur="subdomainPrefix = normalizeSubdomainPrefix(subdomainPrefix)"
                >
                  <template #append>.{{ mallDomainSuffix }}</template>
                </ElInput>
                <p class="mt-2 text-sm leading-6 text-g-500">
                  平台已配置商城主域名 `{{ mallDomainSuffix }}`，这里只需要填写子域名前缀。
                </p>
              </template>
              <ElAlert
                v-else
                title="主站尚未配置商城主域名"
                type="warning"
                :closable="false"
                show-icon
              />
            </div>

            <div class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">店铺状态</p>
                  <p class="mt-1 text-sm text-g-500">关闭后商城仍存在，但前台将不再对外销售。</p>
                </div>
                <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
              </div>
            </div>
          </div>
        </section>

        <section class="art-card-sm p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">商城配置</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">集中维护会员注册、返利、公告和客服配置。</p>
          </div>

          <div class="mt-5 space-y-4">
            <article class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">显示分类栏</p>
                  <p class="mt-1 text-sm text-g-500">关闭后首页不显示分类切换栏，商品仍正常展示。</p>
                </div>
                <ElSwitch v-model="mallConfig.show_categories" />
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">开放会员注册</p>
                  <p class="mt-1 text-sm text-g-500">访客可在商城自助注册会员账号。</p>
                </div>
                <ElSwitch v-model="mallConfig.register_enabled" />
              </div>
            </article>

            <article class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">推广返利</p>
                  <p class="mt-1 text-sm text-g-500">会员带客下单后按支付金额返利。</p>
                </div>
                <ElSwitch v-model="mallConfig.promotion_enabled" />
              </div>
              <div class="mt-4 max-w-[220px]">
                <label class="mb-2 block text-sm font-medium text-g-800">返利比例</label>
                <ElInputNumber v-model="mallConfig.commission_rate" class="w-full" :min="0" :max="100" :precision="2" />
              </div>
            </article>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">弹窗公告（支持 HTML）</label>
              <ElInput
                v-model="mallConfig.popup_notice_html"
                type="textarea"
                :rows="6"
                placeholder="<p>输入商城首页弹窗公告内容</p>"
              />
            </div>

            <article class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">客服入口</p>
                  <p class="mt-1 text-sm text-g-500">首页、查进度页和我的页面都会复用这里的配置。</p>
                </div>
                <ElSwitch v-model="mallConfig.customer_service.enabled" />
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">客服方式</label>
                  <ElSelect v-model="mallConfig.customer_service.type" class="w-full">
                    <ElOption label="微信" value="wechat" />
                    <ElOption label="QQ" value="qq" />
                    <ElOption label="电话" value="phone" />
                    <ElOption label="链接" value="link" />
                  </ElSelect>
                </div>
                <div>
                  <label class="mb-2 block text-sm font-medium text-g-800">按钮文案</label>
                  <ElInput v-model="mallConfig.customer_service.label" placeholder="例如：联系客服" />
                </div>
              </div>

              <div class="mt-4">
                <label class="mb-2 block text-sm font-medium text-g-800">客服内容</label>
                <ElInput v-model="mallConfig.customer_service.value" placeholder="填写微信号、QQ号、手机号或链接" />
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import {
    fetchTenantMallOpenPrice,
    fetchTenantShop,
    openTenantMall,
    saveTenantMallConfig,
    saveTenantShop,
    type LegacyTenantInfo
  } from '@/api/legacy/tenant'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'TenantShopPage' })

  interface TenantMallConfig {
    commission_rate: number
    customer_service: {
      enabled: boolean
      label: string
      type: string
      value: string
    }
    popup_notice_html: string
    promotion_enabled: boolean
    register_enabled: boolean
    show_categories: boolean
  }

  const loading = ref(false)
  const saving = ref(false)
  const opening = ref(false)
  const opened = ref(false)
  const openPrice = ref(99)
  const openShopName = ref('')
  const mallDomainSuffix = ref('')
  const subdomainPrefix = ref('')

  const form = reactive<Partial<LegacyTenantInfo>>({
    tid: 0,
    shop_name: '',
    shop_logo: '',
    shop_desc: '',
    status: 1,
    domain: ''
  })

  const createDefaultMallConfig = (): TenantMallConfig => ({
    commission_rate: 0,
    customer_service: {
      enabled: false,
      label: '联系客服',
      type: 'wechat',
      value: ''
    },
    popup_notice_html: '',
    promotion_enabled: false,
    register_enabled: false,
    show_categories: true
  })

  const mallConfig = reactive<TenantMallConfig>(createDefaultMallConfig())

  const mallDomainEnabled = computed(() => Boolean(mallDomainSuffix.value))
  const effectiveDomain = computed(() => {
    if (!mallDomainEnabled.value) return ''
    const prefix = normalizeSubdomainPrefix(subdomainPrefix.value)
    return prefix ? `${prefix}.${mallDomainSuffix.value}` : ''
  })

  const mallUrl = computed(() => {
    const domain = effectiveDomain.value
    if (domain) return `https://${domain}/mall/`
    if (form.tid) return `${window.location.origin}/mall/${form.tid}/`
    return ''
  })

  const moneyLabel = (value?: number | string) => `¥${Number(value || 0).toFixed(2)}`

  function normalizeDomain(raw?: string) {
    return String(raw || '')
      .trim()
      .toLowerCase()
      .replace(/^https?:\/\//, '')
      .replace(/\/.*$/, '')
      .replace(/:\d+$/, '')
  }

  function normalizeSubdomainPrefix(raw?: string) {
    return String(raw || '')
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9-.]/g, '')
      .replace(/\.+/g, '.')
      .replace(/^\.+|\.+$/g, '')
  }

  function extractPrefixFromDomain(domain: string, suffix: string) {
    const normalizedDomain = normalizeDomain(domain)
    const normalizedSuffix = normalizeDomain(suffix)
    if (!normalizedDomain || !normalizedSuffix) return ''
    if (normalizedDomain === normalizedSuffix) return ''
    const suffixToken = `.${normalizedSuffix}`
    if (!normalizedDomain.endsWith(suffixToken)) return ''
    return normalizedDomain.slice(0, -suffixToken.length)
  }

  function resetMallConfig() {
    Object.assign(mallConfig, createDefaultMallConfig())
  }

  async function loadOpenPrice() {
    const result = await fetchTenantMallOpenPrice()
    if (typeof result === 'number') {
      openPrice.value = result
      return
    }
    openPrice.value = Number(result?.price || 99)
  }

  async function loadData() {
    loading.value = true
    try {
      const result = await fetchTenantShop()
      opened.value = true
      Object.assign(form, {
        tid: 0,
        shop_name: '',
        shop_logo: '',
        shop_desc: '',
        status: 1,
        domain: ''
      }, result || {})

      mallDomainSuffix.value = normalizeDomain(result?.mall_domain_suffix)
      form.domain = normalizeDomain(result?.domain)
      subdomainPrefix.value = extractPrefixFromDomain(String(form.domain || ''), mallDomainSuffix.value)

      resetMallConfig()
      if (result?.mall_config) {
        try {
          const parsed = JSON.parse(result.mall_config)
          Object.assign(mallConfig, createDefaultMallConfig(), {
            commission_rate: Number(parsed?.commission_rate || 0),
            customer_service: {
              enabled: Boolean(parsed?.customer_service?.enabled),
              label: String(parsed?.customer_service?.label || '联系客服'),
              type: String(parsed?.customer_service?.type || 'wechat'),
              value: String(parsed?.customer_service?.value || '')
            },
            popup_notice_html: String(parsed?.popup_notice_html || ''),
            promotion_enabled: Boolean(parsed?.promotion_enabled),
            register_enabled: Boolean(parsed?.register_enabled),
            show_categories: parsed?.show_categories !== false
          })
        } catch {
          resetMallConfig()
        }
      }
    } catch {
      opened.value = false
      resetMallConfig()
      await loadOpenPrice()
    } finally {
      loading.value = false
    }
  }

  async function handleOpen() {
    if (!openShopName.value.trim()) {
      ElMessage.warning('请先填写店铺名称')
      return
    }

    try {
      await ElMessageBox.confirm(
        `将从余额扣除 ${moneyLabel(openPrice.value)} 开通商城，确认继续？`,
        '开通商城',
        {
          type: 'warning',
          confirmButtonText: '确认开通',
          cancelButtonText: '取消'
        }
      )
    } catch {
      return
    }

    opening.value = true
    try {
      await openTenantMall({ shop_name: openShopName.value.trim() })
      ElMessage.success('商城开通成功')
      await loadData()
    } finally {
      opening.value = false
    }
  }

  async function handleSave() {
    if (!String(form.shop_name || '').trim()) {
      ElMessage.warning('请先填写店铺名称')
      return
    }

    if (mallDomainEnabled.value) {
      subdomainPrefix.value = normalizeSubdomainPrefix(subdomainPrefix.value)
      if (!subdomainPrefix.value) {
        ElMessage.warning('请先填写子域名前缀')
        return
      }
      form.domain = `${subdomainPrefix.value}.${mallDomainSuffix.value}`
    } else {
      form.domain = ''
    }

    saving.value = true
    try {
      await saveTenantShop({ ...form, shop_name: String(form.shop_name || '').trim() })
      await saveTenantMallConfig(
        JSON.stringify({
          commission_rate: Number(mallConfig.commission_rate || 0),
          customer_service: mallConfig.customer_service,
          popup_notice_html: mallConfig.popup_notice_html,
          promotion_enabled: mallConfig.promotion_enabled,
          register_enabled: mallConfig.register_enabled,
          show_categories: mallConfig.show_categories
        })
      )
      ElMessage.success('店铺设置已保存')
      await loadData()
    } finally {
      saving.value = false
    }
  }

  async function copyMallUrl() {
    if (!mallUrl.value) return
    if (!navigator?.clipboard?.writeText) {
      ElMessage.warning('当前环境不支持自动复制，请手动复制链接')
      return
    }
    await navigator.clipboard.writeText(mallUrl.value)
    ElMessage.success('商城地址已复制')
  }

  function openMallUrl() {
    if (!mallUrl.value) return
    window.open(mallUrl.value, '_blank', 'noopener,noreferrer')
  }

  onMounted(() => {
    loadData()
  })
</script>
