<template>
  <div v-if="message" :class="error ? theme.errorDescription : theme.description">{{ message }}</div>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import type { Theme } from './common'

export default {
  props: {
    theme: {
      type: Object as PropType<Theme>,
      required: true,
    },
    message: String,
    error: Boolean,
  }
}
</script>
