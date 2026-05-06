<template>
  <section class="art-page-header art-card-sm">
    <div class="art-page-header__main">
      <div class="art-page-header__content">
        <p v-if="eyebrow" class="art-page-header__eyebrow">{{ eyebrow }}</p>
        <h1 class="art-page-header__title">{{ title }}</h1>
        <p v-if="description" class="art-page-header__description">{{ description }}</p>
      </div>

      <div v-if="$slots.actions" class="art-page-header__actions">
        <slot name="actions" />
      </div>
    </div>

    <div v-if="$slots.meta" class="art-page-header__meta">
      <slot name="meta" />
    </div>

    <div v-if="$slots.stats" class="art-page-header__stats">
      <slot name="stats" />
    </div>
  </section>
</template>

<script setup lang="ts">
  defineOptions({ name: 'ArtPageHeader' })

  interface Props {
    title: string
    description?: string
    eyebrow?: string
  }

  defineProps<Props>()
</script>

<style scoped lang="scss">
  .art-page-header {
    padding: 18px 20px;
    margin-bottom: 16px;

    &__main {
      display: flex;
      gap: 14px;
      align-items: flex-start;
      justify-content: space-between;
    }

    &__content {
      min-width: 0;
    }

    &__eyebrow {
      margin-bottom: 6px;
      font-size: 11px;
      font-weight: 600;
      color: var(--theme-color);
      letter-spacing: 0.12em;
      text-transform: uppercase;
    }

    &__title {
      margin: 0;
      font-size: 26px;
      font-weight: 600;
      line-height: 1.25;
      color: var(--art-gray-900);
    }

    &__description {
      max-width: 760px;
      margin-top: 8px;
      font-size: 13px;
      line-height: 1.7;
      color: var(--art-gray-600);
    }

    &__actions {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
      justify-content: flex-end;
      flex-shrink: 0;
    }

    &__meta {
      margin-top: 14px;
    }

    &__stats {
      padding-top: 14px;
      margin-top: 14px;
      border-top: 1px solid var(--art-card-border);
    }
  }

  @media (width <= 768px) {
    .art-page-header {
      padding: 16px;

      &__main {
        flex-direction: column;
      }

      &__actions {
        justify-content: flex-start;
      }

      &__title {
        font-size: 22px;
      }
    }
  }
</style>
