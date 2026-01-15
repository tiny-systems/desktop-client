<template>
  <span v-if="!required && !isReadOnly">
    <button :class="theme.button" type="button" @click="$emit('toggleOptional')">
      {{ value === undefined ? 'Define' : 'Clear' }}
    </button>
  </span>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import type { Theme, Locale } from './common'

export default {
  emits: ['toggleOptional'],
  props: {
    required: Boolean,
    value: null,
    isReadOnly: Boolean,
    theme: {
      type: Object as PropType<Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<Locale>,
      required: true,
    },
  }
}
</script>
