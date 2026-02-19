<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  deepLinkData: Object, // { token, api } or { legacyUrl }
  ctx: Object // { name: contextName, ns: namespace } from App.vue
})

const emit = defineEmits(['close', 'success'])

const GoApp = window.go?.main?.App

// Steps: fetch → select → import → done
const step = ref('fetch') // 'fetch' | 'select' | 'import' | 'done' | 'error' | 'no-context'
const error = ref('')
const solutionJSON = ref('')
const solutionInfo = ref(null)
const importMessage = ref('')

// Project selection only — context/namespace comes from parent
const projects = ref([])
const selectedProject = ref('')
const newProjectName = ref('')
const useNewProject = ref(false)
const loadingProjects = ref(false)

// Custom dropdown state
const openDropdown = ref('')

const toggleDropdown = (name) => {
  openDropdown.value = openDropdown.value === name ? '' : name
}

const selectProject = (name) => {
  selectedProject.value = name
  openDropdown.value = ''
}

const handleClickOutside = (e) => {
  if (!e.target.closest('.custom-dropdown')) {
    openDropdown.value = ''
  }
}

// Fetch solution JSON on mount
onMounted(async () => {
  document.addEventListener('click', handleClickOutside)

  // Check if context is selected
  if (!props.ctx?.name || !props.ctx?.ns) {
    step.value = 'no-context'
    return
  }

  try {
    let json
    if (props.deepLinkData?.token && props.deepLinkData?.api) {
      json = await GoApp.FetchSolutionExport(props.deepLinkData.token, props.deepLinkData.api)
    } else if (props.deepLinkData?.legacyUrl) {
      json = await GoApp.FetchSolutionJSON(props.deepLinkData.legacyUrl)
    } else {
      throw new Error('Missing token/api or legacy URL in deep link data')
    }
    solutionJSON.value = json

    // Parse to extract info
    const data = JSON.parse(json)
    if (!data.version || !data.tinyFlows || !data.elements) {
      throw new Error('Invalid solution format. Expected version, tinyFlows, and elements fields.')
    }

    solutionInfo.value = {
      title: data.title || 'Untitled Solution',
      description: data.description || '',
      flowCount: data.tinyFlows?.length || 0,
      nodeCount: data.elements?.filter(e => e.type === 'tinyNode')?.length || 0,
      pageCount: data.pages?.length || 0,
    }

    // Load projects for current context/namespace
    await loadProjects()

    step.value = 'select'
  } catch (e) {
    error.value = e?.message || String(e)
    step.value = 'error'
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  stopListening()
})

const loadProjects = async () => {
  loadingProjects.value = true
  projects.value = []
  selectedProject.value = ''
  try {
    const projs = await GoApp.GetProjects(props.ctx.name, props.ctx.ns)
    projects.value = projs || []
  } catch (e) {
    // Not critical — user can create a new project
  } finally {
    loadingProjects.value = false
  }
}

// Import progress listener
const startListening = () => {
  EventsOn('import:progress', (msg) => {
    importMessage.value = msg
  })
}

const stopListening = () => {
  EventsOff('import:progress')
}

const canDeploy = () => {
  if (useNewProject.value) return !!newProjectName.value.trim()
  return !!selectedProject.value
}

const deploy = async () => {
  error.value = ''
  importMessage.value = ''
  step.value = 'import'
  startListening()

  try {
    const projectName = useNewProject.value ? newProjectName.value.trim() : selectedProject.value

    if (useNewProject.value) {
      await GoApp.CreateProject(props.ctx.name, props.ctx.ns, projectName)
    }

    await GoApp.ImportProject(props.ctx.name, props.ctx.ns, projectName, solutionJSON.value)
    step.value = 'done'
    importMessage.value = importMessage.value || 'Deploy complete!'
  } catch (e) {
    error.value = e?.message || String(e)
    step.value = 'error'
  } finally {
    stopListening()
  }
}

const close = () => {
  stopListening()
  emit('close')
}

const confirmSuccess = () => {
  emit('success')
  close()
}
</script>

<template>
  <div class="fixed inset-0 z-[100] flex items-center justify-center p-4" @keydown.escape="close">
    <!-- Backdrop -->
    <div class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm" @click="close"></div>

    <!-- Modal -->
    <div class="relative transform rounded-lg bg-white text-left shadow-xl w-full max-w-2xl mx-auto dark:bg-gray-900 dark:border dark:border-gray-700 dark:text-gray-300">
      <!-- Header -->
      <div class="px-6 pt-5 pb-3 border-b border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white">Deploy Solution to Cluster</h3>
      </div>

      <!-- No context selected -->
      <div v-if="step === 'no-context'" class="px-6 py-8">
        <div class="flex flex-col items-center gap-3">
          <svg class="h-10 w-10 text-amber-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
          </svg>
          <p class="text-sm text-gray-700 dark:text-gray-300 text-center">Please select a Kubernetes context and namespace first, then try again.</p>
        </div>
        <div class="flex justify-end mt-4">
          <button
            @click="close"
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 rounded-md transition-colors"
          >
            OK
          </button>
        </div>
      </div>

      <!-- Fetching step -->
      <div v-else-if="step === 'fetch'" class="px-6 py-8">
        <div class="flex flex-col items-center gap-3">
          <svg class="animate-spin h-8 w-8 text-sky-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <p class="text-sm text-gray-600 dark:text-gray-400">Fetching solution...</p>
        </div>
      </div>

      <!-- Select target step -->
      <div v-else-if="step === 'select'" class="px-6 py-5 space-y-4">
        <!-- Solution info -->
        <div v-if="solutionInfo" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 border border-gray-200 dark:border-gray-700">
          <p class="font-medium text-gray-900 dark:text-white">{{ solutionInfo.title }}</p>
          <p v-if="solutionInfo.description" class="text-xs text-gray-500 mt-1 line-clamp-2">{{ solutionInfo.description }}</p>
          <div class="flex gap-4 mt-2 text-xs text-gray-500">
            <span>{{ solutionInfo.flowCount }} flow{{ solutionInfo.flowCount !== 1 ? 's' : '' }}</span>
            <span>{{ solutionInfo.nodeCount }} node{{ solutionInfo.nodeCount !== 1 ? 's' : '' }}</span>
            <span v-if="solutionInfo.pageCount">{{ solutionInfo.pageCount }} page{{ solutionInfo.pageCount !== 1 ? 's' : '' }}</span>
          </div>
        </div>

        <!-- Current context info -->
        <div class="text-xs text-gray-500 dark:text-gray-400">
          Deploying to <span class="font-medium text-gray-700 dark:text-gray-300">{{ ctx.name }}</span> / <span class="font-medium text-gray-700 dark:text-gray-300">{{ ctx.ns }}</span>
        </div>

        <!-- Project selector -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Project</label>
          <div v-if="loadingProjects" class="text-sm text-gray-500">Loading projects...</div>
          <template v-else>
            <div class="flex items-center gap-2 mb-2">
              <button
                @click="useNewProject = false"
                :class="['px-3 py-1 text-xs rounded-full border transition-colors', !useNewProject ? 'bg-sky-100 dark:bg-sky-900 border-sky-300 dark:border-sky-700 text-sky-700 dark:text-sky-300' : 'border-gray-300 dark:border-gray-600 text-gray-500']"
              >
                Existing
              </button>
              <button
                @click="useNewProject = true"
                :class="['px-3 py-1 text-xs rounded-full border transition-colors', useNewProject ? 'bg-sky-100 dark:bg-sky-900 border-sky-300 dark:border-sky-700 text-sky-700 dark:text-sky-300' : 'border-gray-300 dark:border-gray-600 text-gray-500']"
              >
                New Project
              </button>
            </div>
            <div v-if="!useNewProject" class="custom-dropdown relative">
              <button
                type="button"
                @click="toggleDropdown('project')"
                class="w-full flex items-center justify-between px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-white hover:border-gray-400 dark:hover:border-gray-500 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-sky-500 transition-colors"
              >
                <span :class="selectedProject ? '' : 'text-gray-400'">{{ (selectedProject && projects.find(p => p.name === selectedProject)?.title) || selectedProject || 'Select project...' }}</span>
                <svg class="w-4 h-4 text-gray-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
              </button>
              <div v-if="openDropdown === 'project'" class="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-md shadow-lg max-h-48 overflow-auto">
                <div
                  v-for="p in projects"
                  :key="p.name"
                  @click="selectProject(p.name)"
                  :class="['px-3 py-2 text-sm cursor-pointer', p.name === selectedProject ? 'bg-sky-50 dark:bg-sky-900/30 text-sky-700 dark:text-sky-300' : 'text-gray-900 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700']"
                >
                  {{ p.title || p.name }}
                </div>
                <div v-if="projects.length === 0" class="px-3 py-2 text-sm text-gray-400">
                  No projects found
                </div>
              </div>
            </div>
            <input
              v-else
              v-model="newProjectName"
              type="text"
              placeholder="New project name"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-sky-500"
            />
          </template>
        </div>

        <!-- Buttons -->
        <div class="flex justify-end gap-2 pt-2">
          <button
            @click="close"
            type="button"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="deploy"
            type="button"
            :disabled="!canDeploy()"
            class="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Deploy
          </button>
        </div>
      </div>

      <!-- Importing step -->
      <div v-else-if="step === 'import'" class="px-6 py-8">
        <div class="flex flex-col items-center gap-3">
          <svg class="animate-spin h-8 w-8 text-sky-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <p class="text-sm text-gray-600 dark:text-gray-400">{{ importMessage || 'Deploying solution...' }}</p>
        </div>
      </div>

      <!-- Done step -->
      <div v-else-if="step === 'done'" class="px-6 py-8">
        <div class="flex flex-col items-center gap-3">
          <svg class="h-10 w-10 text-green-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="text-sm text-gray-700 dark:text-gray-300">{{ importMessage }}</p>
        </div>
        <div class="flex justify-end mt-4">
          <button
            @click="confirmSuccess"
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 rounded-md transition-colors"
          >
            OK
          </button>
        </div>
      </div>

      <!-- Error step -->
      <div v-else-if="step === 'error'" class="px-6 py-6">
        <div class="flex flex-col items-center gap-3">
          <svg class="h-10 w-10 text-red-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
          </svg>
          <p class="text-sm text-red-600 dark:text-red-400 text-center max-w-sm">{{ error }}</p>
        </div>
        <div class="flex justify-end gap-2 mt-4">
          <button
            @click="step = 'select'; error = ''"
            type="button"
            v-if="solutionJSON"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Back
          </button>
          <button
            @click="close"
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 rounded-md transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
