<!-- 登录页面 -->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h2 class="title">{{ $t('login.title') }}</h2>
          <p class="sub-title">{{ $t('login.subTitle') }}</p>
          <ElForm
            ref="formRef"
            :model="formData"
            :rules="rules"
            :key="formKey"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.username')"
                v-model.trim="formData.username"
              />
            </ElFormItem>
            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.password')"
                v-model.trim="formData.password"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <!-- 推拽验证 -->
            <div class="relative pb-5 mt-6">
              <div
                class="relative z-[2] overflow-hidden select-none rounded-lg border border-transparent tad-300"
                :class="{ '!border-[var(--el-color-danger)]': !isPassing && isClickPass }"
              >
                <ArtDragVerify
                  ref="dragVerify"
                  v-model:value="isPassing"
                  :text="$t('login.sliderText')"
                  textColor="var(--art-gray-700)"
                  :successText="$t('login.sliderSuccessText')"
                  progressBarBg="var(--main-color)"
                  :background="isDark ? '#26272F' : '#F1F1F4'"
                  handlerBg="var(--default-box-color)"
                />
              </div>
              <p
                class="slider-error absolute top-0 z-[1] px-px mt-2 text-xs tad-300"
                :class="{ 'translate-y-10': !isPassing && isClickPass }"
              >
                {{ $t('login.placeholder.slider') }}
              </p>
            </div>

            <div class="flex-cb mt-2 text-sm">
              <ElCheckbox v-model="formData.staySignedIn">{{
                $t('login.rememberPwd')
              }}</ElCheckbox>
              <RouterLink class="auth-link" :to="{ name: 'ForgetPassword' }">{{
                $t('login.forgetPwd')
              }}</RouterLink>
            </div>

            <div style="margin-top: 30px">
              <ElButton
                class="w-full custom-height login-submit"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                v-ripple
              >
                {{ $t('login.btnText') }}
              </ElButton>
            </div>

            <div class="mt-5 text-sm text-g-700">
              <span>{{ $t('login.noAccount') }}</span>
              <RouterLink class="auth-link auth-inline-link" :to="{ name: 'Register' }">{{
                $t('login.register')
              }}</RouterLink>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useUserStore } from '@/store/modules/user'
  import { useSiteStore } from '@/store/modules/site'
  import { useI18n } from 'vue-i18n'
  import { HttpError } from '@/utils/http/error'
  import { fetchLogin } from '@/api/auth'
  import {
    ElMessage,
    ElMessageBox,
    ElNotification,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useSettingStore } from '@/store/modules/setting'

  defineOptions({ name: 'Login' })

  const settingStore = useSettingStore()
  const { isDark } = storeToRefs(settingStore)
  const { t, locale } = useI18n()
  const formKey = ref(0)

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const dragVerify = ref()

  const userStore = useUserStore()
  const siteStore = useSiteStore()
  const router = useRouter()
  const route = useRoute()
  const isPassing = ref(false)
  const isClickPass = ref(false)

  const { systemName } = storeToRefs(siteStore)
  const formRef = ref<FormInstance>()

  const formData = reactive({
    username: '',
    password: '',
    staySignedIn: true
  })

  const rules = computed<FormRules>(() => ({
    username: [{ required: true, message: t('login.placeholder.username'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.placeholder.password'), trigger: 'blur' }]
  }))

  const loading = ref(false)

  const requestAdminPass2 = async () => {
    try {
      const { value } = await ElMessageBox.prompt(
        '检测到管理员账号，请输入二级密码继续登录',
        '管理员二级密码验证',
        {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          inputType: 'password',
          inputPlaceholder: '请输入二级密码',
          inputValidator: (value) => Boolean(value?.trim()) || '请填写二级密码',
          customClass: 'admin-pass2-dialog'
        }
      )

      return value.trim()
    } catch {
      return ''
    }
  }

  const submitLogin = async (params: Api.Auth.LoginParams): Promise<Api.Auth.LoginResponse> => {
    const result = await fetchLogin(params)

    if (result?.code === 5) {
      const pass2 = await requestAdminPass2()
      if (!pass2) {
        return {}
      }

      return submitLogin({
        ...params,
        pass2
      })
    }

    return result
  }

  // 登录
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      // 表单验证
      const valid = await formRef.value.validate()
      if (!valid) return

      // 拖拽验证
      if (!isPassing.value) {
        isClickPass.value = true
        return
      }

      loading.value = true

      // 登录请求
      const { username, password } = formData

      const result = await submitLogin({
        userName: username,
        password
      })

      if (!result?.accessToken) {
        ElMessage.warning(
          result?.message || result?.error || '登录未完成，请检查账号信息后重试'
        )
        return
      }

      // 存储 token 和登录状态
      userStore.setStaySignedIn(formData.staySignedIn)
      userStore.setToken(result.accessToken, result.refreshToken)
      userStore.setLoginStatus(true)

      // 登录成功处理
      showLoginSuccessNotice()

      // 优先跳转登录前目标页，否则进入默认工作台入口。
      const redirect = route.query.redirect as string
      router.push(redirect || '/dashboard/console')
    } catch (error) {
      // 处理 HttpError
      if (error instanceof HttpError) {
        // console.log(error.code)
      } else {
        // 处理非 HttpError
        // ElMessage.error('登录失败，请稍后重试')
        console.error('[Login] Unexpected error:', error)
      }
    } finally {
      loading.value = false
      resetDragVerify()
    }
  }

  // 重置拖拽验证
  const resetDragVerify = () => {
    dragVerify.value.reset()
  }

  // 登录成功提示
  const showLoginSuccessNotice = () => {
    setTimeout(() => {
      ElNotification({
        title: t('login.success.title'),
        type: 'success',
        duration: 2500,
        zIndex: 10000,
        message: `${t('login.success.message')}, ${systemName.value}!`
      })
    }, 1000)
  }
</script>

<style scoped>
  @import './style.css';
</style>

<style lang="scss" scoped>
  :deep(.el-select__wrapper) {
    height: 40px !important;
  }

  :deep(.login-submit.el-button--primary) {
    --el-button-bg-color: var(--auth-accent);
    --el-button-border-color: var(--auth-accent);
    --el-button-hover-bg-color: var(--auth-accent-hover);
    --el-button-hover-border-color: var(--auth-accent-hover);
    --el-button-active-bg-color: var(--auth-accent-active);
    --el-button-active-border-color: var(--auth-accent-active);
  }

  :deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
    color: var(--auth-link);
  }

  :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
    background-color: var(--auth-accent);
    border-color: var(--auth-accent);
  }

  :deep(.el-form-item__error) {
    color: var(--auth-danger);
  }
</style>
