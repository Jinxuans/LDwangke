<template>
  <div class="menu-page art-full-height">
    <ArtSearchBar
      v-model="formFilters"
      :items="formItems"
      :showExpand="false"
      @reset="handleReset"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ElTabs v-model="activeTab" class="menu-tabs">
        <ElTabPane label="前台菜单" name="frontend">
          <ArtTableHeader
            :show-zebra="false"
            :loading="loading"
            layout="refresh,size,fullscreen,columns,settings"
            v-model:columns="menuColumnChecks"
            @refresh="loadData"
          >
            <template #left>
              <ElButton @click="toggleExpand" v-ripple>
                {{ isExpanded ? '收起' : '展开' }}
              </ElButton>
            </template>
          </ArtTableHeader>

          <ArtTable
            ref="menuTableRef"
            row-key="menu_key"
            :loading="loading"
            :columns="menuColumns"
            :data="filteredFrontendMenus"
            :stripe="false"
            :tree-props="{ children: 'children' }"
            :default-expand-all="false"
            :row-class-name="menuRowClassName"
          />
        </ElTabPane>

        <ElTabPane label="后台菜单" name="backend">
          <ArtTableHeader
            :show-zebra="false"
            :loading="loading"
            layout="refresh,size,fullscreen,columns,settings"
            v-model:columns="menuColumnChecks"
            @refresh="loadData"
          >
            <template #left>
              <ElButton @click="toggleExpand" v-ripple>
                {{ isExpanded ? '收起' : '展开' }}
              </ElButton>
            </template>
          </ArtTableHeader>

          <ArtTable
            ref="menuTableRef"
            row-key="menu_key"
            :loading="loading"
            :columns="menuColumns"
            :data="filteredBackendMenus"
            :stripe="false"
            :tree-props="{ children: 'children' }"
            :default-expand-all="false"
            :row-class-name="menuRowClassName"
          />
        </ElTabPane>

        <ElTabPane label="扩展菜单" name="ext">
          <ArtTableHeader
            :show-zebra="false"
            :loading="extLoading"
            layout="refresh,size,fullscreen,columns,settings"
            v-model:columns="extColumnChecks"
            @refresh="loadExtMenus"
          >
            <template #left>
              <ElButton v-auth="'add'" @click="openExtAdd" v-ripple>
                <template #icon>
                  <ArtSvgIcon icon="ri:add-line" />
                </template>
                添加扩展菜单
              </ElButton>
            </template>
          </ArtTableHeader>

          <ArtTable
            row-key="id"
            :loading="extLoading"
            :columns="extColumns"
            :data="filteredExtMenus"
            :stripe="false"
          />
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <MenuDialog
      v-model:visible="menuDialogVisible"
      :edit-data="menuEditData"
      :parent-options="menuParentOptions"
      :saving="menuSaving"
      @submit="handleMenuSubmit"
    />

    <ElDialog
      v-model="extDialogVisible"
      :title="extForm.id ? '编辑扩展菜单' : '添加扩展菜单'"
      width="620px"
      align-center
      class="ext-menu-dialog"
      @closed="resetExtForm"
    >
      <ArtForm
        ref="extFormRef"
        v-model="extForm"
        :items="extFormItems"
        :rules="extRules"
        :span="width > 640 ? 12 : 24"
        :gutter="20"
        label-width="96px"
        :show-reset="false"
        :show-submit="false"
      />

      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="extDialogVisible = false">取 消</ElButton>
          <ElButton type="primary" :loading="extSaving" @click="handleExtSubmit">保 存</ElButton>
        </span>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, nextTick, onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
  import { useWindowSize } from '@vueuse/core'
  import { formatMenuTitle } from '@/utils/router'
  import type { AppRouteRecord } from '@/types/router'
  import type { ColumnOption } from '@/types/component'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import ArtForm from '@/components/core/forms/art-form/index.vue'
  import MenuDialog, {
    type MenuDialogData,
    type MenuParentOption
  } from './modules/menu-dialog.vue'
  import { useMenuStore } from '@/store/modules/menu'
  import {
    deleteLegacyExtMenu,
    fetchLegacyExtMenus,
    saveLegacyExtMenu,
    saveLegacyMenuConfigs
  } from '@/api/legacy/menu'
  import { fetchGetMenuListWithConfigs } from '@/api/system-manage'
  import type { LegacyExtMenuItem, LegacyMenuConfigItem } from '@/types/legacy-contract'

  defineOptions({ name: 'Menus' })

  type MenuScope = 'frontend' | 'backend'

  interface MenuNode {
    key: string
    menu_key: string
    parent_key: string
    title: string
    icon: string
    path: string
    sort_order: number
    visible: number
    scope: MenuScope
    children: MenuNode[]
  }

  interface ExtMenuForm {
    id: number
    title: string
    icon: string
    url: string
    sort_order: number
    visible: boolean
    scope: MenuScope
  }

  const { width } = useWindowSize()
  const loading = ref(false)
  const extLoading = ref(false)
  const activeTab = ref<'frontend' | 'backend' | 'ext'>('frontend')
  const isExpanded = ref(false)
  const menuTableRef = ref<any>()
  const menuDialogVisible = ref(false)
  const menuSaving = ref(false)
  const menuEditData = ref<MenuDialogData | null>(null)
  const menuScope = ref<MenuScope>('frontend')
  const menuStore = useMenuStore()
  const extDialogVisible = ref(false)
  const extSaving = ref(false)
  const extFormRef = ref<FormInstance>()

  const formFilters = reactive({
    name: '',
    route: ''
  })

  const appliedFilters = reactive({
    name: '',
    route: ''
  })

  const menuTree = ref<Record<MenuScope, MenuNode[]>>({
    frontend: [],
    backend: []
  })
  const extMenus = ref<LegacyExtMenuItem[]>([])
  const extEditId = ref(0)
  const extForm = reactive<ExtMenuForm>({
    id: 0,
    title: '',
    icon: 'mdi:puzzle-outline',
    url: '',
    sort_order: 0,
    visible: true,
    scope: 'backend'
  })

  const formItems = computed<FormItem[]>(() => [
    {
      label: '菜单名称',
      key: 'name',
      type: 'input',
      props: { clearable: true, placeholder: '按名称搜索' }
    },
    {
      label: '路由地址',
      key: 'route',
      type: 'input',
      props: { clearable: true, placeholder: '按路径搜索' }
    }
  ])

  const menuColumnChecks = ref(
    getMenuColumns().map((item) => ({ ...item, checked: item.visible ?? true, visible: item.visible ?? true }))
  )
  const extColumnChecks = ref(
    getExtColumns().map((item) => ({ ...item, checked: item.visible ?? true, visible: item.visible ?? true }))
  )

  const menuColumns = computed(() =>
    menuColumnChecks.value.filter((item) => item.visible !== false && item.checked !== false)
  )
  const extColumns = computed(() =>
    extColumnChecks.value.filter((item) => item.visible !== false && item.checked !== false)
  )

  const extRules: FormRules = {
    title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
    url: [{ required: true, message: '请输入地址', trigger: 'blur' }]
  }

  const extFormItems = computed<FormItem[]>(() => [
    { label: '标题', key: 'title', type: 'input', props: { maxlength: 80 } },
    {
      label: '地址',
      key: 'url',
      type: 'input',
      props: { placeholder: '/php-api/ext/index.php 或完整URL' }
    },
    { label: '图标', key: 'icon', type: 'input', props: { placeholder: 'mdi:puzzle-outline' } },
    {
      label: '位置',
      key: 'scope',
      type: 'select',
      props: {
        options: [
          { label: '前台', value: 'frontend' },
          { label: '后台', value: 'backend' }
        ],
        style: { width: '100%' }
      }
    },
    {
      label: '排序',
      key: 'sort_order',
      type: 'number',
      props: { min: -99, max: 999, controlsPosition: 'right', style: { width: '100%' } }
    },
    { label: '是否显示', key: 'visible', type: 'switch', span: width.value > 640 ? 6 : 12 }
  ])

  const filteredFrontendMenus = computed(() => filterMenuTree(menuTree.value.frontend))
  const filteredBackendMenus = computed(() => filterMenuTree(menuTree.value.backend))
  const filteredExtMenus = computed(() => {
    const name = appliedFilters.name.trim().toLowerCase()
    const route = appliedFilters.route.trim().toLowerCase()
    return extMenus.value.filter((item) => {
      const title = String(item.title || '').toLowerCase()
      const url = String(item.url || '').toLowerCase()
      return (!name || title.includes(name)) && (!route || url.includes(route))
    })
  })

  const menuParentOptions = computed<MenuParentOption[]>(() => {
    const tree = menuScope.value === 'backend' ? menuTree.value.backend : menuTree.value.frontend
    const current = findNode(tree, menuEditData.value?.menu_key || '')
    const excludeKeys = new Set<string>()
    const options: MenuParentOption[] = [{ value: '', label: '顶级菜单' }]

    const collect = (node: MenuNode) => {
      excludeKeys.add(node.key)
      node.children.forEach(collect)
    }

    if (current) {
      collect(current)
    }

    const walk = (nodes: MenuNode[], depth = 0) => {
      nodes.forEach((node) => {
        if (!excludeKeys.has(node.key)) {
          options.push({
            value: node.menu_key,
            label: `${'　'.repeat(depth)}${node.title}`
          })
        }
        if (node.children.length) {
          walk(node.children, depth + 1)
        }
      })
    }

    walk(tree)
    return options
  })

  function getMenuColumns(): ColumnOption<MenuNode>[] {
    return [
      {
        prop: 'title',
        label: '菜单名称',
        minWidth: 220,
        formatter: (row) => renderMenuName(row.title, row.icon)
      },
      {
        prop: 'path',
        label: '路由',
        minWidth: 180,
        formatter: (row) => row.path
      },
      {
        prop: 'menu_key',
        label: 'Key',
        minWidth: 160,
        formatter: (row) => row.menu_key
      },
      {
        prop: 'sort_order',
        label: '排序',
        width: 90,
        align: 'center',
        formatter: (row) => String(row.sort_order)
      },
      {
        prop: 'visible',
        label: '状态',
        width: 90,
        align: 'center',
        formatter: (row) => h(ElTag, { type: row.visible === 1 ? 'success' : 'danger' }, () => (row.visible === 1 ? '启用' : '停用'))
      },
      {
        prop: 'operation',
        label: '操作',
        width: 110,
        align: 'center',
        formatter: (row) =>
          h('div', { class: 'flex items-center justify-center' }, [
            h(ArtButtonTable, {
              type: 'edit',
              onClick: () => openMenuEdit(row, row.scope)
            })
          ])
      }
    ]
  }

  function getExtColumns(): ColumnOption<LegacyExtMenuItem>[] {
    return [
      {
        prop: 'title',
        label: '标题',
        minWidth: 220,
        formatter: (row) => renderMenuName(row.title, row.icon || 'mdi:puzzle-outline')
      },
      { prop: 'url', label: '地址', minWidth: 220 },
      {
        prop: 'scope',
        label: '位置',
        width: 90,
        align: 'center',
        formatter: (row) =>
          h(ElTag, { type: row.scope === 'frontend' ? 'primary' : 'warning' }, () =>
            row.scope === 'frontend' ? '前台' : '后台'
          )
      },
      { prop: 'sort_order', label: '排序', width: 90, align: 'center' },
      {
        prop: 'visible',
        label: '显示',
        width: 90,
        align: 'center',
        formatter: (row) =>
          h(ElTag, { type: row.visible === 1 ? 'success' : 'info' }, () =>
            row.visible === 1 ? '显示' : '隐藏'
          )
      },
      {
        prop: 'operation',
        label: '操作',
        width: 120,
        align: 'center',
        formatter: (row) =>
          h('div', { class: 'flex items-center justify-center' }, [
            h(ArtButtonTable, { type: 'edit', onClick: () => openExtEdit(row) }),
            h(ArtButtonTable, { type: 'delete', onClick: () => handleExtDelete(row) })
          ])
      }
    ]
  }

  function renderMenuName(title: string, icon: string) {
    return h('div', { class: 'menu-name-cell' }, [
      icon
        ? h(ArtSvgIcon, { icon, class: 'menu-name-icon' })
        : h('span', { class: 'menu-name-icon is-empty' }),
      h('span', { class: 'menu-name-text' }, title)
    ])
  }

  function inferScope(route: AppRouteRecord, inherited?: MenuScope): MenuScope {
    if (inherited) return inherited
    const path = String(route.path || '')
    return path.startsWith('/admin') || path.startsWith('/system') ? 'backend' : 'frontend'
  }

  function buildMenuItems(
    routes: AppRouteRecord[],
    configMap: Map<string, LegacyMenuConfigItem>,
    parentKey = '',
    inheritedScope?: MenuScope
  ): MenuNode[] {
    const items: MenuNode[] = []

    routes.forEach((route, index) => {
      const routeName = String(route.name || '').trim()
      const saved = routeName ? configMap.get(routeName) : undefined
      const rawTitle = route.meta?.title ? formatMenuTitle(String(route.meta.title)) : ''
      const title = saved?.title || rawTitle || routeName || String(route.path || '')
      const scope = inheritedScope || (saved?.scope as MenuScope | undefined) || inferScope(route)
      const routeOrder = Number(route.meta?.order)
      const sortOrder = saved?.sort_order ?? (Number.isFinite(routeOrder) ? routeOrder : index)

      if (routeName && title) {
        items.push({
          key: routeName,
          menu_key: routeName,
          parent_key: parentKey,
          title,
          icon: saved?.icon || String(route.meta?.icon || ''),
          path: String(route.path || ''),
          sort_order: sortOrder,
          visible: saved ? Number(saved.visible) : route.meta?.isHide ? 0 : 1,
          scope,
          children: []
        })
      }

      if (route.children?.length) {
        items.push(
          ...buildMenuItems(route.children as AppRouteRecord[], configMap, routeName || parentKey, scope)
        )
      }
    })

    return items
  }

  function buildTree(nodes: MenuNode[]): MenuNode[] {
    const map = new Map<string, MenuNode>()
    const roots: MenuNode[] = []

    nodes.forEach((node) => {
      map.set(node.menu_key, { ...node, children: [] })
    })

    nodes.forEach((node) => {
      const current = map.get(node.menu_key)
      if (!current) return
      if (node.parent_key && map.has(node.parent_key)) {
        map.get(node.parent_key)!.children.push(current)
      } else {
        roots.push(current)
      }
    })

    return sortMenuTree(roots)
  }

  function sortMenuTree(nodes: MenuNode[]): MenuNode[] {
    nodes.sort((a, b) => a.sort_order - b.sort_order)
    nodes.forEach((node) => {
      if (node.children.length) {
        node.children = sortMenuTree(node.children)
      }
    })
    return nodes
  }

  function filterMenuTree(nodes: MenuNode[]): MenuNode[] {
    const name = appliedFilters.name.trim().toLowerCase()
    const route = appliedFilters.route.trim().toLowerCase()

    const walk = (list: MenuNode[]): MenuNode[] => {
      const result: MenuNode[] = []

      list.forEach((node) => {
        const children = walk(node.children)
        const matched =
          (!name || node.title.toLowerCase().includes(name)) &&
          (!route || node.path.toLowerCase().includes(route))

        if (matched || children.length) {
          result.push({ ...node, children })
        }
      })

      return result
    }

    return walk(nodes)
  }

  function findNode(nodes: MenuNode[], key: string): MenuNode | null {
    for (const node of nodes) {
      if (node.key === key) return node
      if (node.children.length) {
        const found = findNode(node.children, key)
        if (found) return found
      }
    }
    return null
  }

  function findParentKey(nodes: MenuNode[], key: string, parentKey = ''): string {
    for (const node of nodes) {
      if (node.key === key) return parentKey
      if (node.children.length) {
        const found = findParentKey(node.children, key, node.menu_key)
        if (found !== '__NOT_FOUND__') return found
      }
    }
    return '__NOT_FOUND__'
  }

  function detachNode(nodes: MenuNode[], key: string): MenuNode | null {
    for (let i = 0; i < nodes.length; i += 1) {
      if (nodes[i].key === key) {
        const [removed] = nodes.splice(i, 1)
        return removed || null
      }
      if (nodes[i].children.length) {
        const found = detachNode(nodes[i].children, key)
        if (found) return found
      }
    }
    return null
  }

  function attachNode(nodes: MenuNode[], parentKey: string, node: MenuNode): boolean {
    if (!parentKey) {
      nodes.push(node)
      return true
    }

    for (const current of nodes) {
      if (current.key === parentKey) {
        current.children.push(node)
        return true
      }
      if (current.children.length && attachNode(current.children, parentKey, node)) {
        return true
      }
    }

    return false
  }

  function flattenTree(nodes: MenuNode[], scope: MenuScope, parentKey = ''): LegacyMenuConfigItem[] {
    const result: LegacyMenuConfigItem[] = []

    nodes.forEach((node) => {
      result.push({
        menu_key: node.menu_key,
        parent_key: parentKey,
        title: node.title,
        icon: node.icon,
        sort_order: node.sort_order,
        visible: node.visible,
        scope
      })
      result.push(...flattenTree(node.children, scope, node.menu_key))
    })

    return result
  }

  function menuRowClassName({ row }: { row: MenuNode }) {
    return row.visible === 0 ? 'menu-row--hidden' : ''
  }

  async function loadData() {
    loading.value = true
    try {
      const { routes, menuConfigs } = await fetchGetMenuListWithConfigs()
      const configMap = new Map(menuConfigs.map((item) => [item.menu_key, item]))
      const sourceRoutes = menuStore.menuList.length ? menuStore.menuList : routes
      const items = buildMenuItems(sourceRoutes, configMap)
      menuTree.value.frontend = buildTree(items.filter((item) => item.scope === 'frontend'))
      menuTree.value.backend = buildTree(items.filter((item) => item.scope === 'backend'))
    } catch (error) {
      console.error('加载菜单失败', error)
      ElMessage.error('加载菜单失败')
    } finally {
      loading.value = false
    }
  }

  function handleReset() {
    Object.assign(formFilters, {
      name: '',
      route: ''
    })
    Object.assign(appliedFilters, {
      name: '',
      route: ''
    })
  }

  function handleSearch() {
    Object.assign(appliedFilters, { ...formFilters })
  }

  function toggleExpand() {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      const table = menuTableRef.value?.elTableRef
      const rows = activeTab.value === 'backend' ? filteredBackendMenus.value : filteredFrontendMenus.value
      if (!table || !rows.length) return

      const processRows = (list: MenuNode[]) => {
        list.forEach((row) => {
          if (row.children?.length) {
            table.toggleRowExpansion(row, isExpanded.value)
            processRows(row.children)
          }
        })
      }

      processRows(rows)
    })
  }

  function openMenuEdit(row: MenuNode, scope: MenuScope) {
    menuScope.value = scope
    menuEditData.value = {
      menu_key: row.menu_key,
      parent_key: findParentKey(scope === 'backend' ? menuTree.value.backend : menuTree.value.frontend, row.menu_key),
      path: row.path,
      name: row.title,
      icon: row.icon,
      sort: row.sort_order,
      isEnable: row.visible === 1
    }
    menuDialogVisible.value = true
  }

  async function handleMenuSubmit(form: MenuDialogData) {
    const tree = menuScope.value === 'backend' ? menuTree.value.backend : menuTree.value.frontend
    const node = findNode(tree, form.menu_key)
    if (!node) {
      ElMessage.error('菜单不存在')
      return
    }

    const originalParent = findParentKey(tree, node.menu_key)
    const newParent = form.parent_key

    node.title = form.name
    node.icon = form.icon
    node.sort_order = Number(form.sort || 0)
    node.visible = form.isEnable ? 1 : 0

    if (newParent !== originalParent) {
      const moved = detachNode(tree, node.menu_key)
      if (moved) {
        moved.parent_key = newParent
        if (!attachNode(tree, newParent, moved)) {
          moved.parent_key = ''
          tree.push(moved)
        }
      }
    }

    menuSaving.value = true
    try {
      const allItems = [
        ...flattenTree(menuTree.value.frontend, 'frontend'),
        ...flattenTree(menuTree.value.backend, 'backend')
      ]
      await saveLegacyMenuConfigs(allItems)
      ElMessage.success('菜单已保存')
      menuDialogVisible.value = false
      await loadData()
    } catch (error) {
      console.error('保存菜单失败', error)
      ElMessage.error('保存菜单失败')
      await loadData()
    } finally {
      menuSaving.value = false
    }
  }

  function openExtAdd() {
    extEditId.value = 0
    extForm.id = 0
    extForm.title = ''
    extForm.icon = 'mdi:puzzle-outline'
    extForm.url = ''
    extForm.sort_order = 0
    extForm.visible = true
    extForm.scope = 'backend'
    extDialogVisible.value = true
  }

  function openExtEdit(item: LegacyExtMenuItem) {
    extEditId.value = item.id
    extForm.id = item.id
    extForm.title = item.title
    extForm.icon = item.icon
    extForm.url = item.url
    extForm.sort_order = item.sort_order
    extForm.visible = item.visible === 1
    extForm.scope = item.scope as MenuScope
    extDialogVisible.value = true
  }

  function resetExtForm() {
    extEditId.value = 0
    extForm.id = 0
    extForm.title = ''
    extForm.icon = 'mdi:puzzle-outline'
    extForm.url = ''
    extForm.sort_order = 0
    extForm.visible = true
    extForm.scope = 'backend'
  }

  async function loadExtMenus() {
    extLoading.value = true
    try {
      const list = await fetchLegacyExtMenus()
      extMenus.value = Array.isArray(list) ? list : []
    } catch (error) {
      console.error('加载扩展菜单失败', error)
      extMenus.value = []
    } finally {
      extLoading.value = false
    }
  }

  async function handleExtSubmit() {
    if (!extFormRef.value) return
    await extFormRef.value.validate()

    extSaving.value = true
    try {
      await saveLegacyExtMenu({
        id: extForm.id,
        title: extForm.title.trim(),
        icon: extForm.icon.trim(),
        url: extForm.url.trim(),
        sort_order: extForm.sort_order,
        visible: extForm.visible ? 1 : 0,
        scope: extForm.scope
      })
      ElMessage.success('扩展菜单已保存')
      extDialogVisible.value = false
      await loadExtMenus()
    } catch (error) {
      console.error('保存扩展菜单失败', error)
      ElMessage.error('保存扩展菜单失败')
    } finally {
      extSaving.value = false
    }
  }

  async function handleExtDelete(item: LegacyExtMenuItem) {
    try {
      await ElMessageBox.confirm(`确定删除扩展菜单「${item.title}」？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await deleteLegacyExtMenu(item.id)
      ElMessage.success('已删除')
      await loadExtMenus()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除扩展菜单失败', error)
        ElMessage.error('删除失败')
      }
    }
  }

  onMounted(async () => {
    await loadData()
    await loadExtMenus()
  })
</script>

<style scoped>
.menu-page {
  min-height: 0;
  overflow: hidden;
}

:deep(.art-table-card) {
  min-height: 0;
}

:deep(.art-table-card > .el-card__body) {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

:deep(.menu-row--hidden) {
  opacity: 0.55;
}

:deep(.menu-name-cell) {
  display: inline-flex;
  max-width: 100%;
  align-items: center;
  gap: 8px;
  vertical-align: middle;
}

:deep(.menu-name-icon) {
  width: 18px;
  min-width: 18px;
  height: 18px;
  font-size: 18px;
  color: var(--el-text-color-secondary);
}

:deep(.menu-name-icon.is-empty) {
  opacity: 0;
}

:deep(.menu-name-text) {
  min-width: 0;
  overflow: hidden;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.menu-tabs) {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

:deep(.menu-tabs > .el-tabs__content) {
  min-height: 0;
  flex: 1;
  overflow: hidden;
}

:deep(.menu-tabs > .el-tabs__content > .el-tab-pane) {
  display: flex;
  height: 100%;
  min-height: 0;
  flex-direction: column;
}

:deep(.menu-tabs .art-table) {
  min-height: 0;
  flex: 1;
}
</style>
