<script setup>
import { ref, shallowRef, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import PageTabs from './PageTabs.vue'
import WidgetGrid from './WidgetGrid.vue'
import JsonSchemaEditor from '../json-schema-editor'
import { PencilIcon, CheckIcon, XMarkIcon, ArrowUturnLeftIcon } from '@heroicons/vue/24/outline'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
})

const emit = defineEmits(['error'])

const pages = ref([])
const widgets = ref([])
const activePage = ref('')
const editMode = ref(false)
const loading = ref(false)
const pendingLayoutChanges = ref([])
const editingSchemaWidget = ref(null)
// Use shallowRef to prevent Vue from deeply proxying the schema object
// This allows json-schema-editor to mutate it directly without interference
const schemaEditorValue = shallowRef(null)
const showResetConfirm = ref(null)
const schemaDialogRef = ref(null)

// Load widget pages
const loadPages = async () => {
  if (!GoApp) return
  try {
    const result = await GoApp.GetWidgetPages(props.ctx, props.ns, props.projectName)
    pages.value = result || []
    if (pages.value.length > 0 && !activePage.value) {
      activePage.value = pages.value[0].resourceName
    }
  } catch (err) {
    emit('error', `Failed to load pages: ${err}`)
  }
}

// Load widgets for the active page
const loadWidgets = async () => {
  if (!GoApp || !activePage.value) return
  loading.value = true
  try {
    const result = await GoApp.GetWidgets(props.ctx, props.ns, props.projectName, activePage.value)
    // Add _updateTime to each widget for reactivity tracking
    widgets.value = (result || []).map(w => ({ ...w, _updateTime: Date.now() }))
  } catch (err) {
    emit('error', `Failed to load widgets: ${err}`)
  } finally {
    loading.value = false
  }
}

// Handle page selection
const selectPage = (pageName) => {
  activePage.value = pageName
}

// Handle page creation
const handleCreatePage = async (title) => {
  if (!GoApp) return
  try {
    const newPage = await GoApp.CreateDashboardPage(props.ctx, props.ns, props.projectName, title)
    if (newPage) {
      pages.value.push(newPage)
      activePage.value = newPage.resourceName
    }
  } catch (err) {
    emit('error', `Failed to create page: ${err}`)
  }
}

// Handle page deletion
const handleDeletePage = async (pageResourceName) => {
  if (!GoApp) return
  try {
    await GoApp.DeleteDashboardPage(props.ctx, props.ns, pageResourceName)
    // Remove page from list
    const idx = pages.value.findIndex(p => p.resourceName === pageResourceName)
    if (idx >= 0) {
      pages.value.splice(idx, 1)
    }
    // If deleted page was active, switch to first available page
    if (activePage.value === pageResourceName) {
      activePage.value = pages.value.length > 0 ? pages.value[0].resourceName : ''
    }
  } catch (err) {
    emit('error', `Failed to delete page: ${err}`)
  }
}

// Handle widget action (send signal)
const handleAction = async (actionData) => {
  if (!GoApp) return
  try {
    const dataStr = JSON.stringify(actionData.data)
    await GoApp.SendSignal(props.ctx, props.ns, actionData.nodeName, actionData.port, dataStr)
  } catch (err) {
    emit('error', `Failed to send signal: ${err}`)
  }
}

// Handle layout changes
const handleLayoutChange = (changes) => {
  pendingLayoutChanges.value = changes
}

// Save all pending changes (layout, title, pages)
const saveLayout = async () => {
  if (!GoApp || pendingLayoutChanges.value.length === 0) return
  try {
    // Get widgets with all pending changes merged (layout positions from pendingLayoutChanges,
    // title/pages already updated in widgets.value)
    const updatedWidgets = getWidgetsWithCurrentLayout()
    await GoApp.SaveWidgets(props.ctx, props.ns, props.projectName, activePage.value, updatedWidgets)
    widgets.value = updatedWidgets
    pendingLayoutChanges.value = []
  } catch (err) {
    emit('error', `Failed to save changes: ${err}`)
  }
}

