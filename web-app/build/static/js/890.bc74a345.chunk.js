"use strict";(self.webpackChunkweb_app=self.webpackChunkweb_app||[]).push([[890],{2237:(e,t,a)=>{a.d(t,{A:()=>s});var n=a(5043),l=a(579);const s=function(e){let t=arguments.length>1&&void 0!==arguments[1]?arguments[1]:null;return function(a){return(0,l.jsx)(n.Suspense,{fallback:t,children:(0,l.jsx)(e,{...a})})}}},4681:(e,t,a)=>{a.d(t,{A:()=>s});a(5043);var n=a(9923),l=a(579);const s=e=>{let{placeholder:t="",onChange:a,overrideClass:s,value:i,id:r="search-resource",label:o="",sx:c}=e;return(0,l.jsx)(n.cl_,{placeholder:t,className:s||"",id:r,label:o,onChange:e=>{a(e.target.value)},value:i,startIcon:(0,l.jsx)(n.WIv,{}),sx:c})}},8890:(e,t,a)=>{a.r(t),a.d(t,{default:()=>N});var n=a(5043),l=a(9923),s=a(4159),i=a(3216),r=a(2961),o=a(9607),c=a(8296),d=a(6483),u=a(4574),x=a(3097),h=a.n(x),m=a(579);const g=u.Ay.div((e=>{let{theme:t}=e;return{margin:"0px 20px","& .value":{fontSize:18,color:h()(t,"mutedText","#87888d"),fontWeight:400,"&.normal":{color:h()(t,"fontColor","#000")}},"& .unit":{fontSize:12,color:h()(t,"secondaryText","#5B5C5C"),fontWeight:"bold"},"& .label":{textAlign:"center",color:h()(t,"mutedText","#87888d"),fontSize:12,whiteSpace:"nowrap","&.normal":{color:h()(t,"secondaryText","#5B5C5C")}}}})),p=e=>{let{label:t,value:a,unit:s,variant:i="normal"}=e;return(0,m.jsxs)(g,{children:[(0,m.jsxs)(l.azJ,{style:{textAlign:"center"},children:[(0,m.jsx)("span",{className:`value ${i}`,children:a}),s&&(0,m.jsxs)(n.Fragment,{children:[" ",(0,m.jsx)("span",{className:"unit",children:s})]})]}),(0,m.jsx)(l.azJ,{className:"label",children:t})]})};var f=a(6129);const v=u.Ay.div((e=>{let{theme:t}=e;return{border:`${h()(t,"borderColor","#eaeaea")} 1px solid`,borderRadius:3,padding:15,cursor:"pointer","&.disabled":{backgroundColor:h()(t,"signalColors.danger","red")},"&:hover":{backgroundColor:h()(t,"boxBackground","#FBFAFA")},"& .tenantTitle":{display:"flex",alignItems:"center",justifyContent:"space-between",gap:10,"& h1":{padding:0,margin:0,marginBottom:5,fontSize:22,color:h()(t,"screenTitle.iconColor","#07193E"),[`@media (max-width: ${l.nmC.md}px)`]:{marginBottom:0}}},"& .tenantDetails":{display:"flex",gap:40,"& span":{fontSize:14},[`@media (max-width: ${l.nmC.md}px)`]:{flexFlow:"column-reverse",gap:5}},"& .tenantMetrics":{display:"flex",alignItems:"center",marginTop:20,gap:25,borderTop:`${h()(t,"borderColor","#E2E2E2")} 1px solid`,paddingTop:20,"& svg.tenantIcon":{color:h()(t,"screenTitle.iconColor","#07193E"),fill:h()(t,"screenTitle.iconColor","#07193E")},"& .metric":{"& .min-icon":{color:h()(t,"fontColor","#000"),width:13,marginRight:5}},"& .metricLabel":{fontSize:14,fontWeight:"bold",color:h()(t,"fontColor","#000")},"& .metricText":{fontSize:24,fontWeight:"bold"},"& .unit":{fontSize:12,fontWeight:"normal"},"& .status":{fontSize:12,color:h()(t,"mutedText","#87888d")},[`@media (max-width: ${l.nmC.md}px)`]:{marginTop:8,paddingTop:8}},"& .namespaceLabel":{display:"inline-flex",color:h()(t,"signalColors.dark","#000"),backgroundColor:h()(t,"borderColor","#E2E2E2"),borderRadius:2,padding:"4px 8px",fontSize:10,marginRight:20},"& .redState":{color:h()(t,"signalColors.danger","#C51B3F"),"& .min-icon":{width:16,height:16,float:"left",marginRight:4}},"& .yellowState":{color:h()(t,"signalColors.warning","#FFBD62"),"& .min-icon":{width:16,height:16,float:"left",marginRight:4}},"& .greenState":{color:h()(t,"signalColors.good","#4CCB92"),"& .min-icon":{width:16,height:16,float:"left",marginRight:4}},"& .greyState":{color:h()(t,"signalColors.disabled","#E6EBEB"),"& .min-icon":{width:16,height:16,float:"left",marginRight:4}}}})),j=e=>{let{tenant:t}=e;const a=(0,r.jL)(),s=(0,i.Zp)();let u={value:"n/a",unit:""},x={value:"n/a",unit:""},h={value:"n/a",unit:""},g={value:"n/a",unit:""},j={value:"n/a",unit:""};if(t.capacity_raw){const e=(0,d.nO)(`${t.capacity_raw}`,!0).split(" ");u.value=e[0],u.unit=e[1]}if(t.capacity){const e=(0,d.nO)(`${t.capacity}`,!0).split(" ");x.value=e[0],x.unit=e[1]}if(t.capacity_usage){const e=(0,d.qO)(t.capacity_usage,!0).split(" ");h.value=e[0],h.unit=e[1]}let y=[];if(t.tiers&&0!==t.tiers.length){var C,b;y=null===(C=t.tiers)||void 0===C?void 0:C.map((e=>({value:e.size,variant:e.name})));let e=null===(b=t.tiers)||void 0===b?void 0:b.filter((e=>"internal"===e.type)).reduce(((e,t)=>e+t.size),0),a=t.tiers.filter((e=>"internal"!==e.type)).reduce(((e,t)=>e+t.size),0);const n=(0,d.qO)(a,!0).split(" ");j.value=n[0],j.unit=n[1];const l=(0,d.qO)(e,!0).split(" ");g.value=l[0],g.unit=l[1]}else y=[{value:t.capacity_usage||0,variant:"STANDARD"}];return(0,m.jsx)(n.Fragment,{children:(0,m.jsx)(v,{id:`list-tenant-${t.name}`,onClick:()=>{a((0,o.s1)({name:t.name,namespace:t.namespace})),a((0,c.X)()),s(`/namespaces/${t.namespace}/tenants/${t.name}/summary`)},children:(0,m.jsxs)(l.xA9,{container:!0,children:[(0,m.jsxs)(l.xA9,{item:!0,xs:12,className:"tenantTitle",children:[(0,m.jsx)(l.azJ,{children:(0,m.jsx)("h1",{children:t.name})}),(0,m.jsx)(l.azJ,{children:(0,m.jsxs)("span",{className:"namespaceLabel",children:["Namespace:\xa0",t.namespace]})})]}),(0,m.jsx)(l.xA9,{item:!0,xs:12,sx:{marginTop:2},children:(0,m.jsxs)(l.xA9,{container:!0,children:[(0,m.jsx)(l.xA9,{item:!0,xs:2,children:(0,m.jsx)(f.A,{totalCapacity:t.capacity||0,usedSpaceVariants:y,statusClass:(e=>{switch(e){case"red":return"redState";case"yellow":return"yellowState";case"green":return"greenState";default:return"greyState"}})(t.health_status)})}),(0,m.jsxs)(l.xA9,{item:!0,xs:7,children:[(0,m.jsxs)(l.xA9,{item:!0,xs:!0,sx:{display:"flex",justifyContent:"flex-start",alignItems:"center",marginTop:"10px"},children:[(0,m.jsx)(p,{label:"Raw Capacity",value:u.value,unit:u.unit}),(0,m.jsx)(p,{label:"Usable Capacity",value:x.value,unit:x.unit}),(0,m.jsx)(p,{label:"Pools",value:`${t.pool_count}`})]}),(0,m.jsx)(l.xA9,{item:!0,xs:12,sx:{paddingLeft:"20px",marginTop:"15px"},children:(0,m.jsxs)("span",{className:"status",children:[(0,m.jsx)("strong",{children:"State:"})," ",t.currentState]})})]}),(0,m.jsx)(l.xA9,{item:!0,xs:3,children:(0,m.jsx)(n.Fragment,{children:(0,m.jsxs)(l.xA9,{container:!0,sx:{gap:20},children:[(0,m.jsxs)(l.xA9,{item:!0,xs:2,sx:{display:"flex",flexDirection:"column",alignItems:"center"},children:[(0,m.jsx)(l.JUN,{className:"muted",style:{width:25}}),(0,m.jsx)(l.azJ,{className:"muted",sx:{fontSize:12,fontWeight:"400"},children:"Usage"})]}),(0,m.jsxs)(l.xA9,{item:!0,xs:9,sx:{paddingTop:8},children:[(!t.tiers||0===t.tiers.length)&&(0,m.jsxs)(l.azJ,{sx:{fontSize:14,fontWeight:400},children:[(0,m.jsx)("span",{className:"muted",children:"Internal: "})," ",`${h.value} ${h.unit}`]}),t.tiers&&t.tiers.length>0&&(0,m.jsxs)(n.Fragment,{children:[(0,m.jsxs)(l.azJ,{sx:{fontSize:14,fontWeight:400},children:[(0,m.jsx)("span",{className:"muted",children:"Internal: "})," ",`${g.value} ${g.unit}`]}),(0,m.jsxs)(l.azJ,{sx:{fontSize:14,fontWeight:400},children:[(0,m.jsx)("span",{className:"muted",children:"Tiered: "})," ",`${j.value} ${j.unit}`]})]})]})]})})})]})})]})})})};var y=a(2237),C=a(5271),b=a(3461),A=a(5098);let S={};const w=e=>{let{rowRenderFunction:t,totalItems:a,defaultHeight:l}=e;const s=e=>{let{index:a,style:n}=e;return(0,m.jsx)("div",{style:n,children:t(a)})};return(0,m.jsx)(n.Fragment,{children:(0,m.jsx)(b.A,{isItemLoaded:e=>!!S[e],loadMoreItems:(e,t)=>{for(let a=e;a<=t;a++)S[a]=1;for(let a=e;a<=t;a++)S[a]=2},itemCount:a,children:e=>{let{onItemsRendered:t,ref:n}=e;return(0,m.jsx)(A.t$,{children:e=>{let{width:i,height:r}=e;return(0,m.jsx)(C.Y1,{itemSize:l||220,height:r,itemCount:a,width:i,ref:n,onItemsRendered:t,children:s})}})}})})};var T=a(4681),z=a(6681),_=a(4770),F=a(7984);const $=(0,y.A)(n.lazy((()=>a.e(619).then(a.bind(a,8619))))),N=()=>{const e=(0,r.jL)(),t=(0,i.Zp)(),[a,o]=(0,n.useState)(!1),[c,d]=(0,n.useState)(""),[u,x]=(0,n.useState)([]),[h,g]=(0,n.useState)(!1),[p,f]=(0,n.useState)(null),[v,y]=(0,n.useState)("name"),C=u.filter((e=>""===c||e.name.indexOf(c)>=0));C.sort(((e,t)=>{switch(v){case"capacity":return e.capacity&&t.capacity?e.capacity>t.capacity?1:e.capacity<t.capacity?-1:0:0;case"usage":return e.capacity_usage&&t.capacity_usage?e.capacity_usage>t.capacity_usage?1:e.capacity_usage<t.capacity_usage?-1:0:0;case"active_status":return"red"===e.health_status&&"red"!==t.health_status?1:"red"!==e.health_status&&"red"===t.health_status?-1:0;case"failing_status":return"green"===e.health_status&&"green"!==t.health_status?1:"green"!==e.health_status&&"green"===t.health_status?-1:0;default:return e.name>t.name?1:e.name<t.name?-1:0}})),(0,n.useEffect)((()=>{if(a){(()=>{F.F.tenants.listAllTenants().then((e=>{var t;if(!e.data)return void o(!1);let a=null!==(t=e.data.tenants)&&void 0!==t?t:[];x(a),o(!1)})).catch((t=>{e((0,s.C9)(t)),o(!1)}))})()}}),[a,e]),(0,n.useEffect)((()=>{o(!0)}),[]);return(0,m.jsxs)(n.Fragment,{children:[h&&(0,m.jsx)($,{newServiceAccount:p,open:h,closeModal:()=>{g(!1),f(null)},entity:"Tenant"}),(0,m.jsx)(_.A,{label:"Tenants",middleComponent:(0,m.jsx)(T.A,{placeholder:"Filter Tenants",onChange:e=>{d(e)},value:c}),actions:(0,m.jsxs)(l.xA9,{item:!0,xs:12,sx:{display:"flex",justifyContent:"flex-end"},children:[(0,m.jsx)(z.A,{tooltip:"Refresh Tenant List",children:(0,m.jsx)(l.$nd,{id:"refresh-tenant-list",onClick:()=>{o(!0)},icon:(0,m.jsx)(l.fNY,{}),variant:"regular"})}),(0,m.jsx)(z.A,{tooltip:"Create Tenant",children:(0,m.jsx)(l.$nd,{id:"create-tenant",label:"Create Tenant",onClick:()=>{t("/tenants/add")},icon:(0,m.jsx)(l.REV,{}),variant:"callAction"})})]})}),(0,m.jsx)(l.Mxu,{variant:"constrained",children:(0,m.jsxs)(l.xA9,{item:!0,xs:12,style:{height:"calc(100vh - 195px)"},children:[a&&(0,m.jsx)(l.z21,{}),!a&&(0,m.jsxs)(n.Fragment,{children:[0!==C.length&&(0,m.jsxs)(n.Fragment,{children:[(0,m.jsx)(l.xA9,{item:!0,xs:12,style:{display:"flex",justifyContent:"flex-end",marginBottom:10},children:(0,m.jsx)("div",{style:{maxWidth:200,width:"95%",display:"flex",flexDirection:"row",alignItems:"center"},children:(0,m.jsx)(l.l6P,{id:"sort-by",label:"Sort by",value:v,onChange:e=>{y(e)},name:"sort-by",options:[{label:"Name",value:"name"},{label:"Capacity",value:"capacity"},{label:"Usage",value:"usage"},{label:"Active Status",value:"active_status"},{label:"Failing Status",value:"failing_status"}],noLabelMinWidth:!0})})}),(0,m.jsx)(w,{rowRenderFunction:e=>{const t=C[e]||null;return t?(0,m.jsx)(j,{tenant:t}):null},totalItems:C.length})]}),0===C.length&&(0,m.jsx)(l.xA9,{container:!0,sx:{display:"flex",justifyContent:"center",alignItems:"center"},children:(0,m.jsx)(l.xA9,{item:!0,xs:8,children:(0,m.jsx)(l.lVp,{iconComponent:(0,m.jsx)(l.fmr,{}),title:"Tenants",help:(0,m.jsxs)(n.Fragment,{children:["Tenant is the logical structure to represent a MinIO deployment. A tenant can have different size and configurations from other tenants, even a different storage class.",(0,m.jsx)("br",{}),(0,m.jsx)("br",{}),"To get started,\xa0",(0,m.jsx)(l.t53,{onClick:()=>{t("/tenants/add")},children:"Create a Tenant."})]})})})})]})]})})]})}},6129:(e,t,a)=>{a.d(t,{A:()=>h});a(5043);var n=a(9923),l=a(6568),s=a(3439),i=a(7869),r=a(6483),o=a(579);const c=e=>{let{totalValue:t,sizeItems:a,bgColor:n="#ededed"}=e;return(0,o.jsx)("div",{style:{width:"100%",height:12,backgroundColor:n,borderRadius:30,display:"flex",transitionDuration:"0.3s",overflow:"hidden"},children:a.map(((e,a)=>{const n=100*e.value/t;return(0,o.jsx)("div",{style:{width:`${n}%`,height:"100%",backgroundColor:e.color,transitionDuration:"0.3s"}},`itemSize-${a.toString()}`)}))})};var d=a(4574),u=a(3097),x=a.n(u);const h=e=>{let{totalCapacity:t,usedSpaceVariants:a,statusClass:u,render:h="pie"}=e;const m=["#8dacd3","#bca1ea","#92e8d2","#efc9ac","#97f274","#f7d291","#71ACCB","#f28282","#e28cc1","#2781B0"],g=(0,d.DP)(),p=`${x()(g,"borderColor","#ededed")}70`,f=a.reduce(((e,t)=>e+t.value),0),v=t-f;let j=[];const y=a.find((e=>"STANDARD"===e.variant))||{value:0,variant:"empty"};if(a.length>10){j=[{value:f-y.value,color:"#2781B0",label:"Total Tiers Space"}]}else j=a.filter((e=>"STANDARD"!==e.variant)).map(((e,t)=>({value:e.value,color:m[t],label:`Tier - ${e.variant}`})));let C=x()(g,"signalColors.main","#07193E");const b=100*y.value/t;b>=90?C=x()(g,"signalColors.danger","#C83B51"):b>=75&&(C=x()(g,"signalColors.warning","#FFAB0F"));const A=[{value:y.value,color:C,label:"Used Space by Tenant"},...j,{value:v,color:"bar"===h?p:"transparent",label:"Empty Space"}];if("bar"===h){const e=A.map((e=>({value:e.value,color:e.color,itemName:e.label})));return(0,o.jsx)("div",{style:{width:"100%",marginBottom:15},children:(0,o.jsx)(c,{totalValue:t,sizeItems:e,bgColor:p})})}return(0,o.jsxs)("div",{style:{position:"relative",width:110,height:110},children:[(0,o.jsx)("div",{style:{position:"absolute",right:-5,top:15,zIndex:400},className:u,children:(0,o.jsx)(n.GQ2,{style:{border:"#fff 2px solid",borderRadius:"100%",width:20,height:20}})}),(0,o.jsx)("span",{style:{position:"absolute",top:"50%",left:"50%",transform:"translate(-50%, -50%)",fontWeight:"bold",fontSize:11},children:isNaN(f)?"N/A":(0,r.qO)(f)}),(0,o.jsx)("div",{children:(0,o.jsxs)(l.r,{width:110,height:110,children:[(0,o.jsx)(s.F,{data:[{value:100}],cx:"50%",cy:"50%",dataKey:"value",outerRadius:50,innerRadius:40,fill:p,isAnimationActive:!1,stroke:"none"}),(0,o.jsx)(s.F,{data:A,cx:"50%",cy:"50%",dataKey:"value",outerRadius:50,innerRadius:40,children:A.map(((e,t)=>(0,o.jsx)(i.f,{fill:e.color,stroke:"none"},`cellCapacity-${t}`)))})]})})]})}}}]);
//# sourceMappingURL=890.bc74a345.chunk.js.map