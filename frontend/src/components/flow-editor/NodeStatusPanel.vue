<script setup>
import { ref, computed, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import { CheckCircleIcon, ExclamationCircleIcon, ClockIcon, ArrowPathIcon } from '@heroicons/vue/24/outline'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'

const props = defineProps({
  node: Object
})

const emit = defineEmits(['error'])

const flowStore = useFlowStore()
const selectedPortId = ref('')
const portData = ref(null)
const loadingPort = ref(false)

// Node info computed
const nodeLabel = computed(() => props.node?.data?.label || props.node?.id || '')
const nodeComponent = computed(() => props.node?.data?.component || '')
const nodeModule = computed(() => props.node?.data?.module || '')
const nodeStatus = computed(() => props.node?.data?.status || 'unknown')
const nodeError = computed(() => props.node?.data?.error || false)
const nodeDescription = computed(() => props.node?.data?.description || '')
const nodeInfo = computed(() => props.node?.data?.info || '')
const hasDashboard = computed(() => props.node?.data?.dashboard || false)

// Ports (include _settings, exclude only _control)
const ports = computed(() => {
  const handles = props.node?.data?.handles || []
  return handles.filter(h => h.id !== '_control')
})

// Settings and control handles
const settingsHandle = computed(() => {
  const handles = props.node?.data?.handles || []
  return handles.find(h => h.id === '_settings')
})

const controlHandle = computed(() => {
  const handles = props.node?.data?.handles || []
  return handles.find(h => h.id === '_control')
})

// Status icon component
const statusClass = computed(() => {
  if (nodeError.value) return 'text-red-500'
  if (nodeStatus.value === 'running' || nodeStatus.value === 'ready') return 'text-green-500'
  if (nodeStatus.value === 'pending') return 'text-yellow-500'
  return 'text-gray-400'
})

// Load port data when selected
const loadPortData = async (portId) => {
  if (!portId || !props.node?.id) return

  loadingPort.value = true
  try {
    portData.value = await flowStore.inspectNodePort(props.node.id, portId)
  } catch (err) {
    emit('error', `Failed to load port data: ${err}`)
    portData.value = null
  } finally {
    loadingPort.value = false
  }
}

// Select port
const selectPort = (portId) => {
  selectedPortId.value = portId
  loadPortData(portId)
}

// Watch node changes to reset selection
watch(() => props.node?.id, () => {
  selectedPortId.value = ''
  portData.value = null
})

// Toggle dashboard
const toggleDashboard = async () => {
  if (!props.node?.id) return
  try {
    await flowStore.toggleNodeDashboard(props.node.id, !hasDashboard.value)
  } catch (err) {
    emit('error', `Failed to toggle dashboard: ${err}`)
  }
}

// Rotate node
const rotateNode = async () => {
  if (!props.node?.id) return
  try {
    await flowStore.rotateNode(props.node.id)
  } catch (err) {
    emit('error', `Failed to rotate node: ${err}`)
  }
}
</script>

<template>
  <div class="node-status-panel p-4 space-y-4">
    <!-- Node info header -->
    <div class="space-y-2">
      <div class="flex items-start justify-between">
        <div>
          <h3 class="font-medium text-gray-900 dark:text-white">{{ nodeLabel }}</h3>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ nodeComponent }}</p>
        </div>
        <div class="flex items-center space-x-1">
          <component
            :is="nodeError ? ExclamationCircleIcon : nodeStatus === 'running' || nodeStatus === 'ready' ? CheckCircleIcon : ClockIcon"
            :class="['w-5 h-5', statusClass]"
          />
        </div>
      </div>
      <p v-if="nodeDescription" class="text-sm text-gray-600 dark:text-gray-400">
        {{ nodeDescription }}
      </p>
      <p v-if="nodeInfo" class="text-xs text-gray-400 dark:text-gray-500">
        {{ nodeInfo }}
      </p>
    </div>

    <!-- Quick actions -->
    <div class="flex items-center space-x-2 pt-2 border-t border-gray-200 dark:border-gray-700">
      <button
        @click="rotateNode"
        class="flex items-center space-x-1 px-2 py-1 text-xs text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
        title="Rotate node"
      >
        <ArrowPathIcon class="w-4 h-4" />
        <span>Rotate</span>
      </button>
      <button
        v-if="controlHandle"
        @click="toggleDashboard"
        :class="[
          'flex items-center space-x-1 px-2 py-1 text-xs rounded transition-colors',
          hasDashboard
            ? 'text-sky-600 dark:text-sky-400 bg-sky-50 dark:bg-sky-900/20'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
        ]"
        title="Toggle dashboard widget"
      >
        <span>{{ hasDashboard ? 'On Dashboard' : 'Add to Dashboard' }}</span>
      </button>
    </div>

    <!-- Ports list -->
    <div v-if="ports.length > 0" class="space-y-2 pt-2 border-t border-gray-200 dark:border-gray-700">
      <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Ports</h4>
      <div class="space-y-1">
        <button
          v-for="port in ports"
          :key="port.id"
          @click="selectPort(port.id)"
          :class="[
            'w-full text-left px-3 py-2 rounded text-sm transition-colors',
            selectedPortId === port.id
              ? 'bg-sky-50 dark:bg-sky-900/20 text-sky-700 dark:text-sky-300'
              : 'hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300'
          ]"
        >
          <div class="flex items-center justify-between">
            <span>{{ port.label || port.id }}</span>
            <span :class="['text-xs', port.type === 'source' ? 'text-green-500' : 'text-blue-500']">
              {{ port.type === 'source' ? 'out' : 'in' }}
            </span>
          </div>
        </button>
      </div>
    </div>

    <!-- Port data preview -->
    <div v-if="selectedPortId" class="space-y-2 pt-2 border-t border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between">
        <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
          Port Data: {{ selectedPortId }}
        </h4>
        <button
          @click="loadPortData(selectedPortId)"
          class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded"
          title="Refresh"
        >
          <ArrowPathIcon :class="['w-4 h-4', loadingPort ? 'animate-spin' : '']" />
        </button>
      </div>
      <div v-if="loadingPort" class="text-sm text-gray-400">Loading...</div>
      <div v-else-if="portData" class="bg-gray-50 dark:bg-gray-800 rounded p-2 overflow-auto max-h-64">
        <VueJsonPretty
          :data="portData"
          :deep="2"
          :show-length="true"
          :show-line="false"
          class="text-xs"
        />
      </div>
      <div v-else class="text-sm text-gray-400">No data available</div>
    </div>
  </div>
</template>
