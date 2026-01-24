<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import { CreateTracker, DeleteTracker } from '../../../wailsjs/go/main/App'
import ControlPanel from './ControlPanel.vue'
import FlowCanvas from './FlowCanvas.vue'
import FlowAddComponent from './FlowAddComponent.vue'
import FlowImportModal from './FlowImportModal.vue'
import FlowExportModal from './FlowExportModal.vue'
import SidePanel from './SidePanel.vue'
import Telemetry from './Telemetry.vue'
import Trace from './Trace.vue'

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
  flowResourceName: String,
})

const emit = defineEmits(['close'])

const flowStore = useFlowStore()
const error = ref('')
const sidePanelOpen = ref(true)

// Add component modal state
const showAddComponent = ref(false)
const newNodePosition = ref({ x: 100, y: 100 })

// Import/Export modal state
const showImportModal = ref(false)
const showExportModal = ref(false)

// Auto-open side panel when something is selected
const hasSelection = computed(() => flowStore.selectedNode || flowStore.selectedEdge)
watch(hasSelection, (val) => {
  if (val) {
    sidePanelOpen.value = true
  }
})

// Rename dialog state
const showRenameDialog = ref(false)
const renameNode = ref(null)
const newNodeLabel = ref('')

// Delete node dialog state
const showDeleteDialog = ref(false)
const deleteNode = ref(null)
const deleting = ref(false)

// Delete edge dialog state
const showDeleteEdgeDialog = ref(false)
const deleteEdge = ref(null)
const deletingEdge = ref(false)

const loadFlow = async () => {
  try {
    await flowStore.load(props.ctx, props.ns, props.projectName, props.flowResourceName)
    await flowStore.startWatching()
    // Create tracker for telemetry collection
    try {
      await CreateTracker(props.ctx, props.ns, props.projectName)
    } catch (e) {
      console.warn('Failed to create tracker:', e)
    }
  } catch (err) {
    error.value = `Failed to load flow: ${err}`
  }
}

const handleClose = async () => {
  // Delete tracker when leaving flow
  try {
    await DeleteTracker(props.ctx, props.ns)
  } catch (e) {
    console.warn('Failed to delete tracker:', e)
  }
  flowStore.stopWatching()
  flowStore.clean()
  emit('close')
}

const handleError = (err) => {
  error.value = err
  setTimeout(() => {
    error.value = ''
  }, 5000)
}

// Handle rename from SidePanel
const handleRename = (node) => {
  renameNode.value = node
  newNodeLabel.value = node?.data?.label || ''
  showRenameDialog.value = true
}

const submitRename = async () => {
  if (!renameNode.value || !newNodeLabel.value.trim()) return
  try {
    await flowStore.updateNodeLabel(renameNode.value.id, newNodeLabel.value.trim())
    showRenameDialog.value = false
  } catch (err) {
    handleError(`Failed to rename: ${err}`)
  }
}

// Handle delete from SidePanel
const handleDelete = (node) => {
  if (!node) return
  deleteNode.value = node
  showDeleteDialog.value = true
}

const submitDelete = async () => {
  if (!deleteNode.value) return
  deleting.value = true
  try {
    await flowStore.deleteNode(deleteNode.value.id)
    showDeleteDialog.value = false
    deleteNode.value = null
  } catch (err) {
    handleError(`Failed to delete: ${err}`)
  } finally {
    deleting.value = false
  }
}

const cancelDelete = () => {
  showDeleteDialog.value = false
  deleteNode.value = null
}

// Handle delete edge from FlowCanvas (keyboard Delete key)
const handleDeleteEdge = (edge) => {
  if (!edge) return
  deleteEdge.value = edge
  showDeleteEdgeDialog.value = true
}

const submitDeleteEdge = async () => {
  if (!deleteEdge.value) return
  deletingEdge.value = true
  try {
    await flowStore.disconnectNodes(deleteEdge.value.source, deleteEdge.value.id)
    showDeleteEdgeDialog.value = false
    deleteEdge.value = null
  } catch (err) {
    handleError(`Failed to delete edge: ${err}`)
  } finally {
    deletingEdge.value = false
  }
}

const cancelDeleteEdge = () => {
  showDeleteEdgeDialog.value = false
  deleteEdge.value = null
}

// Handle settings from SidePanel (not implemented yet)
const handleSettings = (node) => {
  console.log('Settings for node:', node)
}

// Handle add node from FlowCanvas (double-click or plus button)
const handleAddNode = (position) => {
  newNodePosition.value = position
  showAddComponent.value = true
}

// Handle import from ControlPanel
const handleImport = () => {
  showImportModal.value = true
}

// Handle export from ControlPanel
const handleExport = () => {
  showExportModal.value = true
}

onMounted(() => {
  loadFlow()
})

onUnmounted(() => {
  // Clean up tracker when component unmounts
  DeleteTracker(props.ctx, props.ns).catch(e => {
    console.warn('Failed to delete tracker on unmount:', e)
  })
  flowStore.stopWatching()
  flowStore.clean()
})

// Reload when flow changes
watch(() => props.flowResourceName, async () => {
  // Delete tracker for old flow
  try {
    await DeleteTracker(props.ctx, props.ns)
  } catch (e) {
    console.warn('Failed to delete tracker on flow change:', e)
  }
  flowStore.clean()
  loadFlow()
})
</script>

