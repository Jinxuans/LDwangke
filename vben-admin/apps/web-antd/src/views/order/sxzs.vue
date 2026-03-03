<script setup lang="ts">
defineOptions({ name: 'SXZSIndex' });
</script>

<template>
  <div class="iframe-container">
    <iframe
      src="/php-api/sxzsredis/sxzs.php"
      title="实习助手"
      class="iframe-content"
      frameborder="0"
      allowfullscreen
    />
  </div>
</template>

<style scoped>
.iframe-container {
  width: 100%;
  height: calc(100vh - 100px);
  overflow: hidden;
}
.iframe-content {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
