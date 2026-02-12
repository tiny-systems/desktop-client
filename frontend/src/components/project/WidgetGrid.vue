<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick, createApp, h } from 'vue'
import { GridStack } from 'gridstack'
import 'gridstack/dist/gridstack.min.css'
import Widget from './Widget.vue'

const props = defineProps({
  widgets: {
    type: Array,
    default: () => []
  },
  editMode: {
    type: Boolean,
    default: false
  },
  pages: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['action', 'layoutChange', 'edit-schema', 'reset-schema', 'update-title', 'update-pages', 'update-content'])

let grid = null
const gridElement = ref(null)
const widgetInstances = new Map()
let isSyncing = false  // Flag to prevent change events during sync

onMounted(() => {
  initGrid()
})

onUnmounted(() => {
  // Clean up widget instances
  widgetInstances.forEach((instance, id) => {
    if (instance.app) {
      instance.app.unmount()
    }
  })
  widgetInstances.clear()

  if (grid) {
    grid.destroy(false)
    grid = null
  }
})

// Watch for widget changes, but skip initial call since initGrid handles that
let initialSyncDone = false
watch(() => props.widgets, (newWidgets) => {
  if (grid && initialSyncDone) {
    syncWidgets(newWidgets)
  }
}, { deep: true })

watch(() => props.editMode, (newEditMode) => {
  if (grid) {
    grid.setStatic(!newEditMode)
  }
})

const initGrid = () => {
  if (!gridElement.value) return

  // Initialize empty GridStack
  grid = GridStack.init({
    column: 6,
    cellHeight: 80,
    margin: 6,
    disableOneColumnMode: true,
    float: true,
    staticGrid: !props.editMode,
  }, gridElement.value)

  grid.on('change', (event, items) => {
    // Skip change events during programmatic sync
    if (isSyncing) return
    if (props.editMode && items) {
      // Get ALL grid items, not just the changed ones
      // This ensures widgets moved as side-effects are also saved
      const allItems = grid.getGridItems()
      const layoutChanges = allItems.map(el => {
        const node = el.gridstackNode
        return {
          id: node.id,
          gridX: node.x,
          gridY: node.y,
          gridW: node.w,
          gridH: node.h
        }
      })
      emit('layoutChange', layoutChanges)
    }
  })

  // Add initial widgets
  syncWidgets(props.widgets)
  initialSyncDone = true
}

const syncWidgets = (widgets) => {
  if (!grid) return

  // Set flag to prevent change events during sync
  isSyncing = true

  const currentIds = new Set(widgets.map(w => w.id))

  // Remove widgets that no longer exist
  widgetInstances.forEach((instance, id) => {
    if (!currentIds.has(id)) {
      grid.removeWidget(instance.element)
      if (instance.app) {
        instance.app.unmount()
      }
      widgetInstances.delete(id)
    }
  })

  // Batch update to prevent multiple change events
  grid.batchUpdate()

  // Add or update widgets
  widgets.forEach(widget => {
    if (!widgetInstances.has(widget.id)) {
      addWidget(widget)
    } else {
      // Update existing widget data and position
      const instance = widgetInstances.get(widget.id)
      updateWidgetContent(instance, widget)
      // Also sync GridStack position
      grid.update(instance.element, {
        x: widget.gridX,
        y: widget.gridY,
        w: widget.gridW || 3,
        h: widget.gridH || 4
      })
    }
  })

  grid.batchUpdate(false)
  isSyncing = false
}

const addWidget = (widget) => {
  // Create container for Vue component
  const container = document.createElement('div')
  container.style.height = '100%'
  container.style.width = '100%'

  // Create a Vue app for each widget to ensure proper reactivity
  const app = createApp({
    render() {
      return h(Widget, {
        widget: widget,
        readonly: false,  // Always allow interaction with widget controls
        editMode: props.editMode,
        pages: props.pages,
        onAction: handleAction,
        onEditSchema: handleEditSchema,
        onResetSchema: handleResetSchema,
        onUpdateTitle: handleUpdateTitle,
        onUpdatePages: handleUpdatePages,
        onUpdateContent: handleUpdateContent
      })
    }
  })
  app.mount(container)

  // Add to GridStack
  const element = grid.addWidget({
    id: widget.id,
    x: widget.gridX,
    y: widget.gridY,
    w: widget.gridW || 3,
    h: widget.gridH || 4,
    content: ''
  })

  // Put our Vue component inside the grid-stack-item-content
  const contentEl = element.querySelector('.grid-stack-item-content')
  if (contentEl) {
    contentEl.appendChild(container)
  }

  widgetInstances.set(widget.id, { element, container, app })
}

const updateWidgetContent = (instance, widget) => {
  // Unmount old app and create new one with updated props
  if (instance.app) {
    instance.app.unmount()
  }

  const app = createApp({
    render() {
      return h(Widget, {
        widget: widget,
        readonly: false,  // Always allow interaction with widget controls
        editMode: props.editMode,
        pages: props.pages,
        onAction: handleAction,
        onEditSchema: handleEditSchema,
        onResetSchema: handleResetSchema,
        onUpdateTitle: handleUpdateTitle,
        onUpdatePages: handleUpdatePages,
        onUpdateContent: handleUpdateContent
      })
    }
  })
  app.mount(instance.container)
  instance.app = app
}

const handleAction = (actionData) => {
  emit('action', actionData)
}

const handleEditSchema = (widget) => {
  emit('edit-schema', widget)
}

const handleResetSchema = (widget) => {
  emit('reset-schema', widget)
}

const handleUpdatePages = (data) => {
  emit('update-pages', data)
}

const handleUpdateContent = (data) => {
  emit('update-content', data)
}

const handleUpdateTitle = (data) => {
  emit('update-title', data)
}
</script>

<template>
  <div class="widget-grid-container h-full overflow-auto p-4">
    <div ref="gridElement" class="grid-stack"></div>
    <div v-if="widgets.length === 0" class="flex items-center justify-center h-64 text-gray-400 dark:text-gray-500">
      No widgets on this page
    </div>
  </div>
</template>

<style>
.widget-grid-container {
  min-height: 400px;
}

.grid-stack {
  background: transparent;
}

.grid-stack-item-content {
  overflow: hidden;
  border-radius: 8px;
}
</style>
