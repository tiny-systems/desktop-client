<template>
  <Menu as="div" class="relative inline-block text-left">
    <MenuButton
      class="inline-flex items-center gap-1 text-lg leading-tight font-medium text-gray-900 dark:text-gray-300 px-2 hover:text-sky-600 dark:hover:text-sky-400 cursor-pointer transition-colors"
    >
      <span class="truncate max-w-64">{{ currentFlowName }}</span>
      <ChevronDownIcon class="w-4 h-4 flex-shrink-0" />
    </MenuButton>

    <transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
    >
      <MenuItems
        class="absolute left-0 z-50 mt-2 w-80 max-h-[70vh] overflow-y-auto origin-top-left rounded-lg shadow-lg ring-1 ring-black dark:ring-gray-700 ring-opacity-5 focus:outline-none bg-white dark:bg-gray-900"
      >
        <div class="p-2 pb-2">
          <!-- Action buttons -->
          <div class="flex items-center justify-between mb-3 pb-2 border-b border-gray-200 dark:border-gray-700">
            <button
              @click.stop="loadFlows"
              :disabled="loading"
              class="flex items-center gap-1 text-xs text-gray-500 hover:text-sky-600 dark:text-gray-400 dark:hover:text-sky-400 transition-colors disabled:opacity-50"
              title="Refresh flow list"
            >
              <ArrowPathIcon :class="['w-4 h-4', loading ? 'animate-spin' : '']" />
              <span>Refresh</span>
            </button>
            <button
              @click.stop="openNewFlowInput"
              class="flex items-center gap-1 text-xs text-sky-600 hover:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300 transition-colors"
            >
              <PlusIcon class="w-4 h-4" />
              <span>New Flow</span>
            </button>
          </div>

          <!-- New flow input -->
          <div v-if="showNewFlowInput" class="mb-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
            <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Flow Name</label>
            <div class="flex gap-2">
              <input
                ref="newFlowInputRef"
                v-model="newFlowName"
                @keyup.enter="createNewFlow"
                @keyup.escape="cancelNewFlow"
                type="text"
                placeholder="Enter flow name"
                class="flex-1 text-sm px-2 py-1 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-sky-500"
                :disabled="creatingFlow"
              />
              <button
                @click.stop="createNewFlow"
                :disabled="!newFlowName.trim() || creatingFlow"
                class="p-1 text-green-600 hover:text-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                title="Create"
              >
                <CheckIcon class="w-5 h-5" />
              </button>
              <button
                @click.stop="cancelNewFlow"
                :disabled="creatingFlow"
                class="p-1 text-gray-400 hover:text-gray-500 disabled:opacity-50"
                title="Cancel"
              >
                <XMarkIcon class="w-5 h-5" />
              </button>
            </div>
          </div>

          <div v-if="loading" class="text-center py-4 text-sm text-gray-500">
            Loading flows...
          </div>
          <div v-else-if="flows.length === 0" class="text-center py-4 text-sm text-gray-500">
            No flows in this project
          </div>
          <div v-else class="space-y-3">
            <MenuItem v-for="flow in flows" :key="flow.resourceName" v-slot="{ active }">
              <button
                @click="switchToFlow(flow)"
                :class="[
                  'w-full rounded-md transition-colors',
                  active ? 'bg-gray-100 dark:bg-gray-800' : '',
                  isCurrentFlow(flow) ? 'ring-2 ring-sky-500' : ''
                ]"
              >
                <div class="flex flex-col">
                  <!-- Flow preview -->
                  <div
                    class="h-28 w-full overflow-hidden rounded-t-md bg-gray-100 dark:bg-gray-800"
                  >
                    <FlowPreview
                      v-if="flow.graph && hasNodes(flow.graph)"
                      :graph="flow.graph"
                      :id="'switcher-' + flow.resourceName"
                    />
                    <div v-else class="h-full flex items-center justify-center text-gray-400 text-xs">
                      No preview
                    </div>
                  </div>
                  <!-- Flow info -->
                  <div class="px-3 py-2 text-left bg-gray-50 dark:bg-gray-800/50 rounded-b-md">
                    <div
                      :class="[
                        'text-sm font-medium truncate',
                        isCurrentFlow(flow) ? 'text-sky-600 dark:text-sky-400' : 'text-gray-900 dark:text-gray-200'
                      ]"
                    >
                      {{ flow.name }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{ flow.resourceName }}
                    </div>
                  </div>
                </div>
              </button>
            </MenuItem>
          </div>
        </div>
      </MenuItems>
    </transition>
  </Menu>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { ChevronDownIcon, PlusIcon, CheckIcon, XMarkIcon } from '@heroicons/vue/24/solid'
import { ArrowPathIcon } from '@heroicons/vue/24/outline'
import FlowPreview from '../flow/FlowPreview.vue'

const GoApp = window.go?.main?.App

const props = defineProps({
  currentFlowName: {
    type: String,
    default: ''
  },
  currentFlowResourceName: {
    type: String,
    default: ''
  },
  projectName: {
    type: String,
    required: true
  },
  contextName: {
    type: String,
    required: true
  },
  namespace: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['switch', 'new-flow'])

const flows = ref([])
const loading = ref(false)
const showNewFlowInput = ref(false)
const newFlowName = ref('')
const creatingFlow = ref(false)
const newFlowInputRef = ref(null)

const isCurrentFlow = (flow) => {
  return flow.resourceName === props.currentFlowResourceName
}

const hasNodes = (graph) => {
  const nodes = graph?.nodes || []
  return nodes.length > 0
}

const loadFlows = async () => {
  if (!GoApp || !props.projectName) return

  loading.value = true
  try {
    const result = await GoApp.GetFlowsWithGraphs(props.contextName, props.namespace, props.projectName)
    flows.value = result || []
  } catch (err) {
    console.error('Failed to load flows:', err)
    flows.value = []
  } finally {
    loading.value = false
  }
}

const switchToFlow = (flow) => {
  if (isCurrentFlow(flow)) return
  emit('switch', flow.resourceName)
}

const openNewFlowInput = async () => {
  showNewFlowInput.value = true
  newFlowName.value = ''
  await nextTick()
  newFlowInputRef.value?.focus()
}

const cancelNewFlow = () => {
  showNewFlowInput.value = false
  newFlowName.value = ''
}

const createNewFlow = async () => {
  if (!newFlowName.value.trim() || !GoApp) return

  creatingFlow.value = true
  try {
    const newFlow = await GoApp.CreateFlow(props.contextName, props.namespace, props.projectName, newFlowName.value.trim())
    showNewFlowInput.value = false
    newFlowName.value = ''
    await loadFlows()
    if (newFlow?.resourceName) {
      emit('switch', newFlow.resourceName)
    }
  } catch (err) {
    console.error('Failed to create flow:', err)
  } finally {
    creatingFlow.value = false
  }
}

onMounted(() => {
  loadFlows()
})

watch(() => props.projectName, () => {
  loadFlows()
})
</script>
