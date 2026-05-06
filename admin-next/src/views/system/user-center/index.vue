<!-- 个人中心页面 -->
<template>
  <div class="system-user-center-page art-full-height">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-4 border-b-d pb-4">
        <div class="flex items-center gap-3">
          <img class="h-12 w-12 rounded-full border-full-d object-cover" src="@imgs/user/avatar.webp" />
          <div>
            <p class="text-base font-semibold text-g-900">{{ userInfo.userName }}</p>
            <p class="mt-1 text-sm text-g-500">维护个人资料、联系方式和登录信息。</p>
          </div>
        </div>
        <div class="flex flex-wrap gap-2">
          <ElTag effect="plain">邮箱 {{ form.email }}</ElTag>
          <ElTag type="success" effect="plain">手机 {{ form.mobile }}</ElTag>
          <ElTag type="info" effect="plain">地址 {{ form.address }}</ElTag>
        </div>
      </div>

      <div class="mt-5 grid gap-3 md:grid-cols-3">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <p class="text-xs text-g-400">真实姓名</p>
          <p class="mt-2 text-sm font-medium text-g-900">{{ form.realName }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <p class="text-xs text-g-400">昵称</p>
          <p class="mt-2 text-sm font-medium text-g-900">{{ form.nikeName }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-3">
          <p class="text-xs text-g-400">性别</p>
          <p class="mt-2 text-sm font-medium text-g-900">{{ form.sex === '1' ? '男' : '女' }}</p>
        </article>
      </div>
    </section>

    <div class="mt-4 space-y-4">
      <section class="art-card-sm">
        <div class="border-b-d px-5 py-4">
          <h2 class="text-lg font-semibold text-g-900">基本设置</h2>
        </div>

        <ElForm
          ref="ruleFormRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="p-5 [&_.el-input]:w-full [&_.el-select]:w-full"
        >
          <div class="grid gap-4 md:grid-cols-2">
            <ElFormItem label="姓名" prop="realName">
              <ElInput v-model="form.realName" :disabled="!isEdit" />
            </ElFormItem>
            <ElFormItem label="性别" prop="sex">
              <ElSelect v-model="form.sex" placeholder="请选择性别" :disabled="!isEdit">
                <ElOption v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="昵称" prop="nikeName">
              <ElInput v-model="form.nikeName" :disabled="!isEdit" />
            </ElFormItem>
            <ElFormItem label="邮箱" prop="email">
              <ElInput v-model="form.email" :disabled="!isEdit" />
            </ElFormItem>
            <ElFormItem label="手机" prop="mobile">
              <ElInput v-model="form.mobile" :disabled="!isEdit" />
            </ElFormItem>
            <ElFormItem label="地址" prop="address">
              <ElInput v-model="form.address" :disabled="!isEdit" />
            </ElFormItem>
          </div>

          <ElFormItem class="mt-1" label="个人介绍" prop="des">
            <ElInput v-model="form.des" type="textarea" :rows="4" :disabled="!isEdit" />
          </ElFormItem>

          <div class="flex justify-end">
            <ElButton type="primary" @click="edit">
              {{ isEdit ? '保存' : '编辑资料' }}
            </ElButton>
          </div>
        </ElForm>
      </section>

      <section class="art-card-sm">
        <div class="border-b-d px-5 py-4">
          <h2 class="text-lg font-semibold text-g-900">更改密码</h2>
        </div>

        <ElForm :model="pwdForm" label-position="top" class="p-5">
          <div class="grid gap-4 md:grid-cols-3">
            <ElFormItem label="当前密码" prop="password">
              <ElInput v-model="pwdForm.password" type="password" :disabled="!isEditPwd" show-password />
            </ElFormItem>
            <ElFormItem label="新密码" prop="newPassword">
              <ElInput v-model="pwdForm.newPassword" type="password" :disabled="!isEditPwd" show-password />
            </ElFormItem>
            <ElFormItem label="确认新密码" prop="confirmPassword">
              <ElInput v-model="pwdForm.confirmPassword" type="password" :disabled="!isEditPwd" show-password />
            </ElFormItem>
          </div>

          <div class="flex justify-end">
            <ElButton type="primary" @click="editPwd">
              {{ isEditPwd ? '保存密码' : '修改密码' }}
            </ElButton>
          </div>
        </ElForm>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useUserStore } from '@/store/modules/user'
  import type { FormInstance, FormRules } from 'element-plus'

  defineOptions({ name: 'UserCenter' })

  const userStore = useUserStore()
  const userInfo = computed(() => userStore.getUserInfo)

  const isEdit = ref(false)
  const isEditPwd = ref(false)
  const ruleFormRef = ref<FormInstance>()

  /**
   * 用户信息表单
   */
  const form = reactive({
    realName: 'John Snow',
    nikeName: '皮卡丘',
    email: '59301283@mall.com',
    mobile: '18888888888',
    address: '广东省深圳市宝安区西乡街道101栋201',
    sex: '2',
    des: '维护个人资料、联系方式和常用账号信息。'
  })

  /**
   * 密码修改表单
   */
  const pwdForm = reactive({
    password: '123456',
    newPassword: '123456',
    confirmPassword: '123456'
  })

  /**
   * 表单验证规则
   */
  const rules = reactive<FormRules>({
    realName: [
      { required: true, message: '请输入姓名', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    nikeName: [
      { required: true, message: '请输入昵称', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
    mobile: [{ required: true, message: '请输入手机号码', trigger: 'blur' }],
    address: [{ required: true, message: '请输入地址', trigger: 'blur' }],
    sex: [{ required: true, message: '请选择性别', trigger: 'blur' }]
  })

  /**
   * 性别选项
   */
  const options = [
    { value: '1', label: '男' },
    { value: '2', label: '女' }
  ]

  /**
   * 切换用户信息编辑状态
   */
  const edit = () => {
    isEdit.value = !isEdit.value
  }

  /**
   * 切换密码编辑状态
   */
  const editPwd = () => {
    isEditPwd.value = !isEditPwd.value
  }
</script>
