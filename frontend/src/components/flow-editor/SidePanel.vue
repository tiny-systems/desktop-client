<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useFlowStore } from '../../stores/flow'
import {
  Menu,
  MenuButton,
  MenuItem,
  MenuItems
} from '@headlessui/vue'
import { EllipsisVerticalIcon } from '@heroicons/vue/24/solid'
import {
  PencilIcon,
  TrashIcon,
  XMarkIcon,
  Cog6ToothIcon,
  ArrowPathIcon,
  ClipboardDocumentIcon,
  ClipboardDocumentCheckIcon,
  Square3Stack3DIcon
} from '@heroicons/vue/24/outline'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'
import SchemaForm from './SchemaForm.vue'
import FlowDataLookupModal from './FlowDataLookupModal.vue'
import PortTabs from './PortTabs.vue'
import FlowTransferModal from './FlowTransferModal.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import UnsavedChangesDialog from './UnsavedChangesDialog.vue'

const emit = defineEmits(['close', 'error', 'rename', 'settings', 'delete'])

const flowStore = useFlowStore()

// Unsaved changes tracking
const originalNodeConfigValue = ref(null)
const originalEdgeConfigValue = ref(null)
const nodeConfigDirty = ref(false)
const edgeConfigDirty = ref(false)
const showUnsavedChangesDialog = ref(false)
const pendingSelectionAction = ref(null)
const isRestoringSelection = ref(false)
const configurationResetKey = ref(0) // Incremented on discard to force SchemaForm remount
// Store the dirty element reference when dialog shows (because selection might not restore properly)
const pendingDirtyNode = ref(null)
const pendingDirtyEdge = ref(null)
// Also store the form values at time of showing dialog (watchers might reset them)
const pendingNodeFormValue = ref(null)
const pendingEdgeFormValue = ref(null)

// Safety: ensure isRestoringSelection is cleared when dialog closes
watch(showUnsavedChangesDialog, (show) => {
  if (!show) {
    // Dialog closed - ensure flag is cleared
    isRestoringSelection.value = false
  }
})

// Helper to decode HTML entities
const decodeHtmlEntities = (text) => {
  if (!text) return ''
  const textarea = document.createElement('textarea')
  textarea.innerHTML = text
  return textarea.value
}

// Dynamic width class for side panel
const hasSelection = computed(() => flowStore.selectedNode || flowStore.selectedEdge || flowStore.selectedNodes.length > 0)
const panelWidthClass = computed(() => {
  if (flowStore.selectedNodes.length > 1) return 'w-1/4' // Multi-node selection - smaller panel
  if (flowStore.selectedEdge) return 'w-2/3' // Edge configuration - wider panel
  if (hasSelection.value) return 'w-1/2' // Single node selection
  return 'w-1/5 min-w-[300px]'
})

// Tab state
const statusTab = ref({ id: 'status', name: '', current: true })
const configurationTab = ref({ id: 'configuration', name: 'Configuration', current: false })

const setCurrentTab = (tabId) => {
  statusTab.value.current = tabId === 'status'
  configurationTab.value.current = tabId === 'configuration'
}

// Computed selection state
const selectedNode = computed(() => flowStore.selectedNode)
const selectedEdge = computed(() => flowStore.selectedEdge)
const selectedNodes = computed(() => flowStore.selectedNodes)

// Multi-node selection actions
const showDeleteMultipleModal = ref(false)
const showTransferModal = ref(false)

const deleteMultipleNodes = async () => {
  flowStore.deleteSelected()
  await flowStore.save()
  showDeleteMultipleModal.value = false
}

const onTransferNodes = () => {
  // Check if any selected nodes are blocked (shared from other flows)
  const hasBlockedNodes = flowStore.selectedNodes.some(n => n.data?.blocked)
  if (hasBlockedNodes) {
    emit('error', 'Some selected nodes are shared from other flows and cannot be transferred')
    return
  }
  showTransferModal.value = true
}

const onTransferSingleNode = () => {
  // For single node, check if it's blocked
  if (selectedNode.value?.data?.blocked) {
    emit('error', 'This node is shared from another flow and cannot be transferred')
    return
  }
  showTransferModal.value = true
}

const clearSelection = () => {
  if (nodeConfigDirty.value || edgeConfigDirty.value) {
    pendingSelectionAction.value = () => flowStore.selectElement(null)
    showUnsavedChangesDialog.value = true
    return
  }
  flowStore.selectElement(null)
}

const deselectNode = (nodeId) => {
  const node = flowStore.getElement(nodeId)
  if (node) {
    node.selected = false
  }
}

// Port inspection
const inspect = ref(null)
const inspectReady = ref(false)
const selectedHandleId = ref(null)
const inspectCopied = ref(false)

