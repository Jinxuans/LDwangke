<!-- 注册页面 -->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ $t('register.title') }}</h3>
          <p class="sub-title">{{ $t('register.subTitle') }}</p>
          <ElForm
            class="mt-7.5"
            ref="formRef"
            :model="formData"
            :rules="rules"
            label-position="top"
            :key="formKey"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                v-model.trim="formData.username"
                :placeholder="$t('register.placeholder.username')"
              />
            </ElFormItem>
            <ElFormItem prop="nickname">
              <ElInput
                class="custom-height"
                v-model.trim="formData.nickname"
                :placeholder="$t('register.placeholder.nickname')"
              />
            </ElFormItem>

            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                v-model.trim="formData.password"
                :placeholder="$t('register.placeholder.password')"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <ElFormItem prop="confirmPassword">
              <ElInput
                class="custom-height"
                v-model.trim="formData.confirmPassword"
                :placeholder="$t('register.placeholder.confirmPassword')"
                type="password"
                autocomplete="off"
                @keyup.enter="register"
                show-password
              />
            </ElFormItem>
            <ElFormItem prop="inviteCode">
              <ElInput
                class="custom-height"
                v-model.trim="formData.inviteCode"
                :placeholder="
                  inviteCodeRequired
                    ? $t('register.placeholder.inviteCode')
                    : $t('register.placeholder.inviteCodeOptional')
                "
              />
            </ElFormItem>

            <template v-if="emailVerifyEnabled">
              <ElFormItem prop="email">
                <ElInput
                  class="custom-height"
                  v-model.trim="formData.email"
                  :placeholder="$t('register.placeholder.email')"
                />
              </ElFormItem>

              <ElFormItem prop="verifyCode">
                <ElInput
                  class="custom-height verify-code-input"
                  v-model.trim="formData.verifyCode"
                  :placeholder="$t('register.placeholder.verifyCode')"
                >
                  <template #append>
                    <ElButton
                      native-type="button"
                      :loading="sendingCode"
                      :disabled="countdown > 0"
                      @click="sendRegisterCode"
                    >
                      {{ countdown > 0 ? `${countdown}s` : $t('register.sendCode') }}
                    </ElButton>
                  </template>
                </ElInput>
              </ElFormItem>
            </template>

            <ElFormItem prop="agreement">
              <ElCheckbox v-model="formData.agreement">
                {{ $t('register.agreeText') }}
                <RouterLink
                  style="color: var(--theme-color); text-decoration: none"
                  to="/privacy-policy"
                  >{{ $t('register.privacyPolicy') }}</RouterLink
                >
              </ElCheckbox>
            </ElFormItem>

            <div style="margin-top: 15px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="register"
                :loading="loading"
                v-ripple
              >
                {{ $t('register.submitBtnText') }}
              </ElButton>
            </div>

            <div class="mt-5 text-sm text-g-600">
              <span>{{ $t('register.hasAccount') }}</span>
              <RouterLink class="text-theme" :to="{ name: 'Login' }">{{
                $t('register.toLogin')
              }}</RouterLink>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { fetchRegister, fetchSendRegisterCode } from '@/api/auth'
  import { useSiteStore } from '@/store/modules/site'
  import { useI18n } from 'vue-i18n'
  import type { FormInstance, FormRules } from 'element-plus'

  defineOptions({ name: 'Register' })

  interface RegisterForm {
    username: string
    nickname: string
    password: string
    confirmPassword: string
    inviteCode: string
    email: string
    verifyCode: string
    agreement: boolean
  }

  const PASSWORD_MIN_LENGTH = 6
  const REDIRECT_DELAY = 1000

  const { t, locale } = useI18n()
  const router = useRouter()
  const route = useRoute()
  const siteStore = useSiteStore()
  const formRef = ref<FormInstance>()

  const loading = ref(false)
  const sendingCode = ref(false)
  const countdown = ref(0)
  const formKey = ref(0)
  let countdownTimer: ReturnType<typeof setInterval> | undefined

  const emailVerifyEnabled = computed(() => siteStore.config.login_email_verify === '1')
  const inviteCodeRequired = computed(() => siteStore.config.user_yqzc === '1')

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const formData = reactive<RegisterForm>({
    username: '',
    nickname: '',
    password: '',
    confirmPassword: '',
    inviteCode: typeof route.query.invite === 'string' ? route.query.invite : '',
    email: '',
    verifyCode: '',
    agreement: false
  })

  onMounted(() => {
    siteStore.initPublicConfig()
  })

  onUnmounted(() => {
    if (countdownTimer) {
      clearInterval(countdownTimer)
    }
  })

  /**
   * 验证密码
   * 当密码输入后，如果确认密码已填写，则触发确认密码的验证
   */
  const validatePassword = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (!value) {
      callback(new Error(t('register.placeholder.password')))
      return
    }

    if (formData.confirmPassword) {
      formRef.value?.validateField('confirmPassword')
    }

    callback()
  }

  /**
   * 验证确认密码
   * 检查确认密码是否与密码一致
   */
  const validateConfirmPassword = (
    _rule: any,
    value: string,
    callback: (error?: Error) => void
  ) => {
    if (!value) {
      callback(new Error(t('register.rule.confirmPasswordRequired')))
      return
    }

    if (value !== formData.password) {
      callback(new Error(t('register.rule.passwordMismatch')))
      return
    }

    callback()
  }

  /**
   * 验证用户协议
   * 确保用户已勾选同意协议
   */
  const validateAgreement = (_rule: any, value: boolean, callback: (error?: Error) => void) => {
    if (!value) {
      callback(new Error(t('register.rule.agreementRequired')))
      return
    }
    callback()
  }

  const validateQQAccount = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (!value) {
      callback(new Error(t('register.placeholder.username')))
      return
    }

    if (!/^[1-9][0-9]{4,10}$/.test(value)) {
      callback(new Error(t('register.rule.usernameQQ')))
      return
    }

    callback()
  }

  const rules = computed<FormRules<RegisterForm>>(() => ({
    username: [
      {
        required: true,
        validator: validateQQAccount,
        trigger: 'blur'
      }
    ],
    nickname: [{ required: true, message: t('register.placeholder.nickname'), trigger: 'blur' }],
    password: [
      { required: true, validator: validatePassword, trigger: 'blur' },
      { min: PASSWORD_MIN_LENGTH, message: t('register.rule.passwordLength'), trigger: 'blur' }
    ],
    confirmPassword: [{ required: true, validator: validateConfirmPassword, trigger: 'blur' }],
    inviteCode: inviteCodeRequired.value
      ? [{ required: true, message: t('register.rule.inviteCodeRequired'), trigger: 'blur' }]
      : [],
    email: emailVerifyEnabled.value
      ? [
          { required: true, message: t('register.placeholder.email'), trigger: 'blur' },
          { type: 'email', message: t('register.rule.emailInvalid'), trigger: 'blur' }
        ]
      : [],
    verifyCode: emailVerifyEnabled.value
      ? [{ required: true, message: t('register.placeholder.verifyCode'), trigger: 'blur' }]
      : [],
    agreement: [{ validator: validateAgreement, trigger: 'change' }]
  }))

  const startCountdown = () => {
    countdown.value = 60
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && countdownTimer) {
        clearInterval(countdownTimer)
        countdownTimer = undefined
      }
    }, 1000)
  }

  const sendRegisterCode = async () => {
    if (!formRef.value || sendingCode.value || countdown.value > 0) return

    try {
      await formRef.value.validateField('email')
      sendingCode.value = true
      await fetchSendRegisterCode(formData.email)
      startCountdown()
    } finally {
      sendingCode.value = false
    }
  }

  /**
   * 注册用户
   * 验证表单后提交注册请求
   */
  const register = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      loading.value = true

      await fetchRegister({
        user: formData.username,
        pass: formData.password,
        name: formData.nickname,
        yqm: formData.inviteCode,
        email: emailVerifyEnabled.value ? formData.email : undefined,
        verify_code: emailVerifyEnabled.value ? formData.verifyCode : undefined
      })

      toLogin()
    } catch (error) {
      console.error('表单验证失败:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 跳转到登录页面
   */
  const toLogin = () => {
    setTimeout(() => {
      router.push({ name: 'Login' })
    }, REDIRECT_DELAY)
  }
</script>

<style scoped>
  @import '../login/style.css';
</style>
