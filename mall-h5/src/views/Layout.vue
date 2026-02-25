<script setup lang="ts">
import { watch, computed } from "vue";
import { useRoute } from "vue-router";
import { Tabbar, TabbarItem } from "vant";
import { setTid } from "../api";

const route = useRoute();
const tid = computed(() => route.params.tid as string);

watch(
  () => route.params.tid as string,
  (tid) => { if (tid) setTid(tid); },
  { immediate: true },
);

// Determine active tab based on current route
const activeTab = computed(() => {
  const path = route.path;
  if (path.includes('/query')) return 'query';
  if (path.includes('/orders')) return 'orders';
  return 'home';
});

// Hide tabbar on Product and PayResult pages
const showTabbar = computed(() => {
  const path = route.path;
  // Only show on Home, Query, Orders pages
  return !path.includes('/product/') && !path.includes('/pay-result');
});
</script>

<template>
  <div class="mall-layout">
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in" appear>
        <component :is="Component" :key="route.fullPath" />
      </transition>
    </router-view>
    
    <!-- Bottom Tabbar -->
    <Tabbar 
      v-if="showTabbar && tid" 
      v-model="activeTab" 
      class="bottom-tabbar"
      active-color="var(--primary-color)"
      inactive-color="var(--text-muted)"
    >
      <TabbarItem name="home" icon="home-o" :to="`/${tid}`">
        首页
      </TabbarItem>
      <TabbarItem name="query" icon="search" :to="`/${tid}/query`">
        查订单
      </TabbarItem>
      <TabbarItem name="orders" icon="orders-o" :to="`/${tid}/orders`">
        我的订单
      </TabbarItem>
    </Tabbar>
  </div>
</template>

<style scoped>
.mall-layout {
  max-width: 480px;
  margin: 0 auto;
  min-height: 100vh;
  background: var(--bg-primary);
  overflow-x: hidden;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
@media (prefers-reduced-motion: reduce) {
  .fade-enter-active,
  .fade-leave-active {
    transition-duration: 0.01ms;
  }
}
.bottom-tabbar {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 480px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-light);
  z-index: 100;
}
</style>
