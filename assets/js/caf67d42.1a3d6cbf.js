"use strict";(self.webpackChunkstencil=self.webpackChunkstencil||[]).push([[794],{9742:function(e,t,n){n.r(t),n.d(t,{frontMatter:function(){return o},contentTitle:function(){return a},metadata:function(){return c},toc:function(){return u},default:function(){return p}});var i=n(7462),r=n(3366),l=(n(7294),n(3905)),s=["components"],o={},a="Overview",c={unversionedId:"clients/overview",id:"clients/overview",isDocsHomePage:!1,title:"Overview",description:"Stencil clients abstracts handling of descriptorset file on client side. Currently we officially support Stencil client in Java, Go, JS languages.",source:"@site/docs/clients/overview.md",sourceDirName:"clients",slug:"/clients/overview",permalink:"/stencil/docs/clients/overview",editUrl:"https://github.com/odpf/stencil/edit/master/docs/docs/clients/overview.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"API",permalink:"/stencil/docs/server/api"},next:{title:"Go",permalink:"/stencil/docs/clients/go"}},u=[{value:"Features",id:"features",children:[]},{value:"A note on configuring Stencil clients",id:"a-note-on-configuring-stencil-clients",children:[]},{value:"Languages",id:"languages",children:[]}],d={toc:u};function p(e){var t=e.components,n=(0,r.Z)(e,s);return(0,l.kt)("wrapper",(0,i.Z)({},d,n,{components:t,mdxType:"MDXLayout"}),(0,l.kt)("h1",{id:"overview"},"Overview"),(0,l.kt)("p",null,"Stencil clients abstracts handling of descriptorset file on client side. Currently we officially support Stencil client in Java, Go, JS languages."),(0,l.kt)("h2",{id:"features"},"Features"),(0,l.kt)("ul",null,(0,l.kt)("li",{parentName:"ul"},"downloading of descriptorset file from server"),(0,l.kt)("li",{parentName:"ul"},"parse API to deserialize protobuf encoded messages"),(0,l.kt)("li",{parentName:"ul"},"lookup API to find proto descriptors"),(0,l.kt)("li",{parentName:"ul"},"inbuilt strategies to refresh protobuf schema definitions.")),(0,l.kt)("h2",{id:"a-note-on-configuring-stencil-clients"},"A note on configuring Stencil clients"),(0,l.kt)("ul",null,(0,l.kt)("li",{parentName:"ul"},"Stencil server provides API to download latest descriptor file. If new version is available latest file will point to new descriptor file. Always use latest version proto descriptor url for stencil client if you want to refresh schema definitions in runtime."),(0,l.kt)("li",{parentName:"ul"},"Keep the refresh intervals relatively large (eg: 24hrs or 12 hrs) to reduce the number of calls depending on how fast systems produce new messages using new proto schema."),(0,l.kt)("li",{parentName:"ul"},"You can refresh descriptor file only if unknowns fields are faced by the client while parsing. This reduces unneccessary frequent calls made by clients. Currently this feature supported in JAVA and GO clients.")),(0,l.kt)("h2",{id:"languages"},"Languages"),(0,l.kt)("ul",null,(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("a",{parentName:"li",href:"java"},"Java")),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("a",{parentName:"li",href:"go"},"Go")),(0,l.kt)("li",{parentName:"ul"},(0,l.kt)("a",{parentName:"li",href:"js"},"Javascript")),(0,l.kt)("li",{parentName:"ul"},"Ruby - Coming soon"),(0,l.kt)("li",{parentName:"ul"},"Pytho - Coming soon")))}p.isMDXComponent=!0}}]);