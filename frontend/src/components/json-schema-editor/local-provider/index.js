const langs = {
  en_US: {
    'title': 'Title',
    'import_json': 'Import JSON',
    'base_setting': 'Base Setting',
    'all_setting': 'Source Code',
    'default': 'Default',
    'description': 'Description',
    'adv_setting': 'Advanced Settings',
    'add_child_node': 'Add child node',
    'add_sibling_node': 'Add sibling nodes',
    'add_node': 'Add sibling/child nodes',
    'remove_node': 'Remove node',
    'child_node': 'Child node',
    'sibling_node': 'Sibling node',
    'ok': 'OK',
    'cancel': 'Cancel',
    'minLength': 'Min length',
    'maxLength': 'Max length',
    'pattern': 'MUST be a valid regular expression.',
    'exclusiveMinimum': 'Value strictly less than',
    'exclusiveMaximum': 'Value strictly more than',
    'minimum': 'Min',
    'maximum': 'Max',
    'uniqueItems': 'Unique Items',
    'minItems': 'MinItems',
    'maxItems': 'MaxItems',
    'minProperties': 'MinProperties',
    'maxProperties': 'MaxProperties',
    'additionalProperties': 'Allow additional properties',
    'valid_json': 'Not valid json',
    'enum': 'Enum',
    'enum_msg': 'One value per line',
    'enum_desc': 'desc',
    'enum_desc_msg': 'enum description',
    'required': 'Required',
    'mock': 'mock',
    'mockLink': 'Help',
    'format': 'Format',
    'none': 'None',
    'preview': 'Preview',
    'tableMode': 'Display: Table mode',
    'readonly': 'Readonly'
  }
}

export default (lang) => {
  return langs[lang] || langs.en_US
}
