<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  url: String
})

const emit = defineEmits(['close', 'success'])

const GoApp = window.go?.main?.App

// Steps: fetch → select → import → done
const step = ref('fetch') // 'fetch' | 'select' | 'import' | 'done' | 'error'
const error = ref('')
const solutionJSON = ref('')
const solutionInfo = ref(null)
const importMessage = ref('')

// Context/namespace/project selection
const contexts = ref([])
const selectedContext = ref('')
const namespaces = ref([])
const selectedNamespace = ref('')
const projects = ref([])
const selectedProject = ref('')
const newProjectName = ref('')
const useNewProject = ref(false)
const loadingNamespaces = ref(false)
const loadingProjects = ref(false)

// Fetch solution JSON on mount
onMounted(async () => {
  try {
    const json = await GoApp.FetchSolutionJSON(props.url)
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

    // Load contexts
    const ctxs = await GoApp.GetKubeContexts()
    contexts.value = ctxs || []

    // Restore saved preferences
    const prefs = await GoApp.GetPreferences()
    if (prefs?.lastContext) {
      selectedContext.value = prefs.lastContext
      if (prefs.lastNamespace) {
        selectedNamespace.value = prefs.lastNamespace
      }
      await loadNamespaces(prefs.lastContext)
    }

    step.value = 'select'
  } catch (e) {
    error.value = e?.message || String(e)
    step.value = 'error'
  }
})

const loadNamespaces = async (contextName) => {
  if (!contextName) return
  loadingNamespaces.value = true
  namespaces.value = []
  projects.value = []
  selectedProject.value = ''
  try {
    const nss = await GoApp.GetNamespaces(contextName)
    namespaces.value = nss || []
    if (selectedNamespace.value && nss?.includes(selectedNamespace.value)) {
      await loadProjects(contextName, selectedNamespace.value)
    }
  } catch (e) {
    error.value = `Failed to load namespaces: ${e?.message || e}`
  } finally {
    loadingNamespaces.value = false
  }
}

const loadProjects = async (contextName, namespace) => {
  if (!contextName || !namespace) return
  loadingProjects.value = true
  projects.value = []
  selectedProject.value = ''
  try {
    const projs = await GoApp.GetProjects(contextName, namespace)
    projects.value = projs || []
  } catch (e) {
    // Not critical — user can create a new project
  } finally {
    loadingProjects.value = false
  }
}

watch(selectedContext, async (val) => {
  if (val) await loadNamespaces(val)
})

watch(selectedNamespace, async (val) => {
  if (val && selectedContext.value) await loadProjects(selectedContext.value, val)
})

// Import progress listener
const startListening = () => {
  EventsOn('import:progress', (msg) => {
    importMessage.value = msg
  })
}

const stopListening = () => {
  EventsOff('import:progress')
}

onUnmounted(stopListening)

const canDeploy = () => {
  if (!selectedContext.value || !selectedNamespace.value) return false
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
      await GoApp.CreateProject(selectedContext.value, selectedNamespace.value, projectName)
    }

    await GoApp.ImportProject(selectedContext.value, selectedNamespace.value, projectName, solutionJSON.value)
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
    <div class="relative transform rounded-lg bg-white text-left shadow-xl w-full max-w-lg mx-auto dark:bg-gray-900 dark:border dark:border-gray-700 dark:text-gray-300">
      <!-- Header -->
      <div class="px-5 pt-5 pb-3 border-b border-gray-200 dark:border-gray-700">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white">Deploy Solution to Desktop</h3>
      </div>

      <!-- Fetching step -->
      <div v-if="step === 'fetch'" class="px-5 py-8">
        <div class="flex flex-col items-center gap-3">
          <svg class="animate-spin h-8 w-8 text-sky-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <p class="text-sm text-gray-600 dark:text-gray-400">Fetching solution...</p>
        </div>
      </div>

      <!-- Select target step -->
      <div v-else-if="step === 'select'" class="px-5 py-4 space-y-4">
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

        <!-- Context selector -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Kubernetes Context</label>
          <select
            v-model="selectedContext"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
          >
            <option value="" disabled>Select context...</option>
            <option v-for="c in contexts" :key="c.name" :value="c.name">{{ c.name }}</option>
          </select>
        </div>

        <!-- Namespace selector -->
        <div v-if="selectedContext">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Namespace</label>
          <div v-if="loadingNamespaces" class="text-sm text-gray-500">Loading namespaces...</div>
          <select
            v-else
            v-model="selectedNamespace"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
          >
            <option value="" disabled>Select namespace...</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>
        </div>

        <!-- Project selector -->
        <div v-if="selectedNamespace">
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
            <select
              v-if="!useNewProject"
              v-model="selectedProject"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
            >
              <option value="" disabled>Select project...</option>
              <option v-for="p in projects" :key="p.name" :value="p.name">{{ p.title || p.name }}</option>
            </select>
            <input
              v-else
              v-model="newProjectName"
              type="text"
              placeholder="New project name"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-sky-500"
            />
          </template>
        </div>

        <!-- Buttons -->
        <div class="flex justify-end gap-2 pt-2">
          <button
            @click="close"
            type="button"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Cancel
          </button>
          <button
            @click="deploy"
            type="button"
            :disabled="!canDeploy()"
            class="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Deploy
          </button>
        </div>
      </div>

      <!-- Importing step -->
      <div v-else-if="step === 'import'" class="px-5 py-8">
        <div class="flex flex-col items-center gap-3">
          <svg class="animate-spin h-8 w-8 text-sky-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <p class="text-sm text-gray-600 dark:text-gray-400">{{ importMessage || 'Deploying solution...' }}</p>
        </div>
      </div>

      <!-- Done step -->
      <div v-else-if="step === 'done'" class="px-5 py-8">
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
            class="px-4 py-2 text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 rounded-lg transition-colors"
          >
            OK
          </button>
        </div>
      </div>

      <!-- Error step -->
      <div v-else-if="step === 'error'" class="px-5 py-6">
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
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            Back
          </button>
          <button
            @click="close"
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