// Toggle edit mode (save changes when exiting)
const toggleEditMode = async () => {
  if (editMode.value && pendingLayoutChanges.value.length > 0) {
    await saveLayout()
  }
  editMode.value = !editMode.value
}

// Reset layout changes (discard and reload)
const resetLayout = async () => {
  pendingLayoutChanges.value = []
  await loadWidgets()
  editMode.value = false
}

// Handle edit schema
const handleEditSchema = (widget) => {
  editingSchemaWidget.value = JSON.parse(JSON.stringify(widget))
  // Use customized schema if available, otherwise use defaultSchema
  const effectiveSchema = widget.schema && Object.keys(widget.schema).length > 0
    ? widget.schema
    : widget.defaultSchema

  // Resolve $ref if present to get the actual schema structure
  let schemaToEdit = effectiveSchema ? JSON.parse(JSON.stringify(effectiveSchema)) : { type: 'object' }
  if (schemaToEdit['$ref'] && schemaToEdit['$defs']) {
    const ref = schemaToEdit['$ref'].replace('#/$defs/', '')
    if (ref && schemaToEdit['$defs'][ref]) {
      schemaToEdit = schemaToEdit['$defs'][ref]
    }
  }

  schemaEditorValue.value = schemaToEdit

  // Auto-focus dialog for keyboard events
  nextTick(() => {
    schemaDialogRef.value?.focus()
  })
}

// Helper to get widgets with current layout positions merged
const getWidgetsWithCurrentLayout = () => {
  return widgets.value.map(w => {
    const change = pendingLayoutChanges.value.find(c => c.id === w.id)
    if (change) {
      return { ...w, ...change }
    }
    return w
  })
}

// Save schema changes - store locally, save on Done
const saveSchemaChanges = () => {
  if (!editingSchemaWidget.value) return

  // Get the original schema structure to preserve $ref and $defs
  const originalSchema = editingSchemaWidget.value.schema && Object.keys(editingSchemaWidget.value.schema).length > 0
    ? editingSchemaWidget.value.schema
    : editingSchemaWidget.value.defaultSchema

  let schemaToSave
  if (originalSchema && originalSchema['$ref'] && originalSchema['$defs']) {
    // Preserve the $ref structure and update the definition
    const ref = originalSchema['$ref'].replace('#/$defs/', '')
    schemaToSave = JSON.parse(JSON.stringify(originalSchema))
    if (ref && schemaToSave['$defs'][ref]) {
      schemaToSave['$defs'][ref] = schemaEditorValue.value
    }
  } else {
    schemaToSave = schemaEditorValue.value
  }

  // Update local widgets array - will be saved when user clicks Done
  const idx = widgets.value.findIndex(w => w.id === editingSchemaWidget.value.id)
  if (idx >= 0) {
    widgets.value[idx] = { ...widgets.value[idx], schema: schemaToSave }
  }

  // Mark as having changes
  if (!pendingLayoutChanges.value.find(c => c.id === editingSchemaWidget.value.id)) {
    pendingLayoutChanges.value.push({ id: editingSchemaWidget.value.id })
  }

  editingSchemaWidget.value = null
  schemaEditorValue.value = null
}

// Cancel schema editing
const cancelSchemaEdit = () => {
  editingSchemaWidget.value = null
  schemaEditorValue.value = null
}

// Handle reset schema request
const handleResetSchema = (widget) => {
  showResetConfirm.value = widget
}

// Confirm reset schema - store locally, save on Done
const confirmResetSchema = () => {
  if (!showResetConfirm.value) return

  // Update local widgets array - will be saved when user clicks Done
  const idx = widgets.value.findIndex(w => w.id === showResetConfirm.value.id)
  if (idx >= 0) {
    widgets.value[idx] = { ...widgets.value[idx], schema: null }
  }

  // Mark as having changes
  if (!pendingLayoutChanges.value.find(c => c.id === showResetConfirm.value.id)) {
    pendingLayoutChanges.value.push({ id: showResetConfirm.value.id })
  }

  showResetConfirm.value = null
}

