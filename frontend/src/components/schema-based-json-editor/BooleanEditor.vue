<template>
  <div :class="className">
    <div :class="theme.title">
      <div class="flex space-x-1 w-full">
        <div class="pl-1" v-if="titleToShow != '' && schema.format != 'button'">{{titleToShow}}</div>
      </div>
      <div :class="theme.buttonGroup" :style="buttonGroupStyle">
        <optional :required="required"
                  :value="value"
                  :isReadOnly="isReadOnly"
                  :theme="theme"
                  :locale="locale"
                  @toggleOptional="toggleOptional()">
        </optional>
        <button v-if="hasDeleteButtonFunction" type="button" class="w-4 text-sky-500 cursor-pointer ml-1" @click="$emit('delete')" title="Delete">
          <XCircleIcon></XCircleIcon>
        </button>
      </div>
    </div>
    <div v-if="value !== undefined" class="flex justify-between mx-1">
      <div class="w-full">
        <label :class="theme.label" v-if="schema.format === 'checkbox'">
          <input type="checkbox"
                 :class="theme.checkboxInput"
                 @change="onChange()"
                 :checked="value"
                 :disabled="isReadOnly" />
          {{locale.info.true}}
        </label>
        <select v-else-if="schema.format === 'select'" :class="theme.select"
                :value="value"
                :disabled="isReadOnly"
                @change="onChange()">
          <option :value="true">{{locale.info.true}}</option>
          <option :value="false">{{locale.info.false}}</option>
        </select>
        <template v-else-if="schema.format === 'button'">
          <button type="button" :class="[theme.button]" :value="value.toString()" @click="emitAction">{{titleToShow}}</button>
        </template>
        <template v-else>
          <span :class="theme.radiobox">
            <label :class="theme.radioboxLabel">
              <input type="radio"
                     @change="onChange()"
                     :checked="value"
                     :value="true"
                     :class="theme.radioboxInput"
                     :disabled="isReadOnly" />
              {{locale.info.true}}
            </label>
          </span>
          <span :class="theme.radiobox">
            <label :class="theme.radioboxLabel">
              <input type="radio"
                     @change="onChange()"
                     :checked="!value"
                     :value="false"
                     :disabled="isReadOnly" />
                {{locale.info.false}}
            </label>
          </span>
        </template>
      </div>
    </div>
    <description :theme="theme" :message="schema.description"></description>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue'
import Optional from './Optional.vue'
import Description  from './Description.vue'
import * as common from './common'
import { XCircleIcon } from '@heroicons/vue/24/solid'

export default {
  emits: ['update-value', 'delete'],
  components: {
    XCircleIcon,
    optional: Optional,
    description: Description
  },
  props: {
    schema: {
      type: Object as PropType<common.BooleanSchema>,
      required: true,
    },
    allowEditSchema: Boolean,
    allowLookup: Boolean,
    plainStruct: Boolean,
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
      value: false as boolean | undefined,
      buttonGroupStyle: common.buttonGroupStyleString,
    }
  },
  beforeMount() {
    this.value = this.getValue()
    if (this.value !== undefined) {
      this.emitValue()
    }
  },
  computed: {
    isReadOnly(): boolean | undefined {
      return this.readonly || this.schema.readonly
    },
    hasDeleteButtonFunction(): boolean {
      return this.hasDeleteButton && !this.isReadOnly
    },
    titleToShow(): string {
      return common.getTitle(this.schema.title, this.title)
    },
    className(): string {
      const rowClass = this.theme.row
      return this.schema.className ? rowClass + ' ' + this.schema.className : rowClass
    },
  },
  methods: {
    onChange() {
      this.value = !this.value
      this.emitValue()
    },
    toggleOptional() {
      this.value = common.toggleOptional(this.value, this.schema, this.initialValue) as boolean | undefined
      this.emitValue()
    },
    emitValue() {
      this.$emit('update-value', { value: this.getAllValue(), isValid: true })
    },
    emitAction() {
      this.value = true
      this.$emit('update-value', { value: this.getAllValue(), isValid: true, isAction: true })
    },
    getAllValue() {
      if(this.value === undefined || this.plainStruct) {
        return this.value
      }
      return !!this.value
    },
    getValue(): boolean {
      return common.getDefaultValue(this.required, this.schema, this.initialValue) as boolean
    },
  }
}
</script>
