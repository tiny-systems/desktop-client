<script setup>
import { ref, onMounted } from 'vue'
import FlowCard from './FlowCard.vue'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  projectName: String,
})

const emit = defineEmits(['error'])

const flows = ref([])
const loading = ref(true)

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

onMounted(() => {
  loadFlows()
})
</script>

<template>
  <div class="flows-tab h-full overflow-auto p-4">
    <div v-if="loading" class="flex items-center justify-center h-64">
      <span class="text-gray-500 dark:text-gray-400">Loading flows...</span>
    </div>
    <div v-else-if="flows.length === 0" class="flex items-center justify-center h-64">
      <span class="text-gray-500 dark:text-gray-400">No flows in this project</span>
    </div>
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <FlowCard
        v-for="flow in flows"
        :key="flow.resourceName"
        :ctx="ctx"
        :ns="ns"
        :project-name="projectName"
        :flow="flow"
      />
    </div>
  </div>
</template>
