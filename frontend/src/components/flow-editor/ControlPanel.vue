<script setup>
import { ArrowDownTrayIcon, ArrowUpTrayIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  flowName: String,
  flowResourceName: String,
  projectName: String,
  loading: Boolean,
})

const emit = defineEmits(['close', 'error', 'import', 'export'])
</script>

<template>
  <div class="lg:items-center lg:justify-center p-1 lg:p-3 border-b border-gray-200 dark:border-gray-700">
    <div class="flex justify-between align-middle">
      <div class="flex justify-left align-middle">
        <!-- Back button -->
        <button
          type="button"
          @click="emit('close')"
          title="Back to project"
          class="text-sky-600 border border-sky-600 hover:bg-sky-600 hover:text-white focus:ring-4 focus:outline-none focus:ring-sky-300 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center mr-2 dark:border-sky-500 dark:text-sky-500 dark:hover:text-white dark:focus:ring-sky-800 dark:hover:bg-sky-600"
        >
          <svg
            aria-hidden="true"
            class="w-5 h-5 rotate-180"
            fill="currentColor"
            viewBox="0 0 20 20"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              fill-rule="evenodd"
              d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z"
              clip-rule="evenodd"
            />
          </svg>
          <span class="sr-only">Back to project</span>
        </button>

        <!-- Flow name -->
        <div class="inline-flex items-center max-w-96 truncate">
          <h3 class="text-lg inline-flex leading-4 font-medium text-gray-900 dark:text-gray-300 px-2">
            {{ flowName || flowResourceName }}
            <template v-if="projectName"> - {{ projectName }}</template>
          </h3>
        </div>

        <!-- Saving indicator -->
        <div v-if="loading" class="inline-flex items-center">
          <span class="text-sky-600 ml-3 text-sm">saving...</span>
        </div>
      </div>

      <!-- Import/Export buttons -->
      <div class="flex items-center gap-2">
        <!-- Import button -->
        <button
          type="button"
          @click="emit('import')"
          title="Import flow JSON"
          class="text-sky-600 border border-sky-600 hover:bg-sky-600 hover:text-white focus:ring-4 focus:outline-none focus:ring-sky-300 font-medium rounded-lg text-sm p-2 text-center inline-flex items-center dark:border-sky-500 dark:text-sky-500 dark:hover:text-white dark:focus:ring-sky-800 dark:hover:bg-sky-600"
        >
          <ArrowUpTrayIcon class="w-4 h-4" />
          <span class="sr-only">Import</span>
        </button>

        <!-- Export button -->
        <button
          type="button"
          @click="emit('export')"
          title="Export flow JSON"
          class="text-sky-600 border border-sky-600 hover:bg-sky-600 hover:text-white focus:ring-4 focus:outline-none focus:ring-sky-300 font-medium rounded-lg text-sm p-2 text-center inline-flex items-center dark:border-sky-500 dark:text-sky-500 dark:hover:text-white dark:focus:ring-sky-800 dark:hover:bg-sky-600"
        >
          <ArrowDownTrayIcon class="w-4 h-4" />
          <span class="sr-only">Export</span>
        </button>
      </div>
    </div>
  </div>
</template>
