<script setup>
import { CheckIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Confirmation.'
  },
  message: {
    type: String,
    default: 'Are you sure?'
  },
  detail: {
    type: String,
    default: ''
  },
  confirmText: {
    type: String,
    default: "Yes, I'm sure"
  },
  cancelText: {
    type: String,
    default: 'No, cancel'
  }
})

const emit = defineEmits(['confirm', 'cancel'])

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
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
      <div class="relative rounded-lg bg-white dark:bg-gray-900 dark:border dark:border-gray-700 shadow-xl p-4 max-w-sm w-full text-center">
        <!-- Checkmark icon -->
        <div class="mx-auto w-12 h-12 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mb-3">
          <CheckIcon class="w-6 h-6 text-green-500" />
        </div>

        <!-- Title -->
        <h3 class="text-base font-semibold text-gray-900 dark:text-white mb-1">
          {{ title }}
        </h3>

        <!-- Message -->
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">
          {{ message }}
        </p>

        <!-- Detail -->
        <p v-if="detail" class="text-sm text-gray-900 dark:text-white font-medium mb-4">
          {{ detail }}
        </p>

        <!-- Buttons -->
        <div class="flex gap-2">
          <button
            type="button"
            @click="handleCancel"
            class="flex-1 px-3 py-2 text-xs font-medium rounded-md border border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
          >
            {{ cancelText }}
          </button>
          <button
            type="button"
            @click="handleConfirm"
            class="flex-1 px-3 py-2 text-xs font-medium rounded-md bg-red-500 text-white hover:bg-red-600"
          >
            {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
