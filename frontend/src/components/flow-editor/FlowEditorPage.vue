<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import ControlPanel from './ControlPanel.vue'
import FlowCanvas from './FlowCanvas.vue'
import FlowAddComponent from './FlowAddComponent.vue'
import SidePanel from './SidePanel.vue'
import Telemetry from './Telemetry.vue'

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

const loadFlow = async () => {
  try {
    await flowStore.load(props.ctx, props.ns, props.projectName, props.flowResourceName)
    await flowStore.startWatching()
  } catch (err) {
    error.value = `Failed to load flow: ${err}`
  }
}

const handleClose = () => {
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
const handleDelete = async (node) => {
  if (!node) return
  if (confirm(`Delete node "${node.data?.label || node.id}"?`)) {
    try {
      await flowStore.deleteNode(node.id)
    } catch (err) {
      handleError(`Failed to delete: ${err}`)
    }
  }
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

onMounted(() => {
  loadFlow()
})

onUnmounted(() => {
  flowStore.stopWatching()
  flowStore.clean()
})

// Reload when flow changes
watch(() => props.flowResourceName, () => {
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
      />

      <!-- Error banner -->
      <div v-if="error" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
        <span class="text-sm text-red-600 dark:text-red-400">{{ error }}</span>
      </div>

      <!-- Editor area -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Canvas and Side Panel -->
        <div class="flex-1 flex overflow-hidden relative">
          <!-- Canvas -->
          <div class="flex-1">
            <FlowCanvas @error="handleError" @add-node="handleAddNode" />
          </div>

          <!-- Add Component Modal -->
          <FlowAddComponent
            v-model="showAddComponent"
            :position="newNodePosition"
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
          :flow-name="flowStore.flowResourceName"
          :project-name="projectName"
          @trace="(traceId) => flowStore.highlightTrace(traceId)"
        />
      </div>
    </template>

    <!-- Rename Node Dialog -->
    <div
      v-if="showRenameDialog"
      class="fixed inset-0 z-50 overflow-y-auto"
      @keydown.enter.prevent="submitRename"
      @keydown.escape="showRenameDialog = false"
    >
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="showRenameDialog = false"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Rename Node</h3>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Node Name</label>
            <input
              v-model="newNodeLabel"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-sky-500 focus:border-transparent"
              autofocus
            />
          </div>
          <div class="flex justify-end gap-2">
            <button
              @click="showRenameDialog = false"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              @click="submitRename"
              :disabled="!newNodeLabel.trim()"
              class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50"
            >
              Rename
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
