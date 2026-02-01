<template>
  <div class="p-4">
    <!-- Loading state -->
    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-500 mx-auto mb-4"></div>
        <span class="text-gray-500 dark:text-gray-400">Loading modules...</span>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="flex items-center justify-center h-64">
      <div class="text-center">
        <svg class="w-12 h-12 mx-auto text-red-500 dark:text-red-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
        <p class="text-red-600 dark:text-red-400">{{ error }}</p>
      </div>
    </div>

    <!-- Modules list -->
    <div v-else-if="modules.length > 0" class="space-y-4">
      <div
        v-for="mod in modules"
        :key="mod.name"
        class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
      >
        <!-- Module header -->
        <div
          class="px-4 py-3 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700/50 transition-colors"
          @click="toggleModule(mod.name)"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <!-- Expand/collapse icon -->
              <svg
                class="w-4 h-4 text-gray-500 dark:text-gray-400 transition-transform"
                :class="{ 'rotate-90': expandedModules[mod.name] }"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
              </svg>
              <!-- Module icon -->
              <div class="p-2 rounded-lg bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
                </svg>
              </div>
              <div>
                <h3 class="text-base font-medium text-gray-900 dark:text-white">
                  {{ mod.name }}
                </h3>
                <div class="flex items-center space-x-2 mt-0.5">
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    v{{ mod.version }}
                  </span>
                  <span v-if="mod.sdkVersion" class="text-xs text-gray-400 dark:text-gray-500">
                    SDK {{ mod.sdkVersion }}
                  </span>
                </div>
              </div>
            </div>
            <span class="text-sm text-gray-500 dark:text-gray-400">
              {{ mod.components?.length || 0 }} component{{ (mod.components?.length || 0) !== 1 ? 's' : '' }}
            </span>
          </div>
        </div>

        <!-- Components list (collapsible) -->
        <div v-show="expandedModules[mod.name]" class="divide-y divide-gray-100 dark:divide-gray-700/50">
          <div
            v-for="comp in mod.components"
            :key="comp.name"
            class="px-4 py-3 pl-14 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center space-x-2">
                  <span class="font-medium text-gray-900 dark:text-white">{{ comp.name }}</span>
                  <div v-if="comp.tags && comp.tags.length > 0" class="flex flex-wrap gap-1">
                    <span
                      v-for="tag in comp.tags"
                      :key="tag"
                      class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300"
                    >
                      {{ tag }}
                    </span>
                  </div>
                </div>
                <p v-if="comp.description" class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ comp.description }}
                </p>
                <p v-if="comp.info" class="mt-1 text-xs text-gray-400 dark:text-gray-500">
                  {{ comp.info }}
                </p>
              </div>
            </div>
          </div>
          <!-- Empty components state -->
          <div v-if="!mod.components || mod.components.length === 0" class="px-4 py-6 pl-14 text-center">
            <p class="text-sm text-gray-400 dark:text-gray-500">No components available</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="flex items-center justify-center h-64">
      <div class="text-center">
        <svg class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/>
        </svg>
        <p class="text-gray-500 dark:text-gray-400">No modules installed in this namespace</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, reactive } from 'vue'

const props = defineProps({
  ctx: Object
})

const GoApp = window.go.main.App

const modules = ref([])
const isLoading = ref(false)
const error = ref('')
const expandedModules = reactive({})

const toggleModule = (moduleName) => {
  expandedModules[moduleName] = !expandedModules[moduleName]
}

const loadModules = async () => {
  if (!props.ctx) return

  isLoading.value = true
  error.value = ''
  modules.value = []

  try {
    const fetchedModules = await GoApp.GetModules(props.ctx.name, props.ctx.ns)
    modules.value = fetchedModules || []
    // Modules start collapsed by default
  } catch (e) {
    error.value = `Failed to load modules: ${e}`
    console.error('Error loading modules:', e)
  } finally {
    isLoading.value = false
  }
}

// Watch for context changes
watch(() => props.ctx, (newCtx) => {
  if (newCtx) {
    loadModules()
  }
}, { immediate: true })

// Expose reload method for parent
defineExpose({ loadModules })
</script>
