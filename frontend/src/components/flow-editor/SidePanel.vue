<script setup>
import { ref, computed, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import {
  Menu,
  MenuButton,
  MenuItem,
  MenuItems,
  Listbox,
  ListboxButton,
  ListboxLabel,
  ListboxOption,
  ListboxOptions
} from '@headlessui/vue'
import {
  EllipsisVerticalIcon,
  ArrowsUpDownIcon,
  CheckIcon
} from '@heroicons/vue/24/solid'
import {
  PencilIcon,
  TrashIcon,
  Cog6ToothIcon,
  ArrowPathIcon
} from '@heroicons/vue/24/outline'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'
import SchemaForm from './SchemaForm.vue'

const emit = defineEmits(['close', 'error', 'rename', 'settings', 'delete'])

// Dynamic width class for side panel - half screen width
const panelWidthClass = 'w-1/2'

const flowStore = useFlowStore()

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

// Port inspection
const inspect = ref(null)
const inspectReady = ref(false)
const selectedHandleId = ref(null)

// Get handles for port selector (include _settings, exclude only _control)
const selectedNodeHandles = computed(() => {
  const handles = selectedNode.value?.data?.handles || []
  return handles.filter(h => h.id !== '_control')
})

// Get selected handle
const selectedHandle = computed(() => {
  if (!selectedNodeHandles.value.length) return null
  const sel = selectedHandleId.value || selectedNode.value?.data?.trace?.port
  return selectedNodeHandles.value.find(h => h.id === sel) || selectedNodeHandles.value[0]
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

const settingsConfiguration = computed(() => {
  if (!settingsHandle.value?.configuration) return '{}'
  const config = settingsHandle.value.configuration
  return typeof config === 'string' ? config : JSON.stringify(config, null, 2)
})

// Settings schema for the form
const settingsSchema = computed(() => {
  if (!settingsHandle.value?.schema) return null
  const schema = settingsHandle.value.schema
  if (typeof schema === 'string') {
    try {
      return JSON.parse(schema)
    } catch {
      return null
    }
  }
  return schema
})

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

// Watch settings changes for raw editor
watch(settingsConfiguration, (val) => {
  editorValue.value = val
}, { immediate: true })

// Watch settings changes for form
watch(settingsConfigObject, (val) => {
  formValue.value = { ...val }
}, { immediate: true, deep: true })

// Update form value handler
const updateFormValue = (newValue) => {
  formValue.value = newValue
  editorValue.value = JSON.stringify(newValue, null, 2)
}

// Watch selected handle for port inspection
watch(selectedHandle, async (handle) => {
  if (!handle || !selectedNode.value) {
    inspect.value = null
    return
  }
  inspectReady.value = false
  try {
    const data = await flowStore.inspectNodePort(selectedNode.value.id, handle.id)
    inspect.value = data
  } catch (e) {
    inspect.value = { error: e.message || String(e) }
  } finally {
    inspectReady.value = true
  }
}, { immediate: true })

// Watch node change to reset tab
watch(() => selectedNode.value?.id, () => {
  setCurrentTab('status')
  selectedHandleId.value = null
})

// Node info expiring check
const selectedNodeExpiring = computed(() => {
  if (!selectedNode.value?.data?.lastUpdateTime) return false
  const lastUpdate = new Date(selectedNode.value.data.lastUpdateTime).getTime()
  return (Date.now() - lastUpdate) > 10 * 60 * 1000
})

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

// Edge configuration
const edgeConfiguration = computed(() => {
  if (!selectedEdge.value?.data?.configuration) return '{}'
  const config = selectedEdge.value.data.configuration
  return typeof config === 'string' ? config : JSON.stringify(config, null, 2)
})

// Edge configuration as object for form
const edgeConfigObject = computed(() => {
  const config = selectedEdge.value?.data?.configuration
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

const edgeEditorValue = ref('{}')
const edgeFormValue = ref({})

watch(edgeConfiguration, (val) => {
  edgeEditorValue.value = val
}, { immediate: true })

watch(edgeConfigObject, (val) => {
  edgeFormValue.value = { ...val }
}, { immediate: true, deep: true })

// Update edge form value handler
const updateEdgeFormValue = (newValue) => {
  edgeFormValue.value = newValue
  edgeEditorValue.value = JSON.stringify(newValue, null, 2)
}

// Actions
const handleDeselect = () => {
  flowStore.selectElement(null)
}

const handleDeleteEdge = async () => {
  if (!selectedEdge.value) return
  try {
    await flowStore.disconnectNodes(selectedEdge.value.source, selectedEdge.value.id)
  } catch (err) {
    emit('error', `Failed to delete edge: ${err}`)
  }
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

// Save configuration
const saveConfiguration = async () => {
  if (!selectedNode.value || !settingsHandle.value) return
  saving.value = true
  try {
    await flowStore.updateNodeConfiguration(
      selectedNode.value.id,
      '_settings',
      editorValue.value,
      ''
    )
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
    await flowStore.updateEdgeConfiguration(
      selectedEdge.value.source,
      selectedEdge.value.sourceHandle,
      targetTo,
      edgeEditorValue.value,
      selectedEdge.value.data?.flowID || flowStore.flowResourceName
    )
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
    :class="['relative text-sm flex flex-col dark:text-gray-300 flex-shrink-0 bg-gray-50 dark:bg-black border-l border-gray-200 dark:border-gray-700 h-full', panelWidthClass]"
  >
    <form @submit.prevent="saveEdgeConfiguration" class="bg-white dark:bg-gray-900 shadow rounded-lg text-xs h-full flex flex-col">
      <!-- Tab header -->
      <nav class="relative border-b border-gray-200 dark:border-gray-700 flex divide-x divide-gray-200 dark:divide-gray-700">
        <a
          href="#"
          @click.prevent="handleDeselect"
          class="text-gray-600 dark:text-gray-300 group relative min-w-0 flex-1 overflow-hidden bg-white dark:bg-gray-900 py-2 px-2 text-sm font-medium text-center hover:bg-gray-50 dark:hover:bg-gray-800 focus:z-10 whitespace-nowrap"
        >
          <span>{{ edgeSourceLabel }} &rarr; {{ edgeTargetLabel }} {{ selectedEdge.targetHandle?.toUpperCase() }}</span>
          <span class="bg-sky-500 absolute inset-x-0 bottom-0 h-0.5"></span>
        </a>
      </nav>

      <!-- Edge info -->
      <div class="bg-white dark:bg-gray-900 shadow rounded-lg text-xs relative px-2 py-2 flex justify-start items-center">
        <button
          type="button"
          @click="handleDeleteEdge"
          class="text-red-400 border-red-400 border px-3 py-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20"
        >
          Delete Edge
        </button>
        <div class="px-2">
          <h3 class="font-semibold dark:text-gray-300 text-gray-600">EdgeID</h3>
          <p class="font-light dark:text-gray-400 text-gray-500 text-xs truncate max-w-48">{{ selectedEdge.id }}</p>
        </div>
      </div>

      <!-- Configuration form or JSON editor -->
      <div class="flex-1 overflow-y-auto p-2">
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Configuration</label>
        <!-- Schema-based form when schema is available -->
        <SchemaForm
          v-if="edgeSchema && (edgeSchema.properties || edgeSchema.type || edgeSchema.$ref)"
          :schema="edgeSchema"
          :model-value="edgeFormValue"
          @update:model-value="updateEdgeFormValue"
          :readonly="selectedEdge.data?.blocked"
        />
        <!-- Fallback to raw JSON editor -->
        <textarea
          v-else
          v-model="edgeEditorValue"
          class="w-full h-48 p-2 text-xs font-mono bg-gray-50 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded focus:ring-sky-500 focus:border-sky-500 dark:text-gray-300"
          placeholder="{}"
        />
        <!-- Save button -->
        <div class="pt-3 text-right">
          <button
            type="submit"
            :disabled="saving || selectedEdge.data?.blocked"
            class="px-4 py-2 text-xs font-medium rounded-md text-sky-700 bg-sky-100 hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-sky-500 dark:bg-sky-900 dark:hover:bg-sky-800 dark:text-sky-300 disabled:opacity-50"
          >
            {{ saving ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </form>
  </aside>

  <!-- Multiple nodes selected -->
  <aside
    v-else-if="selectedNodes.length > 1"
    :class="['relative text-sm flex flex-col flex-shrink-0 bg-gray-50 dark:bg-black border-l border-gray-200 dark:border-gray-700 h-full', panelWidthClass]"
  >
    <div class="p-4 dark:text-gray-300">Multiple nodes selected: {{ selectedNodes.length }}</div>
  </aside>

  <!-- Single node selected -->
  <aside
    v-else-if="selectedNode"
    :class="['relative text-sm flex flex-col flex-shrink-0 bg-gray-50 dark:bg-black border-l border-gray-200 dark:border-gray-700 h-full', panelWidthClass]"
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

      <!-- Settings form -->
      <form v-if="settingsHandle" @submit.prevent="saveConfiguration" class="flex-1 overflow-y-auto bg-white dark:bg-gray-900">
        <div class="p-3">
          <!-- Schema-based form when schema is available (any valid schema, including $ref) -->
          <SchemaForm
            v-if="settingsSchema && (settingsSchema.properties || settingsSchema.type || settingsSchema.$ref)"
            :schema="settingsSchema"
            :model-value="formValue"
            @update:model-value="updateFormValue"
            :readonly="selectedNode.data?.blocked"
          />
          <!-- Fallback to raw JSON editor -->
          <div v-else>
            <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Settings (JSON)</label>
            <textarea
              v-model="editorValue"
              class="w-full h-64 p-2 text-xs font-mono bg-gray-50 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded focus:ring-sky-500 focus:border-sky-500 dark:text-gray-300"
              placeholder="{}"
            />
          </div>
          <!-- Save button -->
          <div class="pt-3 text-right">
            <button
              type="submit"
              :disabled="saving || selectedNode.data?.blocked"
              class="px-4 py-2 text-xs font-medium rounded-md text-sky-700 bg-sky-100 hover:bg-sky-200 focus:outline-none focus:ring-2 focus:ring-sky-500 dark:bg-sky-900 dark:hover:bg-sky-800 dark:text-sky-300 disabled:opacity-50"
            >
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
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
                <p v-if="selectedNode.data?.description">{{ selectedNode.data.description }}</p>
                <p>Module: <span class="font-semibold">{{ selectedNode.data?.module }}</span></p>
                <p>Component: <span class="font-semibold">{{ selectedNode.data?.component }}</span></p>
                <p :class="selectedNodeExpiring ? 'text-red-500' : ''">
                  Last update:
                  <span v-if="selectedNode.data?.lastUpdateTime" class="font-semibold">
                    {{ new Date(selectedNode.data.lastUpdateTime).toLocaleString() }}
                  </span>
                  <span v-else class="font-semibold text-red-500">Never</span>
                </p>
              </div>
            </div>

            <!-- Actions menu -->
            <Menu as="div" class="ml-3 relative inline-block text-left">
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
                <MenuItems class="origin-top-right absolute z-40 right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none dark:ring-gray-600 dark:bg-gray-900">
                  <div class="py-1">
                    <MenuItem v-slot="{ active }">
                      <button
                        type="button"
                        @click="handleRotateNode"
                        :class="[active ? 'bg-gray-100 text-gray-900 dark:bg-gray-700' : 'text-gray-700 dark:text-gray-300', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <ArrowPathIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Rotate</span>
                      </button>
                    </MenuItem>
                    <MenuItem v-slot="{ active }">
                      <button
                        type="button"
                        @click="handleRenameNode"
                        :class="[active ? 'bg-gray-100 text-gray-900 dark:bg-gray-700' : 'text-gray-700 dark:text-gray-300', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <PencilIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Rename</span>
                      </button>
                    </MenuItem>
                    <MenuItem v-slot="{ active }">
                      <button
                        type="button"
                        @click="handleNodeSettings"
                        :class="[active ? 'bg-gray-100 text-gray-900 dark:bg-gray-700' : 'text-gray-700 dark:text-gray-300', 'w-full flex px-4 py-2 text-sm']"
                      >
                        <Cog6ToothIcon class="mr-2 h-4 w-4 text-gray-400" />
                        <span>Settings</span>
                      </button>
                    </MenuItem>
                    <MenuItem v-slot="{ active }">
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
        <div v-if="controlHandle && controlHandleSchema" class="bg-white dark:bg-gray-900 shadow rounded text-xs m-1 p-2">
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Control</label>
          <SchemaForm
            :schema="controlHandleSchema"
            :model-value="controlFormValue"
            @update:model-value="updateControlFormValue"
            @action="handleControlUpdate"
          />
          <div v-if="controlHandle.error" class="mt-2 p-2 text-xs bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
            {{ controlHandle.error }}
          </div>
        </div>

        <!-- Port selector -->
        <Listbox
          v-if="selectedNodeHandles.length > 0"
          as="div"
          :model-value="selectedHandleId"
          @update:model-value="selectedHandleId = $event"
          class="px-2"
        >
          <div class="mt-1 relative">
            <ListboxLabel class="block pb-1 pt-1 text-xs font-medium text-gray-900 dark:text-gray-300 text-left">
              Port data preview
            </ListboxLabel>
            <ListboxButton
              v-if="selectedHandle"
              class="bg-white dark:bg-gray-900 relative w-full border border-gray-300 dark:border-gray-700 rounded-md shadow-sm pl-3 pr-10 py-2 text-left cursor-default focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 text-sm dark:text-gray-300"
            >
              <span class="block truncate text-xs">{{ selectedHandle.label || selectedHandle.id }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <ArrowsUpDownIcon class="h-4 w-4 text-gray-400" />
              </span>
            </ListboxButton>
            <transition
              leave-active-class="transition ease-in duration-100"
              leave-from-class="opacity-100"
              leave-to-class="opacity-0"
            >
              <ListboxOptions class="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none text-sm">
                <ListboxOption
                  v-for="handle in selectedNodeHandles"
                  :key="handle.id"
                  :value="handle.id"
                  v-slot="{ active, selected }"
                  as="template"
                >
                  <li :class="[active ? 'text-white bg-sky-600' : 'text-gray-900 dark:text-gray-300', 'text-xs cursor-default select-none relative py-2 pl-3 pr-9']">
                    <span :class="[selected ? 'font-semibold' : 'font-normal', 'block truncate']">
                      {{ handle.label || handle.id }}
                    </span>
                    <span v-if="selected" :class="[active ? 'text-white' : 'text-sky-600', 'absolute inset-y-0 right-0 flex items-center pr-4']">
                      <CheckIcon class="h-4 w-4" />
                    </span>
                  </li>
                </ListboxOption>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>

        <!-- Port data display -->
        <div class="bg-white dark:bg-gray-900 dark:text-gray-300 shadow rounded text-xs overflow-y-auto m-1 min-h-48">
          <div class="p-2">
            <div v-if="!inspectReady" class="text-center text-gray-400 py-4">Loading...</div>
            <div v-else-if="inspect?.dataError" class="text-center text-red-400 py-4 text-xs">{{ inspect.dataError }}</div>
            <VueJsonPretty
              v-else-if="inspect?.data !== undefined"
              :data="inspect.data"
              :deep="3"
              :show-length="true"
              class="text-xs"
            />
            <div v-else class="text-center text-gray-400 py-4">No data</div>
          </div>
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
    :class="['relative text-sm flex flex-col dark:text-gray-300 flex-shrink-0 bg-gray-50 dark:bg-black border-l border-gray-200 dark:border-gray-700 h-full', panelWidthClass]"
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
</template>

<style scoped>
/* VueJsonPretty size override */
:deep(.vjs-tree) {
  font-size: 11px !important;
  line-height: 1.4 !important;
}
</style>
