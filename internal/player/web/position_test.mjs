import { readFileSync } from 'node:fs';
import { runInNewContext } from 'node:vm';
import test from 'node:test';
import assert from 'node:assert/strict';
const html=readFileSync(new URL('./musickit.html',import.meta.url),'utf8');
const start=html.indexOf('function notifyStateIfChanged()');
const end=html.indexOf('\n        }',start)+'\n        }'.length;
test('subsecond playback positions reach the lyric timeline',()=>{
 const music={isPlaying:true,playbackState:2,currentPlaybackTime:1.01,nowPlayingItem:{id:'song'},volume:1,repeatMode:0,shuffleMode:0};
 let calls=0;const env={_m:()=>music,_lastStateJSON:'',notifyState:()=>calls++};
 runInNewContext(html.slice(start,end),env);
 env.notifyStateIfChanged();music.currentPlaybackTime=1.08;env.notifyStateIfChanged();
 assert.equal(calls,2);
 env.notifyStateIfChanged();assert.equal(calls,2);
 assert.match(html,/setInterval\(notifyStateIfChanged, 50\)/);
});
