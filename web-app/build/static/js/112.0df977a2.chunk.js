"use strict";(self.webpackChunkweb_app=self.webpackChunkweb_app||[]).push([[112],{112:(e,n,s)=>{s.r(n),s.d(n,{default:()=>h});var t=s(5043),r=s(9456),a=s(9923),o=s(3216),l=s(6483),c=s(2961),d=s(4159),i=s(649),x=s(6746),j=s(579);const h=()=>{const e=(0,c.jL)(),n=(0,o.g)(),s=(0,r.d4)((e=>e.tenants.loadingTenant)),[h,g]=(0,t.useState)([]),[p,m]=(0,t.useState)(!0),u=n.tenantName||"",A=n.tenantNamespace||"";return(0,t.useEffect)((()=>{s&&m(!0)}),[s]),(0,t.useEffect)((()=>{p&&i.A.invoke("GET",`/api/v1/namespaces/${A}/tenants/${u}/events`).then((e=>{for(let n=0;n<e.length;n++){let s=Date.now()/1e3|0;e[n].seen=(0,l.hr)((s-e[n].last_seen).toString())}g(e),m(!1)})).catch((n=>{e((0,d.C9)(n)),m(!1)}))}),[p,A,u,e]),(0,j.jsxs)(t.Fragment,{children:[(0,j.jsx)(a._xt,{separator:!0,sx:{marginBottom:15},children:"Events"}),(0,j.jsx)(a.xA9,{item:!0,xs:12,children:(0,j.jsx)(x.A,{events:h,loading:p})})]})}},6746:(e,n,s)=>{s.d(n,{A:()=>l});var t=s(5043),r=s(9923),a=s(579);const o=e=>{const{event:n}=e,[s,o]=t.useState(!1);return(0,a.jsxs)(t.Fragment,{children:[(0,a.jsxs)(r.Hjg,{sx:{cursor:"pointer"},children:[(0,a.jsx)(r.TlP,{scope:"row",onClick:()=>o(!s),sx:{borderBottom:0},children:n.event_type}),(0,a.jsx)(r.nA6,{onClick:()=>o(!s),sx:{borderBottom:0},children:n.reason}),(0,a.jsx)(r.nA6,{onClick:()=>o(!s),sx:{borderBottom:0},children:n.seen}),(0,a.jsx)(r.nA6,{onClick:()=>o(!s),sx:{borderBottom:0},children:n.message.length>=30?`${n.message.slice(0,30)}...`:n.message}),(0,a.jsx)(r.nA6,{onClick:()=>o(!s),sx:{borderBottom:0},children:s?(0,a.jsx)(r.FUY,{}):(0,a.jsx)(r.QpL,{})})]}),(0,a.jsx)(r.Hjg,{children:(0,a.jsx)(r.nA6,{style:{paddingBottom:0,paddingTop:0},colSpan:5,children:s&&(0,a.jsx)(r.azJ,{useBackground:!0,sx:{padding:10,marginBottom:10},children:n.message})})})]})},l=e=>{let{events:n,loading:s}=e;return s?(0,a.jsx)(r.z21,{}):(0,a.jsx)(r.azJ,{withBorders:!0,customBorderPadding:"0px",children:(0,a.jsxs)(r.XIK,{"aria-label":"collapsible table",children:[(0,a.jsx)(r.ndF,{children:(0,a.jsxs)(r.Hjg,{children:[(0,a.jsx)(r.nA6,{children:"Type"}),(0,a.jsx)(r.nA6,{children:"Reason"}),(0,a.jsx)(r.nA6,{children:"Age"}),(0,a.jsx)(r.nA6,{children:"Message"}),(0,a.jsx)(r.nA6,{})]})}),(0,a.jsx)(r.BFY,{children:n.map((e=>(0,a.jsx)(o,{event:e},`${e.event_type}-${e.seen}`)))})]})})}}}]);
//# sourceMappingURL=112.0df977a2.chunk.js.map