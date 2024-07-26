"use strict";(self.webpackChunkstencil=self.webpackChunkstencil||[]).push([[234],{3905:function(e,t,n){n.d(t,{Zo:function(){return p},kt:function(){return h}});var a=n(7294);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function s(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?r(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function o(e,t){if(null==e)return{};var n,a,i=function(e,t){if(null==e)return{};var n,a,i={},r=Object.keys(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var c=a.createContext({}),l=function(e){var t=a.useContext(c),n=t;return e&&(n="function"==typeof e?e(t):s(s({},t),e)),n},p=function(e){var t=l(e.components);return a.createElement(c.Provider,{value:t},e.children)},d={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},m=a.forwardRef((function(e,t){var n=e.components,i=e.mdxType,r=e.originalType,c=e.parentName,p=o(e,["components","mdxType","originalType","parentName"]),m=l(n),h=i,u=m["".concat(c,".").concat(h)]||m[h]||d[h]||r;return n?a.createElement(u,s(s({ref:t},p),{},{components:n})):a.createElement(u,s({ref:t},p))}));function h(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var r=n.length,s=new Array(r);s[0]=m;var o={};for(var c in t)hasOwnProperty.call(t,c)&&(o[c]=t[c]);o.originalType=e,o.mdxType="string"==typeof e?e:i,s[1]=o;for(var l=2;l<r;l++)s[l]=n[l];return a.createElement.apply(null,s)}return a.createElement.apply(null,n)}m.displayName="MDXCreateElement"},113:function(e,t,n){n.r(t),n.d(t,{assets:function(){return p},contentTitle:function(){return c},default:function(){return h},frontMatter:function(){return o},metadata:function(){return l},toc:function(){return d}});var a=n(7462),i=n(3366),r=(n(7294),n(3905)),s=["components"],o={},c="Detect Schema Change",l={unversionedId:"guides/schema_change",id:"guides/schema_change",title:"Detect Schema Change",description:"Enabling Schema Change",source:"@site/docs/guides/5_schema_change.md",sourceDirName:"guides",slug:"/guides/schema_change",permalink:"/stencil/docs/guides/schema_change",editUrl:"https://github.com/goto/stencil/edit/master/docs/docs/guides/5_schema_change.md",tags:[],version:"current",sidebarPosition:5,frontMatter:{}},p={},d=[{value:"Enabling Schema Change",id:"enabling-schema-change",level:2},{value:"Create a schema",id:"create-a-schema",level:2},{value:"Handling Failures",id:"handling-failures",level:2},{value:"Depth",id:"depth",level:2},{value:"Sample SchemaChangedEvent Object",id:"sample-schemachangedevent-object",level:2},{value:"Showing MR information",id:"showing-mr-information",level:2},{value:"Reconciliation API",id:"reconciliation-api",level:2}],m={toc:d};function h(e){var t=e.components,n=(0,i.Z)(e,s);return(0,r.kt)("wrapper",(0,a.Z)({},m,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"detect-schema-change"},"Detect Schema Change"),(0,r.kt)("h2",{id:"enabling-schema-change"},"Enabling Schema Change"),(0,r.kt)("p",null,"To enable schema change toggle on the flag ",(0,r.kt)("inlineCode",{parentName:"p"},"SCHEMACHANGE_ENABLE")," and provide ",(0,r.kt)("inlineCode",{parentName:"p"},"KAFKAPRODUCER_BOOTSTRAPSERVER"),"\nand ",(0,r.kt)("inlineCode",{parentName:"p"},"SCHEMACHANGE_KAFKATOPIC"),"."),(0,r.kt)("h2",{id:"create-a-schema"},"Create a schema"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-bash"},'# create namespace named "quickstart"\ncurl -X POST http://localhost:8000/v1beta1/namespaces -H \'Content-Type: application/json\' -d \'{"id": "quickstart", "format": "FORMAT_PROTOBUF", "compatibility": "COMPATIBILITY_BACKWARD", "description": "This field can be used to store namespace description"}\'\n\n# create descriptor file v1\nprotoc --descriptor_set_out=./protos/1.desc --include_imports ./guide/protos/1.proto\n\nIt will be version 1\n# upload generated proto descriptor file to server with schema name as `example` under `quickstart` namespace.\ncurl -H "X-SourceURL:www.github.com/some-repo" -H "X-CommitSHA:some-commit-sha" -X POST http://localhost:8000/v1beta1/namespaces/quickstart/schemas/example --data-binary "@1.desc"\n\n# create descriptor file v2\nprotoc --descriptor_set_out=./protos/1.desc --include_imports ./guide/protos/2.proto\n\nIt will be version 2\n\n# upload generated proto descriptor file to server with schema name as `example` under `quickstart` namespace.\ncurl -X POST http://localhost:8000/v1beta1/namespaces/quickstart/schemas/example --data-binary "@2.desc"\n\n# Detect Schema Change\nNow, when the schema is uploaded in DB. Below things happened parallely in separate go routine\n1. It will get the previous schema data corresponding to that namespace and schema name.\n2. It will check all the messages in that schema whether they are impacted(directly/indireclty) or not. \n3. For change, it will check if any fileds got added/deleted. In this case only `address` field is added in `friend` message. So, `friend` proto is only impacted.\n4. In this way it gets all impacted messages. And do below steps for all messages.\n5. Then it will get the all indirectly impacted messages due to the message `friend`. \n6. So, it will get `User` as it is using `Friend` message as compostion.\n7. It will do the same thing for `User` also until all impacted messages got proccessed.\n8. Then it will make `SchemChangedEvent` object and push it to a kafka topic.\n9. After that it will update the `notification_events` database table with `success=true` after pushing the object to kafka. \n')),(0,r.kt)("h2",{id:"handling-failures"},"Handling Failures"),(0,r.kt)("p",null,"In case of any failure while publishing to Kafka or other issues, the ",(0,r.kt)("inlineCode",{parentName:"p"},"notification_events")," table can be utilized to\nmanage these failures. If the ",(0,r.kt)("inlineCode",{parentName:"p"},"success")," field is ",(0,r.kt)("inlineCode",{parentName:"p"},"false")," for a specific ",(0,r.kt)("inlineCode",{parentName:"p"},"namespace_id"),", ",(0,r.kt)("inlineCode",{parentName:"p"},"schema_id"),", and ",(0,r.kt)("inlineCode",{parentName:"p"},"version_id"),",\nit indicates that the notification was not sent. In such scenarios, the ",(0,r.kt)("a",{parentName:"p",href:"/stencil/docs/reference/api#reconciliation-api"},"Reconciliation API")," can be used to resend the\nnotifications."),(0,r.kt)("h2",{id:"depth"},"Depth"),(0,r.kt)("p",null,"In schema changes, we use the ",(0,r.kt)("inlineCode",{parentName:"p"},"depth")," parameter to determine the level of impacted schemas to retrieve. If ",(0,r.kt)("inlineCode",{parentName:"p"},"depth")," is\nset to an empty string ",(0,r.kt)("inlineCode",{parentName:"p"},'""')," or ",(0,r.kt)("inlineCode",{parentName:"p"},"-1"),", it will fetch all indirectly impacted schemas. However, if a specific depth is\nprovided, it will only retrieve schemas up to that specified level. For example, if ",(0,r.kt)("inlineCode",{parentName:"p"},"depth")," is set to ",(0,r.kt)("inlineCode",{parentName:"p"},"3"),", it will fetch\nup to three levels of impacted schemas, even if there are more levels of actual impacted schemas.\nE.g. If depth is ",(0,r.kt)("inlineCode",{parentName:"p"},"2")," then the ",(0,r.kt)("inlineCode",{parentName:"p"},"impacted_schemas")," for ",(0,r.kt)("inlineCode",{parentName:"p"},"Address")," in below sample would contain only ",(0,r.kt)("inlineCode",{parentName:"p"},"Address")," and ",(0,r.kt)("inlineCode",{parentName:"p"},"Friend")," message."),(0,r.kt)("h2",{id:"sample-schemachangedevent-object"},"Sample SchemaChangedEvent Object"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'{\n  "namespace_name": "quickstart",\n  "schema_name": "example",\n  "updated_schemas": [\n    "test.stencil.main.Friend",\n    "test.stencil.main.Address",\n  ],\n  "updated_fields": {\n    "test.stencil.main.Friend": {\n      "field_names": [\n        "address"\n      ]\n    },\n    "test.stencil.main.Address": {\n      "field_names": [\n        "city"\n      ]\n    }\n  },\n  "impacted_schemas": {\n    "test.stencil.main.Address": {\n      "schema_names": [\n        "test.stencil.main.Address",\n        "test.stencil.main.Friend",\n        "test.stencil.main.User",\n      ]\n    },\n    "test.stencil.main.Friend": {\n      "schema_names": [\n        "test.stencil.main.Friend",\n        "test.stencil.main.User",\n      ]\n    }\n  },\n  "version": 1,\n  "metadata": {\n    "source_url": "https://github.com/some-repo",\n    "commit_sha": "some-commit-sha"\n  }\n}\n')),(0,r.kt)("h2",{id:"showing-mr-information"},"Showing MR information"),(0,r.kt)("p",null,"While calculating the impacted schemas, the ",(0,r.kt)("inlineCode",{parentName:"p"},"SchemaChangedEvent")," will also include information about the source URL and commit SHA. The ",(0,r.kt)("inlineCode",{parentName:"p"},"source_url")," represents the repository URL, and the ",(0,r.kt)("inlineCode",{parentName:"p"},"commit_sha")," corresponds to the commit SHA associated with that version."),(0,r.kt)("h2",{id:"reconciliation-api"},"Reconciliation API"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-bash"},"$ curl -X GET http://localhost:8080/v1beta1/schema/detect-change/quickstart/example?from=1&to=2;\n")))}h.isMDXComponent=!0}}]);