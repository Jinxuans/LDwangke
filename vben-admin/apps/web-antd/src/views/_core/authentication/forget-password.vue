<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationForgetPassword, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { forgotPasswordApi, resetPasswordApi } from '#/api/core/auth';

defineOptions({ name: 'ForgetPassword' });

const loading = ref(false);
const step = ref(1);
const savedEmail = ref('');
const countdown = ref(0);
const message = ref('');
let countdownTimer: ReturnType<typeof setInterval> | null = null;
const router = useRouter();

function startCountdown() {
  countdown.value = 60;
  countdownTimer = setInterval(() => {
    countdown.value--;
    if (countdown.value <= 0) {
      clearInterval(countdownTimer!);
      countdownTimer = null;
    }
  }, 1000);
}

const formSchema = computed((): VbenFormSchema[] => {
  if (step.value === 1) {
    return [
      {
        component: 'VbenInput',
        componentProps: {
          placeholder: '请输入绑定的邮箱地址',
        },
        fieldName: 'email',
        label: '邮箱地址',
        rules: z
          .string()
          .min(1, { message: '请输入邮箱' })
          .email('请输入有效的邮箱地址'),
      },
    ];
  }
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: '请输入验证码',
      },
      fieldName: 'code',
      label: '验证码',
      rules: z.string().min(1, { message: '请输入验证码' }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: '请输入新密码（至少6位）',
      },
      fieldName: 'password',
      label: '新密码',
      rules: z.string().min(6, { message: '密码至少6位' }),
    },
  ];
});

const buttonText = computed(() => {
  if (step.value === 1) {
    return countdown.value > 0 ? `${countdown.value}s 后可重新发送` : '发送验证码';
  }
  return '重置密码';
});

async function handleSubmit(values: Recordable<any>) {
  loading.value = true;
  message.value = '';
  try {
    if (step.value === 1) {
      await forgotPasswordApi(values.email);
      savedEmail.value = values.email;
      startCountdown();
      step.value = 2;
      message.value = '验证码已发送到您的邮箱';
    } else {
      await resetPasswordApi(savedEmail.value, values.code, values.password);
      message.value = '密码重置成功，即将跳转登录...';
      setTimeout(() => router.push('/auth/login'), 2000);
    }
  } catch (e: any) {
    message.value = e?.message || '操作失败';
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <AuthenticationForgetPassword
    :form-schema="formSchema"
    :loading="loading"
    :submit-button-text="buttonText"
    :sub-title="step === 2 ? `验证码已发送至 ${savedEmail}` : '输入您绑定的邮箱地址来重置密码'"
    @submit="handleSubmit"
  />
  <p v-if="message" class="mt-2 text-center text-sm text-green-600">
    {{ message }}
  </p>
</template>
