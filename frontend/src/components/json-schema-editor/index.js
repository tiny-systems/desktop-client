import JsonSchemaEditor from './Main.vue'

JsonSchemaEditor.install = function (vue) {
  vue.component(JsonSchemaEditor.name, JsonSchemaEditor)
}

export default JsonSchemaEditor
