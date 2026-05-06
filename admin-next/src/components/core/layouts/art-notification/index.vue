<!-- 通知组件 -->
<template>
  <div
    class="art-notification-panel art-card-sm !shadow-xl"
    :style="{
      transform: show ? 'scaleY(1)' : 'scaleY(0.9)',
      opacity: show ? 1 : 0
    }"
    v-show="visible"
    @click.stop
  >
    <div class="flex-cb px-3.5 mt-3.5">
      <span class="text-base font-medium text-g-800">{{ $t('notice.title') }}</span>
      <span
        class="text-xs text-g-800 px-1.5 py-1 c-p select-none rounded hover:bg-g-200"
        @click="emit('read-all')"
      >
        {{ $t('notice.btnRead') }}
      </span>
    </div>

    <ul class="box-border flex items-end w-full h-12.5 px-3.5 border-b-d">
      <li
        v-for="(item, index) in barList"
        :key="index"
        class="h-12 leading-12 mr-5 overflow-hidden text-[13px] text-g-700 c-p select-none"
        :class="{ 'bar-active': barActiveIndex === index }"
        @click="changeBar(index)"
      >
        {{ item.name }} ({{ item.num }})
      </li>
    </ul>

    <div class="w-full h-[calc(100%-95px)]">
      <div class="h-[calc(100%-60px)] overflow-y-scroll scrollbar-thin">
        <ul v-show="barActiveIndex === 0">
          <li
            v-for="item in noticeList"
            :key="item.id"
            class="box-border flex-c px-3.5 py-3.5 c-p last:border-b-0 hover:bg-g-200/60"
            @click="emit('notice-click', item.id)"
          >
            <div
              class="size-9 leading-9 text-center rounded-lg flex-cc"
              :class="[getNoticeStyle(item.type).iconClass]"
            >
              <ArtSvgIcon class="text-lg !bg-transparent" :icon="getNoticeStyle(item.type).icon" />
            </div>
            <div class="w-[calc(100%-45px)] ml-3.5">
              <h4 class="text-sm font-normal leading-5.5 text-g-900">{{ item.title }}</h4>
              <p class="mt-1.5 text-xs text-g-500">{{ item.time }}</p>
            </div>
          </li>
        </ul>

        <ul v-show="barActiveIndex === 1">
          <li
            v-for="item in msgList"
            :key="item.id"
            class="box-border flex-c px-3.5 py-3.5 c-p last:border-b-0 hover:bg-g-200/60"
            @click="emit('message-click', item.id)"
          >
            <div class="w-9 h-9">
              <img :src="item.avatar" class="w-full h-full rounded-lg object-cover" />
            </div>
            <div class="w-[calc(100%-45px)] ml-3.5">
              <h4 class="text-xs font-normal leading-5.5">{{ item.title }}</h4>
              <p class="mt-1.5 text-xs text-g-500">{{ item.time }}</p>
            </div>
          </li>
        </ul>

        <ul v-show="barActiveIndex === 2">
          <li
            v-for="item in pendingList"
            :key="item.id"
            class="box-border px-5 py-3.5 last:border-b-0"
          >
            <h4>{{ item.title }}</h4>
            <p class="text-xs text-g-500">{{ item.time }}</p>
          </li>
        </ul>

        <div
          v-show="currentTabIsEmpty"
          class="relative top-25 h-full text-g-500 text-center !bg-transparent"
        >
          <ArtSvgIcon icon="system-uicons:inbox" class="text-5xl" />
          <p class="mt-3.5 text-xs !bg-transparent"
            >{{ $t('notice.text[0]') }}{{ barList[barActiveIndex].name }}</p
          >
        </div>
      </div>

      <div class="relative box-border w-full px-3.5">
        <ElButton class="w-full mt-3" @click="handleViewAll" v-ripple>
          {{ $t('notice.viewAll') }}
        </ElButton>
      </div>
    </div>

    <div class="h-25"></div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'ArtNotification' })

  type NoticeType = 'email' | 'message' | 'collection' | 'user' | 'notice'
  type NotificationTab = 'notice' | 'message' | 'pending'

  interface NoticeItem {
    id: number | string
    title: string
    time: string
    type?: NoticeType
  }

  interface MessageItem {
    id: number | string
    title: string
    time: string
    avatar: string
    unreadCount?: number
  }

  interface PendingItem {
    id: number | string
    title: string
    time: string
  }

  interface NoticeStyle {
    icon: string
    iconClass: string
  }

  const { t } = useI18n()

  const props = withDefaults(
    defineProps<{
      value: boolean
      noticeList?: NoticeItem[]
      msgList?: MessageItem[]
      pendingList?: PendingItem[]
    }>(),
    {
      noticeList: () => [],
      msgList: () => [],
      pendingList: () => []
    }
  )

  const emit = defineEmits<{
    'update:value': [value: boolean]
    'read-all': []
    'message-click': [id: number | string]
    'notice-click': [id: number | string]
    'view-all': [tab: NotificationTab]
  }>()

  const show = ref(false)
  const visible = ref(false)
  const barActiveIndex = ref(0)

  const barList = computed(() => [
    { name: t('notice.bar[0]'), num: props.noticeList.length },
    { name: t('notice.bar[1]'), num: props.msgList.length },
    { name: t('notice.bar[2]'), num: props.pendingList.length }
  ])

  const currentTabIsEmpty = computed(() => {
    const tabDataMap = [props.noticeList, props.msgList, props.pendingList]
    return tabDataMap[barActiveIndex.value].length === 0
  })

  const noticeStyleMap: Record<NoticeType, NoticeStyle> = {
    email: {
      icon: 'ri:mail-line',
      iconClass: 'bg-warning/12 text-warning'
    },
    message: {
      icon: 'ri:volume-down-line',
      iconClass: 'bg-success/12 text-success'
    },
    collection: {
      icon: 'ri:heart-3-line',
      iconClass: 'bg-danger/12 text-danger'
    },
    user: {
      icon: 'ri:user-3-line',
      iconClass: 'bg-info/12 text-info'
    },
    notice: {
      icon: 'ri:notification-3-line',
      iconClass: 'bg-theme/12 text-theme'
    }
  }

  const getNoticeStyle = (type: NoticeType = 'notice'): NoticeStyle => {
    return (
      noticeStyleMap[type] || {
        icon: 'ri:arrow-right-circle-line',
        iconClass: 'bg-theme/12 text-theme'
      }
    )
  }

  const changeBar = (index: number) => {
    barActiveIndex.value = index
  }

  const handleViewAll = () => {
    const tabs: NotificationTab[] = ['notice', 'message', 'pending']
    emit('view-all', tabs[barActiveIndex.value])
    emit('update:value', false)
  }

  watch(
    () => props.value,
    (open) => {
      if (open) {
        visible.value = true
        setTimeout(() => {
          show.value = true
        }, 5)
      } else {
        show.value = false
        setTimeout(() => {
          visible.value = false
        }, 350)
      }
    }
  )
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .art-notification-panel {
    @apply absolute 
    top-14.5 
    right-5 
    w-90 
    h-125 
    overflow-hidden 
    transition-all 
    duration-300
    origin-top 
    will-change-[top,left] 
    max-[640px]:top-[65px]
    max-[640px]:right-0
    max-[640px]:w-full 
    max-[640px]:h-[80vh];
  }

  .bar-active {
    color: var(--theme-color) !important;
    border-bottom: 2px solid var(--theme-color);
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 5px !important;
  }

  .dark .scrollbar-thin::-webkit-scrollbar-track {
    background-color: var(--default-box-color);
  }

  .dark .scrollbar-thin::-webkit-scrollbar-thumb {
    background-color: #222 !important;
  }
</style>
