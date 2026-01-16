<template>
  <any-editor v-if="hasType(realSchema.type, '') || hasType(realSchema.type, undefined) || realSchema.configure"
    :schema="realSchemaFull(realSchema)"
    :key="property"
    :initial-value="initialValue"
    :title="title"
    :property="property"
    :theme="theme"
    :locale="locale"
    :readonly="readonly"
    :allow-lookup="allowLookup"
    :plain-struct="plainStruct"
    :allow-edit-schema="allowEditSchema"
    :required="required"
    :table-mode="tableMode"
    :has-delete-button="hasDeleteButton"
    @lookup="lookup"
    @update-value="updateValue">
  </any-editor>
  <object-editor v-else-if="hasType(realSchema.type, 'object')"
     :schema="realSchema"
     :initial-value="initialValue"
     :key="property"
     :title="title"
     :allow-lookup="allowLookup"
     :plain-struct="plainStruct"
     :allow-edit-schema="allowEditSchema"
     :getReference="getReference"
     :theme="theme"
     :no-border="noBorder"
     :locale="locale"
     :readonly="readonly"
     :required="required"
     :has-delete-button="hasDeleteButton"
     :disableCollapse="disableCollapse"
     :table-mode="tableMode"
     :minItemCountIfNeedFilter="minItemCountIfNeedFilter"
     @lookup="lookup"
     @update-value="updateValue">
  </object-editor>
  <array-editor v-else-if="hasType(realSchema.type, 'array')"
    :schema="realSchema"
    :initial-value="initialValue"
    :title="title"
    :getReference="getReference"
    :theme="theme"
    :allow-edit-schema="allowEditSchema"
    :allow-lookup="allowLookup"
    :plain-struct="plainStruct"
    :locale="locale"
    :readonly="readonly"
    :required="required"
    :has-delete-button="hasDeleteButton"
    :disableCollapse="disableCollapse"
    :table-mode="tableMode"
    :minItemCountIfNeedFilter="minItemCountIfNeedFilter"
    @lookup="lookup"
    @update-value="updateValue">
  </array-editor>
  <number-editor v-else-if="hasType(realSchema.type,'number') || hasType(realSchema.type, 'integer')"
     :schema="realSchema"
     :initial-value="initialValue"
     :title="title"
     :theme="theme"
     :allow-lookup="allowLookup"
     :plain-struct="plainStruct"
     :allow-edit-schema="allowEditSchema"
     :locale="locale"
     :readonly="readonly"
     :required="required"
     :table-mode="tableMode"
     :has-delete-button="hasDeleteButton"
     @lookup="lookup"
     @update-value="updateValue">
  </number-editor>
  <boolean-editor v-else-if="hasType(realSchema.type, 'boolean')"
    :schema="realSchema"
    :initial-value="initialValue"
    :title="title"
    :allow-lookup="allowLookup"
    :plain-struct="plainStruct"
    :allow-edit-schema="allowEditSchema"
    :theme="theme"
    :locale="locale"
    :readonly="readonly"
    :required="required"
    :table-mode="tableMode"
    :has-delete-button="hasDeleteButton"
    @lookup="lookup"
    @update-value="updateValue">
  </boolean-editor>
  <string-editor v-else-if="hasType(realSchema.type, 'string')"
     :schema="realSchema"
     :initial-value="initialValue"
     :title="title"
     :theme="theme"
     :allow-lookup="allowLookup"
     :plain-struct="plainStruct"
     :allow-edit-schema="allowEditSchema"
     :locale="locale"
     :readonly="readonly"
     :required="required"
     :table-mode="tableMode"
     :has-delete-button="hasDeleteButton"
     @lookup="lookup"
     @update-value="updateValue">
  </string-editor>
  <null-editor v-else-if="hasType(realSchema.type,  'null') || hasType(realSchema.type,  null)"
     :schema="realSchema"
     :initial-value="initialValue"
     :title="title"
     :theme="theme"
     :locale="locale"
     :readonly="readonly"
     :required="required"
     :table-mode="tableMode"
     :has-delete-button="hasDeleteButton">
  </null-editor>
</template>
<script lang="ts">
import {isProxy, toRaw} from 'vue'
import type {PropType} from 'vue'
import * as common from './common'
import type {Schema} from './common'
import { defineAsyncComponent } from 'vue'
import BooleanEditor from './BooleanEditor.vue'
import NullEditor from './NullEditor.vue'
import NumberEditor from './NumberEditor.vue'
import StringEditor from './StringEditor.vue'
import {isObject} from 'lodash';

// Use defineAsyncComponent to break circular dependencies
const AnyEditor = defineAsyncComponent(() => import('./AnyEditor.vue'))
const ArrayEditor = defineAsyncComponent(() => import('./ArrayEditor.vue'))
const ObjectEditor = defineAsyncComponent(() => import('./ObjectEditor.vue'))

export default {
  components: {
    'any-editor': AnyEditor,
    'array-editor': ArrayEditor,
    'boolean-editor': BooleanEditor,
    'null-editor': NullEditor,
    'number-editor': NumberEditor,
    'object-editor': ObjectEditor,
    'string-editor': StringEditor,
  },
  emits: ['lookup', 'update-value'],
  props: {
    schema: {
      required: true,
    },
    allowLookup: Boolean,
    plainStruct: Boolean,
    noBorder: Boolean,
    initialValue: null,
    title: [String, Number],
    getReference: {
      type: Function as unknown as PropType<(ref: string) => Schema | undefined>,
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
    readonly: Boolean,
    required: Boolean,
    allowEditSchema: Boolean,
    tableMode: Boolean,
    hasDeleteButton: {
      type: Boolean,
      required: false,
    },
    property: String,
    disableCollapse: Boolean,
    minItemCountIfNeedFilter: Number
  },
  computed: {
    realSchema(): Schema {
      //@ts-ignore
      return this.getRealSchemaRecursive(this.schema)
    },

  },
  methods: {
    realSchemaFull(s: any): Schema {
      if (!isObject(s)) {
        return s
      }

      s = this.getRealSchemaRecursive(s)
      //@ts-ignore
      for (const property in s) {
        //@ts-ignore
        if (isObject(s[property])) {
          //@ts-ignore
          s[property] = this.realSchemaFull(s[property])
        }
      }
      return s
    },
    getRealSchemaRecursive(s: Schema): Schema {
      if ( s && s.$ref) {
        // ref overrides defs in some props
        const reference = this.getReference(s.$ref)
        if (reference) {
          if (s.title !== undefined) {
            reference.title = s.title
          }
          if (s.description !== undefined) {
            reference.description = s.description
          }
          if (s.readonly) {
            reference.readonly = s.readonly
          }
          if (s.configurable) {
            reference.configurable = s.configurable
          }
          return this.getRealSchemaRecursive(reference)
        }
      }
      return s
    },
    //@ts-ignore
    hasType(types, check) {

      let val = types

      if (isProxy(val)) {
        val = toRaw(val)
      }
      if (Array.isArray(val)) {
        return val.includes(check)
      }
      return val == check
    },
    lookup(...args: any[]) {
      this.$emit('lookup', ...args)
    },
    updateValue(...args: any[]) {
      this.$emit('update-value', ...args)
    }
  }
}
</script>
