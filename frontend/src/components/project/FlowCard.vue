<script setup>
import { ref, onMounted } from 'vue'
import FlowPreview from '../flow/FlowPreview.vue'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
  flow: {
    type: Object,
    required: true
  }
})

const graph = ref(null)
const loading = ref(true)

onMounted(async () => {
  if (!GoApp) return
  try {
    const result = await GoApp.GetFlowGraph(props.ctx, props.ns, props.projectName, props.flow.resourceName)
    graph.value = result
  } catch (err) {
    console.error('Failed to load flow graph:', err)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="flow-card bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden hover:shadow-md transition-shadow">
    <div class="flow-preview h-40 bg-gray-50 dark:bg-gray-800">
      <div v-if="loading" class="flex items-center justify-center h-full">
        <span class="text-xs text-gray-400">Loading...</span>
      </div>
      <FlowPreview v-else-if="graph" :graph="graph" :id="flow.resourceName" />
      <div v-else class="flex items-center justify-center h-full">
        <span class="text-xs text-gray-400">No preview</span>
      </div>
    </div>
    <div class="p-3">
      <h3 class="text-sm font-medium text-gray-900 dark:text-white truncate">
        {{ flow.name }}
      </h3>
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
        {{ flow.nodeCount }} nodes
      </p>
    </div>
  </div>
</template>
