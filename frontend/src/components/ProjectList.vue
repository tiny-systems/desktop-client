<template>
  <div class="h-full flex flex-col bg-white dark:bg-gray-900">
    <!-- Header -->
    <header class="sticky top-0 bg-white dark:bg-gray-900 z-50 border-b border-gray-200 dark:border-gray-700">
      <div class="px-4 py-3 flex items-center justify-between">
        <ContextSelector @select="onSelect" @contexts-loaded="onContextsLoaded" :ctx="ctx"/>
        <div v-if="ctx && statusClass !== 'error' && activeTab === 'projects'" class="flex items-center space-x-3">
          <button
            @click="openSolutionsDirectory"
            class="flex items-center space-x-2 px-4 py-2 text-sky-600 hover:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300 text-sm font-medium transition-colors"
          >
            <span>Solutions</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
            </svg>
          </button>
          <button
            @click="showCreateDialog = true"
            class="flex items-center space-x-2 px-4 py-2 bg-sky-600 hover:bg-sky-700 text-white rounded-lg text-sm font-medium transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
            </svg>
            <span>New Project</span>
          </button>
        </div>
        <button
          v-if="ctx && activeTab === 'modules'"
          @click="openModulesDirectory"
          class="flex items-center space-x-2 px-4 py-2 text-sky-600 hover:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300 text-sm font-medium transition-colors"
        >
          <span>Modules Directory</span>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
          </svg>
        </button>
      </div>
      <!-- Tabs -->
      <div v-if="ctx" class="px-4 border-t border-gray-100 dark:border-gray-800">
        <nav class="flex space-x-6" aria-label="Tabs">
          <button
            @click="activeTab = 'projects'"
            :class="[
              'py-3 px-1 border-b-2 font-medium text-sm transition-colors',
              activeTab === 'projects'
                ? 'border-sky-500 text-sky-600 dark:text-sky-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
            ]"
          >
            <div class="flex items-center space-x-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
              </svg>
              <span>Projects</span>
            </div>
          </button>
          <button
            @click="activeTab = 'modules'"
            :class="[
              'py-3 px-1 border-b-2 font-medium text-sm transition-colors',
              activeTab === 'modules'
                ? 'border-sky-500 text-sky-600 dark:text-sky-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
            ]"
          >
            <div class="flex items-center space-x-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
              </svg>
              <span>Modules</span>
            </div>
          </button>
        </nav>
      </div>
    </header>

    <!-- Main content -->
    <main v-if="ctx" class="flex-1 overflow-auto">
      <!-- Projects Tab -->
      <div v-show="activeTab === 'projects'">
      <div class="p-4">
        <!-- Loading state -->
        <div v-if="isLoading" class="flex items-center justify-center h-64">
          <div class="text-center">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-500 mx-auto mb-4"></div>
            <span class="text-gray-500 dark:text-gray-400">Loading projects...</span>
          </div>
        </div>

        <!-- Projects grid -->
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" v-else-if="projects.length > 0">
          <div
            v-for="p in projects"
            :key="p.name"
            class="group bg-white dark:bg-gray-800 p-5 rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer hover:border-sky-300 dark:hover:border-sky-600 hover:shadow-md transition-all duration-200"
            tabindex="0"
            @click="handleProjectSelection(p)"
            @keydown.enter="handleProjectSelection(p)"
          >
            <div class="flex items-center space-x-3">
              <div class="p-2.5 rounded-lg bg-sky-100 dark:bg-sky-900/30 text-sky-600 dark:text-sky-400 group-hover:bg-sky-200 dark:group-hover:bg-sky-900/50 transition-colors">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                </svg>
              </div>
              <h3 class="text-lg font-medium text-gray-900 dark:text-white truncate">
                {{ p.title }}
              </h3>
            </div>
            <p class="mt-3 text-gray-500 dark:text-gray-400 text-sm line-clamp-3" v-if="p.description">
              {{ p.description }}
            </p>
          </div>
        </div>

        <!-- Empty state -->
        <div v-else class="flex items-center justify-center" :class="isCrdError ? 'min-h-64 py-8' : 'h-64'">
          <div class="text-center">
            <!-- CRD not installed — friendly setup guide -->
            <div v-if="statusClass === 'error' && isCrdError" class="max-w-lg mx-auto text-left">
              <div class="flex items-center gap-3 mb-4">
                <div class="p-2 rounded-full bg-amber-100 dark:bg-amber-900/30">
                  <svg class="w-6 h-6 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                  </svg>
                </div>
                <div>
                  <h3 class="text-base font-semibold text-gray-900 dark:text-white">TinySystems CRDs not installed</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">Run these commands in your terminal to get started:</p>
                </div>
              </div>
              <div class="bg-gray-900 dark:bg-gray-950 rounded-lg p-4 text-sm font-mono text-gray-300 overflow-x-auto relative group">
                <button
                  @click="copyCrdCommands"
                  class="absolute top-2 right-2 p-1.5 rounded text-gray-500 hover:text-white hover:bg-gray-700 opacity-0 group-hover:opacity-100 transition-opacity"
                  title="Copy to clipboard"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
                <pre class="whitespace-pre leading-relaxed"><span class="text-gray-500"># Add Tiny Systems Helm repository</span>
