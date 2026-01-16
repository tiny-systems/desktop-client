<template>
  <label :class="theme.label" v-if="hasOptionalCheckbox && !isReadOnly">
    <input type="checkbox" :class="theme.checkboxInput" @change="$emit('toggleOptional')" :checked="value === undefined" :disabled="isReadOnly" />
    {{locale.info.notExists}}
  </label>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import * as common from './common'

export default {
  emits: ['toggleOptional'],
  props: {
    required: Boolean,
    value: null,
    isReadOnly: Boolean,
    theme: {
      type: Object as PropType<common.Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<common.Locale>,
      required: true,
    },
  },
  computed: {
    hasOptionalCheckbox(): boolean {
      return !this.required && (this.value === undefined || !this.isReadOnly)
    },
  }
}
</script>
