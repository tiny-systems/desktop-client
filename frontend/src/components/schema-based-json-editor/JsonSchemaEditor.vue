<template>
  <JsonSchemaEditorItem
    :value="value"
    :root="root"
    :lang="lang"
  />
</template>

<script lang="ts">
import JsonSchemaEditorItem from './json-schema-editor/Index.vue'

export default {
  name: 'JsonSchemaEditor',
  components: {
    JsonSchemaEditorItem
  },
  props: {
    value: {
      type: Object,
      default: () => ({})
    },
    root: {
      type: Boolean,
      default: false
    },
    lang: {
      type: String,
      default: 'en_US'
    }
  }
}
</script>
