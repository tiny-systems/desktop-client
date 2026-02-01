<script setup>
/**
 * JSON Schema Editor wrapper component
 * Renders form controls based on JSON Schema
 */
import { computed } from 'vue'
import Editor from './Editor.vue'
import * as common from './common'

const props = defineProps({
  schema: {
    type: Object,
    default: () => ({})
  },
  initialValue: {
    type: [Object, String, Number, Boolean, Array],
    default: undefined
  },
  readonly: {
    type: Boolean,
    default: false
  },
  plainStruct: {
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
  disableCollapse: {
    type: Boolean,
    default: false
  },
  hideRootLookup: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update-value', 'lookup'])

// Get theme object
const themeObj = computed(() => common.getTheme(props.theme))

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
  emit('update-value', event)
}

// Handle lookup events
const handleLookup = (...args) => {
  emit('lookup', ...args)
}
</script>

<template>
  <div class="json-editor">
    <Editor
      v-if="schema && Object.keys(schema).length > 0"
      :schema="schema"
      :initial-value="initialValue"
      :theme="themeObj"
      :locale="locale"
      :getReference="getReference"
      :readonly="readonly"
      :required="true"
      :no-border="noBorder"
      :plain-struct="plainStruct"
      :disable-collapse="disableCollapse"
      :hide-root-lookup="hideRootLookup"
      @update-value="handleUpdateValue"
      @lookup="handleLookup"
    />
    <div v-else class="text-center py-2 text-gray-500 dark:text-gray-400 text-sm">
      No configurable properties
    </div>
  </div>
</template>
