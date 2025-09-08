import{r as f}from"./index-Kjc2dqvG.js";import{a as n,u as l}from"./KcPage-BcwNXE78.js";import{w as g}from"./waitForElementMountedOnDom-qpCjLZnq.js";n();n();function w(r){const{authButtonId:t,kcContext:s,i18n:i}=r,{url:o,isUserIdentified:a,challenge:u,userVerification:c,rpId:m,createTimeout:p}=s,{msgStr:h,isFetchingTranslations:e}=i,{insertScriptTags:d}=l({componentOrHookName:"WebauthnAuthenticate",scriptTags:[{type:"module",textContent:()=>`

                    import { authenticateByWebAuthn } from "${o.resourcesPath}/js/webauthnAuthenticate.js";
                    const authButton = document.getElementById('${t}');
                    authButton.addEventListener("click", function() {
                        const input = {
                            isUserIdentified : ${a},
                            challenge : '${u}',
                            userVerification : '${c}',
                            rpId : '${m}',
                            createTimeout : ${p},
                            errmsg : ${JSON.stringify(h("webauthn-unsupported-browser-text"))}
                        };
                        authenticateByWebAuthn(input);
                    });
                `}]});f.useEffect(()=>{e||(async()=>(await g({elementId:t}),d()))()},[e])}export{w as u};
