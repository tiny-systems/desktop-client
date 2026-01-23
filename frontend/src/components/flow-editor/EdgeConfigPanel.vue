<script setup>
import { ref, computed, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import { debounce } from 'lodash'

const props = defineProps({
  edge: Object
})

const emit = defineEmits(['error'])

const flowStore = useFlowStore()
const saving = ref(false)

// Edge info
const edgeSource = computed(() => props.edge?.source || '')
const edgeSourceHandle = computed(() => props.edge?.sourceHandle || '')
const edgeTarget = computed(() => props.edge?.target || '')
const edgeTargetHandle = computed(() => props.edge?.targetHandle || '')

// Validation errors from edge data
const validationError = computed(() => props.edge?.data?.error || null)
const validationErrors = computed(() => {
  const errors = props.edge?.data?.errors
  if (!errors || typeof errors !== 'object') return []
  return Object.entries(errors).map(([path, message]) => ({
    path,
    message
  }))
})
const isValid = computed(() => props.edge?.data?.valid !== false)
const targetTo = computed(() => `${edgeTarget.value}:${edgeTargetHandle.value}`)

// Configuration from edge data or target handle
const configuration = computed(() => {
  // First check if edge has its own configuration
  if (props.edge?.data?.configuration) {
    const config = props.edge.data.configuration
    if (typeof config === 'string') return config
    return JSON.stringify(config, null, 2)
  }

  // Use the flowStore's selectedConfiguration getter
  const config = flowStore.selectedConfiguration
  if (!config || config === '{}') return '{}'
  if (typeof config === 'string') return config
  return JSON.stringify(config, null, 2)
})

const schema = computed(() => {
  // First check edge data
  if (props.edge?.data?.schema) {
    const sch = props.edge.data.schema
    if (typeof sch === 'string') {
      try {
        return JSON.parse(sch)
      } catch {
        return null
      }
    }
    return sch
  }

  // Use flowStore's selectedSchema getter
  const sch = flowStore.selectedSchema
  if (!sch || sch === '{}') return null
  if (typeof sch === 'string') {
    try {
      return JSON.parse(sch)
    } catch {
      return null
    }
  }
  return sch
})

// Local editor value
const editorValue = ref(configuration.value)

// Update local value when edge changes
watch(() => props.edge?.id, () => {
  editorValue.value = configuration.value
})

watch(configuration, (newVal) => {
  editorValue.value = newVal
})

// Monaco editor options
const editorOptions = {
  minimap: { enabled: false },
  lineNumbers: 'off',
  scrollBeyondLastLine: false,
  folding: false,
  fontSize: 12,
  wordWrap: 'on',
  automaticLayout: true,
  tabSize: 2,
}

// Debounced save
const debouncedSave = debounce(async (value) => {
  if (!props.edge?.id) return

  saving.value = true
  try {
    // Validate JSON
    JSON.parse(value)
    await flowStore.updateEdgeConfiguration(
      edgeSource.value,
      edgeSourceHandle.value,
      targetTo.value,
      value,
      props.edge.data?.flowID || flowStore.flowResourceName
    )
  } catch (err) {
    if (err instanceof SyntaxError) {
      // JSON parse error - don't save
      return
    }
    emit('error', `Failed to save edge configuration: ${err}`)
  } finally {
    saving.value = false
  }
}, 1000)

// Handle editor change
const handleEditorChange = (value) => {
  editorValue.value = value
  debouncedSave(value)
}

// Format JSON
const formatJson = () => {
  try {
    const parsed = JSON.parse(editorValue.value)
    editorValue.value = JSON.stringify(parsed, null, 2)
  } catch {
    // Invalid JSON, can't format
  }
}
</script>

<template>
  <div class="edge-config-panel h-full flex flex-col">
    <!-- Edge info -->
    <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 space-y-1">
      <div class="text-xs text-gray-500 dark:text-gray-400">
        <span class="font-medium">From:</span> {{ edgeSource }}:{{ edgeSourceHandle }}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400">
        <span class="font-medium">To:</span> {{ edgeTarget }}:{{ edgeTargetHandle }}
      </div>
    </div>

    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center space-x-2">
        <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Edge Configuration</span>
        <span v-if="saving" class="text-xs text-sky-500">(saving...)</span>
      </div>
      <button
        @click="formatJson"
        class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
      >
        Format
      </button>
    </div>

    <!-- Editor -->
    <div class="flex-1 overflow-hidden">
      <vue-monaco-editor
        v-model:value="editorValue"
        language="json"
        theme="vs-dark"
        :options="editorOptions"
        @change="handleEditorChange"
        class="h-full"
      />
    </div>

    <!-- Validation errors -->
    <div v-if="!isValid && (validationError || validationErrors.length > 0)" class="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
      <div class="text-sm font-medium text-red-500 mb-2">Server validation errors:</div>
      <div v-if="validationErrors.length > 0" class="space-y-1">
        <div v-for="err in validationErrors" :key="err.path" class="text-sm text-red-400">
          <span class="font-mono">{{ err.path }}</span>&nbsp;&nbsp;{{ err.message }}
        </div>
      </div>
      <div v-else-if="validationError" class="text-sm text-red-400">
        {{ validationError }}
      </div>
    </div>

    <!-- Warning message -->
    <div class="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
      <p class="text-sm text-orange-400">
        Do not store sensitive information if you plan sharing your project as a solution.
      </p>
    </div>

    <!-- Schema info -->
    <div v-if="schema" class="px-4 py-2 border-t border-gray-200 dark:border-gray-700">
      <details class="text-xs">
        <summary class="text-gray-500 dark:text-gray-400 cursor-pointer hover:text-gray-700 dark:hover:text-gray-300">
          View Schema
        </summary>
        <pre class="mt-2 p-2 bg-gray-50 dark:bg-gray-800 rounded overflow-auto max-h-32 text-gray-600 dark:text-gray-400">{{ JSON.stringify(schema, null, 2) }}</pre>
      </details>
    </div>
  </div>
</template>
