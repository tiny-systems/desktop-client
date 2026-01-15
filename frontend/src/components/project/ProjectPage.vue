<script setup>
import { ref, onMounted } from 'vue'
import ProjectHeader from './ProjectHeader.vue'
import ProjectStatsBar from './ProjectStatsBar.vue'
import WidgetsTab from './WidgetsTab.vue'
import FlowsTab from './FlowsTab.vue'

const GoApp = window.go.main.App

const props = defineProps({
  ctx: String,
  ns: String,
  name: String,
})

const emit = defineEmits(['close'])

const projectDetails = ref(null)
const stats = ref({
  widgetsCount: 0,
  flowsCount: 0,
  nodesCount: 0
})
const activeTab = ref('widgets')
const loading = ref(true)
const error = ref('')

const loadProjectDetails = async () => {
  if (!GoApp) {
    error.value = 'Wails runtime not available'
    return
  }

  try {
    const details = await GoApp.GetProjectDetails(props.ctx, props.ns, props.name)
    projectDetails.value = details
  } catch (err) {
    error.value = `Failed to load project: ${err}`
  }
}

const loadStats = async () => {
  if (!GoApp) return

  try {
    const projectStats = await GoApp.GetProjectStats(props.ctx, props.ns, props.name)
    stats.value = projectStats
  } catch (err) {
    console.error('Failed to load stats:', err)
  }
}

const handleError = (err) => {
  error.value = err
  setTimeout(() => {
    error.value = ''
  }, 5000)
}

const handleDeleteProject = async () => {
  if (!GoApp) return

  try {
    await GoApp.DeleteProject(props.ctx, props.ns, props.name)
    emit('close')
  } catch (err) {
    handleError(`Failed to delete project: ${err}`)
  }
}

const handleRenameProject = async (newName) => {
  if (!GoApp) return

  try {
    await GoApp.RenameProject(props.ctx, props.ns, props.name, newName)
    if (projectDetails.value) {
      projectDetails.value.title = newName
    }
  } catch (err) {
    handleError(`Failed to rename project: ${err}`)
  }
}

onMounted(async () => {
  loading.value = true
  await Promise.all([loadProjectDetails(), loadStats()])
  loading.value = false
})
</script>

<template>
  <div class="project-page h-full flex flex-col bg-white dark:bg-gray-900">
    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center h-full">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-500 mx-auto mb-4"></div>
        <span class="text-gray-500 dark:text-gray-400">Loading project...</span>
      </div>
    </div>

    <!-- Main content -->
    <template v-else>
      <ProjectHeader
        :title="projectDetails?.title || name"
        :cluster-name="projectDetails?.clusterName"
        :project-name="name"
        @close="emit('close')"
        @delete-project="handleDeleteProject"
        @rename-project="handleRenameProject"
      />

      <ProjectStatsBar
        :stats="stats"
        :active-tab="activeTab"
        @select-tab="(tab) => activeTab = tab"
      />

      <!-- Error banner -->
      <div v-if="error" class="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
        <span class="text-sm text-red-600 dark:text-red-400">{{ error }}</span>
      </div>

      <!-- Tab content -->
      <div class="flex-1 overflow-hidden">
        <WidgetsTab
          v-if="activeTab === 'widgets'"
          :ctx="ctx"
          :ns="ns"
          :project-name="name"
          @error="handleError"
        />
        <FlowsTab
          v-else-if="activeTab === 'flows'"
          :ctx="ctx"
          :ns="ns"
          :project-name="name"
          @error="handleError"
        />
      </div>
    </template>
  </div>
</template>
