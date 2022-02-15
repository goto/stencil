"use strict";(self.webpackChunkstencil=self.webpackChunkstencil||[]).push([[413],{4607:function(e,t,a){a.r(t),a.d(t,{frontMatter:function(){return o},contentTitle:function(){return d},metadata:function(){return s},toc:function(){return E},default:function(){return p}});var n=a(7462),r=a(3366),l=(a(7294),a(3905)),i=["components"],o={},d="Compatability rules",s={unversionedId:"server/rules",id:"server/rules",isDocsHomePage:!1,title:"Compatability rules",description:"Stencil server provides following compatibility rules.",source:"@site/docs/server/rules.md",sourceDirName:"server",slug:"/server/rules",permalink:"/stencil/docs/server/rules",editUrl:"https://github.com/odpf/stencil/edit/master/docs/docs/server/rules.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"Stencil server",permalink:"/stencil/docs/server/overview"},next:{title:"API",permalink:"/stencil/docs/server/api"}},E=[{value:"Feature support matrix",id:"feature-support-matrix",children:[]},{value:"Protobuf compatibility rules",id:"protobuf-compatibility-rules",children:[{value:"Rules",id:"rules",children:[]},{value:"List of Checks",id:"list-of-checks",children:[]}]}],u={toc:E};function p(e){var t=e.components,a=(0,r.Z)(e,i);return(0,l.kt)("wrapper",(0,n.Z)({},u,a,{components:t,mdxType:"MDXLayout"}),(0,l.kt)("h1",{id:"compatability-rules"},"Compatability rules"),(0,l.kt)("p",null,"Stencil server provides following compatibility rules."),(0,l.kt)("ul",null,(0,l.kt)("li",{parentName:"ul"},"BACKWARD_COMPATABILITY"),(0,l.kt)("li",{parentName:"ul"},"FORWARD_COMPATABILITY"),(0,l.kt)("li",{parentName:"ul"},"FULL_COMPATABILITY")),(0,l.kt)("p",null,"Stencil currently supports protobuf, avro and json schema formats. Compatibility rules for each schema format has been built separately considering each schema format's features."),(0,l.kt)("h2",{id:"feature-support-matrix"},"Feature support matrix"),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Compatability rule"),(0,l.kt)("th",{parentName:"tr",align:null},"Protobuf"),(0,l.kt)("th",{parentName:"tr",align:null},"Avro"),(0,l.kt)("th",{parentName:"tr",align:null},"JSON"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"BACKWARD_COMPATABILITY"),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"No")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FORWARD_COMPATABILITY"),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"No")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FULL_COMPATABILITY"),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"Yes"),(0,l.kt)("td",{parentName:"tr",align:null},"No")))),(0,l.kt)("h2",{id:"protobuf-compatibility-rules"},"Protobuf compatibility rules"),(0,l.kt)("p",null,"Protobuf compatability rules composed of ",(0,l.kt)("a",{parentName:"p",href:"#list-of-checks"},"compatability checks"),"."),(0,l.kt)("h3",{id:"rules"},"Rules"),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Compatibility name"),(0,l.kt)("th",{parentName:"tr",align:null},"List of checks"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"BACKWARD_COMPATABILITY"),(0,l.kt)("td",{parentName:"tr",align:null},"SYNTAX_CHANGE, MESSAGE_DELETE, NON_INCLUSIVE_RESERVED_RANGE, NON_INCLUSIVE_RESERVED_NAMES, FIELD_DELETE, FIELD_JSON_NAME_CHANGE, FIELD_LABEL_CHANGE, FIELD_KIND_CHANGE, FIELD_TYPE_CHANGE, ENUM_DELETE, ENUM_VALUE_DELETE, ENUM_VALUE_NUMBER_CHANGE")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FORWARD_COMPATABILITY"),(0,l.kt)("td",{parentName:"tr",align:null},"SYNTAX_CHANGE, MESSAGE_DELETE, NON_INCLUSIVE_RESERVED_RANGE, NON_INCLUSIVE_RESERVED_NAMES, FIELD_JSON_NAME_CHANGE, FIELD_LABEL_CHANGE, FIELD_KIND_CHANGE, FIELD_TYPE_CHANGE, FIELD_DELETE_WITHOUT_RESERVED_NUMBER, FIELD_DELETE_WITHOUT_RESERVED_NAME, ENUM_DELETE, ENUM_VALUE_NUMBER_CHANGE, ENUM_VALUE_DELETE_WITHOUT_RESERVEDNUMBER, ENUM_VALUE_DELETE_WITHOUT_RESERVEDNAME")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FULL_COMPATABILITY"),(0,l.kt)("td",{parentName:"tr",align:null},"SYNTAX_CHANGE, MESSAGE_DELETE, NON_INCLUSIVE_RESERVED_RANGE, NON_INCLUSIVE_RESERVED_NAMES, FIELD_DELETE, FIELD_JSON_NAME_CHANGE, FIELD_LABEL_CHANGE, FIELD_KIND_CHANGE, FIELD_TYPE_CHANGE, ENUM_DELETE, ENUM_VALUE_DELETE, ENUM_VALUE_NUMBER_CHANGE")))),(0,l.kt)("h3",{id:"list-of-checks"},"List of Checks"),(0,l.kt)("table",null,(0,l.kt)("thead",{parentName:"table"},(0,l.kt)("tr",{parentName:"thead"},(0,l.kt)("th",{parentName:"tr",align:null},"Check"),(0,l.kt)("th",{parentName:"tr",align:null},"Description"))),(0,l.kt)("tbody",{parentName:"table"},(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"SYNTAX_CHANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"checks if proto file syntax does not switch between proto2 and proto3, including going to/from unset (which assumes proto2) to set to proto3. Changing the syntax results in differences in generated code for many languages.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"MESSAGE_DELETE"),(0,l.kt)("td",{parentName:"tr",align:null},"checks that messages are deleted from a given file. Deleting a message will delete the corresponding generated type, which could be referenced in source code. Instead of deleting these types, deprecate them using ",(0,l.kt)("a",{parentName:"td",href:"https://developers.google.com/protocol-buffers/docs/proto3#options"},(0,l.kt)("inlineCode",{parentName:"a"},"deprecated")," option"),".")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"NON_INCLUSIVE_RESERVED_NAMES"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if current reserved names contains all previous reserved names.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"NON_INCLUSIVE_RESERVED_RANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if current reserve range inclusive of previous reserved range. This check ensures previous reserved tag numbers haven't been deleted")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_DELETE"),(0,l.kt)("td",{parentName:"tr",align:null},"checks that no message field is deleted. Deleting message field will result in the field being deleted from the generated source code, which could be referenced. Instead of deleting these, deprecate them using ",(0,l.kt)("a",{parentName:"td",href:"https://developers.google.com/protocol-buffers/docs/proto3#options"},(0,l.kt)("inlineCode",{parentName:"a"},"deprecated")," option"),".")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_JSON_NAME_CHANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if the json_name for field does not change, which would break JSON compatibility.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_LABEL_CHANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"checks that no field changes it's label, i.e. ",(0,l.kt)("inlineCode",{parentName:"td"},"optional"),", ",(0,l.kt)("inlineCode",{parentName:"td"},"required"),", ",(0,l.kt)("inlineCode",{parentName:"td"},"repeated"),". Changing to/from optional/required and repeated will be a generated source code and JSON breaking change. Changing to/from optional and repeated is actually not a wire-breaking change, however changing to/from optional and required is. Given that it's unlikely to be advisable in any situation to change your label, and that there is only one exception, we find it best to just outlaw this entirely.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_KIND_CHANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"checks that a field has the same type. Changing the type of a field can affect the type in the generated source code, wire compatibility, and JSON compatibility.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_TYPE_CHANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if message/enum field it's message/enum type has changed from previous version. This rule only applies to message kind and enum kind.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_DELETE_WITHOUT_RESERVED_NUMBER"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if field is deleted, it's tag number should be added to reserved numbers. This will ensure deleted field tag number won't be used in future.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"FIELD_DELETE_WITHOUT_RESERVED_NAME"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if field is deleted, it's tag name should be added to reserved names. This will help to keep the JSON compatibility.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"ENUM_DELETE"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks that no enum is deleted from current version. Deleting an enum will delete the corresponding generated type, which could be referenced in source code. Instead of deleting these types, deprecate them using ",(0,l.kt)("inlineCode",{parentName:"td"},"deprecated")," option.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"ENUM_VALUE_DELETE"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks that no enum value is deleted. Deleting an enum value will result in the corresponding value being deleted from the generated source code, which could be referenced. Instead of deleting these, deprecate them.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"ENUM_VALUE_DELETE_WITHOUT_RESERVEDNUMBER"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if enum value deleted, it's enum number should be added to reserved numbers.")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"ENUM_VALUE_DELETE_WITHOUT_RESERVEDNAME"),(0,l.kt)("td",{parentName:"tr",align:null},"Checks if enum value deleted, it's enum name should be added to reserved names. This will help to keep the JSON compatability")),(0,l.kt)("tr",{parentName:"tbody"},(0,l.kt)("td",{parentName:"tr",align:null},"ENUM_VALUE_NUMBER_CHANGE"),(0,l.kt)("td",{parentName:"tr",align:null},"Check if enum number has changed between current, previous versions. For example You cannot change FOO_ONE = 1 to FOO_ONE = 2. Doing so will result in potential JSON incompatibilites and broken source code.")))))}p.isMDXComponent=!0}}]);