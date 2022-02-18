"use strict";(self.webpackChunkstencil=self.webpackChunkstencil||[]).push([[930],{2379:function(e,t,n){n.r(t),n.d(t,{frontMatter:function(){return a},contentTitle:function(){return l},metadata:function(){return c},toc:function(){return p},default:function(){return g}});var r=n(7462),s=n(3366),i=(n(7294),n(3905)),o=["components"],a={},l="Go",c={unversionedId:"clients/go",id:"clients/go",isDocsHomePage:!1,title:"Go",description:"Go Reference",source:"@site/docs/clients/go.md",sourceDirName:"clients",slug:"/clients/go",permalink:"/stencil/docs/clients/go",editUrl:"https://github.com/odpf/stencil/edit/master/docs/docs/clients/go.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"Overview",permalink:"/stencil/docs/clients/overview"},next:{title:"Java",permalink:"/stencil/docs/clients/java"}},p=[{value:"Requirements",id:"requirements",children:[]},{value:"Installation",id:"installation",children:[]},{value:"Usage",id:"usage",children:[{value:"Creating a client",id:"creating-a-client",children:[]},{value:"Get Descriptor",id:"get-descriptor",children:[]},{value:"Parse protobuf message.",id:"parse-protobuf-message",children:[]},{value:"Serialize data.",id:"serialize-data",children:[]},{value:"Enable auto refresh of schemas",id:"enable-auto-refresh-of-schemas",children:[]},{value:"Using VersionBasedRefresh strategy",id:"using-versionbasedrefresh-strategy",children:[]}]}],u={toc:p};function g(e){var t=e.components,n=(0,s.Z)(e,o);return(0,i.kt)("wrapper",(0,r.Z)({},u,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h1",{id:"go"},"Go"),(0,i.kt)("p",null,(0,i.kt)("a",{parentName:"p",href:"https://pkg.go.dev/github.com/odpf/stencil/clients/go"},(0,i.kt)("img",{parentName:"a",src:"https://pkg.go.dev/badge/github.com/odpf/stencil/clients/go.svg",alt:"Go Reference"}))),(0,i.kt)("p",null,"Stencil go client package provides a store to lookup protobuf descriptors and options to keep the protobuf descriptors upto date."),(0,i.kt)("p",null,"It has following features"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Deserialize protobuf messages directly by specifying protobuf message name"),(0,i.kt)("li",{parentName:"ul"},"Serialize data by specifying protobuf message name"),(0,i.kt)("li",{parentName:"ul"},"Ability to refresh protobuf descriptors in specified intervals"),(0,i.kt)("li",{parentName:"ul"},"Support to download descriptors from multiple urls")),(0,i.kt)("h2",{id:"requirements"},"Requirements"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"go 1.16")),(0,i.kt)("h2",{id:"installation"},"Installation"),(0,i.kt)("p",null,"Use ",(0,i.kt)("inlineCode",{parentName:"p"},"go get")),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"go get github.com/odpf/stencil/clients/go\n")),(0,i.kt)("p",null,"Then import the stencil package into your own code as mentioned below"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n')),(0,i.kt)("h2",{id:"usage"},"Usage"),(0,i.kt)("h3",{id:"creating-a-client"},"Creating a client"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n\nurl := "http://localhost:8000/v1beta1/namespaces/{test-namespace}/schemas/{schema-name}"\nclient, err := stencil.NewClient([]string{url}, stencil.Options{})\n')),(0,i.kt)("h3",{id:"get-descriptor"},"Get Descriptor"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n\nurl := "http://localhost:8000/v1beta1/namespaces/{test-namespace}/schemas/{schema-name}"\nclient, err := stencil.NewClient([]string{url}, stencil.Options{})\nif err != nil {\n    return\n}\ndesc, err := client.GetDescriptor("google.protobuf.DescriptorProto")\n')),(0,i.kt)("h3",{id:"parse-protobuf-message"},"Parse protobuf message."),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n\nurl := "http://localhost:8000/v1beta1/namespaces/{test-namespace}/schemas/{schema-name}"\nclient, err := stencil.NewClient([]string{url}, stencil.Options{})\nif err != nil {\n    return\n}\ndata := []byte("")\nparsedMsg, err := client.Parse("google.protobuf.DescriptorProto", data)\n')),(0,i.kt)("h3",{id:"serialize-data"},"Serialize data."),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n\nurl := "http://url/to/proto/descriptorset/file"\nclient, err := stencil.NewClient([]string{url}, stencil.Options{})\nif err != nil {\n    return\n}\ndata := map[string]interface{}{}\nserializedMsg, err := client.Serialize("google.protobuf.DescriptorProto", data)\n')),(0,i.kt)("h3",{id:"enable-auto-refresh-of-schemas"},"Enable auto refresh of schemas"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n\nurl := "http://localhost:8000/v1beta1/namespaces/{test-namespace}/schemas/{schema-name}"\n// Configured to refresh schema every 12 hours\nclient, err := stencil.NewClient([]string{url}, stencil.Options{AutoRefresh: true, RefreshInterval: time.Hours * 12})\nif err != nil {\n    return\n}\ndesc, err := client.GetDescriptor("google.protobuf.DescriptorProto")\n')),(0,i.kt)("h3",{id:"using-versionbasedrefresh-strategy"},"Using VersionBasedRefresh strategy"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-go"},'import stencil "github.com/odpf/stencil/clients/go"\n\nurl := "http://localhost:8000/v1beta1/namespaces/{test-namespace}/schemas/{schema-name}"\n// Configured to refresh schema every 12 hours\nclient, err := stencil.NewClient([]string{url}, stencil.Options{AutoRefresh: true, RefreshInterval: time.Hours * 12, RefreshStrategy: stencil.VersionBasedRefresh})\nif err != nil {\n    return\n}\ndesc, err := client.GetDescriptor("google.protobuf.DescriptorProto")\n')),(0,i.kt)("p",null,"Refer to ",(0,i.kt)("a",{parentName:"p",href:"https://pkg.go.dev/github.com/odpf/stencil/clients/go"},"go documentation")," for all available methods and options."))}g.isMDXComponent=!0}}]);