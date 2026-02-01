<script setup>
import { ref, computed, watch } from 'vue'
import Editor from '../schema-based-json-editor/Editor.vue'
import * as common from '../schema-based-json-editor/common'

const props = defineProps({
  schema: {
    type: Object,
    default: () => ({})
  },
  modelValue: {
    type: [Object, String, Number, Boolean, Array],
    default: () => ({})
  },
  readonly: {
    type: Boolean,
    default: false
  },
  theme: {
    type: String,
    default: 'small'
  },
  noBorder: {
    type: Boolean,
    default: false
  },
  allowLookup: {
    type: Boolean,
    default: false
  },
  allowEditSchema: {
    type: Boolean,
    default: false
  },
  hideRootLookup: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'action', 'lookup'])

// Get theme object
const themeObj = computed(() => common.getTheme(props.theme))

// Handle lookup event from editor
const handleLookup = (...args) => {
  emit('lookup', ...args)
}

// Get locale object
const locale = computed(() => common.defaultLocale)

// Get reference function for resolving $ref in schemas
const getReference = (ref) => {
  if (!props.schema?.$defs) return undefined
  const refName = ref.replace('#/$defs/', '')
  return props.schema.$defs[refName]
}

// Handle value updates from editor
const handleUpdateValue = (event) => {
  // If this is an action button click, emit the action event
  if (event?.isAction) {
    emit('action', event)
    return
  }
  emit('update:modelValue', event.value)
}
</script>

<template>
  <div class="schema-form">
    <Editor
      v-if="schema && Object.keys(schema).length > 0"
      :schema="schema"
      :initial-value="modelValue"
      :theme="themeObj"
      :locale="locale"
      :getReference="getReference"
      :readonly="readonly"
      :required="true"
      :no-border="noBorder"
      :plain-struct="true"
      :allow-lookup="allowLookup"
      :hide-root-lookup="hideRootLookup"
      :allow-edit-schema="allowEditSchema"
      @update-value="handleUpdateValue"
      @lookup="handleLookup"
    />
    <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
      No configurable properties
    </div>
  </div>
</template>
