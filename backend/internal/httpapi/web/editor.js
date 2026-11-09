(() => {
  const el=id=>document.getElementById(id);
  const f={token:el("api-token"),room:el("room-id"),family:el("font-family"),size:el("font-size"),color:el("font-color"),bold:el("bold"),italic:el("italic"),underline:el("underline"),message:el("preview-text")};
  const feedback=el("feedback"),status=el("connection-status");let source="global_default";
  function style(){return{fontFamily:f.family.value,fontSize:Number(f.size.value),color:f.color.value.toUpperCase(),bold:f.bold.checked,italic:f.italic.checked,underline:f.underline.checked}}
  function preview(){const s=style(),p=el("message-preview");p.textContent=f.message.value;p.style.fontFamily=s.fontFamily;p.style.fontSize=s.fontSize+"px";p.style.color=s.color;p.style.fontWeight=s.bold?"700":"400";p.style.fontStyle=s.italic?"italic":"normal";p.style.textDecoration=s.underline?"underline":"none";el("color-value").value=s.color;el("style-summary").textContent=s.fontFamily+" · "+s.fontSize+" px";el("source-label").textContent=source==="room"?"Room override":"Global default"}
  function message(text,kind=""){feedback.textContent=text;feedback.className="feedback "+kind}
  function headers(){const token=f.token.value;if(token.length<32)throw new Error("Enter an API token with at least 32 characters.");return{Authorization:"Bearer "+token,"Content-Type":"application/json"}}
  async function api(path,options={}){const r=await fetch(path,{...options,headers:{...headers(),...(options.headers||{})},cache:"no-store",credentials:"same-origin"});const data=await r.json().catch(()=>({}));if(!r.ok)throw new Error(data.error||"Request failed ("+r.status+").");return data}
  function apply(s,origin){f.family.value=s.fontFamily;f.size.value=String(s.fontSize);f.color.value=s.color;f.bold.checked=s.bold;f.italic.checked=s.italic;f.underline.checked=s.underline;source=origin;preview()}
  function room(){const id=f.room.value.trim();if(!/^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/.test(id))throw new Error("Enter a valid room identifier.");return id}
  async function load(){try{el("load-button").disabled=true;const data=await api("/api/v1/rooms/"+encodeURIComponent(room())+"/message-style");apply(data.style,data.source);status.textContent="Connected";message(data.source==="room"?"Room style loaded.":"No room override; global default loaded.","success")}catch(e){status.textContent="Connection error";message(e.message,"error")}finally{el("load-button").disabled=false}}
  f.room.addEventListener("change",load);f.message.addEventListener("input",preview);[f.family,f.size,f.color,f.bold,f.italic,f.underline].forEach(x=>{x.addEventListener("input",preview);x.addEventListener("change",preview)});
  el("load-button").addEventListener("click",load);
  el("style-form").addEventListener("submit",async e=>{e.preventDefault();try{el("save-button").disabled=true;const data=await api("/api/v1/rooms/"+encodeURIComponent(room())+"/message-style",{method:"PUT",body:JSON.stringify(style())});apply(data.style,data.source);status.textContent="Connected";message("Room style saved.","success")}catch(err){message(err.message,"error")}finally{el("save-button").disabled=false}});
  el("reset-button").addEventListener("click",async()=>{try{el("reset-button").disabled=true;const data=await api("/api/v1/rooms/"+encodeURIComponent(room())+"/message-style",{method:"DELETE"});apply(data.style,data.source);message("Room override removed. Global default is active.","success")}catch(err){message(err.message,"error")}finally{el("reset-button").disabled=false}});
  preview();
})();
