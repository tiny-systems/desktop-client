<script lang="ts">
import type { PropType } from 'vue'
import * as common from './common'
import Editor from './Editor.vue'

export default {
  components: {
    editor: Editor
  },
  emits: ['update-value'],
  props: {
    schema: {
      type: Object as PropType<common.Schema>,
      required: true,
    },
    initialValue: null,
    locale: Object as PropType<common.Locale>,
    theme: null,
    readonly: Boolean,
    allowLookup: Boolean,
    plainStruct: Boolean,
    noBorder: Boolean,
    disableCollapse: Boolean,
    allowEditSchema: Boolean,
    minItemCountIfNeedFilter: Number,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    },
  },
  computed: {
    themeObject(): common.Theme {
      return common.getTheme(this.theme)
    },
    localeObject(): common.Locale {
      return common.getLocale(this.locale)
    }
  },
  methods: {
    getReference (name: string)  {
      if (this.schema && this.schema.$defs) {
        const key = name.substring('#/$defs/'.length)
        const result = this.schema.$defs[key]
        if (result) {
          // Deep clone to prevent mutation of original $defs
          return JSON.parse(JSON.stringify(result))
        }
        return result
      }
      return undefined
    },
    updateValue(...args: any[]) {
      this.$emit('update-value', ...args)
    }
  },
}
</script>
<template>
  <editor :schema="schema"
    :initial-value="initialValue"
    :getReference="getReference"
    :theme="themeObject"
    :locale="localeObject"
    :readonly="readonly"
    :allow-edit-schema="allowEditSchema"
    :required="true"
    :allow-lookup="allowLookup"
    :no-border="noBorder"
    :plain-struct="plainStruct"
    :has-delete-button="hasDeleteButton"
    :disableCollapse="disableCollapse"
    :minItemCountIfNeedFilter="minItemCountIfNeedFilter"
    @update-value="updateValue">
  </editor>
</template>
