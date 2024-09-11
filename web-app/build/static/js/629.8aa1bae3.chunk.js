"use strict";(self.webpackChunkweb_app=self.webpackChunkweb_app||[]).push([[629],{2237:(e,t,n)=>{n.d(t,{A:()=>l});var s=n(5043),a=n(579);const l=function(e){let t=arguments.length>1&&void 0!==arguments[1]?arguments[1]:null;return function(n){return(0,a.jsx)(s.Suspense,{fallback:t,children:(0,a.jsx)(e,{...n})})}}},2629:(e,t,n)=>{n.r(t),n.d(t,{default:()=>_});var s=n(5043),a=n(9456),l=n(9923),i=n(3216),o=n(4574),r=n(3097),d=n.n(r),c=n(2961),h=n(6483),x=n(9161),j=n(4159),m=n(9607),p=n(8296),b=n(7612),u=n(2237),y=n(6681),g=n(4770),f=n(579);const C=(0,u.A)(s.lazy((()=>n.e(122).then(n.bind(n,8122))))),A=(0,u.A)(s.lazy((()=>Promise.all([n.e(568),n.e(979)]).then(n.bind(n,5979))))),v=(0,u.A)(s.lazy((()=>Promise.all([n.e(241),n.e(481)]).then(n.bind(n,8481))))),z=(0,u.A)(s.lazy((()=>n.e(415).then(n.bind(n,2415))))),q=(0,u.A)(s.lazy((()=>n.e(414).then(n.bind(n,3414))))),$=(0,u.A)(s.lazy((()=>n.e(112).then(n.bind(n,112))))),k=(0,u.A)(s.lazy((()=>n.e(943).then(n.bind(n,3943))))),P=(0,u.A)(s.lazy((()=>n.e(732).then(n.bind(n,732))))),T=(0,u.A)(s.lazy((()=>n.e(204).then(n.bind(n,2204))))),E=(0,u.A)(s.lazy((()=>n.e(104).then(n.bind(n,104))))),N=(0,u.A)(s.lazy((()=>n.e(728).then(n.bind(n,3728))))),B=(0,u.A)(s.lazy((()=>n.e(682).then(n.bind(n,4682))))),w=(0,u.A)(s.lazy((()=>Promise.all([n.e(241),n.e(723)]).then(n.bind(n,3104))))),F=(0,u.A)(s.lazy((()=>Promise.all([n.e(241),n.e(713)]).then(n.bind(n,8094))))),M=(0,u.A)(s.lazy((()=>n.e(461).then(n.bind(n,8461))))),S=(0,u.A)(s.lazy((()=>Promise.all([n.e(98),n.e(583)]).then(n.bind(n,5367))))),L=(0,u.A)(s.lazy((()=>n.e(641).then(n.bind(n,9641))))),R=o.Ay.div((e=>{let{theme:t}=e;return{position:"relative",fontSize:10,left:26,height:10,top:4,"& .statusIcon":{color:d()(t,"signalColors.disabled","#E6EBEB"),"&.red":{color:d()(t,"signalColors.danger","#C51B3F")},"&.yellow":{color:d()(t,"signalColors.warning","#FFBD62")},"&.green":{color:d()(t,"signalColors.good","#4CCB92")}}}})),_=()=>{var e;const t=(0,c.jL)(),n=(0,i.g)(),o=(0,i.Zp)(),{pathname:r=""}=(0,i.zy)(),d=(0,a.d4)((e=>e.tenants.loadingTenant)),u=(0,a.d4)((e=>e.tenants.currentTenant)),_=(0,a.d4)((e=>e.tenants.currentNamespace)),D=(0,a.d4)((e=>e.tenants.tenantInfo)),I=n.tenantName||"",O=n.tenantNamespace||"",[V,Y]=(0,s.useState)(!1);(0,s.useEffect)((()=>{_===O&&u===I||(t((0,m.s1)({name:I,namespace:O})),t((0,p.X)()))}),[u,_,t,I,O]);const G=e=>`/namespaces/${O}/tenants/${I}/${e}`;return(0,f.jsxs)(s.Fragment,{children:[V&&null!==D&&(0,f.jsx)(M,{deleteOpen:V,selectedTenant:D,closeDeleteModalAndRefresh:e=>{Y(!1),e&&(t((0,j.Hk)("Tenant Deleted")),o("/tenants"))}}),(0,f.jsx)(g.A,{label:(0,f.jsx)(s.Fragment,{children:(0,f.jsx)(l.EGL,{label:"Tenants",onClick:()=>o(x.zZ.TENANTS)})}),actions:(0,f.jsx)(s.Fragment,{})}),(0,f.jsxs)(l.Mxu,{variant:"constrained",children:[d&&(0,f.jsx)(l.xA9,{item:!0,xs:12,children:(0,f.jsx)(l.z21,{})}),(0,f.jsx)(l.azJ,{withBorders:!0,customBorderPadding:"0px",sx:{borderBottom:0},children:(0,f.jsx)(l.lcx,{icon:(0,f.jsxs)(s.Fragment,{children:[(0,f.jsx)(R,{children:D&&D.status&&(0,f.jsx)("span",{className:`statusIcon ${null===(e=D.status)||void 0===e?void 0:e.health_status}`,children:(0,f.jsx)(l.GQ2,{style:{width:15,height:15}})})}),(0,f.jsx)(l.fmr,{})]}),title:I,subTitle:(0,f.jsxs)(s.Fragment,{children:["Namespace: ",O," / Capacity:"," ",(0,h.nO)(((null===D||void 0===D?void 0:D.total_size)||0).toString(10))]}),actions:(0,f.jsxs)(l.azJ,{sx:{display:"flex",justifyContent:"flex-end",gap:10},children:[(0,f.jsx)(y.A,{tooltip:"Delete",children:(0,f.jsx)(l.$nd,{id:"delete-tenant",variant:"secondary",onClick:()=>{Y(!0)},color:"secondary",icon:(0,f.jsx)(l.ucK,{})})}),(0,f.jsx)(y.A,{tooltip:"Edit YAML",children:(0,f.jsx)(l.$nd,{icon:(0,f.jsx)(l.qUP,{}),id:"yaml_button",variant:"regular","aria-label":"Edit YAML",onClick:()=>{o(G("summary/yaml"))}})}),(0,f.jsx)(y.A,{tooltip:"Management Console",children:(0,f.jsx)(l.$nd,{id:"tenant-hop",onClick:()=>{o(`/namespaces/${O}/tenants/${I}/hop`)},disabled:!D||!(0,b.an)(D),variant:"regular",icon:(0,f.jsx)(l.$2v,{style:{height:16}})})}),(0,f.jsx)(y.A,{tooltip:"Refresh",children:(0,f.jsx)(l.$nd,{id:"tenant-refresh",variant:"regular","aria-label":"Refresh List",onClick:()=>{t((0,p.X)())},icon:(0,f.jsx)(l.fNY,{})})})]})})}),(0,f.jsx)(l.tUM,{currentTabOrPath:r,useRouteTabs:!0,onTabClick:e=>o(e),routes:(0,f.jsxs)(i.BV,{children:[(0,f.jsx)(i.qh,{path:"summary",element:(0,f.jsx)(A,{})}),(0,f.jsx)(i.qh,{path:"configuration",element:(0,f.jsx)(L,{})}),(0,f.jsx)(i.qh,{path:"summary/yaml",element:(0,f.jsx)(C,{})}),(0,f.jsx)(i.qh,{path:"metrics",element:(0,f.jsx)(T,{})}),(0,f.jsx)(i.qh,{path:"trace",element:(0,f.jsx)(E,{})}),(0,f.jsx)(i.qh,{path:"identity-provider",element:(0,f.jsx)(B,{})}),(0,f.jsx)(i.qh,{path:"security",element:(0,f.jsx)(w,{})}),(0,f.jsx)(i.qh,{path:"encryption",element:(0,f.jsx)(F,{})}),(0,f.jsx)(i.qh,{path:"pools",element:(0,f.jsx)(z,{})}),(0,f.jsx)(i.qh,{path:"pods/:podName",element:(0,f.jsx)(S,{})}),(0,f.jsx)(i.qh,{path:"pods",element:(0,f.jsx)(q,{})}),(0,f.jsx)(i.qh,{path:"pvcs/:PVCName",element:(0,f.jsx)(N,{})}),(0,f.jsx)(i.qh,{path:"volumes",element:(0,f.jsx)(P,{})}),(0,f.jsx)(i.qh,{path:"license",element:(0,f.jsx)(v,{})}),(0,f.jsx)(i.qh,{path:"events",element:(0,f.jsx)($,{})}),(0,f.jsx)(i.qh,{path:"csr",element:(0,f.jsx)(k,{})}),(0,f.jsx)(i.qh,{path:"/",element:(0,f.jsx)(i.C5,{to:`/namespaces/${O}/tenants/${I}/summary`})})]}),options:[{tabConfig:{label:"Summary",id:"details-summary",to:G("summary")}},{tabConfig:{label:"Configuration",id:"details-configuration",to:G("configuration")}},{tabConfig:{label:"Metrics",id:"details-metrics",to:G("metrics")}},{tabConfig:{label:"Identity Provider",id:"details-idp",to:G("identity-provider")}},{tabConfig:{label:"Security",id:"details-security",to:G("security")}},{tabConfig:{label:"Encryption",id:"details-encryption",to:G("encryption")}},{tabConfig:{label:"Pools",id:"details-pools",to:G("pools")}},{tabConfig:{label:"Pods",id:"tenant-pod-tab",to:G("pods")}},{tabConfig:{label:"Volumes",id:"details-volumes",to:G("volumes")}},{tabConfig:{label:"Events",id:"details-events",to:G("events")}},{tabConfig:{label:"Certificate Requests",id:"details-csr",to:G("csr")}},{tabConfig:{label:"License",id:"details-license",to:G("license")}}]})]})]})}}}]);
//# sourceMappingURL=629.8aa1bae3.chunk.js.map