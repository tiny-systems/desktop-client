const value = {
    description: null,
    maximum: null,
    minimum: null,
    exclusiveMaximum:null,
    exclusiveMinimum:null,
    enum:[]
}
const attr = {
    description: {
      name: 'Description',
      type: 'string',
    },
    maximum:{
        name:'Max value',
        type:'integer'
    },
    minimum:{
        name:'Min value',
        type:'integer'
    },
    exclusiveMaximum:{
        name:'Exclusive maximum',
        type:'integer'
    },
    exclusiveMinimum:{
        name:'Exclusive minimum',
        type:'integer'
    },
    enum:{
        name:'Enum',
        type:'array'
    }
}
const wrapper = {value, attr}
export default wrapper