<template>
  <div class="flow-editor-page h-full flex flex-col bg-white dark:bg-gray-900">
    <!-- Loading state -->
    <div v-if="flowStore.loading" class="flex items-center justify-center h-full">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-500 mx-auto mb-4"></div>
        <span class="text-gray-500 dark:text-gray-400">Loading flow...</span>
      </div>
    </div>

    <!-- Main content -->
    <template v-else>
      <ControlPanel
        :flow-name="flowStore.flowName"
        :flow-resource-name="flowStore.flowResourceName"
        :project-name="flowStore.projectName"
        :loading="flowStore.loadingAlt"
        @close="handleClose"
        @error="handleError"
        @import="handleImport"
        @export="handleExport"
      />

      <!-- Error banner -->
      <div v-if="error" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
        <span class="text-sm text-red-600 dark:text-red-400">{{ error }}</span>
      </div>

      <!-- Editor area -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Canvas and Side Panel -->
        <div class="flex-1 flex overflow-hidden relative">
          <!-- Trace overlay -->
          <Trace
            v-if="flowStore.trace"
            :trace="flowStore.trace"
            :ctx="ctx"
            :ns="ns"
            :project-name="projectName"
            @close="flowStore.clearTrace()"
          />

          <!-- Canvas -->
          <div class="flex-1 relative">
            <FlowCanvas
              @error="handleError"
              @add-node="handleAddNode"
              @delete-node="handleDelete"
              @delete-edge="handleDeleteEdge"
            />
            <!-- Loading overlay for trace/save operations -->
            <div v-if="flowStore.loadingAlt" class="absolute inset-0 bg-white/50 dark:bg-gray-900/50 flex items-center justify-center z-40">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-500"></div>
            </div>
          </div>

          <!-- Add Component Modal -->
          <FlowAddComponent
            v-model="showAddComponent"
            :position="newNodePosition"
          />

          <!-- Import Modal -->
          <FlowImportModal
            v-model="showImportModal"
            @error="handleError"
          />

          <!-- Export Modal -->
          <FlowExportModal
            v-model="showExportModal"
            @error="handleError"
          />

          <!-- Side Panel -->
          <SidePanel
            v-if="sidePanelOpen"
            @error="handleError"
            @rename="handleRename"
            @settings="handleSettings"
            @delete="handleDelete"
            @close="sidePanelOpen = false"
          />
        </div>

        <!-- Telemetry section at the bottom -->
        <Telemetry
          :ctx="ctx"
          :ns="ns"
          :flow-name="flowStore.flowResourceName"
          :project-name="projectName"
        />
      </div>
    </template>

    <!-- Rename Node Dialog -->
    <div
      v-if="showRenameDialog"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      @keydown.enter.prevent="submitRename"
      @keydown.escape="showRenameDialog = false"
    >
      <div class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm" @click="showRenameDialog = false"></div>
      <div class="relative bg-white dark:bg-gray-900 dark:border dark:border-gray-700 rounded-lg shadow-xl w-full max-w-sm p-4">
        <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-3">Rename Node</h3>
        <div class="mb-3">
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Node Name</label>
          <input
            v-model="newNodeLabel"
            type="text"
            class="w-full px-2 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-1 focus:ring-sky-500 focus:border-sky-500"
            autofocus
          />
        </div>
        <div class="flex gap-2">
          <button
            @click="showRenameDialog = false"
            class="flex-1 px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 border border-gray-600 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800"
          >
            Cancel
          </button>
          <button
            @click="submitRename"
            :disabled="!newNodeLabel.trim()"
            class="flex-1 px-3 py-2 text-xs font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50"
          >
            Rename
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Node Confirmation Dialog -->
    <div
      v-if="showDeleteDialog"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      @keydown.enter.prevent="submitDelete"
      @keydown.escape="cancelDelete"
    >
      <div class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm" @click="cancelDelete"></div>
      <div class="relative bg-white dark:bg-gray-900 dark:border dark:border-gray-700 rounded-lg shadow-xl w-full max-w-sm p-4 text-center">
        <!-- Warning icon -->
        <div class="mx-auto w-12 h-12 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </div>
        <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">Delete Node</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">
          Are you sure you want to delete this node?
        </p>
        <p class="text-sm text-gray-900 dark:text-white font-medium mb-4">
          {{ deleteNode?.data?.label || deleteNode?.id }}
        </p>
        <div class="flex gap-2">
          <button
            @click="cancelDelete"
            :disabled="deleting"
            class="flex-1 px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 border border-gray-600 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            @click="submitDelete"
            :disabled="deleting"
            class="flex-1 px-3 py-2 text-xs font-medium text-white bg-red-500 rounded-md hover:bg-red-600 disabled:opacity-50"
          >
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Edge Confirmation Dialog -->
    <div
      v-if="showDeleteEdgeDialog"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      @keydown.enter.prevent="submitDeleteEdge"
      @keydown.escape="cancelDeleteEdge"
    >
      <div class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm" @click="cancelDeleteEdge"></div>
      <div class="relative bg-white dark:bg-gray-900 dark:border dark:border-gray-700 rounded-lg shadow-xl w-full max-w-sm p-4 text-center">
        <!-- Warning icon -->
        <div class="mx-auto w-12 h-12 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </div>
        <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">Delete Connection</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
          Are you sure you want to delete this connection?
        </p>
        <div class="flex gap-2">
          <button
            @click="cancelDeleteEdge"
            :disabled="deletingEdge"
            class="flex-1 px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 border border-gray-600 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            @click="submitDeleteEdge"
            :disabled="deletingEdge"
            class="flex-1 px-3 py-2 text-xs font-medium text-white bg-red-500 rounded-md hover:bg-red-600 disabled:opacity-50"
          >
            {{ deletingEdge ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
