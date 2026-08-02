import { readFileSync } from "node:fs";

const app = readFileSync("internal/api/ui/app.js", "utf8");
const html = readFileSync("internal/api/ui/index.html", "utf8");
const css = readFileSync("internal/api/ui/styles.css", "utf8");

const dictionaryLiteral = app.match(/^const dictionaries = ([\s\S]*?);\n\nconst state =/m)?.[1];
if (!dictionaryLiteral) throw new Error("could not locate UI dictionaries");
const dictionaries = Function(`"use strict"; return (${dictionaryLiteral});`)();

const englishKeys = Object.keys(dictionaries.en).sort();
const chineseKeys = Object.keys(dictionaries["zh-CN"]).sort();
if (englishKeys.join("\n") !== chineseKeys.join("\n")) {
  const english = new Set(englishKeys);
  const chinese = new Set(chineseKeys);
  throw new Error(`translation key mismatch: missing zh=${englishKeys.filter(key => !chinese.has(key))}; missing en=${chineseKeys.filter(key => !english.has(key))}`);
}

const staticTranslationKeys = [...html.matchAll(/\bdata-i18n(?:-aria-label|-placeholder|-label)?="([^"]+)"/g)].map(match => match[1]);
const missingTranslations = [...new Set(staticTranslationKeys.filter(key => !(key in dictionaries.en)))];
if (missingTranslations.length) throw new Error(`missing static translations: ${missingTranslations.join(", ")}`);

const ids = [...html.matchAll(/\bid="([^"]+)"/g)].map(match => match[1]);
const duplicateIDs = ids.filter((id, index) => ids.indexOf(id) !== index);
if (duplicateIDs.length) throw new Error(`duplicate HTML ids: ${[...new Set(duplicateIDs)].join(", ")}`);
const idSet = new Set(ids);

const missingScriptIDs = [...new Set([...app.matchAll(/\bbyId\("([^"]+)"\)/g)].map(match => match[1]).filter(id => !idSet.has(id)))];
if (missingScriptIDs.length) throw new Error(`script references missing ids: ${missingScriptIDs.join(", ")}`);

for (const match of html.matchAll(/\b(?:for|aria-labelledby|aria-describedby)="([^"]+)"/g)) {
  for (const reference of match[1].split(/\s+/)) {
    if (!idSet.has(reference)) throw new Error(`missing referenced id: ${reference}`);
  }
}

for (const match of html.matchAll(/<label\b[^>]*>[\s\S]*?<\/label>/g)) {
  if (/<button\b/.test(match[0])) throw new Error("interactive button must not be nested inside a label");
}
if (!/<button[^>]+id="newJobButton"[^>]*data-i18n-aria-label="newJob"[^>]*>[\s\S]*?<span aria-hidden="true">＋<\/span>/.test(html)) {
  throw new Error("new-job button needs a localized accessible name and a decorative hidden plus");
}

for (const match of html.matchAll(/\bpattern="([^"]+)"/g)) new RegExp(`^(?:${match[1]})$`, "v");

const requiredDynamicKeys = [
  "summaryUnit_folder", "summaryUnit_show", "summaryUnit_season", "summaryUnit_depth",
  "summaryConflict_fail", "summaryConflict_merge_immutable", "summaryVerify_checksum", "summaryVerify_size"
];
for (const key of requiredDynamicKeys) {
  if (!(key in dictionaries.en)) throw new Error(`missing dynamic translation: ${key}`);
}

const requiredBehaviorMarkers = [
  "readDestinations()", "setSettleDuration(", "settleSecondsFromForm()",
  "updateJobFormUI()", "state.jobFormDirty", "resetTokenVisibility()",
  "nextDestinationName()", "state.jobFormDirty || state.busy.save", "event.returnValue = \"\""
];
for (const marker of requiredBehaviorMarkers) {
  if (!app.includes(marker)) throw new Error(`missing form behavior: ${marker}`);
}

for (const marker of ["@media (max-width: 720px)", "@media (max-width: 560px)", "@media (max-height: 500px)", "prefers-reduced-motion"]) {
  if (!css.includes(marker)) throw new Error(`missing responsive/accessibility rule: ${marker}`);
}

if (/media objects/i.test(`${app}\n${html}`)) throw new Error("media-only object wording remains in the general UI");
if (/<section class="review-box"[^>]*aria-live/.test(html)) throw new Error("live preview must not announce the entire review on every input");
if (!/<input id="jobConcurrency"[^>]*\brequired\b/.test(html)) throw new Error("concurrency must be explicitly required");
for (const marker of [
  'name.setAttribute("aria-describedby", nameHint.id)',
  'destinationPath.setAttribute("aria-describedby", pathHint.id)',
  'weight.setAttribute("aria-describedby", weightHint.id)'
]) {
  if (!app.includes(marker)) throw new Error(`missing dynamic destination description: ${marker}`);
}
const beforeUnloadHandler = app.match(/window\.addEventListener\("beforeunload", event => \{([\s\S]*?)\n\}\);/)?.[1];
if (!beforeUnloadHandler) throw new Error("missing beforeunload dirty-form guard");
if (/eventController|stopAnalysisPolling/.test(beforeUnloadHandler)) throw new Error("beforeunload must not tear down live connections before navigation is confirmed");

console.log(`UI contract OK: ${englishKeys.length} bilingual keys, ${ids.length} static ids, ${staticTranslationKeys.length} translated nodes`);
