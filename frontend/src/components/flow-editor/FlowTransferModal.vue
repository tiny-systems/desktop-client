<script setup>
import { ref, computed, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'

const GoApp = window.go?.main?.App

const props = defineProps({
  modelValue: Boolean,
  projectName: String,
  contextName: String,
  namespace: String
})

const emit = defineEmits(['update:modelValue', 'error', 'success'])

const flowStore = useFlowStore()
const flows = ref([])
const selectedFlowResourceName = ref('')
const loading = ref(false)
const error = ref('')

// Compute the node IDs to transfer
const nodeIds = computed(() => flowStore.selectedNodes.map(n => n.id))

// Compute the title based on selection
const title = computed(() => {
  const count = flowStore.selectedNodes.length
  if (count > 1) {
    return `Transfer ${count} nodes to another flow`
  }
  const label = flowStore.selectedNode?.data?.label || flowStore.selectedNode?.id
  return `Transfer node ${label} to another flow`
})

// Filter out current flow from the list
const availableFlows = computed(() => {
  return flows.value.filter(f => f.resourceName !== flowStore.flowResourceName)
})

// Load flows when modal opens
watch(() => props.modelValue, async (isOpen) => {
  if (isOpen) {
    error.value = ''
    selectedFlowResourceName.value = ''
    await loadFlows()
  }
})

const loadFlows = async () => {
  if (!GoApp) return

  loading.value = true
  try {
    const result = await GoApp.GetFlows(props.contextName, props.namespace, props.projectName)
    flows.value = result || []
  } catch (e) {
    error.value = e.message || 'Failed to load flows'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  emit('update:modelValue', false)
  selectedFlowResourceName.value = ''
  error.value = ''
}

const transferNodes = async () => {
  if (!selectedFlowResourceName.value || nodeIds.value.length === 0) return

  loading.value = true
  error.value = ''

  try {
    await flowStore.transferNodesToFlow(selectedFlowResourceName.value, nodeIds.value)
    emit('success')
    closeModal()
  } catch (e) {
    error.value = e.message || 'Failed to transfer nodes'
    emit('error', error.value)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 md:p-20"
    @keydown.escape="closeModal"
  >
    <!-- Backdrop -->
    <div
      class="fixed inset-0 bg-gray-500 bg-opacity-25 dark:bg-black dark:bg-opacity-75 backdrop-blur-sm"
      @click="closeModal"
    ></div>

    <!-- Modal -->
    <div class="relative transform rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 p-4 w-full max-w-lg mx-auto dark:bg-black dark:border dark:border-gray-800 dark:text-gray-300">
      <!-- Icon -->
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 dark:bg-green-900">
        <svg class="h-6 w-6 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
        </svg>
      </div>

      <!-- Title -->
      <h3 class="text-center mt-3 font-medium text-gray-900 dark:text-gray-100">
        {{ title }}
      </h3>

      <!-- Description -->
      <p class="text-center mt-2 text-sm text-gray-500 dark:text-gray-400">
        Connected nodes will be automatically shared with the destination flow.
      </p>

      <!-- Flow selector -->
      <div class="mt-4">
        <label for="targetFlow" class="block text-sm text-gray-500 dark:text-gray-400">
          Destination Flow
        </label>
        <select
          id="targetFlow"
          v-model="selectedFlowResourceName"
          :disabled="loading"
          class="mt-1 border-sky-600 placeholder-gray-400 focus:ring-sky-600 appearance-none border rounded w-full py-3 px-3 text-gray-700 leading-tight transition duration-150 ease-in-out sm:text-sm sm:leading-5 dark:bg-gray-900 dark:text-gray-300"
        >
          <option value="" disabled>Select a flow...</option>
          <option
            v-for="flow in availableFlows"
            :key="flow.resourceName"
            :value="flow.resourceName"
          >
            {{ flow.name || flow.resourceName }}
          </option>
        </select>
      </div>

      <!-- Error message -->
      <div v-if="error" class="text-red-500 text-sm py-2 mt-2">
        {{ error }}
      </div>

      <!-- No flows message -->
      <div v-if="availableFlows.length === 0 && !loading" class="text-gray-500 text-sm py-2 mt-2 text-center">
        No other flows available in this project.
      </div>

      <!-- Buttons -->
      <div class="mt-5 sm:grid sm:grid-cols-2 sm:gap-3">
        <button
          @click="closeModal"
          type="button"
          class="w-full text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-lg border border-gray-200 text-sm font-medium px-5 py-2 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
        >
          Cancel
        </button>
        <button
          @click="transferNodes"
          :disabled="!selectedFlowResourceName || loading"
          type="button"
          :class="[
            selectedFlowResourceName && !loading
              ? 'text-white bg-sky-600 hover:bg-sky-800 focus:ring-4'
              : 'text-sky-500 bg-sky-200 cursor-not-allowed',
            'w-full focus:outline-none focus:ring-sky-300 dark:focus:ring-sky-800 font-medium rounded-lg text-sm px-5 py-2 text-center'
          ]"
        >
          {{ loading ? 'Transferring...' : 'Transfer' }}
        </button>
      </div>
    </div>
  </div>
</template>
