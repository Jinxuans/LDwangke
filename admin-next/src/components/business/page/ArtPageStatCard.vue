<template>
  <article class="art-page-stat-card art-card-xs" :class="toneClass">
    <div class="art-page-stat-card__head">
      <p class="art-page-stat-card__label">{{ label }}</p>
      <span class="art-page-stat-card__tone" aria-hidden="true"></span>
    </div>
    <h3 class="art-page-stat-card__value">{{ value }}</h3>
    <p v-if="helper" class="art-page-stat-card__helper">{{ helper }}</p>
  </article>
</template>

<script setup lang="ts">
  defineOptions({ name: 'ArtPageStatCard' })

  type Tone = 'primary' | 'success' | 'warning' | 'info'

  interface Props {
    label: string
    value: number | string
    helper?: string
    tone?: Tone
  }

  const props = withDefaults(defineProps<Props>(), {
    helper: '',
    tone: 'primary'
  })

  const toneClass = computed(() => `art-page-stat-card--${props.tone}`)
</script>

<style scoped lang="scss">
  .art-page-stat-card {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 104px;
    padding: 16px 18px;
    overflow: hidden;

    &__head {
      display: flex;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
    }

    &__label {
      margin: 0;
      font-size: 13px;
      color: var(--art-gray-600);
    }

    &__tone {
      width: 8px;
      height: 8px;
      flex-shrink: 0;
      background: var(--art-page-stat-accent);
      border-radius: 999px;
      box-shadow: 0 0 0 5px var(--art-page-stat-accent-soft);
    }

    &__value {
      margin: 0;
      font-size: 24px;
      font-weight: 600;
      line-height: 1.2;
      color: var(--art-gray-900);
    }

    &__helper {
      margin: 0;
      font-size: 12px;
      color: var(--art-gray-500);
      line-height: 1.5;
    }
  }

  .art-page-stat-card--primary {
    --art-page-stat-accent: var(--el-color-primary);
    --art-page-stat-accent-soft: var(--el-color-primary-light-8);
  }

  .art-page-stat-card--success {
    --art-page-stat-accent: var(--el-color-success);
    --art-page-stat-accent-soft: var(--el-color-success-light-8);
  }

  .art-page-stat-card--warning {
    --art-page-stat-accent: var(--el-color-warning);
    --art-page-stat-accent-soft: var(--el-color-warning-light-8);
  }

  .art-page-stat-card--info {
    --art-page-stat-accent: var(--el-color-info);
    --art-page-stat-accent-soft: var(--el-color-info-light-8);
  }
</style>
