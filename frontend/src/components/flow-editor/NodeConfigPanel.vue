<script setup>
import { ref, computed, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'
import { useMonaco } from '@guolao/vue-monaco-editor'
import { debounce } from 'lodash'

const props = defineProps({
  node: Object
})

const emit = defineEmits(['error'])

const flowStore = useFlowStore()
const saving = ref(false)

// Get settings handle for configuration
const settingsHandle = computed(() => {
  const handles = props.node?.data?.handles || []
  return handles.find(h => h.id === '_settings')
})

// Node error from status
const nodeError = computed(() => props.node?.data?.error || null)

// Configuration and schema
const configuration = computed(() => {
  if (!settingsHandle.value) return '{}'
  const config = settingsHandle.value.configuration
  if (!config) return '{}'
  if (typeof config === 'string') return config
  return JSON.stringify(config, null, 2)
})

const schema = computed(() => {
  if (!settingsHandle.value) return null
  const sch = settingsHandle.value.schema
  if (!sch) return null
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

// Update local value when node changes
watch(() => props.node?.id, () => {
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
  if (!props.node?.id) return

  saving.value = true
  try {
    // Validate JSON
    JSON.parse(value)
    await flowStore.updateNodeConfiguration(
      props.node.id,
      '_settings',
      value,
      schema.value ? JSON.stringify(schema.value) : ''
    )
  } catch (err) {
    if (err instanceof SyntaxError) {
      // JSON parse error - don't save, just ignore
      return
    }
    emit('error', `Failed to save configuration: ${err}`)
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
  <div class="node-config-panel h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center space-x-2">
        <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Settings Configuration</span>
        <span v-if="saving" class="text-xs text-sky-500">(saving...)</span>
      </div>
      <button
        @click="formatJson"
        class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
      >
        Format
      </button>
    </div>

    <!-- No settings port -->
    <div v-if="!settingsHandle" class="flex-1 flex items-center justify-center">
      <p class="text-sm text-gray-400 dark:text-gray-500 px-4 text-center">
        This node does not have a settings port to configure.
      </p>
    </div>

    <!-- Editor -->
    <div v-else class="flex-1 overflow-hidden">
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
    <div v-if="nodeError" class="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
      <div class="text-sm font-medium text-red-500 mb-2">Server validation errors:</div>
      <div class="text-sm text-red-400">
        {{ nodeError }}
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
