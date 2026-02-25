<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed, defineComponent, h, markRaw, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { AuthenticationRegister, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { getSiteConfigApi } from '#/api/admin';
import { sendCodeApi } from '#/api/core/auth';
import { useAuthStore } from '#/store';

defineOptions({ name: 'Register' });

const route = useRoute();
const router = useRouter();
const inviteCode = (route.query.invite as string) || '';
const authStore = useAuthStore();
const emailVerifyEnabled = ref(false);
const countdown = ref(0);
const sendingCode = ref(false);
const registerRef = ref<any>(null);
let countdownTimer: ReturnType<typeof setInterval> | null = null;

onMounted(async () => {
  try {
    const cfg = await getSiteConfigApi();
    if (cfg?.login_email_verify === '1') {
      emailVerifyEnabled.value = true;
    }
  } catch { /* ignore */ }
});

async function handleSendCode() {
  const formApi = registerRef.value?.getFormApi?.();
  const values = formApi ? await formApi.getValues() : {};
  const email = values.email || '';
  if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    return;
  }
  sendingCode.value = true;
  try {
    await sendCodeApi(email, 'register');
    countdown.value = 60;
    countdownTimer = setInterval(() => {
      countdown.value--;
      if (countdown.value <= 0) {
        clearInterval(countdownTimer!);
        countdownTimer = null;
      }
    }, 1000);
  } catch {
    /* ignore */
  } finally {
    sendingCode.value = false;
  }
}

const VerifyCodeInput = markRaw(defineComponent({
  props: { modelValue: { type: String, default: '' } },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () =>
      h('div', { class: 'flex w-full items-center gap-2' }, [
        h('input', {
          value: props.modelValue,
          onInput: (e: Event) =>
            emit('update:modelValue', (e.target as HTMLInputElement).value),
          class:
            'border-input bg-background placeholder:text-muted-foreground flex h-10 w-full flex-1 rounded-md border px-3 py-2 text-sm outline-none focus:border-primary',
          placeholder: '请输入验证码',
          type: 'text',
        }),
        h(
          'button',
          {
            type: 'button',
            disabled: countdown.value > 0 || sendingCode.value,
            onClick: handleSendCode,
            class:
              'bg-primary text-primary-foreground h-10 shrink-0 rounded-md px-4 text-sm font-medium disabled:opacity-50',
          },
          countdown.value > 0 ? `${countdown.value}s` : '发送验证码',
        ),
      ]);
  },
}));

const formSchema = computed((): VbenFormSchema[] => {
  const schema: VbenFormSchema[] = [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: '请输入您的昵称',
      },
      fieldName: 'name',
      label: '用户昵称',
      rules: z.string().min(1, { message: '昵称不能为空' }),
    },
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: '请输入 QQ 号作为账号',
      },
      fieldName: 'user',
      label: '登录账号',
      rules: z
        .string()
        .min(5, { message: '账号最少 5 位' })
        .max(11, { message: '账号最多 11 位' })
        .regex(/^\d+$/, { message: '账号必须为纯数字 QQ 号' }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: '请输入登录密码',
      },
      fieldName: 'pass',
      label: '登录密码',
      rules: z.string().min(1, { message: '密码不能为空' }),
    },
  ];

  if (emailVerifyEnabled.value) {
    schema.push(
      {
        component: 'VbenInput',
        componentProps: {
          placeholder: '请输入邮箱地址',
          type: 'email',
        },
        fieldName: 'email',
        label: '邮箱',
        rules: z.string().email({ message: '请输入有效邮箱' }),
      },
      {
        component: VerifyCodeInput as any,
        fieldName: 'verify_code',
        label: '验证码',
        rules: z.string().min(1, { message: '请输入验证码' }),
      },
    );
  }

  schema.push({
    component: 'VbenInput',
    componentProps: {
      placeholder: '请输入邀请码',
    },
    defaultValue: inviteCode,
    fieldName: 'yqm',
    label: '邀请码',
    rules: z.string().min(1, { message: '邀请码不能为空' }),
  });

  schema.push({
    component: 'VbenCheckbox',
    fieldName: 'agreePolicy',
    renderComponentContent: () => ({
      default: () =>
        h('span', [
          '我已阅读并同意',
          h(
            'a',
            {
              class: 'vben-link ml-1 ',
              href: '',
            },
            '用户协议与隐私政策',
          ),
        ]),
    }),
    rules: z.boolean().refine((value) => !!value, {
      message: '请先同意协议',
    }),
  });

  return schema;
});

async function handleSubmit(values: Recordable<any>) {
  await authStore.authRegister(values);
}
</script>

<template>
  <div>
    <AuthenticationRegister
      ref="registerRef"
      :form-schema="formSchema"
      :loading="authStore.loginLoading"
      @submit="handleSubmit"
    />
  </div>
</template>
