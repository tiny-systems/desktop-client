<template>
  <div class="fixed inset-0 z-50 overflow-y-auto" @keydown.escape="$emit('close')">
    <div class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm" @click="$emit('close')"></div>
    <div class="flex items-center justify-center min-h-full p-4">
      <div class="relative bg-white dark:bg-black dark:border dark:border-gray-800 rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:max-w-lg sm:w-full sm:p-6" @click.stop>
        <form @submit.prevent="saveSettings">
          <div>
            <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
              <Cog6ToothIcon class="h-6 w-6 text-green-600" />
            </div>
            <div class="mt-3 text-center sm:mt-5">
              <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100">
                {{ node?.data?.label || node?.id }}
              </h3>

              <!-- Shared node toggle -->
              <div class="mt-4 text-left">
                <label class="dark:text-gray-500 whitespace-nowrap flex items-center">
                  <input
                    type="checkbox"
                    v-model="sharedToggle"
                    class="w-4 h-4 text-blue-600 bg-gray-100 rounded border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600"
                  />
                  <span class="text-sm pl-2">Shared node</span>
                </label>
                <p class="text-xs dark:text-gray-300 mt-1">Shared nodes are accessible across selected flows of the project.</p>
              </div>

              <!-- Flow selection when shared -->
              <div v-if="sharedToggle" class="mt-4 text-left">
                <label class="block text-sm text-gray-500 dark:text-gray-400">Shared with other flows</label>
                <select
                  v-if="otherFlows.length > 0"
                  v-model="sharedFlows"
                  multiple
                  class="mt-1 border-sky-600 placeholder-gray-400 focus:ring-sky-600 appearance-none border rounded w-full py-3 px-3 text-gray-700 leading-tight transition duration-150 ease-in-out sm:text-sm dark:bg-gray-900 dark:text-gray-300"
                >
                  <option v-for="f in otherFlows" :key="f.resourceName" :value="f.resourceName">
                    {{ f.name }}
                  </option>
                </select>
                <p v-else class="mt-1 text-xs text-gray-400 dark:text-gray-500 py-3">No other flows in this project.</p>
              </div>

              <!-- Dashboard toggle (only if node has _control port) -->
              <div v-if="hasControlPort" class="mt-4 text-left">
                <label class="dark:text-gray-500 whitespace-nowrap flex items-center">
                  <input
                    type="checkbox"
                    v-model="dashboard"
                    class="w-4 h-4 text-blue-600 bg-gray-100 rounded border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-800 dark:border-gray-600"
                  />
                  <span class="text-sm pl-2">Add to dashboard</span>
                </label>
                <p class="text-xs dark:text-gray-300 mt-1">Display control port form on the project's dashboard.</p>
              </div>

              <!-- Advanced settings -->
              <div class="mt-4 text-left">
                <a href="#" @click.prevent="showAdvanced = !showAdvanced" class="text-sm text-sky-500 py-2 flex items-center">
                  <ChevronDownIcon v-if="!showAdvanced" class="w-4 h-4 mr-1" />
                  <ChevronUpIcon v-else class="w-4 h-4 mr-1" />
                  Advanced settings
                </a>
                <div v-if="showAdvanced" class="mt-2 space-y-3">
                  <div>
                    <label class="block text-xs font-medium text-sky-500">Module name</label>
                    <input
                      v-model="moduleName"
                      class="mt-1 appearance-none bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:text-white"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-medium text-sky-500">Module version</label>
                    <input
                      v-model="moduleVersion"
                      class="mt-1 appearance-none bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:text-white"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-medium text-sky-500">Component name</label>
                    <input
                      v-model="componentName"
                      class="mt-1 appearance-none bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2 dark:bg-gray-800 dark:border-gray-600 dark:text-white"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Buttons -->
          <div class="mt-5 sm:mt-6 grid grid-cols-2 gap-3">
            <button
              type="button"
              @click="$emit('close')"
              class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-lg border border-gray-200 text-sm font-medium px-5 py-2 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:bg-gray-600"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="saving"
              class="text-white bg-sky-600 hover:bg-sky-800 focus:ring-4 focus:outline-none focus:ring-sky-300 font-medium rounded-lg text-sm px-5 py-2 disabled:opacity-50"
            >
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Cog6ToothIcon, ChevronDownIcon, ChevronUpIcon } from '@heroicons/vue/24/outline'
import { useFlowStore } from '../../stores/flow'

const props = defineProps({
  node: {
    type: Object,
    required: true
  },
  flows: {
    type: Array,
    default: () => []
  },
  currentFlowResourceName: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['close', 'saved'])

const flowStore = useFlowStore()
const saving = ref(false)
const showAdvanced = ref(false)

// Dashboard setting
const dashboard = ref(props.node?.data?.dashboard === 'true')

// Check if node has control port
const hasControlPort = computed(() => {
  const handles = props.node?.data?.handles || []
  return handles.some(h => h.id === '_control')
})

// Shared flows
const sharedFlowsInitial = computed(() => {
  const flows = props.node?.data?.shared_with_flows
  if (!flows) return []
  return flows.split(',').filter(f => f && f.trim())
})

const sharedFlows = ref([...sharedFlowsInitial.value])

// Shared toggle
const sharedToggle = ref(sharedFlowsInitial.value.length > 0)

// Other flows (excluding current)
const otherFlows = computed(() => {
  return props.flows.filter(f => f.resourceName !== props.currentFlowResourceName)
})

// Advanced settings
const moduleName = ref(props.node?.data?.module || '')
const moduleVersion = ref(props.node?.data?.module_version || '')
const componentName = ref(props.node?.data?.component || '')

// Save settings
const saveSettings = async () => {
  saving.value = true
  try {
    const GoApp = window.go?.main?.App
    if (!GoApp) throw new Error('Wails runtime not available')

    await GoApp.UpdateNodeSettings(
      flowStore.contextName,
      flowStore.namespace,
      props.node.id,
      {
        sharedWithFlows: sharedToggle.value ? sharedFlows.value.join(',') : '',
        dashboard: dashboard.value,
        module: moduleName.value,
        moduleVersion: moduleVersion.value,
        component: componentName.value
      }
    )

    emit('saved')
    emit('close')
  } catch (err) {
    console.error('Failed to save node settings:', err)
  } finally {
    saving.value = false
  }
}
</script>
