<script setup lang="ts">
import { useRouter } from "vue-router";
const router = useRouter();
</script>

<template>
  <div class="not-found-page">
    <div class="nf-body animate-fade-in-up">
      <div class="nf-icon">
        <van-icon name="warn-o" size="40" color="var(--primary-color)" />
      </div>
      <div class="nf-code">404</div>
      <p class="nf-title">页面不存在</p>
      <p class="nf-sub">您访问的页面已移除或地址有误</p>
      <div class="nf-actions">
        <van-button round type="primary" block class="nf-btn" @click="router.back()">返回上一页</van-button>
        <van-button round plain type="primary" block class="nf-btn" @click="router.push('/')">返回首页</van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.not-found-page {
  min-height: 100vh;
  background: var(--bg-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}
.nf-body {
  text-align: center;
  max-width: 320px;
  width: 100%;
}
.nf-icon {
  width: 72px;
  height: 72px;
  background: var(--primary-bg);
  border-radius: var(--radius-xl);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}
.nf-code {
  font-size: 64px;
  font-weight: 800;
  color: var(--text-disabled);
  line-height: 1;
  margin-bottom: 16px;
  letter-spacing: -2px;
}
.nf-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}
.nf-sub {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 32px;
  line-height: 1.6;
}
.nf-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.nf-btn {
  height: 44px !important;
  font-size: 14px !important;
  font-weight: 600 !important;
}
</style>
