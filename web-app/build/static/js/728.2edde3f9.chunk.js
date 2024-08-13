"use strict";(self.webpackChunkweb_app=self.webpackChunkweb_app||[]).push([[728],{6746:(e,n,a)=>{a.d(n,{A:()=>o});var s=a(5043),t=a(9923),c=a(579);const l=e=>{const{event:n}=e,[a,l]=s.useState(!1);return(0,c.jsxs)(s.Fragment,{children:[(0,c.jsxs)(t.Hjg,{sx:{cursor:"pointer"},children:[(0,c.jsx)(t.TlP,{scope:"row",onClick:()=>l(!a),sx:{borderBottom:0},children:n.event_type}),(0,c.jsx)(t.nA6,{onClick:()=>l(!a),sx:{borderBottom:0},children:n.reason}),(0,c.jsx)(t.nA6,{onClick:()=>l(!a),sx:{borderBottom:0},children:n.seen}),(0,c.jsx)(t.nA6,{onClick:()=>l(!a),sx:{borderBottom:0},children:n.message.length>=30?"".concat(n.message.slice(0,30),"..."):n.message}),(0,c.jsx)(t.nA6,{onClick:()=>l(!a),sx:{borderBottom:0},children:a?(0,c.jsx)(t.FUY,{}):(0,c.jsx)(t.QpL,{})})]}),(0,c.jsx)(t.Hjg,{children:(0,c.jsx)(t.nA6,{style:{paddingBottom:0,paddingTop:0},colSpan:5,children:a&&(0,c.jsx)(t.azJ,{useBackground:!0,sx:{padding:10,marginBottom:10},children:n.message})})})]})},o=e=>{let{events:n,loading:a}=e;return a?(0,c.jsx)(t.z21,{}):(0,c.jsx)(t.azJ,{withBorders:!0,customBorderPadding:"0px",children:(0,c.jsxs)(t.XIK,{"aria-label":"collapsible table",children:[(0,c.jsx)(t.ndF,{children:(0,c.jsxs)(t.Hjg,{children:[(0,c.jsx)(t.nA6,{children:"Type"}),(0,c.jsx)(t.nA6,{children:"Reason"}),(0,c.jsx)(t.nA6,{children:"Age"}),(0,c.jsx)(t.nA6,{children:"Message"}),(0,c.jsx)(t.nA6,{})]})}),(0,c.jsx)(t.BFY,{children:n.map((e=>(0,c.jsx)(l,{event:e},"".concat(e.event_type,"-").concat(e.seen))))})]})})}},3728:(e,n,a)=>{a.r(n),a.d(n,{default:()=>v});var s=a(5043),t=a(9923),c=a(3216),l=a(5475),o=a(4159),r=a(2961),i=a(6483),d=a(6746),m=a(649),x=a(579);const j={display:"grid",gridTemplateColumns:"2fr 1fr",gridAutoFlow:"row",gap:2,padding:15,["@media (max-width: ".concat(t.nmC.sm,"px)")]:{gridTemplateColumns:"1fr",gridAutoFlow:"dense"}},p=e=>{let{title:n}=e;return(0,x.jsx)(t.azJ,{sx:{borderBottom:"1px solid #eaeaea",margin:0,marginBottom:"20px"},children:(0,x.jsx)("h3",{children:n})})},b=e=>{let{describeInfo:n}=e;return(0,x.jsx)(s.Fragment,{children:(0,x.jsxs)("div",{id:"pvc-describe-summary-content",children:[(0,x.jsx)(p,{title:"Summary"}),(0,x.jsxs)(t.azJ,{sx:{...j},children:[(0,x.jsx)(t.mZW,{label:"Name",value:n.name}),(0,x.jsx)(t.mZW,{label:"Namespace",value:n.namespace}),(0,x.jsx)(t.mZW,{label:"Capacity",value:n.capacity}),(0,x.jsx)(t.mZW,{label:"Status",value:n.status}),(0,x.jsx)(t.mZW,{label:"Storage Class",value:n.storageClass}),(0,x.jsx)(t.mZW,{label:"Access Modes",value:n.accessModes.join(", ")}),(0,x.jsx)(t.mZW,{label:"Finalizers",value:n.finalizers.join(", ")}),(0,x.jsx)(t.mZW,{label:"Volume",value:n.volume}),(0,x.jsx)(t.mZW,{label:"Volume Mode",value:n.volumeMode})]})]})})},u=e=>{let{annotations:n}=e;return(0,x.jsx)(s.Fragment,{children:(0,x.jsxs)("div",{id:"pvc-describe-annotations-content",children:[(0,x.jsx)(p,{title:"Annotations"}),(0,x.jsx)(t.azJ,{children:n.map(((e,n)=>(0,x.jsx)(t.vwO,{id:"".concat(e.key,"-").concat(e.value),sx:{margin:"0.5%"},label:"".concat(e.key,": ").concat(e.value)},n)))})]})})},h=e=>{let{labels:n}=e;return(0,x.jsx)(s.Fragment,{children:(0,x.jsxs)("div",{id:"pvc-describe-labels-content",children:[(0,x.jsx)(p,{title:"Labels"}),(0,x.jsx)(t.azJ,{children:n.map(((e,n)=>(0,x.jsx)(t.vwO,{id:"".concat(e.key,"-").concat(e.value),sx:{margin:"0.5%"},label:"".concat(e.key,": ").concat(e.value)},n)))})]})})},g=e=>{let{tenant:n,namespace:a,pvcName:c,propLoading:l}=e;const[i,d]=(0,s.useState)(),[j,p]=(0,s.useState)(!0),[g,v]=(0,s.useState)("pvc-describe-summary"),C=(0,r.jL)();return(0,s.useEffect)((()=>{l&&p(!0)}),[l]),(0,s.useEffect)((()=>{j&&m.A.invoke("GET","/api/v1/namespaces/".concat(a,"/tenants/").concat(n,"/pvcs/").concat(c,"/describe")).then((e=>{d(e),p(!1)})).catch((e=>{C((0,o.C9)(e)),p(!1)}))}),[j,c,a,n,C]),(0,x.jsx)(s.Fragment,{children:i&&(0,x.jsx)(t.tUM,{currentTabOrPath:g,onTabClick:e=>{v(e)},options:[{tabConfig:{id:"pvc-describe-summary",label:"Summary"},content:(0,x.jsx)(b,{describeInfo:i})},{tabConfig:{id:"pvc-describe-annotations",label:"Annotations"},content:(0,x.jsx)(u,{annotations:i.annotations})},{tabConfig:{id:"pvc-describe-labels",label:"Labels"},content:(0,x.jsx)(h,{labels:i.labels})}],horizontal:!0,horizontalBarBackground:!1})})},v=()=>{const e=(0,r.jL)(),{tenantName:n,PVCName:a,tenantNamespace:j}=(0,c.g)(),[p,b]=(0,s.useState)("simple-tab-0"),[u,h]=(0,s.useState)(!0),[v,C]=(0,s.useState)([]);return(0,s.useEffect)((()=>{u&&m.A.invoke("GET","/api/v1/namespaces/".concat(j,"/tenants/").concat(n,"/pvcs/").concat(a,"/events")).then((e=>{for(let n=0;n<e.length;n++){let a=Date.now()/1e3|0;e[n].seen=(0,i.hr)((a-e[n].last_seen).toString())}C(e),h(!1)})).catch((n=>{e((0,o.C9)(n)),h(!1)}))}),[u,a,j,n,e]),(0,x.jsxs)(s.Fragment,{children:[(0,x.jsxs)(t._xt,{separator:!0,sx:{marginBottom:15},children:[(0,x.jsx)(l.N_,{to:"/namespaces/".concat(j,"/tenants/").concat(n,"/volumes"),children:"PVCs"})," ","> ",a]}),(0,x.jsx)(t.tUM,{options:[{tabConfig:{id:"simple-tab-0",label:"Events"},content:(0,x.jsxs)(s.Fragment,{children:[(0,x.jsx)(t._xt,{separator:!0,sx:{marginBottom:15},children:"Events"}),(0,x.jsx)(d.A,{events:v,loading:u})]})},{tabConfig:{id:"simple-tab-1",label:"Describe"},content:(0,x.jsx)(g,{tenant:n||"",namespace:j||"",pvcName:a||"",propLoading:u})}],currentTabOrPath:p,onTabClick:e=>{b(e)},horizontal:!0})]})}}}]);
//# sourceMappingURL=728.2edde3f9.chunk.js.map