helm repo add tinysystems https://tiny-systems.github.io/module/
helm repo update

<span class="text-gray-500"># Install CRDs (required once per cluster)</span>
helm upgrade --install tinysystems-crd tinysystems/tinysystems-crd \
  --namespace {{ ctx?.ns || 'default' }} \
  --create-namespace</pre>
              </div>
              <p class="text-xs text-gray-400 dark:text-gray-500 mt-3">After installing, click the refresh button above to reload.</p>
            </div>
            <!-- Generic error -->
            <div v-else-if="statusClass === 'error'" class="text-red-500 dark:text-red-400 mb-2">
              <svg class="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
              <p class="text-red-600 dark:text-red-400">{{ statusMessage }}</p>
            </div>
            <div v-else>
              <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
              <p class="text-gray-500 dark:text-gray-400 mb-4">No projects found</p>
              <button
                @click="showCreateDialog = true"
                class="px-4 py-2 bg-sky-600 hover:bg-sky-700 text-white rounded-lg text-sm font-medium transition-colors"
              >
                Create your first project
              </button>
            </div>
          </div>
        </div>
      </div>
      </div>

      <!-- Modules Tab -->
      <ModuleList v-show="activeTab === 'modules'" :ctx="ctx" ref="moduleListRef" />
    </main>

    <!-- No context selected -->
    <main v-else-if="hasContexts" class="flex-1 flex items-center justify-center">
      <div class="text-center text-gray-500 dark:text-gray-400">
        <svg class="w-16 h-16 mx-auto mb-4 text-gray-300 dark:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
        </svg>
        <p class="text-lg">Select a context and namespace to get started</p>
      </div>
    </main>

    <!-- No contexts available -->
    <main v-else class="flex-1"></main>

    <!-- Create Project Dialog -->
    <div v-if="showCreateDialog" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="fixed inset-0 bg-black/40 backdrop-blur-md" @click="closeCreateDialog"></div>
      <div class="flex min-h-full items-center justify-center p-4">
        <div class="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Create New Project</h3>
          <form @submit.prevent="createProject">
            <div class="mb-4">
              <label for="project-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Project Name
              </label>
              <input
                ref="projectNameInput"
                id="project-name"
                v-model="newProjectName"
                type="text"
                placeholder="My Awesome Project"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent"
                :disabled="isCreating"
              />
            </div>
            <div v-if="createError" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg">
              <p class="text-sm text-red-600 dark:text-red-400">{{ createError }}</p>
            </div>
            <div class="flex justify-end space-x-3">
              <button
                type="button"
                @click="closeCreateDialog"
                :disabled="isCreating"
                class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                :disabled="isCreating || !newProjectName.trim()"
                class="px-4 py-2 text-sm font-medium text-white bg-sky-600 rounded-lg hover:bg-sky-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center space-x-2"
              >
                <div v-if="isCreating" class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                <span>{{ isCreating ? 'Creating...' : 'Create Project' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import {computed, onMounted, ref, watch, nextTick} from 'vue';
import {BrowserOpenURL} from '../../wailsjs/runtime/runtime.js';
import ContextSelector from "./ContextSelector.vue";
import ModuleList from "./ModuleList.vue";

const props = defineProps({
  ctx: Object
})

const activeTab = ref('projects')
const moduleListRef = ref(null)

const openModulesDirectory = () => {
  BrowserOpenURL('https://tinysystems.io/modules')
}

const openSolutionsDirectory = () => {
  BrowserOpenURL('https://tinysystems.io/solutions')
}

const ctx = ref(props.ctx)
const hasContexts = ref(false)

const onSelect = function (c) {
  ctx.value = c
  emit('select-context', c)
  if (!c) {
    return
  }
  loadProjects(c.name, c.ns)
}

const onContextsLoaded = function (hasAny) {
  hasContexts.value = hasAny
}
const emit = defineEmits(['select-project', 'select-context'])

// Define the path to your Go backend functions.
const GoApp = window.go.main.App;

const statusMessage = ref('Initializing...');
const statusClass = ref('');
const isLoading = ref(false)
const projects = ref([]);

const isCrdError = computed(() => {
  const msg = statusMessage.value || ''
  return msg.includes('operator.tinysystems.io') || msg.includes('no matches for')
})

const copyCrdCommands = async () => {
  const ns = ctx.value?.ns || 'default'
  const commands = `# Add Tiny Systems Helm repository
helm repo add tinysystems https://tiny-systems.github.io/module/
helm repo update

# Install CRDs (required once per cluster)
helm upgrade --install tinysystems-crd tinysystems/tinysystems-crd \\
  --namespace ${ns} \\
  --create-namespace`
  try {
    await navigator.clipboard.writeText(commands)
  } catch (e) {
    // fallback — ignore
  }
}

// Create project dialog state
const showCreateDialog = ref(false)
const newProjectName = ref('')
const isCreating = ref(false)
const createError = ref('')
const projectNameInput = ref(null)

// Focus input when dialog opens
watch(showCreateDialog, (newVal) => {
  if (newVal) {
    nextTick(() => {
      projectNameInput.value?.focus()
    })
  }
})

const handleProjectSelection = (prj) => {
  statusMessage.value = ``;
  emit('select-project', prj)
};

const loadProjects = async (name, ns) => {
  statusMessage.value = 'Attempting to read TinyProjects...';
  isLoading.value = true
  projects.value = []
  try {
    const fetchedProjects = await GoApp.GetProjects(name, ns);
    if (!fetchedProjects || fetchedProjects.length === 0) {
      statusMessage.value = 'No projects found in ' + ns + ' namespace';
      statusClass.value = ''
      return
    }

    projects.value = fetchedProjects;
    statusClass.value = 'success'
    statusMessage.value = 'found ' + fetchedProjects.length + ' projects'

  } catch (error) {
    statusMessage.value = `Error loading projects: ${error}`;
    statusClass.value = 'error';
    console.error('Error in getProjects:', error);
  } finally {
    isLoading.value = false;
  }
}

const closeCreateDialog = () => {
  if (isCreating.value) return
  showCreateDialog.value = false
  newProjectName.value = ''
  createError.value = ''
}

const createProject = async () => {
  if (!newProjectName.value.trim() || !ctx.value) return

  isCreating.value = true
  createError.value = ''

  try {
    const newProject = await GoApp.CreateProject(ctx.value.name, ctx.value.ns, newProjectName.value.trim())
    // Add the new project to the list
    projects.value.push(newProject)
    isCreating.value = false
    closeCreateDialog()
  } catch (error) {
    createError.value = `Failed to create project: ${error}`
    console.error('Error creating project:', error)
    isCreating.value = false
  }
}

onMounted(() => {
  if (GoApp) {
    return
  }
  statusMessage.value = "Wails Go runtime not ready.";
  statusClass.value = 'error';
});
</script>
