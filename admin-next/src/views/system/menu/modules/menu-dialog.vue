<template>
  <ElDialog
    title="编辑菜单"
    :model-value="visible"
    width="760px"
    align-center
    class="menu-dialog"
    @update:model-value="handleCancel"
    @closed="handleClosed"
  >
    <ArtForm
      ref="formRef"
      v-model="form"
      :items="formItems"
      :rules="rules"
      :span="width > 640 ? 12 : 24"
      :gutter="20"
      label-width="96px"
      :show-reset="false"
      :show-submit="false"
    />

    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="handleCancel">取 消</ElButton>
        <ElButton type="primary" :loading="saving" @click="handleSubmit">保 存</ElButton>
      </span>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormRules } from 'element-plus'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import ArtForm from '@/components/core/forms/art-form/index.vue'
  import { useWindowSize } from '@vueuse/core'

  export interface MenuParentOption {
    label: string
    value: string
  }

  export interface MenuDialogData {
    menu_key: string
    parent_key: string
    path: string
    name: string
    icon: string
    sort: number
    isEnable: boolean
  }

  interface Props {
    visible: boolean
    editData?: Partial<MenuDialogData> | null
    parentOptions?: MenuParentOption[]
    saving?: boolean
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', data: MenuDialogData): void
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    editData: null,
    parentOptions: () => [],
    saving: false
  })

  const emit = defineEmits<Emits>()
  const { width } = useWindowSize()
  const formRef = ref()

  const form = reactive<MenuDialogData>({
    menu_key: '',
    parent_key: '',
    path: '',
    name: '',
    icon: '',
    sort: 0,
    isEnable: true
  })

  const rules = reactive<FormRules>({
    name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }]
  })

  const formItems = computed<FormItem[]>(() => [
    {
      label: '路由 Key',
      key: 'menu_key',
      type: 'input',
      props: { disabled: true }
    },
    {
      label: '路由地址',
      key: 'path',
      type: 'input',
      props: { disabled: true }
    },
    {
      label: '上级菜单',
      key: 'parent_key',
      type: 'select',
      props: {
        clearable: false,
        disabled: true,
        filterable: true,
        options: props.parentOptions,
        style: { width: '100%' }
      }
    },
    {
      label: '菜单名称',
      key: 'name',
      type: 'input',
      props: { maxlength: 80, placeholder: '菜单名称' }
    },
    {
      label: '图标',
      key: 'icon',
      type: 'input',
      props: { maxlength: 80, placeholder: '如：ri:menu-line、mdi:cog-outline' }
    },
    {
      label: '菜单排序',
      key: 'sort',
      type: 'number',
      props: { min: -99, max: 999, controlsPosition: 'right', style: { width: '100%' } }
    },
    { label: '是否启用', key: 'isEnable', type: 'switch', span: width.value < 640 ? 12 : 6 }
  ])

  const loadFormData = (): void => {
    const data = props.editData
    if (!data) return

    form.menu_key = data.menu_key || ''
    form.parent_key = data.parent_key || ''
    form.path = data.path || ''
    form.name = data.name || ''
    form.icon = data.icon || ''
    form.sort = Number(data.sort || 0)
    form.isEnable = data.isEnable ?? true
  }

  const resetForm = (): void => {
    formRef.value?.reset()
    form.menu_key = ''
    form.parent_key = ''
    form.path = ''
    form.name = ''
    form.icon = ''
    form.sort = 0
    form.isEnable = true
  }

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return
    await formRef.value.validate()
    emit('submit', { ...form, name: form.name.trim(), icon: form.icon.trim() })
  }

  const handleCancel = (): void => {
    emit('update:visible', false)
  }

  const handleClosed = (): void => {
    resetForm()
  }

  watch(
    () => props.visible,
    (visible) => {
      if (visible) {
        nextTick(loadFormData)
      }
    }
  )
</script>