// Cancel reset
const cancelReset = () => {
  showResetConfirm.value = null
}

// Handle title update - store locally, save on Done
const handleUpdateTitle = ({ widget, title }) => {
  // Update local widgets array - will be saved when user clicks Done
  const idx = widgets.value.findIndex(w => w.id === widget.id)
  if (idx >= 0) {
    widgets.value[idx] = { ...widgets.value[idx], title }
  }
  // Mark as having changes (use empty object in pendingLayoutChanges if not already tracking this widget)
  if (!pendingLayoutChanges.value.find(c => c.id === widget.id)) {
    pendingLayoutChanges.value.push({ id: widget.id })
  }
}

// Handle page assignment update - store locally, save on Done
const handleUpdatePages = ({ widget, pages: newPages }) => {
  // Update local widgets array - will be saved when user clicks Done
  const idx = widgets.value.findIndex(w => w.id === widget.id)
  if (idx >= 0) {
    widgets.value[idx] = { ...widgets.value[idx], pages: newPages }
  }
  // Mark as having changes
  if (!pendingLayoutChanges.value.find(c => c.id === widget.id)) {
    pendingLayoutChanges.value.push({ id: widget.id })
  }
}

// Calculate next available grid position for a new widget
const calculateNextGridPosition = () => {
  const existingWidgets = widgets.value
  if (existingWidgets.length === 0) {
    return { gridX: 0, gridY: 0 }
  }

  // Find the maximum Y position and widgets at that row
  let maxY = 0
  let widgetsAtMaxY = 0
  for (const w of existingWidgets) {
    if (w.gridY > maxY) {
      maxY = w.gridY
      widgetsAtMaxY = 1
    } else if (w.gridY === maxY) {
      widgetsAtMaxY++
    }
  }

  // If there's room in the current row (2 widgets per row in 6-col grid with w=3)
  if (widgetsAtMaxY < 2) {
    return { gridX: 3, gridY: maxY }
  }

  // Start a new row
  return { gridX: 0, gridY: maxY + 4 }
}

// Handle real-time node updates
const handleNodeUpdate = (update) => {
  if (!update.widget) return

  const idx = widgets.value.findIndex(w => w.nodeName === update.widget.nodeName)
  if (update.eventType === 'DELETED') {
    if (idx >= 0) {
      widgets.value.splice(idx, 1)
    }
  } else if (idx >= 0) {
    // Update existing widget - update schema and data from real-time, preserve positions
    const existing = widgets.value[idx]
    const newDefaultSchema = update.widget.defaultSchema || existing.defaultSchema

    const updatedWidget = {
      ...existing,
      // Always update schema from real-time updates (node schema reflects current state)
      // The node's schema includes dynamic changes like button titles
      defaultSchema: newDefaultSchema,
      schema: newDefaultSchema,  // Use node's schema directly - it has the current state
      data: update.widget.data || existing.data,
      // Preserve these fields (positions, page assignments, custom title)
      gridX: existing.gridX,
      gridY: existing.gridY,
      gridW: existing.gridW,
      gridH: existing.gridH,
      pages: existing.pages,
      title: existing.title,
      // Track update time for reactivity (forces JSONEditor to re-render)
      _updateTime: Date.now(),
    }
    // Use splice to ensure Vue detects the change
    widgets.value.splice(idx, 1, updatedWidget)
  } else {
    // Add new widget with calculated position and current page assignment
    const { gridX, gridY } = calculateNextGridPosition()
    const newWidget = {
      ...update.widget,
      gridX,
      gridY,
      gridW: update.widget.gridW || 3,
      gridH: update.widget.gridH || 4,
      pages: [activePage.value], // Assign to current page
      _updateTime: Date.now(),
    }
    widgets.value.push(newWidget)
  }
}

// Watch for page changes
watch(activePage, () => {
  loadWidgets()
})

