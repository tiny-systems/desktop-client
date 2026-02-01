<script setup>
import { ExclamationTriangleIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['cancel', 'discard', 'save'])

const handleCancel = () => {
  emit('cancel')
}

const handleDiscard = () => {
  emit('discard')
}

const handleSave = () => {
  emit('save')
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
    >
      <!-- Backdrop -->
      <div
        class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm"
        @click="handleCancel"
      />

      <!-- Dialog -->
      <div class="relative rounded-lg bg-white dark:bg-gray-900 dark:border dark:border-gray-700 shadow-xl p-6 max-w-md w-full">
        <div class="flex items-start gap-4">
          <!-- Warning icon -->
          <div class="flex-shrink-0 w-12 h-12 rounded-full bg-yellow-100 dark:bg-yellow-900/30 flex items-center justify-center">
            <ExclamationTriangleIcon class="w-7 h-7 text-yellow-500" />
          </div>

          <div class="flex-1">
            <!-- Title -->
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Unsaved Changes
            </h3>

            <!-- Message -->
            <p class="text-sm text-gray-500 dark:text-gray-400">
              You have unsaved changes in your configuration. Would you like to save them before continuing?
            </p>
          </div>
        </div>

        <!-- Buttons -->
        <div class="flex justify-end gap-3 mt-6">
          <button
            type="button"
            @click="handleCancel"
            class="px-4 py-2 text-sm font-medium rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
          <button
            type="button"
            @click="handleDiscard"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-red-500 text-white hover:bg-red-600"
          >
            Discard
          </button>
          <button
            type="button"
            @click="handleSave"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-sky-500 text-white hover:bg-sky-600"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
