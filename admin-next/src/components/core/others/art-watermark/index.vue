<!-- 水印组件 -->
<template>
  <div
    v-if="watermarkVisible"
    class="fixed left-0 top-0 h-screen w-screen pointer-events-none"
    :style="{ zIndex: zIndex }"
  >
    <ElWatermark
      :content="watermarkContent"
      :font="{ fontSize: fontSize, color: fontColor, fontGap: fontGap }"
      :rotate="rotate"
      :gap="[gapX, gapY]"
      :offset="[offsetX, offsetY]"
      :height="watermarkHeight"
      :width="watermarkWidth"
    >
      <div style="height: 100vh"></div>
    </ElWatermark>
  </div>
</template>

<script setup lang="ts">
  import { useSettingStore } from '@/store/modules/setting'
  import { useSiteStore } from '@/store/modules/site'

  defineOptions({ name: 'ArtWatermark' })

  const settingStore = useSettingStore()
  const siteStore = useSiteStore()
  const { watermarkVisible } = storeToRefs(settingStore)
  const { systemName } = storeToRefs(siteStore)

  interface WatermarkProps {
    /** 水印内容 */
    content?: string | string[]
    /** 水印是否可见 */
    visible?: boolean
    /** 水印字体大小 */
    fontSize?: number
    /** 水印行间距 */
    fontGap?: number
    /** 水印字体颜色 */
    fontColor?: string
    /** 水印旋转角度 */
    rotate?: number
    /** 水印间距X */
    gapX?: number
    /** 水印间距Y */
    gapY?: number
    /** 水印偏移X */
    offsetX?: number
    /** 水印偏移Y */
    offsetY?: number
    /** 水印层级 */
    zIndex?: number
  }

  const props = withDefaults(defineProps<WatermarkProps>(), {
    content: '',
    visible: false,
    fontSize: 16,
    fontGap: 8,
    fontColor: 'rgba(128, 128, 128, 0.2)',
    rotate: -22,
    gapX: 100,
    gapY: 100,
    offsetX: 50,
    offsetY: 50,
    zIndex: 3100
  })

  const watermarkContent = computed(() => {
    const content = Array.isArray(props.content) ? props.content : [props.content]
    const filteredContent = content.filter(Boolean)

    return filteredContent.length ? filteredContent : [systemName.value]
  })

  const watermarkWidth = computed(() => {
    const longestLine = watermarkContent.value.reduce(
      (max, line) => Math.max(max, String(line).length),
      0
    )

    return Math.max(120, Math.ceil(longestLine * props.fontSize * 0.8) + 32)
  })

  const watermarkHeight = computed(() => {
    const lineCount = watermarkContent.value.length || 1
    return Math.max(64, lineCount * props.fontSize + (lineCount - 1) * props.fontGap + 24)
  })
</script>
