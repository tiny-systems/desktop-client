<script setup>
import { ref, onMounted, nextTick } from 'vue'
import FlowCard from './FlowCard.vue'
import { PlusIcon, XMarkIcon } from '@heroicons/vue/24/outline'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
})

const emit = defineEmits(['error'])

const flows = ref([])
const loading = ref(true)
const showCreateDialog = ref(false)
const newFlowName = ref('')
const creating = ref(false)
const createDialogRef = ref(null)

const loadFlows = async () => {
  if (!GoApp) return
  loading.value = true
  try {
    const result = await GoApp.GetFlows(props.ctx, props.ns, props.projectName)
    flows.value = result || []
  } catch (err) {
    emit('error', `Failed to load flows: ${err}`)
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  newFlowName.value = ''
  showCreateDialog.value = true
  nextTick(() => {
    createDialogRef.value?.focus()
  })
}

const closeCreateDialog = () => {
  showCreateDialog.value = false
  newFlowName.value = ''
}

const createFlow = async () => {
  if (!GoApp || !newFlowName.value.trim()) return

  creating.value = true
  try {
    const newFlow = await GoApp.CreateFlow(props.ctx, props.ns, props.projectName, newFlowName.value.trim())
    if (newFlow) {
      flows.value.push(newFlow)
    }
    closeCreateDialog()
  } catch (err) {
    emit('error', `Failed to create flow: ${err}`)
  } finally {
    creating.value = false
  }
}

const handleUndeploy = (resourceName) => {
  flows.value = flows.value.filter(f => f.resourceName !== resourceName)
}

const handleRename = (resourceName, newName) => {
  const flow = flows.value.find(f => f.resourceName === resourceName)
  if (flow) {
    flow.name = newName
  }
}

onMounted(() => {
  loadFlows()
})
</script>

<template>
  <div class="flows-tab h-full flex flex-col">
    <!-- Header with create button -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700">
      <h2 class="text-sm font-medium text-gray-700 dark:text-gray-300">Flows</h2>
      <button
        @click="openCreateDialog"
        class="flex items-center space-x-1 px-3 py-1.5 text-sm font-medium text-white bg-sky-600 rounded-lg hover:bg-sky-700 transition-colors"
      >
        <PlusIcon class="w-4 h-4" />
        <span>New Flow</span>
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4">
      <div v-if="loading" class="flex items-center justify-center h-64">
        <span class="text-gray-500 dark:text-gray-400">Loading flows...</span>
      </div>
      <div v-else-if="flows.length === 0" class="flex flex-col items-center justify-center h-64 space-y-4">
        <span class="text-gray-500 dark:text-gray-400">No flows in this project</span>
        <button
          @click="openCreateDialog"
          class="flex items-center space-x-1 px-4 py-2 text-sm font-medium text-sky-600 dark:text-sky-400 border border-sky-600 dark:border-sky-400 rounded-lg hover:bg-sky-50 dark:hover:bg-sky-900/20 transition-colors"
        >
          <PlusIcon class="w-4 h-4" />
          <span>Create your first flow</span>
        </button>
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <FlowCard
          v-for="flow in flows"
          :key="flow.resourceName"
          :ctx="ctx"
          :ns="ns"
          :project-name="projectName"
          :flow="flow"
          @error="(err) => emit('error', err)"
          @undeploy="handleUndeploy"
          @rename="handleRename"
        />
      </div>
    </div>

    <!-- Create Flow Dialog -->
    <div
      v-if="showCreateDialog"
      ref="createDialogRef"
      tabindex="-1"
      class="fixed inset-0 z-50 overflow-y-auto outline-none"
      @keydown.enter.prevent="createFlow"
      @keydown.escape="closeCreateDialog"
    >
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeCreateDialog"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
              Create New Flow
            </h3>
            <button @click="closeCreateDialog" class="text-gray-400 hover:text-gray-500">
              <XMarkIcon class="w-5 h-5" />
            </button>
          </div>
          <div class="mb-4">
            <label for="flow-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Flow Name
            </label>
            <input
              id="flow-name"
              v-model="newFlowName"
              type="text"
              placeholder="My Awesome Flow"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-sky-500 focus:border-transparent"
              autofocus
            />
          </div>
          <div class="flex justify-end gap-2">
            <button
              @click="closeCreateDialog"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              @click="createFlow"
              :disabled="!newFlowName.trim() || creating"
              class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
