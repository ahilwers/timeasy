import{r as O}from"./index-Kjc2dqvG.js";import{a as i,u as b}from"./KcPage-BcwNXE78.js";import{w as J}from"./waitForElementMountedOnDom-qpCjLZnq.js";i();i();function x(n){const{authButtonId:e,kcContext:s,i18n:a}=n,{url:o,challenge:u,userid:c,username:g,signatureAlgorithms:m,rpEntityName:f,rpId:l,attestationConveyancePreference:p,authenticatorAttachment:y,requireResidentKey:d,userVerificationRequirement:h,createTimeout:$,excludeCredentialIds:S}=s,{msgStr:t,isFetchingTranslations:r}=a,{insertScriptTags:N}=b({componentOrHookName:"LoginRecoveryAuthnCodeConfig",scriptTags:[{type:"module",textContent:()=>`
                    import { registerByWebAuthn } from "${o.resourcesPath}/js/webauthnRegister.js";
                    const registerButton = document.getElementById('${e}');
                    registerButton.addEventListener("click", function() {
                        const input = {
                            challenge : '${u}',
                            userid : '${c}',
                            username : '${g}',
                            signatureAlgorithms : ${JSON.stringify(m)},
                            rpEntityName : ${JSON.stringify(f)},
                            rpId : ${JSON.stringify(l)},
                            attestationConveyancePreference : ${JSON.stringify(p)},
                            authenticatorAttachment : ${JSON.stringify(y)},
                            requireResidentKey : ${JSON.stringify(d)},
                            userVerificationRequirement : ${JSON.stringify(h)},
                            createTimeout : ${$},
                            excludeCredentialIds : ${JSON.stringify(S)},
                            initLabel : ${JSON.stringify(t("webauthn-registration-init-label"))},
                            initLabelPrompt : ${JSON.stringify(t("webauthn-registration-init-label-prompt"))},
                            errmsg : ${JSON.stringify(t("webauthn-unsupported-browser-text"))}
                        };
                        registerByWebAuthn(input);
                    });
                `}]});O.useEffect(()=>{r||(async()=>(await J({elementId:e}),N()))()},[r])}export{x as u};
