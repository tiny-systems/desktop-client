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
  emits: ['update-value'],
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
      const schema = this.getRealSchemaRecursive(this.schema as Schema)
      // Infer type from schema structure if not explicitly set
      return this.inferSchemaType(schema)
    },
  },
  methods: {
    inferSchemaType(s: Schema): Schema {
      if (!s || s.type) {
        return s
      }
      // Infer type: "object" if schema has properties
      if (s.properties) {
        return { ...s, type: 'object' }
      }
      // Infer type: "array" if schema has items
      if (s.items) {
        return { ...s, type: 'array' }
      }
      return s
    },
    realSchemaFull(s: any): Schema {
      if (!isObject(s)) {
        return s
      }

      s = this.getRealSchemaRecursive(s)
      s = this.inferSchemaType(s)
      for (const property in s) {
        if (isObject((s as any)[property])) {
          (s as any)[property] = this.realSchemaFull((s as any)[property])
        }
      }
      return s
    },
    getRealSchemaRecursive(s: Schema): Schema {
      if ( s && s.$ref) {
        // ref overrides defs in some props
        const reference = this.getReference(s.$ref)
        if (reference) {
          // Copy override properties from the $ref usage to the resolved reference
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
    hasType(types: any, check: any) {
      let val = types

      if (isProxy(val)) {
        val = toRaw(val)
      }
      if (Array.isArray(val)) {
        return val.includes(check)
      }
      return val == check
    },
    updateValue(...args: any[]) {
      this.$emit('update-value', ...args)
    }
  }
}
</script>
