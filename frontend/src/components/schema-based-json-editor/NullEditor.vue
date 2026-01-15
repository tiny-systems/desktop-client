<template>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import * as common from './common'

export default {
  props: {
    schema: {
      type: Object as PropType<common.NullSchema>,
      required: true,
    },
    initialValue: null,
    title: [String, Number],
    theme: {
      type: Object as PropType<common.Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<common.Locale>,
      required: true,
    },
    readonly: Boolean,
    required: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    },
  },
  data: () => {
    return {
      value: null as null | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
}
</script>
