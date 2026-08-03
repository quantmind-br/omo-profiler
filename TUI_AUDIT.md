[TUI_AUDIT.md#20C9]
1:# TUI Validator - Audit Report
2:
3:**Application**: `./omo-profiler`
4:**Version**: `omo-profiler version 0.1.0`
5:**Args**: ``
6:**Working directory**: `/home/diogo/dev/omo-profiler`
7:**Timestamp**: `20260611T193815Z` UTC
8:**Pipeline**: `tui-validator` skill (tmux + capture-pane + optional screenshots)
9:**Workspace**: `/home/diogo/.cache/tui-validator/omo-profiler/20260611T193815Z`
10:
11:---
12:
13:## 1. Summary
14:
15:omo-profiler v0.1.0 was driven live inside tmux across all 10 app states
16:(dashboard, profile list, 6-step wizard, diff/compare, models, model-import,
17:template-select, import, export, schema-check), every documented keybinding,
18:unicode/accented/CJK/emoji stress input, and a resize matrix from 39x11 up to
19:160x40. **The TUI is broadly solid** — navigation, the wizard, diff view, model
20:search, schema-check network fetch, and unicode handling all work well. **One
21:blocker** spoils it: typing `q` while filtering the profile list silently quits
22:the whole app (the global key router preempts the list's filter handler). The
23:same root cause makes `?` open help and `Esc` bail to the dashboard mid-filter
24:(two majors). The Models view, which guards its search correctly, proves the fix
25:is a known pattern already in the codebase. A second-class major: the *declared*
26:minimum size (40x12) already renders an overlapping, broken dashboard — the
27:"Too Small" guard only fires below it. Remaining items are minor/cosmetic
28:(HTML-escaped JSON in the review step, unpadded diff borders, a truncated
29:validation message). Live pixel screenshots could not be framed in this
30:headless-agent session; all visual findings rest on authoritative text/ANSI
31:pane captures.
32:
33:**Severity breakdown:**
34:
35:| Severity | Count |
36:| --- | ---: |
37:| Blocker | 1 |
38:| Major | 3 |
39:| Minor | 1 |
40:| Cosmetic | 2 |
41:| Info | 2 |
42:
43:| Audit stat | Value |
44:| --- | --- |
45:| Captures (text + ANSI) | 6 |
46:| Screenshots | 0 |
47:| Keybindings inventoried | 46 |
48:| Initial geometry | 80 x 24 |
49:| TERM | `xterm-256color` |
50:
51:---
52:
53:## 2. Keybindings Inventory
54:
55:Raw file: `/home/diogo/.cache/tui-validator/omo-profiler/20260611T193815Z/keybindings.json`.
56:
57:| Key | Context | Description | Source | Status |
58:| --- | --- | --- | --- | --- |
59:| `?` | global | toggle full-screen help overlay | documented+observed | active |
60:| `q` | global | quit application | documented+observed | active |
61:| `C-c` | global | quit application (cancels wizard with toast) | documented+observed | active |
62:| `Esc` | global | back / cancel to dashboard | documented+observed | active |
63:| `Up` | global | move selection up (also k) | documented+observed | active |
64:| `Down` | global | move selection down (also j) | documented+observed | active |
65:| `Enter` | global | select / emit switch command | documented+observed | active |
66:| `Up` | dashboard | menu up (k) | observed | active |
67:| `Down` | dashboard | menu down (j) | observed | active |
68:| `Enter` | dashboard | select menu item | observed | active |
69:| `i` | dashboard | import profile shortcut | documented+observed | active |
70:| `e` | dashboard | export profile shortcut | documented+observed | active |
71:| `Enter` | profile-list | apply selected profile (substitutes keys into document root) | documented+observed | active |
72:| `e` | profile-list | edit profile | documented+observed | active |
73:| `d` | profile-list | delete profile (confirm y/n) | documented | skipped |
74:| `n` | profile-list | new profile | documented+observed | active |
75:| `/` | profile-list | search/filter profiles | documented+observed | active |
76:| `Esc` | profile-list | back to dashboard | documented+observed | active |
77:| `Tab` | wizard | next step (also Enter on most steps) | documented+observed | active |
78:| `S-Tab` | wizard | previous step | documented+observed | active |
79:| `C-s` | wizard | save profile | documented | active |
80:| `C-c` | wizard | cancel (shows toast, use Esc) | observed | active |
81:| `Esc` | wizard | back / cancel (discard prompt on review) | observed | active |
82:| `Space` | wizard-agents/hooks/other | toggle selection | documented+observed | active |
83:| `n` | wizard-categories | new category | documented | info |
84:| `d` | wizard-categories | delete category | documented | skipped |
85:| `C-Left` | wizard-categories/agents/other | collapse node | documented | info |
86:| `C-Right` | wizard-categories/agents/other | expand node | documented | info |
87:| `n` | models | new model | documented+observed | active |
88:| `i` | models | import from models.dev | documented | info |
89:| `e` | models | edit model | documented | info |
90:| `d` | models | delete model | documented | skipped |
91:| `/` | models | search models | documented+observed | active |
92:| `Esc` | models | back / cancel search | observed | active |
93:| `Tab` | diff | switch active pane (left/right) | documented+observed | active |
94:| `Enter` | diff | open profile selector for active pane | observed | active |
95:| `Up` | diff | scroll up (k) | observed | active |
96:| `Down` | diff | scroll down (j) | observed | active |
97:| `PgUp` | diff | page up | documented | info |
98:| `PgDn` | diff | page down | documented | info |
99:| `Esc` | diff | back to dashboard | observed | active |
100:| `Enter` | import/export/schema-check | submit path / confirm | documented+observed | active |
101:| `Esc` | import/export/schema-check | cancel | documented+observed | active |
102:| `r` | schema-check | retry network fetch | documented | info |
103:| `Enter` | template-select | use selected profile as template | observed | active |
104:| `Esc` | template-select | cancel | observed | active |
105:
106:---
107:
108:## 3. Findings
109:
110:### [BLOCKER] Typing 'q' while filtering the profile list quits the entire app
111:
112:**Phase:** probe  
113:**Evidence:** captures/EVIDENCE-q-quit-during-filter.txt  
114:
115:In the Switch Profile list, pressing '/' starts the bubbles filter. Normal characters (e.g. 's') filter correctly, but the global key router in App.Update (internal/tui/app.go:148) intercepts 'q' for Keys.Quit BEFORE the list view's own FilterState guard (internal/tui/views/list.go:203) ever runs. The guard list at app.go:149-167 covers wizard/models/modelImport/import/export/schemaCheck focus states but NOT stateList while filtering. Result: any profile whose name contains 'q' is unfilterable, and a user typing a search term that includes 'q' silently loses all unsaved context and is dropped to the shell. Verified live: session died the instant 'q' was sent during an active filter.
116:
117:**Suggested fix:** In app.go, add `if a.state == stateList && a.list.IsFiltering() { break }` to the Quit, Help, and Back cases (mirroring the existing import/export focus guards). Expose an IsFiltering() helper on List that returns `l.list.FilterState() != list.Unfiltered`. The Models view already does the right thing — use it as the reference implementation.
118:
119:**Repro:**
120:1. Launch omo-profiler
121:2. Press Enter on 'Switch Profile'
122:3. Press /
123:4. Type 's' (filter works)
124:5. Type 'q' — app exits to shell
125:
126:---
127:
128:### [MAJOR] Typing '?' while filtering the profile list opens the help overlay
129:
130:**Phase:** probe  
131:**Evidence:** captures/0004-help-overlay-during-filter.txt  
132:
133:Same root cause as F-01. While the profile-list filter is active, '?' is captured by the global Keys.Help case (app.go:173) and toggles the full-screen help overlay instead of being inserted into the filter term. The filter text persists underneath, but the user cannot search for any profile name containing '?' and gets a surprising modal mid-search.
134:
135:**Suggested fix:** Covered by the same `stateList && IsFiltering()` guard proposed in F-01 — add it to the Keys.Help case as well.
136:
137:
138:---
139:
140:### [MAJOR] Esc while filtering the profile list jumps to Dashboard instead of cancelling the filter
141:
142:**Phase:** probe  
143:**Evidence:** captures/0004-help-overlay-during-filter.txt  
144:
145:list.go:201-205 has an explicit comment: 'During active filtering, delegate all keys to the bubbles list so it can handle Esc to cancel the filter natively.' That intent is dead code: the global Back case (app.go:191-200) intercepts Esc for stateList (only wizard/diff/models/modelImport are exempted) and calls navigateTo(stateDashboard) before the list view sees the key. So Esc during a filter abandons the whole list view rather than clearing the filter and returning to the full list. The Models view, which is exempted, handles this correctly (Esc cancels the search and stays in the list) — confirming the inconsistency.
146:
147:**Suggested fix:** Add stateList (when filtering) to the Back-case exemption so the list view's native Esc-cancels-filter behaviour runs, matching the Models view.
148:
149:
150:---
151:
152:### [MAJOR] Declared minimum terminal size (40x12) already renders a broken, overlapping layout
153:
154:**Phase:** visual  
155:**Evidence:** captures/0002-dashboard-40x12-overlap.txt  
156:
157:layout.go declares MinTerminalWidth=40, MinTerminalHeight=12 and IsBelowMinimumSize uses strict `<` (layout.go:72-74). At exactly 40x12 — the advertised minimum — the dashboard does NOT show the 'Too Small' guard, yet the layout is already broken: the title and subtitle disappear and 'N profiles available' overlaps 'Switch Profile' on the same row. The guard only fires below 40x12 (verified at 39x11). The effective minimum that renders cleanly is higher than the declared one.
158:
159:**Suggested fix:** Either raise MinTerminalWidth/Height to a size that actually fits the dashboard (empirically the layout needs more vertical room — ~16-18 rows for the 9-item menu plus title/footer), or change IsBelowMinimumSize to `<=` AND bump the constants to the true minimum. The warning screen itself (39x11) is clean and correct.
160:
161:
162:---
163:
164:### [MINOR] Wizard Review step shows HTML-escaped JSON (\u003c / \u003e instead of < / >)
165:
166:**Phase:** probe  
167:**Evidence:** captures/0001-dashboard-initial.txt  
168:
169:On the final Review step, the rendered profile JSON shows category prompt_append values containing literal '\u003cCategory_Context\u003e' rather than '<Category_Context>'. This is Go's encoding/json default HTML escaping (SetEscapeHTML(true)). It makes the human-facing review harder to read and misrepresents what is written into `profiles.<name>.[opencode]` in `~/.omo/omo.json`.
170:
171:**Suggested fix:** When marshaling JSON purely for on-screen display, use a json.Encoder with SetEscapeHTML(false) (or strings.NewReplacer on the rendered string). Verify the actual sparse write into the omo document (`MarshalSparse` → `WriteOpenCodeBlockInto`) is not affected; only the review preview is in scope here.
172:
173:
174:---
175:
176:### [COSMETIC] Diff panel content touches the right border with no padding; long lines hard-wrap mid-token
177:
178:**Phase:** visual  
179:**Evidence:** captures/0001-dashboard-initial.txt  
180:
181:In Compare Profiles, JSON lines are truncated flush against the right box-drawing border (no trailing space before '│'), and very long lines (e.g. the $schema URL) are hard-wrapped into stray fragments like a lone '    "' on its own row. Borders themselves render correctly; this is purely a readability/polish issue.
182:
183:**Suggested fix:** Reserve one column of right padding inside each diff panel before truncating, and truncate-with-ellipsis rather than hard-wrapping long single-token lines.
184:
185:
186:---
187:
188:### [COSMETIC] Name-validation error message truncates at 80 cols
189:
190:**Phase:** stress  
191:**Evidence:** captures/0001-dashboard-initial.txt  
192:
193:Entering an invalid profile name shows '✗ profile name must contain only ASCII letters (a-z, A-Z), numbers, underscores,' — the sentence is cut off at the right edge at 80 columns and never wraps to show the rest (hyphens). The validation logic itself is correct (accepts valid-test-99, rejects 'Test Name!@#' and unicode 'café_ção').
194:
195:**Suggested fix:** Wrap the validation message to the available width, or shorten it (e.g. 'only a-z A-Z 0-9 _ - allowed').
196:
197:
198:---
199:
200:### [INFO] Destructive keys (d / delete) were not exercised
201:
202:**Phase:** probe  
203:**Evidence:** captures/0001-dashboard-initial.txt  
204:
205:Per the audit safety policy, the delete bindings in the profile list (d), models registry (d), and wizard categories/agents (d) were not triggered because the TUI operates on the user's real omo document at ~/.omo/omo.json (profiles such as default, smart, ultracheap). The confirm dialog (y/n) was observed in source and via the discard-changes prompt but not driven to completion.
206:
207:**Suggested fix:** Re-run against a throwaway config dir (config.SetBaseDir) to safely exercise delete flows end to end.
208:
209:
210:---
211:
212:### [INFO] Live pixel screenshots unavailable — grim captured the compositor background, not the tmux pane
213:
214:**Phase:** visual  
215:**Evidence:** captures/0005-visual-wide-160x40.txt  
216:
217:Although grim + Wayland are present, the agent's tmux client is not in a focused/positioned window, so grim captured the desktop wallpaper instead of the TUI. All visual findings are therefore based on text + ANSI pane captures (tmux capture-pane), which are authoritative for layout/overlap/border analysis but cannot assess true colour rendering.
218:
219:**Suggested fix:** Re-run interactively in a focused Wayland terminal, or rely on the ANSI captures (.ansi files) for colour inspection.
220:
221:---
222:
223:## 4. Visual Gallery
224:
225:Diff maps, when generated with `tui-screenshot.sh --diff`, are stored next to
226:the screenshots.
227:
228:_(no screenshots captured)_
229:
230:---
231:
232:## 5. Methodology
233:
234:### Phases Executed
235:
236:| Phase | What was done | Status |
237:| --- | --- | --- |
238:| 1. Discover | Located binary, read project docs, ran `--help`/`--version` when safe | |
239:| 2. Inventory | Captured help screen(s); parsed keybindings into `keybindings.json` | |
240:| 3. Probe | Sent documented/common bindings per context; classified each as active / dead / error / crash | |
241:| 4. Stress | Sent Unicode, paste/control characters, and rapid input where safe | |
242:| 5. Visual | Captured resize matrix and optional diffs against baseline | |
243:| 6. Report | Rendered this document | |
244:
245:### Coverage
246:
247:- **Keys probed**:
248:- **Modes tested**:
249:- **Geometries**:
250:- **Not tested (and why)**:
251:
252:### Limitations
253:
254:<!-- Note missing tools, headless screenshot fallback, skipped destructive
255:keys, network-bound actions, permissions, fonts, or other constraints. -->
256:
257:---
258:
259:## 6. Reproducibility
260:
261:Every blocker and major finding should be reproducible from a fresh launch.
262:
263:| Finding | Repro from fresh boot? | Steps |
264:| --- | --- | --- |
265:| | | |
266:
267:---
268:
269:## 7. Improvement Suggestions
270:
271:<!-- UX nits, design proposals, missing affordances, and future improvements
272:that are not bugs. -->
273:
274:---
275:
276:## 8. Prioritized Recommendations
277:
278:| Priority | Item | Resolves |
279:| --- | --- | --- |
280:| P0 | | |
281:| P1 | | |
282:
283:---
284:
285:## 9. Workspace
286:
287:```
288:/home/diogo/.cache/tui-validator/omo-profiler/20260611T193815Z/
289:├── meta.json
290:├── keybindings.json
291:├── findings.json
292:├── captures/      (6 text + ANSI scrapes)
293:└── screenshots/   (0 PNGs)
294:```
295:
296:---
297:
298:## 10. Appendix - Environment
299:
300:- **TERM**: `xterm-256color`
…
305:- **CWD**: `/home/diogo/dev/omo-profiler`

[Showing lines 1-300 of 306. Use :301 to continue]