onMounted(async () => {
  await loadPages()
  if (activePage.value) {
    await loadWidgets()
  }

  // Start watching for updates
  if (GoApp) {
    EventsOn('nodeUpdate', handleNodeUpdate)
    try {
      await GoApp.WatchProjectNodes(props.ctx, props.ns, props.projectName)
    } catch (err) {
      console.error('Failed to start node watcher:', err)
    }
  }
})

onUnmounted(async () => {
  EventsOff('nodeUpdate')
  if (GoApp) {
    try {
      await GoApp.StopWatchProjectNodes()
    } catch (err) {
      console.error('Failed to stop node watcher:', err)
    }
  }
})
</script>

<template>
  <div class="widgets-tab h-full flex flex-col">
    <div class="flex items-center justify-between">
      <PageTabs
        :pages="pages"
        :active-page="activePage"
        :edit-mode="editMode"
        @select-page="selectPage"
        @create-page="handleCreatePage"
        @delete-page="handleDeletePage"
      />
      <div class="flex items-center space-x-2 px-4 py-2">
        <button
          v-if="editMode"
          @click="resetLayout"
          class="flex items-center space-x-1 px-3 py-1.5 rounded-lg text-sm transition-colors text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
          title="Discard changes"
        >
          <ArrowUturnLeftIcon class="w-4 h-4" />
          <span>Reset</span>
        </button>
        <button
          @click="toggleEditMode"
          :class="[
            'flex items-center space-x-1 px-3 py-1.5 rounded-lg text-sm transition-colors',
            editMode
              ? 'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300'
              : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
          ]"
        >
          <PencilIcon v-if="!editMode" class="w-4 h-4" />
          <CheckIcon v-else class="w-4 h-4" />
          <span>{{ editMode ? 'Done' : 'Edit' }}</span>
        </button>
      </div>
    </div>
    <div class="flex-1 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-full">
        <span class="text-gray-500 dark:text-gray-400">Loading widgets...</span>
      </div>
      <WidgetGrid
        v-else
        :widgets="widgets"
        :edit-mode="editMode"
        :pages="pages"
        @action="handleAction"
        @layout-change="handleLayoutChange"
        @edit-schema="handleEditSchema"
        @reset-schema="handleResetSchema"
        @update-title="handleUpdateTitle"
        @update-pages="handleUpdatePages"
      />
    </div>

    <!-- Schema Editor Dialog -->
    <div
      v-if="editingSchemaWidget"
      ref="schemaDialogRef"
      tabindex="-1"
      class="fixed inset-0 z-50 overflow-y-auto outline-none"
      @keydown.enter.prevent="saveSchemaChanges"
      @keydown.escape="cancelSchemaEdit"
    >
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="cancelSchemaEdit"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-4xl max-h-[80vh] flex flex-col">
          <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
            <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
              Edit Widget Schema: {{ editingSchemaWidget.title || editingSchemaWidget.nodeName }}
            </h3>
            <button @click="cancelSchemaEdit" class="text-gray-400 hover:text-gray-500">
              <XMarkIcon class="w-6 h-6" />
            </button>
          </div>
          <div class="flex-1 overflow-auto p-4">
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">
              Edit the schema structure below. You can add, remove, or modify fields.
            </p>
            <JsonSchemaEditor
              v-if="schemaEditorValue"
              :value="{ root: schemaEditorValue }"
              :root="true"
              lang="en_US"
            />
          </div>
          <div class="flex justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
            <button
              @click="cancelSchemaEdit"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              @click="saveSchemaChanges"
              class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700"
              title="Enter"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Reset Confirm Dialog -->
    <div v-if="showResetConfirm" class="fixed inset-0 z-50 overflow-y-auto" @keydown.enter="confirmResetSchema" @keydown.escape="cancelReset">
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="cancelReset"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
            Reset Widget Schema?
          </h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
            This will reset the widget schema to its default. Any customizations will be lost.
          </p>
          <div class="flex justify-end gap-2">
            <button
              @click="cancelReset"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              @click="confirmResetSchema"
              class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700"
            >
              Reset
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
