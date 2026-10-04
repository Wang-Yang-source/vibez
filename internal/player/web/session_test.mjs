import { readFileSync } from 'node:fs';
import { runInNewContext } from 'node:vm';
import test from 'node:test';
import assert from 'node:assert/strict';
const html=readFileSync(new URL('./musickit.html',import.meta.url),'utf8');
function restore(music, token){
 const start=html.indexOf('async function restoreSavedSession(');
 assert.ok(start>=0,'missing saved-session restoration helper');
 const end=html.indexOf('\n      }',start)+'\n      }'.length;
 const env={music};runInNewContext(html.slice(start,end),env);
 return env.restoreSavedSession(token);
}
test('saved session refreshes SDK account region before catalog playback',async()=>{
 const sdk={storefrontId:'cn',storefrontCountryCode:'us',musicUserToken:''};
 const storekit={async requestStorefrontCountryCode(){assert.equal(sdk.musicUserToken,'saved');sdk.storefrontCountryCode='cn';}};
 sdk.getPlaybackController=()=>({storekit});
 await restore(sdk,'saved');
 assert.equal(sdk.storefrontCountryCode,sdk.storefrontId);
});
test('region refresh failure rejects startup rather than declaring playback ready',async()=>{
 const sdk={getPlaybackController:()=>({storekit:{requestStorefrontCountryCode:async()=>{throw new Error('region request failed')}}})};
 await assert.rejects(restore(sdk,'saved'),/region request failed/);
});
