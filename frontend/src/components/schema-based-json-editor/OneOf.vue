<template>
    <div v-for="(p, i) in schema.oneOf" >
      <editor :schema="p"
        :title="title"
        :getReference="getReference"
        :initial-value="initialValue"
        :theme="theme"
        :property="property"
        :locale="locale"
        :allow-lookup="allowLookup"
        :plain-struct="plainStruct"
        :allow-edit-schema="allowEditSchema"
        :required="required"
        :has-delete-button="hasDeleteButton"
        :disable-collapse="disableCollapse"
        :minItemCountIfNeedFilter="minItemCountIfNeedFilter"
        @update-value="onChange"
        @lookup="lookup">
      </editor>
    </div>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import * as common from './common'
import Editor from './Editor.vue'

export default {
  emits: ['update-value', 'delete', 'lookup'],
  components: {
    editor: Editor
  },
  props: {
    schema: {
      type: Object as PropType<common.ObjectSchema>,
      required: true,
    },
    property: String,
    allowLookup: Boolean,
    plainStruct: Boolean,
    initialValue: Object as PropType<{ [name: string]: common.ValueType }>,
    title: [String, Number],
    getReference: {
      type: Function as unknown as PropType<(ref: string) => common.Schema | undefined>,
      required: true,
    },
    theme: {
      type: Object as PropType<common.Theme>,
      required: true,
    },
    locale: {
      type: Object as PropType<common.Locale>,
      required: true,
    },
    allowEditSchema: Boolean,
    readonly: Boolean,
    required: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    },
    disableCollapse: Boolean,
    minItemCountIfNeedFilter: Number
  },
  data: () => {
    return {
    }
  },
  beforeMount() {
  },
  computed: {
  },
  methods: {
    lookup(data, schema, cb) {
      this.$emit('lookup', data, schema, cb)
    },
    onChange({ value, isValid }: common.ValidityValue<common.ValueType>) {
      this.$emit('update-value', { value, isValid})
    }
  },
  watch: {
  },
}
</script>