const copyInspectToClipboard = async () => {
  if (!inspect.value?.data) return
  try {
    await navigator.clipboard.writeText(JSON.stringify(inspect.value.data, null, 2))
    inspectCopied.value = true
    setTimeout(() => {
      inspectCopied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}

// Get all handles for port tabs
const selectedNodeHandles = computed(() => {
  return selectedNode.value?.data?.handles || []
})

// Get visible handles (exclude those starting with underscore)
const visibleNodeHandles = computed(() => {
  return selectedNodeHandles.value.filter(h => h.id && !h.id.startsWith('_'))
})

// Get selected handle (only from visible handles)
const selectedHandle = computed(() => {
  if (!visibleNodeHandles.value.length) return null
  const sel = selectedHandleId.value || selectedNode.value?.data?.trace?.port
  // Only return handle if it's in visible handles
  const found = visibleNodeHandles.value.find(h => h.id === sel)
  return found || visibleNodeHandles.value[0]
})

// Settings handle
const settingsHandle = computed(() => {
  const handles = selectedNode.value?.data?.handles || []
  return handles.find(h => h.id === '_settings')
})

// Control handle
const controlHandle = computed(() => {
  const handles = selectedNode.value?.data?.handles || []
  return handles.find(h => h.id === '_control')
})

// Control handle schema for the form
const controlHandleSchema = computed(() => {
  if (!controlHandle.value?.schema) return null
  const schema = controlHandle.value.schema
  if (typeof schema === 'string') {
    try {
      return JSON.parse(schema)
    } catch {
      return null
    }
  }
  return schema
})

// Control handle configuration
const controlConfigObject = computed(() => {
  const config = controlHandle.value?.configuration
  if (!config) return {}
  if (typeof config === 'string') {
    try {
      return JSON.parse(config)
    } catch {
      return {}
    }
  }
  return config
})

// Control form values
const controlFormValue = ref({})

// Watch control config changes
watch(controlConfigObject, (val) => {
  controlFormValue.value = { ...val }
}, { immediate: true, deep: true })

// Handle control form update (send action to node)
const handleControlUpdate = async (event) => {
  // Only handle action button clicks
  if (!event?.isAction) return
  if (!selectedNode.value) return

  try {
    await flowStore.runAction(selectedNode.value.id, '_control', event.value || event)
  } catch (err) {
    emit('error', `Action failed: ${err}`)
  }
}

// Update control form value handler
const updateControlFormValue = (newValue) => {
  controlFormValue.value = newValue
}

// Configuration for settings
const editorValue = ref('{}')
const formValue = ref({})
const saving = ref(false)
const configurationReady = ref(false)

const settingsConfiguration = computed(() => {
  if (!settingsHandle.value?.configuration) return '{}'
  const config = settingsHandle.value.configuration
  return typeof config === 'string' ? config : JSON.stringify(config, null, 2)
})

// Helper to parse schema from handle
const parseSchemaFromHandle = () => {
  if (!settingsHandle.value?.schema) return null
  const schema = settingsHandle.value.schema
  if (typeof schema === 'string') {
    try {
      return JSON.parse(schema)
    } catch {
      return null
    }
  }
  // Deep clone to avoid mutating the original
  return JSON.parse(JSON.stringify(schema))
}

// Editable settings schema (local copy that can be mutated by the form)
const editableSchema = ref(null)

// Settings schema for the form - uses the editable local copy
const settingsSchema = computed(() => editableSchema.value)

// Watch for schema changes from the handle and update local copy
// Use stringified schema as watch source to avoid false triggers from Vue reactivity
// Preserve the 'configure' flag which is UI-only state
watch(() => {
  const schema = settingsHandle.value?.schema
  return typeof schema === 'string' ? schema : JSON.stringify(schema || null)
}, (newSchemaStr, oldSchemaStr) => {
  // Skip if content is identical (watcher might fire due to reactivity even with same content)
  if (newSchemaStr === oldSchemaStr && editableSchema.value !== null) {
    return
  }

  const newParsed = parseSchemaFromHandle()

  // If we have an existing schema with configure=true, preserve it unless schema actually changed
  if (editableSchema.value?.configure) {
    // Compare without the configure flag
    const existingWithoutConfigure = { ...editableSchema.value }
    delete existingWithoutConfigure.configure

    const newWithoutConfigure = newParsed ? { ...newParsed } : null
    if (newWithoutConfigure) delete newWithoutConfigure.configure

    // Only update if content actually changed
    if (JSON.stringify(existingWithoutConfigure) === JSON.stringify(newWithoutConfigure)) {
      return // Keep existing schema with configure flag
    }

    // Schema changed but preserve configure flag
    if (newParsed) {
      newParsed.configure = true
    }
  }

  editableSchema.value = newParsed
}, { immediate: true })

// Parse settings configuration into object for form
const settingsConfigObject = computed(() => {
  const config = settingsHandle.value?.configuration
  if (!config) return {}
  if (typeof config === 'string') {
    try {
      return JSON.parse(config)
    } catch {
      return {}
    }
  }
  return config
})

// Watch settings changes for raw editor - only when not dirty
watch(settingsConfiguration, (val) => {
  if (!nodeConfigDirty.value) {
    editorValue.value = val
  }
}, { immediate: true })

// Watch settings changes for form - only when not dirty
watch(settingsConfigObject, (val, oldVal) => {
  if (!nodeConfigDirty.value) {
    formValue.value = { ...val }
  }
  configurationReady.value = true
}, { immediate: true, deep: true })

// Update form value handler
const updateFormValue = (newValue) => {
  formValue.value = newValue
  editorValue.value = JSON.stringify(newValue, null, 2)
  // Track dirty state
  nodeConfigDirty.value = JSON.stringify(newValue) !== JSON.stringify(originalNodeConfigValue.value)
}

// Watch node and handle for port inspection
watch(
  [() => selectedNode.value?.id, () => selectedHandle.value?.id],
  async ([nodeId, handleId]) => {
    if (!handleId || !nodeId) {
      inspect.value = null
      inspectReady.value = true
      return
    }

    inspectReady.value = false
    try {
      const data = await flowStore.inspectNodePort(nodeId, handleId)
      // Only update if still on same selection
      if (selectedNode.value?.id === nodeId && selectedHandle.value?.id === handleId) {
        inspect.value = data
      }
    } catch (e) {
      if (selectedNode.value?.id === nodeId && selectedHandle.value?.id === handleId) {
        inspect.value = { error: e.message || String(e) }
      }
    } finally {
      if (selectedNode.value?.id === nodeId && selectedHandle.value?.id === handleId) {
        inspectReady.value = true
      }
    }
  },
  { immediate: true }
)

// Watch node change to reset tab, selected handle, and trigger re-fetch
watch(() => selectedNode.value?.id, (newId, oldId) => {
  if (newId !== oldId) {
    setCurrentTab('status')
    selectedHandleId.value = null
    // Only update form value if not restoring selection AND not dirty (dialog flow preserves form value)
    if (!isRestoringSelection.value && !nodeConfigDirty.value) {
      formValue.value = { ...settingsConfigObject.value }
    }
    configurationReady.value = true
  }
}, { immediate: true })

// Node info expiring check
const selectedNodeExpiring = computed(() => {
  if (!selectedNode.value?.data?.last_status_update) return false
  // last_status_update is Unix timestamp in seconds
  const lastUpdate = selectedNode.value.data.last_status_update * 1000
  return (Date.now() - lastUpdate) > 10 * 60 * 1000
})

// Time ago formatter
const timeAgo = (timestamp) => {
  if (!timestamp) return 'Never'
  const seconds = Math.floor((Date.now() - timestamp * 1000) / 1000)

  if (seconds < 5) return 'just now'
  if (seconds < 60) return `${seconds} seconds ago`

  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return minutes === 1 ? '1 minute ago' : `${minutes} minutes ago`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return hours === 1 ? '1 hour ago' : `${hours} hours ago`

  const days = Math.floor(hours / 24)
  if (days < 30) return days === 1 ? '1 day ago' : `${days} days ago`

  const months = Math.floor(days / 30)
  if (months < 12) return months === 1 ? '1 month ago' : `${months} months ago`

  const years = Math.floor(months / 12)
  return years === 1 ? '1 year ago' : `${years} years ago`
}

// Edge info
const edgeSourceLabel = computed(() => {
  if (!selectedEdge.value) return ''
  // Try to find source node label
  const sourceNode = flowStore.getElement(selectedEdge.value.source)
  return sourceNode?.data?.label || selectedEdge.value.source
})

const edgeTargetLabel = computed(() => {
  if (!selectedEdge.value) return ''
  const targetNode = flowStore.getElement(selectedEdge.value.target)
  return targetNode?.data?.label || selectedEdge.value.target
})

// Edge configuration - get from edge data or fall back to target handle
const edgeConfiguration = computed(() => {
  if (!selectedEdge.value) return '{}'

  // First try edge data configuration
  let config = selectedEdge.value.data?.configuration

  // Fall back to target handle's configuration
  if (!config) {
    const targetNode = flowStore.getElement(selectedEdge.value.target)
    if (targetNode?.data?.handles) {
      const targetHandle = targetNode.data.handles.find(h => h.id === selectedEdge.value.targetHandle)
      config = targetHandle?.configuration
    }
  }

  if (!config) return '{}'
  return typeof config === 'string' ? config : JSON.stringify(config, null, 2)
})

// Edge configuration as object for form
const edgeConfigObject = computed(() => {
  if (!selectedEdge.value) return {}

  // First try edge data configuration
  let config = selectedEdge.value.data?.configuration

  // Fall back to target handle's configuration
  if (!config) {
    const targetNode = flowStore.getElement(selectedEdge.value.target)
    if (targetNode?.data?.handles) {
      const targetHandle = targetNode.data.handles.find(h => h.id === selectedEdge.value.targetHandle)
      config = targetHandle?.configuration
    }
  }

  if (!config) return {}
  if (typeof config === 'string') {
    try {
      return JSON.parse(config)
    } catch {
      return {}
    }
  }
  return config
})

// Edge schema - get from target port
const edgeSchema = computed(() => {
  if (!selectedEdge.value) return null
  // Try to get schema from edge data first
  if (selectedEdge.value.data?.schema) {
    const schema = selectedEdge.value.data.schema
    if (typeof schema === 'string') {
      try {
        return JSON.parse(schema)
      } catch {
        return null
      }
    }
    return schema
  }
  // Fall back to getting schema from target node's target port
  const targetNode = flowStore.getElement(selectedEdge.value.target)
  if (!targetNode?.data?.handles) return null
  const targetHandle = targetNode.data.handles.find(h => h.id === selectedEdge.value.targetHandle)
  if (!targetHandle?.schema) return null
  const schema = targetHandle.schema
  if (typeof schema === 'string') {
    try {
      return JSON.parse(schema)
    } catch {
      return null
    }
  }
  return schema
})

// Edge validation errors
const edgeValidationError = computed(() => selectedEdge.value?.data?.error || null)
const edgeValidationErrors = computed(() => {
  const errors = selectedEdge.value?.data?.errors
  if (!errors || typeof errors !== 'object') return []
  return Object.entries(errors).map(([path, message]) => ({
    path,
    message
  }))
})
const edgeIsValid = computed(() => selectedEdge.value?.data?.valid !== false)

const edgeEditorValue = ref('{}')
const edgeFormValue = ref({})

// Edge source data (left column) and preview result (right column)
const edgeSourceData = ref({})
const edgeSourceSchema = ref({})
const edgeSourceLoading = ref(false)
const edgePreviewResult = ref({})
const edgePreviewLoading = ref(false)
const edgePreviewError = ref('')

// Fetch source port data when edge is selected
watch(() => selectedEdge.value?.id, async (edgeId) => {
  if (!edgeId || !selectedEdge.value) {
    edgeSourceData.value = {}
    edgeSourceSchema.value = {}
    edgePreviewResult.value = {}
    return
  }

  edgeSourceLoading.value = true
  try {
    const data = await flowStore.inspectNodePort(selectedEdge.value.source, selectedEdge.value.sourceHandle)
    edgeSourceData.value = data?.data || {}

    // Get source schema from handle
    const sourceNode = flowStore.getElement(selectedEdge.value.source)
    if (sourceNode?.data?.handles) {
      const sourceHandle = sourceNode.data.handles.find(h => h.id === selectedEdge.value.sourceHandle)
      if (sourceHandle?.schema) {
        edgeSourceSchema.value = typeof sourceHandle.schema === 'string'
          ? JSON.parse(sourceHandle.schema)
          : sourceHandle.schema
      }
    }
  } catch (e) {
    console.error('Failed to fetch edge source data:', e)
    edgeSourceData.value = {}
  } finally {
    edgeSourceLoading.value = false
  }
}, { immediate: true })

// Compute preview result when source data or configuration changes
const computeEdgePreview = async () => {
  if (!selectedEdge.value || !edgeSourceData.value || Object.keys(edgeSourceData.value).length === 0) {
    edgePreviewResult.value = {}
    return
  }

  edgePreviewLoading.value = true
  edgePreviewError.value = ''

  try {
    // Use previewEdgeMapping to apply the edge configuration mapping
    const result = await flowStore.previewEdgeMapping(
      edgeEditorValue.value,
      edgeSourceData.value
    )
    if (result?.result) {
      edgePreviewResult.value = JSON.parse(result.result)
    } else {
      edgePreviewResult.value = {}
    }
    // Show errors if any
    if (result?.errors && result.errors.length > 0) {
      edgePreviewError.value = result.errors.join('; ')
    }
  } catch (e) {
    edgePreviewError.value = e.message || 'Preview failed'
    edgePreviewResult.value = {}
  } finally {
    edgePreviewLoading.value = false
  }
}

// Debounced preview computation
let previewTimeout = null
watch([edgeEditorValue, edgeSourceData], () => {
  if (previewTimeout) clearTimeout(previewTimeout)
  previewTimeout = setTimeout(computeEdgePreview, 300)
}, { deep: true })

// Watch edge config changes - only when not dirty
watch(edgeConfiguration, (val) => {
  if (!edgeConfigDirty.value) {
    edgeEditorValue.value = val
  }
}, { immediate: true })

watch(edgeConfigObject, (val) => {
  if (!edgeConfigDirty.value) {
    edgeFormValue.value = { ...val }
  }
}, { immediate: true, deep: true })

// Update edge form value handler
const updateEdgeFormValue = (newValue) => {
  edgeFormValue.value = newValue
  edgeEditorValue.value = JSON.stringify(newValue, null, 2)
  // Track dirty state
  edgeConfigDirty.value = JSON.stringify(newValue) !== JSON.stringify(originalEdgeConfigValue.value)
}

// ==========================================
// Unsaved Changes Tracking (platform-style)
// ==========================================

// Deep copy helper
const deepCopy = (obj) => JSON.parse(JSON.stringify(obj))

// Watch selected node for unsaved changes
watch(selectedNode, (newNode, oldNode) => {
  if (isRestoringSelection.value) {
    return
  }

  const newId = newNode?.id || null
  const oldId = oldNode?.id || null

  // If selection changed and we have unsaved changes
  if (newId !== oldId && nodeConfigDirty.value && !flowStore.readOnly && oldId !== null) {
    const targetId = newId
    isRestoringSelection.value = true

    // Store deep copy of the dirty node AND form value (watchers might clear them)
    pendingDirtyNode.value = deepCopy(oldNode)
    pendingNodeFormValue.value = deepCopy(formValue.value)
    pendingDirtyEdge.value = null
    pendingEdgeFormValue.value = null

    // Restore old selection using flowStore
    flowStore.selectElement(oldId)

    // Set pending action (isRestoringSelection will be cleared in dialog handlers)
    pendingSelectionAction.value = () => {
      isRestoringSelection.value = true
      if (targetId) {
        flowStore.selectElement(targetId)
      } else {
        flowStore.selectElement(null)
      }
    }
    showUnsavedChangesDialog.value = true
  }
})

// Watch selected edge for unsaved changes
watch(() => flowStore.selectedEdge, (newEdge, oldEdge) => {
  if (isRestoringSelection.value) {
    return
  }

  const newId = newEdge?.id || null
  const oldId = oldEdge?.id || null

  // If selection changed and we have unsaved changes
  if (newId !== oldId && edgeConfigDirty.value && !flowStore.readOnly && oldId !== null) {
    const targetId = newId
    isRestoringSelection.value = true

    // Store deep copy of the dirty edge AND form value (watchers might clear them)
    console.log('Storing dirty edge:', oldEdge)
    console.log('Storing edge form value:', edgeFormValue.value)
    pendingDirtyEdge.value = deepCopy(oldEdge)
    pendingEdgeFormValue.value = deepCopy(edgeFormValue.value)
    console.log('Stored pendingDirtyEdge:', pendingDirtyEdge.value)
    console.log('Stored pendingEdgeFormValue:', pendingEdgeFormValue.value)
    pendingDirtyNode.value = null
    pendingNodeFormValue.value = null

    // Restore old selection using flowStore
    flowStore.selectElement(oldId)

    // Set pending action (isRestoringSelection will be cleared in dialog handlers)
    pendingSelectionAction.value = () => {
      isRestoringSelection.value = true
      if (targetId) {
        flowStore.selectElement(targetId)
      } else {
        flowStore.selectElement(null)
      }
    }
    showUnsavedChangesDialog.value = true
  }
})

// Save original node config when node selection changes
watch(() => selectedNode.value?.id, (newId, oldId) => {
  // Skip if we're restoring selection after cancel
  if (isRestoringSelection.value) {
    return
  }
  if (newId && newId !== oldId) {
    // New node selected - save original config
    originalNodeConfigValue.value = deepCopy(settingsConfigObject.value)
    nodeConfigDirty.value = false
  }
}, { immediate: true })

// Save original edge config when edge selection changes
watch(() => selectedEdge.value?.id, (newId, oldId) => {
  // Skip if we're restoring selection after cancel
  if (isRestoringSelection.value) {
    return
  }
  if (newId && newId !== oldId) {
    // New edge selected - save original config
    originalEdgeConfigValue.value = deepCopy(edgeConfigObject.value)
    edgeConfigDirty.value = false
  }
}, { immediate: true })

// Handle unsaved changes dialog actions
const handleUnsavedCancel = () => {
  showUnsavedChangesDialog.value = false
  pendingSelectionAction.value = null
  pendingDirtyNode.value = null
  pendingDirtyEdge.value = null
  pendingNodeFormValue.value = null
  pendingEdgeFormValue.value = null
  // Clear restoring flag since we're staying on current selection
  nextTick(() => {
    isRestoringSelection.value = false
  })
}

const handleUnsavedDiscard = () => {
  // Revert form values to original
  if (originalNodeConfigValue.value !== null) {
    formValue.value = deepCopy(originalNodeConfigValue.value)
    editorValue.value = JSON.stringify(formValue.value, null, 2)
  }
  if (originalEdgeConfigValue.value !== null) {
    edgeFormValue.value = deepCopy(originalEdgeConfigValue.value)
    edgeEditorValue.value = JSON.stringify(edgeFormValue.value, null, 2)
  }
  // Increment reset key to force SchemaForm remount
  configurationResetKey.value++
  nodeConfigDirty.value = false
  edgeConfigDirty.value = false
  pendingDirtyNode.value = null
  pendingDirtyEdge.value = null
  pendingNodeFormValue.value = null
  pendingEdgeFormValue.value = null
  showUnsavedChangesDialog.value = false
  if (pendingSelectionAction.value) {
    pendingSelectionAction.value()
    pendingSelectionAction.value = null
  }
  // Clear restoring flag after pending action completes
  nextTick(() => {
    isRestoringSelection.value = false
  })
}

const handleUnsavedSave = async () => {
  console.log('handleUnsavedSave called')
  console.log('pendingDirtyNode:', pendingDirtyNode.value)
  console.log('pendingNodeFormValue:', pendingNodeFormValue.value)
  console.log('pendingDirtyEdge:', pendingDirtyEdge.value)
  console.log('pendingEdgeFormValue:', pendingEdgeFormValue.value)

  // Use stored dirty references AND form values - selection/watchers might have cleared them
  if (pendingDirtyNode.value && pendingNodeFormValue.value) {
    console.log('Saving node...')
    // Save node configuration using the stored reference and form value
    const node = pendingDirtyNode.value
    const configToSave = pendingNodeFormValue.value
    const handles = node?.data?.handles || []
    const settings = handles.find(h => h.id === '_settings')
    if (settings) {
      saving.value = true
      try {
        const schemaStr = settingsSchema.value ? JSON.stringify(settingsSchema.value) : ''
        const configStr = JSON.stringify(configToSave)
        console.log('Calling updateNodeConfiguration:', { nodeId: node.id, configStr })
        await flowStore.updateNodeConfiguration(
          node.id,
          '_settings',
          configStr,
          schemaStr
        )
        console.log('Node save completed successfully')

        // Update local node data in store so it reflects the new config
        const localNode = flowStore.getElement(node.id)
        if (localNode && localNode.data?.handles) {
          const settingsIdx = localNode.data.handles.findIndex(h => h.id === '_settings')
          if (settingsIdx !== -1) {
            localNode.data.handles[settingsIdx].configuration = configToSave
            console.log('Updated local node data')
          }
        }

        originalNodeConfigValue.value = deepCopy(configToSave)
        nodeConfigDirty.value = false
      } catch (err) {
        console.error('Node save failed:', err)
        emit('error', `Failed to save: ${err}`)
      } finally {
        saving.value = false
      }
    } else {
      console.log('No _settings handle found on node')
    }
  } else {
    console.log('Node save skipped - pendingDirtyNode:', !!pendingDirtyNode.value, 'pendingNodeFormValue:', !!pendingNodeFormValue.value)
  }
  if (pendingDirtyEdge.value && pendingEdgeFormValue.value) {
    console.log('Saving edge...')
    // Save edge configuration using the stored reference and form value
    const edge = pendingDirtyEdge.value
    const configToSave = pendingEdgeFormValue.value
    // Validate edge has required properties
    if (!edge.source || !edge.sourceHandle || !edge.target || !edge.targetHandle) {
      console.log('Edge validation failed:', edge)
      emit('error', 'Cannot save: edge data is incomplete')
    } else {
      saving.value = true
      try {
        const targetTo = `${edge.target}:${edge.targetHandle}`
        const configStr = JSON.stringify(configToSave)
        console.log('Calling updateEdgeConfiguration:', {
          source: edge.source,
          sourceHandle: edge.sourceHandle,
          targetTo,
          configStr,
          flowID: edge.data?.flowID || flowStore.flowResourceName
        })
        await flowStore.updateEdgeConfiguration(
          edge.source,
          edge.sourceHandle,
          targetTo,
          configStr,
          edge.data?.flowID || flowStore.flowResourceName
        )
        console.log('Edge save completed successfully')

        // Update local edge data in store so it reflects the new config
        const localEdge = flowStore.getElement(edge.id)
        if (localEdge && localEdge.data) {
          localEdge.data.configuration = configToSave
          console.log('Updated local edge data')
        }

        originalEdgeConfigValue.value = deepCopy(configToSave)
        edgeConfigDirty.value = false
      } catch (err) {
        console.error('Edge save failed:', err, err?.stack)
        emit('error', `Failed to save edge: ${err?.message || err}`)
      } finally {
        saving.value = false
      }
    }
  } else {
    console.log('Edge save skipped - pendingDirtyEdge:', !!pendingDirtyEdge.value, 'pendingEdgeFormValue:', !!pendingEdgeFormValue.value)
  }
  // Clear stored references
  pendingDirtyNode.value = null
  pendingDirtyEdge.value = null
  pendingNodeFormValue.value = null
  pendingEdgeFormValue.value = null
  showUnsavedChangesDialog.value = false

  // Small delay to let backend sync before navigation
  await new Promise(resolve => setTimeout(resolve, 100))

  if (pendingSelectionAction.value) {
    console.log('Executing pending selection action')
    pendingSelectionAction.value()
    pendingSelectionAction.value = null
  }
  // Clear restoring flag after pending action completes
  nextTick(() => {
    isRestoringSelection.value = false
  })
}

// Lookup modal state
const showLookupModal = ref(false)
const lookupCallback = ref(null)
const lookupTargetSchema = ref({})
const lookupFullSchema = ref({})
const lookupInitialExpression = ref('')
const lookupPortName = ref('')
const lookupSourceData = ref({})
const lookupSourceSchema = ref({})
const lookupLoading = ref(false)
const lookupFieldTitle = ref('')

// Handle lookup event from schema editor
const handleEdgeLookup = async (currentValue, schema, callback) => {
  if (!selectedEdge.value) return

  lookupTargetSchema.value = schema || {}
  lookupFullSchema.value = edgeSchema.value || {}
  lookupCallback.value = callback
  lookupLoading.value = true
  lookupFieldTitle.value = schema?.title || 'Context'

  // Extract existing expression if present
  if (typeof currentValue === 'string' && currentValue.startsWith('{{') && currentValue.endsWith('}}')) {
    lookupInitialExpression.value = currentValue.slice(2, -2)
  } else {
    lookupInitialExpression.value = '$'
  }
  lookupPortName.value = selectedEdge.value?.sourceHandle || ''

  // Get source node and port info
  const sourceNode = flowStore.getElement(selectedEdge.value.source)
  const sourceHandleId = selectedEdge.value.sourceHandle

  // Get source schema from handle
  if (sourceNode?.data?.handles) {
    const sourceHandle = sourceNode.data.handles.find(h => h.id === sourceHandleId)
    if (sourceHandle?.schema) {
      try {
        lookupSourceSchema.value = typeof sourceHandle.schema === 'string'
          ? JSON.parse(sourceHandle.schema)
          : sourceHandle.schema
      } catch {
        lookupSourceSchema.value = {}
      }
    }
  }

  // Fetch actual port data via inspection
  try {
    const inspectData = await flowStore.inspectNodePort(selectedEdge.value.source, sourceHandleId)
    if (inspectData?.data) {
      lookupSourceData.value = inspectData.data
    } else {
      // Fallback to schema default or empty object
      lookupSourceData.value = lookupSourceSchema.value?.default || {}
    }
  } catch (err) {
    console.error('Failed to inspect port:', err)
    lookupSourceData.value = lookupSourceSchema.value?.default || {}
  }

  lookupLoading.value = false
  showLookupModal.value = true
}

// Apply expression from lookup modal
const applyLookupExpression = (expression, portName) => {
  if (lookupCallback.value && typeof lookupCallback.value === 'function') {
    lookupCallback.value(expression, portName)
  }
  closeLookupModal()
}

// Close lookup modal
const closeLookupModal = () => {
  showLookupModal.value = false
  lookupCallback.value = null
  lookupTargetSchema.value = {}
  lookupFullSchema.value = {}
  lookupInitialExpression.value = ''
  lookupPortName.value = ''
  lookupSourceData.value = {}
  lookupSourceSchema.value = {}
  lookupFieldTitle.value = ''
}

// Actions
const handleDeselect = () => {
  flowStore.selectElement(null)
}

// Edge delete confirmation dialog
const showDeleteEdgeConfirm = ref(false)

const deleteEdgeDetail = computed(() => {
  if (!selectedEdge.value) return ''
  const targetHandle = selectedEdge.value.targetHandle?.toUpperCase() || ''
  return `${edgeSourceLabel.value} → ${edgeTargetLabel.value} ${targetHandle}`.trim()
})

const handleDeleteEdgeClick = () => {
  if (!selectedEdge.value) return
  showDeleteEdgeConfirm.value = true
}

const handleDeleteEdgeConfirm = async () => {
  if (!selectedEdge.value) return
  showDeleteEdgeConfirm.value = false
  try {
    await flowStore.disconnectNodes(selectedEdge.value.source, selectedEdge.value.id)
  } catch (err) {
    emit('error', `Failed to delete edge: ${err}`)
  }
}

const handleDeleteEdgeCancel = () => {
  showDeleteEdgeConfirm.value = false
}

const handleDeleteNode = () => {
  emit('delete', selectedNode.value)
}

const handleRenameNode = () => {
  emit('rename', selectedNode.value)
}

const handleNodeSettings = () => {
  emit('settings', selectedNode.value)
}

const handleRotateNode = async () => {
  if (!selectedNode.value) return
  try {
    await flowStore.rotateNode(selectedNode.value.id)
  } catch (err) {
    emit('error', `Failed to rotate node: ${err}`)
  }
}

// Save configuration and schema
const saveConfiguration = async () => {
  if (!selectedNode.value || !settingsHandle.value) return
  saving.value = true
  try {
    // Serialize the modified schema (includes any configurable changes made by the user)
    const schemaStr = settingsSchema.value ? JSON.stringify(settingsSchema.value) : ''
    // Use formValue directly to ensure we save the current form state
    const configStr = JSON.stringify(formValue.value)
    await flowStore.updateNodeConfiguration(
      selectedNode.value.id,
      '_settings',
      configStr,
      schemaStr
    )

    // Update local node data in store so it reflects the new config
    const localNode = flowStore.getElement(selectedNode.value.id)
    if (localNode && localNode.data?.handles) {
      const settingsIdx = localNode.data.handles.findIndex(h => h.id === '_settings')
      if (settingsIdx !== -1) {
        localNode.data.handles[settingsIdx].configuration = formValue.value
      }
    }

    // Reset dirty state after save
    originalNodeConfigValue.value = deepCopy(formValue.value)
    nodeConfigDirty.value = false
  } catch (err) {
    emit('error', `Failed to save: ${err}`)
  } finally {
    saving.value = false
  }
}

const saveEdgeConfiguration = async () => {
  if (!selectedEdge.value) return
  saving.value = true
  try {
    const targetTo = `${selectedEdge.value.target}:${selectedEdge.value.targetHandle}`
    // Use edgeFormValue directly to ensure we save the current form state
    const configStr = JSON.stringify(edgeFormValue.value)
    await flowStore.updateEdgeConfiguration(
      selectedEdge.value.source,
      selectedEdge.value.sourceHandle,
      targetTo,
      configStr,
      selectedEdge.value.data?.flowID || flowStore.flowResourceName
    )

    // Update local edge data in store so it reflects the new config
    const localEdge = flowStore.getElement(selectedEdge.value.id)
    if (localEdge && localEdge.data) {
      localEdge.data.configuration = edgeFormValue.value
    }

    // Reset dirty state after save
    originalEdgeConfigValue.value = deepCopy(edgeFormValue.value)
    edgeConfigDirty.value = false
  } catch (err) {
    emit('error', `Failed to save: ${err}`)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <!-- Edge selected -->
  <aside
    v-if="selectedEdge && selectedNodes.length === 0"
    :class="['relative text-sm flex flex-col dark:text-gray-300 flex-shrink-0 bg-gray-50 dark:bg-black border-l-2 border-gray-300 dark:border-gray-600 h-full', panelWidthClass]"
  >
    <!-- Header -->
    <nav class="relative border-b border-gray-200 dark:border-gray-700 flex divide-x divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900">
      <a
        href="#"
        @click.prevent="handleDeselect"
        class="text-gray-600 dark:text-gray-300 group relative min-w-0 flex-1 overflow-hidden py-2 px-2 text-sm font-medium text-center hover:bg-gray-50 dark:hover:bg-gray-800 focus:z-10 whitespace-nowrap"
      >
        <span>{{ edgeSourceLabel }} &rarr; {{ edgeTargetLabel }} {{ selectedEdge.targetHandle?.toUpperCase() }}</span>
        <span class="bg-sky-500 absolute inset-x-0 bottom-0 h-0.5"></span>
      </a>
    </nav>

    <!-- Edge info bar -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 text-xs px-2 py-2 flex justify-start items-center gap-2">
      <button
        type="button"
        @click="handleDeleteEdgeClick"
        :disabled="flowStore.readOnly"
        :title="flowStore.readOnly ? 'You are in read only mode' : ''"
        class="text-red-400 border-red-400 border px-3 py-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50"
      >
        Delete Edge
      </button>
      <div class="px-2">
        <h3 class="font-semibold dark:text-gray-300 text-gray-600">EdgeID</h3>
        <p class="font-light dark:text-gray-400 text-gray-500 text-xs">{{ selectedEdge.id }}</p>
      </div>
    </div>

    <!-- 3-column layout: Source Node | Edge Configuration | Target Node -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Left column: Source data (1/4) -->
      <div class="w-1/4 flex-shrink-0 flex flex-col border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
        <div class="px-2 py-1.5 border-b border-gray-200 dark:border-gray-700 ">
          <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300">{{ edgeSourceLabel }}</h3>
          <p class="text-xs text-gray-400 dark:text-gray-500 truncate">{{ selectedEdge.sourceHandle }}</p>
        </div>
        <div class="flex-1 overflow-y-auto p-2">
          <div v-if="edgeSourceLoading" class="text-center text-gray-400 py-4 text-xs">Loading...</div>
          <VueJsonPretty
            v-else-if="edgeSourceData && Object.keys(edgeSourceData).length > 0"
            :data="edgeSourceData"
            :deep="2"
            :show-length="true"
            class="text-xs"
          />
          <div v-else class="text-center text-gray-400 py-4 text-xs">No data</div>
        </div>
      </div>

      <!-- Middle column: Configuration form (1/2) -->
      <div class="w-1/2 flex-shrink-0 flex flex-col bg-white dark:bg-gray-900">
        <div class="px-2 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
          <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300 flex items-center gap-1">
            <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
              <path d="M20 12l-8-8v5H4v6h8v5l8-8z" />
            </svg>
            Edge configuration
          </h3>
          <p class="text-xs text-gray-400 dark:text-gray-500">Map input data to target port</p>
        </div>
        <form @submit.prevent="saveEdgeConfiguration" class="flex-1 flex flex-col overflow-hidden">
          <div class="flex-1 overflow-y-auto">
            <!-- Schema-based form when schema is available -->
            <SchemaForm
              v-if="edgeSchema && (edgeSchema.properties || edgeSchema.type || edgeSchema.$ref)"
              :key="'edge-form-' + selectedEdge?.id + '-' + configurationResetKey"
              :schema="edgeSchema"
              :model-value="edgeFormValue"
              @update:model-value="updateEdgeFormValue"
              @lookup="handleEdgeLookup"
              :readonly="flowStore.readOnly || selectedEdge.data?.blocked"
              :allow-lookup="true"
              :hide-root-lookup="true"
              :no-border="false"
            />
            <!-- Fallback to raw JSON editor -->
            <div v-else class="p-2">
              <textarea
                v-model="edgeEditorValue"
                class="w-full h-48 p-2 text-xs font-mono bg-gray-50 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded focus:ring-sky-500 focus:border-sky-500 dark:text-gray-300"
                placeholder="{}"
              />
            </div>
            <!-- Validation errors -->
            <div v-if="!edgeIsValid && (edgeValidationError || edgeValidationErrors.length > 0)" class="px-2 pt-2">
              <div class="text-xs font-medium text-red-500 mb-1">Validation errors:</div>
              <div v-if="edgeValidationErrors.length > 0" class="space-y-1">
                <div v-for="err in edgeValidationErrors" :key="err.path" class="text-xs text-red-400">
                  <span class="font-mono">{{ err.path }}</span>&nbsp;&nbsp;{{ err.message }}
                </div>
              </div>
              <div v-else-if="edgeValidationError" class="text-xs text-red-400">
                {{ edgeValidationError }}
              </div>
            </div>
          </div>
          <!-- Warning message and Save button -->
          <div class="text-right px-2 pt-2 pb-3 border-t border-gray-200 dark:border-gray-700">
            <p class="text-xs text-orange-600 pb-2 text-left">Do not store sensitive information if you plan sharing your project as a solution.</p>
            <button
              type="submit"
              :disabled="flowStore.readOnly || saving || selectedEdge.data?.blocked"
              class="px-4 py-2 text-xs font-medium rounded-md text-sky-700 bg-sky-100 hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-sky-500 dark:bg-gray-900 dark:hover:bg-gray-800 dark:text-sky-500 disabled:opacity-50"
            >
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Right column: Preview result (1/4) -->
      <div class="w-1/4 flex-shrink-0 flex flex-col border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
        <div class="px-2 py-1.5 border-b border-gray-200 dark:border-gray-700 ">
          <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300">{{ edgeTargetLabel }}</h3>
          <p class="text-xs text-gray-400 dark:text-gray-500 truncate">{{ selectedEdge.targetHandle }}</p>
        </div>
        <div class="flex-1 overflow-y-auto p-2">
          <div v-if="edgePreviewLoading" class="text-center text-gray-400 py-4 text-xs">Computing...</div>
          <div v-else-if="edgePreviewError" class="text-center text-red-400 py-4 text-xs">{{ edgePreviewError }}</div>
          <VueJsonPretty
            v-else-if="edgePreviewResult && Object.keys(edgePreviewResult).length > 0"
            :data="edgePreviewResult"
            :deep="2"
            :show-length="true"
            class="text-xs"
          />
          <div v-else class="text-center text-gray-400 py-4 text-xs">No preview</div>
        </div>
      </div>
    </div>
  </aside>

  <!-- Multiple nodes selected -->
  <aside
    v-else-if="selectedNodes.length > 1"
    :class="['relative text-sm flex flex-col flex-shrink-0 bg-gray-50 dark:bg-black border-l-2 border-gray-300 dark:border-gray-600 h-full', panelWidthClass]"
  >
    <!-- Toolbar -->
    <div class="flex items-center gap-2 px-3 py-2 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
      <span class="text-sm text-gray-600 dark:text-gray-300">{{ selectedNodes.length }} selected</span>
      <div class="flex-1" />
      <button
        @click="onTransferNodes"
        type="button"
        :disabled="flowStore.readOnly"
        :title="flowStore.readOnly ? 'You are in read only mode' : 'Transfer selected nodes to another flow'"
        class="p-1.5 text-sky-600 dark:text-sky-400 hover:bg-sky-50 dark:hover:bg-sky-900/20 rounded disabled:opacity-50"
      >
        <Square3Stack3DIcon class="h-5 w-5" />
      </button>
      <button
        @click="showDeleteMultipleModal = true"
        type="button"
        :disabled="flowStore.readOnly"
        :title="flowStore.readOnly ? 'You are in read only mode' : 'Delete selected nodes'"
        class="p-1.5 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded disabled:opacity-50"
      >
        <TrashIcon class="h-5 w-5" />
      </button>
      <button
        @click="clearSelection"
        type="button"
        title="Cancel selection"
        class="p-1.5 text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded"
      >
        <XMarkIcon class="h-5 w-5" />
      </button>
    </div>
    <!-- Selected nodes list -->
    <div class="p-2 overflow-y-auto">
      <ul class="space-y-1">
        <li v-for="node in selectedNodes" :key="node.id" class="flex items-center gap-1 px-2 py-1.5 bg-sky-50 dark:bg-sky-900/30 rounded text-sm text-gray-700 dark:text-gray-200">
          <span class="truncate flex-1">{{ node.data?.label || node.id }}</span>
          <button
            @click="deselectNode(node.id)"
            type="button"
            title="Remove from selection"
            class="p-0.5 text-gray-400 hover:text-red-500 dark:hover:text-red-400 flex-shrink-0"
          >
            <XMarkIcon class="h-4 w-4" />
          </button>
        </li>
      </ul>
    </div>

    <!-- Delete multiple nodes confirmation dialog -->
    <ConfirmDialog
      :show="showDeleteMultipleModal"
      :title="`Delete ${selectedNodes.length} nodes?`"
      message="Are you sure you want to delete these nodes? This action cannot be undone."
      confirm-text="Delete all"
      cancel-text="Cancel"
      @confirm="deleteMultipleNodes"
      @cancel="showDeleteMultipleModal = false"
    />

  </aside>

  <!-- Single node selected -->
  <aside
    v-else-if="selectedNode"
    :class="['relative text-sm flex flex-col flex-shrink-0 bg-gray-50 dark:bg-black h-full', panelWidthClass]"
    style="border-left: 1px solid #4b5563;"
  >
    <!-- Configuration tab active -->
    <div
      v-if="configurationTab.current"
      :class="['relative flex flex-col h-full', selectedNode.data?.error ? 'bg-red-50 dark:bg-red-950' : '']"
    >
      <!-- Tab navigation -->
      <nav class="relative z-0 border-b border-gray-200 dark:border-gray-700 flex divide-x divide-gray-200 dark:divide-gray-700">
        <a
          href="#"
          @click.prevent="setCurrentTab('status')"
          :class="[
            statusTab.current ? 'text-gray-600 dark:text-gray-300' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-400',
            'group relative min-w-0 flex-1 overflow-hidden bg-white dark:bg-gray-900 py-2 px-2 text-sm font-medium text-center hover:bg-gray-50 dark:hover:bg-gray-800 focus:z-10 whitespace-nowrap'
          ]"
        >
          <span>{{ selectedNode.data?.label || selectedNode.id }}</span>
          <span :class="[statusTab.current ? 'bg-sky-500' : 'bg-transparent', 'absolute inset-x-0 bottom-0 h-0.5']"></span>
        </a>
        <a
          href="#"
          @click.prevent="setCurrentTab('configuration')"
          :class="[
            selectedNode.data?.error
              ? (configurationTab.current ? 'text-gray-600 dark:text-red-300 bg-red-50 dark:bg-red-950' : 'text-gray-500 hover:text-red-700 bg-red-50')
              : (configurationTab.current ? 'text-gray-600 dark:text-gray-300' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-400'),
            'group relative min-w-0 flex-1 overflow-hidden py-2 px-2 text-sm font-medium text-center focus:z-10 whitespace-nowrap hover:bg-gray-50 dark:hover:bg-gray-800'
          ]"
        >
          <span>Configuration</span>
          <span :class="[configurationTab.current ? 'bg-sky-500' : 'bg-transparent', 'absolute inset-x-0 bottom-0 h-0.5']"></span>
        </a>
      </nav>

      <!-- Error message -->
      <p v-if="selectedNode.data?.error" class="font-light p-4 text-center text-red-600 dark:text-red-300">
        Error: {{ selectedNode.data?.status }}
      </p>

      <!-- Blocked node notice -->
      <div v-if="selectedNode.data?.blocked" class="mx-2 mt-2 p-3 bg-gray-100 dark:bg-gray-800 rounded-lg text-xs text-gray-600 dark:text-gray-400">
        <p class="font-medium">This node is shared from flow: <span class="text-emerald-600 dark:text-emerald-400">{{ selectedNode.data?.flow_id }}</span></p>
        <p class="mt-1">Configuration is read-only. Edit this node in its original flow.</p>
      </div>

      <!-- Settings form -->
      <form v-if="settingsHandle" @submit.prevent="saveConfiguration">
        <div class="overflow-y-auto">
          <!-- Schema-based form when both schema and configuration are available -->
          <SchemaForm
            v-if="settingsSchema && (settingsSchema.properties || settingsSchema.type || settingsSchema.$ref) && configurationReady"
            :key="'node-form-' + selectedNode?.id + '-' + configurationResetKey"
            :schema="settingsSchema"
            :model-value="formValue"
            @update:model-value="updateFormValue"
            :readonly="flowStore.readOnly || selectedNode.data?.blocked"
            :allow-edit-schema="!flowStore.readOnly && !selectedNode.data?.blocked"
          />
          <div v-else-if="settingsSchema && (settingsSchema.properties || settingsSchema.type || settingsSchema.$ref)" class="text-center text-gray-400 py-4">
            Loading configuration...
          </div>
          <!-- Fallback to raw JSON editor -->
          <div v-else class="p-2">
            <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Settings (JSON)</label>
            <textarea
              v-model="editorValue"
              class="w-full h-64 p-2 text-xs font-mono bg-gray-50 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded focus:ring-sky-500 focus:border-sky-500 dark:text-gray-300"
              placeholder="{}"
            />
          </div>
        </div>
        <!-- Warning message and Save button -->
        <div class="text-right px-2 pt-2 pb-4">
          <p class="text-xs text-orange-600 pb-2 text-left">Do not store sensitive information if you plan sharing your project as a solution.</p>
          <button
            type="submit"
            :disabled="flowStore.readOnly || saving || selectedNode.data?.blocked"
            class="px-4 py-2 text-xs font-medium rounded-md text-sky-700 bg-sky-100 hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-sky-500 dark:bg-gray-900 dark:hover:bg-gray-800 dark:text-sky-500 disabled:opacity-50"
          >
            {{ saving ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </form>
      <div v-else class="p-4 pt-5 text-center dark:text-gray-400">
        No configuration needed for the selected node.
      </div>
    </div>

    <!-- Status tab active -->
    <div v-if="statusTab.current" class="flex flex-col h-full">
      <!-- Tab navigation -->
      <nav class="relative z-0 border-b border-gray-200 dark:border-gray-700 flex divide-x divide-gray-200 dark:divide-gray-700">
        <a
          href="#"
          @click.prevent="setCurrentTab('status')"
          :class="[
            statusTab.current ? 'text-gray-600 dark:text-gray-300' : 'text-gray-500 hover:text-gray-700',
            'group relative min-w-0 flex-1 overflow-hidden bg-white dark:bg-gray-900 py-2 px-2 text-sm font-medium text-center hover:bg-gray-50 dark:hover:bg-gray-800 focus:z-10 whitespace-nowrap'
          ]"
        >
          <span>{{ selectedNode.data?.label || selectedNode.id }}</span>
          <span v-if="selectedNode.data?.blocked" class="text-xs text-gray-400"> [shared]</span>
          <span :class="[statusTab.current ? 'bg-sky-500' : 'bg-transparent', 'absolute inset-x-0 bottom-0 h-0.5']"></span>
        </a>
        <a
          href="#"
          @click.prevent="setCurrentTab('configuration')"
          :class="[
            selectedNode.data?.error
              ? (configurationTab.current ? 'text-gray-600 dark:text-red-300 bg-red-50' : 'text-gray-500 hover:text-red-700 bg-red-50 dark:bg-red-900')
              : (configurationTab.current ? 'text-gray-600 dark:text-gray-300' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-400'),
            'group relative min-w-0 flex-1 overflow-hidden py-2 px-2 text-sm font-medium text-center focus:z-10 whitespace-nowrap hover:bg-gray-50 dark:hover:bg-gray-800'
          ]"
        >
          <span>Configuration</span>
          <span :class="[configurationTab.current ? 'bg-sky-500' : 'bg-transparent', 'absolute inset-x-0 bottom-0 h-0.5']"></span>
        </a>
      </nav>

      <!-- Node status content -->
      <div class="flex flex-col h-full overflow-y-auto">
        <!-- Node info card -->
        <div class="bg-white m-1 dark:bg-gray-900 shadow rounded-lg text-xs">
          <div class="px-2 py-2 flex justify-between">
            <div class="w-full">
              <!-- Node ID -->
              <input
                type="text"
                readonly
                class="w-full m-1 p-1 text-xs border-transparent dark:bg-gray-900 focus:border-transparent focus:ring-0 dark:text-gray-300 text-gray-600"
                :value="selectedNode.id"
              />
              <!-- Module/Component info -->
              <div class="font-light px-2 dark:text-gray-300 text-gray-600">
                <p v-if="selectedNode.data?.description">{{ decodeHtmlEntities(selectedNode.data.description) }}</p>
                <p>Module: <span class="font-semibold">{{ selectedNode.data?.module }}</span></p>
                <p>Component: <span class="font-semibold">{{ selectedNode.data?.component }}</span></p>
                <!-- Show source flow for blocked/shared nodes -->
                <p v-if="selectedNode.data?.blocked && selectedNode.data?.flow_id" class="text-emerald-600 dark:text-emerald-400">
                  Flow: <span class="font-semibold">{{ selectedNode.data?.flow_id }}</span>
                </p>
                <p :class="selectedNodeExpiring ? 'text-red-500' : ''">
                  Last update:
                  <span class="font-semibold" :class="{ 'text-red-500': !selectedNode.data?.last_status_update }">
                    {{ timeAgo(selectedNode.data?.last_status_update) }}
                  </span>
                </p>
              </div>
            </div>

            <!-- Actions menu -->
            <Menu v-if="!flowStore.readOnly" as="div" class="ml-3 relative inline-block text-left">
              <MenuButton class="-my-2 p-2 rounded-full flex items-center text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none">
                <EllipsisVerticalIcon class="h-5 w-5" />
              </MenuButton>
              <transition
                enter-active-class="transition ease-out duration-100"
                enter-from-class="transform opacity-0 scale-95"
                enter-to-class="transform opacity-100 scale-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
              >
                <MenuItems class="origin-top-right absolute z-40 right-0 mt-2 w-48 rounded-md shadow-lg bg-white border border-gray-200 focus:outline-none dark:border-gray-700 dark:bg-gray-900">
                  <div class="py-1">
                    <!-- Blocked node message -->
                    <div v-if="selectedNode.data?.blocked" class="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
                      This node is shared from another flow. Edit it in the original flow.
                    </div>
                    <MenuItem v-slot="{ active }" v-if="!selectedNode.data?.blocked && !flowStore.readOnly">
                      <button
                        type="button"
                        @click="onTransferSingleNode"
                        :class="[active ? 'bg-gray-100 text-gray-900 dark:bg-gray-700' : 'text-gray-700 dark:text-gray-300', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <Square3Stack3DIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Transfer node</span>
                      </button>
                    </MenuItem>
                    <MenuItem v-slot="{ active }" v-if="!selectedNode.data?.blocked && !flowStore.readOnly">
                      <button
                        type="button"
                        @click="handleRenameNode"
                        :class="[active ? 'bg-gray-100 text-gray-900 dark:bg-gray-700' : 'text-gray-700 dark:text-gray-300', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <PencilIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Rename</span>
                      </button>
                    </MenuItem>
                    <MenuItem v-slot="{ active }" v-if="!selectedNode.data?.blocked && !flowStore.readOnly">
                      <button
                        type="button"
                        @click="handleNodeSettings"
                        :class="[active ? 'bg-gray-100 text-gray-900 dark:bg-gray-700' : 'text-gray-700 dark:text-gray-300', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <Cog6ToothIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Settings</span>
                      </button>
                    </MenuItem>
                    <MenuItem v-slot="{ active }" v-if="!selectedNode.data?.blocked && !flowStore.readOnly">
                      <button
                        type="button"
                        @click="handleDeleteNode"
                        :class="[active ? 'bg-gray-100 text-orange-500 dark:bg-gray-700' : 'text-orange-600 dark:text-orange-400', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <TrashIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Delete</span>
                      </button>
                    </MenuItem>
                  </div>
                </MenuItems>
              </transition>
            </Menu>
          </div>
        </div>

        <!-- Control port form -->
        <div v-if="controlHandle && controlHandleSchema" :key="'control-' + selectedNode?.id" class="bg-white dark:bg-gray-900 shadow rounded text-xs m-1 p-2">
          <SchemaForm
            :key="'control-form-' + selectedNode?.id"
            :schema="controlHandleSchema"
            :model-value="controlFormValue"
            @update:model-value="updateControlFormValue"
            @action="handleControlUpdate"
          />
          <div v-if="controlHandle.error" class="mt-2 p-2 text-xs bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
            {{ controlHandle.error }}
          </div>
        </div>

        <!-- Port tabs and data display -->
        <div class="px-2 mt-2 flex-1 flex flex-col min-h-0">
          <PortTabs
            :handles="selectedNodeHandles"
            :selected-handle-id="selectedHandle?.id"
            @select="selectedHandleId = $event"
            class="flex-1 min-h-[250px]"
          >
            <div class="relative bg-white dark:bg-gray-900 dark:text-gray-300 shadow rounded-none text-xs h-full flex flex-col border border-gray-300 dark:border-gray-600">
              <!-- Copy button -->
              <button
                v-if="inspectReady && inspect?.data !== undefined"
                @click="copyInspectToClipboard"
                class="absolute top-2 right-2 z-10 p-1.5 rounded-md transition-colors
                  bg-gray-100 hover:bg-gray-200 text-gray-600
                  dark:bg-gray-800 dark:hover:bg-gray-700 dark:text-gray-400
                  focus:outline-none focus:ring-2 focus:ring-sky-500"
                :title="inspectCopied ? 'Copied!' : 'Copy to clipboard'"
              >
                <ClipboardDocumentCheckIcon v-if="inspectCopied" class="h-4 w-4 text-green-500" />
                <ClipboardDocumentIcon v-else class="h-4 w-4" />
              </button>
              <div class="p-2 flex-1 overflow-y-auto">
                <div v-if="!inspectReady" class="text-center text-gray-400 py-4">Loading...</div>
                <div v-else-if="inspect?.dataError" class="text-center text-red-400 py-4 text-xs">{{ inspect.dataError }}</div>
                <VueJsonPretty
                  v-else-if="inspect?.data !== undefined"
                  :data="inspect.data"
                  :deep="2"
                  :show-length="true"
                  class="text-xs"
                />
                <div v-else class="text-center text-gray-400 py-4">No data</div>
              </div>
            </div>
          </PortTabs>
        </div>

        <!-- Info text -->
        <div v-if="!selectedNode.data?.trace" class="p-2 min-h-12">
          <p class="text-xs dark:text-gray-500 text-gray-400">
            Data structure generated automatically based on component meta information. Values may be explanatory.
          </p>
        </div>
      </div>
    </div>
  </aside>

  <!-- No selection - flow info -->
  <aside
    v-else
    :class="['relative text-sm flex flex-col dark:text-gray-300 flex-shrink-0 bg-gray-50 dark:bg-black border-l-2 border-gray-300 dark:border-gray-600 h-full', panelWidthClass]"
  >
    <div class="flex flex-col h-full">
      <div class="relative z-20 border-b border-gray-200 dark:border-gray-700 flex divide-x divide-gray-200 dark:divide-gray-700">
        <div class="text-gray-600 dark:text-gray-300 group relative min-w-0 flex-1 bg-white dark:bg-gray-900 py-2 px-3 text-sm font-medium text-center focus:z-10 whitespace-nowrap">
          <span>{{ flowStore.flowName || 'Flow' }}</span>
        </div>
      </div>
      <div class="flex-1 flex items-center justify-center">
        <div class="text-center text-gray-400 dark:text-gray-500 px-4">
          <p class="text-sm">Select a node or edge to view its properties</p>
          <p class="text-xs mt-2">Double-click on canvas to add components</p>
        </div>
      </div>
    </div>
  </aside>

  <!-- Expression lookup modal -->
  <FlowDataLookupModal
    :show="showLookupModal"
    :source-data="lookupSourceData"
    :source-schema="lookupSourceSchema"
    :target-schema="lookupTargetSchema"
    :full-schema="lookupFullSchema"
    :initial-expression="lookupInitialExpression"
    :port-name="lookupPortName"
    :field-title="lookupFieldTitle"
    @close="closeLookupModal"
    @apply="applyLookupExpression"
  />

  <!-- Delete edge confirmation dialog -->
  <ConfirmDialog
    :show="showDeleteEdgeConfirm"
    title="Confirmation."
    message="Are you sure you want to delete this edge"
    :detail="deleteEdgeDetail"
    @confirm="handleDeleteEdgeConfirm"
    @cancel="handleDeleteEdgeCancel"
  />

  <!-- Unsaved changes dialog -->
  <UnsavedChangesDialog
    :show="showUnsavedChangesDialog"
    @cancel="handleUnsavedCancel"
    @discard="handleUnsavedDiscard"
    @save="handleUnsavedSave"
  />

  <!-- Transfer nodes modal (used for both single and multi-node transfer) -->
  <FlowTransferModal
    v-model="showTransferModal"
    :project-name="flowStore.projectResourceName"
    :context-name="flowStore.contextName"
    :namespace="flowStore.namespace"
    @error="(msg) => emit('error', msg)"
  />
</template>

<style scoped>
/* VueJsonPretty size override */
:deep(.vjs-tree) {
  font-size: 11px !important;
  line-height: 1.4 !important;
}

/* VueJsonPretty hover styles for dark mode */
:deep(.vjs-tree-node:hover) {
  background-color: rgba(59, 130, 246, 0.15) !important;
}

.dark :deep(.vjs-tree-node:hover) {
  background-color: rgba(59, 130, 246, 0.25) !important;
}

.dark :deep(.vjs-tree-node.is-highlight) {
  background-color: rgba(59, 130, 246, 0.35) !important;
}
</style>
