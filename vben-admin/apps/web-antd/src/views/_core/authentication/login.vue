<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, markRaw, onMounted, ref } from 'vue';

import { AuthenticationLogin, SliderCaptcha, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { getSiteConfigApi } from '#/api/admin';
import { useAuthStore } from '#/store';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const sliderEnabled = ref(true);

onMounted(async () => {
  try {
    const cfg = await getSiteConfigApi();
    if (cfg?.login_slider_verify === '0') {
      sliderEnabled.value = false;
    }
  } catch { /* ignore */ }
});

const formSchema = computed((): VbenFormSchema[] => {
  const schema: VbenFormSchema[] = [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: '请输入账号',
      },
      fieldName: 'username',
      label: '登录账号',
      rules: z.string().min(1, { message: '请输入账号' }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: '请输入密码',
      },
      fieldName: 'password',
      label: '登录密码',
      rules: z.string().min(1, { message: '请输入密码' }),
    },
  ];
  if (sliderEnabled.value) {
    schema.push({
      component: markRaw(SliderCaptcha),
      fieldName: 'captcha',
      rules: z.boolean().refine((value) => value, {
        message: '请完成验证',
      }),
    });
  }
  return schema;
});
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    @submit="authStore.authLogin"
  />
</template>
