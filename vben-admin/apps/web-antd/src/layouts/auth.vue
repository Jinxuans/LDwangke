<script lang="ts" setup>
import { computed, ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';

import { AuthPageLayout } from '@vben/layouts';
import { preferences, updatePreferences } from '@vben/preferences';

import { $t } from '#/locales';
import { getSiteConfigApi } from '#/api/admin';

const route = useRoute();
const appName = computed(() => preferences.app.name);
const logo = computed(() => preferences.logo.source);
const maintenanceMsg = ref('');

onMounted(async () => {
  try {
    const cfg = await getSiteConfigApi();
    if (cfg?.sitename) {
      updatePreferences({ app: { name: cfg.sitename } });
      document.title = cfg.sitename;
    }
    if (cfg?.logo) {
      updatePreferences({ logo: { source: cfg.logo } });
    }
    if (cfg?.bz === '1') {
      maintenanceMsg.value = '系统维护中，仅管理员可登录使用';
    }
  } catch { /* ignore */ }
  if (route.query.msg === 'maintenance') {
    maintenanceMsg.value = '系统维护中，仅管理员可登录使用';
  }
});
</script>

<template>
  <AuthPageLayout
    :app-name="appName"
    :logo="logo"
    :page-description="$t('authentication.pageDesc')"
    :page-title="$t('authentication.pageTitle')"
  >
    <template #toolbar>
      <div v-if="maintenanceMsg" class="w-full rounded-lg border border-orange-300 bg-orange-50 px-4 py-2 text-center text-sm text-orange-700">
        ⚠️ {{ maintenanceMsg }}
      </div>
    </template>
  </AuthPageLayout>
</template>
