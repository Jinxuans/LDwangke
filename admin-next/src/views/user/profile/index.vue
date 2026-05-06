<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="grid gap-5 xl:grid-cols-[1.15fr_0.85fr]">
      <article class="art-card-sm p-5">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">{{ displayName }}</h2>
            <p class="mt-1 text-sm text-g-500">UID {{ profile?.uid || 0 }} / 账号 {{ profile?.user || '-' }}</p>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <ElTag effect="plain">{{ profile?.grade_name || '未识别等级' }}</ElTag>
            <ElTag type="info" effect="plain">费率 {{ Number(profile?.addprice || 0).toFixed(2) }}</ElTag>
            <ElButton plain :loading="loading" @click="loadProfile">刷新</ElButton>
          </div>
        </div>

        <div class="mt-4 flex flex-wrap gap-3">
          <ElTag :type="profile?.khcz ? 'success' : 'info'" effect="plain">
            {{ profile?.khcz ? '已开通跨户充值' : '跨户充值关闭' }}
          </ElTag>
          <ElTag v-if="profile?.sjuser" type="warning" effect="plain">上级：{{ profile.sjuser }}</ElTag>
        </div>

        <div class="mt-5 grid gap-3">
          <button
            type="button"
            class="flex w-full items-center justify-between rounded-custom-sm border-full-d bg-box px-4 py-4 text-left transition hover:border-[var(--el-color-primary-light-5)]"
            @click="copyText(String(profile?.uid || ''))"
          >
            <div>
              <p class="text-xs font-medium text-g-400">接口账号</p>
              <p class="mt-2 text-base font-semibold text-g-900">{{ profile?.uid || '-' }}</p>
            </div>
            <span class="text-sm text-[var(--el-color-primary)]">复制</span>
          </button>

          <button
            type="button"
            class="flex w-full items-center justify-between rounded-custom-sm border-full-d bg-box px-4 py-4 text-left transition hover:border-[var(--el-color-primary-light-5)]"
            @click="copyText(profile?.key && profile.key !== '0' ? profile.key : '')"
          >
            <div>
              <p class="text-xs font-medium text-g-400">接口密钥</p>
              <p class="mt-2 text-sm font-semibold text-g-900 break-all">
                {{ profile?.key && profile.key !== '0' ? profile.key : '未开通' }}
              </p>
            </div>
            <span class="text-sm text-g-500">
              {{ profile?.key && profile.key !== '0' ? '复制' : '待开通' }}
            </span>
          </button>

          <button
            type="button"
            class="flex w-full items-center justify-between rounded-custom-sm border-full-d bg-box px-4 py-4 text-left transition hover:border-[var(--el-color-primary-light-5)]"
            @click="copyText(inviteUrl)"
          >
            <div>
              <p class="text-xs font-medium text-g-400">邀请链接</p>
              <p class="mt-2 text-sm font-semibold text-g-900 break-all">
                {{ inviteUrl || '请先设置邀请码' }}
              </p>
            </div>
            <span class="text-sm text-[var(--el-color-primary)]">{{ inviteUrl ? '复制' : '未生成' }}</span>
          </button>
        </div>

        <div class="mt-5 flex flex-wrap gap-3">
          <ElButton type="primary" @click="passwordVisible = true">修改密码</ElButton>
          <ElButton v-if="profile?.grade === '3'" plain @click="pass2Visible = true">二级密码</ElButton>
          <ElButton plain @click="openInviteRateDialog">邀请等级</ElButton>
          <ElButton v-if="isAdminGrade" plain @click="openSelfGradeDialog">设置我的等级</ElButton>
          <ElButton v-if="migrateEnabled" plain @click="migrateVisible = true">上级迁移</ElButton>
        </div>
      </article>

      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">邀请与分销</h2>
            <p class="mt-1 text-sm text-g-500">邀请码、邀请等级和推送 Token 集中维护。</p>
          </div>
          <ElTag effect="plain">{{ siteStore.systemName }}</ElTag>
        </div>

        <div class="mt-5 space-y-4">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">邀请码</p>
            <div class="mt-3 flex items-center justify-between gap-3">
              <div>
                <p class="text-lg font-semibold text-g-900">{{ profile?.yqm || '未设置' }}</p>
                <p class="mt-1 text-sm text-g-500">分享注册链接时会自动带上这个邀请码。</p>
              </div>
              <ElButton text type="primary" @click="copyText(profile?.yqm || '')">复制</ElButton>
            </div>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">邀请费率</p>
            <div class="mt-3 flex items-start justify-between gap-3">
              <div>
                <p class="text-lg font-semibold text-g-900">{{ inviteGradeText }}</p>
                <p class="mt-1 text-sm text-g-500">受邀用户注册后会应用这里选定的费率等级。</p>
              </div>
              <ElButton text type="primary" @click="openInviteRateDialog">设置</ElButton>
            </div>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">推送 Token</p>
            <div class="mt-3 flex items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-g-900 break-all">
                  {{ profile?.push_token || '未设置' }}
                </p>
                <p class="mt-1 text-sm text-g-500">订单状态变化可通过该 Token 接收消息推送。</p>
              </div>
              <ElButton text type="primary" @click="openPushTokenDialog">编辑</ElButton>
            </div>
          </article>
        </div>

        <div
          v-if="profile?.dailitongji"
          class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4"
        >
          <div class="flex flex-wrap gap-3">
            <ElTag effect="plain">代理总数 {{ profile.dailitongji.dlzs || 0 }}</ElTag>
            <ElTag type="success" effect="plain">今日活跃 {{ profile.dailitongji.dldl || 0 }}</ElTag>
            <ElTag type="primary" effect="plain">今日新增 {{ profile.dailitongji.dlzc || 0 }}</ElTag>
            <ElTag type="warning" effect="plain">今日交单 {{ profile.dailitongji.jrjd || 0 }}</ElTag>
          </div>
        </div>
      </article>
    </section>

    <section v-if="profile?.notice || profile?.sjnotice" class="grid gap-5 xl:grid-cols-2">
      <article v-if="profile?.notice" class="art-card-sm p-5">
        <p class="text-sm font-semibold text-g-800">站点公告</p>
        <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-g-700">{{ profile.notice }}</p>
      </article>
      <article v-if="profile?.sjnotice" class="art-card-sm p-5">
        <p class="text-sm font-semibold text-g-800">上级公告</p>
        <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-g-700">{{ profile.sjnotice }}</p>
      </article>
    </section>

    <section class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
      <article class="art-card-sm p-5" v-loading="loading">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">账户信息</h2>
            <p class="mt-1 text-sm text-g-500">基础资料、等级、邀请码和联系方式都在这里集中维护。</p>
          </div>
          <ElTag type="info" effect="plain">基础资料</ElTag>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">基础信息</p>
            <div class="mt-3 space-y-3 text-sm text-g-700">
              <p>昵称：{{ profile?.name || '-' }}</p>
              <p>邮箱：{{ profile?.email || '-' }}</p>
              <p>手机号：{{ profile?.phone || '-' }}</p>
              <p>总充值：¥{{ formatMoney(profile?.zcz) }}</p>
              <p>今日订单：{{ profile?.today_orders || 0 }}</p>
            </div>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">等级信息</p>
            <div class="mt-3 space-y-3 text-sm text-g-700">
              <p>当前等级：{{ profile?.grade_name || '-' }}</p>
              <p>邀请等级：{{ inviteGradeText }}</p>
              <p>邀请码：{{ profile?.yqm || '未设置' }}</p>
              <p>储值金额：¥{{ formatMoney(profile?.cdmoney) }}</p>
              <p>商城冻结：¥{{ formatMoney(profile?.mall_cdmoney) }}</p>
            </div>
          </article>
        </div>

      </article>

      <article class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-g-900">接口与安全</h2>
            <p class="mt-1 text-sm text-g-500">接口密钥、邀请码和推送配置在这里集中维护。</p>
          </div>
          <ElTag effect="plain">安全设置</ElTag>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">密钥状态</p>
          <p class="mt-3 text-sm font-semibold text-g-900 break-all">
            {{ profile?.key && profile.key !== '0' ? profile.key : '未开通' }}
          </p>
          <p class="mt-2 text-sm text-g-500">
            {{ profile?.key && profile.key !== '0' ? '更换后旧密钥会立即失效。' : '开通接口密钥需要扣费，余额满足条件时可免扣。' }}
          </p>
        </div>

        <div class="mt-5 flex flex-wrap gap-3">
          <ElButton plain @click="openInviteCodeDialog">
            {{ profile?.yqm ? '修改邀请码' : '设置邀请码' }}
          </ElButton>
          <ElButton plain @click="openPushTokenDialog">推送 Token</ElButton>
          <ElPopconfirm
            :title="profile?.key && profile.key !== '0' ? '确认更换接口密钥？旧密钥会立即失效。' : '开通接口密钥需要扣除 5 元，余额满 100 免扣，确认继续？'"
            width="300"
            @confirm="handleSecretKeyAction(profile?.key && profile.key !== '0' ? 3 : 1)"
          >
            <template #reference>
              <ElButton type="primary" plain>
                {{ profile?.key && profile.key !== '0' ? '更换接口密钥' : '开通接口密钥' }}
              </ElButton>
            </template>
          </ElPopconfirm>
        </div>
      </article>
    </section>

    <ElDialog v-model="passwordVisible" title="修改密码" width="520px">
      <div class="space-y-4 py-2">
        <ElInput v-model="passwordForm.oldpass" type="password" show-password placeholder="旧密码" />
        <ElInput v-model="passwordForm.newpass" type="password" show-password placeholder="新密码，至少 6 位" />
        <ElInput
          v-model="passwordForm.confirm"
          type="password"
          show-password
          placeholder="确认新密码"
          @keyup.enter="submitPassword"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="passwordVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="passwordSubmitting" @click="submitPassword">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="pass2Visible" title="设置二级密码" width="520px">
      <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
        二级密码用于管理员登录二次验证。首次设置时旧密码可以留空。
      </div>
      <div class="mt-4 space-y-4">
        <ElInput v-model="pass2Form.oldPass2" type="password" show-password placeholder="旧二级密码，可留空" />
        <ElInput v-model="pass2Form.newPass2" type="password" show-password placeholder="新二级密码，至少 6 位" />
        <ElInput
          v-model="pass2Form.confirm"
          type="password"
          show-password
          placeholder="确认新二级密码"
          @keyup.enter="submitPass2"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="pass2Visible = false">取消</ElButton>
          <ElButton type="primary" :loading="pass2Submitting" @click="submitPass2">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="inviteCodeVisible" title="设置邀请码" width="480px">
      <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
        邀请码最少 4 位，其他人通过注册链接注册后会归属到你的名下。
      </div>
      <ElInput
        v-model="inviteCodeForm"
        class="mt-4"
        clearable
        maxlength="32"
        placeholder="输入邀请码"
        @keyup.enter="submitInviteCode"
      />
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="inviteCodeVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="inviteCodeSubmitting" @click="submitInviteCode">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="pushTokenVisible" title="设置推送 Token" width="520px">
      <ElInput
        v-model="pushTokenForm"
        type="textarea"
        :rows="4"
        resize="none"
        placeholder="输入推送 Token"
      />
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="pushTokenVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="pushTokenSubmitting" @click="submitPushToken">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="inviteRateVisible" title="设置邀请等级" width="520px">
      <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
        受邀用户注册后，会继承这里选中的等级费率。
      </div>
      <ElSelect v-model="inviteRateForm.gradeId" class="mt-4 w-full" filterable placeholder="请选择邀请等级">
        <ElOption
          v-for="item in gradeOptions"
          :key="item.id"
          :label="`${item.name}（费率 ${item.rate}）`"
          :value="item.id"
        />
      </ElSelect>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="inviteRateVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="inviteRateSubmitting" @click="submitInviteRate">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="selfGradeVisible" title="设置我的等级" width="520px">
      <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
        仅管理员可直接调整自己的等级，设置后会立即影响费率和定价。
      </div>
      <ElSelect v-model="selfGradeForm.gradeId" class="mt-4 w-full" filterable placeholder="请选择等级">
        <ElOption
          v-for="item in gradeOptions"
          :key="item.id"
          :label="`${item.name}（费率 ${item.rate}）`"
          :value="item.id"
        />
      </ElSelect>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="selfGradeVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="selfGradeSubmitting" @click="submitSelfGrade">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="migrateVisible" title="上级迁移" width="560px">
      <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3 text-sm text-g-700">
        输入新上级 UID 和邀请码。原上级需满足系统规则后才能迁移成功。
      </div>
      <div class="mt-4 space-y-4">
        <ElInputNumber v-model="migrateForm.uid" class="w-full" :min="1" placeholder="新上级 UID" />
        <ElInput v-model="migrateForm.yqm" clearable placeholder="新上级邀请码" @keyup.enter="submitMigrate" />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="migrateVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="migrateSubmitting" @click="submitMigrate">确认迁移</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { storeToRefs } from 'pinia'
  import { ElMessage } from 'element-plus'
  import {
    changeLegacyPass2,
    changeLegacyPassword,
    changeLegacySecretKey,
    fetchLegacyUserGrades,
    fetchLegacyUserProfile,
    migrateLegacySuperior,
    setLegacyInviteCode,
    setLegacyInviteRate,
    setLegacyMyGrade,
    setLegacyPushToken,
    type LegacyGradeOption,
    type LegacyUserProfile
  } from '@/api/legacy/user-center'
  import { useSiteStore } from '@/store/modules/site'

  defineOptions({ name: 'UserProfilePage' })

  const siteStore = useSiteStore()
  const { config } = storeToRefs(siteStore)

  const loading = ref(false)
  const profile = ref<LegacyUserProfile | null>(null)
  const gradeOptions = ref<LegacyGradeOption[]>([])
  const gradeOptionsLoading = ref(false)

  const passwordVisible = ref(false)
  const passwordSubmitting = ref(false)
  const passwordForm = reactive({
    oldpass: '',
    newpass: '',
    confirm: ''
  })

  const pass2Visible = ref(false)
  const pass2Submitting = ref(false)
  const pass2Form = reactive({
    oldPass2: '',
    newPass2: '',
    confirm: ''
  })

  const inviteCodeVisible = ref(false)
  const inviteCodeSubmitting = ref(false)
  const inviteCodeForm = ref('')

  const pushTokenVisible = ref(false)
  const pushTokenSubmitting = ref(false)
  const pushTokenForm = ref('')

  const inviteRateVisible = ref(false)
  const inviteRateSubmitting = ref(false)
  const inviteRateForm = reactive({
    gradeId: undefined as number | undefined
  })

  const selfGradeVisible = ref(false)
  const selfGradeSubmitting = ref(false)
  const selfGradeForm = reactive({
    gradeId: undefined as number | undefined
  })

  const migrateVisible = ref(false)
  const migrateSubmitting = ref(false)
  const migrateForm = reactive({
    uid: undefined as number | undefined,
    yqm: ''
  })

  const displayName = computed(
    () => profile.value?.name || profile.value?.user || siteStore.systemName || '未命名用户'
  )

  const isAdminGrade = computed(() => ['2', '3'].includes(String(profile.value?.grade || '')))
  const migrateEnabled = computed(() => String(config.value.sjqykg || '0') === '1')

  const inviteGradeText = computed(() => {
    if (!profile.value?.invite_grade_name) {
      return '未设置'
    }
    return `${profile.value.invite_grade_name}（费率 ${Number(profile.value.invite_addprice || 0).toFixed(2)}）`
  })

  const inviteUrl = computed(() => {
    if (!profile.value?.yqm) {
      return ''
    }
    const invite = encodeURIComponent(profile.value.yqm)
    return `${window.location.origin}${window.location.pathname}#/auth/register?invite=${invite}`
  })

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  const resetPasswordForm = () => {
    passwordForm.oldpass = ''
    passwordForm.newpass = ''
    passwordForm.confirm = ''
  }

  const resetPass2Form = () => {
    pass2Form.oldPass2 = ''
    pass2Form.newPass2 = ''
    pass2Form.confirm = ''
  }

  const loadProfile = async () => {
    loading.value = true
    try {
      profile.value = await fetchLegacyUserProfile()
    } finally {
      loading.value = false
    }
  }

  const ensureGradeOptions = async () => {
    if (gradeOptions.value.length || gradeOptionsLoading.value) {
      return
    }
    gradeOptionsLoading.value = true
    try {
      gradeOptions.value = await fetchLegacyUserGrades()
    } finally {
      gradeOptionsLoading.value = false
    }
  }

  const copyText = async (text: string) => {
    if (!text) {
      ElMessage.warning('暂无可复制内容')
      return
    }

    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动复制')
    }
  }

  const submitPassword = async () => {
    if (!passwordForm.oldpass.trim() || !passwordForm.newpass.trim()) {
      ElMessage.warning('请填写完整密码信息')
      return
    }
    if (passwordForm.newpass.trim().length < 6) {
      ElMessage.warning('新密码至少 6 位')
      return
    }
    if (passwordForm.newpass !== passwordForm.confirm) {
      ElMessage.warning('两次输入的新密码不一致')
      return
    }

    passwordSubmitting.value = true
    try {
      await changeLegacyPassword(passwordForm.oldpass.trim(), passwordForm.newpass.trim())
      ElMessage.success('密码修改成功')
      passwordVisible.value = false
      resetPasswordForm()
    } finally {
      passwordSubmitting.value = false
    }
  }

  const submitPass2 = async () => {
    if (!pass2Form.newPass2.trim()) {
      ElMessage.warning('请填写新二级密码')
      return
    }
    if (pass2Form.newPass2.trim().length < 6) {
      ElMessage.warning('二级密码至少 6 位')
      return
    }
    if (pass2Form.newPass2 !== pass2Form.confirm) {
      ElMessage.warning('两次输入的二级密码不一致')
      return
    }

    pass2Submitting.value = true
    try {
      await changeLegacyPass2(pass2Form.oldPass2.trim(), pass2Form.newPass2.trim())
      ElMessage.success('二级密码修改成功')
      pass2Visible.value = false
      resetPass2Form()
    } finally {
      pass2Submitting.value = false
    }
  }

  const openInviteCodeDialog = () => {
    inviteCodeForm.value = profile.value?.yqm || ''
    inviteCodeVisible.value = true
  }

  const submitInviteCode = async () => {
    if (!inviteCodeForm.value.trim() || inviteCodeForm.value.trim().length < 4) {
      ElMessage.warning('邀请码最少 4 位')
      return
    }

    inviteCodeSubmitting.value = true
    try {
      await setLegacyInviteCode(inviteCodeForm.value.trim())
      ElMessage.success('邀请码设置成功')
      inviteCodeVisible.value = false
      await loadProfile()
    } finally {
      inviteCodeSubmitting.value = false
    }
  }

  const openPushTokenDialog = () => {
    pushTokenForm.value = profile.value?.push_token || ''
    pushTokenVisible.value = true
  }

  const submitPushToken = async () => {
    pushTokenSubmitting.value = true
    try {
      await setLegacyPushToken(pushTokenForm.value.trim())
      ElMessage.success('推送 Token 已更新')
      pushTokenVisible.value = false
      await loadProfile()
    } finally {
      pushTokenSubmitting.value = false
    }
  }

  const openInviteRateDialog = async () => {
    await ensureGradeOptions()
    inviteRateForm.gradeId = profile.value?.invite_grade_id || profile.value?.grade_id || undefined
    inviteRateVisible.value = true
  }

  const submitInviteRate = async () => {
    if (!inviteRateForm.gradeId) {
      ElMessage.warning('请选择邀请等级')
      return
    }

    inviteRateSubmitting.value = true
    try {
      await setLegacyInviteRate(inviteRateForm.gradeId)
      ElMessage.success('邀请等级已更新')
      inviteRateVisible.value = false
      await loadProfile()
    } finally {
      inviteRateSubmitting.value = false
    }
  }

  const openSelfGradeDialog = async () => {
    await ensureGradeOptions()
    selfGradeForm.gradeId = profile.value?.grade_id || undefined
    selfGradeVisible.value = true
  }

  const submitSelfGrade = async () => {
    if (!selfGradeForm.gradeId) {
      ElMessage.warning('请选择等级')
      return
    }

    selfGradeSubmitting.value = true
    try {
      await setLegacyMyGrade(selfGradeForm.gradeId)
      ElMessage.success('等级已更新')
      selfGradeVisible.value = false
      await loadProfile()
    } finally {
      selfGradeSubmitting.value = false
    }
  }

  const handleSecretKeyAction = async (type: number) => {
    const result = await changeLegacySecretKey(type)
    ElMessage.success(type === 1 ? '接口密钥已开通' : '接口密钥已更换')
    if (result?.key) {
      await copyText(result.key)
    }
    await loadProfile()
  }

  const submitMigrate = async () => {
    if (!migrateForm.uid || !migrateForm.yqm.trim()) {
      ElMessage.warning('请填写完整的迁移信息')
      return
    }

    migrateSubmitting.value = true
    try {
      await migrateLegacySuperior(migrateForm.uid, migrateForm.yqm.trim())
      ElMessage.success('迁移成功')
      migrateVisible.value = false
      migrateForm.uid = undefined
      migrateForm.yqm = ''
      await loadProfile()
    } finally {
      migrateSubmitting.value = false
    }
  }

  watch(passwordVisible, (value) => {
    if (!value) resetPasswordForm()
  })

  watch(pass2Visible, (value) => {
    if (!value) resetPass2Form()
  })

  onMounted(async () => {
    await siteStore.initPublicConfig()
    await loadProfile()
  })
</script>
